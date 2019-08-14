package connectors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
)

// MicrocksClient allows interacting with Mcirocks APIs
type MicrocksClient interface {
	GetKeycloakURL() (string, error)
	SetOAuthToken(oauthToken string)
	CreateTestResult(serviceID string, testEndpoint string, runnerType string, operationsHeaders string) (string, error)
	GetTestResult(testResultID string) (*TestResultSummary, error)
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

type microcksClient struct {
	APIURL     *url.URL
	OAuthToken string

	httpClient *http.Client
}

// NewMicrocksClient build a new MicrocksClient implementation
func NewMicrocksClient(apiURL string) MicrocksClient {
	mc := microcksClient{}

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

func (c *microcksClient) GetKeycloakURL() (string, error) {
	// Ensure we have a correct URL for retrieving Keycloal configuration.
	rel := &url.URL{Path: "keycloak/config"}
	u := c.APIURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var configResp map[string]interface{}
	if err := json.Unmarshal(body, &configResp); err != nil {
		panic(err)
	}

	// Retrieve auth server url and realm name.
	authServerURL := configResp["auth-server-url"].(string)
	realmName := configResp["realm"].(string)

	// Return a proper URL.
	return authServerURL + "/realms/" + realmName + "/", nil
}

func (c *microcksClient) SetOAuthToken(oauthToken string) {
	c.OAuthToken = oauthToken
}

func (c *microcksClient) CreateTestResult(serviceID string, testEndpoint string, runnerType string, operationsHeaders string) (string, error) {
	// Ensure we have a correct URL.
	rel := &url.URL{Path: "tests"}
	u := c.APIURL.ResolveReference(rel)

	// Prepare an input string as body.
	var input = "{"
	input += ("\"serviceId\": \"" + serviceID + "\", ")
	input += ("\"testEndpoint\": \"" + testEndpoint + "\", ")
	input += ("\"runnerType\": \"" + runnerType + "\"")
	if len(operationsHeaders) > 0 && ensureValid(operationsHeaders) {
		input += (", \"operationsHeaders\": " + operationsHeaders)
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

func ensureValid(operationsHeaders string) bool {
	// Unmarshal using a generic interface
	var headers = map[string][]HeaderDTO{}
	err := json.Unmarshal([]byte(operationsHeaders), &headers)
	if err != nil {
		fmt.Println("Error parsing JSON in operationsHeaders: ", err)
		return false
	}
	return true
}
