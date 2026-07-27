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
	"path/filepath"
	"runtime"
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
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific path test")
	}

	// Clear MICROCKS_CONFIG_DIR to ensure home dir logic is used
	t.Setenv("MICROCKS_CONFIG_DIR", "")

	// Set both HOME and USERPROFILE to mock a Windows-style home directory
	windowsHome := `C:\Users\JohnDoe`
	t.Setenv("HOME", windowsHome)
	t.Setenv("USERPROFILE", windowsHome)

	dir, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDir := `C:\Users\JohnDoe\.config\microcks`
	if dir != expectedDir {
		t.Errorf("expected config dir to be %q, got %q", expectedDir, dir)
	}

	// Verify DefaultLocalConfigPath
	configPath, err := DefaultLocalConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedConfigPath := `C:\Users\JohnDoe\.config\microcks\config`
	if configPath != expectedConfigPath {
		t.Errorf("expected config path to be %q, got %q", expectedConfigPath, configPath)
	}

	// Verify DefaultLocalWatchPath
	watchPath, err := DefaultLocalWatchPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedWatchPath := `C:\Users\JohnDoe\.config\microcks\watch`
	if watchPath != expectedWatchPath {
		t.Errorf("expected watch path to be %q, got %q", expectedWatchPath, watchPath)
	}
}

func TestDefaultConfigDir_WithUnixHomeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific path test")
	}

	// Clear MICROCKS_CONFIG_DIR to ensure home dir logic is used
	t.Setenv("MICROCKS_CONFIG_DIR", "")

	// Set HOME to mock a Unix-style home directory
	unixHome := "/home/johndoe"
	t.Setenv("HOME", unixHome)

	dir, err := DefaultConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDir := "/home/johndoe/.config/microcks"
	if dir != expectedDir {
		t.Errorf("expected config dir to be %q, got %q", expectedDir, dir)
	}

	// Verify DefaultLocalConfigPath
	configPath, err := DefaultLocalConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedConfigPath := "/home/johndoe/.config/microcks/config"
	if configPath != expectedConfigPath {
		t.Errorf("expected config path to be %q, got %q", expectedConfigPath, configPath)
	}

	// Verify DefaultLocalWatchPath
	watchPath, err := DefaultLocalWatchPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedWatchPath := "/home/johndoe/.config/microcks/watch"
	if watchPath != expectedWatchPath {
		t.Errorf("expected watch path to be %q, got %q", expectedWatchPath, watchPath)
	}
}
