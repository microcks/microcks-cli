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
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTestCommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		// valid
		{
			name:    "valid HTTP runner",
			args:    []string{"API:1.0", "http://localhost:8080", "HTTP"},
			wantErr: false,
		},
		{
			name:    "valid POSTMAN runner",
			args:    []string{"Petstore:1.0", "http://localhost:9090", "POSTMAN"},
			wantErr: false,
		},
		{
			name:    "valid OPEN_API_SCHEMA runner",
			args:    []string{"API:2.0", "http://endpoint", "OPEN_API_SCHEMA"},
			wantErr: false,
		},

		// missing args
		{
			name:        "no args",
			args:        []string{},
			wantErr:     true,
			errContains: testCommandUsageError,
		},
		{
			name:        "missing runner",
			args:        []string{"API:1.0", "http://endpoint"},
			wantErr:     true,
			errContains: testCommandUsageError,
		},

		// args look like flags
		{
			name:        "serviceRef starts with dash",
			args:        []string{"-flag", "http://endpoint", "HTTP"},
			wantErr:     true,
			errContains: testCommandUsageError,
		},
		{
			name:        "testEndpoint starts with dash",
			args:        []string{"API:1.0", "--endpoint", "HTTP"},
			wantErr:     true,
			errContains: testCommandUsageError,
		},

		// invalid runner
		{
			name:        "invalid runner INVALID",
			args:        []string{"API:1.0", "http://endpoint", "INVALID"},
			wantErr:     true,
			errContains: testCommandRunnerError,
		},
		{
			name:        "runner lowercase http",
			args:        []string{"API:1.0", "http://endpoint", "http"},
			wantErr:     true,
			errContains: testCommandRunnerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := validateTestCommandArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTestCommandArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if err.Error() != tt.errContains {
					t.Errorf("error message = %q, want %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestParseWaitForMilliseconds(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		// valid
		{"500milli", 500, false},
		{"1sec", 1000, false},
		{"2min", 120000, false},
		{"0milli", 0, false},
		{"30sec", 30000, false},

		// wrong suffix
		{"5seconds", 0, true},
		{"5mins", 0, true},
		{"5ms", 0, true},
		{"", 0, true},

		// bad numeric prefix
		{"abcsec", 0, true},
		{"sec", 0, true},    // missing number
		{"milli", 0, true},  // missing number
		{"1.5sec", 0, true}, // float not accepted
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseWaitForMilliseconds(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseWaitForMilliseconds(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseWaitForMilliseconds(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewTestCommand(t *testing.T) {
	clientOpts := &connectors.ClientOptions{}
	cmd := NewTestCommand(clientOpts)

	assert.Equal(t, "test <apiName:apiVersion> <testEndpoint> <runner>", cmd.Use)
	assert.Equal(t, "Run tests on Microcks", cmd.Short)

	waitForFlag := cmd.Flags().Lookup("waitFor")
	require.NotNil(t, waitForFlag)
	assert.Equal(t, "5sec", waitForFlag.DefValue)

	secretFlag := cmd.Flags().Lookup("secretName")
	require.NotNil(t, secretFlag)

	filteredOperationsFlag := cmd.Flags().Lookup("filteredOperations")
	require.NotNil(t, filteredOperationsFlag)

	operationsHeadersFlag := cmd.Flags().Lookup("operationsHeaders")
	require.NotNil(t, operationsHeadersFlag)

	oauth2ContextFlag := cmd.Flags().Lookup("oAuth2Context")
	require.NotNil(t, oauth2ContextFlag)
}
