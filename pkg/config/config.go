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
	"fmt"
	"os"
	"path/filepath"
	strings "strings"
)

var (
	// InsecureTLS defines if TLS transport should accept insecure certs.
	InsecureTLS bool = false
	// CaCertPaths defines extra paths (comma-separated) of CRT files to add to system CA Roots.
	CaCertPaths string

	ConfigPath = filepath.Join(os.Getenv("HOME"), ".microcks-cli", "config.yaml")
)

// CreateTLSConfig wraps the creation of tls.Config object for use with HTTP Client for example.
func CreateTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{}
	if InsecureTLS {
		tlsConfig.InsecureSkipVerify = true
	}
	if len(CaCertPaths) > 0 {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
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
