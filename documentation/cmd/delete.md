## `microcks delete` – Delete an API from Microcks
Deletes a specific API (service + version) from the Microcks server.

### Usage
```bash
microcks delete <serviceName:version> [flags]
```

### Example
```bash
# Delete the local 'Simple' API version '1.1'
microcks delete "Simple:1.1"

# Delete without previously logining to microcks
microcks delete "Simple:1.1" \
        --microcksURL <microcks-url> \
        --keycloakClientId <client-id> \
        --keycloakClientSecret <client-secret>
```

### Options
| Flag                   | Description                                                                         |
| ---------------------- | ----------------------------------------------------------------------------------- |
| `-h, --help`           | help for delete                                                                     |

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
