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
	"text/tabwriter"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewTestsCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	testsCmd := &cobra.Command{
		Use:   "tests",
		Short: "Manage Microcks test results",
		Long:  `Manage Microcks test results`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	testsCmd.AddCommand(newTestsListCommand(globalClientOpts))
	return testsCmd
}

func newTestsListCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		page int
		size int
	)
	listCmd := &cobra.Command{
		Use:   "list <serviceRef>",
		Short: "List test results for a service",
		Long:  `List test results for a service`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("tests list command requires <serviceRef> arg")
				os.Exit(1)
			}
			serviceRef := args[0]

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
				_, err = localConfig.ResolveContext(globalClientOpts.Context)
				errors.CheckError(err)
			}

			results, err := mc.GetTestResults(serviceRef, page, size)
			if err != nil {
				fmt.Printf("Got error when invoking Microcks client listing test results: %s", err)
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tDATE\tRUNNER\tENDPOINT\tSUCCESS")
			for _, r := range results {
				date := time.UnixMilli(r.TestDate).Format("2006-01-02 15:04:05")
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%v\n",
					r.ID,
					date,
					r.RunnerType,
					r.TestedEndpoint,
					r.Success,
				)
			}
			w.Flush()
		},
	}
	listCmd.Flags().IntVar(&page, "page", 0, "Page of results to retrieve")
	listCmd.Flags().IntVar(&size, "size", 20, "Number of results per page")
	return listCmd
}
