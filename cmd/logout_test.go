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
	"path/filepath"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/stretchr/testify/require"
)

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
