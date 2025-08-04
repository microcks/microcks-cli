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
			// Parse subcommand args first.
			if len(os.Args) < 4 {
				fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
				os.Exit(1)
			}

			serviceRef := args[0]
			testEndpoint := args[1]
			runnerType := args[2]

			// Validate presence and values of args.
			if len(serviceRef) == 0 || strings.HasPrefix(serviceRef, "-") {
				fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
				os.Exit(1)
			}
			if len(testEndpoint) == 0 || strings.HasPrefix(testEndpoint, "-") {
				fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
				os.Exit(1)
			}
			if len(runnerType) == 0 || strings.HasPrefix(runnerType, "-") {
				fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
				os.Exit(1)
			}
			if _, validChoice := runnerChoices[runnerType]; !validChoice {
				fmt.Println("<runner> should be one of: HTTP, SOAP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA")
				os.Exit(1)
			}

			// Validate presence and values of flags.
			if !strings.HasSuffix(waitFor, "milli") && !strings.HasSuffix(waitFor, "sec") && !strings.HasSuffix(waitFor, "min") {
				fmt.Println("--waitFor format is wrong. Applying default 5sec")
			}

			// Collect optional HTTPS transport flags.
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

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

				var oauthToken string = "unauthentifed-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						fmt.Printf("Got error when invoking Keycloack client: %s", err)
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
					fmt.Println("Please login to perform opertion...")
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
			testResultID, err := mc.CreateTestResult(serviceRef, testEndpoint, runnerType, secretName, waitForMilliseconds, filteredOperations, operationsHeaders, oAuth2Context)
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

func nowInMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
