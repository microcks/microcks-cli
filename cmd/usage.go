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
	"strings"

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

func wrapUsageError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	var ue *usageError
	if stderrors.As(err, &ue) {
		return err
	}
	usage := ""
	if cmd != nil {
		usage = cmd.UsageString()
	}
	return &usageError{
		err:   errors.Wrap(errors.KindUsage, err),
		usage: usage,
	}
}

func unknownCommandError(cmd *cobra.Command, arg string) error {
	var sb strings.Builder
	if cmd != nil {
		if suggestions := cmd.SuggestionsFor(arg); len(suggestions) > 0 {
			sb.WriteString("\n\nDid you mean this?\n")
			for _, s := range suggestions {
				fmt.Fprintf(&sb, "\t%v\n", s)
			}
		}
	}
	cmdPath := "microcks"
	if cmd != nil {
		cmdPath = cmd.CommandPath()
	}
	return wrapUsageError(cmd, fmt.Errorf("unknown command %q for %q%s", arg, cmdPath, sb.String()))
}

func setupUsageErrors(root *cobra.Command) {
	// Child commands inherit FlagErrorFunc from the root command in Cobra.
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return wrapUsageError(cmd, err)
	})

	wrapArgs(root)
}

func wrapArgs(c *cobra.Command) {
	if c.Args != nil {
		origArgs := c.Args
		c.Args = func(cmd *cobra.Command, args []string) error {
			if err := origArgs(cmd, args); err != nil {
				return wrapUsageError(cmd, err)
			}
			return nil
		}
	} else if !c.HasParent() {
		c.Args = func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownCommandError(cmd, args[0])
			}
			return nil
		}
	}

	for _, child := range c.Commands() {
		wrapArgs(child)
	}
}
