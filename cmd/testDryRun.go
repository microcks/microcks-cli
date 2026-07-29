/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/testcontainers/testcontainers-go"
	microcks "microcks.io/testcontainers-go"
)

const defaultDryRunImage = "quay.io/microcks/microcks-uber:latest-native"

// hostAccessHostname is the hostname Testcontainers exposes inside the
// container to reach ports on the host machine.
const hostAccessHostname = "host.testcontainers.internal"

type dryRunOptions struct {
	artifact     string
	image        string
	readyTimeout time.Duration
	watch        bool
	driver       string
	params       testParams
}

// configureDriver points testcontainers-go at the right container runtime.
// Docker is its default (honoring DOCKER_HOST); Podman needs its socket wired
// into DOCKER_HOST and Ryuk disabled. An empty driver auto-detects.
func configureDriver(driver string) error {
	switch driver {
	case "podman":
		return setupPodman()
	case "docker":
		return nil // testcontainers-go's default, via DOCKER_HOST
	case "":
		if shouldUsePodman() {
			return setupPodman()
		}
		return nil
	default:
		return fmt.Errorf("unsupported --driver %q (use 'docker' or 'podman')", driver)
	}
}

func shouldUsePodman() bool {
	if os.Getenv("DOCKER_HOST") != "" {
		return false // respect an explicitly configured endpoint
	}
	_, podErr := exec.LookPath("podman")
	_, dockErr := exec.LookPath("docker")
	return podErr == nil && dockErr != nil
}

func setupPodman() error {
	if err := connectors.ConfigurePodmanHost(); err != nil {
		return err
	}
	// testcontainers-go silently falls back to Docker when the podman endpoint
	// isn't reachable, which would make "--driver podman" a lie. Verify the
	// connection now and fail loudly instead.
	if err := connectors.PingDockerHost(); err != nil {
		return fmt.Errorf("--driver podman selected but the podman endpoint is not reachable. "+
			"Start it with 'podman machine start' (macOS/Windows) or "+
			"'systemctl --user start podman.socket' (Linux). Underlying error: %w", err)
	}
	// Ryuk (Testcontainers' reaper) needs privileges rootless Podman doesn't
	// grant; our signal-driven Terminate already guarantees cleanup, so disable
	// it for the Podman path.
	return os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func validateDryRunOptions(opts dryRunOptions) error {
	if opts.artifact == "" {
		return fmt.Errorf("--artifact is required with --dry-run")
	}
	if _, err := os.Stat(opts.artifact); err != nil {
		return fmt.Errorf("cannot read --artifact file %q: %v", opts.artifact, err)
	}
	// The uber-native flavor runs without Keycloak, which is what makes the
	// zero-config dry-run possible. Fail fast on other flavors.
	if !strings.Contains(opts.image, "-native") {
		return fmt.Errorf("--dry-run requires the uber-native image variant (got %q). "+
			"Use the default or pass --image with a *-native tag", opts.image)
	}
	return nil
}

// rewriteLocalEndpoint maps a localhost test endpoint to the hostname that
// resolves back to the host from inside the Microcks container. Returns the
// rewritten endpoint, the host port to expose, and whether a rewrite happened.
func rewriteLocalEndpoint(testEndpoint string) (string, int, bool) {
	u, err := url.Parse(testEndpoint)
	if err != nil {
		return testEndpoint, 0, false
	}
	hostname := u.Hostname()
	if hostname != "localhost" && hostname != "127.0.0.1" {
		return testEndpoint, 0, false
	}
	portStr := u.Port()
	if portStr == "" {
		if u.Scheme == "https" {
			portStr = "443"
		} else {
			portStr = "80"
		}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return testEndpoint, 0, false
	}
	u.Host = hostAccessHostname + ":" + portStr
	return u.String(), port, true
}

func runDryRunTest(opts dryRunOptions) bool {
	// Progress/diagnostics go to stderr for machine output formats so stdout
	// carries only the formatted result.
	progress := progressWriter(opts.params.outputFormat)

	if err := validateDryRunOptions(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	// Select the container runtime (docker default, podman wired via DOCKER_HOST).
	if err := configureDriver(opts.driver); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	// Ctrl+C / SIGTERM cancels the context so teardown still runs.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	containerOpts := []testcontainers.ContainerCustomizer{
		microcks.WithMainArtifact(opts.artifact),
	}

	// A localhost test endpoint refers to the user's machine, not the
	// container: expose the port and point Microcks at the host gateway.
	if rewritten, hostPort, ok := rewriteLocalEndpoint(opts.params.testEndpoint); ok {
		fmt.Fprintf(progress, "Test endpoint %s is local: reaching it from the container as %s\n", opts.params.testEndpoint, rewritten)
		opts.params.testEndpoint = rewritten
		containerOpts = append(containerOpts, testcontainers.WithHostPortAccess(hostPort))
	}

	fmt.Fprintf(progress, "Starting ephemeral Microcks container (%s)...\n", opts.image)
	startCtx, startCancel := context.WithTimeout(ctx, opts.readyTimeout)
	defer startCancel()

	container, err := microcks.Run(startCtx, opts.image, containerOpts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start ephemeral Microcks container: %s\n", err)
		fmt.Fprintln(os.Stderr, "Check that the container runtime is running, the port is free and the image is reachable (or raise --ready-timeout).")
		if container != nil {
			terminateContainer(container, progress)
		}
		return false
	}
	defer terminateContainer(container, progress)

	endpoint, err := container.HttpEndpoint(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve ephemeral Microcks endpoint: %s\n", err)
		return false
	}
	fmt.Fprintf(progress, "Ephemeral Microcks is ready at %s\n", endpoint)

	// The uber-native image runs without Keycloak: a headless client with
	// the unauthenticated token is enough.
	mc := connectors.NewMicrocksClient(endpoint)
	mc.SetOAuthToken("unauthenticated-token")

	success, testResultID, err := runTestAndWait(mc, opts.params)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	if !opts.watch {
		return success
	}
	printDetailsLink(progress, endpoint, testResultID)
	return watchAndRerun(ctx, mc, endpoint, opts)
}

func terminateContainer(container *microcks.MicrocksContainer, progress io.Writer) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fmt.Fprintln(progress, "Tearing down ephemeral Microcks container...")
	if err := container.Terminate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to terminate container %s: %s\n", container.GetContainerID(), err)
	}
}

func watchAndRerun(ctx context.Context, mc connectors.MicrocksClient, serverAddr string, opts dryRunOptions) bool {
	progress := progressWriter(opts.params.outputFormat)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file watcher: %s\n", err)
		return false
	}
	defer watcher.Close()

	// Watch the directory, not the file: editors replace files on save
	// (rename + create), which silently drops a watch set on the file itself.
	artifactPath, err := filepath.Abs(opts.artifact)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve artifact path: %s\n", err)
		return false
	}
	if err := watcher.Add(filepath.Dir(artifactPath)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to watch %s: %s\n", filepath.Dir(artifactPath), err)
		return false
	}

	fmt.Fprintf(progress, "\nWatching %s for changes — press Ctrl+C to stop.\n", opts.artifact)

	rerun := make(chan struct{}, 1)
	var debounce *time.Timer

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(progress, "\nStopping watch mode.")
			return true

		case event, ok := <-watcher.Events:
			if !ok {
				return true
			}
			eventPath, err := filepath.Abs(event.Name)
			if err != nil || eventPath != artifactPath {
				continue
			}
			if !event.Op.Has(fsnotify.Write) && !event.Op.Has(fsnotify.Create) && !event.Op.Has(fsnotify.Rename) {
				continue
			}
			// Debounce editor save bursts (write + chmod, save-twice patterns).
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(300*time.Millisecond, func() {
				select {
				case rerun <- struct{}{}:
				default:
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return true
			}
			fmt.Fprintf(os.Stderr, "Watch error: %s\n", err)

		case <-rerun:
			fmt.Fprintln(progress, strings.Repeat("-", 60))
			fmt.Fprintf(progress, "Artifact changed, re-importing %s ...\n", opts.artifact)
			if _, err := mc.UploadArtifact(opts.artifact, true); err != nil {
				// Invalid spec mid-edit is normal in a TDD loop: report and
				// keep watching, the next valid save recovers.
				fmt.Fprintf(os.Stderr, "Re-import failed, waiting for next change: %s\n", err)
				continue
			}
			success, testResultID, err := runTestAndWait(mc, opts.params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Test run failed, waiting for next change: %s\n", err)
				continue
			}
			printDetailsLink(progress, serverAddr, testResultID)
			if success {
				fmt.Fprintln(progress, "Contract test PASSED — waiting for next change.")
			} else {
				fmt.Fprintln(progress, "Contract test FAILED — waiting for next change.")
			}
		}
	}
}

func printDetailsLink(progress io.Writer, serverAddr, testResultID string) {
	fmt.Fprintf(progress, "Test details (live while watching): %s/#/tests/%s\n", serverAddr, testResultID)
}
