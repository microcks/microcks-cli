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

// NewValidateCommand builds the "validate" command which runs pre-flight
// checks on the local configuration and connectivity to a Microcks server,
// without performing any import or test action.
func NewValidateCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	var validateCmd = &cobra.Command{

		Use:   "validate",
		Short: "Validate current Microcks CLI configuration and connectivity",
		Long: `Runs pre-flight checks on the active Microcks context:
   - config file is present and parses correctly
   - a context is configured and resolvable
   - the auth token is valid (refreshed if needed)
   - the target Microcks server is reachable
 
 Exits with code 0 if all checks pass, or 1 if any check fails.`,
		Example: `microcks validate
 microcks validate --microcks-context staging`,

		Run: func(cmd *cobra.Command, args []string) {
			ok := true

			// 1. Config file presence and parsing.
			configFile := globalClientOpts.ConfigPath
			localConfig, err := config.ReadLocalConfig(configFile)
			if err != nil {
				fmt.Printf("✗ Config file: invalid — %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ Config file: parsed successfully")

			if localConfig == nil || localConfig.IsEmpty() {
				fmt.Println("✗ No contexts configured. Run 'microcks login <server>' first.")
				os.Exit(1)
			}

			// 2. Context resolution.
			ctx, err := localConfig.ResolveContext(globalClientOpts.Context)
			if err != nil {
				fmt.Printf("✗ Context resolution: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Context resolved: %q (server: %s)\n", ctx.Name, ctx.Server.Server)

			// 3. Build a client for this context. NewClient() attempts a token
			// refresh internally, so a failure here reveals an expired or
			// invalid credential.
			mc, err := connectors.NewClient(*globalClientOpts)
			if err != nil {
				fmt.Printf("✗ Authentication: %v\n", err)
				ok = false
			} else {
				fmt.Println("✓ Authentication: token is valid")
			}

			// 4. Server reachability check (also confirms Keycloak config
			// endpoint responds, covering both plain connectivity and
			// Keycloak-backed setups).
			if mc != nil {
				if _, err := mc.GetKeycloakURL(); err != nil {
					fmt.Printf("✗ Server reachability: %v\n", err)
					ok = false
				} else {
					fmt.Printf("✓ Server reachable: %s\n", ctx.Server.Server)
				}
			}

			if !ok {
				fmt.Println("\nOne or more checks failed.")
				os.Exit(1)
			}
			fmt.Println("\nAll checks passed ✓ — ready to run imports/tests.")
		},
	}

	return validateCmd
}
