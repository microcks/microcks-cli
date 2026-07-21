package connectors

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKeycloakConnectAndGetToken_Success(t *testing.T) {
	const expectedToken = "mock-access-token-123"
	const clientID = "test-client"
	const clientSecret = "test-secret"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/protocol/openid-connect/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		// Verify Content-Type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}

		// Verify Basic Auth Header
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))
		if authHeader != expectedAuth {
			t.Errorf("unexpected auth header: %s, want: %s", authHeader, expectedAuth)
		}

		// Verify Request Body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		bodyStr := string(bodyBytes)
		if bodyStr != "grant_type=client_credentials" {
			t.Errorf("unexpected body content: %s", bodyStr)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{
			"access_token": expectedToken,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewKeycloakClient(server.URL, clientID, clientSecret)
	token, err := client.ConnectAndGetToken()
	if err != nil {
		t.Fatalf("ConnectAndGetToken returned error: %v", err)
	}

	if token != expectedToken {
		t.Errorf("expected token %q, got %q", expectedToken, token)
	}
}

func TestKeycloakConnectAndGetTokenAndRefreshToken_Success(t *testing.T) {
	const expectedAccessToken = "mock-access-token-abc"
	const expectedRefreshToken = "mock-refresh-token-xyz"
	const clientID = "test-client"
	const clientSecret = "test-secret"
	const username = "john_doe"
	const password = "secure_password"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/protocol/openid-connect/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		// Verify Content-Type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}

		// Verify Body Parameters
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.FormValue("client_id") != clientID {
			t.Errorf("unexpected client_id: %s", r.FormValue("client_id"))
		}
		if r.FormValue("client_secret") != clientSecret {
			t.Errorf("unexpected client_secret: %s", r.FormValue("client_secret"))
		}
		if r.FormValue("username") != username {
			t.Errorf("unexpected username: %s", r.FormValue("username"))
		}
		if r.FormValue("password") != password {
			t.Errorf("unexpected password: %s", r.FormValue("password"))
		}
		if r.FormValue("grant_type") != "password" {
			t.Errorf("unexpected grant_type: %s", r.FormValue("grant_type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{
			"access_token":  expectedAccessToken,
			"refresh_token": expectedRefreshToken,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewKeycloakClient(server.URL, clientID, clientSecret)
	accessToken, refreshToken, err := client.ConnectAndGetTokenAndRefreshToken(username, password)
	if err != nil {
		t.Fatalf("ConnectAndGetTokenAndRefreshToken returned error: %v", err)
	}

	if accessToken != expectedAccessToken {
		t.Errorf("expected access token %q, got %q", expectedAccessToken, accessToken)
	}
	if refreshToken != expectedRefreshToken {
		t.Errorf("expected refresh token %q, got %q", expectedRefreshToken, refreshToken)
	}
}

func TestKeycloakGetOIDCConfig_Success(t *testing.T) {
	const expectedAuthEndpoint = "https://keycloak.example.com/auth"
	const expectedTokenEndpoint = "https://keycloak.example.com/token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{
			"authorization_endpoint": expectedAuthEndpoint,
			"token_endpoint":         expectedTokenEndpoint,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewKeycloakClient(server.URL, "clientID", "clientSecret")
	config, err := client.GetOIDCConfig()
	if err != nil {
		t.Fatalf("GetOIDCConfig returned error: %v", err)
	}

	if config.Endpoint.AuthURL != expectedAuthEndpoint {
		t.Errorf("expected auth URL %q, got %q", expectedAuthEndpoint, config.Endpoint.AuthURL)
	}
	if config.Endpoint.TokenURL != expectedTokenEndpoint {
		t.Errorf("expected token URL %q, got %q", expectedTokenEndpoint, config.Endpoint.TokenURL)
	}
}
