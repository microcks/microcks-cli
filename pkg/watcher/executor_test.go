package watcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
)

func TestTriggerImport_NewClientErrorNoPanic(t *testing.T) {
	// Create a temp directory for the config dir.
	tmpDir := t.TempDir()

	// Set MICROCKS_CONFIG_DIR env var to point to our temp directory
	os.Setenv("MICROCKS_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("MICROCKS_CONFIG_DIR")

	// We create an invalid config file that fails validation or context resolution.
	// This will cause connectors.NewClient to fail when resolving config context.
	configPath := filepath.Join(tmpDir, "config")
	invalidYAML := []byte(`
currentContext: non-existent-context
contexts:
  - name: invalid
`)
	if err := os.WriteFile(configPath, invalidYAML, 0600); err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	// TriggerImport should run and log the failure but not panic.
	entry := config.WatchEntry{
		FilePath:     "dummy.yaml",
		MainArtifact: true,
		Context:      []string{"non-existent-context"},
	}

	// Wrap in recover to assert no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TriggerImport panicked: %v", r)
		}
	}()

	TriggerImport(entry)
}

func TestTriggerImport_NoConfigFile_NewMicrocksClientErrorNoPanic(t *testing.T) {
	// Create a temp directory for the config dir to ensure config doesn't exist.
	tmpDir := t.TempDir()

	// Set MICROCKS_CONFIG_DIR env var to point to our empty temp directory.
	os.Setenv("MICROCKS_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("MICROCKS_CONFIG_DIR")

	// Since there is no config file in the temp directory, TriggerImport will fall back
	// to creating a headless client via connectors.NewMicrocksClient(context).
	// We pass a context containing a control character, which causes url.Parse to fail,
	// ensuring NewMicrocksClient returns an error and does not panic.
	entry := config.WatchEntry{
		FilePath:     "dummy.yaml",
		MainArtifact: true,
		Context:      []string{"http://\x7f.com"},
	}

	// Wrap in recover to assert no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TriggerImport panicked: %v", r)
		}
	}()

	TriggerImport(entry)
}
