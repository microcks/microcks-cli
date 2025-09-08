## `microcks start` – Start a Local Microcks Instance
Starts a Microcks instance using Docker or Podman and configures it as the current CLI context.

### Usage
```bash
microcks start [flags]
```

### Example
```bash
# Start a Microcks instance
microcks start

# Define your port (by default 8585)
microcks start --port [Port you want]

# Define your driver (by default docker)
microcks start --driver [driver you wnat either 'docker' or 'podman']

# Define name of your microcks container/instance
microcks start --name [name of you container/instance]

# Auto remove the container on exit
microcks start --rm
```

### Options
| Flag        | Description                                                                      |
| ----------- | -------------------------------------------------------------------------------- |
| `-h, --help`| help for start                                                                   |
| `--name`    | Name for the Microcks instance (default: `microcks`)                             |
| `--port`    | Host port to expose Microcks (default: `8585`)                                   |
| `--image`   | Container image to use (default: `quay.io/microcks/microcks-uber:latest-native`) |
| `--rm`      | Auto-remove the container when it exits (like Docker `--rm`)                     |
| `--driver`  | Container driver to use (`docker` or `podman`, default: `docker`)                |

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

