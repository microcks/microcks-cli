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
	"io"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestParseImportURLArg(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		expectedURL          string
		expectedMainArtifact bool
		expectedSecret       string
	}{
		{
			name:                 "standard URL without suffixes",
			input:                "https://example.com/openapi.yaml",
			expectedURL:          "https://example.com/openapi.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "standard URL with mainArtifact true",
			input:                "https://example.com/spec1.yaml:true",
			expectedURL:          "https://example.com/spec1.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "standard URL with mainArtifact false",
			input:                "https://example.com/spec1.yaml:false",
			expectedURL:          "https://example.com/spec1.yaml",
			expectedMainArtifact: false,
			expectedSecret:       "",
		},
		{
			name:                 "standard URL with mainArtifact and secret",
			input:                "https://example.com/spec1.yaml:true:my-secret",
			expectedURL:          "https://example.com/spec1.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "my-secret",
		},
		{
			name:                 "standard URL with mainArtifact false and secret",
			input:                "https://example.com/spec1.yaml:false:my-secret-token",
			expectedURL:          "https://example.com/spec1.yaml",
			expectedMainArtifact: false,
			expectedSecret:       "my-secret-token",
		},
		{
			name:                 "URL with port and no suffixes",
			input:                "http://localhost:8585/spec.yaml",
			expectedURL:          "http://localhost:8585/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "URL with port and mainArtifact true",
			input:                "http://localhost:8585/spec.yaml:true",
			expectedURL:          "http://localhost:8585/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "URL with port and mainArtifact false",
			input:                "http://localhost:8585/spec.yaml:false",
			expectedURL:          "http://localhost:8585/spec.yaml",
			expectedMainArtifact: false,
			expectedSecret:       "",
		},
		{
			name:                 "URL with port, mainArtifact and secret",
			input:                "http://localhost:8585/spec.yaml:true:my-secret-token",
			expectedURL:          "http://localhost:8585/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "my-secret-token",
		},
		{
			name:                 "URL with port, mainArtifact false and secret",
			input:                "http://localhost:8585/spec.yaml:false:my-secret-token",
			expectedURL:          "http://localhost:8585/spec.yaml",
			expectedMainArtifact: false,
			expectedSecret:       "my-secret-token",
		},
		{
			name:                 "malformed bool suffix preserves URL",
			input:                "http://localhost:8585/spec.yaml:tru",
			expectedURL:          "http://localhost:8585/spec.yaml:tru",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "malformed bool with secret preserves URL",
			input:                "http://localhost:8585/spec.yaml:tru:mysecret",
			expectedURL:          "http://localhost:8585/spec.yaml:tru:mysecret",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "URL with path and no port",
			input:                "https://raw.githubusercontent.com/org/repo/main/spec.yaml",
			expectedURL:          "https://raw.githubusercontent.com/org/repo/main/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "short URL without path",
			input:                "http://localhost:true",
			expectedURL:          "http://localhost",
			expectedMainArtifact: true,
			expectedSecret:       "",
		},
		{
			name:                 "boolean-like secret is not consumed as primary flag",
			input:                "http://example.com/api:true:false",
			expectedURL:          "http://example.com/api",
			expectedMainArtifact: true,
			expectedSecret:       "false",
		},
		{
			name:                 "secret containing colons",
			input:                "http://example.com/api:true:my:secret:token",
			expectedURL:          "http://example.com/api",
			expectedMainArtifact: true,
			expectedSecret:       "my:secret:token",
		},
		{
			name:                 "URL with port and secret containing colons",
			input:                "http://localhost:8080/api:false:auth:basic:user:pass",
			expectedURL:          "http://localhost:8080/api",
			expectedMainArtifact: false,
			expectedSecret:       "auth:basic:user:pass",
		},
		{
			name:                 "boolean host with port does not get misparsed as primary flag",
			input:                "http://true:8080/api:false:secret",
			expectedURL:          "http://true:8080/api",
			expectedMainArtifact: false,
			expectedSecret:       "secret",
		},
		{
			name:                 "uppercase HTTP scheme with suffix",
			input:                "HTTP://example.com/spec.yaml:false:mysecret",
			expectedURL:          "HTTP://example.com/spec.yaml",
			expectedMainArtifact: false,
			expectedSecret:       "mysecret",
		},
		{
			name:                 "mixed-case Https scheme with suffix",
			input:                "Https://example.com/spec.yaml:true:mysecret",
			expectedURL:          "Https://example.com/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:       "mysecret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, mainArtifact, secret := parseImportURLArg(tt.input)
			assert.Equal(t, tt.expectedURL, url)
			assert.Equal(t, tt.expectedMainArtifact, mainArtifact)
			assert.Equal(t, tt.expectedSecret, secret)
		})
	}
}

func TestValidateImportURLArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:           "missing args",
			args:           []string{},
			wantErr:        true,
			expectedErrMsg: "import-url requires a <specificationFileURL[:primary[:secret]]> argument",
		},
		{
			name:    "valid standard URL",
			args:    []string{"https://example.com/spec.yaml"},
			wantErr: false,
		},
		{
			name:    "valid uppercase HTTP scheme",
			args:    []string{"HTTP://example.com/spec.yaml"},
			wantErr: false,
		},
		{
			name:    "valid mixed-case Https scheme",
			args:    []string{"Https://example.com/spec.yaml:false:secret"},
			wantErr: false,
		},
		{
			name:           "relative path without scheme",
			args:           []string{"spec.yaml"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'spec.yaml': must start with http:// or https://",
		},
		{
			name:           "relative path with suffixes without scheme",
			args:           []string{"spec.yaml:true:mysecret"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'spec.yaml:true:mysecret': must start with http:// or https://",
		},
		{
			name:           "unsupported ftp scheme",
			args:           []string{"ftp://example.com/spec.yaml"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'ftp://example.com/spec.yaml': must start with http:// or https://",
		},
		{
			name:           "file scheme",
			args:           []string{"file:///tmp/spec.yaml"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'file:///tmp/spec.yaml': must start with http:// or https://",
		},
		{
			name:           "empty string",
			args:           []string{""},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL '': must start with http:// or https://",
		},
		{
			name:           "bare http scheme",
			args:           []string{"http://"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'http://': must start with http:// or https://",
		},
		{
			name:           "bare https scheme",
			args:           []string{"https://"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'https://': must start with http:// or https://",
		},
		{
			name:           "bare uppercase HTTP scheme",
			args:           []string{"HTTP://"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'HTTP://': must start with http:// or https://",
		},
		{
			name:           "multiple positional arguments",
			args:           []string{"https://example.com/spec1.yaml", "https://example.com/spec2.yaml"},
			wantErr:        true,
			expectedErrMsg: "import-url accepts only 1 positional argument; specify multiple URLs as a comma-separated list: <url1,url2>",
		},
		{
			name:    "comma-separated valid URLs",
			args:    []string{"https://example.com/spec1.yaml,http://example.com/spec2.yaml"},
			wantErr: false,
		},
		{
			name:           "comma-separated with second URL invalid fails upfront",
			args:           []string{"https://example.com/spec1.yaml,spec2.yaml:true"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'spec2.yaml:true': must start with http:// or https://",
		},
		{
			name:           "comma-separated with first URL invalid fails upfront",
			args:           []string{"spec1.yaml,https://example.com/spec2.yaml"},
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'spec1.yaml': must start with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImportURLArgs(nil, tt.args)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErrMsg)
				assert.Equal(t, errors.KindUsage, errors.KindOf(err))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestImportURLCommandExecutionRejectsInvalidArgsBeforeRunE(t *testing.T) {
	// Create the command with dummy client options.
	// If Args validator did not reject the arguments upfront, RunE would proceed
	// to newCommandClient, which would fail with a missing microcksURL connection/usage error.
	cmd := NewImportURLCommand(&connectors.ClientOptions{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"https://example.com/spec1.yaml,spec2.yaml:true"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, errors.KindUsage, errors.KindOf(err))
	assert.EqualError(t, err, "invalid artifact URL 'spec2.yaml:true': must start with http:// or https://")
}
