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
	stderrors "errors"
	"net"
	"net/http"
	"testing"

	clierrors "github.com/microcks/microcks-cli/pkg/errors"
)

func TestNewOAuthCallbackServerUsesIsolatedMuxes(t *testing.T) {
	firstStatus := startOAuthCallbackTestServer(t, http.StatusCreated)
	secondStatus := startOAuthCallbackTestServer(t, http.StatusAccepted)

	if firstStatus != http.StatusCreated {
		t.Fatalf("first callback server returned status %d, want %d", firstStatus, http.StatusCreated)
	}
	if secondStatus != http.StatusAccepted {
		t.Fatalf("second callback server returned status %d, want %d", secondStatus, http.StatusAccepted)
	}
}

func TestNewOAuthCallbackServerReturnsErrorWhenPortIsOccupied(t *testing.T) {
	occupiedListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := occupiedListener.Close(); err != nil {
			t.Errorf("closing occupied listener: %v", err)
		}
	})

	server, listener, err := newOAuthCallbackServer(occupiedListener.Addr().String(), http.NotFound)
	if err == nil {
		if server != nil {
			if closeErr := server.Close(); closeErr != nil {
				t.Errorf("closing unexpected callback server: %v", closeErr)
			}
		}
		if listener != nil {
			if closeErr := listener.Close(); closeErr != nil {
				t.Errorf("closing unexpected callback listener: %v", closeErr)
			}
		}
		t.Fatal("expected an error when the callback port is already occupied")
	}
	if got := clierrors.KindOf(err); got != clierrors.KindEnvironment {
		t.Fatalf("error kind = %v, want %v", got, clierrors.KindEnvironment)
	}
}

func startOAuthCallbackTestServer(t *testing.T, status int) int {
	t.Helper()

	server, listener, err := newOAuthCallbackServer("127.0.0.1:0", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	if err != nil {
		t.Fatal(err)
	}

	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- server.Serve(listener)
	}()
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("closing callback server: %v", err)
		}
		if err := <-serveErrors; !stderrors.Is(err, http.ErrServerClosed) {
			t.Errorf("callback server returned unexpected error: %v", err)
		}
	})

	response, err := http.Get("http://" + listener.Addr().String() + "/auth/callback")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("closing callback response body: %v", err)
		}
	}()
	return response.StatusCode
}
