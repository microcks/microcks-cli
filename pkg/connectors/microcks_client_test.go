package connectors

import (
	"net/http"
	"net/http/httptest"
	"os"
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

func TestLoggingTransportPassesThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	lt := &loggingTransport{transport: http.DefaultTransport}

	old := os.Stderr
	pr, pw, _ := os.Pipe()
	os.Stderr = pw

	resp, err := lt.RoundTrip(req)

	pw.Close()
	os.Stderr = old
	pr.Close()

	if err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLoggingTransportRedactsAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Authorization", "Bearer supersecret")

	old := os.Stderr
	pr, pw, _ := os.Pipe()
	os.Stderr = pw

	lt := &loggingTransport{transport: http.DefaultTransport}
	_, _ = lt.RoundTrip(req)

	pw.Close()
	os.Stderr = old

	buf := make([]byte, 4096)
	n, _ := pr.Read(buf)
	out := string(buf[:n])

	if strings.Contains(out, "supersecret") {
		t.Fatal("loggingTransport leaked the Authorization token")
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatal("loggingTransport did not redact the Authorization header")
	}
}
