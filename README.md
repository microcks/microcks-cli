# microcks-cli

Simple CLI for interacting with Microcks server APIs.
It allows to launch tests or import API artifacts with minimal dependencies.

 [![Join the chat on Zulip](https://img.shields.io/badge/chat-on_zulip-pink.svg?color=ff69b4&style=for-the-badge&logo=zulip)](https://microcksio.zulipchat.com/)

## Build Status

Current development version is `0.4.0-SNAPSHOT`. [![Build Status](https://travis-ci.com/microcks/microcks-cli.png?branch=master)](https://travis-ci.com/microcks/microcks-cli)

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
* `<runner>` : Test strategy (one of: `HTTP`, `SOAP`, `SOAP_UI`, `POSTMAN`, `OPEN_API_SCHEMA`, `ASYNC_API_SCHEMA`, `GRPC_PROTOBUF`)")

The flags:
* `--microcksURL` for the Microcks API endpoint,
* `--waitFor` for the time to wait for test to finish (int + one of: milli, sec, min),
* `--keycloakClientId` for the Keycloak Realm Service Account ClientId,
* `--keycloakClientSecret` for the Keycloak Realm Service Account ClientSecret.

Real life example command and execution:
```
$ ./microcks-cli test 'Beer Catalog API:0.9' http://localhost:9090/api/ POSTMAN \
        --microcksURL=http://localhost:8080/api/ \
        --keycloakClientId=microcks-serviceaccount \
        --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 \
        --waitFor=3sec
[...]
MicrocksClient got status for test "5c1781cf6310d94f8169384e" - success: false, inProgress: true
MicrocksTester waiting for 2 seconds before checking again.
MicrocksClient got status for test "5c1781cf6310d94f8169384e" - success: true, inProgress: false
```

#### Advanced options

The `test` command provides additional flags for advanced usages and options:
* `--verbose` allows to dump on standard output all the HTTP requests and responses,
* `--insecure` allows to interact with Microcks and Keycloak instances through HTTPS without checking certificates issuer CA,
* `--caCerts=<path1,path2>` allows to specify additional certificates CRT files to add to trusted roots ones,
* `--secretName='<Secret Name>'` is an optional flag specifying the name of a Secret to use for connecting endpoint
* `--operationsHeaders=<JSON>` allows to override some operations headers for the tests to launch.

Overriden test operations headers is a JSON strings where 1st level keys are operation name (eg. `GET /beer`) or `globals` for header applying to all the operations of the API. Headers are specified as an array of objects defining `key` and `values` properties.

Here's below an example of using some of this flags:

```
./microcks-cli test 'Beer Catalog API:0.9' http://localhost:9090/api/ POSTMAN \                           
        --microcksURL=http://localhost:8080/api/ \
        --keycloakClientId=microcks-serviceaccount \
        --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 \
        --insecure --verbose  --waitFor=3sec \
        --operationsHeaders='{"globals": [{"name": "x-api-key", "values": "my-values"}], "GET /beer": [{"name": "x-trace-id", "values": "xcvbnsdfghjklm"}]}'
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

Real life example command and execution:
```
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

The `microcks-cli` is now available as a container image as version `0.2.0`. So that you'd be able to easily use it from a GitLab CI or a [Tekton pipeline](https://github.com/tektoncd/pipeline). The hosting repository is now on Docker Hub [here](https://hub.docker.com/r/microcks/microcks-cli).

Below a sample on how using the image without getting the CLI binary:

```
$ docker run -it quay.io/microcks/microcks-cli:latest microcks-cli test 'Beer Catalog API:0.9' http://beer-catalog-impl-beer-catalog-dev.apps.144.76.24.92.nip.io/api/ POSTMAN --microcksURL=http://microcks.apps.144.76.24.92.nip.io/api/ --keycloakClientId=microcks-serviceaccount --keycloakClientSecret=7deb71e8-8c80-4376-95ad-00a399ee3ca1 --waitFor=8sec  --operationsHeaders='{"globals": [{"name": "x-api-key", "values": "my-values"}], "GET /beer": [{"name": "x-trace-id", "values": "xcvbnsdfghjklm"}]}'
```

> Previous versions of the `microcks-cli` container image (< `0.3.0`) where hosted on Docker Hub and you can retrieve them using `microcks/microcks:0.2.0` version for example.


## Tekton tasks

This repository also contains different [Tekton](https://tekton.dev/) tasks definition and sample pipelines. You'll find under the `/tekton` folder the resource for current `v1beta1` Tekton API version and the older `v1alpha1` under `tekton/v1alpha1`.
