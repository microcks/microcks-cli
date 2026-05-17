package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func waitForReadiness(port string, timeoutSecs int) error {
	healthURL := fmt.Sprintf("http://localhost:%s/api/health", port)
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(time.Duration(timeoutSecs) * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("Microcks did not become ready within %d seconds", timeoutSecs)
}

func NewStartCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		name        string
		hostPort    string
		imageName   string
		autoRemove  bool
		driver      string
		wait        bool
		waitTimeout int
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

			if instance.Status == "Running" {
				fmt.Printf("Microcks instance with name %s is already running", name)
				return
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

			if wait {
				fmt.Print("Waiting for Microcks to be ready")
				if err := waitForReadiness(instance.Port, waitTimeout); err != nil {
					fmt.Printf("\n%s\n", err)
					os.Exit(1)
				}
				fmt.Println()
			}

			fmt.Printf("Microcks started successfully at %s\n", server)
		},
	}
	startCmd.Flags().StringVar(&name, "name", "microcks", "name for your Microcks instance")
	startCmd.Flags().StringVar(&hostPort, "port", "8585", "Host port to expose Microcks")
	startCmd.Flags().StringVar(&imageName, "image", "quay.io/microcks/microcks-uber:latest-native", "image which will be used to create a container")
	startCmd.Flags().BoolVar(&autoRemove, "rm", false, "mimic of '--rm' flag of Docker to automatically remove the container when it exits")
	startCmd.Flags().StringVar(&driver, "driver", "docker", "use --driver to change driver from docker to podman")
	startCmd.Flags().BoolVar(&wait, "wait", false, "wait for Microcks to be ready before returning")
	startCmd.Flags().IntVar(&waitTimeout, "wait-timeout", 30, "seconds to wait for readiness before giving up")
	return startCmd
}
