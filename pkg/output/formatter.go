package output

import (
	"fmt"
	"io"
	"os"

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

type Writer struct {
	infoWriter     io.Writer
	progressWriter io.Writer
}

func NewWriter(format OutputFormat) *Writer {
	if format == OutputFormatText {
		return &Writer{
			infoWriter:     os.Stdout,
			progressWriter: os.Stdout,
		}
	}
	return &Writer{
		infoWriter:     os.Stderr,
		progressWriter: os.Stderr,
	}
}

func (w *Writer) Infof(format string, args ...any) {
	fmt.Fprintf(w.infoWriter, format, args...)
}

func (w *Writer) Progressf(format string, args ...any) {
	fmt.Fprintf(w.progressWriter, format, args...)
}
