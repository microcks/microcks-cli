/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package connectors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	microckserrors "github.com/microcks/microcks-cli/pkg/errors"
)

func TestConnectAndGetTokenReturnsErrorOnNon200(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"unauthorized", http.StatusUnauthorized, `{"error":"unauthorized","error_description":"Invalid credentials"}`},
		{"forbidden", http.StatusForbidden, `{"error":"access_denied"}`},
		{"server error", http.StatusInternalServerError, `internal server error`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			kc, err := NewKeycloakClient(server.URL, "client-id", "client-secret")
			if err != nil {
				t.Fatalf("NewKeycloakClient returned unexpected error: %v", err)
			}

			token, err := kc.ConnectAndGetToken()

			if token != "" {
				t.Errorf("expected empty token on error, got %q", token)
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if microckserrors.KindOf(err) != microckserrors.KindAPI {
				t.Errorf("expected KindAPI, got %v", microckserrors.KindOf(err))
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("%d", tc.statusCode)) {
				t.Errorf("expected error to contain status code %d, got: %s", tc.statusCode, err.Error())
			}
		})
	}
}

func TestNewKeycloakClient_NormalizesTrailingSlash(t *testing.T) {
	cases := []struct {
		name     string
		inputURL string
		wantPath string
	}{
		{
			name:     "subpath without trailing slash",
			inputURL: "http://keycloak.example.com/realms/microcks",
			wantPath: "/realms/microcks/",
		},
		{
			name:     "subpath with trailing slash",
			inputURL: "http://keycloak.example.com/realms/microcks/",
			wantPath: "/realms/microcks/",
		},
		{
			name:     "root URL without slash",
			inputURL: "http://keycloak.example.com",
			wantPath: "/",
		},
		{
			name:     "nested subpath without trailing slash",
			inputURL: "http://keycloak.example.com/auth/realms/microcks",
			wantPath: "/auth/realms/microcks/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kc, err := NewKeycloakClient(tc.inputURL, "client", "secret")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			client, ok := kc.(*keycloakClient)
			if !ok {
				t.Fatal("expected *keycloakClient implementation")
			}
			if client.BaseURL.Path != tc.wantPath {
				t.Errorf("Path = %q, want %q", client.BaseURL.Path, tc.wantPath)
			}
		})
	}
}

func TestKeycloakClient_EndpointsPreserveRealmPath(t *testing.T) {
	var requestedPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/realms/microcks/protocol/openid-connect/token":
			fmt.Fprint(w, `{"access_token":"dummy-token","refresh_token":"dummy-refresh"}`)
		case "/realms/microcks/.well-known/openid-configuration":
			fmt.Fprint(w, `{"authorization_endpoint":"http://auth","token_endpoint":"http://token"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Provide realmURL without trailing slash
	realmURL := server.URL + "/realms/microcks"
	kc, err := NewKeycloakClient(realmURL, "client-id", "client-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. ConnectAndGetToken
	token, err := kc.ConnectAndGetToken()
	if err != nil {
		t.Fatalf("ConnectAndGetToken failed: %v", err)
	}
	if token != "dummy-token" {
		t.Errorf("expected dummy-token, got %q", token)
	}

	// 2. GetOIDCConfig
	oidcConfig, err := kc.GetOIDCConfig()
	if err != nil {
		t.Fatalf("GetOIDCConfig failed: %v", err)
	}
	if oidcConfig.Endpoint.AuthURL != "http://auth" {
		t.Errorf("unexpected auth URL: %s", oidcConfig.Endpoint.AuthURL)
	}

	// 3. ConnectAndGetTokenAndRefreshToken
	tok, refTok, err := kc.ConnectAndGetTokenAndRefreshToken("user", "pass")
	if err != nil {
		t.Fatalf("ConnectAndGetTokenAndRefreshToken failed: %v", err)
	}
	if tok != "dummy-token" || refTok != "dummy-refresh" {
		t.Errorf("unexpected tokens: %q, %q", tok, refTok)
	}

	// Verify all requests preserved the /realms/microcks realm path
	for _, p := range requestedPaths {
		if !strings.HasPrefix(p, "/realms/microcks/") {
			t.Errorf("request path %q stripped the realm segment", p)
		}
	}
}
