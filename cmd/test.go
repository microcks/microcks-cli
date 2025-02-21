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
package cmd

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
	Short: "Execute a test on a Microcks server",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		serviceRef, testEndpoint, runnerType := args[0], args[1], args[2]

// Validate presence and values of args.
if strings.HasPrefix(serviceRef, "-") || serviceRef == "" {
			fmt.Println("test command requires <apiName:apiVersion> <testEndpoint> <runner> args")
			os.Exit(1)
		}
		if strings.HasPrefix(testEndpoint, "-") || testEndpoint == "" {
			fmt.Println("test command requires <apiName:apiVersion> <testEndpoint> <runner> args")
			os.Exit(1)
		}
		if strings.HasPrefix(runnerType, "-") || runnerType == "" {
			fmt.Println("test command requires <apiName:apiVersion> <testEndpoint> <runner> args")
			os.Exit(1)
		}
		if _, validChoice := runnerChoices[runnerType]; !validChoice {
			fmt.Println("<runner> should be one of: HTTP, SOAP_HTTP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA")
			os.Exit(1)
		}

	// Then parse flags.
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

	// Validate presence and values of flags.
	if len(microcksURL) == 0 {
		fmt.Println("--microcksURL flag is mandatory. Check Usage.")
		os.Exit(1)
	}
	if len(keycloakClientID) == 0 {
		fmt.Println("--keycloakClientId flag is mandatory. Check Usage.")
		os.Exit(1)
	}
	if len(keycloakClientSecret) == 0 {
		fmt.Println("--keycloakClientSecret flag is mandatory. Check Usage.")
		os.Exit(1)
	}
	if &waitFor == nil || (!strings.HasSuffix(waitFor, "milli") && !strings.HasSuffix(waitFor, "sec") && !strings.HasSuffix(waitFor, "min")) {
		fmt.Println("--waitFor format is wrong. Applying default 5sec")
		waitFor = "5sec"
	}

	// Collect optional HTTPS transport flags.
	if insecureTLS {
		config.InsecureTLS = true
	}
	if len(caCertPaths) > 0 {
		config.CaCertPaths = caCertPaths
	}
	if verbose {
		config.Verbose = true
	}

	// Compute time to wait in milliseconds.
	var waitForMilliseconds int64 = 5000
	if strings.HasSuffix(waitFor, "milli") {
		waitForMilliseconds, _ = strconv.ParseInt(waitFor[:len(waitFor)-5], 0, 64)
	} else if strings.HasSuffix(waitFor, "sec") {
		waitForMilliseconds, _ = strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
		waitForMilliseconds = waitForMilliseconds * 1000
	} else if strings.HasSuffix(waitFor, "min") {
		waitForMilliseconds, _ = strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
		waitForMilliseconds = waitForMilliseconds * 60 * 1000
	}

	// Now we seems to be good ...
	// First - retrieve the Keycloak URL from Microcks configuration.
	mc := connectors.NewMicrocksClient(microcksURL)
	keycloakURL, err := mc.GetKeycloakURL()
	if err != nil {
		fmt.Printf("Got error when invoking Microcks client retrieving config: %s", err)
		os.Exit(1)
	}

	var oauthToken string = "unauthentifed-token"
	if keycloakURL != "null" {
		// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
		kc := connectors.NewKeycloakClient(keycloakURL, keycloakClientID, keycloakClientSecret)

		oauthToken, err = kc.ConnectAndGetToken()
		if err != nil {
			fmt.Printf("Got error when invoking Keycloack client: %s", err)
			os.Exit(1)
		}
		//fmt.Printf("Retrieve OAuthToken: %s", oauthToken)
	}

	// Then - launch the test on Microcks Server.
	mc.SetOAuthToken(oauthToken)

	var testResultID string
	testResultID, err = mc.CreateTestResult(serviceRef, testEndpoint, runnerType, secretName, waitForMilliseconds, filteredOperations, operationsHeaders, oAuth2Context)
	if err != nil {
		fmt.Printf("Got error when invoking Microcks client creating Test: %s", err)
		os.Exit(1)
	}
	//fmt.Printf("Retrieve TestResult ID: %s", testResultID)

	// Finally - wait before checking and loop for some time
	time.Sleep(1 * time.Second)

	// Add 10.000ms to wait time as it's now representing the server timeout.
	now := nowInMilliseconds()
	future := now + waitForMilliseconds + 10000

	var success = false
	for nowInMilliseconds() < future {
		testResultSummary, err := mc.GetTestResult(testResultID)
		if err != nil {
			fmt.Printf("Got error when invoking Microcks client check TestResult: %s", err)
			os.Exit(1)
		}
		success = testResultSummary.Success
		inProgress := testResultSummary.InProgress
		fmt.Printf("MicrocksClient got status for test \"%s\" - success: %s, inProgress: %s \n", testResultID, fmt.Sprint(success), fmt.Sprint(inProgress))

		if !inProgress {
			break
		}

		fmt.Println("MicrocksTester waiting for 2 seconds before checking again or exiting.")
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("Full TestResult details are available here: %s/#/tests/%s \n", strings.Split(microcksURL, "/api")[0], testResultID)

	if !success {
		os.Exit(1)
	}
},
}

func nowInMilliseconds() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().String("microcksURL", "", "Microcks API URL (required)")
	testCmd.Flags().String("keycloakClientId", "", "Keycloak Client ID (required)")
	testCmd.Flags().String("keycloakClientSecret", "", "Keycloak Client Secret (required)")
	testCmd.Flags().String("waitFor", "5sec", "Time to wait for test completion")
	testCmd.Flags().String("secretName", "", "Secret for connecting to the test endpoint")
	testCmd.Flags().String("filteredOperations", "", "List of operations to test")
	testCmd.Flags().String("operationsHeaders", "", "Override operations headers as JSON")
	testCmd.Flags().String("oAuth2Context", "", "OAuth2 client context as JSON")
	testCmd.Flags().Bool("insecure", false, "Allow insecure HTTPS connections")
	testCmd.Flags().String("caCerts", "", "Comma-separated paths to CA certificates")
	testCmd.Flags().Bool("verbose", false, "Enable verbose output")
}
