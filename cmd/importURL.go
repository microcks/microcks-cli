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

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewImportURLCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var outputFormat string

	var importURLCmd = &cobra.Command{
		Use:   "import-url",
		Short: "import API artifacts from URL on Microcks server",
		Long:  `import API artifacts from URL on Microcks server`,
		Args:  validateImportURLArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}

			specificationFiles := args[0]

			mc, _, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}
			sepSpecificationFiles := strings.Split(specificationFiles, ",")
			results := make([]urlArtifactImportResult, 0, len(sepSpecificationFiles))
			for _, f := range sepSpecificationFiles {
				f, mainArtifact, secret := parseImportURLArg(f)

				// Try downloading the artifact.
				msg, err := mc.DownloadArtifact(f, mainArtifact, secret)
				if err != nil {
					return err
				}
				action := "discovered"
				if !mainArtifact {
					action = "completed"
				}
				results = append(results, urlArtifactImportResult{
					URL:     f,
					ID:      msg,
					Primary: mainArtifact,
					Action:  action,
				})
				if outputFormat == "text" {
					if _, err := fmt.Printf("Microcks has discovered '%s'\n", msg); err != nil {
						return errors.Wrap(errors.KindEnvironment, err)
					}
				}
			}
			if outputFormat == "json" {
				return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, results))
			}
			return nil
		},
	}

	importURLCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return importURLCmd
}

type urlArtifactImportResult struct {
	URL     string `json:"url"`
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
	Action  string `json:"action"`
}

func validateImportURLArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.Wrapf(errors.KindUsage, "import-url requires a <specificationFileURL[:primary[:secret]]> argument")
	}
	if len(args) > 1 {
		return errors.Wrapf(errors.KindUsage, "import-url accepts only 1 positional argument; specify multiple URLs as a comma-separated list: <url1,url2>")
	}

	for _, f := range strings.Split(args[0], ",") {
		lower := strings.ToLower(f)
		if (!strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "http://")) || lower == "http://" || lower == "https://" {
			return errors.Wrapf(
				errors.KindUsage,
				"invalid artifact URL '%s': must start with http:// or https://",
				f,
			)
		}
	}
	return nil
}

func parseImportURLArg(f string) (string, bool, string) {
	mainArtifact := true
	secret := ""

	parts := strings.Split(f, ":")
	n := len(parts)

	for i := 2; i < n; i++ {
		if val, parseErr := strconv.ParseBool(parts[i]); parseErr == nil {
			mainArtifact = val
			if i+1 < n {
				secret = strings.Join(parts[i+1:], ":")
			}
			f = strings.Join(parts[:i], ":")
			break
		}
	}
	return f, mainArtifact, secret
}
