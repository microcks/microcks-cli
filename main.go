package main

import (
	"os"

	"github.com/microcks/microcks-cli/cmd"
)

var (
//KeycloakURL      = flag.String("keycloakURL", "", "Keycloak Realm URL")
//KeycloakUsername = flag.String("keycloakUsername", "", "Keycloak Realm ServiceAccount ")
//KeycloakPassword = flag.String("keycloakPassword", "", "Keycloak Realm Account Password")
)

func main() {
	var c cmd.Command

	switch os.Args[1] {
	case "version":
		c = cmd.NewVersionCommand()
	case "help":
		c = cmd.NewHelpCommand()
	case "test":
		c = cmd.NewTestCommand()
	default:
		cmd.NewHelpCommand().Execute()
		os.Exit(1)
	}

	c.Execute()
	return

	/*
		var keycloakURL string
		var keycloakUsername string
		var keycloakPassword string

		flag.StringVar(&keycloakURL, "keycloakURL", "", "Keycloak Realm URL")
		flag.StringVar(&keycloakUsername, "keycloakUsername", "", "Keycloak Realm ServiceAccount ")
		flag.StringVar(&keycloakPassword, "keycloakPassword", "", "Keycloak Realm Account Password")
		flag.Parse()

		if len(keycloakURL) == 0 {
			fmt.Fprintf(os.Stderr, "You must specify a Keycloak Realm URL")
		}

		fmt.Printf("The Keycloak Realm URL is: %s \n", keycloakURL)
		fmt.Printf("The Keycloak Realm ServiceAccount is: %s \n", keycloakUsername)
		fmt.Printf("The Keycloak Realm Account Password is: %s \n", keycloakPassword)

		kc := NewKeycloakClient(keycloakURL, keycloakUsername, keycloakPassword)

		var oauthToken string
		oauthToken, err := kc.ConnectAndGetToken()
		if err != nil {
			fmt.Printf("Got error: %s", err)
		}

		fmt.Printf("Retrieve OAuthToken: %s", oauthToken)
	*/
}
