package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"gopkg.in/yaml.v2"
)

func TestTextFormatter_FormatTestResult(t *testing.T) {
	formatter := NewTextFormatter()

	result := &connectors.TestResult{
		ID:      "test-123",
		Success: true,
		TestCases: []connectors.TestCaseResult{
			{
				Success:       true,
				OperationName: "GET /api/users",
				TestStepResults: []connectors.TestStepResult{
					{
						Success:     true,
						RequestName: "Test Case 1",
						Message:     "",
					},
				},
			},
		},
	}

	output := formatter.FormatTestResult(result)
	if output != "" && !strings.Contains(output, "success: true") {
		t.Errorf("TextFormatter should include success status, got: %s", output)
	}
}

func TestTextFormatter_FormatTestResult_WithFailures(t *testing.T) {
	formatter := NewTextFormatter()

	result := &connectors.TestResult{
		ID:      "test-456",
		Success: false,
		TestCases: []connectors.TestCaseResult{
			{
				Success:       false,
				OperationName: "POST /api/users",
				TestStepResults: []connectors.TestStepResult{
					{
						Success:     false,
						RequestName: "Invalid Request",
						Message:     "expected status 201, got 400",
					},
				},
			},
		},
	}

	output := formatter.FormatTestResult(result)
	if !strings.Contains(output, "POST /api/users") {
		t.Errorf("TextFormatter should include operation name, got: %s", output)
	}
	if !strings.Contains(output, "expected status 201, got 400") {
		t.Errorf("TextFormatter should include failure message, got: %s", output)
	}
}

func TestJSONFormatter_FormatTestResult(t *testing.T) {
	formatter := NewJSONFormatter()

	result := &connectors.TestResult{
		ID:         "test-789",
		Version:    1,
		TestNumber: 1,
		Success:    true,
		TestCases: []connectors.TestCaseResult{
			{
				Success:       true,
				OperationName: "GET /api/health",
				TestStepResults: []connectors.TestStepResult{
					{
						Success:     true,
						RequestName: "Health Check",
						ElapsedTime: 50,
					},
				},
			},
		},
	}

	output := formatter.FormatTestResult(result)

	var parsed TestResultJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("JSONFormatter output should be valid JSON: %v", err)
	}

	if parsed.ID != "test-789" {
		t.Errorf("JSONFormatter should preserve ID field, got: %s", parsed.ID)
	}
	if len(parsed.TestCases) != 1 {
		t.Errorf("JSONFormatter should include testCases, got: %d", len(parsed.TestCases))
	}
}

func TestYAMLFormatter_FormatTestResult(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &connectors.TestResult{
		ID:             "test-yaml-001",
		Version:        1,
		TestNumber:     11,
		TestDate:       1700000000,
		TestedEndpoint: "https://example.org/api",
		ServiceID:      "svc-123",
		ElapsedTime:    99,
		Success:        true,
		InProgress:     false,
		RunnerType:     "HTTP",
		TestCases: []connectors.TestCaseResult{
			{
				Success:       true,
				OperationName: "GET /health",
				TestStepResults: []connectors.TestStepResult{
					{
						Success:     true,
						RequestName: "check",
						Message:     "ok",
					},
				},
			},
		},
	}

	output := formatter.FormatTestResult(result)

	var parsed TestResultJSON
	if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("YAMLFormatter output should be valid YAML: %v", err)
	}
	if parsed.ID != "test-yaml-001" {
		t.Errorf("YAMLFormatter should preserve ID field, got: %s", parsed.ID)
	}
	if len(parsed.TestCases) != 1 {
		t.Errorf("YAMLFormatter should include testCaseResults, got: %d", len(parsed.TestCases))
	}
}

func TestJSONFormatter_FormatTestCaseResult(t *testing.T) {
	formatter := NewJSONFormatter()

	testCase := &connectors.TestCaseResult{
		Success:       false,
		ElapsedTime:   100,
		OperationName: "DELETE /api/users/1",
		TestStepResults: []connectors.TestStepResult{
			{
				Success: false,
				Message: "connection refused",
			},
		},
	}

	output := formatter.FormatTestCaseResult(testCase)

	var parsed connectors.TestCaseResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("JSONFormatter output should be valid JSON: %v", err)
	}

	if parsed.Success {
		t.Errorf("JSONFormatter should preserve success=false")
	}
	if !strings.Contains(output, "DELETE /api/users/1") {
		t.Errorf("JSONFormatter should include operationName, got: %s", output)
	}
}

func TestNewTextFormatter_ReturnsPointer(t *testing.T) {
	formatter := NewTextFormatter()
	if formatter == nil {
		t.Error("NewTextFormatter should not return nil")
	}
}

func TestNewJSONFormatter_ReturnsPointer(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter == nil {
		t.Error("NewJSONFormatter should not return nil")
	}
}

func TestNewYAMLFormatter_ReturnsPointer(t *testing.T) {
	formatter := NewYAMLFormatter()
	if formatter == nil {
		t.Error("NewYAMLFormatter should not return nil")
	}
}

func TestNewFormatter(t *testing.T) {
	formatter, err := NewFormatter(OutputFormatText)
	if err != nil || formatter == nil {
		t.Fatalf("NewFormatter text should return formatter, err=%v", err)
	}

	formatter, err = NewFormatter(OutputFormatJSON)
	if err != nil || formatter == nil {
		t.Fatalf("NewFormatter json should return formatter, err=%v", err)
	}

	formatter, err = NewFormatter(OutputFormatYAML)
	if err != nil || formatter == nil {
		t.Fatalf("NewFormatter yaml should return formatter, err=%v", err)
	}

	formatter, err = NewFormatter(OutputFormat("csv"))
	if err == nil || formatter != nil {
		t.Fatalf("NewFormatter should fail on unknown format")
	}
}

func TestWriterRouting(t *testing.T) {
	textWriter := NewWriter(OutputFormatText)
	var textBuf bytes.Buffer
	textWriter.out = &textBuf
	textWriter.Infof("hello %s", "text")
	if textBuf.String() != "hello text" {
		t.Fatalf("text writer should write to stdout buffer, got %q", textBuf.String())
	}

	jsonWriter := NewWriter(OutputFormatJSON)
	var jsonBuf bytes.Buffer
	jsonWriter.out = &jsonBuf
	jsonWriter.Progressf("hello %s", "json")
	if jsonBuf.String() != "hello json" {
		t.Fatalf("json writer should write to stderr buffer, got %q", jsonBuf.String())
	}
}
