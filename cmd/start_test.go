package cmd

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForReadyImmediate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/keycloak/config" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := waitForReady(server.URL, 5*time.Second); err != nil {
		t.Errorf("expected ready, got error: %v", err)
	}
}

func TestWaitForReadyAfterRetries(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := waitForReady(server.URL, 10*time.Second); err != nil {
		t.Errorf("expected ready after retries, got error: %v", err)
	}
	if calls.Load() < 3 {
		t.Errorf("expected at least 3 polls, got %d", calls.Load())
	}
}

func TestWaitForReadyTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	if err := waitForReady(server.URL, 1*time.Second); err == nil {
		t.Error("expected timeout error, got nil")
	}
}
