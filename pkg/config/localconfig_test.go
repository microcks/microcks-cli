package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfigDir_WithEnvVar(t *testing.T) {
	// Set the environment variable MICROCKS_CONFIG_DIR
	customDir := filepath.Join("C:", "Users", "JohnDoe", "custom-config")
	t.Setenv("MICROCKS_CONFIG_DIR", customDir)

	dir, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dir != customDir {
		t.Errorf("expected config dir to be %q, got %q", customDir, dir)
	}
}

func TestDefaultConfigDir_WithWindowsHomeDir(t *testing.T) {
	// Clear MICROCKS_CONFIG_DIR to ensure home dir logic is used
	t.Setenv("MICROCKS_CONFIG_DIR", "")

	// Set both HOME and USERPROFILE to mock a Windows-style home directory
	windowsHome := filepath.Join("C:", "Users", "JohnDoe")
	t.Setenv("HOME", windowsHome)
	t.Setenv("USERPROFILE", windowsHome)

	dir, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDir := filepath.Join(windowsHome, ".config", "microcks")
	if dir != expectedDir {
		t.Errorf("expected config dir to be %q, got %q", expectedDir, dir)
	}

	// Verify DefaultLocalConfigPath
	configPath, err := DefaultLocalConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedConfigPath := filepath.Join(expectedDir, "config")
	if configPath != expectedConfigPath {
		t.Errorf("expected config path to be %q, got %q", expectedConfigPath, configPath)
	}

	// Verify DefaultLocalWatchPath
	watchPath, err := DefaultLocalWatchPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedWatchPath := filepath.Join(expectedDir, "watch")
	if watchPath != expectedWatchPath {
		t.Errorf("expected watch path to be %q, got %q", expectedWatchPath, watchPath)
	}
}
