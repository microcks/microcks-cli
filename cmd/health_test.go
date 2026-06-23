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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

func TestHealthyServer(t *testing.T) {
	// Mock Keycloak and Health endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/keycloak/config") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"enabled":false}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/health") || strings.HasSuffix(r.URL.Path, "/q/health") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"UP","checks":[{"name":"Database connection health check","status":"UP"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Override exitFunc
	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	opts := connectors.ClientOptions{
		ServerAddr: server.URL,
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute returned error: %v", err)
	}

	if exitedCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitedCode)
	}

	output := buf.String()
	if !strings.Contains(output, "✓ Reachable") {
		t.Errorf("output does not contain reachability info: %q", output)
	}
	if !strings.Contains(output, "Overall Status: HEALTHY ✅") {
		t.Errorf("output does not contain healthy overall status: %q", output)
	}
	if !strings.Contains(output, "✓ Keycloak: disabled") {
		t.Errorf("output does not contain keycloak disabled info: %q", output)
	}
	if !strings.Contains(output, "✓ Database: connected") {
		t.Errorf("output does not contain database connected info: %q", output)
	}
}

func TestUnreachableServer(t *testing.T) {
	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	// Use an invalid local address to guarantee connection failure
	opts := connectors.ClientOptions{
		ServerAddr: "http://127.0.0.1:58999/api",
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute returned error: %v", err)
	}

	if exitedCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitedCode)
	}

	output := buf.String()
	if !strings.Contains(output, "✗ Unreachable") {
		t.Errorf("output does not contain unreachable info: %q", output)
	}
	if !strings.Contains(output, "Overall Status: UNHEALTHY ❌") {
		t.Errorf("output does not contain unhealthy status: %q", output)
	}
}

func TestHealthResponseParsing(t *testing.T) {
	// Mock Spring Boot Actuator components format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/keycloak/config") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"enabled":false}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/health") || strings.HasSuffix(r.URL.Path, "/actuator/health") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"UP","components":{"db":{"status":"UP"},"rabbit":{"status":"DOWN"}}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	opts := connectors.ClientOptions{
		ServerAddr: server.URL,
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute returned error: %v", err)
	}

	if exitedCode != 2 {
		t.Errorf("expected exit code 2 (degraded due to rabbit DOWN component), got %d", exitedCode)
	}

	output := buf.String()
	if !strings.Contains(output, "✓ Database: connected") {
		t.Errorf("output should show Database connected: %q", output)
	}
	if !strings.Contains(output, "Overall Status: DEGRADED ⚠️") {
		t.Errorf("output should show degraded status: %q", output)
	}
}

func TestJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/keycloak/config") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"enabled":true,"auth-server-url":"http://keycloak","realm":"microcks"}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/health") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"UP","checks":[{"name":"Database connection health check","status":"UP"},{"name":"Async Minion","status":"UP"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	opts := connectors.ClientOptions{
		ServerAddr: server.URL,
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute returned error: %v", err)
	}

	output := buf.String()
	var jo JSONOutput
	if err := json.Unmarshal([]byte(output), &jo); err != nil {
		t.Fatalf("invalid JSON output: %q, error: %v", output, err)
	}

	if jo.Status != "UP" {
		t.Errorf("expected status 'UP', got %q", jo.Status)
	}
	if len(jo.Checks) == 0 {
		t.Errorf("expected parsed checks, got none")
	}

	foundDb := false
	foundKc := false
	for _, check := range jo.Checks {
		if check.Name == "Database" {
			foundDb = true
			if check.Status != "UP" {
				t.Errorf("expected Database check to be UP, got %q", check.Status)
			}
		}
		if check.Name == "Keycloak" {
			foundKc = true
		}
	}

	if !foundDb {
		t.Errorf("Database check not found in JSON checks")
	}
	if !foundKc {
		t.Errorf("Keycloak check not found in JSON checks")
	}
}

func TestWatchMode(t *testing.T) {
	// Set up mock server
	var reqCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		reqCount++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"UP","checks":[]}`))
	}))
	defer server.Close()

	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	// Inject signal channel to abort watch loop
	watchSignalChan = make(chan os.Signal, 1)
	defer func() { watchSignalChan = nil }()

	opts := connectors.ClientOptions{
		ServerAddr: server.URL,
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--watch", "--interval", "10ms"})

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	// Wait for a few iterations
	time.Sleep(sleepTimeOutFallback(time.Millisecond, 100))

	// Stop watch loop by sending signal
	watchSignalChan <- syscall.SIGINT

	err := <-errCh
	if err != nil {
		t.Fatalf("cmd.Execute returned error: %v", err)
	}

	mu.Lock()
	count := reqCount
	mu.Unlock()

	if count < 2 {
		t.Errorf("expected multiple executions in watch mode, got only %d", count)
	}
}

func TestExitCodeHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	}))
	defer server.Close()

	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	opts := connectors.ClientOptions{
		ServerAddr: server.URL,
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	_ = cmd.Execute()
	if exitedCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitedCode)
	}
}

func TestExitCodeUnhealthy(t *testing.T) {
	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	opts := connectors.ClientOptions{
		ServerAddr: "http://127.0.0.1:58999/api",
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	_ = cmd.Execute()
	if exitedCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitedCode)
	}
}

func TestExitCodeDegraded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"UP","checks":[{"name":"database","status":"DOWN"}]}`))
	}))
	defer server.Close()

	var exitedCode int
	exitFunc = func(code int) {
		exitedCode = code
	}
	defer func() { exitFunc = os.Exit }()

	opts := connectors.ClientOptions{
		ServerAddr: server.URL,
	}
	cmd := NewHealthCommand(&opts)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	_ = cmd.Execute()
	if exitedCode != 2 {
		t.Errorf("expected exit code 2, got %d", exitedCode)
	}
}

// helper function to scale sleep times in test safely
func sleepTimeOutFallback(d time.Duration, factor int) time.Duration {
	return d * time.Duration(factor)
}
