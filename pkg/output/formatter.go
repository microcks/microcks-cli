package output

import (
	"fmt"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

type Formatter interface {
	FormatTestResult(result *connectors.TestResult) string
	FormatTestCaseResult(testCase *connectors.TestCaseResult) string
}

type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
)

func NewFormatter(format OutputFormat) (Formatter, error) {
	switch format {
	case OutputFormatText:
		return NewTextFormatter(), nil
	case OutputFormatJSON:
		return NewJSONFormatter(), nil
	case OutputFormatYAML:
		return NewYAMLFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported output format %q", format)
	}
}
