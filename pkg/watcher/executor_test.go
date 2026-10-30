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

package watcher

import (
	"os"
	"path"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
)

// TestTriggerImportSkipsContextWhenClientFails ensures TriggerImport does not
// panic when the Microcks client cannot be created for a context (for example
// an unknown or unresolvable context). Before the fix it dereferenced a nil
// client and crashed the watcher process.
func TestTriggerImportSkipsContextWhenClientFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Write an empty local config so that resolving any context fails and
	// NewClient returns a nil client with an error.
	cfgDir := path.Join(home, ".config", "microcks")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(path.Join(cfgDir, "config"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	entry := config.WatchEntry{
		FilePath:     path.Join(home, "spec.json"),
		Context:      []string{"missing-context"},
		MainArtifact: true,
	}

	// Should return cleanly instead of panicking on a nil client.
	TriggerImport(entry)
}
