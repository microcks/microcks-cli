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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"golang.org/x/oauth2"
)

// KeycloakClient defines methods for cinteracting with Keycloak
type KeycloakClient interface {
	ConnectAndGetToken() (string, error)
	ConnectAndGetTokenAndRefreshToken(string, string) (string, string, error)
	GetOIDCConfig() (*oauth2.Config, error)
}

type keycloakClient struct {
	BaseURL  *url.URL
	Username string
	Password string

	httpClient *http.Client
}

// NewKeycloakClient build a new KeycloakClient implementation
func NewKeycloakClient(realmURL string, username string, password string) KeycloakClient {
	kc := keycloakClient{}

	u, err := url.Parse(realmURL)
	if err != nil {
		panic(err)
	}
	kc.BaseURL = u
	kc.Username = username
	kc.Password = password

	if config.InsecureTLS || len(config.CaCertPaths) > 0 {
		tlsConfig := config.CreateTLSConfig()
		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		kc.httpClient = &http.Client{Transport: tr}
	} else {
		kc.httpClient = http.DefaultClient
	}
	return &kc
}

// ConnectAndGetToken implementation on keycloakClient structure
func (c *keycloakClient) ConnectAndGetToken() (string, error) {
	rel := &url.URL{Path: "protocol/openid-connect/token"}
	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(url.Values{"grant_type": {"client_credentials"}}.Encode()))
	if err != nil {
		return "", err
	}

	credential := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+credential)

	// Dump request if verbose required.
	config.DumpRequestIfRequired("Keycloak for getting token", req, false)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Dump response if verbose required.
	config.DumpResponseIfRequired("Keycloak for getting token", resp, true)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var openIDResp map[string]interface{}
	if err := json.Unmarshal(body, &openIDResp); err != nil {
		panic(err)
	}

	accessToken := openIDResp["access_token"].(string)
	return accessToken, err
}

func (c *keycloakClient) GetOIDCConfig() (*oauth2.Config, error) {
	rel := &url.URL{Path: ".well-known/openid-configuration"}
	u := c.BaseURL.ResolveReference(rel)

	// Create HTTP request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var openIDResp map[string]interface{}
	if err := json.Unmarshal(body, &openIDResp); err != nil {
		panic(err)
	}

	authURL := openIDResp["authorization_endpoint"].(string)
	tokenURL := openIDResp["token_endpoint"].(string)

	return &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}, nil
}

func (c *keycloakClient) ConnectAndGetTokenAndRefreshToken(username, password string) (string, string, error) {

	rel := &url.URL{Path: "protocol/openid-connect/token"}
	u := c.BaseURL.ResolveReference(rel)

	data := url.Values{}
	data.Set("client_id", c.Username)
	data.Set("client_secret", c.Password)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("grant_type", "password")
	// Create HTTP request
	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var openIDResp map[string]interface{}
	if err := json.Unmarshal(body, &openIDResp); err != nil {
		panic(err)
	}

	authToken := openIDResp["access_token"].(string)
	refershToken := openIDResp["refresh_token"].(string)

	return authToken, refershToken, nil
}
