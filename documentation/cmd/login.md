## `microcks login` – Authenticate with a Microcks Instance
Log in to a Microcks instance using username/password or SSO. Creates or updates a CLI context with authentication details.

### Examples
```bash
# Login to microcks using a username and password
microcks login http://locahost:8080

# Provide name to your logged in context (Defautl context name is server name)
microcks login http://localhost:8080 --name

# Provide username and password as flags
microcks login http://localhost:8080 --username --password

# Perform SSO login
microcks login http://localhost:8080 --sso

# Change port callback server for SSO login
microcks login http://localhost:8080 --sso --sso-port

# Get OAuth URI instead of getting redirect to browser for SSO login
microcks login http://localhost:8080 --sso --sso-launch-browser=false
```

### Options
| Flag                   | Description                                                  |
| ---------------------- | ------------------------------------------------------------ |
| `-h, --help`           | help for login                                               |
| `--name`               | Name to assign the context (default: server URL)             |
| `--username`           | Username for login                                           |
| `--password`           | Password for login                                           |
| `--sso`                | Perform Single Sign-On (OIDC-based) login                    |
| `--sso-launch-browser` | Launch system browser for SSO (default: `true`)              |
| `--sso-port`           | Local port to use for SSO callback server (default: `58085`) |

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
