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
package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	// InsecureTLS defines if TLS transport should accept insecure certs.
	InsecureTLS bool = false
	// CaCertPaths defines extra paths (comma-separated) of CRT files to add to system CA Roots.
	CaCertPaths string
	// Verbose represents a debug flag for HTTP Exchanges
	Verbose bool = false
)

const redactedValue = "[REDACTED]"

// sensitiveHeaderPattern matches headers whose whole value is a credential.
var sensitiveHeaderPattern = regexp.MustCompile(
	`(?im)^((?:Authorization|Proxy-Authorization|X-Auth-Token|Cookie|Set-Cookie):[ \t]*)[^\r\n]*`,
)

// contentTypePattern extracts the media type from a dumped header block.
var contentTypePattern = regexp.MustCompile(`(?im)^Content-Type:[ \t]*([^\r\n]+)`)

// sensitiveTextPattern matches "key=value", `"key":"value"` and quoted variants.
// It covers query strings in request lines and redirect headers, plus bodies
// that cannot be parsed structurally (chunked framing, unknown encodings). Key
// membership is checked in the replacement callback.
//
// '?' is excluded from the value class so that a URL such as
// "http://host?access_token=x" does not let the leading "http://host" match
// swallow the query string before its parameters are examined.
var sensitiveTextPattern = regexp.MustCompile(
	`(["']?)([A-Za-z0-9_-]+)(["']?[ \t]*[:=][ \t]*)(["']?)([^"'&,}?\r\n\s]+)(["']?)`,
)

// sensitiveValueKeys holds the normalized parameter and field names whose
// values must never reach verbose output, whatever encoding carries them.
// Matching is on the name, not on the delimiter, so form bodies
// (access_token=...) and JSON bodies ("accessToken": "...") are covered alike.
var sensitiveValueKeys = map[string]struct{}{
	"accesstoken":   {},
	"refreshtoken":  {},
	"idtoken":       {},
	"token":         {},
	"clientsecret":  {},
	"password":      {},
	"secret":        {},
	"code":          {},
	"codeverifier":  {},
	"authorization": {},
	"apikey":        {},
}

// normalizeKey folds a name so snake_case, camelCase, kebab-case and
// capitalized spellings of the same field compare equal.
func normalizeKey(key string) string {
	var b strings.Builder
	for _, r := range key {
		if r == '_' || r == '-' || r == ' ' {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func isSensitiveKey(key string) bool {
	_, ok := sensitiveValueKeys[normalizeKey(key)]
	return ok
}

// CreateTLSConfig wraps the creation of tls.Config object for use with HTTP Client for example.
func CreateTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{}
	if InsecureTLS {
		tlsConfig.InsecureSkipVerify = true
	}
	if len(CaCertPaths) > 0 {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, err := x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		sepCaFiles := strings.Split(CaCertPaths, ",")
		for _, f := range sepCaFiles {
			// Read in the cert file
			certs, err := os.ReadFile(f)
			if err != nil {
				fmt.Println("Unable to read cert file from CaCertPaths: " + f)
			}

			// Append our cert to the system pool
			if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
				fmt.Println("Unable to append cert file from CaCertPaths: " + f)
			}
		}
		tlsConfig.RootCAs = rootCAs
	}
	return tlsConfig
}

// DumpRequestIfRequired takes care of dumping request if configured that way
func DumpRequestIfRequired(name string, req *http.Request, body bool) {
	if Verbose {
		fmt.Printf("\nDumping request '%s':\n", name)
		dump, err := httputil.DumpRequestOut(req, body)
		if err != nil {
			fmt.Println("Got error while dumping request out")
		}
		fmt.Printf("%s", redactSensitiveContent(string(dump)))
	}
}

// DumpResponseIfRequired takes care of dumping request if configured that way
func DumpResponseIfRequired(name string, resp *http.Response, body bool) {
	if Verbose {
		fmt.Printf("\nDumping response '%s':\n", name)
		dump, err := httputil.DumpResponse(resp, body)
		if err != nil {
			fmt.Println("Got error while dumping response")
		}
		fmt.Printf("%s", redactSensitiveContent(string(dump)))
		if body {
			fmt.Println("")
		}
	}
}

// redactSensitiveContent masks OAuth tokens and credentials in HTTP dump
// output. Headers and body are redacted separately: the body is parsed
// according to its Content-Type so that credentials are matched by field name
// rather than by wire delimiter.
func redactSensitiveContent(dump string) string {
	head, sep, body := splitHTTPMessage(dump)
	contentType := contentTypeOf(head)

	// Credential-bearing headers are masked whole. The rest of the head still
	// needs scanning: the request line and redirect targets carry OAuth
	// parameters in their query string.
	head = sensitiveHeaderPattern.ReplaceAllString(head, "${1}"+redactedValue)
	head = redactText(head)

	if sep == "" {
		return head
	}
	return head + sep + redactBody(contentType, body)
}

// splitHTTPMessage divides a dumped HTTP message into its header block, the
// blank-line separator and its body. sep is empty when there is no body.
func splitHTTPMessage(dump string) (head, sep, body string) {
	// "\r\n\r\n" is checked first: it contains no "\n\n", so a CRLF message can
	// never be split on the LF-only boundary by mistake.
	for _, candidate := range []string{"\r\n\r\n", "\n\n"} {
		if before, after, found := strings.Cut(dump, candidate); found {
			return before, candidate, after
		}
	}
	return dump, "", ""
}

func contentTypeOf(head string) string {
	if m := contentTypePattern.FindStringSubmatch(head); m != nil {
		return strings.ToLower(m[1])
	}
	return ""
}

// redactBody masks credential-bearing fields in a dumped body, falling back to
// text matching whenever the body cannot be parsed as its declared type.
func redactBody(contentType, body string) string {
	switch {
	case strings.Contains(contentType, "json"):
		if out, ok := redactJSONBody(body); ok {
			return out
		}
	case strings.Contains(contentType, "x-www-form-urlencoded"):
		if out, ok := redactFormBody(body); ok {
			return out
		}
	}
	return redactText(body)
}

// redactJSONBody rewrites a JSON body with sensitive members masked. It reports
// false when the body is not a single well-formed JSON value, so the caller can
// fall back rather than emit a truncated re-encoding.
func redactJSONBody(body string) (string, bool) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return body, true
	}

	decoder := json.NewDecoder(strings.NewReader(trimmed))
	// Preserve the original number formatting instead of round-tripping
	// every number through float64.
	decoder.UseNumber()

	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		return "", false
	}
	// Trailing content means this was not a bare JSON body (chunked transfer
	// framing, for instance); re-encoding would silently drop it.
	if decoder.More() {
		return "", false
	}

	out, err := json.Marshal(redactJSONValue(value))
	if err != nil {
		return "", false
	}
	return string(out), true
}

func redactJSONValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		for k, v := range typed {
			if isSensitiveKey(k) {
				typed[k] = redactedValue
				continue
			}
			typed[k] = redactJSONValue(v)
		}
		return typed
	case []interface{}:
		for i, v := range typed {
			typed[i] = redactJSONValue(v)
		}
		return typed
	}
	return value
}

// redactFormBody rewrites a form-urlencoded body with sensitive parameters
// masked. It reports false when the body cannot be parsed as a query string.
func redactFormBody(body string) (string, bool) {
	trimmed := strings.TrimRight(body, "\r\n")
	if trimmed == "" {
		return body, true
	}

	values, err := url.ParseQuery(trimmed)
	if err != nil {
		return "", false
	}
	for key, vals := range values {
		if !isSensitiveKey(key) {
			continue
		}
		for i := range vals {
			vals[i] = redactedValue
		}
	}
	// Preserve whatever trailing newlines the dump carried.
	return values.Encode() + body[len(trimmed):], true
}

// redactText masks sensitive key/value pairs in a body of unknown or
// unparseable encoding.
func redactText(body string) string {
	return sensitiveTextPattern.ReplaceAllStringFunc(body, func(match string) string {
		groups := sensitiveTextPattern.FindStringSubmatch(match)
		if !isSensitiveKey(groups[2]) {
			return match
		}
		return groups[1] + groups[2] + groups[3] + groups[4] + redactedValue + groups[6]
	})
}
