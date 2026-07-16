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
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveStopTargetByContextName(t *testing.T) {
	cfg := &config.LocalConfig{
		Contexts: []config.ContextRef{
			{Name: "dev", Server: "http://dev-server:8080", User: "u1"},
		},
		Servers: []config.Server{
			{Server: "http://dev-server:8080"},
		},
		Users: []config.User{
			{Name: "u1"},
		},
	}

	ctx, err := resolveStopTarget("dev", cfg)
	require.NoError(t, err)
	assert.Equal(t, "dev", ctx.Name)
}

func TestResolveStopTargetByInstanceName(t *testing.T) {
	cfg := &config.LocalConfig{
		Contexts: []config.ContextRef{
			{Name: "http://server", Server: "http://instance-server:8080", User: "u1", Instance: "myinst"},
		},
		Servers: []config.Server{
			{Server: "http://instance-server:8080"},
		},
		Users: []config.User{
			{Name: "u1"},
		},
	}

	ctx, err := resolveStopTarget("myinst", cfg)
	require.NoError(t, err)
	assert.Equal(t, "http://server", ctx.Name)
}

func TestResolveStopTargetNotFound(t *testing.T) {
	cfg := &config.LocalConfig{
		Contexts: []config.ContextRef{
			{Name: "dev", Server: "http://dev-server:8080", User: "u1"},
		},
		Servers: []config.Server{
			{Server: "http://dev-server:8080"},
		},
		Users: []config.User{
			{Name: "u1"},
		},
	}

	_, err := resolveStopTarget("nonexistent", cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestResolveStopTargetCollisionContextWins(t *testing.T) {
	cfg := &config.LocalConfig{
		Contexts: []config.ContextRef{
			{Name: "shared", Server: "http://server-a:8080", User: "u1", Instance: "other"},
			{Name: "different", Server: "http://server-b:8080", User: "u2", Instance: "shared"},
		},
		Servers: []config.Server{
			{Server: "http://server-a:8080"},
			{Server: "http://server-b:8080"},
		},
		Users: []config.User{
			{Name: "u1"},
			{Name: "u2"},
		},
	}

	ctx, err := resolveStopTarget("shared", cfg)
	require.NoError(t, err)
	assert.Equal(t, "shared", ctx.Name)
	assert.Equal(t, "http://server-a:8080", ctx.Server.Server)
}

func TestResolveStopTargetEmptyName(t *testing.T) {
	cfg := &config.LocalConfig{
		CurrentContext: "dev",
		Contexts: []config.ContextRef{
			{Name: "dev", Server: "http://dev-server:8080", User: "u1"},
		},
		Servers: []config.Server{
			{Server: "http://dev-server:8080"},
		},
		Users: []config.User{
			{Name: "u1"},
		},
	}

	ctx, err := resolveStopTarget("", cfg)
	require.NoError(t, err)
	assert.Equal(t, "dev", ctx.Name)
}
