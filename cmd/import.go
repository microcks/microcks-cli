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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/watcher"
	"github.com/spf13/cobra"
)

func parsePathSuffix(arg string) (string, bool) {
	colonIdx := strings.LastIndex(arg, ":")
	if colonIdx < 0 {
		return arg, true
	}
	if runtime.GOOS == "windows" && colonIdx == 1 && len(arg) > 2 {
		if c := arg[0]; (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return arg, true
		}
	}
	pathPart := arg[:colonIdx]
	suffixPart := arg[colonIdx+1:]
	val, err := strconv.ParseBool(suffixPart)
	if err != nil {
		return arg, true
	}
	return pathPart, val
}

func NewImportCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var watch bool

	var importCmd = &cobra.Command{
		Use:   "import <specificationFile1[:primary],specificationFile2[:primary]>",
		Short: "import API artifacts on Microcks server",
		Long:  `import API artifacts on Microcks server`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specificationFiles := args[0]

			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
			if err != nil {
				return err
			}

			var mc connectors.MicrocksClient

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {
				// Create client with server address.
				mc = connectors.NewMicrocksClient(globalClientOpts.ServerAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					return fmt.Errorf("got error when invoking Microcks client retrieving config: %s", err)
				}

				var oauthToken string = "unauthenticated-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						return fmt.Errorf("got error when invoking Keycloak client: %s", err)
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
					return fmt.Errorf("please login to perform operation")
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
			for _, f := range sepSpecificationFiles {
				mainArtifact := true
				var err error

				f, mainArtifact = parsePathSuffix(f)

				// Try uploading this artifact.
				msg, err := mc.UploadArtifact(f, mainArtifact)
				if err != nil {
					return fmt.Errorf("got error when invoking Microcks client importing Artifact: %s", err)
				}
				action := "discovered"
				if !mainArtifact {
					action = "completed"
				}
				fmt.Printf("Microcks has %s '%s'\n", action, msg)

				// If watch flag is provided, update watch config.
				if watch {
					watchFile, err := config.DefaultLocalWatchPath()
					errors.CheckError(err)

					watchCfg, err := config.ReadLocalWatchConfig(watchFile)
					errors.CheckError(err)
					if watchCfg == nil {
						watchCfg = &config.WatchConfig{}
					}

					// Normalize file path to match the watcher fsnotify events format.
					f = strings.TrimPrefix(f, "./")
					f = strings.TrimPrefix(f, ".\\")
					f = filepath.Clean(f)

					// Upsert entry.
					watchCfg.UpsertEntry(config.WatchEntry{
						FilePath:     f,
						Context:      []string{globalClientOpts.Context},
						MainArtifact: mainArtifact,
					})

					// Write watch file.
					err = config.WriteLocalWatchConfig(*watchCfg, watchFile)
					errors.CheckError(err)
				}
			}

			// Start watcher if --watch flag is provided.
			if watch {
				watchFile, err := config.DefaultLocalWatchPath()
				errors.CheckError(err)

				wm, err := watcher.NewWatchManger(watchFile)
				errors.CheckError(err)

				fmt.Println("Watch mode enabled - microcks-watcher started...")
				wm.Run()
			}
			return nil
		},
	}

	importCmd.Flags().BoolVar(&watch, "watch", false, "Keep watch on file changes and re-import it on change")
	return importCmd
}
