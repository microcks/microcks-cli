## `microcks test` – Run Tests on Microcks
Runs contract or integration tests against a deployed API using the selected runner.

### Usage
```bash
microcks test <apiName:apiVersion> <testEndpoint> <runner> [flags]
```

### Example
```bash
# Run a basic HTTP test on the local hello-api version 1.0.0
microcks test hello-api:1.0.0 http://localhost:8080/api HTTP

# Run a POSTMAN test on petstore API version 2.0.0, wait up to 10 seconds for completion
microcks test petstore:2.0.0 https://api.example.com POSTMAN --waitFor 10sec

# Run a POSTMAN test on the local Beer Catalog API version 0.9.0 without logining to microcks
microcks test Beer Catalog API:0.9 http://localhost:9090/api/ POSTMAN \
        --microcksURL <microcks-url> \
        --keycloakClientId <client-id> \
        --keycloakClientSecret <client-secret> \
```

### Runner Options
One of:
`HTTP`|`SOAP_HTTP`|`SOAP_UI`|`POSTMAN`|`OPEN_API_SCHEMA`|`ASYNC_API_SCHEMA`|`GRPC_PROTOBUF`|`GRAPHQL_SCHEMA`

### Options
| Flag                   | Description                                                                         |
| ---------------------- | ----------------------------------------------------------------------------------- |
| `-h, --help`           | help for test                                                                       |
| `--waitFor`            | Time to wait for test result. Format: `5sec`, `2000milli`, `1min` (default: `5sec`) |
| `--secretName`         | Secret name for accessing secured test endpoint                                     |
| `--filteredOperations` | Comma-separated list of operations to test                                          |
| `--operationsHeaders`  | Custom headers for operations as JSON string                                        |
| `--oAuth2Context`      | OAuth2 client context as JSON string                                                |


### Options Inherited from Parent Commands
| Flag                     | Description                                 |
| ------------------------ | ------------------------------------------- |
| `--config`               | Path to Microcks config file                |
| `--microcks-context`     | Name of the Microcks context to use         |
| `--verbose`              | Produce dumps of HTTP exchanges             |
| `--insecure-tls`         | Allow insecure HTTPS connections            |
| `--caCerts`              | Comma-separated paths of CA cert files      |
| `--keycloakClientId`     | Keycloak Realm Service Account ClientId     |
| `--keycloakClientSecret` | Keycloak Realm Service Account ClientSecret |
| `--microcksURL`          | Microcks API URL                            |
