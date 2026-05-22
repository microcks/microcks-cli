package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func findAvailablePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func makeTestOAuth2Config(port int) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    "microcks-app-js",
		RedirectURL: fmt.Sprintf("http://localhost:%d/auth/callback", port),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:0/auth/realms/microcks/protocol/openid-connect/auth",
			TokenURL: "http://localhost:0/auth/realms/microcks/protocol/openid-connect/token",
		},
		Scopes: []string{"openid"},
	}
}

func TestPortHijackDetection(t *testing.T) {
	port := findAvailablePort(t)

	rogueLn, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("failed to start rogue listener: %v", err)
	}
	defer rogueLn.Close()

	rogueSrv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("ROGUE SERVER intercepted request: %s", r.URL)
		w.WriteHeader(200)
		fmt.Fprint(w, "attacker captured route")
	})}
	go rogueSrv.Serve(rogueLn)
	defer rogueSrv.Shutdown(context.Background())

	oauth2conf := makeTestOAuth2Config(port)

	doneCh := make(chan error, 1)
	go func() {
		_, _, err := oauth2login(context.Background(), port, oauth2conf, false)
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Fatal("oauth2login should NOT have returned nil error when the port is hijacked")
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "failed to bind callback port") &&
			!strings.Contains(errMsg, "address already in use") {
			t.Fatalf("expected port-bind error, got: %s", errMsg)
		}
		t.Logf("PASS: Port hijack detected with clear error: %s", errMsg)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: oauth2login did not detect port hijack within 5s")
	}
}

func TestGlobalMuxNotPolluted(t *testing.T) {
	hasCallbackBefore := isRouteInDefaultServeMux("/auth/callback")

	port := findAvailablePort(t)
	oauth2conf := makeTestOAuth2Config(port)

	go func() {
		oauth2login(context.Background(), port, oauth2conf, false)
	}()

	time.Sleep(500 * time.Millisecond)

	hasCallbackAfter := isRouteInDefaultServeMux("/auth/callback")

	if hasCallbackAfter && !hasCallbackBefore {
		t.Fatal("DefaultServeMux was polluted! /auth/callback route was registered globally " +
			"but should use an isolated mux.")
	}
	t.Logf("PASS: DefaultServeMux not polluted (/auth/callback in global mux: before=%v, after=%v)",
		hasCallbackBefore, hasCallbackAfter)
}

func isRouteInDefaultServeMux(path string) bool {
	_, pattern := http.DefaultServeMux.Handler(&http.Request{
		URL:    &url.URL{Path: path},
		Method: "GET",
	})
	return pattern != ""
}

func TestEagerPortBindBeforeBrowser(t *testing.T) {
	port := findAvailablePort(t)
	oauth2conf := makeTestOAuth2Config(port)

	go func() {
		oauth2login(context.Background(), port, oauth2conf, false)
	}()

	time.Sleep(200 * time.Millisecond)

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err == nil {
		ln.Close()
		t.Fatal("Port was NOT bound shortly after oauth2login started — " +
			"server should eagerly bind before opening the browser")
	}

	t.Log("PASS: Port was already bound early (eager bind works)")
}

func TestServerShutdownOnErrorPath(t *testing.T) {
	port := findAvailablePort(t)
	oauth2conf := makeTestOAuth2Config(port)

	go func() {
		oauth2login(context.Background(), port, oauth2conf, false)
	}()

	time.Sleep(500 * time.Millisecond)

	callbackURL := fmt.Sprintf("http://localhost:%d/auth/callback?error=access_denied&error_description=User+cancelled", port)
	_, _ = http.Get(callbackURL)

	time.Sleep(2 * time.Second)

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Server port still bound after error path — server was not shut down properly: %v", err)
	}
	ln.Close()
	t.Log("PASS: Server port released after error path (shutdown on error works)")
}

func TestConcurrentIsolatedMuxCalls(t *testing.T) {
	port1 := findAvailablePort(t)
	port2 := findAvailablePort(t)

	oauth2conf1 := makeTestOAuth2Config(port1)
	oauth2conf2 := makeTestOAuth2Config(port2)

	go func() {
		oauth2login(context.Background(), port1, oauth2conf1, false)
	}()

	go func() {
		oauth2login(context.Background(), port2, oauth2conf2, false)
	}()

	time.Sleep(500 * time.Millisecond)

	ln1, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port1))
	if err == nil {
		ln1.Close()
		t.Fatal("Port 1 was NOT bound — server should eagerly bind")
	}
	ln2, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port2))
	if err == nil {
		ln2.Close()
		t.Fatal("Port 2 was NOT bound — server should eagerly bind")
	}

	t.Log("PASS: Concurrent oauth2login calls both bound their ports independently (isolated muxes)")
}
