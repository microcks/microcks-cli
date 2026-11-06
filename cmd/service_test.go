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
	"slices"
	"strings"
	"testing"
)

func TestServiceListCommandOutputsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode([]map[string]string{{
			"id":      "svc-1",
			"name":    "Catalog API",
			"version": "1.0.0",
			"type":    "REST",
		}}); err != nil {
			t.Fatalf("failed to encode services response: %v", err)
		}
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
		if err := json.NewEncoder(w).Encode(map[string]any{
			"service": map[string]string{
				"id":      "svc-1",
				"name":    "Catalog API",
				"version": "1.0.0",
				"type":    "REST",
			},
		}); err != nil {
			t.Fatalf("failed to encode service response: %v", err)
		}
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
	if !slices.Contains(args, "--config") {
		args = append(args, "--config", t.TempDir()+"/config.yaml")
	}
	command.SetArgs(args)

	execErr := command.Execute()
	if err := w.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	os.Stdout = oldStdout

	out, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("ReadAll returned error: %v", readErr)
	}
	return string(out), execErr
}

func TestServiceDeleteCommandRequiresForceInNonTerminal(t *testing.T) {
	oldIsTerminal := isTerminal
	isTerminal = func() bool { return false }
	defer func() { isTerminal = oldIsTerminal }()

	_, err := executeCLIForTest(t, "service", "delete", "svc-1", "--microcksURL", "http://localhost:8080")
	if err == nil {
		t.Fatal("expected error when running in non-terminal without --force, got nil")
	}
	if !strings.Contains(err.Error(), "--force flag is required") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestServiceDeleteCommandWithForce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/svc-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "service", "delete", "svc-1", "--force", "--microcksURL", server.URL)
	if err != nil {
		t.Fatalf("command returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "Service 'svc-1' deleted successfully") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestServiceDeleteCommandWithConfirmationAccepted(t *testing.T) {
	oldIsTerminal := isTerminal
	oldStdinReader := stdinReader
	isTerminal = func() bool { return true }
	stdinReader = strings.NewReader("y\n")
	defer func() {
		isTerminal = oldIsTerminal
		stdinReader = oldStdinReader
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/svc-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "service", "delete", "svc-1", "--microcksURL", server.URL)
	if err != nil {
		t.Fatalf("command returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "Service 'svc-1' deleted successfully") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestServiceDeleteCommandWithConfirmationAborted(t *testing.T) {
	oldIsTerminal := isTerminal
	oldStdinReader := stdinReader
	isTerminal = func() bool { return true }
	stdinReader = strings.NewReader("n\n")
	defer func() {
		isTerminal = oldIsTerminal
		stdinReader = oldStdinReader
	}()

	out, err := executeCLIForTest(t, "service", "delete", "svc-1", "--microcksURL", "http://localhost:8080")
	if err != nil {
		t.Fatalf("command returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "Service deletion aborted") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestRootDeleteAliasCommandWithForce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services/svc-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "delete", "svc-1", "--force", "--microcksURL", server.URL)
	if err != nil {
		t.Fatalf("command returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "Service 'svc-1' deleted successfully") {
		t.Fatalf("unexpected output: %s", out)
	}
}
