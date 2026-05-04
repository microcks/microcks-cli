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
	"strconv"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

func NewImportURLCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var importURLCmd = &cobra.Command{
		Use:   "import-url <specificationFile1URL[:primary],specificationFile2URL[:primary]>",
		Short: "import API artifacts from URL on Microcks server",
		Long:  `import API artifacts from URL on Microcks server`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specificationFiles := args[0]

			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			var mc connectors.MicrocksClient

			if globalClientOpts.ServerAddr != "" && globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {
				// Create client with server address.
				mc = connectors.NewMicrocksClient(globalClientOpts.ServerAddr)

				keycloakURL, err := mc.GetKeycloakURL()
				if err != nil {
					return fmt.Errorf("got error when invoking Microcks client retrieving config: %s", err)
				}

				var oauthToken string = "unauthenticated-token"
				if keycloakURL != "null" {
					// If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
					kc := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)

					oauthToken, err = kc.ConnectAndGetToken()
					if err != nil {
						return fmt.Errorf("got error when invoking Keycloak client: %s", err)
					}
				}

				// Set Auth token.
				mc.SetOAuthToken(oauthToken)
			} else {

				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					return err
				}

				if localConfig == nil {
					return fmt.Errorf("please login to perform operation")
				}

				if globalClientOpts.Context == "" {
					globalClientOpts.Context = localConfig.CurrentContext
				}

				mc, err = connectors.NewClient(*globalClientOpts)
				if err != nil {
					return err
				}
			}
			sepSpecificationFiles := strings.Split(specificationFiles, ",")
			for _, f := range sepSpecificationFiles {
				mainArtifact := true
				secret := ""

				// Check if URL starts with https or http.
				if strings.HasPrefix(f, "https://") || strings.HasPrefix(f, "http://") {
					urlAndMainAtrifactAndSecretName := strings.Split(f, ":")
					n := len(urlAndMainAtrifactAndSecretName)
					f = urlAndMainAtrifactAndSecretName[0] + ":" + urlAndMainAtrifactAndSecretName[1]
					if n > 2 {
						val, err := strconv.ParseBool(urlAndMainAtrifactAndSecretName[2])
						if err != nil {
							return fmt.Errorf("failed to parse mainArtifact flag: %w", err)
						}
						mainArtifact = val
					}
					if n > 3 {
						secret = urlAndMainAtrifactAndSecretName[3]
					}
				}

				// Try downloading the artifact.
				msg, err := mc.DownloadArtifact(f, mainArtifact, secret)
				if err != nil {
					return fmt.Errorf("got error when invoking Microcks client importing Artifact: %s", err)
				}
				fmt.Printf("Microcks has discovered '%s'\n", msg)
			}
			return nil
		},
	}

	return importURLCmd
}
