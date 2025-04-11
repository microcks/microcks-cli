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
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	strings "strings"

	"gopkg.in/yaml.v2"
)

var (
	// InsecureTLS defines if TLS transport should accept insecure certs.
	InsecureTLS bool = false
	// CaCertPaths defines extra paths (comma-separated) of CRT files to add to system CA Roots.
	CaCertPaths string
	// Verbose represents a debug flag for HTTP Exchanges
	Verbose bool = false

	ConfigPath = filepath.Join(os.Getenv("HOME"), ".microcks-cli", "config.yaml")
)

type Config struct {
	Instance struct {
		Name        string `yaml:"name"`
		Image       string `yaml:"image"`
		Status      string `yaml:"status"`
		Port        string `yaml:"port"`
		ContainerID string `yaml:"containerID"`
		AutoRemove  bool   `yaml:"autoRemove"`
	} `yaml:"instance"`
}

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

// DumpRequestIfRequired takes care of dumping request if configured that way
func DumpRequestIfRequired(name string, req *http.Request, body bool) {
	if Verbose {
		fmt.Printf("\nDumping request '%s':\n", name)
		dump, err := httputil.DumpRequestOut(req, body)
		if err != nil {
			fmt.Println("Got error while dumping request out")
		}
		fmt.Printf("%s", dump)
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
		fmt.Printf("%s", dump)
		if body {
			fmt.Println("")
		}
	}
}

//Functions related to configs

func defaultConfig() *Config {
	return &Config{
		Instance: struct {
			Name        string `yaml:"name"`
			Image       string `yaml:"image"`
			Status      string `yaml:"status"`
			Port        string `yaml:"port"`
			ContainerID string `yaml:"containerID"`
			AutoRemove  bool   `yaml:"autoRemove"`
		}{
			Name:       "microcks",
			Image:      "",
			Status:     "",
			Port:       "",
			AutoRemove: false,
		},
	}
}

func EnsureConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Config not found. Initializing default config.")
		cfg := defaultConfig()
		err := SaveConfig(path, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}
	return LoadConfig(path)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
