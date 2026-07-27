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
	"time"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

// testParams bundles the inputs needed to launch and poll a Microcks test.
// Shared by the regular server path and the --dry-run ephemeral path.
type testParams struct {
	serviceRef         string
	testEndpoint       string
	runnerType         string
	secretName         string
	waitForMillis      int64
	filteredOperations string
	operationsHeaders  string
	oAuth2Context      string
}

// runTestAndWait creates a test on the Microcks server and polls its result
// until completion or timeout. Shared by the regular and --dry-run paths.
func runTestAndWait(mc connectors.MicrocksClient, params testParams) (bool, string, error) {
	testResultID, err := mc.CreateTestResult(params.serviceRef, params.testEndpoint, params.runnerType, params.secretName,
		params.waitForMillis, params.filteredOperations, params.operationsHeaders, params.oAuth2Context)
	if err != nil {
		return false, "", fmt.Errorf("Got error when invoking Microcks client creating Test: %s", err)
	}

	// Finally - wait before checking and loop for some time
	time.Sleep(1 * time.Second)

	// Add 10.000ms to wait time as it's now representing the server timeout.
	now := nowInMilliseconds()
	future := now + params.waitForMillis + 10000

	var success = false
	for nowInMilliseconds() < future {
		testResultSummary, err := mc.GetTestResult(testResultID)
		if err != nil {
			return false, "", fmt.Errorf("Got error when invoking Microcks client check TestResult: %s", err)
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

	return success, testResultID, nil
}

func nowInMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
