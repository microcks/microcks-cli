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
package root

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

var (
	runnerChoices   = map[string]bool{"HTTP": true, "SOAP_HTTP": true, "SOAP_UI": true, "POSTMAN": true, "OPEN_API_SCHEMA": true, "ASYNC_API_SCHEMA": true, "GRPC_PROTOBUF": true, "GRAPHQL_SCHEMA": true}
	timeUnitChoices = map[string]bool{"milli": true, "sec": true, "min": true}
)

var testCmd = &cobra.Command{
	Use:   "test <apiName:apiVersion> <testEndpoint> <runner>",
	Short: "Run tests on Microcks",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		serviceRef := args[0]
		testEndpoint := args[1]
		runnerType := args[2]

		if _, valid := runnerChoices[runnerType]; !valid {
			fmt.Println("Invalid runner type. Should be one of:", keys(runnerChoices))
			os.Exit(1)
		}

		// Retrieve flag values
		microcksURL, _ := cmd.Flags().GetString("microcksURL")
		keycloakClientID, _ := cmd.Flags().GetString("keycloakClientId")
		keycloakClientSecret, _ := cmd.Flags().GetString("keycloakClientSecret")
		waitFor, _ := cmd.Flags().GetString("waitFor")
		secretName, _ := cmd.Flags().GetString("secretName")
		filteredOperations, _ := cmd.Flags().GetString("filteredOperations")
		operationsHeaders, _ := cmd.Flags().GetString("operationsHeaders")
		oAuth2Context, _ := cmd.Flags().GetString("oAuth2Context")
		insecureTLS, _ := cmd.Flags().GetBool("insecure")
		caCertPaths, _ := cmd.Flags().GetString("caCerts")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if microcksURL == "" || keycloakClientID == "" || keycloakClientSecret == "" {
			fmt.Println("--microcksURL, --keycloakClientId, and --keycloakClientSecret are mandatory flags.")
			os.Exit(1)
		}

		if insecureTLS {
			config.InsecureTLS = true
		}
		if caCertPaths != "" {
			config.CaCertPaths = caCertPaths
		}
		if verbose {
			config.Verbose = true
		}

		waitForMilliseconds := parseWaitFor(waitFor)
		mc := connectors.NewMicrocksClient(microcksURL)
		keycloakURL, err := mc.GetKeycloakURL()
		if err != nil {
			fmt.Printf("Error retrieving Microcks config: %s", err)
			os.Exit(1)
		}

		oauthToken := "unauthenticated-token"
		if keycloakURL != "null" {
			kc := connectors.NewKeycloakClient(keycloakURL, keycloakClientID, keycloakClientSecret)
			oauthToken, err = kc.ConnectAndGetToken()
			if err != nil {
				fmt.Printf("Error retrieving OAuth token: %s", err)
				os.Exit(1)
			}
		}

		mc.SetOAuthToken(oauthToken)
		testResultID, err := mc.CreateTestResult(serviceRef, testEndpoint, runnerType, secretName, waitForMilliseconds, filteredOperations, operationsHeaders, oAuth2Context)
		if err != nil {
			fmt.Printf("Error creating test: %s", err)
			os.Exit(1)
		}

		time.Sleep(1 * time.Second)
		success := pollTestResult(mc, testResultID, waitForMilliseconds)

		fmt.Printf("Full TestResult details: %s/#/tests/%s\n", strings.Split(microcksURL, "/api")[0], testResultID)
		if !success {
			os.Exit(1)
		}
	},
}

func init() {
	testCmd.Flags().String("microcksURL", "", "Microcks API URL")
	testCmd.Flags().String("keycloakClientId", "", "Keycloak Client ID")
	testCmd.Flags().String("keycloakClientSecret", "", "Keycloak Client Secret")
	testCmd.Flags().String("waitFor", "5sec", "Time to wait for test to finish")
	testCmd.Flags().String("secretName", "", "Secret for connecting test endpoint")
	testCmd.Flags().String("filteredOperations", "", "List of operations for test")
	testCmd.Flags().String("operationsHeaders", "", "Override operations headers as JSON")
	testCmd.Flags().String("oAuth2Context", "", "OAuth2 client context JSON")
	testCmd.Flags().Bool("insecure", false, "Accept insecure HTTPS connections")
	testCmd.Flags().String("caCerts", "", "Comma-separated CRT file paths")
	testCmd.Flags().Bool("verbose", false, "Enable verbose mode")
}

func parseWaitFor(waitFor string) int64 {
	if strings.HasSuffix(waitFor, "milli") {
		ms, _ := strconv.ParseInt(waitFor[:len(waitFor)-5], 10, 64)
		return ms
	} else if strings.HasSuffix(waitFor, "sec") {
		sec, _ := strconv.ParseInt(waitFor[:len(waitFor)-3], 10, 64)
		return sec * 1000
	} else if strings.HasSuffix(waitFor, "min") {
		min, _ := strconv.ParseInt(waitFor[:len(waitFor)-3], 10, 64)
		return min * 60 * 1000
	}
	return 5000
}

func pollTestResult(mc connectors.MicrocksClient, testResultID string, waitTime int64) bool {
	now := time.Now().UnixMilli()
	future := now + waitTime + 10000
	for time.Now().UnixMilli() < future {
		summary, err := mc.GetTestResult(testResultID)
		if err != nil {
			fmt.Printf("Error checking test result: %s", err)
			os.Exit(1)
		}
		if !summary.InProgress {
			return summary.Success
		}
		fmt.Println("Waiting 2 seconds before checking again...")
		time.Sleep(2 * time.Second)
	}
	return false
}

func keys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
