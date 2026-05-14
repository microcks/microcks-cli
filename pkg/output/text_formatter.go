package output

import (
	"fmt"
	"strings"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

type TextFormatter struct{}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

func (f *TextFormatter) FormatTestResult(result *connectors.TestResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Test \"%s\" success: %v\n", result.ID, result.Success))

	for _, tc := range result.TestCases {
		sb.WriteString(f.FormatTestCaseResult(&tc))
	}

	return sb.String()
}

func (f *TextFormatter) FormatTestCaseResult(testCase *connectors.TestCaseResult) string {
	var sb strings.Builder

	for _, step := range testCase.TestStepResults {
		if step.Message != "" {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", testCase.OperationName, step.RequestName, step.Message))
		}
	}

	return sb.String()
}