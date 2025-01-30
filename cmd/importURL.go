package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
)

type importURLCommand struct {
}

func NewImportURLCommand() Command {
	return new(importURLCommand)
}

func (c *importURLCommand) Execute() {
	// Parse subcommand args first.
	if len(os.Args) < 2 {
		fmt.Println("import-url command require <specificationFile1URL[:primary],specificationFile2URL[:primary]> args")
		os.Exit(1)
	}

	specificationFiles := os.Args[2]

	// Then parse flags.
	importCmd := flag.NewFlagSet("import-url", flag.ExitOnError)

	var microcksURL string
	var keycloakURL string
	var keycloakClientID string
	var keycloakClientSecret string
	var insecureTLS bool
	var caCertPaths string
	var verbose bool

	importCmd.StringVar(&microcksURL, "microcksURL", "", "Microcks API URL")
	importCmd.StringVar(&keycloakClientID, "keycloakClientId", "", "Keycloak Realm Service Account ClientId")
	importCmd.StringVar(&keycloakClientSecret, "keycloakClientSecret", "", "Keycloak Realm Service Account ClientSecret")
	importCmd.BoolVar(&insecureTLS, "insecure", false, "Whether to accept insecure HTTPS connection")
	importCmd.StringVar(&caCertPaths, "caCerts", "", "Comma separated paths of CRT files to add to Root CAs")
	importCmd.BoolVar(&verbose, "verbose", false, "Produce dumps of HTTP exchanges")
	importCmd.Parse(os.Args[3:])

	// Validate presence and values of flags.
	if len(microcksURL) == 0 {
		fmt.Println("--microcksURL flag is mandatory. Check Usage.")
		os.Exit(1)
	}
	if len(keycloakClientID) == 0 {
		fmt.Println("--keycloakClientId flag is mandatory. Check Usage.")
		os.Exit(1)
	}
	if len(keycloakClientSecret) == 0 {
		fmt.Println("--keycloakClientSecret flag is mandatory. Check Usage.")
		os.Exit(1)
	}

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
}
