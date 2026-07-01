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
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

// GitHubActionsFormatter renders the result as GitHub Actions workflow commands:
// a collapsible ::group:: per operation, ::error:: annotations for failures
// (and ::notice:: for passes when MICROCKS_ACTIONS_VERBOSE is set), plus a
// markdown table appended to $GITHUB_STEP_SUMMARY.
type GitHubActionsFormatter struct {
	artifactPath string
}

func (f *GitHubActionsFormatter) Format(r *connectors.TestResult) (string, error) {
	verbose := os.Getenv("MICROCKS_ACTIONS_VERBOSE") != ""

	var b strings.Builder
	for _, tc := range r.TestCaseResults {
		icon := "✅"
		if !tc.Success {
			icon = "❌"
		}
		fmt.Fprintf(&b, "::group::%s %s\n", icon, tc.OperationName)
		for _, s := range tc.TestStepResults {
			switch {
			case !s.Success:
				fmt.Fprintf(&b, "::error %s::%s\n",
					f.errorProperties(tc.OperationName), escapeData(stepMessage(s)))
			case verbose:
				fmt.Fprintf(&b, "::notice title=%s::%s passed\n",
					escapeProperty(tc.OperationName), escapeData(s.RequestName))
			}
		}
		fmt.Fprintf(&b, "::endgroup::\n")
	}

	if r.Success {
		fmt.Fprintf(&b, "::notice title=Microcks contract test::All %d operation(s) conform to the contract\n",
			len(r.TestCaseResults))
	} else {
		fmt.Fprintf(&b, "::error title=Microcks contract test::Contract test failed - see annotations above\n")
	}

	if err := writeStepSummary(r); err != nil {
		// The step summary is best-effort; never fail the run over it.
		fmt.Fprintf(&b, "::warning::could not write GITHUB_STEP_SUMMARY: %s\n", escapeData(err.Error()))
	}

	return b.String(), nil
}

func (f *GitHubActionsFormatter) errorProperties(operationName string) string {
	props := []string{"title=" + escapeProperty(operationName)}
	if f.artifactPath != "" {
		props = append(props, "file="+escapeProperty(f.artifactPath))
		if line := operationLine(f.artifactPath, operationName); line > 0 {
			props = append(props, fmt.Sprintf("line=%d", line))
		}
	}
	return strings.Join(props, ",")
}

// stepMessage returns the failure message, or a sensible default when empty.
func stepMessage(s connectors.TestStepResult) string {
	if strings.TrimSpace(s.Message) != "" {
		return s.Message
	}
	if s.RequestName != "" {
		return s.RequestName + " did not conform to the contract"
	}
	return "did not conform to the contract"
}

// writeStepSummary appends a per-operation markdown table to the GitHub job
// summary file, if GITHUB_STEP_SUMMARY is set.
func writeStepSummary(r *connectors.TestResult) error {
	path := os.Getenv("GITHUB_STEP_SUMMARY")
	if path == "" {
		return nil
	}

	var b strings.Builder
	b.WriteString("## Microcks contract test\n\n")
	b.WriteString(fmt.Sprintf("**Overall:** %s\n\n", passFail(r.Success)))
	b.WriteString("| Operation | Result |\n| --- | --- |\n")
	for _, tc := range r.TestCaseResults {
		b.WriteString(fmt.Sprintf("| %s | %s |\n", tc.OperationName, passFail(tc.Success)))
	}
	b.WriteString("\n")

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(b.String())
	return err
}

func passFail(ok bool) string {
	if ok {
		return "✅ pass"
	}
	return "❌ fail"
}

// escapeData escapes a GitHub Actions command message per the workflow-command spec.
func escapeData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}

// escapeProperty escapes a GitHub Actions command property value.
func escapeProperty(s string) string {
	s = escapeData(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}
