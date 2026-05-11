package output

import (
	"github.com/microcks/microcks-cli/pkg/connectors"
	"gopkg.in/yaml.v2"
)

type YAMLFormatter struct{}

func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

func (f *YAMLFormatter) FormatTestResult(result *connectors.TestResult) string {
	formattedResult := NewTestResultJSON(result)
	bytes, err := yaml.Marshal(formattedResult)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (f *YAMLFormatter) FormatTestCaseResult(testCase *connectors.TestCaseResult) string {
	bytes, err := yaml.Marshal(testCase)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
