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
	stderrors "errors"
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/pkg/errors"
)

// exitCodes maps each Failure Kind to a process exit status. This is the only
// place that knows about exit codes; pkg/* speaks in Failure Kinds.
var exitCodes = map[errors.Kind]int{
	errors.KindGeneric:     20,
	errors.KindUsage:       2,
	errors.KindConnection:  11,
	errors.KindAPI:         12,
	errors.KindNotFound:    13,
	errors.KindEnvironment: 14,
}

// ExitCodeFor returns the process exit code for err (0 when nil). Pure and
// testable; Handle is the side-effecting wrapper main() calls.
func ExitCodeFor(err error) int {
	switch {
	case err == nil:
		return 0
	case stderrors.Is(err, errors.ErrTestFailed):
		// A clean run with a non-conforming result — pytest-style exit 1.
		return 1
	default:
		if code, ok := exitCodes[errors.KindOf(err)]; ok {
			return code
		}
		// A non-nil error must never exit 0; fall back to generic.
		return exitCodes[errors.KindGeneric]
	}
}

// Handle is the single exit point for the CLI. It prints err to stderr (except
// the silent ErrTestFailed sentinel, whose result the command already rendered)
// and exits with the mapped code. main() calls this and nothing else exits.
func Handle(err error) {
	if err == nil {
		return
	}
	if !stderrors.Is(err, errors.ErrTestFailed) {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(ExitCodeFor(err))
}
