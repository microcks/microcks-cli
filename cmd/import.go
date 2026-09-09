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
				return usageErrorf(cmd, "import requires a <specificationFile1[:primary],specificationFile2[:primary]> argument")
			}

			specificationFiles := args[0]

			mc, serverAddr, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}

			watchContext := globalClientOpts.Context
			if watch && watchContext == "" {
				watchContext, err = defaultImportWatchContext(globalClientOpts.ConfigPath, serverAddr)
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
						Context:      []string{watchContext},
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

func defaultImportWatchContext(configPath, serverAddr string) (string, error) {
	localConfig, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return "", errors.Wrap(errors.KindEnvironment, err)
	}
	if localConfig != nil && localConfig.CurrentContext != "" {
		return localConfig.CurrentContext, nil
	}
	return serverAddr, nil
}

type artifactImportResult struct {
	File    string `json:"file"`
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
	Action  string `json:"action"`
}
