## `microcks import-url` – Import API Artifacts from URL
Imports API specification files (OpenAPI, AsyncAPI, etc.) hosted at a remote URL into the Microcks server.

### Usage
```bash
microcks import-url <specURL1[:main][:secret]>,<specURL2[:main][:secret]> [flags]
```

### URL suffix parsing
You can optionally append metadata suffixes to each URL:

- `:<main>` where `<main>` is `true` or `false`
- `:<main>:<secret>` to additionally specify a secret name

The CLI parses these suffixes from the **rightmost** `:` characters only, so normal URLs containing `:` (scheme, ports, etc.) are preserved.

### Example
```bash
# Import a single artifact (marked as main)
microcks import-url https://example.com/openapi.yaml

# Specify mainArtifact flag for each file
microcks import-url https://example.com/spec1.yaml:true,https://example.com/spec2.yaml:false

# URL with port + :main suffix (port/path are preserved)
microcks import-url http://localhost:8080/spec.yaml:true

# URL with port + :main + :secret
microcks import-url http://localhost:8080/spec.yaml:true:mySecret

# Import specification to microcks without logining to microcks
microcks import-url https://example.com/openapi.yaml \
    --microcksURL <microcks-url> \ 
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