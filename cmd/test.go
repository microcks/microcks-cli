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
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	runnerChoices = map[string]bool{"HTTP": true, "SOAP_HTTP": true, "SOAP_UI": true, "POSTMAN": true, "OPEN_API_SCHEMA": true, "ASYNC_API_SCHEMA": true, "GRPC_PROTOBUF": true, "GRAPHQL_SCHEMA": true}
)

const testCommandUsageError = "test command require <apiName:apiVersion> <testEndpoint> <runner> args"
const testCommandRunnerError = "<runner> should be one of: HTTP, SOAP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA"
const testCommandWaitForFormatError = "--waitFor format is wrong. Accepted units are: milli, sec, min (e.g. 500milli, 30sec, 5min)"

func NewTestCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		waitFor            string
		secretName         string
		filteredOperations string
		operationsHeaders  string
		oAuth2Context      string
	)
	var testCmd = &cobra.Command{

		Use:   "test",
		Short: "Run tests on Microcks",
		Long:  `Run tests on Microcks`,
		Run: func(cmd *cobra.Command, args []string) {
			serviceRef, testEndpoint, runnerType, err := validateTestCommandArgs(args)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			// Collect optional HTTPS transport flags.
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			waitForMilliseconds, err := parseWaitForMilliseconds(waitFor)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			var mc connectors.MicrocksClient
			var serverAddr string

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {

				// create client with server address
				serverAddr = globalClientOpts.ServerAddr
				mc = connectors.NewMicrocksClient(serverAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					fmt.Printf("Got error when invoking Microcks client retrieving config: %s", err)
					os.Exit(1)
				}

				var oauthToken string = "unauthenticated-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						fmt.Printf("Got error when invoking Keycloak client: %s", err)
						os.Exit(1)
					}
					//fmt.Printf("Retrieve OAuthToken: %s", oauthToken)
				}

				// Then - launch the test on Microcks Server.
				mc.SetOAuthToken(oauthToken)

			} else {
				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					fmt.Println(err)
					return
				}

				if localConfig == nil {
					fmt.Println("Please login to perform operation...")
					return
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					fmt.Printf("error %v", err)
					return
				}

				ctx, err := localConfig.ResolveContext(globalClientOpts.Context)
				errors.CheckError(err)

				serverAddr = ctx.Server.Server
			}

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

			fmt.Printf("Full TestResult details are available here: %s/#/tests/%s \n", serverAddr, testResultID)

			if !success {
				os.Exit(1)
			}
		},
	}

	testCmd.Flags().StringVar(&waitFor, "waitFor", "5sec", "Time to wait for test to finish")
	testCmd.Flags().StringVar(&secretName, "secretName", "", "Secret to use for connecting test endpoint")
	testCmd.Flags().StringVar(&filteredOperations, "filteredOperations", "", "List of operations to launch a test for")
	testCmd.Flags().StringVar(&operationsHeaders, "operationsHeaders", "", "Override of operations headers as JSON string")
	testCmd.Flags().StringVar(&oAuth2Context, "oAuth2Context", "", "Spec of an OAuth2 client context as JSON string")

	return testCmd
}

func validateTestCommandArgs(args []string) (string, string, string, error) {
	if len(os.Args) < 4 || len(args) < 3 {
		return "", "", "", fmt.Errorf(testCommandUsageError)
	}

	serviceRef := args[0]
	testEndpoint := args[1]
	runnerType := args[2]

	if len(serviceRef) == 0 || strings.HasPrefix(serviceRef, "-") {
		return "", "", "", fmt.Errorf(testCommandUsageError)
	}
	if len(testEndpoint) == 0 || strings.HasPrefix(testEndpoint, "-") {
		return "", "", "", fmt.Errorf(testCommandUsageError)
	}
	if len(runnerType) == 0 || strings.HasPrefix(runnerType, "-") {
		return "", "", "", fmt.Errorf(testCommandUsageError)
	}
	if _, validChoice := runnerChoices[runnerType]; !validChoice {
		return "", "", "", fmt.Errorf(testCommandRunnerError)
	}

	return serviceRef, testEndpoint, runnerType, nil
}

func parseWaitForMilliseconds(waitFor string) (int64, error) {
	if !strings.HasSuffix(waitFor, "milli") && !strings.HasSuffix(waitFor, "sec") && !strings.HasSuffix(waitFor, "min") {
		return 0, fmt.Errorf(testCommandWaitForFormatError)
	}

	if strings.HasSuffix(waitFor, "milli") {
		return parseWaitForValue(waitFor, "milli", 1)
	}
	if strings.HasSuffix(waitFor, "sec") {
		return parseWaitForValue(waitFor, "sec", 1000)
	}
	return parseWaitForValue(waitFor, "min", 60*1000)
}

func parseWaitForValue(waitFor string, suffix string, multiplier int64) (int64, error) {
	n, err := strconv.ParseInt(waitFor[:len(waitFor)-len(suffix)], 0, 64)
	if err != nil {
		return 0, fmt.Errorf("--waitFor value %q is not a valid number", waitFor)
	}
	return n * multiplier, nil
}

func nowInMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
