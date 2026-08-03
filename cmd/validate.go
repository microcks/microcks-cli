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
// checks on configuration and connectivity to a Microcks server, without
// performing any import or test action.
func NewValidateCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	var validateCmd = &cobra.Command{

		Use:   "validate",
		Short: "Validate current Microcks CLI configuration and connectivity",
		Long: `Runs pre-flight checks before running imports/tests:
   - config/credentials are resolvable, either via a local context or via
	 --microcksURL together with --keycloakClientId/--keycloakClientSecret
	 for CI/CD (service account) usage
   - the auth token is valid (refreshed / retrieved as needed)
   - the target Microcks server is reachable
 
 Exits with code 0 if all checks pass, or 1 if any check fails.
 
 Note: this command validates connectivity and authentication only. It
 does not check service-account authorization scopes or Async API Minion
 connectivity, which are tracked as a follow-up.`,
		Example: `microcks validate
 microcks validate --microcks-context staging
 microcks validate --microcksURL http://microcks.example.com/api --keycloakClientId my-sa --keycloakClientSecret my-secret`,

		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(runValidate(globalClientOpts))
		},
	}

	return validateCmd
}

// runValidate executes the pre-flight checks and returns a process exit
// code (0 if everything passed, 1 otherwise). Extracted from Run so it can
// be unit tested without the test process itself exiting.
func runValidate(globalClientOpts *connectors.ClientOptions) int {
	ok := true

	directMode := globalClientOpts.ServerAddr != "" &&
		globalClientOpts.ClientId != "" &&
		globalClientOpts.ClientSecret != ""

	var mc connectors.MicrocksClient
	var serverAddr string

	if directMode {
		// CI/CD direct-connection mode: no local config file involved,
		// mirrors the pattern used by `test`/`import` for --microcksURL.
		serverAddr = globalClientOpts.ServerAddr
		mc = connectors.NewMicrocksClient(serverAddr)
		fmt.Printf("✓ Using direct connection: %s\n", serverAddr)

		keycloakURL, err := mc.GetKeycloakURL()
		if err != nil {
			fmt.Printf("✗ Server reachability: %v\n", err)
			return 1
		}
		fmt.Printf("✓ Server reachable: %s\n", serverAddr)

		oauthToken := "unauthenticated-token"
		if keycloakURL != "null" {
			kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)
			oauthToken, err = kc.ConnectAndGetToken()
			if err != nil {
				fmt.Printf("✗ Authentication: %v\n", err)
				return 1
			}
		}
		mc.SetOAuthToken(oauthToken)
		fmt.Println("✓ Authentication: service account token obtained")

	} else {
		// Local config / context mode.
		localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
		if err != nil {
			fmt.Printf("✗ Config file: invalid — %v\n", err)
			return 1
		}
		if localConfig == nil || localConfig.IsEmpty() {
			fmt.Println("✗ No contexts configured. Run 'microcks login <server>', or pass --microcksURL with --keycloakClientId/--keycloakClientSecret for CI/CD.")
			return 1
		}
		fmt.Println("✓ Config file: parsed successfully")

		ctx, err := localConfig.ResolveContext(globalClientOpts.Context)
		if err != nil {
			fmt.Printf("✗ Context resolution: %v\n", err)
			return 1
		}
		fmt.Printf("✓ Context resolved: %q (server: %s)\n", ctx.Name, ctx.Server.Server)
		serverAddr = ctx.Server.Server

		mc, err = connectors.NewClient(*globalClientOpts)
		if err != nil {
			fmt.Printf("✗ Authentication: %v\n", err)
			ok = false
		} else {
			fmt.Println("✓ Authentication: token is valid")
		}

		if mc != nil {
			if _, err := mc.GetKeycloakURL(); err != nil {
				fmt.Printf("✗ Server reachability: %v\n", err)
				ok = false
			} else {
				fmt.Printf("✓ Server reachable: %s\n", serverAddr)
			}
		}
	}

	if !ok {
		fmt.Println("\nOne or more checks failed.")
		return 1
	}
	fmt.Println("\nAll checks passed ✓ — ready to run imports/tests.")
	return 0
}
