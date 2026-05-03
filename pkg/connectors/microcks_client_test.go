package connectors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetKeycloakURL_Disabled(t *testing.T) {
	// Create a mock server that returns keycloak disabled response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/keycloak/config" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"enabled": false}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Initialize the MicrocksClient with the mock server URL
	client := NewMicrocksClient(ts.URL)

	// Call GetKeycloakURL and check for panic/error
	url, err := client.GetKeycloakURL()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if url != "null" {
		t.Errorf("Expected 'null', got '%s'", url)
	}
}

func TestGetKeycloakURL_Enabled(t *testing.T) {
	// Create a mock server that returns keycloak enabled response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/keycloak/config" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"enabled": true, "auth-server-url": "http://keycloak:8080/auth", "realm": "microcks"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Initialize the MicrocksClient with the mock server URL
	client := NewMicrocksClient(ts.URL)

	// Call GetKeycloakURL and check for panic/error
	url, err := client.GetKeycloakURL()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "http://keycloak:8080/auth/realms/microcks/"
	if url != expected {
		t.Errorf("Expected '%s', got '%s'", expected, url)
	}
}
