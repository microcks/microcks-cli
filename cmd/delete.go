package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

func NewDeleteCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	var deleteCmd = &cobra.Command{
		Use:   "delete <serviceName:version>",
		Short: "Delete an API from Microcks server",
		Long:  "Delete an API (service + version) from Microcks server",
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {

			input := args[0]

			// Validate input format
			if !strings.Contains(input, ":") {
				fmt.Println("delete requires <serviceName:version>")
				os.Exit(1)
			}

			parts := strings.SplitN(input, ":", 2)
			service := parts[0]
			version := parts[1]

			if service == "" || version == "" {
				fmt.Println("delete requires both serviceName and version (neither can be empty)")
				os.Exit(1)
			}

			// Load config (same as import)
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
			if err != nil {
				fmt.Println(err)
				return
			}

			var mc connectors.MicrocksClient

			// Same auth logic as import (DO NOT DUPLICATE BADLY → reuse later)
			if globalClientOpts.ServerAddr != "" &&
				globalClientOpts.ClientId != "" &&
				globalClientOpts.ClientSecret != "" {

				mc = connectors.NewMicrocksClient(globalClientOpts.ServerAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					fmt.Printf("Error retrieving config: %s", err)
					os.Exit(1)
				}

				token := "unauthenticated-token"

				if keycloakURL != "null" {
					kc := connectors.NewKeycloakClient(
						keycloakURL,
						globalClientOpts.ClientId,
						globalClientOpts.ClientSecret,
					)

					token, err = kc.ConnectAndGetToken()
					if err != nil {
						fmt.Printf("Auth error: %s", err)
						os.Exit(1)
					}
				}

				mc.SetOAuthToken(token)

			} else {
				if localConfig == nil {
					fmt.Println("Please login to perform operation...")
					return
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					fmt.Printf("error %v", err)
					return
				}
			}

			err = mc.DeleteService(service, version)
			if err != nil {
				fmt.Printf("Delete failed: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("Deleted service '%s:%s'\n", service, version)
		},
	}

	return deleteCmd
}
