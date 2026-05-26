package connectors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
)

func TestGetKeycloakURLReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetKeycloakURL()
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("expected error to mention HTTP 500, got: %s", err.Error())
	}
}

func TestGetKeycloakURLReturnsErrorOnMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetKeycloakURL()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Fatalf("expected parse error, got: %s", err.Error())
	}
}

func TestGetKeycloakURLReturnsErrorOnMissingFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"enabled": true}`))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetKeycloakURL()
	if err == nil {
		t.Fatal("expected error for missing fields, got nil")
	}
	if !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("expected missing field error, got: %s", err.Error())
	}
}

func TestGetKeycloakURLReturnsNullWhenKeycloakDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"enabled": false}`))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	keycloakURL, err := client.GetKeycloakURL()
	if err != nil {
		t.Fatalf("expected no error for disabled Keycloak config, got: %s", err.Error())
	}
	if keycloakURL != "null" {
		t.Fatalf("expected disabled Keycloak URL to be null, got: %s", keycloakURL)
	}
}

func TestRedeemRefreshTokenReturnsKeycloakConfigError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("config unavailable"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL).(*microcksClient)
	_, _, err := client.redeemRefreshToken(config.Auth{})
	if err == nil {
		t.Fatal("expected error for Keycloak config failure, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("expected error to mention HTTP 500, got: %s", err.Error())
	}
}

func TestCreateTestResultReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/keycloak/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"enabled":         false,
				"auth-server-url": "",
				"realm":           "",
			})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	client.SetOAuthToken("expired-token")

	_, err := client.CreateTestResult("svc:1.0", "http://endpoint", "HTTP", "", 5000, "", "", "")
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("expected error to mention HTTP 401, got: %s", err.Error())
	}
}

func TestCreateTestResultReturnsErrorOnMissingID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "created"}`))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	client.SetOAuthToken("test-token")

	_, err := client.CreateTestResult("svc:1.0", "http://endpoint", "HTTP", "", 5000, "", "", "")
	if err == nil {
		t.Fatal("expected error for missing id field, got nil")
	}
	if !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("expected missing field error, got: %s", err.Error())
	}
}

func TestGetTestResultReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	client.SetOAuthToken("test-token")

	_, err := client.GetTestResult("abc-123")
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("expected error to mention HTTP 403, got: %s", err.Error())
	}
}

func TestDownloadArtifactReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	client.SetOAuthToken("test-token")

	_, err := client.DownloadArtifact("https://example.com/spec.yaml", true, "")
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected error body in message, got: %s", err.Error())
	}
}
