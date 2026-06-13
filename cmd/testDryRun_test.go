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
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

const existingArtifact = "../samples/weather-forecast-openapi.yml"

func TestValidateDryRunOptionsValid(t *testing.T) {
	err := validateDryRunOptions(dryRunOptions{
		artifact: existingArtifact,
		image:    defaultDryRunImage,
	})
	if err != nil {
		t.Errorf("expected valid options, got: %v", err)
	}
}

func TestValidateDryRunOptionsMissingArtifact(t *testing.T) {
	err := validateDryRunOptions(dryRunOptions{
		image: defaultDryRunImage,
	})
	if err == nil || !strings.Contains(err.Error(), "--artifact is required") {
		t.Errorf("expected missing-artifact error, got: %v", err)
	}
}

func TestValidateDryRunOptionsArtifactNotFound(t *testing.T) {
	err := validateDryRunOptions(dryRunOptions{
		artifact: "does-not-exist.yaml",
		image:    defaultDryRunImage,
	})
	if err == nil || !strings.Contains(err.Error(), "cannot read --artifact") {
		t.Errorf("expected unreadable-artifact error, got: %v", err)
	}
}

func TestValidateDryRunOptionsNonNativeImage(t *testing.T) {
	err := validateDryRunOptions(dryRunOptions{
		artifact: existingArtifact,
		image:    "quay.io/microcks/microcks-uber:latest",
	})
	if err == nil || !strings.Contains(err.Error(), "uber-native") {
		t.Errorf("expected image flavor error, got: %v", err)
	}
}

func TestRewriteLocalEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		wantEndpoint string
		wantPort     int
		wantRewrite  bool
	}{
		{"localhost with port", "http://localhost:3000", "http://host.testcontainers.internal:3000", 3000, true},
		{"loopback IP with path", "http://127.0.0.1:8080/v1", "http://host.testcontainers.internal:8080/v1", 8080, true},
		{"localhost no port", "http://localhost", "http://host.testcontainers.internal:80", 80, true},
		{"https localhost no port", "https://localhost", "https://host.testcontainers.internal:443", 443, true},
		{"remote host untouched", "https://api.example.com/v2", "https://api.example.com/v2", 0, false},
		{"other IP untouched", "http://192.168.1.10:3000", "http://192.168.1.10:3000", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, port, rewritten := rewriteLocalEndpoint(tt.endpoint)
			if endpoint != tt.wantEndpoint || port != tt.wantPort || rewritten != tt.wantRewrite {
				t.Errorf("rewriteLocalEndpoint(%q) = (%q, %d, %v), want (%q, %d, %v)",
					tt.endpoint, endpoint, port, rewritten, tt.wantEndpoint, tt.wantPort, tt.wantRewrite)
			}
		})
	}
}

func TestDryRunFlagsRegistered(t *testing.T) {
	cmd := NewTestCommand(&connectors.ClientOptions{})
	for _, flag := range []string{"dry-run", "artifact", "image", "ready-timeout", "watch"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag to be registered", flag)
		}
	}
}
