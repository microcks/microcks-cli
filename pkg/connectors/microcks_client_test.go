package connectors

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		_, _ = w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
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
		_, _ = w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)

	msg, err := client.DownloadArtifact("https://example.com/openapi.yaml", true, "")
	if err != nil {
		t.Fatalf("DownloadArtifact returned error: %v", err)
	}
	if strings.TrimSpace(msg) != expectedBody {
		t.Fatalf("expected response body %q, got %q", expectedBody, msg)
	}
}

func TestUnexpectedServerResponseHandling(t *testing.T) {
	// Create a test server that returns HTML / Bad Gateway
	const htmlError = `<html><body>502 Bad Gateway</body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(htmlError))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)

	// Test GetKeycloakURL
	_, err := client.GetKeycloakURL()
	if err == nil {
		t.Error("expected GetKeycloakURL to return error on Bad Gateway response, got nil")
	} else if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected error message to contain HTTP status code '502', got: %v", err)
	}

	// Test GetTestResult
	_, err = client.GetTestResult("some-id")
	if err == nil {
		t.Error("expected GetTestResult to return error on Bad Gateway response, got nil")
	} else if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected error message to contain HTTP status code '502', got: %v", err)
	}

	// Test DownloadArtifact
	_, err = client.DownloadArtifact("https://example.com/openapi.yaml", true, "")
	if err == nil {
		t.Error("expected DownloadArtifact to return error on Bad Gateway response, got nil")
	}
}

func TestGetKeycloakURLConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		expectedURL   string
		expectError   bool
		wantErrSubstr string
	}{
		{
			name:         "Valid Keycloak enabled config",
			responseBody: `{"enabled": true, "auth-server-url": "http://localhost:8080/auth", "realm": "microcks"}`,
			expectedURL:  "http://localhost:8080/auth/realms/microcks/",
			expectError:  false,
		},
		{
			name:         "Valid Keycloak disabled config",
			responseBody: `{"enabled": false}`,
			expectedURL:  "null",
			expectError:  false,
		},
		{
			name:          "Missing enabled field",
			responseBody:  `{"auth-server-url": "http://localhost:8080/auth", "realm": "microcks"}`,
			expectError:   true,
			wantErrSubstr: "missing 'enabled' field",
		},
		{
			name:          "Invalid type for enabled field",
			responseBody:  `{"enabled": "true", "auth-server-url": "http://localhost:8080/auth", "realm": "microcks"}`,
			expectError:   true,
			wantErrSubstr: "'enabled' field is not a boolean",
		},
		{
			name:          "Enabled but missing auth-server-url",
			responseBody:  `{"enabled": true, "realm": "microcks"}`,
			expectError:   true,
			wantErrSubstr: "missing auth-server-url or realm",
		},
		{
			name:          "Enabled but missing realm",
			responseBody:  `{"enabled": true, "auth-server-url": "http://localhost:8080/auth"}`,
			expectError:   true,
			wantErrSubstr: "missing auth-server-url or realm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/keycloak/config" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewMicrocksClient(server.URL)
			url, err := client.GetKeycloakURL()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if url != tt.expectedURL {
					t.Errorf("expected URL %q, got %q", tt.expectedURL, url)
				}
			}
		})
	}
}
