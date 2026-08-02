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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	microckserrors "github.com/microcks/microcks-cli/pkg/errors"
)

func TestUploadArtifactStreamsWithoutBuffering(t *testing.T) {
	const fileContent = `{"openapi":"3.0.0","info":{"title":"Test API","version":"1.0.0"}}`
	const expectedResponse = "artifact uploaded"

	// Create a temporary file to simulate an API specification.
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")
	if err := os.WriteFile(specPath, []byte(fileContent), 0o600); err != nil {
		t.Fatalf("failed to create temp spec file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/artifact/upload" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		// Verify the multipart form contains the file.
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to get form file: %v", err)
		}
		defer file.Close()

		if header.Filename != "openapi.json" {
			t.Fatalf("unexpected filename: %s", header.Filename)
		}

		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("failed to read uploaded file: %v", err)
		}
		if string(body) != fileContent {
			t.Fatalf("file content mismatch: got %q, want %q", string(body), fileContent)
		}

		// Verify the mainArtifact field.
		if got := r.FormValue("mainArtifact"); got != "true" {
			t.Fatalf("unexpected mainArtifact value: %s", got)
		}

		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(expectedResponse)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewMicrocksClient(server.URL)
	if err != nil {
		t.Fatalf("NewMicrocksClient returned error: %v", err)
	}
	msg, err := client.UploadArtifact(specPath, true)
	if err != nil {
		t.Fatalf("UploadArtifact returned error: %v", err)
	}
	if strings.TrimSpace(msg) != expectedResponse {
		t.Fatalf("expected response %q, got %q", expectedResponse, msg)
	}
}

func TestDownloadArtifactReturnsResponseBody(t *testing.T) {
	const expectedBody = "artifact downloaded"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/artifact/download" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseMultipartForm(1024); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		if got := r.FormValue("url"); got != "https://example.com/openapi.yaml" {
			t.Fatalf("unexpected artifact url: %s", got)
		}
		if got := r.FormValue("mainArtifact"); got != "true" {
			t.Fatalf("unexpected mainArtifact value: %s", got)
		}
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(expectedBody)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewMicrocksClient(server.URL)
	if err != nil {
		t.Fatalf("NewMicrocksClient returned error: %v", err)
	}

	msg, err := client.DownloadArtifact("https://example.com/openapi.yaml", true, "")
	if err != nil {
		t.Fatalf("DownloadArtifact returned error: %v", err)
	}
	if strings.TrimSpace(msg) != expectedBody {
		t.Fatalf("expected response body %q, got %q", expectedBody, msg)
	}
}

func TestGetKeycloakURLRejectsMalformedConfig(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "missing enabled",
			body: `{"auth-server-url":"http://keycloak","realm":"microcks"}`,
			want: "enabled",
		},
		{
			name: "invalid auth server url",
			body: `{"enabled":true,"auth-server-url":42,"realm":"microcks"}`,
			want: "auth-server-url",
		},
		{
			name: "invalid realm",
			body: `{"enabled":true,"auth-server-url":"http://keycloak","realm":42}`,
			want: "realm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/keycloak/config" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(tt.body)); err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client, err := NewMicrocksClient(server.URL)
			if err != nil {
				t.Fatalf("NewMicrocksClient returned error: %v", err)
			}

			_, err = client.GetKeycloakURL()
			if err == nil {
				t.Fatal("GetKeycloakURL returned nil error")
			}
			if got := microckserrors.KindOf(err); got != microckserrors.KindAPI {
				t.Fatalf("KindOf = %v, want %v", got, microckserrors.KindAPI)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %q", err.Error(), tt.want)
			}
		})
	}
}

func TestCreateTestResultClassifiesMalformedResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "invalid json", body: `not-json`, want: "parse test creation response"},
		{name: "missing id", body: `{}`, want: "missing 'id'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/tests" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusCreated)
				if _, err := w.Write([]byte(tt.body)); err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client, err := NewMicrocksClient(server.URL)
			if err != nil {
				t.Fatalf("NewMicrocksClient returned error: %v", err)
			}

			_, err = client.CreateTestResult("service:1.0", "http://example.test", "OPEN_API_SCHEMA", "", 1000, "", "", "")
			if err == nil {
				t.Fatal("CreateTestResult returned nil error")
			}
			if got := microckserrors.KindOf(err); got != microckserrors.KindAPI {
				t.Fatalf("KindOf = %v, want %v", got, microckserrors.KindAPI)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %q", err.Error(), tt.want)
			}
		})
	}
}

func TestCreateTestResultRejectsInvalidFilteredOperations(t *testing.T) {
	client, err := NewMicrocksClient("http://localhost:8585")
	if err != nil {
		t.Fatalf("NewMicrocksClient returned error: %v", err)
	}

	_, err = client.CreateTestResult("service:1.0", "http://example.test", "OPEN_API_SCHEMA", "", 1000, "{", "", "")
	if err == nil {
		t.Fatal("CreateTestResult returned nil error")
	}
	if got := microckserrors.KindOf(err); got != microckserrors.KindUsage {
		t.Fatalf("KindOf = %v, want %v", got, microckserrors.KindUsage)
	}
}
