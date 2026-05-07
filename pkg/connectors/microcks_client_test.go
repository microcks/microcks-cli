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

	msg, err := client.DownloadArtifact("https://example.com/openapi.yaml", true, "", false)
	if err != nil {
		t.Fatalf("DownloadArtifact returned error: %v", err)
	}
	if strings.TrimSpace(msg) != expectedBody {
		t.Fatalf("expected response body %q, got %q", expectedBody, msg)
	}
}

func TestDownloadArtifactRejectsPrivateIP(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		errMsg string
	}{
		{
			name:   "cloud metadata IP",
			url:    "https://169.254.169.254/latest/meta-data/",
			errMsg: "private/reserved address",
		},
		{
			name:   "localhost IP",
			url:    "https://127.0.0.1/api",
			errMsg: "private/reserved address",
		},
		{
			name:   "RFC1918 IP",
			url:    "https://10.0.0.1/api",
			errMsg: "private/reserved address",
		},
		{
			name:   "localhost hostname",
			url:    "https://localhost/api",
			errMsg: "internal address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("server should not be reached for invalid URLs")
			}))
			defer server.Close()

			client := NewMicrocksClient(server.URL)

			_, err := client.DownloadArtifact(tt.url, true, "", true)
			if err == nil {
				t.Fatalf("expected error for URL %q, got nil", tt.url)
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Fatalf("expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestDownloadArtifactRejectsHTTPScheme(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be reached for http:// URLs without allowInsecure")
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)

	_, err := client.DownloadArtifact("http://93.184.216.34/spec.yaml", true, "", false)
	if err == nil {
		t.Fatal("expected error for http:// URL with allowInsecure=false, got nil")
	}
	if !strings.Contains(err.Error(), "http scheme is not allowed") {
		t.Fatalf("expected error about http scheme, got %q", err.Error())
	}
}

func TestDownloadArtifactAllowsHTTPWithFlag(t *testing.T) {
	const expectedBody = "artifact downloaded"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1024); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		if got := r.FormValue("url"); got != "http://93.184.216.34/spec.yaml" {
			t.Fatalf("unexpected artifact url: %s", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)

	msg, err := client.DownloadArtifact("http://93.184.216.34/spec.yaml", true, "", true)
	if err != nil {
		t.Fatalf("DownloadArtifact returned error: %v", err)
	}
	if strings.TrimSpace(msg) != expectedBody {
		t.Fatalf("expected response body %q, got %q", expectedBody, msg)
	}
}
