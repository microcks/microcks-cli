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
	"strings"
	"text/tabwriter"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewServicesCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "Manage Microcks services",
		Long:  `Manage Microcks services`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	servicesCmd.AddCommand(newServicesListCommand(globalClientOpts))
	return servicesCmd
}

func newServicesListCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		page int
		size int
	)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List services imported in Microcks",
		Long:  `List services imported in Microcks`,
		Example: `# List services using current context
microcks services list

# List services with pagination
microcks services list --page 1 --size 10`,
		Run: func(cmd *cobra.Command, args []string) {
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			var mc connectors.MicrocksClient

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {
				mc = connectors.NewMicrocksClient(globalClientOpts.ServerAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					fmt.Printf("Got error when invoking Microcks client retrieving config: %s", err)
					os.Exit(1)
				}

				var oauthToken string = "unauthenticated-token"
				if keycloakURL != "null" {
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)
					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						fmt.Printf("Got error when invoking Keycloak client: %s", err)
						os.Exit(1)
					}
				}
				mc.SetOAuthToken(oauthToken)
			} else {
				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					fmt.Println(err)
					return
				}

				if localConfig == nil {
					fmt.Println("Please login to perform operation...")
					return
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					fmt.Printf("error %v", err)
					return
				}
			}

			services, err := mc.GetServices(page, size)
			if err != nil {
				fmt.Printf("Got error when listing services: %s", err)
				os.Exit(1)
			}

			if len(services) == 0 {
				fmt.Println("No services found")
				return
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer func() { _ = w.Flush() }()
			columnNames := []string{"NAME", "VERSION", "TYPE"}
			_, err = fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))
			errors.CheckError(err)

			for _, svc := range services {
				_, err = fmt.Fprintf(w, "%s\t%s\t%s\n", svc.Name, svc.Version, svc.Type)
				errors.CheckError(err)
			}
		},
	}

	listCmd.Flags().IntVar(&page, "page", 0, "Page of services to retrieve (0-indexed)")
	listCmd.Flags().IntVar(&size, "size", 20, "Number of services per page")
	return listCmd
}
