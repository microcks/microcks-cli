## `microcks health` – Check Microcks Server Health and Diagnostics

Provides quick diagnostics about the configured Microcks server, including reachability, API responsiveness, database status, Keycloak status, Async Minion status, and server version.

### Usage
```bash
microcks health [flags]
```

### Examples
```bash
# Basic health check
microcks health

# Machine-readable JSON output
microcks health --json

# Watch mode (repeated health check every 5 seconds)
microcks health --watch

# Watch mode with custom interval of 10 seconds
microcks health --watch --interval 10s
```

### Options
| Flag | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| `--json` | bool | Output machine-readable JSON | `false` |
| `--watch` | bool | Repeat health checks periodically | `false` |
| `--interval` | duration | Watch interval when --watch is enabled | `5s` |

### Exit Codes
The `microcks health` command returns standard, automation-friendly exit codes:

| Condition | Exit Code | Description |
| :--- | :--- | :--- |
| Healthy | `0` | Server is fully reachable and all critical subsystems are UP. |
| Unhealthy | `1` | Server is unreachable (connection refused, DNS failure, TLS failure, or timeout). |
| Degraded/Partial | `2` | Server is reachable, but the overall status is DOWN or one or more subsystem checks are DOWN (e.g., database disconnected). |

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
