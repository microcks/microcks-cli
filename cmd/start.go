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
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewStartCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		name         string
		hostPort     string
		imageName    string
		autoRemove   bool
		driver       string
		readyTimeout time.Duration
		noWait       bool
		outputFormat string
	)
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "start microcks instance",
		Example: `# Start a Microcks instance
microcks start

# Define your port (by default 8585)
microcks start --port [Port you want]

# Define your driver (by default docker)
microcks start --driver [driver you wnat either 'docker' or 'podman']

# Define name of your microcks container/instance
microcks start --name [name of you container/instance]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}
			progress := progressWriter(outputFormat)

			configFile := globalClientOpts.ConfigPath
			localConfig, err := config.ReadLocalConfig(configFile)
			if err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}

			if localConfig == nil {
				localConfig = &config.LocalConfig{}
			}

			instance, err := localConfig.GetInstance(name)
			if err != nil {
				instance = &config.Instance{}
			}

			// The recorded status can drift from reality: a system restart,
			// autoRemove or a manual `docker rm` removes the container while
			// the config still says Running/Exited. Reconcile before trusting it.
			if instance.Status != "" && instance.ContainerID != "" {
				instanceDriver := instance.Driver
				if instanceDriver == "" {
					instanceDriver = driver
				}
				containerClient, err := connectors.NewContainerClient(instanceDriver)
				if err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
				exists, err := containerClient.ContainerExists(instance.ContainerID)
				closeErr := containerClient.CloseClient()
				if err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
				if closeErr != nil {
					return errors.Wrap(errors.KindEnvironment, fmt.Errorf("closing container client: %w", closeErr))
				}
				if !exists {
					if _, err := fmt.Fprintf(progress, "Container for instance %s no longer exists, recreating it\n", name); err != nil {
						return errors.Wrap(errors.KindEnvironment, err)
					}
					instance.Status = ""
					instance.ContainerID = ""
				}
			}

			switch instance.Status {
			case "Running":
				server := fmt.Sprintf("http://localhost:%s", instance.Port)
				return writeStartResult(outputFormat, instanceStartResult{
					Name: name, Server: server, Context: server, Status: "running",
				})
			case "Exited":
				containerClient, err := connectors.NewContainerClient(instance.Driver)
				if err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
				if err := containerClient.StartContainer(instance.ContainerID); err != nil {
					if closeErr := containerClient.CloseClient(); closeErr != nil {
						return errors.Wrapf(errors.KindEnvironment, "failed to start container: %v; closing container client: %v", err, closeErr)
					}
					return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to start container: %w", err))
				}
				if err := containerClient.CloseClient(); err != nil {
					return errors.Wrap(errors.KindEnvironment, fmt.Errorf("closing container client: %w", err))
				}
				instance.Status = "Running"
			default:
				containerClient, err := connectors.NewContainerClient(driver)
				if err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
				containerId, err := containerClient.CreateContainer(connectors.ContainerOpts{
					Image:      imageName,
					Port:       hostPort,
					Name:       name,
					AutoRemove: autoRemove,
					Output:     progress,
				})
				if err != nil {
					if closeErr := containerClient.CloseClient(); closeErr != nil {
						return errors.Wrapf(errors.KindEnvironment, "failed to create container: %v; closing container client: %v", err, closeErr)
					}
					return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to create container: %w", err))
				}

				if err := containerClient.StartContainer(containerId); err != nil {
					if closeErr := containerClient.CloseClient(); closeErr != nil {
						return errors.Wrapf(errors.KindEnvironment, "failed to start container: %v; closing container client: %v", err, closeErr)
					}
					return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to start container: %w", err))
				}
				if err := containerClient.CloseClient(); err != nil {
					return errors.Wrap(errors.KindEnvironment, fmt.Errorf("closing container client: %w", err))
				}

				instance.ContainerID = containerId
				instance.AutoRemove = autoRemove
				instance.Name = name
				instance.Image = imageName
				instance.Port = hostPort
				instance.Status = "Running"
				instance.Driver = driver
			}

			//Store config and change context
			localConfig.UpsertInstance(config.Instance{
				ContainerID: instance.ContainerID,
				Name:        instance.Name,
				Image:       instance.Image,
				Port:        instance.Port,
				Status:      instance.Status,
				Driver:      instance.Driver,
				AutoRemove:  instance.AutoRemove,
			})

			server := fmt.Sprintf("http://localhost:%s", instance.Port)

			localConfig.UpsertServer(config.Server{
				Name:           name,
				Server:         server,
				InsecureTLS:    true,
				KeycloakEnable: false,
			})

			localConfig.UpsertAuth(config.Auth{
				Server:       server,
				ClientId:     "",
				ClientSecret: "",
			})

			localConfig.UpsertUser(config.User{
				Name:         server,
				AuthToken:    "",
				RefreshToken: "",
			})

			localConfig.CurrentContext = server
			localConfig.UpsertContext(config.ContextRef{
				Name:     server,
				Server:   server,
				User:     server,
				Instance: instance.Name,
			})

			// Save configs to config file
			if err := config.WriteLocalConfig(*localConfig, configFile); err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}

			// The container being up doesn't mean the Microcks server inside
			// is serving traffic yet: wait until HTTP is actually answering
			// so chained commands (import, test) don't race the boot.
			if !noWait {
				if _, err := fmt.Fprintf(progress, "Waiting for Microcks to be ready at %s ...\n", server); err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
				if err := waitForReady(server, readyTimeout); err != nil {
					return errors.Wrapf(errors.KindEnvironment, "Microcks container is started but the server is not ready: %v. "+
						"It may still be booting — retry shortly or raise --ready-timeout", err)
				}
			}

			return writeStartResult(outputFormat, instanceStartResult{
				Name: name, Server: server, Context: server, Status: "running",
			})
		},
	}
	startCmd.Flags().StringVar(&name, "name", "microcks", "name for your Microcks instance")
	startCmd.Flags().StringVar(&hostPort, "port", "8585", "Host port to expose Microcks")
	startCmd.Flags().StringVar(&imageName, "image", "quay.io/microcks/microcks-uber:latest-native", "image which will be used to create a container")
	startCmd.Flags().BoolVar(&autoRemove, "rm", false, "mimic of '--rm' flag of Docker to automatically remove the container when it exits")
	startCmd.Flags().StringVar(&driver, "driver", "docker", "use --driver to change driver from docker to podman")
	startCmd.Flags().DurationVar(&readyTimeout, "ready-timeout", 60*time.Second, "how long to wait for the Microcks server to be ready before failing")
	startCmd.Flags().BoolVar(&noWait, "no-wait", false, "return as soon as the container is started, without waiting for the Microcks server to be ready")
	startCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return startCmd
}

type instanceStartResult struct {
	Name    string `json:"name"`
	Server  string `json:"server"`
	Context string `json:"context"`
	Status  string `json:"status"`
}

func writeStartResult(outputFormat string, result instanceStartResult) error {
	if outputFormat == "json" {
		return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, result))
	}
	_, err := fmt.Printf("Microcks started successfully at %s\n", result.Server)
	return errors.Wrap(errors.KindEnvironment, err)
}

// waitForReady polls the Microcks API until it answers with 200 or the
// timeout elapses. HTTP being up is the signal users care about — the
// Spring Boot app inside the container takes a while after the container
// process itself is running.
func waitForReady(serverURL string, timeout time.Duration) error {
	url := serverURL + "/api/keycloak/config"
	httpClient := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := httpClient.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("not ready after %s", timeout)
}
