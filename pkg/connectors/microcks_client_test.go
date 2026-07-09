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
		_, _ = w.Write([]byte(expectedBody))
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
