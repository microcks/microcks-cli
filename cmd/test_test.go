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
	"os/exec"
	"testing"

	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const missingTestArgsMessage = "accepts 3 arg(s), received 2"

func TestTestCommandMissingRunnerWithGlobalFlagsDoesNotPanic(t *testing.T) {
	if os.Getenv("MICROCKS_CLI_TEST_MISSING_RUNNER") == "1" {
		os.Args = []string{
			os.Args[0],
			"test",
			"--microcksURL=http://localhost:8080",
			"--keycloakClientId=foo",
			"--keycloakClientSecret=bar",
			"MyAPI:1.0",
			"http://localhost:3000",
		}

		command, err := NewCommand()
		if err != nil {
			t.Fatal(err)
		}
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestTestCommandMissingRunnerWithGlobalFlagsDoesNotPanic")
	cmd.Env = append(os.Environ(), "MICROCKS_CLI_TEST_MISSING_RUNNER=1")

	output, err := cmd.CombinedOutput()
	require.Error(t, err)

	exitError, ok := err.(*exec.ExitError)
	require.True(t, ok)
	assert.Equal(t, 1, exitError.ExitCode())
	assert.Contains(t, string(output), missingTestArgsMessage)
	assert.NotContains(t, string(output), "panic:") // core regression guard: must never crash with a stack trace
}

func TestValidateTestEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		wantError bool
	}{
		{name: "http URL", endpoint: "http://example.com/api"},
		{name: "https URL", endpoint: "https://example.com/api"},
		{name: "uppercase HTTP scheme", endpoint: "HTTP://example.com/api"},
		{name: "file URL", endpoint: "file:///etc/passwd", wantError: true},
		{name: "ftp URL", endpoint: "ftp://example.com/api", wantError: true},
		{name: "scheme-less host and port", endpoint: "localhost:8080", wantError: true},
		{name: "malformed URL", endpoint: "http://[::1", wantError: true},
		{name: "missing host", endpoint: "http://", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTestEndpoint(tt.endpoint)
			if tt.wantError {
				assert.Equal(t, errors.KindUsage, errors.KindOf(err))
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
