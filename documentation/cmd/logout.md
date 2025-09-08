## `microcks logout` – Log Out from a Microcks Context
Removes authentication tokens from the specified context.

### Usage
```bash
microcks logout CONTEXT
```

### Example
```bash
# Log out from the specified context
microcks logout http://localhost:8080   

# Log out from a named context
microcks logout dev-context
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
