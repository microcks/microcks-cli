## `microcks service` - List and Inspect Microcks Services
Lists services known by the selected Microcks server and retrieves service details.

### Usage
```bash
microcks service list [flags]
microcks service get <id-or-name-version> [flags]
```

### Examples
```bash
# List services from the current context
microcks service list

# List services as JSON for tools and editor integrations
microcks service list --output json

# Get service details by id
microcks service get 64f1d8c9e4b02c1c4d6c7a90 --output json

# Get service details by name and version
microcks service get "E-Commerce Platform API:2.0.0" --output json
```

### Options
| Flag       | Description                                      |
| ---------- | ------------------------------------------------ |
| `-h, --help` | help for service                               |
| `--output` | Output format: `text` (default) or `json`       |
| `--page`   | Page index to fetch for `service list`          |
| `--size`   | Number of services to fetch for `service list`  |

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

### JSON contracts

`service list --output json` writes a JSON array of service summaries. Each
summary includes `id`, `name`, `version`, and `type`, and may include
`operations`.

`service get --output json` writes an object containing `service` and, when
available, `messagesMap`. Integrations should check for the
`service.list.json` and `service.get.json` capabilities before depending on
these contracts.
