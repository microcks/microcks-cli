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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

func TestExportCommandRequiresServiceOrAll(t *testing.T) {
	_, err := executeCLIForTest(t, "export")
	if err == nil {
		t.Fatal("expected error when neither serviceRef nor --all is provided")
	}
	if !strings.Contains(err.Error(), "missing required argument") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExportCommandExportsArtifacts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/services" && r.Method == http.MethodGet:
			// Service lookup
			if err := json.NewEncoder(w).Encode([]connectors.Service{
				{
					ID:      "svc-123",
					Name:    "Petstore API",
					Version: "1.0",
					Type:    "REST",
				},
			}); err != nil {
				t.Fatalf("failed to encode services: %v", err)
			}
		case r.URL.Path == "/api/services/svc-123":
			if err := json.NewEncoder(w).Encode(map[string]any{
				"service": map[string]string{
					"id":      "svc-123",
					"name":    "Petstore API",
					"version": "1.0",
					"type":    "REST",
				},
			}); err != nil {
				t.Fatalf("failed to encode service detail: %v", err)
			}
		case r.URL.Path == "/api/resources/service/svc-123":
			if err := json.NewEncoder(w).Encode([]connectors.Resource{
				{
					ID:        "res-1",
					Name:      "petstore-openapi.yaml",
					Content:   "openapi: 3.0.0\ninfo:\n  title: Petstore\n  version: 1.0",
					Type:      "OPEN_API_SPEC",
					ServiceID: "svc-123",
				},
			}); err != nil {
				t.Fatalf("failed to encode resources: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	outDir := t.TempDir()
	out, err := executeCLIForTest(t, "export", "Petstore API:1.0", "--output", outDir, "--microcksURL", server.URL)
	if err != nil {
		t.Fatalf("export command returned error: %v, output: %s", err, out)
	}

	expectedFile := filepath.Join(outDir, "petstore-openapi.yaml")
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read exported file %s: %v", expectedFile, err)
	}
	if !strings.Contains(string(content), "openapi: 3.0.0") {
		t.Fatalf("unexpected content in exported file: %s", string(content))
	}
}

func TestExportCommandExportsServiceSnapshot(t *testing.T) {
	const snapshotJSON = `{"services":[{"id":"svc-123","name":"Petstore API","version":"1.0"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/services" && r.Method == http.MethodGet:
			if err := json.NewEncoder(w).Encode([]connectors.Service{
				{
					ID:      "svc-123",
					Name:    "Petstore API",
					Version: "1.0",
					Type:    "REST",
				},
			}); err != nil {
				t.Fatalf("failed to encode services: %v", err)
			}
		case r.URL.Path == "/api/services/svc-123":
			if err := json.NewEncoder(w).Encode(map[string]any{
				"service": map[string]string{
					"id":      "svc-123",
					"name":    "Petstore API",
					"version": "1.0",
					"type":    "REST",
				},
			}); err != nil {
				t.Fatalf("failed to encode service detail: %v", err)
			}
		case r.URL.Path == "/api/export":
			if r.URL.Query().Get("serviceIds") != "svc-123" {
				t.Fatalf("unexpected serviceIds query parameter: %s", r.URL.Query().Get("serviceIds"))
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(snapshotJSON)); err != nil {
				t.Fatalf("failed to write snapshot response: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	outDir := t.TempDir()
	out, err := executeCLIForTest(t, "export", "Petstore API:1.0", "--snapshot", "--output", outDir, "--microcksURL", server.URL)
	if err != nil {
		t.Fatalf("export snapshot returned error: %v, output: %s", err, out)
	}

	expectedFile := filepath.Join(outDir, "Petstore-API-1.0-snapshot.json")
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read exported snapshot file %s: %v", expectedFile, err)
	}
	if string(content) != snapshotJSON {
		t.Fatalf("snapshot content mismatch: got %q, want %q", string(content), snapshotJSON)
	}
}

func TestExportCommandExportsAllSnapshot(t *testing.T) {
	const snapshotJSON = `{"services":[{"id":"svc-1","name":"A","version":"1"},{"id":"svc-2","name":"B","version":"1"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/export" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(snapshotJSON)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	outDir := t.TempDir()
	out, err := executeCLIForTest(t, "export", "--all", "--output", outDir, "--microcksURL", server.URL)
	if err != nil {
		t.Fatalf("export --all returned error: %v, output: %s", err, out)
	}

	expectedFile := filepath.Join(outDir, "microcks-repository.json")
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read exported repo file %s: %v", expectedFile, err)
	}
	if string(content) != snapshotJSON {
		t.Fatalf("snapshot content mismatch: got %q, want %q", string(content), snapshotJSON)
	}
}
