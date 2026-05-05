## `microcks completion` – Generate Shell Completion Scripts
Generates shell completion scripts for bash, zsh, fish, and PowerShell.

### Usage
```bash
microcks completion [bash|zsh|fish|powershell]
```

### Example
```bash
# Generate bash completion
microcks completion bash

# Generate zsh completion
microcks completion zsh

# Generate fish completion
microcks completion fish

# Generate PowerShell completion
microcks completion powershell
```

### Options
| Flag         | Description     |
| ------------ | --------------- |
| `-h, --help` | help for completion |

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
