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

package root

import (
	"fmt"

	"github.com/microcks/microcks-cli/cmd/microcks/internal/version"
	"github.com/spf13/cobra"
)


func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Check this CLI version",
		Long: `Get the version of the Microcks CLI.`,
		Example: ` microcks version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Microcks CLI version: %s\n", version.Version)
			return nil
		},
	}
	return cmd
}