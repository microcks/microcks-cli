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

var sensitiveHeaderPattern = regexp.MustCompile(
	`(?im)^((?:Authorization|Proxy-Authorization|X-Auth-Token|Cookie|Set-Cookie):[ \t]*)[^\r\n]*`,
)

var contentTypePattern = regexp.MustCompile(`(?im)^Content-Type:[ \t]*([^\r\n]+)`)

// sensitiveTextPattern matches key-value pairs in text and query strings.
// '?' is excluded to avoid swallowing query strings in URLs.
var sensitiveTextPattern = regexp.MustCompile(
	`(["']?)([A-Za-z0-9_-]+)(["']?[ \t]*[:=][ \t]*)(["']?)([^"'&,}?\r\n\s]+)(["']?)`,
)

var sensitiveValueKeys = map[string]struct{}{
	"accesstoken":   {},
	"refreshtoken":  {},
	"idtoken":       {},
	"authtoken":     {},
	"token":         {},
	"clientsecret":  {},
	"password":      {},
	"secret":        {},
	"code":          {},
	"codeverifier":  {},
	"authorization": {},
	"apikey":        {},
}

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

func redactSensitiveContent(dump string) string {
	head, sep, body := splitHTTPMessage(dump)
	contentType := contentTypeOf(head)

	head = sensitiveHeaderPattern.ReplaceAllString(head, "${1}"+redactedValue)
	head = redactText(head)

	if sep == "" {
		return head
	}
	return head + sep + redactBody(contentType, body)
}

func splitHTTPMessage(dump string) (head, sep, body string) {
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

func redactJSONBody(body string) (string, bool) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return body, true
	}

	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.UseNumber()

	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		return "", false
	}
	if decoder.More() {
		return "", false
	}

	// json.Marshal reorders keys and HTML-escapes, so dumped bodies are not byte-faithful.
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
	return values.Encode() + body[len(trimmed):], true
}

func redactText(body string) string {
	return sensitiveTextPattern.ReplaceAllStringFunc(body, func(match string) string {
		groups := sensitiveTextPattern.FindStringSubmatch(match)
		if !isSensitiveKey(groups[2]) {
			return match
		}
		return groups[1] + groups[2] + groups[3] + groups[4] + redactedValue + groups[6]
	})
}
