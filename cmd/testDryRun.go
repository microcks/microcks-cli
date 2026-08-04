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
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
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
		return errors.Wrapf(errors.KindUsage, "unsupported --driver %q (use 'docker' or 'podman')", driver)
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
		return errors.Wrap(errors.KindEnvironment, err)
	}
	// testcontainers-go silently falls back to Docker when the podman endpoint
	// isn't reachable, which would make "--driver podman" a lie. Verify the
	// connection now and fail loudly instead.
	if err := connectors.PingDockerHost(); err != nil {
		return errors.Wrapf(errors.KindEnvironment, "--driver podman selected but the podman endpoint is not reachable. "+
			"Start it with 'podman machine start' (macOS/Windows) or "+
			"'systemctl --user start podman.socket' (Linux). Underlying error: %v", err)
	}
	// Ryuk (Testcontainers' reaper) needs privileges rootless Podman doesn't
	// grant; our signal-driven Terminate already guarantees cleanup, so disable
	// it for the Podman path.
	return os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func validateDryRunOptions(opts dryRunOptions) error {
	if opts.artifact == "" {
		return errors.Wrapf(errors.KindUsage, "--artifact is required with --dry-run")
	}
	if _, err := os.Stat(opts.artifact); err != nil {
		return errors.Wrapf(errors.KindUsage, "cannot read --artifact file %q: %v", opts.artifact, err)
	}
	// The uber-native flavor runs without Keycloak, which is what makes the
	// zero-config dry-run possible. Fail fast on other flavors.
	if !strings.Contains(opts.image, "-native") {
		return errors.Wrapf(errors.KindUsage, "--dry-run requires the uber-native image variant (got %q). "+
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

func runDryRunTest(opts dryRunOptions) (resultErr error) {
	// Progress/diagnostics go to stderr for machine output formats so stdout
	// carries only the formatted result.
	progress := progressWriter(opts.params.outputFormat)
	eventMode := opts.watch && opts.params.outputFormat == string(output.FormatJSON)
	var events *dryRunEventWriter
	if eventMode {
		events = newDryRunEventWriter(os.Stdout)
		defer func() {
			if err := events.emit(dryRunWatchEvent{Type: "stopped"}); err != nil {
				resultErr = errors.Wrapf(
					errors.KindEnvironment,
					"writing dry-run stopped event: %v (previous error: %v)",
					err,
					resultErr,
				)
			}
		}()
	}

	if err := validateDryRunOptions(opts); err != nil {
		return err
	}

	// Select the container runtime (docker default, podman wired via DOCKER_HOST).
	if err := configureDriver(opts.driver); err != nil {
		return err
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
		if _, err := fmt.Fprintf(progress, "Test endpoint %s is local: reaching it from the container as %s\n", opts.params.testEndpoint, rewritten); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
		opts.params.testEndpoint = rewritten
		containerOpts = append(containerOpts, testcontainers.WithHostPortAccess(hostPort))
	}

	if _, err := fmt.Fprintf(progress, "Starting ephemeral Microcks container (%s)...\n", opts.image); err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}
	startCtx, startCancel := context.WithTimeout(ctx, opts.readyTimeout)
	defer startCancel()

	container, err := microcks.Run(startCtx, opts.image, containerOpts...)
	if err != nil {
		if container != nil {
			if terminateErr := terminateContainer(container, progress); terminateErr != nil {
				return errors.Wrapf(
					errors.KindEnvironment,
					"failed to start ephemeral Microcks container: %v; cleanup also failed: %v",
					err,
					terminateErr,
				)
			}
		}
		return errors.Wrapf(errors.KindEnvironment, "failed to start ephemeral Microcks container: %v. "+
			"Check that the container runtime is running, the port is free and the image is reachable (or raise --ready-timeout)", err)
	}
	defer func() {
		if err := terminateContainer(container, progress); err != nil {
			resultErr = errors.Wrapf(
				errors.KindEnvironment,
				"tearing down ephemeral Microcks container: %v (previous error: %v)",
				err,
				resultErr,
			)
		}
	}()

	endpoint, err := container.HttpEndpoint(ctx)
	if err != nil {
		return errors.Wrapf(errors.KindEnvironment, "failed to resolve ephemeral Microcks endpoint: %v", err)
	}
	if _, err := fmt.Fprintf(progress, "Ephemeral Microcks is ready at %s\n", endpoint); err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}
	if events != nil {
		if err := events.emit(dryRunWatchEvent{Type: "ready", Endpoint: endpoint}); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
		if err := events.emit(dryRunWatchEvent{
			Type: "imported", Artifact: opts.artifact, Service: opts.params.serviceRef,
		}); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
	}

	// The uber-native image runs without Keycloak: a headless client with
	// the unauthenticated token is enough.
	mc, err := connectors.NewMicrocksClient(endpoint)
	if err != nil {
		return err
	}
	mc.SetOAuthToken("unauthenticated-token")

	params := opts.params
	params.suppressOutput = eventMode
	success, testResultID, err := runTestAndWait(mc, params)
	if err != nil {
		if emitErr := emitDryRunError(events, err); emitErr != nil {
			return emitErr
		}
		return err
	}
	if events != nil {
		if err := events.emitTestResult(mc, testResultID); err != nil {
			return errors.Wrap(errors.KindAPI, err)
		}
		if err := events.emit(dryRunWatchEvent{Type: "waiting"}); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
	}

	if !opts.watch {
		if success {
			return nil
		}
		return errors.ErrTestFailed
	}
	if err := printDetailsLink(progress, endpoint, testResultID); err != nil {
		return err
	}
	return watchAndRerun(ctx, mc, endpoint, opts, events)
}

func terminateContainer(container *microcks.MicrocksContainer, progress io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := fmt.Fprintln(progress, "Tearing down ephemeral Microcks container..."); err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}
	if err := container.Terminate(ctx); err != nil {
		return errors.Wrapf(
			errors.KindEnvironment,
			"failed to terminate container %s: %v",
			container.GetContainerID(),
			err,
		)
	}
	return nil
}

func watchAndRerun(
	ctx context.Context,
	mc connectors.MicrocksClient,
	serverAddr string,
	opts dryRunOptions,
	events *dryRunEventWriter,
) (resultErr error) {
	progress := progressWriter(opts.params.outputFormat)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to create file watcher: %w", err))
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			resultErr = errors.Wrapf(
				errors.KindEnvironment,
				"closing file watcher: %v (previous error: %v)",
				err,
				resultErr,
			)
		}
	}()

	// Watch the directory, not the file: editors replace files on save
	// (rename + create), which silently drops a watch set on the file itself.
	artifactPath, err := filepath.Abs(opts.artifact)
	if err != nil {
		return errors.Wrap(errors.KindUsage, fmt.Errorf("failed to resolve artifact path: %w", err))
	}
	if err := watcher.Add(filepath.Dir(artifactPath)); err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to watch %s: %w", filepath.Dir(artifactPath), err))
	}

	if _, err := fmt.Fprintf(progress, "\nWatching %s for changes — press Ctrl+C to stop.\n", opts.artifact); err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}

	rerun := make(chan struct{}, 1)
	var debounce *time.Timer

	for {
		select {
		case <-ctx.Done():
			if _, err := fmt.Fprintln(progress, "\nStopping watch mode."); err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
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
				return nil
			}
			if _, writeErr := fmt.Fprintf(os.Stderr, "Watch error: %s\n", err); writeErr != nil {
				return errors.Wrap(errors.KindEnvironment, writeErr)
			}
			if emitErr := emitDryRunError(events, err); emitErr != nil {
				return emitErr
			}

		case <-rerun:
			if _, err := fmt.Fprintln(progress, strings.Repeat("-", 60)); err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}
			if _, err := fmt.Fprintf(progress, "Artifact changed, re-importing %s ...\n", opts.artifact); err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}
			if _, err := mc.UploadArtifact(opts.artifact, true); err != nil {
				// Invalid spec mid-edit is normal in a TDD loop: report and
				// keep watching, the next valid save recovers.
				if _, writeErr := fmt.Fprintf(os.Stderr, "Re-import failed, waiting for next change: %s\n", err); writeErr != nil {
					return errors.Wrap(errors.KindEnvironment, writeErr)
				}
				if emitErr := emitDryRunError(events, err); emitErr != nil {
					return emitErr
				}
				continue
			}
			if events != nil {
				if err := events.emit(dryRunWatchEvent{
					Type: "imported", Artifact: opts.artifact, Service: opts.params.serviceRef,
				}); err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
			}
			params := opts.params
			params.suppressOutput = events != nil
			success, testResultID, err := runTestAndWait(mc, params)
			if err != nil {
				if _, writeErr := fmt.Fprintf(os.Stderr, "Test run failed, waiting for next change: %s\n", err); writeErr != nil {
					return errors.Wrap(errors.KindEnvironment, writeErr)
				}
				if emitErr := emitDryRunError(events, err); emitErr != nil {
					return emitErr
				}
				continue
			}
			if events != nil {
				if err := events.emitTestResult(mc, testResultID); err != nil {
					if emitErr := emitDryRunError(events, err); emitErr != nil {
						return emitErr
					}
					continue
				}
				if err := events.emit(dryRunWatchEvent{Type: "waiting"}); err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
			}
			if err := printDetailsLink(progress, serverAddr, testResultID); err != nil {
				return err
			}
			if success {
				if _, err := fmt.Fprintln(progress, "Contract test PASSED — waiting for next change."); err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
			} else {
				if _, err := fmt.Fprintln(progress, "Contract test FAILED — waiting for next change."); err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
			}
		}
	}
}

func emitDryRunError(events *dryRunEventWriter, sourceErr error) error {
	if events == nil {
		return nil
	}
	if err := events.emit(dryRunWatchEvent{Type: "error", Message: sourceErr.Error()}); err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("writing dry-run error event: %w", err))
	}
	return nil
}

func printDetailsLink(progress io.Writer, serverAddr, testResultID string) error {
	_, err := fmt.Fprintf(progress, "Test details (live while watching): %s/#/tests/%s\n", serverAddr, testResultID)
	return errors.Wrap(errors.KindEnvironment, err)
}
