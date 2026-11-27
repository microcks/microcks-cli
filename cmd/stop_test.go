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

package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStopContainerClient struct {
	exists      bool
	existsErr   error
	stopErr     error
	closeErr    error
	stoppedIDs  []string
	closedCalls int
	stopOutputs []io.Writer
}

func (m *mockStopContainerClient) CreateContainer(opts connectors.ContainerOpts) (string, error) {
	return "", nil
}

func (m *mockStopContainerClient) StartContainer(containerID string) error {
	return nil
}

func (m *mockStopContainerClient) StopContainer(containerID string) error {
	return m.StopContainerWithOutput(containerID, os.Stdout)
}

func (m *mockStopContainerClient) StopContainerWithOutput(containerID string, out io.Writer) error {
	m.stopOutputs = append(m.stopOutputs, out)
	if m.stopErr != nil {
		return m.stopErr
	}
	m.stoppedIDs = append(m.stoppedIDs, containerID)
	return nil
}

func (m *mockStopContainerClient) ContainerExists(containerID string) (bool, error) {
	return m.exists, m.existsErr
}

func (m *mockStopContainerClient) CloseClient() error {
	m.closedCalls++
	return m.closeErr
}

func captureStdoutForStopTest(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
	}()

	runErr := fn()
	require.NoError(t, w.Close())

	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	require.NoError(t, copyErr)
	require.NoError(t, r.Close())
	return buf.String(), runErr
}

func writeStopTestConfig(t *testing.T, autoRemove bool) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config")
	cfg := config.LocalConfig{
		CurrentContext: "http://localhost:8585",
		Contexts: []config.ContextRef{
			{
				Name:     "http://localhost:8585",
				Server:   "http://localhost:8585",
				User:     "http://localhost:8585",
				Instance: "microcks",
			},
		},
		Servers: []config.Server{
			{
				Name:   "microcks",
				Server: "http://localhost:8585",
			},
		},
		Users: []config.User{
			{
				Name: "http://localhost:8585",
			},
		},
		Auths: []config.Auth{
			{
				Server: "http://localhost:8585",
			},
		},
		Instances: []config.Instance{
			{
				ContainerID: "container-123",
				Name:        "microcks",
				Port:        "8585",
				Status:      "Running",
				Driver:      "docker",
				AutoRemove:  autoRemove,
			},
		},
	}
	require.NoError(t, config.WriteLocalConfig(cfg, configPath))
	return configPath
}

func TestStopCommandRejectsUnsupportedOutput(t *testing.T) {
	_, err := executeCLIForTest(t, "stop", "--output", "yaml")
	require.Error(t, err)
	assert.Equal(t, errors.KindUsage, errors.KindOf(err))
}

func TestStopOutputsJSONWithoutConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing-config")

	out, err := executeCLIForTest(t, "stop", "--output", "json", "--config", configPath)
	require.NoError(t, err)

	var result instanceStopResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, instanceStopResult{Status: "not_found"}, result)
}

func TestStopOutputsJSONWhenContextHasNoInstance(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	cfg := config.LocalConfig{
		CurrentContext: "http://localhost:8080",
		Contexts: []config.ContextRef{
			{Name: "http://localhost:8080", Server: "http://localhost:8080", User: "http://localhost:8080"},
		},
		Servers: []config.Server{{Server: "http://localhost:8080"}},
		Users:   []config.User{{Name: "http://localhost:8080"}},
	}
	require.NoError(t, config.WriteLocalConfig(cfg, configPath))

	out, err := executeCLIForTest(t, "stop", "--output", "json", "--config", configPath)
	require.NoError(t, err)

	var result instanceStopResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, instanceStopResult{
		Context: "http://localhost:8080",
		Status:  "not_found",
	}, result)
}

func TestRunStopOutputsJSONStopped(t *testing.T) {
	configPath := writeStopTestConfig(t, false)
	mockClient := &mockStopContainerClient{exists: true}

	out, err := captureStdoutForStopTest(t, func() error {
		return runStop(configPath, "", "json", func(driver string) (connectors.ContainerClient, error) {
			assert.Equal(t, "docker", driver)
			return mockClient, nil
		})
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"container-123"}, mockClient.stoppedIDs)
	assert.Equal(t, 1, mockClient.closedCalls)
	assert.Equal(t, []io.Writer{os.Stderr}, mockClient.stopOutputs)

	var result instanceStopResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, instanceStopResult{
		Name:    "microcks",
		Context: "http://localhost:8585",
		Status:  "stopped",
		Removed: false,
	}, result)

	updatedCfg, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	instance, err := updatedCfg.GetInstance("microcks")
	require.NoError(t, err)
	assert.Equal(t, "Exited", instance.Status)
}

func TestRunStopOutputsJSONAutoRemoved(t *testing.T) {
	configPath := writeStopTestConfig(t, true)
	mockClient := &mockStopContainerClient{exists: true}

	out, err := captureStdoutForStopTest(t, func() error {
		return runStop(configPath, "", "json", func(string) (connectors.ContainerClient, error) {
			return mockClient, nil
		})
	})
	require.NoError(t, err)

	var result instanceStopResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, instanceStopResult{
		Name:    "microcks",
		Context: "http://localhost:8585",
		Status:  "removed",
		Removed: true,
	}, result)

	updatedCfg, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	assert.Empty(t, updatedCfg.CurrentContext)
	assert.Empty(t, updatedCfg.Contexts)
	assert.Empty(t, updatedCfg.Instances)
}

func TestRunStopOutputsJSONMissingContainer(t *testing.T) {
	configPath := writeStopTestConfig(t, false)
	mockClient := &mockStopContainerClient{exists: false}

	out, err := captureStdoutForStopTest(t, func() error {
		return runStop(configPath, "", "json", func(string) (connectors.ContainerClient, error) {
			return mockClient, nil
		})
	})
	require.NoError(t, err)
	assert.Empty(t, mockClient.stoppedIDs)

	var result instanceStopResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, instanceStopResult{
		Name:    "microcks",
		Context: "http://localhost:8585",
		Status:  "missing",
	}, result)
}

func TestRunStopTextMode(t *testing.T) {
	configPath := writeStopTestConfig(t, false)
	mockClient := &mockStopContainerClient{exists: true}

	out, err := captureStdoutForStopTest(t, func() error {
		return runStop(configPath, "", "text", func(string) (connectors.ContainerClient, error) {
			return mockClient, nil
		})
	})
	require.NoError(t, err)
	assert.Contains(t, out, "Instance microcks stopped successfully")
	assert.Contains(t, out, "Instance microcks status updated to Exited")
}
