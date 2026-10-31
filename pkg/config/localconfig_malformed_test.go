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

package config

import (
	"os"
	"path/filepath"
	"testing"

	clierrors "github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadLocalConfigMalformedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	want := []byte("current-context: [broken\n")
	require.NoError(t, os.WriteFile(path, want, 0o600))

	config, err := ReadLocalConfig(path)

	assert.Nil(t, config)
	require.Error(t, err)
	assert.Equal(t, clierrors.KindEnvironment, clierrors.KindOf(err))
	assert.Contains(t, err.Error(), "reading local config")
	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, want, got)
}

func TestReadLocalWatchConfigMalformedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "watch")
	want := []byte("entries: [broken\n")
	require.NoError(t, os.WriteFile(path, want, 0o600))

	config, err := ReadLocalWatchConfig(path)

	assert.Nil(t, config)
	require.Error(t, err)
	assert.Equal(t, clierrors.KindEnvironment, clierrors.KindOf(err))
	assert.Contains(t, err.Error(), "reading local watch config")
	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, want, got)
}
