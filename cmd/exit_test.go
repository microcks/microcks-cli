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
	"bytes"
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/errors"
)

func TestExitCodeFor(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"test failed", errors.ErrTestFailed, 1},
		{"test failed wrapped", fmt.Errorf("run: %w", errors.ErrTestFailed), 1},
		{"usage", errors.Wrap(errors.KindUsage, stderrors.New("bad flag")), 2},
		{"connection", errors.Wrap(errors.KindConnection, stderrors.New("refused")), 11},
		{"api", errors.Wrap(errors.KindAPI, stderrors.New("500")), 12},
		{"not found", errors.Wrap(errors.KindNotFound, stderrors.New("404")), 13},
		{"environment", errors.Wrap(errors.KindEnvironment, stderrors.New("no docker")), 14},
		{"generic", stderrors.New("boom"), 20},
	}
	for _, c := range cases {
		if got := ExitCodeFor(c.err); got != c.want {
			t.Errorf("%s: ExitCodeFor = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestPrintErrorIncludesUsageFromUsageError(t *testing.T) {
	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand returned error: %v", err)
	}
	importCmd, _, err := cmd.Find([]string{"import"})
	if err != nil {
		t.Fatalf("could not find import command: %v", err)
	}

	var out bytes.Buffer
	printError(&out, usageErrorf(importCmd, "import requires a file"))

	got := out.String()
	if !strings.Contains(got, "Usage:") {
		t.Fatalf("printError did not include usage: %q", got)
	}
	if !strings.Contains(got, "microcks import") {
		t.Fatalf("printError did not include import usage: %q", got)
	}
	if !strings.Contains(got, "import requires a file") {
		t.Fatalf("printError did not include error: %q", got)
	}
}

func TestUsageErrorIsKindUsage(t *testing.T) {
	err := usageErrorf(nil, "bad args")
	if got := errors.KindOf(err); got != errors.KindUsage {
		t.Fatalf("KindOf = %v, want KindUsage", got)
	}
}

func TestCobraSyntaxAndArgumentErrorsReturnKindUsage(t *testing.T) {
	cases := []struct {
		name             string
		args             []string
		wantErrSubstring string
		wantUsageSnippet string
	}{
		{
			name:             "unknown flag on root",
			args:             []string{"--unknown-flag"},
			wantErrSubstring: "unknown flag: --unknown-flag",
			wantUsageSnippet: "microcks [command]",
		},
		{
			name:             "unknown flag on subcommand",
			args:             []string{"test", "--unknown-flag"},
			wantErrSubstring: "unknown flag: --unknown-flag",
			wantUsageSnippet: "microcks test",
		},
		{
			name:             "missing flag argument",
			args:             []string{"start", "--port"},
			wantErrSubstring: "flag needs an argument: --port",
			wantUsageSnippet: "microcks start",
		},
		{
			name:             "missing required positional arguments",
			args:             []string{"test"},
			wantErrSubstring: "accepts 3 arg(s), received 0",
			wantUsageSnippet: "microcks test",
		},
		{
			name:             "required together flags missing",
			args:             []string{"--keycloakClientId=test"},
			wantErrSubstring: "missing [keycloakClientSecret]",
			wantUsageSnippet: "microcks [command]",
		},
		{
			name:             "unknown command",
			args:             []string{"non-existent-command"},
			wantErrSubstring: `unknown command "non-existent-command" for "microcks"`,
			wantUsageSnippet: "microcks [command]",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd, err := NewCommand()
			if err != nil {
				t.Fatalf("NewCommand returned error: %v", err)
			}
			cmd.SetArgs(c.args)

			execErr := cmd.Execute()
			if execErr == nil {
				t.Fatalf("expected error for args %v, got nil", c.args)
			}

			if gotKind := errors.KindOf(execErr); gotKind != errors.KindUsage {
				t.Errorf("KindOf = %v, want KindUsage", gotKind)
			}

			if gotCode := ExitCodeFor(execErr); gotCode != 2 {
				t.Errorf("ExitCodeFor = %d, want 2", gotCode)
			}

			var out bytes.Buffer
			printError(&out, execErr)
			gotOutput := out.String()

			if !strings.Contains(gotOutput, "Usage:") {
				t.Errorf("printError output missing 'Usage:': %q", gotOutput)
			}
			if !strings.Contains(gotOutput, c.wantUsageSnippet) {
				t.Errorf("printError output missing snippet %q: %q", c.wantUsageSnippet, gotOutput)
			}
			if !strings.Contains(gotOutput, c.wantErrSubstring) {
				t.Errorf("printError output missing error substring %q: %q", c.wantErrSubstring, gotOutput)
			}
		})
	}
}

func TestWrapUsageError(t *testing.T) {
	if got := wrapUsageError(nil, nil); got != nil {
		t.Fatalf("wrapUsageError(nil, nil) = %v, want nil", got)
	}

	origErr := stderrors.New("some error")
	wrapped := wrapUsageError(nil, origErr)
	if gotKind := errors.KindOf(wrapped); gotKind != errors.KindUsage {
		t.Fatalf("KindOf = %v, want KindUsage", gotKind)
	}

	// Double wrapping should return existing *usageError
	doubleWrapped := wrapUsageError(nil, wrapped)
	if doubleWrapped != wrapped {
		t.Fatalf("expected idempotent wrapping, got %v", doubleWrapped)
	}
}
