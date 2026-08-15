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
