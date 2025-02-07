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
package root

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

var (
	microcksURL         string
	keycloakClientID    string
	keycloakClientSecret string
	insecureTLS        bool
	caCertPaths        string
	verbose            bool
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <specificationFile1[:primary],specificationFile2[:primary]>",
	Short: "Import API specifications into Microcks",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importArtifacts(args[0])
	},
}

func init() {

	importCmd.Flags().StringVar(&microcksURL, "microcksURL", "", "Microcks API URL (required)")
	importCmd.Flags().StringVar(&keycloakClientID, "keycloakClientId", "", "Keycloak Realm Service Account ClientId (required)")
	importCmd.Flags().StringVar(&keycloakClientSecret, "keycloakClientSecret", "", "Keycloak Realm Service Account ClientSecret (required)")
	importCmd.Flags().BoolVar(&insecureTLS, "insecure", false, "Whether to accept insecure HTTPS connection")
	importCmd.Flags().StringVar(&caCertPaths, "caCerts", "", "Comma separated paths of CRT files to add to Root CAs")
	importCmd.Flags().BoolVar(&verbose, "verbose", false, "Produce dumps of HTTP exchanges")

	importCmd.MarkFlagRequired("microcksURL")
	importCmd.MarkFlagRequired("keycloakClientId")
	importCmd.MarkFlagRequired("keycloakClientSecret")
}

func importArtifacts(specificationFiles string) {
	mc := connectors.NewMicrocksClient(microcksURL)
	keycloakURL, err := mc.GetKeycloakURL()
	if err != nil {
		fmt.Printf("Error retrieving Microcks config: %s\n", err)
		os.Exit(1)
	}

	var oauthToken = "unauthenticated-token"
	if keycloakURL != "null" {
		kc := connectors.NewKeycloakClient(keycloakURL, keycloakClientID, keycloakClientSecret)
		oauthToken, err = kc.ConnectAndGetToken()
		if err != nil {
			fmt.Printf("Error connecting to Keycloak: %s\n", err)
			os.Exit(1)
		}
	}

	mc.SetOAuthToken(oauthToken)
	sepSpecificationFiles := strings.Split(specificationFiles, ",")
	for _, f := range sepSpecificationFiles {
		mainArtifact := true
		if strings.Contains(f, ":") {
			parts := strings.Split(f, ":")
			f = parts[0]
			mainArtifact, err = strconv.ParseBool(parts[1])
			if err != nil {
				fmt.Printf("Invalid boolean value '%s', defaulting to true\n", parts[1])
			}
		}

		msg, err := mc.UploadArtifact(f, mainArtifact)
		if err != nil {
			fmt.Printf("Error importing artifact: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Microcks discovered: '%s'\n", msg)
	}
}
