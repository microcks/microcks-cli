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
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func runOAuth2LoginWithError(t *testing.T) {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())

	oauth2conf := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://127.0.0.1/auth",
			TokenURL: "http://127.0.0.1/token",
		},
	}

	done := make(chan error, 1)
	go func() {
		_, _, err := oauth2login(context.Background(), port, oauth2conf, false)
		done <- err
	}()

	// Send an error callback so the login flow ends.
	callbackURL := fmt.Sprintf("http://localhost:%d/auth/callback?error=access_denied", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(callbackURL) // #nosec G107 -- local test server
		if err != nil {
			return false
		}
		resp.Body.Close()
		return true
	}, 10*time.Second, 100*time.Millisecond)

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("oauth2login did not return")
	}
}

func TestOAuth2LoginCanRunTwice(t *testing.T) {
	runOAuth2LoginWithError(t)
	// The second login must not panic on a duplicate callback registration.
	runOAuth2LoginWithError(t)
}

func TestLogoutContextResolvesNamedContextUser(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	server := "https://microcks.example"

	localCfg := config.LocalConfig{
		CurrentContext: "staging",
		Contexts: []config.ContextRef{
			{Name: "staging", Server: server, User: server},
		},
		Servers: []config.Server{
			{Server: server, KeycloakEnable: true},
		},
		Users: []config.User{
			{Name: server, AuthToken: "access-token", RefreshToken: "refresh-token"},
		},
	}
	require.NoError(t, config.WriteLocalConfig(localCfg, configPath))

	require.NoError(t, logoutContext("staging", configPath))

	updated, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, updated)

	user, err := updated.GetUser(server)
	require.NoError(t, err)
	require.Empty(t, user.AuthToken)
	require.Empty(t, user.RefreshToken)
}

func TestLogoutContextStillAcceptsStoredUserName(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	server := "https://microcks.example"

	localCfg := config.LocalConfig{
		CurrentContext: "staging",
		Contexts: []config.ContextRef{
			{Name: "staging", Server: server, User: server},
		},
		Servers: []config.Server{
			{Server: server, KeycloakEnable: true},
		},
		Users: []config.User{
			{Name: server, AuthToken: "access-token", RefreshToken: "refresh-token"},
		},
	}
	require.NoError(t, config.WriteLocalConfig(localCfg, configPath))

	require.NoError(t, logoutContext(server, configPath))

	updated, err := config.ReadLocalConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, updated)

	user, err := updated.GetUser(server)
	require.NoError(t, err)
	require.Empty(t, user.AuthToken)
	require.Empty(t, user.RefreshToken)
}
