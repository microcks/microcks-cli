## `microcks import` – Import API Artifacts into Microcks
Uploads one or more API spec files (e.g., OpenAPI, AsyncAPI, Postman) to the Microcks server and optionally watches them for changes.

### Usage
```bash
microcks import <specFile1[:main]>,<specFile2[:main]> [flags]
```

### Examples
```bash
# Import a single artifact (marked as main)
microcks import ./api.yaml

# Specify mainArtifact flag for each file
microcks import ./api.yaml:false,./schema.json:true

# Import and watch file for changes
microcks import ./api.yaml --watch

# Import specification to microcks without logining to microcks
microck import ./api.yaml \
    --micrcoksURL <microcks-url> \ 
    --keycloakClientId <client-id> \
    --keycloakClientSecret <client-secret> 
```

### Options
| Flag        | Description                                         |
| ----------- | --------------------------------------------------- |
| `-h, --help`| help for import                                     |
| `--watch`   | Watch the file(s) and auto-reimport them on changes |

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
