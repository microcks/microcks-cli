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
	"strings"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewImportURLCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var importURLCmd = &cobra.Command{
		Use:   "import-url",
		Short: "import API artifacts from URL on Microcks server",
		Long:  `import API artifacts from URL on Microcks server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse subcommand args first.
			if len(args) == 0 {
				return errors.Wrapf(errors.KindUsage, "import-url requires a <specificationFileURL[:primary[:secret]]> argument")
			}

			specificationFiles := args[0]

			mc, _, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}
			sepSpecificationFiles := strings.Split(specificationFiles, ",")
			for _, f := range sepSpecificationFiles {
				artifactURL, mainArtifact, secret := parseImportURLSpecifier(f)

				// Try downloading the artifcat
				msg, err := mc.DownloadArtifact(artifactURL, mainArtifact, secret)
				if err != nil {
					return err
				}
				fmt.Printf("Microcks has discovered '%s'\n", msg)
			}
			return nil
		},
	}

	return importURLCmd
}
