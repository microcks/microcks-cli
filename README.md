# Microcks CLI

Simple CLI for interacting with Microcks server APIs.
It allows to launch tests or import API artifacts with minimal dependencies.

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/microcks/microcks-cli/build-verify.yml?logo=github&style=for-the-badge)](https://github.com/microcks/microcks-cli/actions)
[![Container](https://img.shields.io/badge/dynamic/json?color=blue&logo=docker&style=for-the-badge&label=Quay.io&query=tags[0].name&url=https://quay.io/api/v1/repository/microcks/microcks-cli/tag/?limit=10&page=1&onlyActiveTags=true)](https://quay.io/repository/microcks/microcks-cli?tab=tags)
[![License](https://img.shields.io/github/license/microcks/microcks-cli?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![Project Chat](https://img.shields.io/badge/discord-microcks-pink.svg?color=7289da&style=for-the-badge&logo=discord)](https://microcks.io/discord-invite/)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/microcks-cli-image&style=for-the-badge)](https://artifacthub.io/packages/search?repo=microcks-cli-image)
[![CNCF Landscape](https://img.shields.io/badge/CNCF%20Landscape-5699C6?style=for-the-badge&logo=cncf)](https://landscape.cncf.io/?item=app-definition-and-development--application-definition-image-build--microcks)


## Build Status

Latest release is `0.5.6`.

Current development version is `0.5.7`.

It is available as a container image named `quay.io/microcks/microcks-cli:nightly`.

#### Fossa license and security scans

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli?ref=badge_shield&issueType=license)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli?ref=badge_shield&issueType=security)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmicrocks%2Fmicrocks-cli?ref=badge_small)

#### OpenSSF best practices on Microcks core

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/7513/badge)](https://bestpractices.coreinfrastructure.org/projects/7513)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/microcks/microcks/badge)](https://securityscorecards.dev/viewer/?uri=github.com/microcks/microcks)

## Community

* [Documentation](https://microcks.io/documentation/tutorials/getting-started/)
* [Microcks Community](https://github.com/microcks/community) and community meeting
* Join us on [Discord](https://microcks.io/discord-invite/), on [GitHub Discussions](https://github.com/orgs/microcks/discussions) or [CNCF Slack #microcks channel](https://cloud-native.slack.com/archives/C05BYHW1TNJ)

To get involved with our community, please make sure you are familiar with the project's [Code of Conduct](./CODE_OF_CONDUCT.md).

## Usage instructions

Usage is simply `microcks-cli [command]`

where `[command]` can be one of the following:

* `version` to check this CLI version,
* `help` to display usage informations,
* `test` to launch new test on Microcks server.
* `import` to import API artifacts on Microcks server.

### Test command

The `test` command has a bunch of arguments and flags so that you can use it that way:

```
microcks-cli test <apiName:apiVersion> <testEndpoint> <runner>
        --microcksURL=<> --waitFor=5sec
        --keycloakClientId=<> --keycloakClientSecret=<>
```

The arguments:

* `<apiName:apiVersion>` : Service to test reference. Example: `'Beer Catalog API:0.9'`
* `<testEndpoint>` : URL where is deployed implementation to test
* `<runner>` : Test strategy (one of: `HTTP`, `SOAP`, `SOAP_UI`, `POSTMAN`, `OPEN_API_SCHEMA`, `ASYNC_API_SCHEMA`, `GRPC_PROTOBUF`, `GRAPHQL_SCHEMA`)

The flags:

* `--microcksURL` for the Microcks API endpoint,
* `--waitFor` for the time to wait for test to finish (int + one of: milli, sec, min),
* `--keycloakClientId` for the Keycloak Realm Service Account ClientId,
* `--keycloakClientSecret` for the Keycloak Realm Service Account ClientSecret.

> Since `0.5.2` release, `microcks-cli` also works in unauthenticated mode, without Keycloak being deployed and configured with your Microcks instance. However, you still have to specify the Keycloak flags on the command even if you put random values there. For eg. `--keycloakClientId=foo --keycloakClientSecret=bar`.

Real life example command and execution:

```sh
$ ./microcks-cli test 'Beer Catalog API:0.9' http://localhost:9090/api/ POSTMAN \
        --microcksURL=http://localhost:8080/api/ \
        --keycloakClientId=microcks-serviceaccount \
        --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 \
        --waitFor=3sec
[...]
MicrocksClient got status for test "5c1781cf6310d94f8169384e" - success: false, inProgress: true
MicrocksTester waiting for 2 seconds before checking again.
MicrocksClient got status for test "5c1781cf6310d94f8169384e" - success: true, inProgress: false
Full TestResult details are available here: http://localhost:8080/#/tests/5c1781cf6310d94f8169384e 
```

#### Advanced options

The `test` command provides additional flags for advanced usages and options:

* `--verbose` allows to dump on standard output all the HTTP requests and responses,
* `--insecure` allows to interact with Microcks and Keycloak instances through HTTPS without checking certificates issuer CA,
* `--caCerts=<path1,path2>` allows to specify additional certificates CRT files to add to trusted roots ones,
* `--secretName='<Secret Name>'` is an optional flag specifying the name of a Secret to use for connecting endpoint,
* `--filteredOperations=<JSON>` allows to filter a list of operations to launch a test for,
* `--operationsHeaders=<JSON>` allows to override some operations headers for the tests to launch,
* `--oAuth2Context=<JSON>` allows specification of an OAuth2 grant flow to execute before launching the test (starts with Microcks version `1.8.0`).

Overriden test operations headers is a JSON strings where 1st level keys are operation name (eg. `GET /beer`) or `globals` for header applying to all the operations of the API. Headers are specified as an array of objects defining `key` and `values` properties.

Here's below an example of using some of this flags:

```sh
$ ./microcks-cli test 'Beer Catalog API:0.9' http://localhost:9090/api/ OPEN_API_SCHEMA \                           
        --microcksURL=http://localhost:8080/api/ \
        --keycloakClientId=microcks-serviceaccount \
        --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 \
        --insecure --verbose  --waitFor=3sec \
        --filteredOperations='["GET /beer", "GET /beer/{name}"]' \
        --operationsHeaders='{"globals": [{"name": "x-api-key", "values": "my-values"}], "GET /beer": [{"name": "x-trace-id", "values": "xcvbnsdfghjklm"}]}' \
        --oAuth2Context='{"clientId": "microcks-test", "clientSecret": "ab54d329-e435-41ae-a900-ec6b3fe15c54", "tokenUri": "https://idp.acme.org/realms/my-app/protocol/openid-connect/token", "grantType": "CLIENT_CREDENTIALS"}'

MicrocksClient got status for test "64c25f7ddec62569f9a0ed95" - success: true, inProgress: false 
Full TestResult details are available here: http://localhost:8080/#/tests/64c25f7ddec62569f9a0ed95 
```

### Import command

The `import` command has one argument and common flags with `test` command. You can use it that way:

```
microcks-cli import <specificationFile1[:primary],specificationFile2[:primary]>
	--microcksURL=<>
	--keycloakClientId=<> --keycloakClientSecret=<>
```

The arguments:

* `<specificationFile1[:primary],specificationFile2[:primary]>` : Comma separated list of API specs to import with flag telling if it's a primary artifact. Example: `'specs/my-openapi.yaml:true,specs/my-postmancollection.json:false'`

The flags:

* `--microcksURL` for the Microcks API endpoint,
* `--keycloakClientId` for the Keycloak Realm Service Account ClientId,
* `--keycloakClientSecret` for the Keycloak Realm Service Account ClientSecret.

> Since `0.5.2` release, `microcks-cli` also works in unauthenticated mode, without Keycloak being deployed and configured with your Microcks instance. However, you still have to specify the Keycloak flags on the command even if you put random values there. For eg. `--keycloakClientId=foo --keycloakClientSecret=bar`.

Real life example command and execution:

```sh
$ ./microcks-cli import 'samples/weather-forecast-openapi.yml:true,samples/weather-forecast-postman.json:false' \
        --microcksURL=http://localhost:8080/api/ \
        --keycloakClientId=microcks-serviceaccount \
        --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1
Microcks has discovered 'WeatherForecast API:1.1.0'
Microcks has discovered 'WeatherForecast API:1.1.0'
```

#### Advanced options

The `import` command provides additional flags for advanced usages and options:

* `--verbose` allows to dump on standard output all the HTTP requests and responses,
* `--insecure` allows to interact with Microcks and Keycloak instances through HTTPS without checking certificates issuer CA,
* `--caCerts=<path1,path2>` allows to specify additional certificates CRT files to add to trusted roots ones,


## Installation

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
