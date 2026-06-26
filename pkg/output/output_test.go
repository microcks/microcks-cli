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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

func sampleResult() *connectors.TestResult {
	return &connectors.TestResult{
		ID:          "abc123",
		Success:     false,
		ElapsedTime: 1500,
		TestCaseResults: []connectors.TestCaseResult{
			{Success: true, OperationName: "GET /products", TestStepResults: []connectors.TestStepResult{
				{Success: true, RequestName: "all"},
			}},
			{Success: false, OperationName: "POST /orders", TestStepResults: []connectors.TestStepResult{
				{Success: false, RequestName: "new", Message: "price: expected number\ngot string"},
			}},
		},
	}
}

func TestNewFormatterAndIsValid(t *testing.T) {
	for _, f := range []string{"text", "json", "yaml", "github-actions"} {
		if !IsValid(f) {
			t.Errorf("expected %q to be valid", f)
		}
		if _, err := NewFormatter(OutputFormat(f)); err != nil {
			t.Errorf("NewFormatter(%q) errored: %v", f, err)
		}
	}
	if IsValid("xml") {
		t.Error("expected xml to be invalid")
	}
	if _, err := NewFormatter("xml"); err == nil {
		t.Error("expected NewFormatter(xml) to error")
	}
}

func TestTextFormatter(t *testing.T) {
	out, err := (&TextFormatter{}).Format(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"FAILURE", "[PASS] GET /products", "[FAIL] POST /orders", "price: expected number"} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q\n%s", want, out)
		}
	}
}

func TestJSONFormatter(t *testing.T) {
	out, err := (&JSONFormatter{}).Format(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"id": "abc123"`) || !strings.Contains(out, `"operationName": "POST /orders"`) {
		t.Errorf("json output unexpected:\n%s", out)
	}
}

func TestYAMLFormatter(t *testing.T) {
	out, err := (&YAMLFormatter{}).Format(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	// Keys should be camelCase (json field names), not Go field names.
	if !strings.Contains(out, "id: abc123") || !strings.Contains(out, "testCaseResults:") {
		t.Errorf("yaml output unexpected:\n%s", out)
	}
}

func TestGitHubActionsFormatter(t *testing.T) {
	out, err := (&GitHubActionsFormatter{}).Format(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	checks := []string{
		"::group::",
		"GET /products",
		"::error title=POST /orders::",
		"price: expected number%0Agot string", // newline escaped in data; colon only escaped in properties
		"::endgroup::",
		"::error title=Microcks contract test::",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("github-actions output missing %q\n%s", want, out)
		}
	}
	// No ::notice:: for the passing op unless verbose.
	if strings.Contains(out, "::notice title=GET /products::") {
		t.Errorf("unexpected ::notice:: without verbose:\n%s", out)
	}
}

func TestGitHubActionsVerbose(t *testing.T) {
	t.Setenv("MICROCKS_ACTIONS_VERBOSE", "1")
	out, err := (&GitHubActionsFormatter{}).Format(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "::notice title=GET /products::") {
		t.Errorf("expected ::notice:: for passing op in verbose mode:\n%s", out)
	}
}

func TestGitHubActionsStepSummary(t *testing.T) {
	summary := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summary)

	if _, err := (&GitHubActionsFormatter{}).Format(sampleResult()); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(summary)
	if err != nil {
		t.Fatalf("step summary not written: %v", err)
	}
	s := string(data)
	for _, want := range []string{"## Microcks contract test", "| Operation | Result |", "GET /products", "POST /orders"} {
		if !strings.Contains(s, want) {
			t.Errorf("step summary missing %q\n%s", want, s)
		}
	}
}

func TestEscaping(t *testing.T) {
	if got := escapeData("a%b\nc\rd"); got != "a%25b%0Ac%0Dd" {
		t.Errorf("escapeData = %q", got)
	}
	if got := escapeProperty("a:b,c"); got != "a%3Ab%2Cc" {
		t.Errorf("escapeProperty = %q", got)
	}
}
