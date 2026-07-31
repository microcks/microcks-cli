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
