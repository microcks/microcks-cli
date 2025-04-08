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
	fmt.Println("It allows to launch tests or import API artifacts with minimal dependencies")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  microcks-cli [command]")
	fmt.Println("")
	fmt.Println("Available Commands:")
	fmt.Println("  version     check this CLI version")
	fmt.Println("  help        display this help message")
	fmt.Println("  test        launch new test on Microcks server")
	fmt.Println("  import      import API artifacts on Microcks server")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("Use: microcks-cli test <apiName:apiVersion> <testEndpoint> <runner> \\")
	fmt.Println("   --microcksURL=<> --waitFor=5sec \\")
	fmt.Println("   --keycloakClientId=<> --keycloakClientSecret=<>")
	fmt.Println("")
	fmt.Println("Args: ")
	fmt.Println("  <apiName:apiVersion>   Exemple: 'Beer Catalog API:0.9'")
	fmt.Println("  <testEndpoint>         URL where is deployed implementation to test")
	fmt.Println("  <runner>               Test strategy (one of: HTTP, SOAP, SOAP_UI, POSTMAN, OPEN_API_SCHEMA, ASYNC_API_SCHEMA, GRPC_PROTOBUF, GRAPHQL_SCHEMA)")
	fmt.Println("")
	fmt.Println("Flags: ")
	fmt.Println("  --microcksURL          Microcks API endpoint")
	fmt.Println("  --waitFor       		  Time to wait for test to finish (int + one of: milli, sec, min)")
	fmt.Println("  --keycloakClientId     Keycloak Realm Service Account ClientId")
	fmt.Println("  --keycloakClientSecret Keycloak Realm Service Account ClientSecret")
	fmt.Println("")
	fmt.Println("Advanced options: ")
	fmt.Println("  --secretName     	  Secret to use for connecting test endpoint")
	fmt.Println("  --filteredOperations   JSON list of operations to launch a test for")
	fmt.Println("  --operationsHeaders    Override of operations headers as JSON string")
	fmt.Println("  --oAuth2Context    	  JSON Spec of an OAuth2 client context as JSON string")
	fmt.Println("  --insecure    	  	  Whether to accept insecure HTTPS connection (false by default)")
	fmt.Println("  --caCerts    	  	  Comma separated paths of CRT files to add to Root CAs")
	fmt.Println("  --verbose    	  	  Produce dumps of HTTP exchanges (false by default)")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("Use: microcks-cli import <specificationFile1[:primary],specificationFile2[:primary]> \\")
	fmt.Println("   --microcksURL=<> \\")
	fmt.Println("   --keycloakClientId=<> --keycloakClientSecret=<>")
	fmt.Println("")
	fmt.Println("Args: ")
	fmt.Println("  <specificationFile1[:primary],specificationFile2[:primary]>   Exemple: 'specs/my-openapi.yaml:true,specs/my-postmancollection.json:false'")
	fmt.Println("")
	fmt.Println("Flags: ")
	fmt.Println("  --microcksURL          Microcks API endpoint")
	fmt.Println("  --keycloakClientId     Keycloak Realm Service Account ClientId")
	fmt.Println("  --keycloakClientSecret Keycloak Realm Service Account ClientSecret")
	fmt.Println("")
	fmt.Println("Advanced options: ")
	fmt.Println("  --insecure    	  	  Whether to accept insecure HTTPS connection (false by default)")
	fmt.Println("  --caCerts    	  	  Comma separated paths of CRT files to add to Root CAs")
	fmt.Println("  --verbose    	  	  Produce dumps of HTTP exchanges (false by default)")
	fmt.Println("")
}
