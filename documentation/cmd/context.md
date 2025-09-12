## `microcks context` – Manage CLI Contexts
Switch between, list, or delete configured Microcks CLI contexts.

### Usage
```bash
microcks context [CONTEXT] [flags]
```

### Examples
```bash
# List all available contexts
microcks context/ctx             

# Switch to this context
microcks context/ctx http://localhost:8080   

# Delete the context
microcks context/ctx http://localhost:8080 --delete/-d 
```
### Options
| Flag           | Description                  |
| -------------- | ---------------------------- |
| `-d, --delete` | Delete the specified context |
| `-h, --help`   | help for context             |

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



