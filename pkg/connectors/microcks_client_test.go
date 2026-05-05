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
package connectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureValidOperationsList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid operations list with multiple entries",
			input:    `["GET /pastries", "GET /pastries/{name}"]`,
			expected: true,
		},
		{
			name:     "valid operations list with single entry",
			input:    `["POST /orders"]`,
			expected: true,
		},
		{
			name:     "valid empty list",
			input:    `[]`,
			expected: true,
		},
		{
			name:     "invalid JSON - not an array",
			input:    `"GET /pastries"`,
			expected: false,
		},
		{
			name:     "invalid JSON - object instead of array",
			input:    `{"operation": "GET /pastries"}`,
			expected: false,
		},
		{
			name:     "invalid JSON - malformed",
			input:    `[GET /pastries]`,
			expected: false,
		},
		{
			name:     "invalid JSON - completely broken",
			input:    `not json at all`,
			expected: false,
		},
		{
			name:     "invalid JSON - missing closing bracket",
			input:    `["GET /pastries"`,
			expected: false,
		},
		{
			name:     "valid list with special characters in operation names",
			input:    `["GET /pastries?limit=10", "POST /orders/{id}/items"]`,
			expected: true,
		},
		{
			name:     "invalid JSON - array of numbers instead of strings",
			input:    `[1, 2, 3]`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureValidOperationsList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureValidOperationsHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid operations headers with single operation",
			input:    `{"GET /pastries": [{"name": "X-Custom-Header", "values": "value1"}]}`,
			expected: true,
		},
		{
			name:     "valid operations headers with multiple operations",
			input:    `{"GET /pastries": [{"name": "Authorization", "values": "Bearer token"}], "POST /orders": [{"name": "Content-Type", "values": "application/json"}]}`,
			expected: true,
		},
		{
			name:     "valid operations headers with multiple headers per operation",
			input:    `{"GET /pastries": [{"name": "X-Header-1", "values": "val1"}, {"name": "X-Header-2", "values": "val2"}]}`,
			expected: true,
		},
		{
			name:     "valid empty map",
			input:    `{}`,
			expected: true,
		},
		{
			name:     "invalid JSON - malformed",
			input:    `{invalid}`,
			expected: false,
		},
		{
			name:     "invalid JSON - plain string",
			input:    `not json`,
			expected: false,
		},
		{
			name:     "invalid JSON - array instead of map",
			input:    `[{"name": "X-Header", "values": "val"}]`,
			expected: false,
		},
		{
			name:     "invalid JSON - wrong value type (string instead of array)",
			input:    `{"GET /pastries": "not-an-array"}`,
			expected: false,
		},
		{
			name:     "valid operation with empty header array",
			input:    `{"GET /pastries": []}`,
			expected: true,
		},
		{
			name:     "invalid JSON - missing closing brace",
			input:    `{"GET /pastries": [{"name": "X-Header", "values": "val"}]`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureValidOperationsHeaders(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureValieOAuth2Context(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "valid PASSWORD grant type",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"username": "user",
				"password": "pass",
				"grantType": "PASSWORD"
			}`,
			expected: true,
		},
		{
			name: "valid CLIENT_CREDENTIALS grant type",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"grantType": "CLIENT_CREDENTIALS"
			}`,
			expected: true,
		},
		{
			name: "valid REFRESH_TOKEN grant type",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"refreshToken": "some-refresh-token",
				"grantType": "REFRESH_TOKEN"
			}`,
			expected: true,
		},
		{
			name: "valid context with optional scopes field",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"grantType": "CLIENT_CREDENTIALS",
				"scopes": "read write"
			}`,
			expected: true,
		},
		{
			name: "invalid - unsupported grant type",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"grantType": "IMPLICIT"
			}`,
			expected: false,
		},
		{
			name: "invalid - empty grant type",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"grantType": ""
			}`,
			expected: false,
		},
		{
			name: "invalid - missing grant type field",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token"
			}`,
			expected: false,
		},
		{
			name:     "invalid JSON - malformed",
			input:    `{not valid json}`,
			expected: false,
		},
		{
			name:     "invalid JSON - plain string",
			input:    `not json at all`,
			expected: false,
		},
		{
			name:     "invalid JSON - empty string",
			input:    ``,
			expected: false,
		},
		{
			name:     "invalid JSON - array instead of object",
			input:    `["PASSWORD"]`,
			expected: false,
		},
		{
			name: "invalid - grant type with wrong casing",
			input: `{
				"clientId": "my-client",
				"clientSecret": "my-secret",
				"tokenUri": "https://keycloak.example.com/token",
				"grantType": "password"
			}`,
			expected: false,
		},
		{
			name: "valid - minimal fields with CLIENT_CREDENTIALS",
			input: `{
				"clientId": "id",
				"clientSecret": "secret",
				"grantType": "CLIENT_CREDENTIALS"
			}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureValieOAuth2Context(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
