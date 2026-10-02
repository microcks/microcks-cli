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
		wantErr              bool
		expectedErrMsg       string
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
			name:           "relative path without scheme",
			input:          "spec.yaml",
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'spec.yaml': must start with http:// or https://",
		},
		{
			name:           "relative path with suffixes without scheme",
			input:          "spec.yaml:true:mysecret",
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'spec.yaml:true:mysecret': must start with http:// or https://",
		},
		{
			name:           "unsupported ftp scheme",
			input:          "ftp://example.com/spec.yaml",
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'ftp://example.com/spec.yaml': must start with http:// or https://",
		},
		{
			name:           "file scheme",
			input:          "file:///tmp/spec.yaml",
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL 'file:///tmp/spec.yaml': must start with http:// or https://",
		},
		{
			name:           "empty string",
			input:          "",
			wantErr:        true,
			expectedErrMsg: "invalid artifact URL '': must start with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, mainArtifact, secret, err := parseImportURLArg(tt.input)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErrMsg)
				assert.Equal(t, errors.KindUsage, errors.KindOf(err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedURL, url)
			assert.Equal(t, tt.expectedMainArtifact, mainArtifact)
			assert.Equal(t, tt.expectedSecret, secret)
		})
	}
}
