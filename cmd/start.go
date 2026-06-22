package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
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
		Run: func(cmd *cobra.Command, args []string) {

			configFile := globalClientOpts.ConfigPath
			localConfig, err := config.ReadLocalConfig(configFile)
			errors.CheckError(err)

			if localConfig == nil {
				localConfig = &config.LocalConfig{}
			}

			instance, _ := localConfig.GetInstance(name)
			if instance == nil {
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
				errors.CheckError(err)
				exists, err := containerClient.ContainerExists(instance.ContainerID)
				containerClient.CloseClient()
				errors.CheckError(err)
				if !exists {
					fmt.Printf("Container for instance %s no longer exists, recreating it\n", name)
					localConfig.RemoveInstance(instance.ContainerID)
					instance.Status = ""
				}
			}

			switch instance.Status {
			case "Running":
				fmt.Printf("Microcks instance with name %s is already running", name)
				return
			case "Exited":
				containerClient, err := connectors.NewContainerClient(instance.Driver)
				errors.CheckError(err)
				defer containerClient.CloseClient()

				if err := containerClient.StartContainer(instance.ContainerID); err != nil {
					log.Fatalf("failed to start container: %v", err)
					return
				}
				instance.Status = "Running"
			default:
				containerClient, err := connectors.NewContainerClient(driver)
				errors.CheckError(err)
				defer containerClient.CloseClient()

				containerId, err := containerClient.CreateContainer(connectors.ContainerOpts{
					Image:      imageName,
					Port:       hostPort,
					Name:       name,
					AutoRemove: autoRemove,
				})
				if err != nil {
					log.Fatalf("Failed to create a container: %v", err)
					return
				}

				if err := containerClient.StartContainer(containerId); err != nil {
					log.Fatalf("failed to start container: %v", err)
					return
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
			err = config.WriteLocalConfig(*localConfig, configFile)
			errors.CheckError(err)

			// The container being up doesn't mean the Microcks server inside
			// is serving traffic yet: wait until HTTP is actually answering
			// so chained commands (import, test) don't race the boot.
			if !noWait {
				fmt.Printf("Waiting for Microcks to be ready at %s ...\n", server)
				if err := waitForReady(server, readyTimeout); err != nil {
					log.Fatalf("Microcks container is started but the server is not ready: %v. "+
						"It may still be booting — retry shortly or raise --ready-timeout.", err)
				}
			}

			fmt.Printf("Microcks started successfully at %s\n", server)
		},
	}
	startCmd.Flags().StringVar(&name, "name", "microcks", "name for your Microcks instance")
	startCmd.Flags().StringVar(&hostPort, "port", "8585", "Host port to expose Microcks")
	startCmd.Flags().StringVar(&imageName, "image", "quay.io/microcks/microcks-uber:latest-native", "image which will be used to create a container")
	startCmd.Flags().BoolVar(&autoRemove, "rm", false, "mimic of '--rm' flag of Docker to automatically remove the container when it exits")
	startCmd.Flags().StringVar(&driver, "driver", "docker", "use --driver to change driver from docker to podman")
	startCmd.Flags().DurationVar(&readyTimeout, "ready-timeout", 60*time.Second, "how long to wait for the Microcks server to be ready before failing")
	startCmd.Flags().BoolVar(&noWait, "no-wait", false, "return as soon as the container is started, without waiting for the Microcks server to be ready")
	return startCmd
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
