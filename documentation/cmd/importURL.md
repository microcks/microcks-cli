# ImportURL Command

The `import-url` command in Microcks CLI is used to upload and register API specification artifacts(like OpenAPI, AsyncAPI, Postman collections, etc.) from URLs into a Microcks server.

📝 Description

The `import-url` command provides a convenient way to register API specifications (OpenAPI, AsyncAPI, Postman collections, etc.) hosted online without needing to download them manually. This is useful for CI/CD pipelines or importing frequently updated remote specs.

📌 Usage
```bash
microcks import-url <url1[:primary]>,<url2[:primary]> \
  --microcksURL <microcks-api-url> \
  --keycloakClientId <client-id> \
  --keycloakClientSecret <client-secret> \
  [--insecure] \
  [--caCerts <cert-paths>] \
  [--verbose]
```

Arguments
-   `<url[:primary]>`:
A comma-separated list of publicly accessible URLs pointing to specification files.
Optionally, each URL can be suffixed with `:true` or `:false` to mark it as a primary artifact.


| Flag                    | Type    | Required | Description                                                                 |
|-------------------------|---------|----------|-----------------------------------------------------------------------------|
| `--microcksURL`         | string  | ✅        | The URL of the Microcks API endpoint.                                      |
| `--keycloakClientId`    | string  | ✅        | The Keycloak Service Account Client ID for OAuth2 authentication.          |
| `--keycloakClientSecret`| string  | ✅        | The Keycloak Service Account Client Secret for authentication.             |
| `--insecure`            | bool    | ❌        | Allow insecure TLS connections (e.g., self-signed certs).                  |
| `--caCerts`             | string  | ❌        | Comma-separated paths to additional CA certificate files (PEM format).     |
| `--verbose`             | bool    | ❌        | Enable verbose mode to dump HTTP requests and responses to the console.    |

🧪 Examples
- Importing from a Single URL
```bash
microcks import-url https://example.com/api/openapi.yaml \
  --microcksURL http://localhost:8080/api \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret
```

- Importing Multiple Remote Artifacts with Primary Designation
```bash
microcks import-url https://example.com/openapi.yaml:true,https://example.com/postman.json:false \
  --microcksURL https://microcks.example.com/api \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret
```

- Using Custom TLS CA Certificates and Verbose Logging
```bash
microcks import-url https://mydomain.com/api/spec.yaml \
  --microcksURL https://microcks.example.com/api \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret \
  --caCerts /etc/ssl/certs/ca1.crt,/etc/ssl/certs/ca2.crt \
  --verbose
```
