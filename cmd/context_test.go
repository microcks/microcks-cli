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
