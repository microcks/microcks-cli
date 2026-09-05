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
	"io"
	"os"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/stretchr/testify/require"
)

const validateTestConfigFilePath = "./testdata/validate.config"

// setupValidateTestConfig writes the shared testConfig (defined in
// context_test.go) to disk so ReadLocalConfig can load it, and returns
// a cleanup func to remove it afterwards.
func setupValidateTestConfig(t *testing.T) func() {
	t.Helper()
	err := os.MkdirAll("./testdata", 0755)
	require.NoError(t, err)
	err = os.WriteFile(validateTestConfigFilePath, []byte(testConfig), 0644)
	require.NoError(t, err)
	return func() {
		_ = os.Remove(validateTestConfigFilePath)
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns
// everything that was printed, so we can assert on runValidate's output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

func TestRunValidate_NoConfigFile_ReturnsFailureAndMessage(t *testing.T) {
	opts := &connectors.ClientOptions{
		ConfigPath: "./testdata/does-not-exist.config",
	}

	var exitCode int
	output := captureStdout(t, func() {
		exitCode = runValidate(opts)
	})

	require.Equal(t, 1, exitCode)
	require.Contains(t, output, "No contexts configured")
}

func TestRunValidate_InvalidConfigFile_ReturnsFailure(t *testing.T) {
	// current-context points to a context name that doesn't exist in the
	// contexts list — this is what ValidateLocalConfig() actually rejects.
	badConfig := `current-context: ghost-context
contexts:
- name: http://localhost:8080
  server: http://localhost:8080
  user: http://localhost:8080
  instance: ""
servers:
- name: ""
  server: http://localhost:8080
  insecureTLS: true
  keycloakEnable: true
users:
- name: http://localhost:8080
  auth-token: ""
  refresh-token: ""`

	badConfigPath := "./testdata/invalid.config"
	err := os.MkdirAll("./testdata", 0755)
	require.NoError(t, err)
	err = os.WriteFile(badConfigPath, []byte(badConfig), 0644)
	require.NoError(t, err)
	defer os.Remove(badConfigPath)

	opts := &connectors.ClientOptions{
		ConfigPath: badConfigPath,
	}

	var exitCode int
	output := captureStdout(t, func() {
		exitCode = runValidate(opts)
	})

	require.Equal(t, 1, exitCode)
	require.Contains(t, output, "Config file: invalid")
}

func TestRunValidate_ValidConfigButUnreachableServer_ReturnsFailure(t *testing.T) {
	cleanup := setupValidateTestConfig(t)
	defer cleanup()

	opts := &connectors.ClientOptions{
		ConfigPath: validateTestConfigFilePath,
	}

	var exitCode int
	output := captureStdout(t, func() {
		exitCode = runValidate(opts)
	})

	// The current context (localhost:8083) is not actually running, so the
	// reachability (or auth) check must fail and the command must report it.
	require.Equal(t, 1, exitCode)
	require.Contains(t, output, "Context resolved")
	require.Contains(t, output, "One or more checks failed")
}

func TestRunValidate_DirectModeIncomplete_FallsBackToLocalConfigPath(t *testing.T) {
	// Only --microcksURL set, missing client id/secret: direct mode must
	// NOT engage, and behavior should fall back to local config handling.
	opts := &connectors.ClientOptions{
		ConfigPath: "./testdata/does-not-exist.config",
		ServerAddr: "http://localhost:9999",
	}

	var exitCode int
	output := captureStdout(t, func() {
		exitCode = runValidate(opts)
	})

	require.Equal(t, 1, exitCode)
	require.Contains(t, output, "No contexts configured")
}
