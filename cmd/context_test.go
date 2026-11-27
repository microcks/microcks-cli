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
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
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

const orphanedContextConfig = `current-context: prod
contexts:
- name: prod
  server: https://microcks-prod.example.com
  user: prod-user
  instance: ""
- name: broken-ctx
  server: https://missing-server.example.com
  user: missing-user
  instance: ""
servers:
- name: ""
  server: https://microcks-prod.example.com
  insecureTLS: true
  keycloakEnable: false
users:
- name: prod-user
  auth-token: valid-token
  refresh-token: valid-refresh`

func TestContextListWithOrphanedContext(t *testing.T) {
	configPath := t.TempDir() + "/config.yaml"
	require.NoError(t, os.WriteFile(configPath, []byte(orphanedContextConfig), 0o600))

	// Test JSON output
	out, err := executeCLIForTest(t, "context", "--output", "json", "--config", configPath)
	require.NoError(t, err)

	var contexts []contextSummary
	require.NoError(t, json.Unmarshal([]byte(out), &contexts))
	require.Len(t, contexts, 2)
	assert.Equal(t, "prod", contexts[0].Name)
	assert.Equal(t, "https://microcks-prod.example.com", contexts[0].Server)
	assert.True(t, contexts[0].Current)

	assert.Equal(t, "broken-ctx", contexts[1].Name)
	assert.Equal(t, "https://missing-server.example.com", contexts[1].Server)
	assert.False(t, contexts[1].Current)

	// Test Text output
	textOut, err := executeCLIForTest(t, "context", "--output", "text", "--config", configPath)
	require.NoError(t, err)
	assert.Contains(t, textOut, "prod")
	assert.Contains(t, textOut, "broken-ctx")
	assert.Contains(t, textOut, "https://missing-server.example.com")
}

func TestDeleteOrphanedContext(t *testing.T) {
	configPath := t.TempDir() + "/config.yaml"
	require.NoError(t, os.WriteFile(configPath, []byte(orphanedContextConfig), 0o600))

	// Deleting the orphaned context should succeed
	err := deleteContext("broken-ctx", configPath)
	require.NoError(t, err)

	// Verify the context was removed and valid context remains
	reloaded, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	require.Len(t, reloaded.Contexts, 1)
	assert.Equal(t, "prod", reloaded.Contexts[0].Name)
	assert.Equal(t, "prod", reloaded.CurrentContext)

	// Verify server and user for remaining context still exist
	assert.Len(t, reloaded.Servers, 1)
	assert.Equal(t, "https://microcks-prod.example.com", reloaded.Servers[0].Server)
	assert.Len(t, reloaded.Users, 1)
	assert.Equal(t, "prod-user", reloaded.Users[0].Name)
}

func TestDeleteOrphanedContextViaCLI(t *testing.T) {
	configPath := t.TempDir() + "/config.yaml"
	require.NoError(t, os.WriteFile(configPath, []byte(orphanedContextConfig), 0o600))

	out, err := executeCLIForTest(t, "context", "broken-ctx", "--delete", "--config", configPath)
	require.NoError(t, err)
	assert.Contains(t, out, "Context 'broken-ctx' deleted")

	reloaded, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.Len(t, reloaded.Contexts, 1)
}

func TestDeleteOrphanedContextWhenOnlyContext(t *testing.T) {
	onlyBrokenConfig := `current-context: ""
contexts:
- name: only-broken
  server: https://missing-server.example.com
  user: missing-user
  instance: ""
servers: []
users: []`

	configPath := t.TempDir() + "/config.yaml"
	require.NoError(t, os.WriteFile(configPath, []byte(onlyBrokenConfig), 0o600))

	err := deleteContext("only-broken", configPath)
	require.NoError(t, err)

	// Config should be deleted as it becomes empty
	_, err = os.Stat(configPath)
	assert.True(t, os.IsNotExist(err))
}
