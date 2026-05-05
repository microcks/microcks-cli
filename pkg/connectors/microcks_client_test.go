package connectors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestDeleteServiceReturnsNoErrorOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/services/search" {
			if got := r.URL.Query().Get("name"); got != "Simple" {
				t.Fatalf("unexpected service name: %s", got)
			}
			if got := r.URL.Query().Get("version"); got != "1.1" {
				t.Fatalf("unexpected service version: %s", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id": "test-id-123", "name": "Simple", "version": "1.1"}]`))
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/api/services/test-id-123" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)

	err := client.DeleteService("Simple", "1.1")
	if err != nil {
		t.Fatalf("DeleteService returned error: %v", err)
	}
}

func TestDeleteServiceReturnsErrorOnNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/services/search" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)

	err := client.DeleteService("Simple", "1.1")
	if err == nil {
		t.Fatalf("expected DeleteService to return error on not found, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected error to contain 'not found', got: %v", err)
	}
}
