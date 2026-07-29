# Microcks CLI

Simple CLI for interacting with Microcks server APIs.
It allows launching tests or import API artifacts with minimal dependencies.

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/microcks/microcks-cli/build-verify.yml?logo=github&style=for-the-badge)](https://github.com/microcks/microcks-cli/actions)
[![Container](https://img.shields.io/badge/dynamic/json?color=blue&logo=docker&style=for-the-badge&label=Quay.io&query=tags[1].name&url=https://quay.io/api/v1/repository/microcks/microcks-cli/tag/?limit=10&page=1&onlyActiveTags=true)](https://quay.io/repository/microcks/microcks-cli?tab=tags)
[![License](https://img.shields.io/github/license/microcks/microcks-cli?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![Project Chat](https://img.shields.io/badge/discord-microcks-pink.svg?color=7289da&style=for-the-badge&logo=discord)](https://microcks.io/discord-invite/)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/microcks-cli-image&style=for-the-badge)](https://artifacthub.io/packages/search?repo=microcks-cli-image)
[![CNCF Landscape](https://img.shields.io/badge/CNCF%20Landscape-5699C6?style=for-the-badge&logo=cncf)](https://landscape.cncf.io/?item=app-definition-and-development--application-definition-image-build--microcks)


## Build Status

Latest release is `1.0.2`.

Current development version is `1.0.3`. It is available as a container image named `quay.io/microcks/microcks-cli:nightly`.

#### Fossa license and security scans

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli?ref=badge_shield&issueType=license)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli?ref=badge_shield&issueType=security)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli?ref=badge_small)

#### Signature, Provenance, SBOM

[![Static Badge](https://img.shields.io/badge/supply_chain-documentation-blue?logo=securityscorecard&label=Supply%20Chain&link=https%3A%2F%2Fmicrocks.io%2Fdocumentation%2Freferences%2Fcontainer-images%23software-supply-chain-security)](https://microcks.io/documentation/references/container-images#software-supply-chain-security)

#### OpenSSF best practices on Microcks core

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/7513/badge)](https://bestpractices.coreinfrastructure.org/projects/7513)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/microcks/microcks/badge)](https://securityscorecards.dev/viewer/?uri=github.com/microcks/microcks)

## Community

* [Documentation](https://microcks.io/documentation/tutorials/getting-started/)
* [Microcks Community](https://github.com/microcks/community) and community meeting
* Join us on [Discord](https://microcks.io/discord-invite/), on [GitHub Discussions](https://github.com/orgs/microcks/discussions) or [CNCF Slack #microcks channel](https://cloud-native.slack.com/archives/C05BYHW1TNJ)

To get involved with our community, please make sure you are familiar with the project's [Code of Conduct](./CODE_OF_CONDUCT.md).

## Installation

Visit the [Release page](https://github.com/microcks/microcks-cli/releases/tag/1.0.2), browse the **Assets** and download the one matching your platform

OR you can use the [Homebrew](https://brew.sh/) package manager on Linux and MacOS that way:

```sh
brew tap microcks/tap
brew install microcks/tap/microcks
```

## Usage instructions

### Usage

```bash
microcks [command] [flags]
```

### Available Commands

| Command      | Description                                              | Documentation                                   |
| ------------ | -------------------------------------------------------- | ----------------------------------------------- |
| `login`      | Log in to a Microcks instance using Keycloak credentials | [`login`](documentation/cmd/login.md)           |
| `logout`     | Log out and remove authentication from a given context   | [`logout`](documentation/cmd/logout.md)         |
| `context`    | Manage CLI contexts (list, use, delete)                  | [`context`](documentation/cmd/context.md)       |
| `start`      | Start a local Microcks instance via Docker/Podman        | [`start`](documentation/cmd/start.md)           |
| `stop`       | Stop a local Microcks instance                           | [`stop`](documentation/cmd/stop.md)             |
| `import`     | Import API spec files from local filesystem              | [`import`](documentation/cmd/import.md)         |
| `import-dir`  | Scan a directory and import API spec files.              | [`import-dir`](documentation/cmd/importDir.md)     |
| `import-url` | Import API spec files directly from a remote URL         | [`import-url`](documentation/cmd/importUrl.md) |
| `test`       | Run tests against a deployed API using selected runner   | [`test`](documentation/cmd/test.md)             |
| `version`    | Print Microcks CLI version                               | [`version`](documentation/cmd/version.md)       |

### Options

| Flag                     | Description                                 |
| ------------------------ | ------------------------------------------- |
| `-h, --help`             | help for microcks command                    |
| `--config`               | Path to Microcks config file                |
| `--microcks-context`     | Name of the Microcks context to use         |
| `--verbose`              | Produce dumps of HTTP exchanges             |
| `--insecure-tls`         | Allow insecure HTTPS connections            |
| `--caCerts`              | Comma-separated paths of CA cert files      |
| `--keycloakClientId`     | Keycloak Realm Service Account ClientId     |
| `--keycloakClientSecret` | Keycloak Realm Service Account ClientSecret |
| `--microcksURL`          | Microcks API URL                            |


### Local contract testing without a server

`microcks test --dry-run` runs a contract test with zero infrastructure: no running Microcks server, no Keycloak credentials, no upfront import. The CLI spins up an ephemeral Microcks container (via [Testcontainers](https://microcks.io/documentation/guides/usage/developing-testcontainers/)), imports your spec, runs the test against your endpoint, prints the result and tears the container down.

```bash
# One-shot: run once, tear down, exit (exit code 0/1 reflects the test result)
microcks test --dry-run \
  --artifact ./openapi.yaml \
  "Pastry API:1.0.0" \
  http://localhost:3000 \
  OPEN_API_SCHEMA

# Watch mode: keep the container alive, re-import + re-run on every save (TDD loop)
microcks test --dry-run --watch \
  --artifact ./openapi.yaml \
  "Pastry API:1.0.0" \
  http://localhost:3000 \
  OPEN_API_SCHEMA
```

| Flag | Default | Description |
| ---- | ------- | ----------- |
| `--dry-run` | `false` | Activate the ephemeral-container path |
| `--artifact` | _(required with `--dry-run`)_ | Local spec file imported as main artifact |
| `--image` | `quay.io/microcks/microcks-uber:latest-native` | Uber image override (must be a `*-native` tag) |
| `--ready-timeout` | `90s` | How long to wait for the container to be ready |
| `--watch` | `false` | Re-run the test when the artifact file changes |

Notes:

- A `localhost`/`127.0.0.1` test endpoint is automatically reachable from inside the container — the CLI exposes the port and rewrites the endpoint for you.
- The container is removed on every exit path, including `Ctrl+C` mid-test.
- Docker is the primary runtime; Podman works through its Docker-compatible socket (`DOCKER_HOST`).

### Building from Source

To build the CLI locally:

```bash
make build-local
```

The resulting binary will be available at:

```bash
/build/dist/microcks
```

You can move it to a location in your $PATH for global usage, for example:

```bash
sudo mv build/dist/microcks /usr/local/bin/microcks
```


### Binary

Binary releases for Linux, MacOS or Windows platform are available on the GitHub [releases page](https://github.com/microcks/microcks-cli/releases). Just download the binary corresponding to your system and put the binary into the `PATH` somewhere ;-)

### Container image

The `microcks-cli` is available as a container image. So that you'd be able to easily use it from a GitLab CI or a [Tekton pipeline](https://github.com/tektoncd/pipeline). The hosting repository is on Quay.io [here](https://quay.io/repository/microcks/microcks-cli).

Below a sample on how to use the image without getting the CLI binary:

```
$ docker run -it quay.io/microcks/microcks-cli:latest microcks test 'Beer Catalog API:0.9' http://beer-catalog-impl-beer-catalog-dev.apps.144.76.24.92.nip.io/api/ POSTMAN --microcksURL=http://microcks.apps.144.76.24.92.nip.io/api/ --keycloakClientId=microcks-serviceaccount --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 --waitFor=8sec  --operationsHeaders='{"globals": [{"name": "x-api-key", "values": "my-values"}], "GET /beer": [{"name": "x-trace-id", "values": "xcvbnsdfghjklm"}]}'
```


## Machine-readable test output

`microcks test` accepts `--output` to control how the result is rendered:

| Value | Output |
| --- | --- |
| `text` (default) | Human-readable summary |
| `json` | The full `TestResult` as JSON |
| `yaml` | The full `TestResult` as YAML |
| `github-actions` | GitHub Actions workflow commands (annotations + log groups + step summary) |

For machine formats (`json`/`yaml`/`github-actions`), progress goes to **stderr**
and only the formatted result is written to **stdout**, so it can be piped or
parsed cleanly:

```bash
microcks test "Pastry API:1.0.0" http://localhost:8080/api OPEN_API_SCHEMA \
  --microcksURL=http://localhost:8585/api --output=json > result.json
```

### GitHub Actions

With `--output=github-actions`, failures surface as `::error::` annotations,
each operation is wrapped in a collapsible `::group::`, and a per-operation table
is appended to the job summary (`$GITHUB_STEP_SUMMARY`). Set
`MICROCKS_ACTIONS_VERBOSE=true` to also emit `::notice::` for passing operations.

```yaml
# .github/workflows/contract-test.yml
name: contract-test
on: [pull_request]
jobs:
  contract-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Microcks contract test
        run: |
          microcks test "Pastry API:1.0.0" "${{ env.API_URL }}" OPEN_API_SCHEMA \
            --microcksURL=${{ secrets.MICROCKS_URL }} \
            --keycloakClientId=${{ secrets.MICROCKS_CLIENT_ID }} \
            --keycloakClientSecret=${{ secrets.MICROCKS_CLIENT_SECRET }} \
            --output=github-actions
```

## Tekton tasks

This repository also contains different [Tekton](https://tekton.dev/) tasks definitions and sample pipelines. You'll find under the `/tekton` folder the resource for current `v1beta1` Tekton API version and the older `v1alpha1` under `tekton/v1alpha1`.
