package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalConfig_UpsertContext(t *testing.T) {
	cfg := &LocalConfig{}
	
	ctx1 := ContextRef{Name: "ctx1", Server: "server1", User: "user1"}
	cfg.UpsertContext(ctx1)
	assert.Len(t, cfg.Contexts, 1)
	assert.Equal(t, ctx1, cfg.Contexts[0])

	// Update existing context
	ctx1Updated := ContextRef{Name: "ctx1", Server: "server1-updated", User: "user1"}
	cfg.UpsertContext(ctx1Updated)
	assert.Len(t, cfg.Contexts, 1)
	assert.Equal(t, ctx1Updated, cfg.Contexts[0])

	// Add second context
	ctx2 := ContextRef{Name: "ctx2", Server: "server2", User: "user2"}
	cfg.UpsertContext(ctx2)
	assert.Len(t, cfg.Contexts, 2)
}

func TestLocalConfig_RemoveContext(t *testing.T) {
	cfg := &LocalConfig{
		Contexts: []ContextRef{
			{Name: "ctx1", Server: "server1"},
			{Name: "ctx2", Server: "server2"},
		},
	}

	server, removed := cfg.RemoveContext("ctx1")
	assert.True(t, removed)
	assert.Equal(t, "server1", server)
	assert.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "ctx2", cfg.Contexts[0].Name)

	server, removed = cfg.RemoveContext("non-existent")
	assert.False(t, removed)
	assert.Equal(t, "", server)
	assert.Len(t, cfg.Contexts, 1)
}

func TestLocalConfig_ResolveContext(t *testing.T) {
	cfg := &LocalConfig{
		CurrentContext: "ctx1",
		Contexts: []ContextRef{
			{Name: "ctx1", Server: "server1", User: "user1"},
		},
		Servers: []Server{
			{Name: "server1", Server: "server1"},
		},
		Users: []User{
			{Name: "user1", AuthToken: "token1"},
		},
	}

	// Resolve current context
	ctx, err := cfg.ResolveContext("")
	require.NoError(t, err)
	assert.Equal(t, "ctx1", ctx.Name)
	assert.Equal(t, "server1", ctx.Server.Server)
	assert.Equal(t, "token1", ctx.User.AuthToken)

	// Resolve specific context
	ctx, err = cfg.ResolveContext("ctx1")
	require.NoError(t, err)
	assert.Equal(t, "ctx1", ctx.Name)

	// Resolve non-existent context
	ctx, err = cfg.ResolveContext("non-existent")
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

func TestReadLocalConfig_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	
	configContent := `current-context: test
contexts:
- name: test
  server: server1
  user: user1
servers:
- name: server1
  server: server1
users:
- name: user1`

	// Create config with wrong permissions
	err := os.WriteFile(path, []byte(configContent), 0755)
	require.NoError(t, err)

	_, err = ReadLocalConfig(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect permission flags")

	// Fix permissions
	err = os.Chmod(path, 0600)
	require.NoError(t, err)

	cfg, err := ReadLocalConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "test", cfg.CurrentContext)
}
