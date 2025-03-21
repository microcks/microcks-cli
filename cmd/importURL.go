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

func NewImportURLCommand() *cobra.Command {
	var (
		microcksURL          string
		keycloakClientID     string
		keycloakClientSecret string
		insecureTLS          bool
		caCertPaths          string
		verbose              bool
	)
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

			// Collect optional HTTPS transport flags.
			if insecureTLS {
				config.InsecureTLS = true
			}
			if len(caCertPaths) > 0 {
				config.CaCertPaths = caCertPaths
			}
			if verbose {
				config.Verbose = true
			}

			// Now we seems to be good ...
			// First - retrieve the Keycloak URL from Microcks configuration.
			mc := connectors.NewMicrocksClient(microcksURL)
			keycloakURL, err := mc.GetKeycloakURL()
			if err != nil {
				fmt.Printf("Got error when invoking Microcks client retrieving config: %s", err)
				os.Exit(1)
			}

			var oauthToken string = "unauthentifed-token"
			if keycloakURL != "null" {
				//  If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
				kc := connectors.NewKeycloakClient(keycloakURL, keycloakClientID, keycloakClientSecret)

				oauthToken, err = kc.ConnectAndGetToken()
				if err != nil {
					fmt.Printf("Got error when invoking Keycloack client: %s", err)
					os.Exit(1)
				}
			}

			// Then - for each specificationFile, upload the artifact on Microcks Server.
			mc.SetOAuthToken(oauthToken)

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
	importURLCmd.Flags().StringVar(&microcksURL, "microcksURL", "", "Microcks API URL")
	importURLCmd.Flags().StringVar(&keycloakClientID, "keycloakClientId", "", "Keycloak Realm Service Account ClientId")
	importURLCmd.Flags().StringVar(&keycloakClientSecret, "keycloakClientSecret", "", "Keycloak Realm Service Account ClientSecret")
	importURLCmd.Flags().BoolVar(&insecureTLS, "insecure", false, "Whether to accept insecure HTTPS connection")
	importURLCmd.Flags().StringVar(&caCertPaths, "caCerts", "", "Comma separated paths of CRT files to add to Root CAs")
	importURLCmd.Flags().BoolVar(&verbose, "verbose", false, "Produce dumps of HTTP exchanges")

	//Marking flags 'required'
	importURLCmd.MarkFlagRequired("microcksURL")
	importURLCmd.MarkFlagRequired("keycloakClientId")
	importURLCmd.MarkFlagRequired("keycloakClientSecret")

	return importURLCmd
}
