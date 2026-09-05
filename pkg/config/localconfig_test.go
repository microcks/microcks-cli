package config

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGetUserReturnsReferenceToStoredEntry(t *testing.T) {
    cfg := LocalConfig{
        Users: []User{
            {Name: "http://localhost:8080", AuthToken: "old-token"},
        },
    }

    user, err := cfg.GetUser("http://localhost:8080")
    require.NoError(t, err)

    user.AuthToken = "new-token"
    assert.Equal(t, "new-token", cfg.Users[0].AuthToken)
}

func TestGetServerReturnsReferenceToStoredEntry(t *testing.T) {
    cfg := LocalConfig{
        Servers: []Server{
            {Server: "http://localhost:8080", KeycloakEnable: true},
        },
    }

    server, err := cfg.GetServer("http://localhost:8080")
    require.NoError(t, err)

    server.KeycloakEnable = false
    assert.False(t, cfg.Servers[0].KeycloakEnable)
}

func TestGetInstanceReturnsReferenceToStoredEntry(t *testing.T) {
    cfg := LocalConfig{
        Instances: []Instance{
            {Name: "microcks", Status: "Exited"},
        },
    }

    instance, err := cfg.GetInstance("microcks")
    require.NoError(t, err)

    instance.Status = "Running"
    assert.Equal(t, "Running", cfg.Instances[0].Status)
}

func TestGetAuthReturnsReferenceToStoredEntry(t *testing.T) {
    cfg := LocalConfig{
        Auths: []Auth{
            {Server: "http://localhost:8080", ClientId: "id-a"},
        },
    }

    auth, err := cfg.GetAuth("http://localhost:8080")
    require.NoError(t, err)

    auth.ClientId = "id-b"
    assert.Equal(t, "id-b", cfg.Auths[0].ClientId)
}
