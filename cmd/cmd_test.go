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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProbeCommand(t *testing.T, onRun func(cmd *cobra.Command)) *cobra.Command {
	t.Helper()
	root, err := NewCommand()
	require.NoError(t, err)
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, args []string) error {
			onRun(cmd)
			return nil
		},
	})
	return root
}

func TestGlobalFlagsFallbackToEnvironmentVariables(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("MICROCKS_CONFIG", cfgPath)
	t.Setenv("MICROCKS_CONTEXT", "staging")
	t.Setenv("MICROCKS_VERBOSE", "true")
	t.Setenv("MICROCKS_INSECURE_TLS", "true")
	t.Setenv("MICROCKS_CA_CERTS", "/tmp/ca1.crt,/tmp/ca2.crt")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_ID", "svc-client")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_SECRET", "svc-secret")
	t.Setenv("MICROCKS_URL", "https://microcks.example.com")

	var gotConfig, gotContext, gotCaCerts, gotClientID, gotClientSecret, gotURL string
	var gotVerbose, gotInsecureTLS bool

	root := newProbeCommand(t, func(cmd *cobra.Command) {
		var err error
		gotConfig, err = cmd.Flags().GetString("config")
		require.NoError(t, err)
		gotContext, err = cmd.Flags().GetString("microcks-context")
		require.NoError(t, err)
		gotVerbose, err = cmd.Flags().GetBool("verbose")
		require.NoError(t, err)
		gotInsecureTLS, err = cmd.Flags().GetBool("insecure-tls")
		require.NoError(t, err)
		gotCaCerts, err = cmd.Flags().GetString("caCerts")
		require.NoError(t, err)
		gotClientID, err = cmd.Flags().GetString("keycloakClientId")
		require.NoError(t, err)
		gotClientSecret, err = cmd.Flags().GetString("keycloakClientSecret")
		require.NoError(t, err)
		gotURL, err = cmd.Flags().GetString("microcksURL")
		require.NoError(t, err)
	})
	root.SetArgs([]string{"probe"})
	require.NoError(t, root.Execute())

	assert.Equal(t, cfgPath, gotConfig)
	assert.Equal(t, "staging", gotContext)
	assert.True(t, gotVerbose)
	assert.True(t, gotInsecureTLS)
	assert.Equal(t, "/tmp/ca1.crt,/tmp/ca2.crt", gotCaCerts)
	assert.Equal(t, "svc-client", gotClientID)
	assert.Equal(t, "svc-secret", gotClientSecret)
	assert.Equal(t, "https://microcks.example.com", gotURL)
}

func TestCLIFlagsOverrideEnvironmentVariables(t *testing.T) {
	t.Setenv("MICROCKS_URL", "https://env.example.com")
	t.Setenv("MICROCKS_CONTEXT", "env-ctx")
	t.Setenv("MICROCKS_INSECURE_TLS", "true")
	t.Setenv("MICROCKS_VERBOSE", "true")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_ID", "env-id")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_SECRET", "env-secret")

	var gotURL, gotContext, gotClientID, gotClientSecret string
	var gotInsecureTLS, gotVerbose bool

	root := newProbeCommand(t, func(cmd *cobra.Command) {
		var err error
		gotURL, err = cmd.Flags().GetString("microcksURL")
		require.NoError(t, err)
		gotContext, err = cmd.Flags().GetString("microcks-context")
		require.NoError(t, err)
		gotInsecureTLS, err = cmd.Flags().GetBool("insecure-tls")
		require.NoError(t, err)
		gotVerbose, err = cmd.Flags().GetBool("verbose")
		require.NoError(t, err)
		gotClientID, err = cmd.Flags().GetString("keycloakClientId")
		require.NoError(t, err)
		gotClientSecret, err = cmd.Flags().GetString("keycloakClientSecret")
		require.NoError(t, err)
	})
	root.SetArgs([]string{
		"probe",
		"--microcksURL", "https://cli.example.com",
		"--microcks-context", "cli-ctx",
		"--insecure-tls=false",
		"--verbose=false",
		"--keycloakClientId", "cli-id",
	})
	require.NoError(t, root.Execute())

	assert.Equal(t, "https://cli.example.com", gotURL)
	assert.Equal(t, "cli-ctx", gotContext)
	assert.False(t, gotInsecureTLS)
	assert.False(t, gotVerbose)
	assert.Equal(t, "cli-id", gotClientID)
	assert.Equal(t, "env-secret", gotClientSecret)
}

func TestEmptyEnvVarsAreIgnored(t *testing.T) {
	t.Setenv("MICROCKS_URL", "")
	t.Setenv("MICROCKS_INSECURE_TLS", "")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_ID", "")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_SECRET", "")

	var gotURL string
	var gotInsecureTLS bool

	root := newProbeCommand(t, func(cmd *cobra.Command) {
		var err error
		gotURL, err = cmd.Flags().GetString("microcksURL")
		require.NoError(t, err)
		gotInsecureTLS, err = cmd.Flags().GetBool("insecure-tls")
		require.NoError(t, err)
	})
	root.SetArgs([]string{"probe"})
	require.NoError(t, root.Execute())

	assert.Empty(t, gotURL)
	assert.False(t, gotInsecureTLS)
}

func TestInvalidBooleanEnvVarReturnsUsageError(t *testing.T) {
	for _, tc := range []struct {
		name    string
		envName string
	}{
		{name: "insecure-tls", envName: "MICROCKS_INSECURE_TLS"},
		{name: "verbose", envName: "MICROCKS_VERBOSE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.envName, "not-a-bool")
			root := newProbeCommand(t, func(cmd *cobra.Command) {
				t.Fatal("command RunE should not be reached on invalid boolean env var")
			})
			root.SetArgs([]string{"probe"})
			err := root.Execute()
			require.Error(t, err)
			assert.Equal(t, errors.KindUsage, errors.KindOf(err))
			assert.Contains(t, err.Error(), tc.envName)
		})
	}
}

func TestHalfSetKeycloakEnvVarsFailWithUsageError(t *testing.T) {
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_ID", "only-id")

	root := newProbeCommand(t, func(cmd *cobra.Command) {
		t.Fatal("command RunE should not be reached when keycloakClientSecret is missing")
	})
	root.SetArgs([]string{"probe"})
	err := root.Execute()
	require.Error(t, err)
	assert.Equal(t, errors.KindUsage, errors.KindOf(err))
	assert.Contains(t, err.Error(), "missing [keycloakClientSecret]")
}

func TestServiceListUsesEnvURLAndKeycloakCredentials(t *testing.T) {
	var gotAuthHeader string
	var serverURL string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/keycloak/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"enabled":         true,
			"realm":           "microcks",
			"auth-server-url": serverURL + "/auth",
		})
		require.NoError(t, err)
	})
	mux.HandleFunc("/auth/realms/microcks/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "ci-client", user)
		assert.Equal(t, "ci-secret", pass)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]string{
			"access_token": "env-token-123",
		})
		require.NoError(t, err)
	})
	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode([]connectors.Service{
			{ID: "svc-1", Name: "Petstore", Version: "1.0.0", Type: "REST"},
		})
		require.NoError(t, err)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	t.Setenv("MICROCKS_URL", server.URL)
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_ID", "ci-client")
	t.Setenv("MICROCKS_KEYCLOAK_CLIENT_SECRET", "ci-secret")

	out, err := executeCLIForTest(t, "service", "list", "--output", "json")
	require.NoError(t, err)
	assert.Equal(t, "Bearer env-token-123", gotAuthHeader)

	var services []connectors.Service
	require.NoError(t, json.Unmarshal([]byte(out), &services))
	require.Len(t, services, 1)
	assert.Equal(t, "Petstore", services[0].Name)
}
