package connectors

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// MicrocksClient allows interacting with Mcirocks APIs
type MicrocksClient interface {
	CreateTestResult(serviceID string, testEndpoint string, runnerType string) (string, error)
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

type microcksClient struct {
	APIURL     *url.URL
	OAuthToken string

	httpClient *http.Client
}

// NewMicrocksClient build a new MicrocksClient implementation
func NewMicrocksClient(apiURL string, oauthToken string) MicrocksClient {
	mc := microcksClient{}

	u, err := url.Parse(apiURL)
	if err != nil {
		panic(err)
	}
	mc.APIURL = u
	mc.OAuthToken = oauthToken
	mc.httpClient = http.DefaultClient
	return &mc
}

func (c *microcksClient) CreateTestResult(serviceID string, testEndpoint string, runnerType string) (string, error) {
	// Ensure we have a correct URL.
	rel := &url.URL{Path: "tests"}
	u := c.APIURL.ResolveReference(rel)

	// Prepare an input string as body.
	var input = "{"
	input += ("\"serviceId\": \"" + serviceID + "\", ")
	input += ("\"testEndpoint\": \"" + testEndpoint + "\", ")
	input += ("\"runnerType\": \"" + runnerType + "\"")
	input += "}"

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(input))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.OAuthToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	result := TestResultSummary{}
	json.Unmarshal([]byte(body), &result)

	return &result, err
}
