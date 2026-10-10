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

package connectors

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelperProcess intercepts execCommand calls during unit tests.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "no command provided")
		os.Exit(2)
	}

	cmd, cmdArgs := args[0], args[1:]
	switch cmd {
	case "docker":
		if os.Getenv("HELPER_DOCKER_FAIL") == "1" {
			os.Exit(1)
		}
		if len(cmdArgs) >= 4 && cmdArgs[0] == "context" && cmdArgs[1] == "inspect" {
			fmt.Print("unix:///mock/active/docker.sock\n")
		}
	default:
		os.Exit(2)
	}
}

func mockExecCommand(envVars ...string) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		cmd.Env = append(cmd.Env, envVars...)
		return cmd
	}
}

func TestConfigureDockerHost_PreservesExistingDockerHost(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix:///existing/custom.sock")

	err := ConfigureDockerHost()
	require.NoError(t, err)
	assert.Equal(t, "unix:///existing/custom.sock", os.Getenv("DOCKER_HOST"))
}

func TestConfigureDockerHost_NonDarwin(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	origGOOS := currentGOOS
	currentGOOS = "linux"
	defer func() { currentGOOS = origGOOS }()

	err := ConfigureDockerHost()
	require.NoError(t, err)
	assert.Empty(t, os.Getenv("DOCKER_HOST"))
}

func TestConfigureDockerHost_DefaultSocketExists(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	origGOOS := currentGOOS
	origDefaultPath := defaultSocketPath
	currentGOOS = "darwin"
	tmpDir := t.TempDir()
	dummySock := filepath.Join(tmpDir, "docker.sock")
	require.NoError(t, os.WriteFile(dummySock, []byte(""), 0o600))
	defaultSocketPath = dummySock
	defer func() {
		currentGOOS = origGOOS
		defaultSocketPath = origDefaultPath
	}()

	err := ConfigureDockerHost()
	require.NoError(t, err)
	assert.Empty(t, os.Getenv("DOCKER_HOST"))
}

func TestConfigureDockerHost_InspectContextSuccess(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	origGOOS := currentGOOS
	origDefaultPath := defaultSocketPath
	origExec := execCommand
	currentGOOS = "darwin"
	defaultSocketPath = filepath.Join(t.TempDir(), "nonexistent.sock")
	execCommand = mockExecCommand()
	defer func() {
		currentGOOS = origGOOS
		defaultSocketPath = origDefaultPath
		execCommand = origExec
	}()

	err := ConfigureDockerHost()
	require.NoError(t, err)
	assert.Equal(t, "unix:///mock/active/docker.sock", os.Getenv("DOCKER_HOST"))
}

func TestConfigureDockerHost_FallbackToUserSocket(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	origGOOS := currentGOOS
	origDefaultPath := defaultSocketPath
	origExec := execCommand
	currentGOOS = "darwin"
	defaultSocketPath = filepath.Join(t.TempDir(), "nonexistent.sock")
	execCommand = mockExecCommand("HELPER_DOCKER_FAIL=1")

	tempHome := t.TempDir()
	userSockDir := filepath.Join(tempHome, ".docker", "run")
	require.NoError(t, os.MkdirAll(userSockDir, 0o755))
	userSockFile := filepath.Join(userSockDir, "docker.sock")
	require.NoError(t, os.WriteFile(userSockFile, []byte(""), 0o600))
	t.Setenv("HOME", tempHome)

	defer func() {
		currentGOOS = origGOOS
		defaultSocketPath = origDefaultPath
		execCommand = origExec
	}()

	err := ConfigureDockerHost()
	require.NoError(t, err)
	assert.Equal(t, "unix://"+userSockFile, os.Getenv("DOCKER_HOST"))
}

func TestConfigureDockerHost_NeitherContextNorUserSocketAvailable(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	origGOOS := currentGOOS
	origDefaultPath := defaultSocketPath
	origExec := execCommand
	currentGOOS = "darwin"
	defaultSocketPath = filepath.Join(t.TempDir(), "nonexistent.sock")
	execCommand = mockExecCommand("HELPER_DOCKER_FAIL=1")
	t.Setenv("HOME", t.TempDir())

	defer func() {
		currentGOOS = origGOOS
		defaultSocketPath = origDefaultPath
		execCommand = origExec
	}()

	err := ConfigureDockerHost()
	require.NoError(t, err)
	assert.Empty(t, os.Getenv("DOCKER_HOST"))
}
