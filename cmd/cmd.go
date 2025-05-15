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
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCommad() *cobra.Command {

	var clientOpts connectors.ClientOptions

	command := &cobra.Command{
		Use:          "microcks",
		Short:        "A CLI tool for Microcks",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	command.AddCommand(NewImportCommand(&clientOpts))
	command.AddCommand(NewVersionCommand())
	command.AddCommand(NewTestCommand())
	command.AddCommand(NewImportURLCommand())
	command.AddCommand(NewStartCommand())
	command.AddCommand(NewStopCommand())
	command.AddCommand(NewContextCommand())
	command.AddCommand(NewLoginCommand(&clientOpts))

	defaultLocalConfigPath, err := config.DefaultLocalConfigPath()
	errors.CheckError(err)
	command.PersistentFlags().StringVar(&clientOpts.ConfigPath, "config", defaultLocalConfigPath, "Path to Microcks config")
	command.PersistentFlags().StringVar(&clientOpts.Context, "microcks-context", "", "Name of the Microcks context to use")
	command.PersistentFlags().BoolVar(&clientOpts.Verbose, "verbose", false, "Produce dumps of HTTP exchanges")
	command.PersistentFlags().BoolVar(&clientOpts.InsecureTLS, "insecure-tls", false, "Whether to accept insecure HTTPS connection")
	command.PersistentFlags().StringVar(&clientOpts.CaCertPaths, "caCerts", "", "Comma separated paths of CRT files to add to Root CAs")
	return command
}
