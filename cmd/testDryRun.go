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
	"net/url"
	"os"
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
	params       testParams
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
	if err := validateDryRunOptions(opts); err != nil {
		fmt.Println(err)
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
		fmt.Printf("Test endpoint %s is local: reaching it from the container as %s\n", opts.params.testEndpoint, rewritten)
		opts.params.testEndpoint = rewritten
		containerOpts = append(containerOpts, testcontainers.WithHostPortAccess(hostPort))
	}

	fmt.Printf("Starting ephemeral Microcks container (%s)...\n", opts.image)
	startCtx, startCancel := context.WithTimeout(ctx, opts.readyTimeout)
	defer startCancel()

	container, err := microcks.Run(startCtx, opts.image, containerOpts...)
	if err != nil {
		fmt.Printf("Failed to start ephemeral Microcks container: %s\n", err)
		fmt.Println("Check that the container runtime is running, the port is free and the image is reachable (or raise --ready-timeout).")
		if container != nil {
			terminateContainer(container)
		}
		return false
	}
	defer terminateContainer(container)

	endpoint, err := container.HttpEndpoint(ctx)
	if err != nil {
		fmt.Printf("Failed to resolve ephemeral Microcks endpoint: %s\n", err)
		return false
	}
	fmt.Printf("Ephemeral Microcks is ready at %s\n", endpoint)

	// The uber-native image runs without Keycloak: a headless client with
	// the unauthenticated token is enough.
	mc := connectors.NewMicrocksClient(endpoint)
	mc.SetOAuthToken("unauthenticated-token")

	success, testResultID, err := runTestAndWait(mc, opts.params)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if !opts.watch {
		return success
	}
	printDetailsLink(endpoint, testResultID)
	return watchAndRerun(ctx, mc, endpoint, opts)
}

func terminateContainer(container *microcks.MicrocksContainer) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fmt.Println("Tearing down ephemeral Microcks container...")
	if err := container.Terminate(ctx); err != nil {
		fmt.Printf("Failed to terminate container %s: %s\n", container.GetContainerID(), err)
	}
}

func watchAndRerun(ctx context.Context, mc connectors.MicrocksClient, serverAddr string, opts dryRunOptions) bool {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Failed to create file watcher: %s\n", err)
		return false
	}
	defer watcher.Close()

	// Watch the directory, not the file: editors replace files on save
	// (rename + create), which silently drops a watch set on the file itself.
	artifactPath, err := filepath.Abs(opts.artifact)
	if err != nil {
		fmt.Printf("Failed to resolve artifact path: %s\n", err)
		return false
	}
	if err := watcher.Add(filepath.Dir(artifactPath)); err != nil {
		fmt.Printf("Failed to watch %s: %s\n", filepath.Dir(artifactPath), err)
		return false
	}

	fmt.Printf("\nWatching %s for changes — press Ctrl+C to stop.\n", opts.artifact)

	rerun := make(chan struct{}, 1)
	var debounce *time.Timer

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nStopping watch mode.")
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
			fmt.Printf("Watch error: %s\n", err)

		case <-rerun:
			fmt.Println(strings.Repeat("-", 60))
			fmt.Printf("Artifact changed, re-importing %s ...\n", opts.artifact)
			if _, err := mc.UploadArtifact(opts.artifact, true); err != nil {
				// Invalid spec mid-edit is normal in a TDD loop: report and
				// keep watching, the next valid save recovers.
				fmt.Printf("Re-import failed, waiting for next change: %s\n", err)
				continue
			}
			success, testResultID, err := runTestAndWait(mc, opts.params)
			if err != nil {
				fmt.Printf("Test run failed, waiting for next change: %s\n", err)
				continue
			}
			printDetailsLink(serverAddr, testResultID)
			if success {
				fmt.Println("Contract test PASSED — waiting for next change.")
			} else {
				fmt.Println("Contract test FAILED — waiting for next change.")
			}
		}
	}
}

func printDetailsLink(serverAddr, testResultID string) {
	fmt.Printf("Test details (live while watching): %s/#/tests/%s\n", serverAddr, testResultID)
}
