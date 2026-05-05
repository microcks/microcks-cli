package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestAuthYAMLTags(t *testing.T) {
	cfg := LocalConfig{
		Auths: []Auth{
			{
				Server:       "local",
				ClientId:     "microcks-serviceaccount",
				ClientSecret: "secret",
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.Contains(t, string(data), "server: local")
	require.Contains(t, string(data), "clientId: microcks-serviceaccount")
	require.Contains(t, string(data), "clientSecret: secret")

	var parsed LocalConfig
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)
	require.Equal(t, cfg.Auths, parsed.Auths)
}

func TestAuthYAMLAcceptsLegacyKeys(t *testing.T) {
	data := []byte(`auths:
- server: local
  clientid: microcks-serviceaccount
  clientsecret: secret
`)

	var parsed LocalConfig
	err := yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)
	require.Equal(t, []Auth{
		{
			Server:       "local",
			ClientId:     "microcks-serviceaccount",
			ClientSecret: "secret",
		},
	}, parsed.Auths)
}
