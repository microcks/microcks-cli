package connectors

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
)

// KeycloakClient defines methods for cinteracting with Keycloak
type KeycloakClient interface {
	ConnectAndGetToken() (string, error)
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
