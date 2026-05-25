package cmd

import (
	"os"
	"path/filepath"
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

const sharedServerConfig = `current-context: dev
contexts:
- name: dev
  server: http://localhost:8585
  user: http://localhost:8585
  instance: microcks
- name: qa
  server: http://localhost:8585
  user: http://localhost:8585
  instance: microcks
servers:
- name: microcks
  server: http://localhost:8585
  insecureTLS: true
  keycloakEnable: false
users:
- name: http://localhost:8585
  auth-token: ""
  refresh-token: ""
instances:
- name: microcks
  image: quay.io/microcks/microcks-uber:latest-native
  status: Running
  port: "8585"
  containerID: abc123
  autoRemove: false
  driver: docker
auths:
- server: http://localhost:8585
  clientid: ""
  clientsecret: ""`

func TestDeleteContext(t *testing.T) {
	testConfigFilePath := filepath.Join(t.TempDir(), "local.config")

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

func TestDeleteContextSharedServerKeepsReferencedEntries(t *testing.T) {
	testConfigFilePath := filepath.Join(t.TempDir(), "local.config")

	err := os.WriteFile(testConfigFilePath, []byte(sharedServerConfig), os.ModePerm)
	require.NoError(t, err)

	err = os.Chmod(testConfigFilePath, 0o600)
	require.NoError(t, err, "Could not change the file permission to 0600 %v", err)

	err = deleteContext("dev", testConfigFilePath)
	require.NoError(t, err)

	localCfg, err := config.ReadLocalConfig(testConfigFilePath)
	require.NoError(t, err)
	require.NotNil(t, localCfg)

	assert.Equal(t, "", localCfg.CurrentContext)
	assert.Len(t, localCfg.Contexts, 1)
	assert.Equal(t, "qa", localCfg.Contexts[0].Name)
	assert.Len(t, localCfg.Servers, 1)
	assert.Equal(t, "http://localhost:8585", localCfg.Servers[0].Server)
	assert.Len(t, localCfg.Users, 1)
	assert.Equal(t, "http://localhost:8585", localCfg.Users[0].Name)
	assert.Len(t, localCfg.Instances, 1)
	assert.Equal(t, "microcks", localCfg.Instances[0].Name)
	assert.Len(t, localCfg.Auths, 1)
	assert.Equal(t, "http://localhost:8585", localCfg.Auths[0].Server)
}
