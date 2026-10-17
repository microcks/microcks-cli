## `microcks test` – Run Tests on Microcks
Runs contract or integration tests against a deployed API using the selected runner.

### Usage
```bash
microcks test <apiName:apiVersion> <testEndpoint> <runner> [flags]
microcks test list [flags]
microcks test get <testResultId> [flags]
```

### Example
```bash
# Run a basic HTTP test on the local hello-api version 1.0.0
microcks test hello-api:1.0.0 http://localhost:8080/api HTTP

# Run a POSTMAN test on petstore API version 2.0.0, wait up to 10 seconds for completion
microcks test petstore:2.0.0 https://api.example.com POSTMAN --waitFor 10sec

# Run a POSTMAN test on the local Beer Catalog API version 0.9.0 without logining to microcks
microcks test Beer Catalog API:0.9 http://localhost:9090/api/ POSTMAN \
        --microcksURL <microcks-url> \
        --keycloakClientId <client-id> \
        --keycloakClientSecret <client-secret> \

# List recent test results as JSON
microcks test list --output json

# List recent test results for a service
microcks test list --serviceId <service-id> --output json

# Get a full test result as JSON
microcks test get <test-result-id> --output json
```

### Runner Options
One of:
`HTTP`|`SOAP_HTTP`|`SOAP_UI`|`POSTMAN`|`OPEN_API_SCHEMA`|`ASYNC_API_SCHEMA`|`GRPC_PROTOBUF`|`GRAPHQL_SCHEMA`

### Local dry-run (ephemeral container)

`--dry-run` runs the contract test without any infrastructure: no running Microcks
server, no Keycloak credentials, no prior import. The CLI starts an ephemeral
Microcks container, imports `--artifact`, runs the test against your endpoint,
prints the result and tears the container down.

```bash
# One-shot: run once, tear down, exit (exit code reflects the test result)
microcks test --dry-run \
  --artifact ./openapi.yaml \
  "Pastry API:1.0.0" \
  http://localhost:3000 \
  OPEN_API_SCHEMA

# Watch mode: keep the container alive, re-import and re-run on every save
microcks test --dry-run --watch \
  --artifact ./openapi.yaml \
  "Pastry API:1.0.0" \
  http://localhost:3000 \
  OPEN_API_SCHEMA
```

| Flag              | Default                                        | Description                                              |
| ----------------- | ---------------------------------------------- | -------------------------------------------------------- |
| `--dry-run`       | `false`                                        | Run against an ephemeral container instead of a server   |
| `--artifact`      | _(required with `--dry-run`)_                  | Local spec file imported as the main artifact            |
| `--image`         | `quay.io/microcks/microcks-uber:latest-native` | Uber image override; must be a `*-native` tag            |
| `--ready-timeout` | `90s`                                          | How long to wait for the container to become ready       |
| `--watch`         | `false`                                        | Re-import and re-run the test when the artifact changes  |
| `--driver`        | _(auto-detect)_                                | Container runtime: `docker` or `podman`                  |

`--artifact`, `--watch` and `--driver` are only meaningful with `--dry-run`;
passing one without it is a usage error (exit code `2`).

The command accepts only the `-native` image flavor. That variant runs without
Keycloak, which is what makes a zero-configuration dry-run possible.

#### Choosing a container runtime

Without `--driver`, the CLI selects Podman only when `DOCKER_HOST` is unset, `podman`
is on the `PATH` and `docker` is not. An explicitly configured `DOCKER_HOST` takes
precedence, so an existing Docker or remote-daemon setup keeps working unchanged.

`--driver podman` verifies the Podman socket before starting the container. When the
socket is unreachable, the command fails with an environment error and exit code
`14` instead of falling back to Docker. Start the socket with `podman machine start`
on macOS and Windows, or `systemctl --user start podman.socket` on Linux.

#### Dry-run behavior

- A `localhost` or `127.0.0.1` test endpoint is rewritten to
  `host.testcontainers.internal` and its port exposed to the container, so an API
  running on your machine is reachable without any change on your side.
- The container is removed on every exit path, including `Ctrl+C` mid-test.
- In watch mode the CLI watches the artifact's _directory_ rather than the file.
  Many editors save by replacing the file, which cancels a watch registered on the
  file itself. The CLI also debounces bursts of save events.
- A spec that is invalid while you are still editing it does not stop the session:
  the CLI reports the failed re-import and keeps watching. The next valid save
  recovers.

Watch mode combined with `--output json` emits a newline-delimited event stream
instead of rendered results. See [JSON output contracts](../json-output.md).

### Options
| Flag                   | Description                                                                         |
| ---------------------- | ----------------------------------------------------------------------------------- |
| `-h, --help`           | help for test                                                                       |
| `--waitFor`            | Time to wait for test result. Format: `5sec`, `2000milli`, `1min` (default: `5sec`) |
| `--secretName`         | Secret name for accessing secured test endpoint                                     |
| `--filteredOperations` | Comma-separated list of operations to test                                          |
| `--operationsHeaders`  | Custom headers for operations as JSON string                                        |
| `--oAuth2Context`      | OAuth2 client context as JSON string                                                |
| `--output`             | Output format: `text` (default), `json`, `yaml`, or `github-actions`                |

The dry-run flags (`--dry-run`, `--artifact`, `--image`, `--ready-timeout`,
`--watch`, `--driver`) are listed under [Local dry-run](#local-dry-run-ephemeral-container) above.

### `test list` and `test get` Options
| Flag          | Description                                      |
| ------------- | ------------------------------------------------ |
| `--output`    | Output format: `text` (default) or `json`        |
| `--serviceId` | Service id to filter `test list` results         |
| `--page`      | Page index to fetch for `test list`              |
| `--size`      | Number of test results to fetch for `test list`  |


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

### Structured output contracts

`test list --output json` writes a JSON array of test-result summaries.
`test get --output json` writes one complete test result, including its
`testCaseResults` when present. `microcks test ... --output json` and one-shot
dry-run tests write the completed test result to stdout while progress and
diagnostics go to stderr.

Field-level schemas, the newline-delimited dry-run watch events, and the
capability identifiers to check before depending on any of them are documented in
[JSON output contracts](../json-output.md).
