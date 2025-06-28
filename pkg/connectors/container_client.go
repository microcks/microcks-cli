package connectors

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/microcks/microcks-cli/pkg/errors"
)

type ContainerClient interface {
	CreateContainer(opts ContainerOpts) (string, error)
	StartContainer(containerId string) error
	StopContainer(continerId string) error
	CloseClient() error
}

type containerClient struct {
	cli *client.Client
}

type ContainerOpts struct {
	Image      string
	Port       string
	AutoRemove bool
	Name       string
}

const (
	MICROCKS_DEFAULT_PORT = "8080"
	LOCALHOST_IP          = "127.0.0.1"
)

func NewContainerClient(driver string) (ContainerClient, error) {
	switch driver {
	case "docker":
		return NewDockerClient()

	case "podman":
		return NewPodmanClient()

	default:
		return nil, fmt.Errorf("unsupported container driver: %s", driver)
	}
}

func NewDockerClient() (*containerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return &containerClient{cli: cli}, nil
}

func NewPodmanClient() (*containerClient, error) {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		remoteSocket, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanPipe.Path}}").Output()
		errors.CheckError(err)
		socketPath := strings.TrimSpace(string(remoteSocket))
		err = os.Setenv("DOCKER_HOST", "npipe:////"+strings.TrimPrefix(socketPath, "\\\\"))
		errors.CheckError(err)

	case "darwin":
		remoteSocket, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}").Output()
		errors.CheckError(err)
		err = os.Setenv("DOCKER_HOST", "unix://"+strings.TrimSpace(string(remoteSocket)))
		errors.CheckError(err)
	case "linux":
		remoteSocket, err := exec.Command("podman", "info", "--format", "{{.Host.RemoteSocket.Path}}").Output()
		errors.CheckError(err)
		err = os.Setenv("DOCKER_HOST", "unix://"+strings.TrimSpace(string(remoteSocket)))
		errors.CheckError(err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return &containerClient{cli: cli}, nil
}

func (cli *containerClient) CreateContainer(opts ContainerOpts) (string, error) {
	ctx := context.Background()

	// Define exposed port and bindings
	exposedPort, _ := nat.NewPort("tcp", "8080")
	portBindings := nat.PortMap{
		exposedPort: []nat.PortBinding{
			{
				HostIP:   LOCALHOST_IP,
				HostPort: opts.Port,
			},
		},
	}

	out, err := cli.cli.ImagePull(ctx, opts.Image, image.PullOptions{})
	if err != nil {
		return "", err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	resp, err := cli.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        opts.Image,
			ExposedPorts: nat.PortSet{exposedPort: struct{}{}},
		},
		&container.HostConfig{
			PortBindings: portBindings,
			AutoRemove:   opts.AutoRemove,
		}, nil, nil, opts.Name)

	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (cli *containerClient) StartContainer(containerId string) error {
	ctx := context.Background()
	return cli.cli.ContainerStart(ctx, containerId, container.StartOptions{})
}

func (cli *containerClient) StopContainer(containerId string) error {
	ctx := context.Background()

	fmt.Print("Stopping container ", containerId, "... ")
	noWaitTimeout := 0 // to not wait for the container to exit gracefully
	return cli.cli.ContainerStop(ctx, containerId, container.StopOptions{Timeout: &noWaitTimeout})
}

func (cli *containerClient) CloseClient() error {
	return cli.cli.Close()
}
