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
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

type usageError struct {
	err   error
	usage string
}

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

func usageErrorf(cmd *cobra.Command, format string, a ...any) error {
	usage := ""
	if cmd != nil {
		usage = cmd.UsageString()
	}
	return &usageError{
		err:   errors.Wrapf(errors.KindUsage, format, a...),
		usage: usage,
	}
}
