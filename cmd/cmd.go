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
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

type envFlagBinding struct {
	flagName string
	envName  string
}

var persistentEnvBindings = []envFlagBinding{
	{flagName: "config", envName: "MICROCKS_CONFIG"},
	{flagName: "microcks-context", envName: "MICROCKS_CONTEXT"},
	{flagName: "verbose", envName: "MICROCKS_VERBOSE"},
	{flagName: "insecure-tls", envName: "MICROCKS_INSECURE_TLS"},
	{flagName: "caCerts", envName: "MICROCKS_CA_CERTS"},
	{flagName: "keycloakClientId", envName: "MICROCKS_KEYCLOAK_CLIENT_ID"},
	{flagName: "keycloakClientSecret", envName: "MICROCKS_KEYCLOAK_CLIENT_SECRET"},
	{flagName: "microcksURL", envName: "MICROCKS_URL"},
}

func applyPersistentFlagEnvFallbacks(cmd *cobra.Command) error {
	rootFlags := cmd.Root().PersistentFlags()
	for _, binding := range persistentEnvBindings {
		flag := rootFlags.Lookup(binding.flagName)
		if flag == nil || flag.Changed {
			continue
		}
		val, ok := os.LookupEnv(binding.envName)
		if !ok || val == "" {
			continue
		}
		if err := rootFlags.Set(binding.flagName, val); err != nil {
			return wrapUsageError(cmd, fmt.Errorf("invalid value for %s: %w", binding.envName, err))
		}
	}
	return nil
}

func NewCommand() (*cobra.Command, error) {

	var clientOpts connectors.ClientOptions

	command := &cobra.Command{
		Use:   "microcks",
		Short: "A CLI tool for Microcks",
		// cmd.Handle (called by main) is the single point that prints errors and
		// exits, so Cobra must not also print them.
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			clientOpts.RequestContext = cmd.Context()
			if err := applyPersistentFlagEnvFallbacks(cmd); err != nil {
				return err
			}
			if err := cmd.ValidateRequiredFlags(); err != nil {
				return wrapUsageError(cmd, err)
			}
			if err := cmd.ValidateFlagGroups(); err != nil {
				return wrapUsageError(cmd, err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	command.AddCommand(NewImportCommand(&clientOpts))
	command.AddCommand(NewImportDirCommand(&clientOpts))
	command.AddCommand(NewVersionCommand())
	command.AddCommand(NewCapabilitiesCommand())
	command.AddCommand(NewTestCommand(&clientOpts))
	command.AddCommand(NewServiceCommand(&clientOpts))
	command.AddCommand(NewImportURLCommand(&clientOpts))
	command.AddCommand(NewStartCommand(&clientOpts))
	command.AddCommand(NewStopCommand(&clientOpts))
	command.AddCommand(NewContextCommand(&clientOpts))
	command.AddCommand(NewLoginCommand(&clientOpts))
	command.AddCommand(NewLogoutCommand(&clientOpts))

	defaultLocalConfigPath, err := config.DefaultLocalConfigPath()
	if err != nil {
		return nil, err
	}
	command.PersistentFlags().StringVar(&clientOpts.ConfigPath, "config", defaultLocalConfigPath, "Path to Microcks config (env: MICROCKS_CONFIG)")
	command.PersistentFlags().StringVar(&clientOpts.Context, "microcks-context", "", "Name of the Microcks context to use (env: MICROCKS_CONTEXT)")
	command.PersistentFlags().BoolVar(&clientOpts.Verbose, "verbose", false, "Produce dumps of HTTP exchanges (env: MICROCKS_VERBOSE)")
	command.PersistentFlags().BoolVar(&clientOpts.InsecureTLS, "insecure-tls", false, "Whether to accept insecure HTTPS connection (env: MICROCKS_INSECURE_TLS)")
	command.PersistentFlags().StringVar(&clientOpts.CaCertPaths, "caCerts", "", "Comma separated paths of CRT files to add to Root CAs (env: MICROCKS_CA_CERTS)")
	command.PersistentFlags().StringVar(&clientOpts.ClientId, "keycloakClientId", "", "Keycloak Realm Service Account ClientId (env: MICROCKS_KEYCLOAK_CLIENT_ID)")
	command.PersistentFlags().StringVar(&clientOpts.ClientSecret, "keycloakClientSecret", "", "Keycloak Realm Service Account ClientSecret (env: MICROCKS_KEYCLOAK_CLIENT_SECRET)")
	command.PersistentFlags().StringVar(&clientOpts.ServerAddr, "microcksURL", "", "Microcks API URL (env: MICROCKS_URL)")
	command.MarkFlagsRequiredTogether("keycloakClientId", "keycloakClientSecret")

	setupUsageErrors(command)

	return command, nil
}
