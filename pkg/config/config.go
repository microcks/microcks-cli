package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	strings "strings"
)

var (
	// InsecureTLS defines if TLS transport should accept insecure certs.
	InsecureTLS bool
	// CaCertPaths defines extra paths (comma-separated) of CRT files to add to system CA Roots.
	CaCertPaths string
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
			certs, err := ioutil.ReadFile(f)
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
