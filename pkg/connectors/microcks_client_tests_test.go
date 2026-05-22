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

func makeTestsServer(t *testing.T, serviceID string, results []TestResultSummary, checkPage, checkSize string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/services":
			_ = json.NewEncoder(w).Encode([]serviceSummary{{ID: serviceID, Name: "WeatherForecast API", Version: "1.1.0"}})
		case "/api/tests/service/" + serviceID:
			if checkPage != "" {
				if got := r.URL.Query().Get("page"); got != checkPage {
					t.Fatalf("unexpected page: %s", got)
				}
			}
			if checkSize != "" {
				if got := r.URL.Query().Get("size"); got != checkSize {
					t.Fatalf("unexpected size: %s", got)
				}
			}
			_ = json.NewEncoder(w).Encode(results)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
}

func TestGetTestResultsReturnsResults(t *testing.T) {
	results := []TestResultSummary{
		{ID: "abc123", ServiceID: "svc001", TestedEndpoint: "http://localhost:8080", RunnerType: "OPEN_API_SCHEMA", Success: true},
		{ID: "def456", ServiceID: "svc001", TestedEndpoint: "http://localhost:8080", RunnerType: "OPEN_API_SCHEMA", Success: false},
	}
	server := makeTestsServer(t, "svc001", results, "0", "20")
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
	server := makeTestsServer(t, "svc001", []TestResultSummary{}, "", "")
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	got, err := client.GetTestResults("WeatherForecast API:1.1.0", 0, 20)
	if err != nil {
		t.Fatalf("GetTestResults returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty results, got %d", len(got))
	}
}

func TestGetTestResultsPaginationParams(t *testing.T) {
	server := makeTestsServer(t, "svc001", []TestResultSummary{}, "2", "5")
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetTestResults("WeatherForecast API:1.1.0", 2, 5)
	if err != nil {
		t.Fatalf("GetTestResults returned error: %v", err)
	}
}

func TestGetTestResultsInvalidServiceRefReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client := NewMicrocksClient(server.URL)
	_, err := client.GetTestResults("NoColonHere", 0, 20)
	if err == nil {
		t.Fatal("expected error for invalid serviceRef format, got nil")
	}
}
