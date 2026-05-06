## `microcks services list` – List Services Imported in Microcks

Lists the services (APIs and mocks) currently imported in the connected Microcks instance.

### Usage
```bash
microcks services list [flags]
```

### Flags
| Flag     | Default | Description                            |
| -------- | ------- | -------------------------------------- |
| `--page` | `0`     | Page of services to retrieve (0-indexed) |
| `--size` | `20`    | Number of services per page            |

### Examples
```bash
# List services using the current context
microcks services list

# List the second page of results with 10 services per page
microcks services list --page 1 --size 10

# List services against a specific Microcks instance
microcks services list --microcksURL http://localhost:8585 \
  --keycloakClientId my-client \
  --keycloakClientSecret my-secret
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
