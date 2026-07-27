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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKeycloakClientNegativeScenarios(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		checkErrSubstr string
	}{
		{
			name:           "Non-200 status code",
			statusCode:     http.StatusInternalServerError,
			responseBody:   "Internal Server Error Details",
			checkErrSubstr: "returned HTTP 500",
		},
		{
			name:           "Non-JSON response",
			statusCode:     http.StatusOK,
			responseBody:   "not-a-json-string",
			checkErrSubstr: "parsing Keycloak",
		},
		{
			name:           "JSON response with missing required fields",
			statusCode:     http.StatusOK,
			responseBody:   `{"some_key": "some_value"}`,
			checkErrSubstr: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Add trailing slash to server.URL so ResolveReference resolves paths properly
			client := NewKeycloakClient(server.URL+"/", "client-id", "client-secret")

			// Test ConnectAndGetToken
			_, err := client.ConnectAndGetToken()
			if err == nil {
				t.Error("expected ConnectAndGetToken to return error, got nil")
			} else if !strings.Contains(err.Error(), tt.checkErrSubstr) {
				t.Errorf("ConnectAndGetToken: expected error containing %q, got: %v", tt.checkErrSubstr, err)
			}

			// Test ConnectAndGetTokenAndRefreshToken
			_, _, err = client.ConnectAndGetTokenAndRefreshToken("user", "pass")
			if err == nil {
				t.Error("expected ConnectAndGetTokenAndRefreshToken to return error, got nil")
			} else if !strings.Contains(err.Error(), tt.checkErrSubstr) {
				t.Errorf("ConnectAndGetTokenAndRefreshToken: expected error containing %q, got: %v", tt.checkErrSubstr, err)
			}

			// Test GetOIDCConfig
			_, err = client.GetOIDCConfig()
			if err == nil {
				t.Error("expected GetOIDCConfig to return error, got nil")
			} else if !strings.Contains(err.Error(), tt.checkErrSubstr) {
				t.Errorf("GetOIDCConfig: expected error containing %q, got: %v", tt.checkErrSubstr, err)
			}
		})
	}
}
