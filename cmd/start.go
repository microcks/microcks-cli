package cmd

import (
	"fmt"
	"log"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	var (
		name       string
		hostPort   string
		imageName  string
		autoRemove bool
		driver     string
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

			configFile, err := config.DefaultLocalConfigPath()
			errors.CheckError(err)
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
				Name:            name,
				Server:          server,
				InsecureTLS:     true,
				KeycloackEnable: false,
			})

			localConfig.UpserAuth(config.Auth{
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
			localConfig.UpserContext(config.ContextRef{
				Name:     server,
				Server:   server,
				User:     server,
				Instance: instance.Name,
			})

			// Save configs to config file
			err = config.WriteLocalConfig(*localConfig, configFile)
			errors.CheckError(err)

			fmt.Printf("Microcks started successfully...")
		},
	}
	startCmd.Flags().StringVar(&name, "name", "microcks", "name for you microcks instance")
	startCmd.Flags().StringVar(&hostPort, "port", "8585", "")
	startCmd.Flags().StringVar(&imageName, "image", "quay.io/microcks/microcks-uber:latest-native", "image which will be used to create a container")
	startCmd.Flags().BoolVar(&autoRemove, "rm", false, "mimic of '--rm' flag of dokcer to automatically remove the container when it exits")
	startCmd.Flags().StringVar(&driver, "driver", "docker", "use --driver to change driver from docker to podman")
	return startCmd
}
