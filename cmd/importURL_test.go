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

	"github.com/stretchr/testify/assert"
)

func TestParseImportURLArg(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedURL         string
		expectedMainArtifact bool
		expectedSecret      string
	}{
		{
			name:                "standard URL without suffixes",
			input:               "https://example.com/openapi.yaml",
			expectedURL:         "https://example.com/openapi.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "standard URL with mainArtifact true",
			input:               "https://example.com/spec1.yaml:true",
			expectedURL:         "https://example.com/spec1.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "standard URL with mainArtifact false",
			input:               "https://example.com/spec1.yaml:false",
			expectedURL:         "https://example.com/spec1.yaml",
			expectedMainArtifact: false,
			expectedSecret:      "",
		},
		{
			name:                "standard URL with mainArtifact and secret",
			input:               "https://example.com/spec1.yaml:true:my-secret",
			expectedURL:         "https://example.com/spec1.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "my-secret",
		},
		{
			name:                "standard URL with mainArtifact false and secret",
			input:               "https://example.com/spec1.yaml:false:my-secret-token",
			expectedURL:         "https://example.com/spec1.yaml",
			expectedMainArtifact: false,
			expectedSecret:      "my-secret-token",
		},
		{
			name:                "URL with port and no suffixes",
			input:               "http://localhost:8585/spec.yaml",
			expectedURL:         "http://localhost:8585/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "URL with port and mainArtifact true",
			input:               "http://localhost:8585/spec.yaml:true",
			expectedURL:         "http://localhost:8585/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "URL with port and mainArtifact false",
			input:               "http://localhost:8585/spec.yaml:false",
			expectedURL:         "http://localhost:8585/spec.yaml",
			expectedMainArtifact: false,
			expectedSecret:      "",
		},
		{
			name:                "URL with port, mainArtifact and secret",
			input:               "http://localhost:8585/spec.yaml:true:my-secret-token",
			expectedURL:         "http://localhost:8585/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "my-secret-token",
		},
		{
			name:                "URL with port, mainArtifact false and secret",
			input:               "http://localhost:8585/spec.yaml:false:my-secret-token",
			expectedURL:         "http://localhost:8585/spec.yaml",
			expectedMainArtifact: false,
			expectedSecret:      "my-secret-token",
		},
		{
			name:                "malformed bool suffix preserves URL",
			input:               "http://localhost:8585/spec.yaml:tru",
			expectedURL:         "http://localhost:8585/spec.yaml:tru",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "malformed bool with secret preserves URL",
			input:               "http://localhost:8585/spec.yaml:tru:mysecret",
			expectedURL:         "http://localhost:8585/spec.yaml:tru:mysecret",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "URL with path and no port",
			input:               "https://raw.githubusercontent.com/org/repo/main/spec.yaml",
			expectedURL:         "https://raw.githubusercontent.com/org/repo/main/spec.yaml",
			expectedMainArtifact: true,
			expectedSecret:      "",
		},
		{
			name:                "short URL without path",
			input:               "http://localhost:true",
			expectedURL:         "http://localhost",
			expectedMainArtifact: true,
			expectedSecret:      "",
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
