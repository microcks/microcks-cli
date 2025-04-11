package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	var (
		hostIP     string = "0.0.0.0"
		name       string
		hostPort   string
		imageName  string
		autoRemove bool
	)
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "start microcks instance",
		Example: `# Start a Microcks instance
microcks start

# Define your port (by default 8585)
microcks start --port [Port you want]`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.EnsureConfig(config.ConfigPath)

			if err != nil {
				log.Fatalf("Error loading config: %v", err)
			}

			if cfg.Instance.Status == "Running" {
				fmt.Println("Microcks is already running.")
				return
			}

			if cfg.Instance.Status == "Stopped" {
				if err := startContainer(cfg.Instance.ContainerID); err != nil {
					fmt.Errorf("failed to start container: %v", err)
				}
				fmt.Println("Microcks started successfully...")
				return
			}

			cfg.Instance.Name = name
			cfg.Instance.Image = imageName
			cfg.Instance.Port = hostPort
			cfg.Instance.AutoRemove = autoRemove

			containerID, err := createContainer(cfg, hostIP)

			if err != nil {
				log.Fatalf("Failed to create a container: %v", err)
			}
			cfg.Instance.ContainerID = containerID

			if err := startContainer(cfg.Instance.ContainerID); err != nil {
				fmt.Errorf("failed to start container: %v", err)
			}

			err = config.SaveConfig(config.ConfigPath, cfg)

			if err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}

			fmt.Printf("Microcks started successfully...")
		},
	}
	startCmd.Flags().StringVar(&name, "name", "microcks", "name for you microcks instance")
	startCmd.Flags().StringVar(&hostPort, "port", "8585", "")
	startCmd.Flags().StringVar(&imageName, "image", "quay.io/microcks/microcks-uber:latest-native", "image which will be used to create a container")
	startCmd.Flags().BoolVar(&autoRemove, "rm", false, "mimic of '--rm' flag of dokcer to automatically remove the container when it exits")
	return startCmd
}

func createContainer(cfg *config.Config, hostIP string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()

	// Define exposed port and bindings
	exposedPort, _ := nat.NewPort("tcp", "8080")
	portBindings := nat.PortMap{
		exposedPort: []nat.PortBinding{
			{
				HostIP:   hostIP,
				HostPort: cfg.Instance.Port,
			},
		},
	}

	out, err := cli.ImagePull(ctx, cfg.Instance.Image, image.PullOptions{})
	if err != nil {
		return "", err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        cfg.Instance.Image,
			ExposedPorts: nat.PortSet{exposedPort: struct{}{}},
		},
		&container.HostConfig{
			PortBindings: portBindings,
			AutoRemove:   cfg.Instance.AutoRemove,
		}, nil, nil, cfg.Instance.Name)

	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func startContainer(cotainerID string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ContainerStart(ctx, cotainerID, container.StartOptions{}); err != nil {
		panic(err)
	}

	return nil
}
