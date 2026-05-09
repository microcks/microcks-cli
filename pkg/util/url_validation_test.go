/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateArtifactURL(t *testing.T) {
	tests := []struct {
		name          string
		rawURL        string
		allowInsecure bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid HTTPS public URL",
			rawURL:        "https://example.com/spec.yaml",
			allowInsecure: false,
			expectError:   false,
		},
		{
			name:          "HTTP without allowInsecure flag",
			rawURL:        "http://example.com/spec.yaml",
			allowInsecure: false,
			expectError:   true,
			errorContains: "http scheme is not allowed",
		},
		{
			name:          "HTTP with allowInsecure flag and public IP",
			rawURL:        "http://93.184.216.34/spec.yaml",
			allowInsecure: true,
			expectError:   false,
		},
		{
			name:          "ftp scheme always rejected",
			rawURL:        "ftp://example.com/spec.yaml",
			allowInsecure: false,
			expectError:   true,
			errorContains: "scheme \"ftp\" is not allowed",
		},
		{
			name:          "ftp scheme rejected even with allowInsecure",
			rawURL:        "ftp://example.com/spec.yaml",
			allowInsecure: true,
			expectError:   true,
			errorContains: "scheme \"ftp\" is not allowed",
		},
		{
			name:          "file scheme rejected - no host",
			rawURL:        "file:///etc/passwd",
			allowInsecure: false,
			expectError:   true,
			errorContains: "URL must have a host",
		},
		{
			name:          "gopher scheme always rejected",
			rawURL:        "gopher://example.com/spec.yaml",
			allowInsecure: true,
			expectError:   true,
			errorContains: "scheme \"gopher\" is not allowed",
		},
		{
			name:          "localhost hostname rejected",
			rawURL:        "https://localhost/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "internal address",
		},
		{
			name:          "localhost with port rejected",
			rawURL:        "https://localhost:8080/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "internal address",
		},
		{
			name:          "127.0.0.1 rejected",
			rawURL:        "https://127.0.0.1/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "127.0.0.1 with port rejected",
			rawURL:        "https://127.0.0.1:8080/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "cloud metadata 169.254.169.254 rejected",
			rawURL:        "http://169.254.169.254/latest/meta-data/",
			allowInsecure: true,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "cloud metadata 169.254.169.254 rejected even with HTTPS",
			rawURL:        "https://169.254.169.254/latest/meta-data/",
			allowInsecure: true,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "RFC1918 10.x rejected",
			rawURL:        "https://10.0.0.1/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "RFC1918 172.16.x rejected",
			rawURL:        "https://172.16.0.1/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "RFC1918 192.168.x rejected",
			rawURL:        "https://192.168.1.1/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "IPv6 loopback rejected",
			rawURL:        "https://[::1]/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "0.0.0.0 rejected",
			rawURL:        "https://0.0.0.0/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "host.docker.internal rejected",
			rawURL:        "https://host.docker.internal/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "internal address",
		},
		{
			name:          "public IP allowed",
			rawURL:        "https://93.184.216.34/spec.yaml",
			allowInsecure: false,
			expectError:   false,
		},
		{
			name:          "malformed URL rejected - no scheme",
			rawURL:        "not-a-url",
			allowInsecure: false,
			expectError:   true,
			errorContains: "URL must have a scheme",
		},
		{
			name:          "empty URL rejected",
			rawURL:        "",
			allowInsecure: false,
			expectError:   true,
			errorContains: "URL must not be empty",
		},
		{
			name:          "link-local 169.254.1.1 rejected",
			rawURL:        "https://169.254.1.1/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "carrier-grade NAT 100.64.0.1 rejected",
			rawURL:        "https://100.64.0.1/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "127.x.x.x loopback range rejected",
			rawURL:        "https://127.255.255.255/api",
			allowInsecure: false,
			expectError:   true,
			errorContains: "private/reserved address",
		},
		{
			name:          "localhost with HTTP and allowInsecure still rejected for private host",
			rawURL:        "http://localhost:8080/api",
			allowInsecure: true,
			expectError:   true,
			errorContains: "internal address",
		},
		{
			name:          "HTTP with allowInsecure and public IP allowed",
			rawURL:        "http://8.8.8.8/spec.yaml",
			allowInsecure: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArtifactURL(tt.rawURL, tt.allowInsecure)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		isPrivate bool
	}{
		{"loopback", "127.0.0.1", true},
		{"loopback alternate", "127.0.0.2", true},
		{"link-local", "169.254.169.254", true},
		{"RFC1918 10.x", "10.0.0.1", true},
		{"RFC1918 172.16.x", "172.16.0.1", true},
		{"RFC1918 192.168.x", "192.168.1.1", true},
		{"unspecified", "0.0.0.0", true},
		{"IPv6 loopback", "::1", true},
		{"IPv6 unspecified", "::", true},
		{"public IP", "93.184.216.34", false},
		{"public IP 2", "8.8.8.8", false},
		{"carrier-grade NAT", "100.64.0.1", true},
		{"carrier-grade NAT end range", "100.127.255.255", true},
		{"just outside carrier-grade NAT", "100.128.0.0", false},
		{"link-local multicast is private", "224.0.0.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := parseIPLiteral(tt.ip)
			assert.Equal(t, tt.isPrivate, isPrivateIP(ip), "isPrivateIP(%s)", tt.ip)
		})
	}
}

func TestIsPrivateHostname(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		isPrivate bool
	}{
		{"localhost", "localhost", true},
		{"LOCALHOST uppercase", "LOCALHOST", true},
		{"Localhost mixed case", "Localhost", true},
		{"host.docker.internal", "host.docker.internal", true},
		{"public host", "example.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isPrivate, isPrivateHostname(tt.host))
		})
	}
}

func parseIPLiteral(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		panic("invalid IP: " + s)
	}
	return ip
}
