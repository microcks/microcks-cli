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
	"strings"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

// TextFormatter renders a human-readable summary of the test result.
type TextFormatter struct{}

func (f *TextFormatter) Format(r *connectors.TestResult) (string, error) {
	var b strings.Builder

	status := "SUCCESS"
	if !r.Success {
		status = "FAILURE"
	}
	fmt.Fprintf(&b, "Test %s: %s (%dms)\n", r.ID, status, r.ElapsedTime)

	for _, tc := range r.TestCaseResults {
		mark := "PASS"
		if !tc.Success {
			mark = "FAIL"
		}
		fmt.Fprintf(&b, "  [%s] %s\n", mark, tc.OperationName)
		for _, s := range tc.TestStepResults {
			if !s.Success && s.Message != "" {
				fmt.Fprintf(&b, "      %s: %s\n", s.RequestName, s.Message)
			}
		}
	}

	return b.String(), nil
}
