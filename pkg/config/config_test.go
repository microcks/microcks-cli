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

func captureStdout(t testing.TB, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	f()

	require.NoError(t, w.Close())
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

func TestDumpRequestAndResponseIfRequired(t *testing.T) {
	oldVerbose := Verbose
	defer func() { Verbose = oldVerbose }()

	req, err := http.NewRequest("GET", "http://example.com?access_token=some_token", nil)
	require.NoError(t, err)
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
	outputReq := captureStdout(t, func() {
		DumpRequestIfRequired("test-req", req, false)
	})
	assert.Empty(t, outputReq)

	outputResp := captureStdout(t, func() {
		DumpResponseIfRequired("test-resp", resp, false)
	})
	assert.Empty(t, outputResp)

	// 2. Verbose = true
	Verbose = true
	outputReq = captureStdout(t, func() {
		DumpRequestIfRequired("test-req", req, false)
	})
	assert.Contains(t, outputReq, "Dumping request 'test-req'")
	assert.Contains(t, outputReq, "Authorization: [REDACTED]")
	assert.Contains(t, outputReq, "access_token=[REDACTED]")
	assert.NotContains(t, outputReq, "sensitive")

	outputResp = captureStdout(t, func() {
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
	u1Mod, err := loadedCfg.GetUser("usr1")
	assert.NoError(t, err)
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

var secretMarkers = []string{
	"eyJLEAKEDACCESS",
	"eyJLEAKEDREFRESH",
	"eyJLEAKEDID",
	"eyJPROBELEAK",
	"LEAKEDCLIENTSECRET",
	"LEAKEDPASSWORD",
	"LEAKEDAUTHCODE",
}

func assertNoCredentialLeaked(t testing.TB, got string) {
	t.Helper()
	for _, marker := range secretMarkers {
		assert.NotContains(t, got, marker, "credential leaked into output")
	}
}

func TestRedactSensitiveContentInBodies(t *testing.T) {
	tests := []struct {
		name        string
		dump        string
		mustHave    []string
		mustNotHave []string
	}{
		{
			name: "json token response",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"access_token":"eyJLEAKEDACCESS","refresh_token":"eyJLEAKEDREFRESH","token_type":"Bearer","expires_in":300}`,
			mustHave: []string{"token_type", "Bearer", "expires_in", "300"},
		},
		{
			name: "json authtoken in body",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"authToken":"eyJPROBELEAK"}`,
			mustHave: []string{`{"authToken":"[REDACTED]"}`},
		},
		{
			name: "json test request with oauth2 context",
			dump: "POST /api/tests HTTP/1.1\r\n" +
				"Content-Type: application/json; charset=utf-8\r\n" +
				"Authorization: Bearer eyJLEAKEDACCESS\r\n" +
				"\r\n" +
				`{"serviceId":"Beer Catalog:0.9","oAuth2Context":{"clientId":"cli","clientSecret":"LEAKEDCLIENTSECRET","username":"bob","password":"LEAKEDPASSWORD","grantType":"PASSWORD"}}`,
			mustHave: []string{"Beer Catalog:0.9", "clientId", "bob", "PASSWORD"},
		},
		{
			name: "form encoded token exchange",
			dump: "POST /token HTTP/1.1\r\n" +
				"Content-Type: application/x-www-form-urlencoded\r\n" +
				"\r\n" +
				"grant_type=authorization_code&code=LEAKEDAUTHCODE&client_secret=LEAKEDCLIENTSECRET",
			mustHave: []string{"grant_type", "authorization_code"},
		},
		{
			name: "form encoded auth-token",
			dump: "POST /api/login HTTP/1.1\r\n" +
				"Content-Type: application/x-www-form-urlencoded\r\n" +
				"\r\n" +
				"auth-token=eyJPROBELEAK",
			mustHave: []string{"auth-token=%5BREDACTED%5D"},
		},
		{
			name: "nested and array json",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"sessions":[{"user":"bob","credentials":{"idToken":"eyJLEAKEDID"}}]}`,
			mustHave: []string{"sessions", "bob"},
		},
		{
			name: "chunked json falls back to text redaction",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"3a\r\n" + `{"access_token":"eyJLEAKEDACCESS"}` + "\r\n0\r\n\r\n",
		},
		{
			name: "oauth code in request line",
			dump: "GET /auth/callback?state=abc&code=LEAKEDAUTHCODE HTTP/1.1\r\n" +
				"Host: localhost:58085\r\n" +
				"\r\n",
			mustHave: []string{"state=abc", "Host: localhost:58085"},
		},
		{
			name: "credentials in both header and body",
			dump: "POST /api/tests HTTP/1.1\r\n" +
				"Authorization: Bearer eyJLEAKEDACCESS\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"password":"LEAKEDPASSWORD"}`,
			mustHave: []string{"Authorization: [REDACTED]"},
		},
		{
			name: "non sensitive body is preserved",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"id":"abc123","success":true,"elapsedTime":42}`,
			mustHave:    []string{"abc123", "true", "42"},
			mustNotHave: []string{"[REDACTED]"},
		},
		{
			name: "header only dump without body",
			dump: "GET /api/keycloak/config HTTP/1.1\r\n" +
				"Accept: application/json",
			mustHave:    []string{"Accept: application/json"},
			mustNotHave: []string{"[REDACTED]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactSensitiveContent(tt.dump)
			assertNoCredentialLeaked(t, got)
			for _, want := range tt.mustHave {
				assert.Contains(t, got, want, "expected value to survive redaction")
			}
			for _, unwanted := range tt.mustNotHave {
				assert.NotContains(t, got, unwanted)
			}
		})
	}
}

func TestRedactSensitiveContentPreservesCRLF(t *testing.T) {
	dump := "GET / HTTP/1.1\r\nAuthorization: Bearer eyJLEAKEDACCESS\r\nAccept: */*\r\n\r\n"
	got := redactSensitiveContent(dump)
	assert.Contains(t, got, "[REDACTED]\r\nAccept:", "CRLF line ending was not preserved around the redacted header")
}

func TestIsSensitiveKeyMatchesSpellingVariants(t *testing.T) {
	sensitive := []string{"access_token", "accessToken", "Access-Token", "ACCESS_TOKEN", "clientSecret", "client_secret", "authToken", "auth-token", "AUTH_TOKEN"}
	for _, key := range sensitive {
		assert.True(t, isSensitiveKey(key), "expected %q to be treated as sensitive", key)
	}

	for _, key := range []string{"token_type", "tokenType", "serviceId", "expires_in"} {
		assert.False(t, isSensitiveKey(key), "did not expect %q to be treated as sensitive", key)
	}
}

func TestDumpHelpersRedactCredentials(t *testing.T) {
	oldVerbose := Verbose
	Verbose = true
	defer func() { Verbose = oldVerbose }()

	t.Run("keycloak token response", func(t *testing.T) {
		body := `{"access_token":"eyJLEAKEDACCESS","refresh_token":"eyJLEAKEDREFRESH"}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Proto:      "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(body)),
			ContentLength: int64(len(body)),
		}
		assertNoCredentialLeaked(t, captureStdout(t, func() {
			DumpResponseIfRequired("Keycloak for getting token", resp, true)
		}))
	})

	t.Run("test creation request", func(t *testing.T) {
		payload := `{"serviceId":"x","oAuth2Context":{"clientSecret":"LEAKEDCLIENTSECRET","password":"LEAKEDPASSWORD"}}`
		req, err := http.NewRequest("POST", "https://microcks.example.com/api/tests", strings.NewReader(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Authorization", "Bearer eyJLEAKEDACCESS")

		assertNoCredentialLeaked(t, captureStdout(t, func() {
			DumpRequestIfRequired("Microcks for creating test", req, true)
		}))
	})
}

func TestDumpResponseLeavesBodyReadable(t *testing.T) {
	oldVerbose := Verbose
	Verbose = true
	defer func() { Verbose = oldVerbose }()

	body := `{"access_token":"eyJLEAKEDACCESS"}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	captureStdout(t, func() { DumpResponseIfRequired("token", resp, true) })

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(got), "body was altered by dump")
}
