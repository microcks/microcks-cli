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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	microckserrors "github.com/microcks/microcks-cli/pkg/errors"
)

func TestConnectAndGetTokenReturnsErrorOnNon200(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"unauthorized", http.StatusUnauthorized, `{"error":"unauthorized","error_description":"Invalid credentials"}`},
		{"forbidden", http.StatusForbidden, `{"error":"access_denied"}`},
		{"server error", http.StatusInternalServerError, `internal server error`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			kc, err := NewKeycloakClient(server.URL, "client-id", "client-secret")
			if err != nil {
				t.Fatalf("NewKeycloakClient returned unexpected error: %v", err)
			}

			token, err := kc.ConnectAndGetToken()

			if token != "" {
				t.Errorf("expected empty token on error, got %q", token)
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if microckserrors.KindOf(err) != microckserrors.KindAPI {
				t.Errorf("expected KindAPI, got %v", microckserrors.KindOf(err))
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("%d", tc.statusCode)) {
				t.Errorf("expected error to contain status code %d, got: %s", tc.statusCode, err.Error())
			}
		})
	}
}
