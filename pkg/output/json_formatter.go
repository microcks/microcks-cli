package output

import (
	"encoding/json"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) FormatTestResult(result *connectors.TestResult) string {
	formattedResult := NewTestResultJSON(result)
	bytes, err := json.MarshalIndent(formattedResult, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (f *JSONFormatter) FormatTestCaseResult(testCase *connectors.TestCaseResult) string {
	bytes, err := json.MarshalIndent(testCase, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
