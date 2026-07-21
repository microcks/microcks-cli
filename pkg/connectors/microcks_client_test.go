package connectors

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
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

	client := NewMicrocksClient(server.URL)
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

	client := NewMicrocksClient(server.URL)

	msg, err := client.DownloadArtifact("https://example.com/openapi.yaml", true, "")
	if err != nil {
		t.Fatalf("DownloadArtifact returned error: %v", err)
	}
	if strings.TrimSpace(msg) != expectedBody {
		t.Fatalf("expected response body %q, got %q", expectedBody, msg)
	}
}

func createDummyJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp)))
	return header + "." + payload + "."
}

func TestRefreshAuthToken_ValidTokenNoRefresh(t *testing.T) {
	// A token with expiration 1 hour in the future
	futureTime := time.Now().Add(1 * time.Hour).Unix()
	dummyToken := createDummyJWT(futureTime)

	// Setup local config
	localCfg := &config.LocalConfig{
		CurrentContext: "test-context",
		Contexts: []config.ContextRef{
			{Name: "test-context", Server: "localhost", User: "test-user"},
		},
		Servers: []config.Server{
			{Name: "localhost", Server: "localhost"},
		},
		Users: []config.User{
			{Name: "test-user", AuthToken: dummyToken, RefreshToken: "some-refresh-token"},
		},
	}

	mc := &microcksClient{
		AuthToken:    dummyToken,
		RefreshToken: "some-refresh-token",
	}

	// Calling refreshAuthToken with a valid token should do nothing and return nil
	err := mc.refreshAuthToken(localCfg, "test-context", "")
	if err != nil {
		t.Fatalf("refreshAuthToken failed: %v", err)
	}

	// Verify token was not modified
	if mc.AuthToken != dummyToken {
		t.Errorf("expected AuthToken to remain %q, got %q", dummyToken, mc.AuthToken)
	}
}

func TestRefreshAuthToken_ExpiredTokenTriggersRefresh(t *testing.T) {
	// A token with expiration 1 hour in the past
	pastTime := time.Now().Add(-1 * time.Hour).Unix()
	expiredToken := createDummyJWT(pastTime)

	// We need a temporary config file path since the function calls WriteLocalConfig
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Setup local config. Note: refreshAuthToken uses the context name ("test-context")
	// as the name of the user to upsert, so we name the user "test-context" to match.
	localCfg := &config.LocalConfig{
		CurrentContext: "test-context",
		Contexts: []config.ContextRef{
			{Name: "test-context", Server: "http://localhost", User: "test-context"},
		},
		Servers: []config.Server{
			{Server: "http://localhost"},
		},
		Users: []config.User{
			{Name: "test-context", AuthToken: expiredToken, RefreshToken: "old-refresh-token"},
		},
		Auths: []config.Auth{
			{Server: "http://localhost", ClientId: "cli", ClientSecret: "secret"},
		},
	}

	// Write initial localconfig to the temp file
	if err := config.WriteLocalConfig(*localCfg, configPath); err != nil {
		t.Fatalf("failed to write local config: %v", err)
	}

	// Spin up mock server handling Microcks client / Keycloak routes
	var mockServer *httptest.Server
	mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/keycloak/config":
			// Return keycloak config pointing to this mock server
			resp := map[string]interface{}{
				"enabled":         true,
				"auth-server-url": mockServer.URL,
				"realm":           "microcks",
			}
			json.NewEncoder(w).Encode(resp)
		case "/realms/microcks/.well-known/openid-configuration":
			// Return OIDC metadata pointing to token endpoint on mock server
			resp := map[string]string{
				"authorization_endpoint": mockServer.URL + "/realms/microcks/protocol/openid-connect/auth",
				"token_endpoint":         mockServer.URL + "/realms/microcks/protocol/openid-connect/token",
			}
			json.NewEncoder(w).Encode(resp)
		case "/realms/microcks/protocol/openid-connect/token":
			// Verify request body for refresh token grant
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			if r.FormValue("grant_type") != "refresh_token" {
				t.Errorf("unexpected grant_type: %q", r.FormValue("grant_type"))
			}
			if r.FormValue("refresh_token") != "old-refresh-token" {
				t.Errorf("unexpected refresh_token: %q", r.FormValue("refresh_token"))
			}
			
			// Return new tokens
			resp := map[string]string{
				"access_token":  "new-access-token",
				"refresh_token": "new-refresh-token",
			}
			json.NewEncoder(w).Encode(resp)
		default:
			t.Fatalf("unexpected request to: %s", r.URL.Path)
		}
	}))
	defer mockServer.Close()

	apiURL, err := url.Parse(mockServer.URL + "/api/")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	mc := &microcksClient{
		APIURL:       apiURL,
		AuthToken:    expiredToken,
		RefreshToken: "old-refresh-token",
		httpClient:   mockServer.Client(),
	}

	err = mc.refreshAuthToken(localCfg, "test-context", configPath)
	if err != nil {
		t.Fatalf("refreshAuthToken failed: %v", err)
	}

	// Verify client tokens were updated
	if mc.AuthToken != "new-access-token" {
		t.Errorf("expected AuthToken to be refreshed to %q, got %q", "new-access-token", mc.AuthToken)
	}
	if mc.RefreshToken != "new-refresh-token" {
		t.Errorf("expected RefreshToken to be refreshed to %q, got %q", "new-refresh-token", mc.RefreshToken)
	}

	// Verify local config was updated and written back to file
	updatedCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		t.Fatalf("failed to read back config: %v", err)
	}
	user, err := updatedCfg.GetUser("test-context")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if user.AuthToken != "new-access-token" {
		t.Errorf("expected config AuthToken to be %q, got %q", "new-access-token", user.AuthToken)
	}
	if user.RefreshToken != "new-refresh-token" {
		t.Errorf("expected config RefreshToken to be %q, got %q", "new-refresh-token", user.RefreshToken)
	}
}
