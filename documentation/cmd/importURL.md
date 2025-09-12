## `microcks import-url` – Import API Artifacts from URL
Imports API specification files (OpenAPI, AsyncAPI, etc.) hosted at a remote URL into the Microcks server.

### Usage
```bash
microcks import-url <specURL1[:main][:secret]>,<specURL2[:main][:secret]> [flags]
```

### Example
```bash
# Import a single artifact (marked as main)
microcks import-url https://example.com/openapi.yaml

# Specify mainArtifact flag for each file
microcks import-url https://example.com/spec1.yaml:true,https://example.com/spec2.yaml:false

# Import specification to microcks without logining to microcks
microck import-url https://example.com/openapi.yaml \
    --micrcoksURL <microcks-url> \ 
    --keycloakClientId <client-id> \
    --keycloakClientSecret <client-secret> 
```

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