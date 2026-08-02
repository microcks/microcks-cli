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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestServiceListCommandOutputsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]map[string]string{{
			"id":      "svc-1",
			"name":    "Catalog API",
			"version": "1.0.0",
			"type":    "REST",
		}})
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "service", "list", "--microcksURL", server.URL, "--output", "json")
	if err != nil {
		t.Fatalf("command returned error: %v", err)
	}
	if !strings.Contains(out, `"id": "svc-1"`) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestServiceGetCommandOutputsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/svc-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": map[string]string{
				"id":      "svc-1",
				"name":    "Catalog API",
				"version": "1.0.0",
				"type":    "REST",
			},
		})
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "service", "get", "svc-1", "--microcksURL", server.URL, "--output", "json")
	if err != nil {
		t.Fatalf("command returned error: %v", err)
	}
	if !strings.Contains(out, `"name": "Catalog API"`) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func executeCLIForTest(t *testing.T, args ...string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = w

	command, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand returned error: %v", err)
	}
	command.SetArgs(append(args, "--config", t.TempDir()+"/config.yaml"))

	execErr := command.Execute()
	_ = w.Close()
	os.Stdout = oldStdout

	out, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("ReadAll returned error: %v", readErr)
	}
	return string(out), execErr
}
