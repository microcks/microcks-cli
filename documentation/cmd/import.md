# Import Command

The `import` command in Microcks CLI is used to upload and register API specification artifacts (like OpenAPI, AsyncAPI, Postman collections, etc.) into a Microcks server.

📝 Description

The `import` command enables developers to push one or multiple API artifacts to a Microcks instance. It supports secure authentication via Keycloak and allows custom TLS configurations for secure communication.

📌 Usage
```bash
microcks import <specificationFile1[:primary]>,<specificationFile2[:primary]> \
  --microcksURL <microcks-api-url> \
  --keycloakClientId <client-id> \
  --keycloakClientSecret <client-secret> \
  [--insecure] \
  [--caCerts <cert-paths>] \
  [--verbose]
```
Arguments
- `<specificationFile[:primary]>`:
A comma-separated list of specification file paths to import.
Optionally, each file can be suffixed with `:true` or `:false` to indicate whether it's the primary artifact.

| Flag                    | Type    | Required | Description                                                                 |
|-------------------------|---------|----------|-----------------------------------------------------------------------------|
| `--microcksURL`         | string  | ✅        | The URL of the Microcks API endpoint.                                      |
| `--keycloakClientId`    | string  | ✅        | The Keycloak Service Account Client ID for OAuth2 authentication.          |
| `--keycloakClientSecret`| string  | ✅        | The Keycloak Service Account Client Secret for authentication.             |
| `--insecure`            | bool    | ❌        | Allow insecure TLS connections (e.g., self-signed certs).                  |
| `--caCerts`             | string  | ❌        | Comma-separated paths to additional CA certificate files (PEM format).     |
| `--verbose`             | bool    | ❌        | Enable verbose mode to dump HTTP requests and responses to the console.    |

🧪 Examples
- Basic Import
```bash
microcks import my-api.yaml \
  --microcksURL http://localhost:8080/api \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret
```
- Import Multiple Files with Primary Indicator
```bash
microcks import openapi.yaml:true,postman.json:false \
  --microcksURL https://microcks.example.com/api \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret
```
- Using Custom TLS CA Certificates and Verbose Logging
```bash
microcks-cli import spec.yaml \
  --microcksURL https://microcks.example.com/api \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret \
  --caCerts /etc/ssl/certs/ca1.crt,/etc/ssl/certs/ca2.crt \
  --verbose
```
