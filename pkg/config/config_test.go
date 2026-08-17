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
	"strings"
	"testing"
)

// secretMarkers are values that must never survive redaction, whatever
// encoding carries them.
var secretMarkers = []string{
	"eyJLEAKEDACCESS",
	"eyJLEAKEDREFRESH",
	"eyJLEAKEDID",
	"LEAKEDCLIENTSECRET",
	"LEAKEDPASSWORD",
	"LEAKEDAUTHCODE",
}

func assertRedacted(t *testing.T, got string) {
	t.Helper()
	for _, marker := range secretMarkers {
		if strings.Contains(got, marker) {
			t.Errorf("credential %q leaked into output:\n%s", marker, got)
		}
	}
}

func TestRedactSensitiveContent(t *testing.T) {
	tests := []struct {
		name        string
		dump        string
		mustNotHave []string
		mustHave    []string
	}{
		{
			// Regression: the Keycloak token endpoint answers in JSON, so the
			// form-encoded "access_token=..." shape never appears on the wire.
			name: "json token response",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"access_token":"eyJLEAKEDACCESS","refresh_token":"eyJLEAKEDREFRESH","token_type":"Bearer","expires_in":300}`,
			mustHave: []string{"token_type", "Bearer", "expires_in", "300"},
		},
		{
			// Regression: oAuth2Context carries a client secret and an end-user
			// password in the body of POST /api/tests.
			name: "json test request with oauth2 context",
			dump: "POST /api/tests HTTP/1.1\r\n" +
				"Content-Type: application/json; charset=utf-8\r\n" +
				"Authorization: Bearer eyJLEAKEDACCESS\r\n" +
				"\r\n" +
				`{"serviceId":"Beer Catalog:0.9","oAuth2Context":{"clientId":"cli","clientSecret":"LEAKEDCLIENTSECRET","username":"bob","password":"LEAKEDPASSWORD","grantType":"PASSWORD"}}`,
			mustHave: []string{"Beer Catalog:0.9", "clientId", "cli", "bob", "PASSWORD"},
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
			name: "nested and array json",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"sessions":[{"user":"bob","credentials":{"idToken":"eyJLEAKEDID"}}]}`,
			mustHave: []string{"sessions", "bob"},
		},
		{
			// Chunked framing defeats structural parsing; the text fallback
			// must still catch the credential.
			name: "chunked json falls back to text redaction",
			dump: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"3a\r\n" + `{"access_token":"eyJLEAKEDACCESS"}` + "\r\n0\r\n\r\n",
		},
		{
			name: "authorization header keeps its scheme",
			dump: "GET /api/tests/1 HTTP/1.1\r\n" +
				"Authorization: Bearer eyJLEAKEDACCESS\r\n" +
				"Accept: application/json\r\n" +
				"\r\n",
			mustHave:    []string{"Bearer [REDACTED]", "Accept: application/json"},
			mustNotHave: []string{"Bearer eyJ"},
		},
		{
			name: "basic auth header",
			dump: "POST /token HTTP/1.1\r\n" +
				"Authorization: Basic dXNlcjpMRUFLRURQQVNTV09SRA==\r\n" +
				"\r\n",
			mustHave:    []string{"Basic [REDACTED]"},
			mustNotHave: []string{"dXNlcjpMRUFLRURQQVNTV09SRA=="},
		},
		{
			// Nothing sensitive: the dump must survive untouched so --verbose
			// stays useful.
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
			assertRedacted(t, got)
			for _, want := range tt.mustHave {
				if !strings.Contains(got, want) {
					t.Errorf("expected %q to survive redaction, got:\n%s", want, got)
				}
			}
			for _, unwanted := range tt.mustNotHave {
				if strings.Contains(got, unwanted) {
					t.Errorf("did not expect %q in output, got:\n%s", unwanted, got)
				}
			}
		})
	}
}

func TestRedactSensitiveContentPreservesCRLF(t *testing.T) {
	dump := "GET / HTTP/1.1\r\nAuthorization: Bearer eyJLEAKEDACCESS\r\nAccept: */*\r\n\r\n"
	got := redactSensitiveContent(dump)
	if !strings.Contains(got, "[REDACTED]\r\nAccept:") {
		t.Errorf("CRLF line ending was not preserved around the redacted header:\n%q", got)
	}
}

func TestNormalizeKeyMatchesSpellingVariants(t *testing.T) {
	sensitive := []string{"access_token", "accessToken", "Access-Token", "ACCESS_TOKEN", "clientSecret", "client_secret"}
	for _, key := range sensitive {
		if !isSensitiveKey(key) {
			t.Errorf("expected %q to be treated as sensitive", key)
		}
	}
	// token_type is metadata, not a credential, and must stay readable.
	for _, key := range []string{"token_type", "tokenType", "serviceId", "expires_in"} {
		if isSensitiveKey(key) {
			t.Errorf("did not expect %q to be treated as sensitive", key)
		}
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = original }()

	f()
	w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	return buf.String()
}

// End-to-end over the dump helpers, on the exact call shapes used by
// test/import/importURL under --verbose.
func TestDumpHelpersRedactCredentials(t *testing.T) {
	previous := Verbose
	Verbose = true
	defer func() { Verbose = previous }()

	t.Run("token response", func(t *testing.T) {
		body := `{"access_token":"eyJLEAKEDACCESS","refresh_token":"eyJLEAKEDREFRESH"}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Proto:      "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(body)),
			ContentLength: int64(len(body)),
		}
		assertRedacted(t, captureStdout(t, func() {
			DumpResponseIfRequired("Keycloak for getting token", resp, true)
		}))
	})

	t.Run("test creation request", func(t *testing.T) {
		payload := `{"serviceId":"x","oAuth2Context":{"clientSecret":"LEAKEDCLIENTSECRET","password":"LEAKEDPASSWORD"}}`
		req, err := http.NewRequest("POST", "https://microcks.example.com/api/tests", strings.NewReader(payload))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Authorization", "Bearer eyJLEAKEDACCESS")

		assertRedacted(t, captureStdout(t, func() {
			DumpRequestIfRequired("Microcks for creating test", req, true)
		}))
	})
}

// The response body must remain readable by the caller after being dumped.
func TestDumpResponseLeavesBodyReadable(t *testing.T) {
	previous := Verbose
	Verbose = true
	defer func() { Verbose = previous }()

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
	if err != nil {
		t.Fatalf("failed to read body after dump: %v", err)
	}
	if string(got) != body {
		t.Errorf("body altered by dump: got %q, want %q", got, body)
	}
}
