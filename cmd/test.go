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
	runnerChoices = map[string]bool{"HTTP": true, "SOAP_HTTP": true, "SOAP_UI": true, "POSTMAN": true, "OPEN_API_SCHEMA": true, "ASYNC_API_SCHEMA": true, "GRPC_PROBUF": true, "GRAPHQL_SCHEMA": true}
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

		Use:   "test <apiName:apiVersion> <testEndpoint> <runner>",
		Short: "Run tests on Microcks",
		Long:  `Run tests on Microcks`,
		Args:  cobra.ExactArgs(3),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			runnerType := args[2]
			if _, validChoice := runnerChoices[runnerType]; !validChoice {
				return fmt.Errorf("<runner> should be one of: HTTP, SOAP_HTTP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROBUF, GRAPHQL_SCHEMA")
			}
			if !strings.HasSuffix(waitFor, "milli") && !strings.HasSuffix(waitFor, "sec") && !strings.HasSuffix(waitFor, "min") {
				fmt.Println("--waitFor format is wrong. Applying default 5sec")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceRef := args[0]
			testEndpoint := args[1]
			runnerType := args[2]

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

				serverAddr = globalClientOpts.ServerAddr
				mc = connectors.NewMicrocksClient(serverAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					return fmt.Errorf("got error when invoking Microcks client retrieving config: %s", err)
				}

				var oauthToken string = "unauthenticated-token"
				if keycloakURL != "null" {
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						return fmt.Errorf("got error when invoking Keycloak client: %s", err)
					}
				}

				mc.SetOAuthToken(oauthToken)

			} else {
				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					return err
				}

				if localConfig == nil {
					return fmt.Errorf("please login to perform operation")
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					return err
				}

				ctx, err := localConfig.ResolveContext(globalClientOpts.Context)
				errors.CheckError(err)

				serverAddr = ctx.Server.Server
			}

			testResultID, err := mc.CreateTestResult(serviceRef, testEndpoint, runnerType, secretName, waitForMilliseconds, filteredOperations, operationsHeaders, oAuth2Context)
			if err != nil {
				return fmt.Errorf("got error when invoking Microcks client creating Test: %s", err)
			}

			time.Sleep(1 * time.Second)

			now := nowInMilliseconds()
			future := now + waitForMilliseconds + 10000

			var success = false
			for nowInMilliseconds() < future {
				testResultSummary, err := mc.GetTestResult(testResultID)
				if err != nil {
					return fmt.Errorf("got error when invoking Microcks client check TestResult: %s", err)
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
				return fmt.Errorf("test failed")
			}
			return nil
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
