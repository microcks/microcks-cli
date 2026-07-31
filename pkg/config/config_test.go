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
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

func TestGetFilePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		validModes := []os.FileMode{0666, 0444}
		for _, m := range validModes {
			fi := mockFileInfo{mode: m}
			err := getFilePermission(fi)
			assert.NoError(t, err, "Mode %v should be valid on Windows", m)
		}

		invalidModes := []os.FileMode{0777, 0600, 0400}
		for _, m := range invalidModes {
			fi := mockFileInfo{mode: m}
			err := getFilePermission(fi)
			assert.Error(t, err, "Mode %v should be invalid on Windows", m)
		}
	} else {
		validModes := []os.FileMode{0600, 0400}
		for _, m := range validModes {
			fi := mockFileInfo{mode: m}
			err := getFilePermission(fi)
			assert.NoError(t, err, "Mode %v should be valid on UNIX", m)
		}

		invalidModes := []os.FileMode{0777, 0666, 0444}
		for _, m := range invalidModes {
			fi := mockFileInfo{mode: m}
			err := getFilePermission(fi)
			assert.Error(t, err, "Mode %v should be invalid on UNIX", m)
		}
	}
}

func TestRedactSensitiveContent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Authorization: Bearer my-secret-token-123\n",
			expected: "Authorization: [REDACTED]\n",
		},
		{
			input:    "Authorization: Basic dXNlcjpwYXNz\n",
			expected: "Authorization: [REDACTED]\n",
		},
		{
			input:    "http://example.com?access_token=foo&refresh_token=bar&id_token=baz&code=qux",
			expected: "http://example.com?access_token=[REDACTED]&refresh_token=[REDACTED]&id_token=[REDACTED]&code=[REDACTED]",
		},
		{
			input:    "No sensitive data here",
			expected: "No sensitive data here",
		},
	}

	for _, tc := range tests {
		got := redactSensitiveContent(tc.input)
		assert.Equal(t, tc.expected, got)
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDumpRequestAndResponseIfRequired(t *testing.T) {
	oldVerbose := Verbose
	defer func() { Verbose = oldVerbose }()

	req, _ := http.NewRequest("GET", "http://example.com?access_token=some_token", nil)
	req.Header.Set("Authorization", "Bearer sensitive")

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"Authorization": []string{"Bearer secret-resp"}},
		Body:       io.NopCloser(strings.NewReader("response body")),
	}

	// 1. Verbose = false
	Verbose = false
	outputReq := captureStdout(func() {
		DumpRequestIfRequired("test-req", req, false)
	})
	assert.Empty(t, outputReq)

	outputResp := captureStdout(func() {
		DumpResponseIfRequired("test-resp", resp, false)
	})
	assert.Empty(t, outputResp)

	// 2. Verbose = true
	Verbose = true
	outputReq = captureStdout(func() {
		DumpRequestIfRequired("test-req", req, false)
	})
	assert.Contains(t, outputReq, "Dumping request 'test-req'")
	assert.Contains(t, outputReq, "Authorization: [REDACTED]")
	assert.Contains(t, outputReq, "access_token=[REDACTED]")
	assert.NotContains(t, outputReq, "sensitive")

	outputResp = captureStdout(func() {
		DumpResponseIfRequired("test-resp", resp, false)
	})
	assert.Contains(t, outputResp, "Dumping response 'test-resp'")
	assert.Contains(t, outputResp, "Authorization: [REDACTED]")
	assert.NotContains(t, outputResp, "secret-resp")
}

func TestCreateTLSConfig(t *testing.T) {
	oldInsecure := InsecureTLS
	oldCaCertPaths := CaCertPaths
	defer func() {
		InsecureTLS = oldInsecure
		CaCertPaths = oldCaCertPaths
	}()

	// 1. Defaults
	InsecureTLS = false
	CaCertPaths = ""
	cfg := CreateTLSConfig()
	assert.NotNil(t, cfg)
	assert.False(t, cfg.InsecureSkipVerify)
	assert.Nil(t, cfg.RootCAs)

	// 2. Insecure TLS
	InsecureTLS = true
	cfg = CreateTLSConfig()
	assert.True(t, cfg.InsecureSkipVerify)

	// 3. CA Cert Paths
	InsecureTLS = false
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "dummy.crt")

	dummyCertPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBtzCCAV2gAwIBAgIJAJ1V7U8W5B5tMA0GCSqGSIb3DQEBCwUAMBAxDjAMBgNV
BAMMBXRlc3RjYTAeFw0yNjA4MDEwMDAwMDBaFw0zNjA4MDEwMDAwMDBaMBAxDjAM
BgNVBAMMBXRlc3RjYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC1jF
-----END CERTIFICATE-----`)

	err := os.WriteFile(certFile, dummyCertPEM, 0o600)
	require.NoError(t, err)

	CaCertPaths = certFile
	cfg = CreateTLSConfig()
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.RootCAs)
}

func TestDefaultPaths(t *testing.T) {
	oldEnv := os.Getenv("MICROCKS_CONFIG_DIR")
	defer os.Setenv("MICROCKS_CONFIG_DIR", oldEnv)

	// 1. Env variable set
	os.Setenv("MICROCKS_CONFIG_DIR", "/custom/path")
	dir, err := DefaultConfigDir()
	assert.NoError(t, err)
	assert.Equal(t, "/custom/path", dir)

	cfgPath, err := DefaultLocalConfigPath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join("/custom/path", "config"), cfgPath)

	watchPath, err := DefaultLocalWatchPath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join("/custom/path", "watch"), watchPath)

	// 2. Env variable unset - should fall back to user home dir
	os.Unsetenv("MICROCKS_CONFIG_DIR")
	dir, err = DefaultConfigDir()
	assert.NoError(t, err)
	assert.Contains(t, dir, ".config")
	assert.Contains(t, dir, "microcks")
}

func TestLocalConfigCRUDAndValidation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "local.config")

	// 1. Read non-existent config
	cfg, err := ReadLocalConfig(configPath)
	assert.NoError(t, err)
	assert.Nil(t, cfg)

	// 2. Write a minimal valid config
	initialCfg := LocalConfig{
		CurrentContext: "ctx1",
		Contexts: []ContextRef{
			{Name: "ctx1", Server: "https://microcks.example.com", User: "usr1", Instance: "inst1"},
		},
		Servers: []Server{
			{Name: "srv1", Server: "https://microcks.example.com", InsecureTLS: true},
		},
		Users: []User{
			{Name: "usr1", AuthToken: "token123", RefreshToken: "refresh123"},
		},
		Instances: []Instance{
			{Name: "inst1", ContainerID: "cont123"},
		},
		Auths: []Auth{
			{Server: "https://microcks.example.com", ClientId: "client123", ClientSecret: "secret123"},
		},
	}

	err = WriteLocalConfig(initialCfg, configPath)
	assert.NoError(t, err)

	// 3. Read it back
	loadedCfg, err := ReadLocalConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedCfg)
	assert.Equal(t, "ctx1", loadedCfg.CurrentContext)

	// 4. Resolve Context
	// 4a. Resolve current (empty argument)
	ctx, err := loadedCfg.ResolveContext("")
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
	assert.Equal(t, "ctx1", ctx.Name)
	assert.Equal(t, "srv1", ctx.Server.Name)
	assert.Equal(t, "usr1", ctx.User.Name)
	assert.Equal(t, "inst1", ctx.Instance.Name)

	// 4b. Resolve by name
	ctx, err = loadedCfg.ResolveContext("ctx1")
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// 4c. Resolve non-existent context
	_, err = loadedCfg.ResolveContext("non-existent")
	assert.Error(t, err)

	// 5. Validation failure: Set current-context to something undefined
	loadedCfg.CurrentContext = "invalid-ctx"
	err = ValidateLocalConfig(*loadedCfg)
	assert.Error(t, err)

	// 6. Test Get/Upsert/Remove Context
	loadedCfg.CurrentContext = "ctx1"
	loadedCfg.UpsertContext(ContextRef{Name: "ctx2", Server: "https://microcks.example.com", User: "usr1"})
	ctx2, err := loadedCfg.ResolveContext("ctx2")
	assert.NoError(t, err)
	assert.Equal(t, "ctx2", ctx2.Name)

	srvName, ok := loadedCfg.RemoveContext("ctx2")
	assert.True(t, ok)
	assert.Equal(t, "https://microcks.example.com", srvName)

	_, ok = loadedCfg.RemoveContext("ctx2")
	assert.False(t, ok)

	// 7. Test Get/Upsert/Remove User
	u, err := loadedCfg.GetUser("usr1")
	assert.NoError(t, err)
	assert.Equal(t, "token123", u.AuthToken)

	loadedCfg.UpsertUser(User{Name: "usr2", AuthToken: "token456"})
	u2, err := loadedCfg.GetUser("usr2")
	assert.NoError(t, err)
	assert.Equal(t, "token456", u2.AuthToken)

	ok = loadedCfg.RemoveUser("usr2")
	assert.True(t, ok)
	_, err = loadedCfg.GetUser("usr2")
	assert.Error(t, err)

	// Token removal
	ok = loadedCfg.RemoveToken("usr1")
	assert.True(t, ok)
	u1Mod, _ := loadedCfg.GetUser("usr1")
	assert.Empty(t, u1Mod.AuthToken)
	assert.Empty(t, u1Mod.RefreshToken)

	// 8. Test Get/Upsert/Remove Server
	s, err := loadedCfg.GetServer("https://microcks.example.com")
	assert.NoError(t, err)
	assert.True(t, s.InsecureTLS)

	loadedCfg.UpsertServer(Server{Name: "srv2", Server: "https://microcks2.example.com"})
	s2, err := loadedCfg.GetServer("https://microcks2.example.com")
	assert.NoError(t, err)
	assert.Equal(t, "srv2", s2.Name)

	ok = loadedCfg.RemoveServer("https://microcks2.example.com")
	assert.True(t, ok)
	_, err = loadedCfg.GetServer("https://microcks2.example.com")
	assert.Error(t, err)

	// 9. Test Get/Upsert/Remove Instance
	inst, err := loadedCfg.GetInstance("inst1")
	assert.NoError(t, err)
	assert.Equal(t, "cont123", inst.ContainerID)

	loadedCfg.UpsertInstance(Instance{Name: "inst2", ContainerID: "cont456"})
	inst2, err := loadedCfg.GetInstance("inst2")
	assert.NoError(t, err)
	assert.Equal(t, "cont456", inst2.ContainerID)

	loadedCfg.UpsertInstance(Instance{Name: "inst2-updated", ContainerID: "cont456"})
	inst2Updated, err := loadedCfg.GetInstance("inst2-updated")
	assert.NoError(t, err)
	assert.Equal(t, "cont456", inst2Updated.ContainerID)

	ok = loadedCfg.RemoveInstance("inst2-updated")
	assert.True(t, ok)

	ok = loadedCfg.RemoveInstance("")
	assert.True(t, ok)

	// 10. Test Get/Upsert/Remove Auth
	a, err := loadedCfg.GetAuth("https://microcks.example.com")
	assert.NoError(t, err)
	assert.Equal(t, "client123", a.ClientId)

	loadedCfg.UpsertAuth(Auth{Server: "https://microcks2.example.com", ClientId: "client456"})
	a2, err := loadedCfg.GetAuth("https://microcks2.example.com")
	assert.NoError(t, err)
	assert.Equal(t, "client456", a2.ClientId)

	ok = loadedCfg.RemoveAuth("https://microcks2.example.com")
	assert.True(t, ok)
	_, err = loadedCfg.GetAuth("https://microcks2.example.com")
	assert.Error(t, err)

	assert.False(t, loadedCfg.IsEmpty())
	emptyCfg := LocalConfig{}
	assert.True(t, emptyCfg.IsEmpty())

	// 11. Delete Local Config
	err = loadedCfg.DeleteLocalConfig(configPath)
	assert.NoError(t, err)
	_, err = os.Stat(configPath)
	assert.True(t, os.IsNotExist(err))
}

func TestReadLocalConfigPermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping OS file permission failure test on Windows since files always have 0666 or 0444 permissions")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "local.config")

	err := os.WriteFile(configPath, []byte("current-context: ctx1"), 0o777)
	require.NoError(t, err)

	_, err = ReadLocalConfig(configPath)
	assert.Error(t, err, "Should fail due to incorrect file permission on UNIX")
}

func TestWatchConfig(t *testing.T) {
	tmpDir := t.TempDir()
	watchPath := filepath.Join(tmpDir, "watch")

	// 1. Read non-existent watch config
	wCfg, err := ReadLocalWatchConfig(watchPath)
	assert.NoError(t, err)
	assert.Nil(t, wCfg)

	// 2. Create watch config
	initialWCfg := WatchConfig{
		Entries: []WatchEntry{
			{FilePath: "file1.yaml", Context: []string{"ctx1"}, MainArtifact: true},
		},
	}

	err = WriteLocalWatchConfig(initialWCfg, watchPath)
	assert.NoError(t, err)

	// 3. Read it back
	loadedWCfg, err := ReadLocalWatchConfig(watchPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedWCfg)
	assert.Len(t, loadedWCfg.Entries, 1)
	assert.Equal(t, "file1.yaml", loadedWCfg.Entries[0].FilePath)

	// 4. Upsert entry (new path)
	loadedWCfg.UpsertEntry(WatchEntry{FilePath: "file2.yaml", Context: []string{"ctx2"}, MainArtifact: false})
	assert.Len(t, loadedWCfg.Entries, 2)

	// 5. Upsert entry (existing path, append context)
	loadedWCfg.UpsertEntry(WatchEntry{FilePath: "file1.yaml", Context: []string{"ctx3"}, MainArtifact: true})
	assert.Len(t, loadedWCfg.Entries, 2)
	for _, e := range loadedWCfg.Entries {
		if e.FilePath == "file1.yaml" {
			assert.Contains(t, e.Context, "ctx1")
			assert.Contains(t, e.Context, "ctx3")
		}
	}
}
