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
 "github.com/spf13/viper"
)

var importCmd = &cobra.Command{
	Use:   "import <specificationFile1[:primary],specificationFile2[:primary]>",
	Short: "Import API specifications into Microcks",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specificationFiles := args[0]

    microcksURL := viper.GetString("microcksURL")
		keycloakClientID := viper.GetString("keycloakClientId")
		keycloakClientSecret := viper.GetString("keycloakClientSecret")
		insecureTLS := viper.GetBool("insecure")
		caCertPaths := viper.GetString("caCerts")
		verbose := viper.GetBool("verbose")

    if microcksURL == "" {
			fmt.Println("--microcksURL flag is mandatory. Check Usage.")
			os.Exit(1)
		}
		if keycloakClientID == "" {
			fmt.Println("--keycloakClientId flag is mandatory. Check Usage.")
			os.Exit(1)
		}
		if keycloakClientSecret == "" {
			fmt.Println("--keycloakClientSecret flag is mandatory. Check Usage.")
			os.Exit(1)
		}

	// Collect optional HTTPS transport flags.
	config.InsecureTLS = insecureTLS
		config.CaCertPaths = caCertPaths
		config.Verbose = verbose

	// Now we seems to be good ...
	// First - retrieve the Keycloak URL from Microcks configuration.
	mc := connectors.NewMicrocksClient(microcksURL)
	keycloakURL, err := mc.GetKeycloakURL()
	if err != nil {
		fmt.Printf("Got error when invoking Microcks client retrieving config: %s", err)
		os.Exit(1)
	}

  oauthToken := "unauthentifed-token"
	if keycloakURL != "null" {
		//  If Keycloak is enabled, retrieve an OAuth token using Keycloak Client.
		kc := connectors.NewKeycloakClient(keycloakURL, keycloakClientID, keycloakClientSecret)

		oauthToken, err = kc.ConnectAndGetToken()
		if err != nil {
			fmt.Printf("Got error when invoking Keycloak client: %s", err)
			os.Exit(1)
		}
	}

	// Then - for each specificationFile, upload the artifact on Microcks Server.
	mc.SetOAuthToken(oauthToken)

	sepSpecificationFiles := strings.Split(specificationFiles, ",")
	for _, f := range sepSpecificationFiles {
		mainArtifact := true

		// Check if mainArtifact flag is provided.
		if strings.Contains(f, ":") {
			pathAndMainArtifact := strings.Split(f, ":")
			f = pathAndMainArtifact[0]
			mainArtifact, err = strconv.ParseBool(pathAndMainArtifact[1])
			if err != nil {
				fmt.Printf("Cannot parse '%s' as Bool, default to true\n", pathAndMainArtifact[1])
			}
		}

		// Try uploading this artifact.
		msg, err := mc.UploadArtifact(f, mainArtifact)
		if err != nil {
			fmt.Printf("Got error when invoking Microcks client importing Artifact: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Microcks has discovered '%s'\n", msg)
	}
},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.PersistentFlags().String("microcksURL", "", "Microcks API URL")
	importCmd.PersistentFlags().String("keycloakClientId", "", "Keycloak Client ID")
	importCmd.PersistentFlags().String("keycloakClientSecret", "", "Keycloak Client Secret")
	importCmd.PersistentFlags().Bool("insecure", false, "Allow insecure HTTPS connections")
	importCmd.PersistentFlags().String("caCerts", "", "Comma-separated paths of CA cert files")
	importCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")

	// Bind flags to Viper
	viper.BindPFlag("microcksURL", importCmd.PersistentFlags().Lookup("microcksURL"))
	viper.BindPFlag("keycloakClientId", importCmd.PersistentFlags().Lookup("keycloakClientId"))
	viper.BindPFlag("keycloakClientSecret", importCmd.PersistentFlags().Lookup("keycloakClientSecret"))
	viper.BindPFlag("insecure", importCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("caCerts", importCmd.PersistentFlags().Lookup("caCerts"))
	viper.BindPFlag("verbose", importCmd.PersistentFlags().Lookup("verbose"))
}
