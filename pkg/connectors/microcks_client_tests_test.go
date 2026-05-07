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
package connectors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTestResultsReturnsResults(t *testing.T) {
	results := []TestResultSummary{
		{ID: "abc123", ServiceID: "WeatherForecast API:1.1.0", TestedEndpoint: "http://localhost:8080", RunnerType: "OPEN_API_SCHEMA", Success: true},
		{ID: "def456", ServiceID: "WeatherForecast API:1.1.0", TestedEndpoint: "http://localhost:8080", RunnerType: "OPEN_API_SCHEMA", Success: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tests" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.URL.Query().Get("serviceId"); got != "WeatherForecast API:1.1.0" {
			t.Fatalf("unexpected serviceId: %s", got)
		}
		if got := r.URL.Query().Get("page"); got != "0" {
			t.Fatalf("unexpected page: %s", got)
		}
		if got := r.URL.Query().Get("size"); got != "20" {
			t.Fatalf("unexpected size: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results)
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	got, err := client.GetTestResults("WeatherForecast API:1.1.0", 0, 20)
	if err != nil {
		t.Fatalf("GetTestResults returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	if got[0].ID != "abc123" {
		t.Fatalf("unexpected first result ID: %s", got[0].ID)
	}
	if got[1].Success {
		t.Fatalf("expected second result to be failed")
	}
}

func TestGetTestResultsReturnsEmptySlice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	got, err := client.GetTestResults("Unknown API:1.0.0", 0, 20)
	if err != nil {
		t.Fatalf("GetTestResults returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty results, got %d", len(got))
	}
}

func TestGetTestResultsPaginationParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("unexpected page: %s", got)
		}
		if got := r.URL.Query().Get("size"); got != "5" {
			t.Fatalf("unexpected size: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetTestResults("SomeAPI:1.0.0", 2, 5)
	if err != nil {
		t.Fatalf("GetTestResults returned error: %v", err)
	}
}

func TestGetTestResultsInvalidJSONReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetTestResults("SomeAPI:1.0.0", 0, 20)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
