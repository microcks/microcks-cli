package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

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
			cfg, err := config.EnsureConfig(config.ConfigPath)

			if err != nil {
				log.Fatalf("Error loading config: %v", err)
			}

			cfg.Instance.Driver = driver

			cli, err := createClient(cfg.Instance.Driver)

			if err != nil {
				fmt.Println(err)
				return
			}

			defer cli.Close()

			if cfg.Instance.Status == "Running" {
				fmt.Println("Microcks is already running.")
				return
			}

			if cfg.Instance.Status == "Stopped" || cfg.Instance.Status == "Created" {
				if err := startContainer(cfg.Instance.ContainerID, cli); err != nil {
					fmt.Errorf("failed to start container: %v", err)
				}
				fmt.Println("Microcks started successfully...")
				return
			}

			cfg.Instance.Name = name
			cfg.Instance.Image = imageName
			cfg.Instance.Port = hostPort
			cfg.Instance.AutoRemove = autoRemove

			containerID, err := createContainer(cfg, hostIP, cli)

			if err != nil {
				log.Fatalf("Failed to create a container: %v", err)
				return
			}
			cfg.Instance.ContainerID = containerID
			cfg.Instance.Status = "Created"

			if err := startContainer(cfg.Instance.ContainerID, cli); err != nil {
				fmt.Errorf("failed to start container: %v", err)
				return
			}
			cfg.Instance.Status = "Running"
			err = config.SaveConfig(config.ConfigPath, cfg)

			if err != nil {
				log.Fatalf("Failed to save config: %v", err)
				return
			}

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

func createClient(driver string) (*client.Client, error) {

	if driver != "docker" {
		out, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}").Output()
		if err != nil {
			fmt.Println(err)
		}

		err = os.Setenv("DOCKER_HOST", "unix://"+strings.TrimSpace(string(out)))
		if err != nil {
			fmt.Println(err)
		}
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return cli, nil
}

func createContainer(cfg *config.Config, hostIP string, cli *client.Client) (string, error) {
	ctx := context.Background()

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

func startContainer(cotainerID string, cli *client.Client) error {
	ctx := context.Background()

	if err := cli.ContainerStart(ctx, cotainerID, container.StartOptions{}); err != nil {
		panic(err)
	}

	return nil
}
