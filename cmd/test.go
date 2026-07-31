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
		dryRun             bool
		artifact           string
		image              string
		readyTimeout       time.Duration
		watch              bool
		driver             string
	)
	var testCmd = &cobra.Command{

		Use:   "test <apiName:apiVersion> <testEndpoint> <runner>",
		Short: "Run tests on Microcks",
		Long:  `Run tests on Microcks`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			serviceRef := args[0]
			testEndpoint := args[1]
			runnerType := args[2]

			// Validate presence and values of args.
			if len(serviceRef) == 0 || strings.HasPrefix(serviceRef, "-") {
				return errors.Wrapf(errors.KindUsage, "missing required argument: <apiName:apiVersion> (e.g. 'my-api:1.0')")
			}
			if len(testEndpoint) == 0 || strings.HasPrefix(testEndpoint, "-") {
				return errors.Wrapf(errors.KindUsage, "missing required argument: <testEndpoint> (e.g. 'http://localhost:8080/api')")
			}
			if len(runnerType) == 0 || strings.HasPrefix(runnerType, "-") {
				return errors.Wrapf(errors.KindUsage, "missing required argument: <runner> (e.g. 'HTTP', 'POSTMAN', 'OPEN_API_SCHEMA')")
			}
			if _, validChoice := runnerChoices[runnerType]; !validChoice {
				return errors.Wrapf(errors.KindUsage, "<runner> should be one of: HTTP, SOAP_HTTP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA")
			}

			// Validate presence and values of flags.
			if !strings.HasSuffix(waitFor, "milli") && !strings.HasSuffix(waitFor, "sec") && !strings.HasSuffix(waitFor, "min") {
				return errors.Wrapf(errors.KindUsage, "--waitFor format is wrong. Accepted units are: milli, sec, min (e.g. 500milli, 30sec, 5min)")
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
					return errors.Wrapf(errors.KindUsage, "--waitFor value %q is not a valid number", waitFor)
				}
				waitForMilliseconds = n
			} else if strings.HasSuffix(waitFor, "sec") {
				n, err := strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
				if err != nil {
					return errors.Wrapf(errors.KindUsage, "--waitFor value %q is not a valid number", waitFor)
				}
				waitForMilliseconds = n * 1000
			} else if strings.HasSuffix(waitFor, "min") {
				n, err := strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
				if err != nil {
					return errors.Wrapf(errors.KindUsage, "--waitFor value %q is not a valid number", waitFor)
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
			}

			if !dryRun {
				if artifact != "" {
					return errors.Wrapf(errors.KindUsage, "--artifact is only valid together with --dry-run")
				}
				if watch {
					return errors.Wrapf(errors.KindUsage, "--watch is only valid together with --dry-run")
				}
				if driver != "" {
					return errors.Wrapf(errors.KindUsage, "--driver is only valid together with --dry-run")
				}
			}

			if dryRun {
				// Ephemeral path: no server, no Keycloak, no prior import needed.
				return runDryRunTest(dryRunOptions{
					artifact:     artifact,
					image:        image,
					readyTimeout: readyTimeout,
					watch:        watch,
					driver:       driver,
					params:       params,
				})
			}

			var mc connectors.MicrocksClient
			var serverAddr string

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {

				// create client with server address
				serverAddr = globalClientOpts.ServerAddr
				var err error
				mc, err = connectors.NewMicrocksClient(serverAddr)
				if err != nil {
					return err
				}

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					return err
				}

				oauthToken := "unauthenticated-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc, err := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)
					if err != nil {
						return err
					}

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						return err
					}
				}

				// Then - launch the test on Microcks Server.
				mc.SetOAuthToken(oauthToken)

			} else {
				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					return err
				}

				if localConfig == nil {
					return errors.Wrapf(errors.KindUsage, "please login to perform this operation")
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					return err
				}

				ctx, err := localConfig.ResolveContext(globalClientOpts.Context)
				if err != nil {
					return errors.Wrap(errors.KindNotFound, err)
				}

				serverAddr = ctx.Server.Server
			}

			success, testResultID, err := runTestAndWait(mc, params)
			if err != nil {
				return err
			}

			fmt.Printf("Full TestResult details are available here: %s/#/tests/%s \n", serverAddr, testResultID)

			if !success {
				return errors.ErrTestFailed
			}
			return nil
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

	return testCmd
}
