package connectors

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/moby/term"
)

type ContainerClient interface {
	CreateContainer(opts ContainerOpts) (string, error)
	StartContainer(containerId string) error
	StopContainer(continerId string) error
	ContainerExists(containerId string) (bool, error)
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

// PingDockerHost verifies the Docker-API endpoint at DOCKER_HOST actually
// answers. The raw client honors DOCKER_HOST with no fallback, so this catches
// an unreachable endpoint (e.g. a stopped Podman machine) before testcontainers-go
// silently resolves to a different runtime.
func PingDockerHost() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = cli.Ping(ctx)
	return err
}

func ConfigurePodmanHost() error {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanPipe.Path}}").Output()
		if err != nil {
			return fmt.Errorf("resolving podman pipe path (is the podman machine running?): %w", err)
		}
		socketPath := strings.TrimSpace(string(out))
		return os.Setenv("DOCKER_HOST", "npipe:////"+strings.TrimPrefix(socketPath, "\\\\"))
	case "darwin":
		out, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}").Output()
		if err != nil {
			return fmt.Errorf("resolving podman socket path (is the podman machine running?): %w", err)
		}
		return os.Setenv("DOCKER_HOST", "unix://"+strings.TrimSpace(string(out)))
	case "linux":
		out, err := exec.Command("podman", "info", "--format", "{{.Host.RemoteSocket.Path}}").Output()
		if err != nil {
			return fmt.Errorf("resolving podman socket path: %w", err)
		}
		return os.Setenv("DOCKER_HOST", "unix://"+strings.TrimSpace(string(out)))
	}
	return nil
}

func NewPodmanClient() (*containerClient, error) {
	if err := ConfigurePodmanHost(); err != nil {
		return nil, err
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

	fd, isTerminal := term.GetFdInfo(os.Stdout)

	err = jsonmessage.DisplayJSONMessagesStream(
		out,
		os.Stdout,
		fd,
		isTerminal,
		nil,
	)
	if err != nil {
		return "", err
	}
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

// ContainerExists reports whether the container is still known to the
// runtime. "No such container" is a regular outcome here, not an error,
// so callers can reconcile stale config entries.
func (cli *containerClient) ContainerExists(containerId string) (bool, error) {
	ctx := context.Background()
	_, err := cli.cli.ContainerInspect(ctx, containerId)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (cli *containerClient) CloseClient() error {
	return cli.cli.Close()
}
