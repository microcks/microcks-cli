# Microcks CLI

Simple CLI for interacting with Microcks server APIs.
It allows to launch tests or import API artifacts with minimal dependencies.

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/microcks/microcks-cli/build-verify.yml?logo=github&style=for-the-badge)](https://github.com/microcks/microcks-cli/actions)
[![Container](https://img.shields.io/badge/dynamic/json?color=blue&logo=docker&style=for-the-badge&label=Quay.io&query=tags[0].name&url=https://quay.io/api/v1/repository/microcks/microcks-cli/tag/?limit=10&page=1&onlyActiveTags=true)](https://quay.io/repository/microcks/microcks-cli?tab=tags)
[![License](https://img.shields.io/github/license/microcks/microcks-cli?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![Project Chat](https://img.shields.io/badge/discord-microcks-pink.svg?color=7289da&style=for-the-badge&logo=discord)](https://microcks.io/discord-invite/)


## Build Status

Latest release is `0.5.5`

Current development version is `0.5.6`. [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/microcks/microcks-cli/build-verify.yml?logo=github&style=for-the-badge)](https://github.com/microcks/microcks-cli/actions).

It is available as a container image named `quay.io/microcks/microcks-cli:nightly`.

## CLI Command Overview

### Usage
```bash
microcks [command] [flags]
```

### Availabel Commands
| Command      | Description                                              | Documentation                                   |
| ------------ | -------------------------------------------------------- | ----------------------------------------------- |
| `login`      | Log in to a Microcks instance using Keycloak credentials | [`login`](documentation/cmd/login.md)           |
| `logout`     | Log out and remove authentication from a given context   | [`logout`](documentation/cmd/logout.md)         |
| `context`    | Manage CLI contexts (list, use, delete)                  | [`context`](documentation/cmd/context.md)       |
| `start`      | Start a local Microcks instance via Docker/Podman        | [`start`](documentation/cmd/start.md)           |
| `stop`       | Stop a local Microcks instance                           | [`stop`](documentation/cmd/stop.md)             |
| `import`     | Import API spec files from local filesystem              | [`import`](documentation/cmd/import.md)         |
| `import-url` | Import API spec files directly from a remote URL         | [`import-url`](documentation/cmd/import-url.md) |
| `test`       | Run tests against a deployed API using selected runner   | [`test`](documentation/cmd/test.md)             |
| `version`    | Print Microcks CLI version                               | [`version`](documentation/cmd/version.md)       |

### Options

| Flag                     | Description                                 |
| ------------------------ | ------------------------------------------- |
| `-h, --help`             | help for microck command                    |
| `--config`               | Path to Microcks config file                |
| `--microcks-context`     | Name of the Microcks context to use         |
| `--verbose`              | Produce dumps of HTTP exchanges             |
| `--insecure-tls`         | Allow insecure HTTPS connections            |
| `--caCerts`              | Comma-separated paths of CA cert files      |
| `--keycloakClientId`     | Keycloak Realm Service Account ClientId     |
| `--keycloakClientSecret` | Keycloak Realm Service Account ClientSecret |
| `--microcksURL`          | Microcks API URL                            |


## Installation

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

Below a sample on how using the image without getting the CLI binary:

```
$ docker run -it quay.io/microcks/microcks-cli:latest microcks-cli test 'Beer Catalog API:0.9' http://beer-catalog-impl-beer-catalog-dev.apps.144.76.24.92.nip.io/api/ POSTMAN --microcksURL=http://microcks.apps.144.76.24.92.nip.io/api/ --keycloakClientId=microcks-serviceaccount --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 --waitFor=8sec  --operationsHeaders='{"globals": [{"name": "x-api-key", "values": "my-values"}], "GET /beer": [{"name": "x-trace-id", "values": "xcvbnsdfghjklm"}]}'
```


## Tekton tasks

This repository also contains different [Tekton](https://tekton.dev/) tasks definition and sample pipelines. You'll find under the `/tekton` folder the resource for current `v1beta1` Tekton API version and the older `v1alpha1` under `tekton/v1alpha1`.
