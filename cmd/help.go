package cmd

import "fmt"

type helpCommand struct {
}

// NewHelpCommand build a new HelpCommand implementation
func NewHelpCommand() Command {
	return new(helpCommand)
}

// Execute implementation on helpCommand structure
func (c *helpCommand) Execute() {
	fmt.Println("")
	fmt.Println("microcks-cli is a CLI for interacting with Microcks server APIs.")
	fmt.Println("It allows to launch tests with minimal dependencies")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  microcks-cli [command]")
	fmt.Println("")
	fmt.Println("Available Commands:")
	fmt.Println("  version     check this CLI version")
	fmt.Println("  help        display this help message")
	fmt.Println("  test        launch new test on Microcks server")
	fmt.Println("")
	fmt.Println("Use: microcks-cli test <apiName:apiVersion> <testEndpoint> <runner> \\")
	fmt.Println("   --microcksURL=<> --waitFor=5sec \\")
	fmt.Println("   --keycloakURL=<> --keycloakUsername=<> --keycloakPassword=<>")
	fmt.Println("")
	fmt.Println("Args: ")
	fmt.Println("  <apiName:apiVersion>   Exemple: 'Beer Catalog API:0.9'")
	fmt.Println("  <testEndpoint>         URL where is deployed implementation to test")
	fmt.Println("  <runner>               Test strategy (one of: HTTP, SOAP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA)")
	fmt.Println("")
	fmt.Println("Flags: ")
	fmt.Println("  --microcksURL        Microcks API endpoint")
	fmt.Println("  --waitFor        	Time to wait for test to finish (int + one of: milli, sec, min)")
	fmt.Println("  --keycloakURL        Keycloak Realm API endpoint for Microcks")
	fmt.Println("  --keycloakUsername   Keycloak Realm ServiceAccount")
	fmt.Println("  --keycloakPassword   Keycloak Realm Account Password")
	fmt.Println("")
}
