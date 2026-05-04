package connectors

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTestResultEscapesSpecialChars(t *testing.T) {
	// serviceID contains quotes + backslash — old string-concat would have produced broken JSON.
	maliciousServiceID := `svc"id\bad`
	maliciousEndpoint := `http://evil.com", "injected":"val`

	var gotPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"test-123"}`))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	id, err := client.CreateTestResult(maliciousServiceID, maliciousEndpoint, "HTTP", "", 5000, "", "", "")
	if err != nil {
		t.Fatalf("CreateTestResult returned error: %v", err)
	}
	if id != "test-123" {
		t.Fatalf("expected id test-123, got %s", id)
	}

	// Values must be preserved exactly — no injection of extra fields.
	if got, ok := gotPayload["serviceId"].(string); !ok || got != maliciousServiceID {
		t.Fatalf("serviceId not preserved: %v", gotPayload["serviceId"])
	}
	if got, ok := gotPayload["testEndpoint"].(string); !ok || got != maliciousEndpoint {
		t.Fatalf("testEndpoint not preserved: %v", gotPayload["testEndpoint"])
	}
	if _, injected := gotPayload["injected"]; injected {
		t.Fatal("JSON injection succeeded — extra field 'injected' found in payload")
	}
}

func TestCreateTestResultOmitsEmptySecretName(t *testing.T) {
	var gotPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"test-456"}`))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.CreateTestResult("svcA", "http://example.com", "HTTP", "", 3000, "", "", "")
	if err != nil {
		t.Fatalf("CreateTestResult returned error: %v", err)
	}
	if _, exists := gotPayload["secretName"]; exists {
		t.Fatal("secretName should be omitted when empty")
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
