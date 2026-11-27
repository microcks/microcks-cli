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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	clierrors "github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextListOutputsJSON(t *testing.T) {
	configPath := t.TempDir() + "/config.yaml"
	require.NoError(t, os.WriteFile(configPath, []byte(testConfig), 0o600))

	out, err := executeCLIForTest(t, "context", "--output", "json", "--config", configPath)
	require.NoError(t, err)

	var contexts []contextSummary
	require.NoError(t, json.Unmarshal([]byte(out), &contexts))
	require.Len(t, contexts, 2)
	assert.Equal(t, "http://localhost:8083", contexts[1].Name)
	assert.True(t, contexts[1].Current)
}

func TestContextListOutputsEmptyJSONArrayWithoutConfig(t *testing.T) {
	configPath := t.TempDir() + "/missing-config.yaml"

	out, err := executeCLIForTest(t, "context", "--output", "json", "--config", configPath)
	require.NoError(t, err)
	assert.JSONEq(t, "[]", out)
}

func TestContextUseOutputsJSON(t *testing.T) {
	configPath := t.TempDir() + "/config.yaml"
	require.NoError(t, os.WriteFile(configPath, []byte(testConfig), 0o600))

	out, err := executeCLIForTest(
		t,
		"context",
		"http://localhost:8080",
		"--output",
		"json",
		"--config",
		configPath,
	)
	require.NoError(t, err)

	var result contextMutationResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "selected", result.Action)
	assert.Equal(t, "http://localhost:8080", result.Server)
}

const testConfig = `current-context: http://localhost:8083
contexts:
- name: http://localhost:8080
  server: http://localhost:8080
  user: http://localhost:8080
  instance: ""
- name: http://localhost:8083
  server: http://localhost:8083
  user: http://localhost:8083
  instance: ""
servers:
- name: ""
  server: http://localhost:8080
  insecureTLS: true
  keycloakEnable: true
- name: ""
  server: http://localhost:8083
  insecureTLS: true
  keycloakEnable: true
users:
- name: http://localhost:8080
  auth-token: vErrYS3c3tReFRe$hToken
  refresh-token: vErrYS3c3tReFRe$hToken
- name: http://localhost:8083
  auth-token: ""
  refresh-token: ""`

const localInstanceConfig = `current-context: http://localhost:8585
contexts:
- name: http://localhost:8585
  server: http://localhost:8585
  user: http://localhost:8585
  instance: demo
servers:
- server: http://localhost:8585
  insecureTLS: true
  keycloakEnable: false
users:
- name: http://localhost:8585
  auth-token: ""
  refresh-token: ""
instances:
- name: demo
  image: quay.io/microcks/microcks-uber:latest-native
  status: Running
  port: "8585"
  containerID: abc123
  autoRemove: false
  driver: docker
auths:
- server: http://localhost:8585
  clientid: ""
  clientsecret: ""
`

func writeLocalConfigForTest(t *testing.T, contents string) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(contents), 0o600))
	return configPath
}

const testConfigFilePath = "./testdata/local.config"

func TestDeleteContext(t *testing.T) {
	//write the test config file
	require.NoError(t, os.MkdirAll("./testdata", 0o750))
	err := os.WriteFile(testConfigFilePath, []byte(testConfig), os.ModePerm)
	require.NoError(t, err)

	err = os.Chmod(testConfigFilePath, 0o600)
	require.NoError(t, err, "Could not change the file permission to 0600 %v", err)
	localCfg, err := config.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8083", localCfg.CurrentContext)
	assert.Contains(t, localCfg.Contexts, config.ContextRef{Name: "http://localhost:8083", Server: "http://localhost:8083", User: "http://localhost:8083", Instance: ""})

	//Delete non-existing context
	err = deleteContext("microcks.io", testConfigFilePath)
	require.EqualError(t, err, `context "microcks.io" does not exist`)

	//Delete non-current context
	err = deleteContext("http://localhost:8080", testConfigFilePath)
	require.NoError(t, err)

	//Delete current context
	err = deleteContext("http://localhost:8083", testConfigFilePath)
	require.NoError(t, err)
	_, err = config.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
}

func TestDeleteContextEmpty(t *testing.T) {
	err := deleteContext("http://localhost:8080", "./testdata/non-existent-file.config")
	require.EqualError(t, err, "nothing to delete")
}

func TestDeleteContextRejectsRunningLocalInstance(t *testing.T) {
	configPath := writeLocalConfigForTest(t, localInstanceConfig)

	err := deleteContext("http://localhost:8585", configPath)
	require.EqualError(t, err, `context "http://localhost:8585" is backed by running local instance "demo"; run 'microcks stop' before deleting it`)
	assert.Equal(t, clierrors.KindUsage, clierrors.KindOf(err))

	_, err = os.Stat(configPath)
	require.NoError(t, err, "config must be unchanged when deletion is rejected")

	localCfg, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, localCfg)
	require.Len(t, localCfg.Contexts, 1)
	require.Len(t, localCfg.Instances, 1)
	assert.Equal(t, "abc123", localCfg.Instances[0].ContainerID)
}

func TestDeleteContextAllowsStoppedLocalInstance(t *testing.T) {
	stoppedConfig := strings.Replace(localInstanceConfig, "status: Running", "status: Exited", 1)
	configPath := writeLocalConfigForTest(t, stoppedConfig)

	err := deleteContext("http://localhost:8585", configPath)
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	require.NoError(t, err, "config must be retained while a local instance still exists")

	localCfg, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, localCfg)
	assert.Empty(t, localCfg.Contexts)
	assert.Empty(t, localCfg.Servers)
	assert.Empty(t, localCfg.Users)
	assert.Empty(t, localCfg.Auths)
	require.Len(t, localCfg.Instances, 1)
	assert.Equal(t, "abc123", localCfg.Instances[0].ContainerID)
}
