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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
)

var (
	grantTypeChoices = map[string]bool{"PASSWORD": true, "CLIENT_CREDENTIALS": true, "REFRESH_TOKEN": true}
)

// MicrocksClient allows interacting with Microcks APIs
type MicrocksClient interface {
	HttpClient() *http.Client
	GetKeycloakURL() (string, error)
	SetOAuthToken(oauthToken string)
	CreateTestResult(serviceID string, testEndpoint string, runnerType string, secretName string, timeout int64, filteredOperations string, operationsHeaders string, oAuth2Context string) (string, error)
	GetTestResult(testResultID string) (*TestResultSummary, error)
	UploadArtifact(specificationFilePath string, mainArtifact bool) (string, error)
	DownloadArtifact(artifactURL string, mainArtifact bool, secret string) (string, error)
}

// TestResultSummary represents a simple view on Microcks TestResult
type TestResultSummary struct {
	ID             string `json:"id"`
	Version        int32  `json:"version"`
	TestNumber     int32  `json:"testNumber"`
	TestDate       int64  `json:"testDate"`
	TestedEndpoint string `json:"testedEndpoint"`
	ServiceID      string `json:"serviceId"`
	ElapsedTime    int32  `json:"elapsedTime"`
	Success        bool   `json:"success"`
	InProgress     bool   `json:"inProgress"`
}

// HeaderDTO represents an operation header passed for Test
type HeaderDTO struct {
	Name   string `json:"name"`
	Values string `json:"values"`
}

// OAuth2ClientContext represents a test request OAuth2 client context
type OAuth2ClientContext struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	TokenURI     string `json:"tokenUri"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RefreshToken string `json:"refreshToken"`
	GrantType    string `json:"grantType"`
	Scopes       string `json:"scopes"`
}

type microcksClient struct {
	APIURL     *url.URL
	OAuthToken string

	httpClient *http.Client
}

// NewMicrocksClient build a new MicrocksClient implementation
func NewMicrocksClient(apiURL string) MicrocksClient {
	mc := microcksClient{}

	if !strings.HasSuffix(apiURL, "/api/") {
		apiURL += "/api/"
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		panic(err)
	}
	mc.APIURL = u

	if config.InsecureTLS || len(config.CaCertPaths) > 0 {
		tlsConfig := config.CreateTLSConfig()
		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		mc.httpClient = &http.Client{Transport: tr}
	} else {
		mc.httpClient = http.DefaultClient
	}
	return &mc
}

func (mc *microcksClient) HttpClient() *http.Client {
	return mc.httpClient
}

func (c *microcksClient) GetKeycloakURL() (string, error) {
	// Ensure we have a correct URL for retrieving Keycloal configuration.
	rel := &url.URL{Path: "keycloak/config"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	// Dump request if verbose required.
	config.DumpRequestIfRequired("Microcks for getting Keycloak config", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Dump request if verbose required.
	config.DumpResponseIfRequired("Microcks for getting Keycloak config", resp, true)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var configResp map[string]interface{}
	if err := json.Unmarshal(body, &configResp); err != nil {
		panic(err)
	}

	// Retrieve auth server url and realm name.
	enabled := configResp["enabled"].(bool)
	authServerURL := configResp["auth-server-url"].(string)
	realmName := configResp["realm"].(string)

	// Return a proper URL or 'null' if Keycloak is disables.
	if enabled {
		return authServerURL + "/realms/" + realmName + "/", nil
	}
	return "null", nil
}

func (c *microcksClient) SetOAuthToken(oauthToken string) {
	c.OAuthToken = oauthToken
}

func (c *microcksClient) CreateTestResult(serviceID string, testEndpoint string, runnerType string, secretName string, timeout int64, filteredOperations string, operationsHeaders string, oAuth2Context string) (string, error) {
	// Ensure we have a correct URL.
	rel := &url.URL{Path: "tests"}
	u := c.APIURL.ResolveReference(rel)

	// Prepare an input string as body.
	var input = "{"
	input += ("\"serviceId\": \"" + serviceID + "\", ")
	input += ("\"testEndpoint\": \"" + testEndpoint + "\", ")
	input += ("\"runnerType\": \"" + runnerType + "\", ")
	input += ("\"timeout\":  " + strconv.FormatInt(timeout, 10))
	if len(secretName) > 0 {
		input += (", \"secretName\": \"" + secretName + "\"")
	}
	if len(filteredOperations) > 0 && ensureValidOperationsList(filteredOperations) {
		input += (", \"filteredOperations\": " + filteredOperations)
	}
	if len(operationsHeaders) > 0 && ensureValidOperationsHeaders(operationsHeaders) {
		input += (", \"operationsHeaders\": " + operationsHeaders)
	}
	if len(oAuth2Context) > 0 && ensureValieOAuth2Context(oAuth2Context) {
		input += (", \"oAuth2Context\": " + oAuth2Context)
	}

	input += "}"

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(input))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.OAuthToken)

	// Dump request if verbose required.
	config.DumpRequestIfRequired("Microcks for creating test", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Dump response if verbose required.
	config.DumpResponseIfRequired("Microcks for creating test", resp, true)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var createTestResp map[string]interface{}
	if err := json.Unmarshal(body, &createTestResp); err != nil {
		panic(err)
	}

	testID := createTestResp["id"].(string)
	return testID, err
}

func (c *microcksClient) GetTestResult(testResultID string) (*TestResultSummary, error) {
	// Ensure we have a correct URL.
	rel := &url.URL{Path: "tests/" + testResultID}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.OAuthToken)

	// Dump request if verbose required.
	config.DumpRequestIfRequired("Microcks for getting status", req, false)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Dump response if verbose required.
	config.DumpResponseIfRequired("Microcks for getting status test", resp, true)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	result := TestResultSummary{}
	json.Unmarshal([]byte(body), &result)

	return &result, err
}

func (c *microcksClient) UploadArtifact(specificationFilePath string, mainArtifact bool) (string, error) {
	// Ensure file exists on fs.
	file, err := os.Open(specificationFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a multipart request body, reading the file.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(specificationFilePath))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err.Error())
	}

	// Add the mainArtifact flag to request.
	_ = writer.WriteField("mainArtifact", strconv.FormatBool(mainArtifact))

	err = writer.Close()
	if err != nil {
		return "", err
	}

	// Ensure we have a correct URL.
	rel := &url.URL{Path: "artifact/upload"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.OAuthToken)

	// Dump request if verbose required.
	config.DumpRequestIfRequired("Microcks for uploading artifact", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Dump response if verbose required.
	config.DumpResponseIfRequired("Microcks for uploading artifact", resp, true)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	// Raise exception if not created.
	if resp.StatusCode != 201 {
		return "", errors.New(string(respBody))
	}

	return string(respBody), err
}

func (c *microcksClient) DownloadArtifact(artifactURL string, mainArtifact bool, secret string) (string, error) {

	// create Multipart Form to add fields
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add all the form fields
	writer.WriteField("url", artifactURL)
	writer.WriteField("mainArtifact", strconv.FormatBool(mainArtifact))
	if secret != "" {
		writer.WriteField("secret", secret)
	}

	err := writer.Close()
	if err != nil {
		return "", err
	}

	// Ensure we have a correct URL.
	rel := &url.URL{Path: "artifact/download"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.OAuthToken)

	// Dump request if verbose required.
	config.DumpRequestIfRequired("Microcks for uploading artifact", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Dump response if verbose required.
	config.DumpResponseIfRequired("Microcks for uploading artifact", resp, true)

	respBody, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err.Error())
	}

	// Raise exception if not created.
	if resp.StatusCode != 201 {
		return "", errors.New(string(respBody))
	}

	return string(respBody), err
}

func ensureValidOperationsList(filteredOperations string) bool {
	// Unmarshal using a generic interface
	var list = []string{}
	err := json.Unmarshal([]byte(filteredOperations), &list)
	if err != nil {
		fmt.Println("Error parsing JSON in filteredOperations: ", err)
		return false
	}
	return true
}

func ensureValidOperationsHeaders(operationsHeaders string) bool {
	// Unmarshal using a generic interface
	var headers = map[string][]HeaderDTO{}
	err := json.Unmarshal([]byte(operationsHeaders), &headers)
	if err != nil {
		fmt.Println("Error parsing JSON in operationsHeaders: ", err)
		return false
	}
	return true
}

func ensureValieOAuth2Context(oAuth2Context string) bool {
	var oContext = OAuth2ClientContext{}
	err := json.Unmarshal([]byte(oAuth2Context), &oContext)
	if err != nil {
		fmt.Println("Error parsing JSON in oAuth2Context: ", err)
		return false
	}
	if !grantTypeChoices[oContext.GrantType] {
		fmt.Println("grantType in oAuth2Context is not supported. OAuth2 is turned off.")
		return false
	}
	return true
}
