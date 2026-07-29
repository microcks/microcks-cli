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
	"github.com/microcks/microcks-cli/pkg/output"
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
		dryRun             bool
		artifact           string
		image              string
		readyTimeout       time.Duration
		watch              bool
		driver             string
		outputFormat       string
	)
	var testCmd = &cobra.Command{

		Use:   "test <apiName:apiVersion> <testEndpoint> <runner>",
		Short: "Run tests on Microcks",
		Long:  `Run tests on Microcks`,
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {

			serviceRef := args[0]
			testEndpoint := args[1]
			runnerType := args[2]

			// Validate presence and values of args.
			if len(serviceRef) == 0 || strings.HasPrefix(serviceRef, "-") {
				fmt.Fprintln(os.Stderr, "missing required argument: <apiName:apiVersion> (e.g. 'my-api:1.0')")
				os.Exit(1)
			}
			if len(testEndpoint) == 0 || strings.HasPrefix(testEndpoint, "-") {
				fmt.Fprintln(os.Stderr, "missing required argument: <testEndpoint> (e.g. 'http://localhost:8080/api')")
				os.Exit(1)
			}
			if len(runnerType) == 0 || strings.HasPrefix(runnerType, "-") {
				fmt.Fprintln(os.Stderr, "missing required argument: <runner> (e.g. 'HTTP', 'POSTMAN', 'OPEN_API_SCHEMA')")
				os.Exit(1)
			}
			if _, validChoice := runnerChoices[runnerType]; !validChoice {
				fmt.Println("<runner> should be one of: HTTP, SOAP_HTTP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA")
				os.Exit(1)
			}

			// Validate presence and values of flags.
			if !strings.HasSuffix(waitFor, "milli") && !strings.HasSuffix(waitFor, "sec") && !strings.HasSuffix(waitFor, "min") {
				fmt.Println("--waitFor format is wrong. Accepted units are: milli, sec, min (e.g. 500milli, 30sec, 5min)")
				os.Exit(1)
			}

			if !output.IsValid(outputFormat) {
				fmt.Fprintln(os.Stderr, "--output must be one of: text, json, yaml, github-actions")
				os.Exit(1)
			}

			// Collect optional HTTPS transport flags.
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			// Compute time to wait in milliseconds.
			var waitForMilliseconds int64
			if strings.HasSuffix(waitFor, "milli") {
				n, err := strconv.ParseInt(waitFor[:len(waitFor)-5], 0, 64)
				if err != nil {
					fmt.Printf("--waitFor value %q is not a valid number\n", waitFor)
					os.Exit(1)
				}
				waitForMilliseconds = n
			} else if strings.HasSuffix(waitFor, "sec") {
				n, err := strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
				if err != nil {
					fmt.Printf("--waitFor value %q is not a valid number\n", waitFor)
					os.Exit(1)
				}
				waitForMilliseconds = n * 1000
			} else if strings.HasSuffix(waitFor, "min") {
				n, err := strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
				if err != nil {
					fmt.Printf("--waitFor value %q is not a valid number\n", waitFor)
					os.Exit(1)
				}
				waitForMilliseconds = n * 60 * 1000
			}

			params := testParams{
				serviceRef:         serviceRef,
				testEndpoint:       testEndpoint,
				runnerType:         runnerType,
				secretName:         secretName,
				waitForMillis:      waitForMilliseconds,
				filteredOperations: filteredOperations,
				operationsHeaders:  operationsHeaders,
				oAuth2Context:      oAuth2Context,
				outputFormat:       outputFormat,
				artifactPath:       artifact,
			}

			if !dryRun {
				if artifact != "" {
					fmt.Println("--artifact is only valid together with --dry-run")
					os.Exit(1)
				}
				if watch {
					fmt.Println("--watch is only valid together with --dry-run")
					os.Exit(1)
				}
				if driver != "" {
					fmt.Println("--driver is only valid together with --dry-run")
					os.Exit(1)
				}
			}

			if dryRun {
				// Ephemeral path: no server, no Keycloak, no prior import needed.
				if !runDryRunTest(dryRunOptions{
					artifact:     artifact,
					image:        image,
					readyTimeout: readyTimeout,
					watch:        watch,
					driver:       driver,
					params:       params,
				}) {
					os.Exit(1)
				}
				return
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
					os.Exit(1)
				}

				if localConfig == nil {
					fmt.Println("Please login to perform operation...")
					os.Exit(1)
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					fmt.Printf("error %v", err)
					os.Exit(1)
				}

				ctx, err := localConfig.ResolveContext(globalClientOpts.Context)
				errors.CheckError(err)

				serverAddr = ctx.Server.Server
			}

			success, testResultID, err := runTestAndWait(mc, params)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Fprintf(progressWriter(outputFormat), "Full TestResult details are available here: %s/#/tests/%s \n", serverAddr, testResultID)

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
	testCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run the test against an ephemeral local Microcks container instead of a server")
	testCmd.Flags().StringVar(&artifact, "artifact", "", "Local spec file to import on the ephemeral server (required with --dry-run)")
	testCmd.Flags().StringVar(&image, "image", defaultDryRunImage, "Microcks uber-native image used for --dry-run")
	testCmd.Flags().DurationVar(&readyTimeout, "ready-timeout", 90*time.Second, "How long to wait for the ephemeral container to be ready (--dry-run only)")
	testCmd.Flags().BoolVar(&watch, "watch", false, "Watch the artifact file and re-run the test on change (--dry-run only)")
	testCmd.Flags().StringVar(&driver, "driver", "", "Container runtime for --dry-run: 'docker' or 'podman' (default: auto-detect)")
	testCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text, json, yaml, or github-actions")

	return testCmd
}

