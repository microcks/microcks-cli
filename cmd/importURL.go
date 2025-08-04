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
	"github.com/spf13/cobra"
)

func NewImportURLCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var importURLCmd = &cobra.Command{
		Use:   "import-url",
		Short: "import API artifacts from URL on Microcks server",
		Long:  `import API artifacts from URL on Microcks server`,
		Run: func(cmd *cobra.Command, args []string) {
			// Parse subcommand args first.
			if len(args) == 0 {
				fmt.Println("import-url command require <specificationFile1URL[:primary],specificationFile2URL[:primary]> args")
				os.Exit(1)
			}

			specificationFiles := args[0]

			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			var mc connectors.MicrocksClient

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {
				// create client with server address
				mc = connectors.NewMicrocksClient(globalClientOpts.ServerAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					fmt.Printf("Got error when invoking Microcks client retrieving config: %s", err)
					os.Exit(1)
				}

				var oauthToken string = "unauthentifed-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						fmt.Printf("Got error when invoking Keycloack client: %s", err)
						os.Exit(1)
					}
					//fmt.Printf("Retrieve OAuthToken: %s", oauthToken)
				}

				//Set Auth token
				mc.SetOAuthToken(oauthToken)
			} else {

				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					fmt.Println(err)
					return
				}

				if localConfig == nil {
					fmt.Println("Please login to perform opertion...")
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
			sepSpecificationFiles := strings.Split(specificationFiles, ",")
			for _, f := range sepSpecificationFiles {
				mainArtifact := true
				secret := ""

				// Check if URL starts with https or http
				if strings.HasPrefix(f, "https://") || strings.HasPrefix(f, "http://") {
					urlAndMainAtrifactAndSecretName := strings.Split(f, ":")
					n := len(urlAndMainAtrifactAndSecretName)
					f = urlAndMainAtrifactAndSecretName[0] + ":" + urlAndMainAtrifactAndSecretName[1]
					if n > 2 {
						val, err := strconv.ParseBool(urlAndMainAtrifactAndSecretName[2])
						if err != nil {
							fmt.Println(err)
						}
						mainArtifact = val
					}
					if n > 3 {
						secret = urlAndMainAtrifactAndSecretName[3]
					}
				}

				// Try downloading the artifcat
				msg, err := mc.DownloadArtifact(f, mainArtifact, secret)
				if err != nil {
					fmt.Printf("Got error when invoking Microcks client importing Artifact: %s", err)
					os.Exit(1)
				}
				fmt.Printf("Microcks has discovered '%s'\n", msg)
			}
		},
	}

	return importURLCmd
}
