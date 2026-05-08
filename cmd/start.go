package cmd

import (
	"fmt"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewStartCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
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
microcks start --driver [driver you want either 'docker' or 'podman']

# Define name of your microcks container/instance
microcks start --name [name of your container/instance]`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			switch instance.Status {
			case "Running":
				fmt.Printf("Microcks instance with name %s is already running\n", name)
				return nil
			case "Exited":
				containerClient, err := connectors.NewContainerClient(instance.Driver)
				errors.CheckError(err)
				defer containerClient.CloseClient()

				if err := containerClient.StartContainer(instance.ContainerID); err != nil {
					return fmt.Errorf("failed to start container: %w", err)
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
					return fmt.Errorf("failed to create container: %w", err)
				}

				if err := containerClient.StartContainer(containerId); err != nil {
					return fmt.Errorf("failed to start container: %w", err)
				}

				instance.ContainerID = containerId
				instance.AutoRemove = autoRemove
				instance.Name = name
				instance.Image = imageName
				instance.Port = hostPort
				instance.Status = "Running"
				instance.Driver = driver
			}

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

			if err := config.WriteLocalConfig(*localConfig, configFile); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			fmt.Printf("Microcks started successfully at %s\n", server)
			return nil
		},
	}
	startCmd.Flags().StringVar(&name, "name", "microcks", "name for your microcks instance")
	startCmd.Flags().StringVar(&hostPort, "port", "8585", "port to expose Microcks on")
	startCmd.Flags().StringVar(&imageName, "image", "quay.io/microcks/microcks-uber:latest-native", "image which will be used to create a container")
	startCmd.Flags().BoolVar(&autoRemove, "rm", false, "automatically remove the container when it exits")
	startCmd.Flags().StringVar(&driver, "driver", "docker", "container driver to use: 'docker' or 'podman'")
	return startCmd
}
