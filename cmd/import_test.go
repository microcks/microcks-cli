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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestImportCommandUsesMicrocksURLWithoutLocalConfig(t *testing.T) {
	artifact := t.TempDir() + "/openapi.yaml"
	if err := os.WriteFile(artifact, []byte("openapi: 3.0.0\ninfo:\n  title: Demo\n  version: 1.0.0\npaths: {}\n"), 0o644); err != nil {
		t.Fatalf("failed to write artifact: %v", err)
	}

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/api/artifact/upload" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		if got := r.MultipartForm.Value["mainArtifact"]; len(got) != 1 || got[0] != "true" {
			t.Fatalf("unexpected mainArtifact: %v", got)
		}
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("Demo:1.0.0")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}

	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "import", artifact, "--microcksURL", server.URL, "--config", t.TempDir()+"/config")
	if err != nil {
		t.Fatalf("command returned error: %v", err)
	}
	if !called {
		t.Fatal("expected import command to call Microcks upload endpoint")
	}
	if !strings.Contains(out, "Microcks has discovered 'Demo:1.0.0'") {
		t.Fatalf("unexpected output: %s", out)
	}
}
