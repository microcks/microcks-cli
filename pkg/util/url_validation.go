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
	"fmt"
	"net"
	"net/url"
	"strings"
)

var allowedSchemes = map[string]bool{
	"https": true,
	"http":  true,
}

var privateHostnames = map[string]bool{
	"localhost":            true,
	"host.docker.internal": true,
}

func ValidateArtifactURL(rawURL string, allowInsecure bool) error {
	if rawURL == "" {
		return fmt.Errorf("URL must not be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	if err := validateScheme(u, allowInsecure); err != nil {
		return err
	}

	if err := validateHost(u.Host); err != nil {
		return err
	}

	return nil
}

func validateScheme(u *url.URL, allowInsecure bool) error {
	scheme := strings.ToLower(u.Scheme)
	if !allowedSchemes[scheme] {
		return fmt.Errorf("scheme %q is not allowed; only https and http are permitted", u.Scheme)
	}
	if scheme == "http" && !allowInsecure {
		return fmt.Errorf("http scheme is not allowed; use https instead or pass the --allow-insecure-url flag")
	}
	return nil
}

func validateHost(host string) error {
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	if isPrivateHostname(hostname) {
		return fmt.Errorf("hostname %q is not allowed as it resolves to an internal address", hostname)
	}

	if ip := net.ParseIP(hostname); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("IP address %q is not allowed as it is a private/reserved address", ip.String())
		}
		return nil
	}

	return resolveAndCheckHost(hostname)
}

func isPrivateHostname(host string) bool {
	return privateHostnames[strings.ToLower(host)]
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

func resolveAndCheckHost(hostname string) error {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname %q: %w", hostname, err)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("hostname %q resolves to private/reserved IP %q which is not allowed", hostname, ip.String())
		}
	}

	return nil
}
