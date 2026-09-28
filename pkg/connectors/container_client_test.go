package connectors

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureDockerHost(t *testing.T) {
	oldDockerHost := os.Getenv("DOCKER_HOST")
	defer func() {
		if oldDockerHost == "" {
			os.Unsetenv("DOCKER_HOST")
		} else {
			os.Setenv("DOCKER_HOST", oldDockerHost)
		}
	}()

	oldExecCommand := execCommand
	defer func() {
		execCommand = oldExecCommand
	}()

	t.Run("sets DOCKER_HOST from docker context", func(t *testing.T) {
		os.Unsetenv("DOCKER_HOST")

		execCommand = func(name string, args ...string) *exec.Cmd {
			cmd := exec.Command(
				os.Args[0],
				"-test.run=TestHelperProcess",
				"--",
				"npipe:////./pipe/dockerDesktopLinuxEngine",
			)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
			return cmd
		}

		require.NoError(t, ConfigureDockerHost())
		assert.Equal(t, "npipe:////./pipe/dockerDesktopLinuxEngine", os.Getenv("DOCKER_HOST"))
	})

	t.Run("does not overwrite existing DOCKER_HOST", func(t *testing.T) {
		require.NoError(t, os.Setenv("DOCKER_HOST", "existing-endpoint"))

		require.NoError(t, ConfigureDockerHost())
		assert.Equal(t, "existing-endpoint", os.Getenv("DOCKER_HOST"))
	})

	t.Run("returns error when docker context command fails", func(t *testing.T) {
		os.Unsetenv("DOCKER_HOST")

		execCommand = func(name string, args ...string) *exec.Cmd {
			cmd := exec.Command(
				os.Args[0],
				"-test.run=TestHelperProcessFailure",
			)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
			return cmd
		}

		err := ConfigureDockerHost()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resolving docker context host")
	})
}
func TestNewDockerClient(t *testing.T) {
	oldDockerHost := os.Getenv("DOCKER_HOST")
	defer func() {
		if oldDockerHost == "" {
			os.Unsetenv("DOCKER_HOST")
		} else {
			os.Setenv("DOCKER_HOST", oldDockerHost)
		}
	}()

	os.Unsetenv("DOCKER_HOST")

	dockerClient, err := NewDockerClient()
	require.NoError(t, err)
	require.NotNil(t, dockerClient)

	defer dockerClient.cli.Close()

	ctx := context.Background()
	_, err = dockerClient.cli.Ping(ctx)
	require.NoError(t, err)
}
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			os.Stdout.WriteString(os.Args[i+1])
			os.Exit(0)
		}
	}

	os.Exit(1)
}
func TestHelperProcessFailure(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	os.Exit(2)
}
