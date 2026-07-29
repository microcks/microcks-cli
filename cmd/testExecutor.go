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
	"io"
	"os"
	"time"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/output"
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
	outputFormat       string
	artifactPath       string
}

// progressWriter returns where human progress/diagnostics should go. For
// machine-readable output formats they go to stderr, leaving stdout for the
// formatted result only.
func progressWriter(format string) io.Writer {
	if format != "" && output.OutputFormat(format) != output.FormatText {
		return os.Stderr
	}
	return os.Stdout
}

// runTestAndWait creates a test on the Microcks server, polls until completion
// or timeout, then renders the result in the requested output format (result to
// stdout, progress to stderr for machine formats). Shared by the regular and
// --dry-run paths.
func runTestAndWait(mc connectors.MicrocksClient, params testParams) (bool, string, error) {
	progress := progressWriter(params.outputFormat)

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
		fmt.Fprintf(progress, "MicrocksClient got status for test \"%s\" - success: %s, inProgress: %s \n", testResultID, fmt.Sprint(success), fmt.Sprint(inProgress))

		if !inProgress {
			break
		}

		fmt.Fprintln(progress, "MicrocksTester waiting for 2 seconds before checking again or exiting.")
		time.Sleep(2 * time.Second)
	}

	if err := renderTestResult(mc, testResultID, params.outputFormat, params.artifactPath); err != nil {
		return false, testResultID, err
	}

	return success, testResultID, nil
}

// renderTestResult fetches the full result and writes it to stdout in the
// requested format. artifactPath (when set) lets the github-actions formatter
// map failures to file:line.
func renderTestResult(mc connectors.MicrocksClient, testResultID, format, artifactPath string) error {
	if format == "" {
		format = string(output.FormatText)
	}
	full, err := mc.GetFullTestResult(testResultID)
	if err != nil {
		return fmt.Errorf("Got error when retrieving full test result: %s", err)
	}
	formatter, err := output.NewFormatter(output.OutputFormat(format), output.WithArtifactPath(artifactPath))
	if err != nil {
		return err
	}
	rendered, err := formatter.Format(full)
	if err != nil {
		return fmt.Errorf("Got error when formatting test result: %s", err)
	}
	if rendered != "" {
		fmt.Println(rendered)
	}
	return nil
}

func nowInMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
