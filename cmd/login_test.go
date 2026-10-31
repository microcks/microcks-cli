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

package cmd

import (
	stderrors "errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	clierrors "github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOAuthCallbackServerUsesIsolatedMuxes(t *testing.T) {
	firstStatus := startOAuthCallbackTestServer(t, http.StatusCreated)
	secondStatus := startOAuthCallbackTestServer(t, http.StatusAccepted)

	if firstStatus != http.StatusCreated {
		t.Fatalf("first callback server returned status %d, want %d", firstStatus, http.StatusCreated)
	}
	if secondStatus != http.StatusAccepted {
		t.Fatalf("second callback server returned status %d, want %d", secondStatus, http.StatusAccepted)
	}
}

func TestNewOAuthCallbackServerReturnsErrorWhenPortIsOccupied(t *testing.T) {
	occupiedListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := occupiedListener.Close(); err != nil {
			t.Errorf("closing occupied listener: %v", err)
		}
	})

	server, listener, err := newOAuthCallbackServer(occupiedListener.Addr().String(), http.NotFound)
	if err == nil {
		if server != nil {
			if closeErr := server.Close(); closeErr != nil {
				t.Errorf("closing unexpected callback server: %v", closeErr)
			}
		}
		if listener != nil {
			if closeErr := listener.Close(); closeErr != nil {
				t.Errorf("closing unexpected callback listener: %v", closeErr)
			}
		}
		t.Fatal("expected an error when the callback port is already occupied")
	}
	if got := clierrors.KindOf(err); got != clierrors.KindEnvironment {
		t.Fatalf("error kind = %v, want %v", got, clierrors.KindEnvironment)
	}
}

func startOAuthCallbackTestServer(t *testing.T, status int) int {
	t.Helper()

	server, listener, err := newOAuthCallbackServer("127.0.0.1:0", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	if err != nil {
		t.Fatal(err)
	}

	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- server.Serve(listener)
	}()
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("closing callback server: %v", err)
		}
		if err := <-serveErrors; !stderrors.Is(err, http.ErrServerClosed) {
			t.Errorf("callback server returned unexpected error: %v", err)
		}
	})

	response, err := http.Get("http://" + listener.Addr().String() + "/auth/callback")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("closing callback response body: %v", err)
		}
	}()
	return response.StatusCode
}

func TestLoginHonorsConfigFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/keycloak/config" {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"enabled": false}`)); err != nil {
				t.Errorf("writing Keycloak config response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	customConfigPath := filepath.Join(t.TempDir(), "custom-config.yaml")

	out, err := executeCLIForTest(t, "login", server.URL, "--config", customConfigPath)
	require.NoError(t, err, "output: %s", out)
	assert.Contains(t, out, "No login required...")
	assert.Contains(t, out, "Context '"+server.URL+"' updated")

	// Verify the custom configuration file was created and populated
	cfg, err := config.ReadLocalConfig(customConfigPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, server.URL, cfg.CurrentContext)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, server.URL, cfg.Contexts[0].Name)
	assert.Equal(t, server.URL, cfg.Contexts[0].Server)

	require.Len(t, cfg.Servers, 1)
	assert.Equal(t, server.URL, cfg.Servers[0].Server)
	assert.False(t, cfg.Servers[0].KeycloakEnable)

	require.Len(t, cfg.Users, 1)
	assert.Equal(t, server.URL, cfg.Users[0].Name)
}

func TestLoginHonorsConfigFlagWithCustomContextName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/keycloak/config" {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"enabled": false}`)); err != nil {
				t.Errorf("writing Keycloak config response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	customConfigPath := filepath.Join(t.TempDir(), "custom-config.yaml")

	out, err := executeCLIForTest(t, "login", server.URL, "--name", "my-env", "--config", customConfigPath)
	require.NoError(t, err, "output: %s", out)
	assert.Contains(t, out, "Context 'my-env' updated")

	cfg, err := config.ReadLocalConfig(customConfigPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "my-env", cfg.CurrentContext)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "my-env", cfg.Contexts[0].Name)
	assert.Equal(t, server.URL, cfg.Contexts[0].Server)
}

func TestLoginPreservesExistingContextsInCustomConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/keycloak/config" {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"enabled": false}`)); err != nil {
				t.Errorf("writing Keycloak config response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	customConfigPath := filepath.Join(t.TempDir(), "custom-config.yaml")

	initialCfg := config.LocalConfig{
		CurrentContext: "existing-env",
		Contexts: []config.ContextRef{
			{Name: "existing-env", Server: "http://existing:8080", User: "existing-user"},
		},
		Servers: []config.Server{
			{Server: "http://existing:8080"},
		},
		Users: []config.User{
			{Name: "existing-user"},
		},
	}
	require.NoError(t, config.WriteLocalConfig(initialCfg, customConfigPath))

	out, err := executeCLIForTest(t, "login", server.URL, "--name", "new-env", "--config", customConfigPath)
	require.NoError(t, err, "output: %s", out)
	assert.Contains(t, out, "Context 'new-env' updated")

	cfg, err := config.ReadLocalConfig(customConfigPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "new-env", cfg.CurrentContext)
	require.Len(t, cfg.Contexts, 2)
	assert.Equal(t, "existing-env", cfg.Contexts[0].Name)
	assert.Equal(t, "new-env", cfg.Contexts[1].Name)
	require.Len(t, cfg.Servers, 2)
	require.Len(t, cfg.Users, 2)
}
