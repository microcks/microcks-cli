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

	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/microcks/microcks-cli/version"
	"github.com/spf13/cobra"
)

const capabilitiesSchemaVersion = "v1"

var supportedCapabilities = []string{
	"auth.login",
	"auth.login.sso",
	"auth.logout",
	"context.list",
	"context.list.json",
	"context.use",
	"context.use.json",
	"context.delete",
	"context.delete.json",
	"instance.start",
	"instance.start.json",
	"instance.stop",
	"artifact.import.file",
	"artifact.import.file.json",
	"artifact.import.file.watch",
	"artifact.import.directory",
	"artifact.import.url",
	"service.list.json",
	"service.get.json",
	"test.run",
	"test.run.output.json",
	"test.run.output.yaml",
	"test.run.output.github-actions",
	"test.dry-run",
	"test.dry-run.watch",
	"test.dry-run.watch.events.json",
	"test.list.json",
	"test.get.json",
}

type capabilitiesDocument struct {
	SchemaVersion string   `json:"schemaVersion"`
	CLIVersion    string   `json:"cliVersion"`
	Capabilities  []string `json:"capabilities"`
}

func NewCapabilitiesCommand() *cobra.Command {
	var outputFormat string

	command := &cobra.Command{
		Use:   "capabilities",
		Short: "List machine-readable Microcks CLI capabilities",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}

			document := capabilitiesDocument{
				SchemaVersion: capabilitiesSchemaVersion,
				CLIVersion:    version.Version,
				Capabilities:  supportedCapabilities,
			}
			if outputFormat == "json" {
				return errors.Wrap(
					errors.KindEnvironment,
					output.WriteJSON(os.Stdout, document),
				)
			}

			for _, capability := range document.Capabilities {
				if _, err := fmt.Fprintln(os.Stdout, capability); err != nil {
					return errors.Wrap(
						errors.KindEnvironment,
						fmt.Errorf("writing capabilities output: %w", err),
					)
				}
			}
			return nil
		},
	}
	command.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return command
}
