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
	"strconv"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/microcks/microcks-cli/pkg/watcher"
	"github.com/spf13/cobra"
)

func NewImportCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var watch bool
	var outputFormat string

	var importCmd = &cobra.Command{
		Use:   "import",
		Short: "import API artifacts on Microcks server",
		Long:  `import API artifacts on Microcks server`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}
			if watch && outputFormat == "json" {
				return errors.Wrapf(errors.KindUsage, "--output json is not supported with --watch")
			}
			// Parse subcommand args first.
			if len(args) == 0 {
				return errors.Wrapf(errors.KindUsage, "import requires a <specificationFile1[:primary],specificationFile2[:primary]> argument")
			}

			specificationFiles := args[0]

			// Initialize config from command options.
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			// Read local config file in case we need some context info.
			localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
			if err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}

			// Prepare Microcks client.
			var mc connectors.MicrocksClient

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {
				// Create client with server address.
				var err error
				mc, err = connectors.NewMicrocksClient(globalClientOpts.ServerAddr)
				if err != nil {
					return err
				}

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					return err
				}

				oauthToken := "unauthenticated-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc, err := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)
					if err != nil {
						return err
					}

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						return err
					}
				}

				// Set Auth token.
				mc.SetOAuthToken(oauthToken)

				// If no context provided use current one from config file or client server address.
				// So that watch config can be updated properly, referencing the right context.
				if globalClientOpts.Context == "" {
					if (localConfig != nil) && (localConfig.CurrentContext != "") {
						globalClientOpts.Context = localConfig.CurrentContext
					} else {
						globalClientOpts.Context = globalClientOpts.ServerAddr
					}
				}

			} else {
				// Create client from config file and using the current or provided context.
				if localConfig == nil {
					return errors.Wrapf(errors.KindUsage, "please login to perform this operation")
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					return err
				}
			}

			// Handle multiple specification files separated by comma.
			sepSpecificationFiles := strings.Split(specificationFiles, ",")
			results := make([]artifactImportResult, 0, len(sepSpecificationFiles))
			for _, f := range sepSpecificationFiles {
				mainArtifact := true
				var err error

				// Check if mainArtifact flag is provided.
				if strings.Contains(f, ":") {
					pathAndMainArtifact := strings.Split(f, ":")
					f = pathAndMainArtifact[0]
					mainArtifact, err = strconv.ParseBool(pathAndMainArtifact[1])
					if err != nil {
						return errors.Wrapf(errors.KindUsage, "cannot parse %q as artifact primary flag", pathAndMainArtifact[1])
					}
				}

				// Try uploading this artifact.
				msg, err := mc.UploadArtifact(f, mainArtifact)
				if err != nil {
					return err
				}
				action := "discovered"
				if !mainArtifact {
					action = "completed"
				}
				results = append(results, artifactImportResult{
					File: f, ID: msg, Primary: mainArtifact, Action: action,
				})
				if outputFormat == "text" {
					if _, err := fmt.Printf("Microcks has %s '%s'\n", action, msg); err != nil {
						return errors.Wrap(errors.KindEnvironment, err)
					}
				}

				// If watch flag is provided, update watch config.
				if watch {
					watchFile, err := config.DefaultLocalWatchPath()
					if err != nil {
						return errors.Wrap(errors.KindEnvironment, err)
					}

					watchCfg, err := config.ReadLocalWatchConfig(watchFile)
					if err != nil {
						return errors.Wrap(errors.KindEnvironment, err)
					}
					if watchCfg == nil {
						watchCfg = &config.WatchConfig{}
					}

					// Normalize file path to match the watcher fsnotify events format.
					if strings.HasPrefix(f, "./") {
						f = strings.TrimPrefix(f, "./")
					}

					// Upsert entry.
					watchCfg.UpsertEntry(config.WatchEntry{
						FilePath:     f,
						Context:      []string{globalClientOpts.Context},
						MainArtifact: mainArtifact,
					})

					// Write watch file.
					if err := config.WriteLocalWatchConfig(*watchCfg, watchFile); err != nil {
						return errors.Wrap(errors.KindEnvironment, err)
					}
				}
			}

			// Start watcher if --watch flag is provided.
			if watch {
				watchFile, err := config.DefaultLocalWatchPath()
				if err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}

				wm, err := watcher.NewWatchManger(watchFile)
				if err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}

				if _, err := fmt.Println("Watch mode enabled - microcks-watcher started..."); err != nil {
					return errors.Wrap(errors.KindEnvironment, err)
				}
				wm.Run()
			}
			if outputFormat == "json" {
				return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, results))
			}
			return nil
		},
	}

	importCmd.Flags().BoolVar(&watch, "watch", false, "Keep watch on file changes and re-import it on change")
	importCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return importCmd
}

type artifactImportResult struct {
	File    string `json:"file"`
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
	Action  string `json:"action"`
}
