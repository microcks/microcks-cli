package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Instance.Name != "microcks" {
		t.Errorf("Expected default name 'microcks', got %s", cfg.Instance.Name)
	}
	if cfg.Instance.Driver != "docker" {
		t.Errorf("Expected default driver 'docker', got %s", cfg.Instance.Driver)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test-config.yaml")

	original := defaultConfig()
	original.Instance.Image = "test-image"

	err := SaveConfig(cfgPath, original)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Instance.Image != "test-image" {
		t.Errorf("Expected image 'test-image', got %s", loaded.Instance.Image)
	}
}

func TestEnsureConfigCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "new-config.yaml")

	cfg, err := EnsureConfig(cfgPath)
	if err != nil {
		t.Fatalf("EnsureConfig failed: %v", err)
	}

	if cfg.Instance.Name != "microcks" {
		t.Errorf("Expected default name 'microcks', got %s", cfg.Instance.Name)
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", cfgPath)
	}
}
