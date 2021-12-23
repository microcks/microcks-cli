package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
)

var (
	runnerChoices   = map[string]bool{"HTTP": true, "SOAP_HTTP": true, "SOAP_UI": true, "POSTMAN": true, "OPEN_API_SCHEMA": true, "ASYNC_API_SCHEMA": true, "GRPC_PROTOBUF": true, "GRAPHQL_SCHEMA": true}
	timeUnitChoices = map[string]bool{"milli": true, "sec": true, "min": true}
)

type testComamnd struct {
}

// NewTestCommand build a new TestCommand implementation
func NewTestCommand() Command {
	return new(testComamnd)
}

// Execute implementation of testCommand structure
func (c *testComamnd) Execute() {

	// Parse subcommand args first.
	if len(os.Args) < 4 {
		fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
		os.Exit(1)
	}

	serviceRef := os.Args[2]
	testEndpoint := os.Args[3]
	runnerType := os.Args[4]

	// Validate presence and values of args.
	if &serviceRef == nil || strings.HasPrefix(serviceRef, "-") {
		fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
		os.Exit(1)
	}
	if &testEndpoint == nil || strings.HasPrefix(testEndpoint, "-") {
		fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
		os.Exit(1)
	}
	if &runnerType == nil || strings.HasPrefix(runnerType, "-") {
		fmt.Println("test command require <apiName:apiVersion> <testEndpoint> <runner> args")
		os.Exit(1)
	}
	if _, validChoice := runnerChoices[runnerType]; !validChoice {
		fmt.Println("<runner> should be one of: HTTP, SOAP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA")
		os.Exit(1)
	}

	// Then parse flags.
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)

	var microcksURL string
	var keycloakURL string
	var keycloakClientID string
	var keycloakClientSecret string
	var waitFor string
	var secretName string
	var operationsHeaders string
	var insecureTLS bool
	var caCertPaths string
	var verbose bool

	testCmd.StringVar(&microcksURL, "microcksURL", "", "Microcks API URL")
	testCmd.StringVar(&keycloakClientID, "keycloakClientId", "", "Keycloak Realm Service Account ClientId")
	testCmd.StringVar(&keycloakClientSecret, "keycloakClientSecret", "", "Keycloak Realm Service Account ClientSecret")
	testCmd.StringVar(&waitFor, "waitFor", "5sec", "Time to wait for test to finish")
	testCmd.StringVar(&secretName, "secretName", "", "Secret to use for connecting test endpoint")
	testCmd.StringVar(&operationsHeaders, "operationsHeaders", "", "Override of operations headers as JSON string")
	testCmd.BoolVar(&insecureTLS, "insecure", false, "Whether to accept insecure HTTPS connection")
	testCmd.StringVar(&caCertPaths, "caCerts", "", "Comma separated paths of CRT files to add to Root CAs")
	testCmd.BoolVar(&verbose, "verbose", false, "Produce dumps of HTTP exchanges")
	testCmd.Parse(os.Args[5:])

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

	// Second - retrieve an OAuth token using Keycloak Client.
	kc := connectors.NewKeycloakClient(keycloakURL, keycloakClientID, keycloakClientSecret)

	var oauthToken string
	oauthToken, err = kc.ConnectAndGetToken()
	if err != nil {
		fmt.Printf("Got error when invoking Keycloack client: %s", err)
		os.Exit(1)
	}
	//fmt.Printf("Retrieve OAuthToken: %s", oauthToken)

	// Then - launch the test on Microcks Server.
	mc.SetOAuthToken(oauthToken)

	var testResultID string
	testResultID, err = mc.CreateTestResult(serviceRef, testEndpoint, runnerType, secretName, waitForMilliseconds, operationsHeaders)
	if err != nil {
		fmt.Printf("Got error when invoking Microcks client creating Test: %s", err)
		os.Exit(1)
	}
	//fmt.Printf("Retrieve TestResult ID: %s", testResultID)

	// Finally - wait for some time
	// Add 500ms to wait time as it's now representing the server timeout.
	now := nowInMilliseconds()
	future := now + waitForMilliseconds + 500

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

	if !success {
		os.Exit(1)
	}
}

func nowInMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
