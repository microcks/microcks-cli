package connectors

import (
	"encoding/json"
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

func TestGetServices(t *testing.T) {
	services := []ServiceSummary{
		{ID: "1", Name: "Petstore API", Version: "1.0", Type: "REST"},
		{ID: "2", Name: "HelloService", Version: "0.9", Type: "SOAP"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.URL.Query().Get("page"); got != "0" {
			t.Fatalf("unexpected page: %s", got)
		}
		if got := r.URL.Query().Get("size"); got != "20" {
			t.Fatalf("unexpected size: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(services)
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	result, err := client.GetServices(0, 20)
	if err != nil {
		t.Fatalf("GetServices returned error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 services, got %d", len(result))
	}
	if result[0].Name != "Petstore API" {
		t.Fatalf("expected first service name %q, got %q", "Petstore API", result[0].Name)
	}
	if result[0].Version != "1.0" {
		t.Fatalf("expected first service version %q, got %q", "1.0", result[0].Version)
	}
	if result[0].Type != "REST" {
		t.Fatalf("expected first service type %q, got %q", "REST", result[0].Type)
	}
}

func TestGetServicesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	result, err := client.GetServices(0, 20)
	if err != nil {
		t.Fatalf("GetServices returned error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %d services", len(result))
	}
}

func TestGetServicesInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetServices(0, 20)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse services response") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGetServicesPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("unexpected page: %s", got)
		}
		if got := r.URL.Query().Get("size"); got != "5" {
			t.Fatalf("unexpected size: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetServices(2, 5)
	if err != nil {
		t.Fatalf("GetServices returned error: %v", err)
	}
}
