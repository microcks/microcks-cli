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
	"context"
	"crypto/tls"
	"encoding/json"
	errs "errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/errors"
	"golang.org/x/oauth2"
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
	GetFullTestResult(testResultID string) (*TestResult, error)
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

// TestResult represents the full Microcks TestResult with nested test case/step details
type TestResult struct {
	ID             string           `json:"id"`
	Version        int32            `json:"version"`
	TestNumber     int32            `json:"testNumber"`
	TestDate       int64            `json:"testDate"`
	TestedEndpoint string           `json:"testedEndpoint"`
	ServiceID      string           `json:"serviceId"`
	ElapsedTime    int64            `json:"elapsedTime"`
	Success        bool             `json:"success"`
	InProgress     bool             `json:"inProgress"`
	RunnerType     string           `json:"runnerType"`
	TestCases      []TestCaseResult `json:"testCaseResults"`
}

// TestCaseResult represents results for a single operation/action within a test
type TestCaseResult struct {
	Success        bool             `json:"success"`
	ElapsedTime    int64            `json:"elapsedTime"`
	OperationName  string           `json:"operationName"`
	TestStepResults []TestStepResult `json:"testStepResults"`
}

// TestStepResult represents results for a single request/message within a test case
type TestStepResult struct {
	Success         bool   `json:"success"`
	ElapsedTime     int64  `json:"elapsedTime"`
	RequestName     string `json:"requestName"`
	Message         string `json:"message"`
	EventMessageName string `json:"eventMessageName"`
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

type ClientOptions struct {
	ServerAddr   string
	Context      string
	ConfigPath   string
	AuthToken    string
	InsecureTLS  bool
	Verbose      bool
	CaCertPaths  string
	ClientId     string
	ClientSecret string
}

type microcksClient struct {
	ServerAddr   string
	APIURL       *url.URL
	AuthToken    string
	CertFile     *tls.Certificate
	InsecureTLS  bool
	RefreshToken string
	Insecure     bool
	Verbose      bool

	httpClient *http.Client
}

func NewClient(opts ClientOptions) (MicrocksClient, error) {
	var c microcksClient
	localCfg, err := config.ReadLocalConfig(opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	var ctxName string

	if localCfg != nil {
		configCtx, err := localCfg.ResolveContext(opts.Context)
		if err != nil {
			return nil, err
		}
		c.ServerAddr = configCtx.Server.Server
		c.Insecure = configCtx.Server.KeycloakEnable
		c.InsecureTLS = configCtx.Server.InsecureTLS
		c.AuthToken = configCtx.User.AuthToken
		c.RefreshToken = configCtx.User.RefreshToken

		apiURL := configCtx.Server.Server

		if strings.HasSuffix(apiURL, "/api") {
			apiURL += "/"
		}
		if !strings.HasSuffix(apiURL, "/api/") {
			apiURL += "/api/"
		}

		u, err := url.Parse(apiURL)
		if err != nil {
			panic(err)
		}
		c.APIURL = u

		ctxName = configCtx.Name
	}

	if opts.Verbose {
		c.Verbose = opts.Verbose
	}

	if config.InsecureTLS || len(config.CaCertPaths) > 0 {
		tlsConfig := config.CreateTLSConfig()
		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		c.httpClient = &http.Client{Transport: tr}
	} else {
		c.httpClient = http.DefaultClient
	}

	if localCfg != nil {
		err = c.refreshAuthToken(localCfg, ctxName, opts.ConfigPath)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// NewMicrocksClient builds a new headless MicrocksClient without any authtoken and all for general purposes
func NewMicrocksClient(apiURL string) MicrocksClient {
	mc := microcksClient{}

	if strings.HasSuffix(apiURL, "/api") {
		apiURL += "/"
	}
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
func (c *microcksClient) HttpClient() *http.Client {
	return c.httpClient
}

func (c *microcksClient) GetKeycloakURL() (string, error) {
	rel := &url.URL{Path: "keycloak/config"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	config.DumpRequestIfRequired("Microcks for getting Keycloak config", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	config.DumpResponseIfRequired("Microcks for getting Keycloak config", resp, true)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var configResp map[string]interface{}
	if err := json.Unmarshal(body, &configResp); err != nil {
		panic(err)
	}

	enabled := configResp["enabled"].(bool)
	authServerURL := configResp["auth-server-url"].(string)
	realmName := configResp["realm"].(string)

	if enabled {
		return authServerURL + "/realms/" + realmName + "/", nil
	}
	return "null", nil
}

func (c *microcksClient) refreshAuthToken(localCfg *config.LocalConfig, ctxName, configPath string) error {
	if c.RefreshToken == "" {
		return nil
	}
	configCtx, err := localCfg.ResolveContext(ctxName)
	if err != nil {
		return err
	}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	var claims jwt.RegisteredClaims
	_, _, err = parser.ParseUnverified(configCtx.User.AuthToken, &claims)
	if err != nil {
		return err
	}
	if claims.Valid() == nil {
		return nil
	}

	log.Printf("Auth token no longer valid. Refreshing")
	auth, err := localCfg.GetAuth(configCtx.Server.Server)
	if err != nil {
		return err
	}
	authToken, refreshToken, err := c.redeemRefreshToken(*auth)
	if err != nil {
		return err
	}
	c.AuthToken = authToken
	c.RefreshToken = refreshToken
	localCfg.UpsertUser(config.User{
		Name:         ctxName,
		AuthToken:    authToken,
		RefreshToken: refreshToken,
	})
	err = config.WriteLocalConfig(*localCfg, configPath)
	if err != nil {
		return err
	}
	return nil
}

func (c *microcksClient) redeemRefreshToken(auth config.Auth) (string, string, error) {
	keyCloakUrl, err := c.GetKeycloakURL()
	errors.CheckError(err)
	kc := NewKeycloakClient(keyCloakUrl, "", "")
	oauth2Conf, err := kc.GetOIDCConfig()
	errors.CheckError(err)
	oauth2Conf.ClientID = auth.ClientId
	oauth2Conf.ClientSecret = auth.ClientSecret

	httpClient := c.httpClient
	ctx := oidc.ClientContext(context.Background(), httpClient)

	t := &oauth2.Token{
		RefreshToken: c.RefreshToken,
	}
	token, err := oauth2Conf.TokenSource(ctx, t).Token()
	if err != nil {
		return "", "", err
	}

	return token.AccessToken, token.RefreshToken, nil
}

func (c *microcksClient) SetOAuthToken(oauthToken string) {
	c.AuthToken = oauthToken
}

func (c *microcksClient) CreateTestResult(serviceID string, testEndpoint string, runnerType string, secretName string, timeout int64, filteredOperations string, operationsHeaders string, oAuth2Context string) (string, error) {
	rel := &url.URL{Path: "tests"}
	u := c.APIURL.ResolveReference(rel)

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
	if len(oAuth2Context) > 0 && ensureValidOAuth2Context(oAuth2Context) {
		input += (", \"oAuth2Context\": " + oAuth2Context)
	}

	input += "}"

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(input))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)

	config.DumpRequestIfRequired("Microcks for creating test", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	config.DumpResponseIfRequired("Microcks for creating test", resp, true)

	body, err := io.ReadAll(resp.Body)
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
	rel := &url.URL{Path: "tests/" + testResultID}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)

	config.DumpRequestIfRequired("Microcks for getting status", req, false)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	config.DumpResponseIfRequired("Microcks for getting status test", resp, true)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	result := TestResultSummary{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse test result response: %w", err)
	}

	return &result, nil
}

func (c *microcksClient) GetFullTestResult(testResultID string) (*TestResult, error) {
	rel := &url.URL{Path: "tests/" + testResultID}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)

	config.DumpRequestIfRequired("Microcks for getting full test result", req, false)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	config.DumpResponseIfRequired("Microcks for getting full test result", resp, true)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := TestResult{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse full test result response: %w", err)
	}

	return &result, nil
}

func (c *microcksClient) UploadArtifact(specificationFilePath string, mainArtifact bool) (string, error) {
	file, err := os.Open(specificationFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Use io.Pipe to stream the multipart form data directly to the HTTP
	// request without buffering the entire file in memory.
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Write the multipart form data in a background goroutine so the pipe
	// reader can be consumed concurrently by the HTTP request.
	errCh := make(chan error, 1)
	go func() {
		defer pw.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(specificationFilePath))
		if err != nil {
			errCh <- err
			return
		}
		if _, err = io.Copy(part, file); err != nil {
			errCh <- err
			return
		}

		// Add the mainArtifact flag to request.
		if err = writer.WriteField("mainArtifact", strconv.FormatBool(mainArtifact)); err != nil {
			errCh <- err
			return
		}

		errCh <- writer.Close()
	}()

	rel := &url.URL{Path: "artifact/upload"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("POST", u.String(), pr)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)

	config.DumpRequestIfRequired("Microcks for uploading artifact", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for errors from the multipart writer goroutine.
	if pipeErr := <-errCh; pipeErr != nil {
		return "", fmt.Errorf("failed to write multipart form: %w", pipeErr)
	}

	// Dump response if verbose required.
	config.DumpResponseIfRequired("Microcks for uploading artifact", resp, true)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload response: %w", err)
	}

	if resp.StatusCode != 201 {
		return "", errs.New(string(respBody))
	}

	return string(respBody), nil
}

func (c *microcksClient) DownloadArtifact(artifactURL string, mainArtifact bool, secret string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("url", artifactURL)
	writer.WriteField("mainArtifact", strconv.FormatBool(mainArtifact))
	if secret != "" {
		writer.WriteField("secret", secret)
	}

	err := writer.Close()
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "artifact/download"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)

	config.DumpRequestIfRequired("Microcks for uploading artifact", req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	config.DumpResponseIfRequired("Microcks for uploading artifact", resp, true)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	if resp.StatusCode != 201 {
		return "", errs.New(string(respBody))
	}

	return string(respBody), err
}

func ensureValidOperationsList(filteredOperations string) bool {
	var list = []string{}
	err := json.Unmarshal([]byte(filteredOperations), &list)
	if err != nil {
		fmt.Println("Error parsing JSON in filteredOperations: ", err)
		return false
	}
	return true
}

func ensureValidOperationsHeaders(operationsHeaders string) bool {
	var headers = map[string][]HeaderDTO{}
	err := json.Unmarshal([]byte(operationsHeaders), &headers)
	if err != nil {
		fmt.Println("Error parsing JSON in operationsHeaders: ", err)
		return false
	}
	return true
}

func ensureValidOAuth2Context(oAuth2Context string) bool {
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