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
	"strings"
	"testing"
)

func TestTestListCommandOutputsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tests" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("serviceId"); got != "svc-1" {
			t.Fatalf("unexpected serviceId: %s", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":        "test-1",
			"serviceId": "svc-1",
			"success":   true,
		}})
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "test", "list", "--microcksURL", server.URL, "--serviceId", "svc-1", "--output", "json")
	if err != nil {
		t.Fatalf("command returned error: %v", err)
	}
	if !strings.Contains(out, `"id": "test-1"`) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestTestGetCommandOutputsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tests/test-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "test-1",
			"serviceId": "svc-1",
			"success":   false,
		})
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "test", "get", "test-1", "--microcksURL", server.URL, "--output", "json")
	if err != nil {
		t.Fatalf("command returned error: %v", err)
	}
	if !strings.Contains(out, `"id": "test-1"`) {
		t.Fatalf("unexpected output: %s", out)
	}
}
