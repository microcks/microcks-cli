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

func NewCompletionCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long:  `Generate shell completion scripts`,
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{
			"bash",
			"zsh",
			"fish",
			"powershell",
		},
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd := cmd.Root()
			var err error

			switch args[0] {
			case "bash":
				err = rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				err = rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				err = rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				err = rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
			}
			errors.CheckError(err)
		},
	}

	return command
}
