package connectors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConnectAndGetTokenReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized_client"}`))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "bad-id", "bad-secret")
	_, err := kc.ConnectAndGetToken()
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("expected error to mention HTTP 401, got: %s", err.Error())
	}
}

func TestConnectAndGetTokenReturnsErrorOnMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>not json</html>"))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "id", "secret")
	_, err := kc.ConnectAndGetToken()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Fatalf("expected parse error, got: %s", err.Error())
	}
}

func TestConnectAndGetTokenReturnsErrorOnMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token_type":"bearer"}`))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "id", "secret")
	_, err := kc.ConnectAndGetToken()
	if err == nil {
		t.Fatal("expected error for missing access_token, got nil")
	}
	if !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("expected missing field error, got: %s", err.Error())
	}
}

func TestGetOIDCConfigReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "", "")
	_, err := kc.GetOIDCConfig()
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("expected error to mention HTTP 404, got: %s", err.Error())
	}
}

func TestGetOIDCConfigReturnsErrorOnMissingFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"issuer":"https://auth.example.com"}`))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "", "")
	_, err := kc.GetOIDCConfig()
	if err == nil {
		t.Fatal("expected error for missing OIDC fields, got nil")
	}
	if !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("expected missing field error, got: %s", err.Error())
	}
}

func TestConnectAndGetTokenAndRefreshTokenReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "client-id", "client-secret")
	_, _, err := kc.ConnectAndGetTokenAndRefreshToken("user", "wrong-pass")
	if err == nil {
		t.Fatal("expected error for 400 response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("expected error to mention HTTP 400, got: %s", err.Error())
	}
}

func TestConnectAndGetTokenAndRefreshTokenReturnsErrorOnMissingFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"tok123"}`))
	}))
	defer server.Close()

	kc := NewKeycloakClient(server.URL+"/", "client-id", "client-secret")
	_, _, err := kc.ConnectAndGetTokenAndRefreshToken("user", "pass")
	if err == nil {
		t.Fatal("expected error for missing refresh_token field, got nil")
	}
	if !strings.Contains(err.Error(), "refresh_token") {
		t.Fatalf("expected error to mention refresh_token, got: %s", err.Error())
	}
}
