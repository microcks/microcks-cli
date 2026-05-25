package cmd

import (
	"os"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	require.EqualError(t, err, "Context microcks.io does not exist")

	//Delete non-current context
	err = deleteContext("http://localhost:8080", testConfigFilePath)
	require.NoError(t, err)

	//Delete current context
	err = deleteContext("http://localhost:8083", testConfigFilePath)
	require.NoError(t, err)
	_, err = config.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
}

func TestDeleteNamedContextRemovesReferencedUserAndAuth(t *testing.T) {
	configPath := t.TempDir() + "/config"
	localConfig := `current-context: staging
contexts:
- name: dev
  server: http://localhost:8080
  user: http://localhost:8080
  instance: ""
- name: staging
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
  auth-token: stale-token
  refresh-token: stale-refresh-token
- name: http://localhost:8083
  auth-token: ""
  refresh-token: ""
auths:
- server: http://localhost:8080
  clientid: my-client
  clientsecret: my-secret
`

	err := os.WriteFile(configPath, []byte(localConfig), os.ModePerm)
	require.NoError(t, err)
	err = os.Chmod(configPath, 0o600)
	require.NoError(t, err)

	err = deleteContext("dev", configPath)
	require.NoError(t, err)

	localCfg, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, localCfg)

	assert.Equal(t, "staging", localCfg.CurrentContext)
	assert.NotContains(t, localCfg.Contexts, config.ContextRef{Name: "dev", Server: "http://localhost:8080", User: "http://localhost:8080", Instance: ""})
	assert.NotContains(t, localCfg.Servers, config.Server{Server: "http://localhost:8080", InsecureTLS: true, KeycloakEnable: true})
	assert.NotContains(t, localCfg.Users, config.User{Name: "http://localhost:8080", AuthToken: "stale-token", RefreshToken: "stale-refresh-token"})
	assert.NotContains(t, localCfg.Auths, config.Auth{Server: "http://localhost:8080", ClientId: "my-client", ClientSecret: "my-secret"})
	assert.Contains(t, localCfg.Contexts, config.ContextRef{Name: "staging", Server: "http://localhost:8083", User: "http://localhost:8083", Instance: ""})
}
