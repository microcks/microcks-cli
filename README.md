# microcks-cli
Simple CLI for interacting with Microcks server APIs.
It allows to launch tests with minimal dependencies

## Build Status

Current development version is `0.1.1-SNAPSHOT`. [![Build Status](https://travis-ci.org/microcks/microcks-cli.png?branch=master)](https://travis-ci.org/microcks/microcks-cli)

## Usage instrcutions

Usage is simply `microcks-cli [command]`

where `[comamnd]` can be one of the following:
* `version` to check this CLI version,
* `help` to display usage informations,
* `test` to launch new test on Microcks server.

The main `test` command has abunch of arguments and flags so that you can use it that way:
```
microcks-cli test <apiName:apiVersion> <testEndpoint> <runner>
	--microcksURL=<> --waitFor=5sec
	--keycloakURL=<> --keycloakUsername=<> --keycloakPassword=<>
```

The arguments:
* `<apiName:apiVersion>` : Service to test reference. Exemple: `'Beer Catalog API:0.9'`
* `<testEndpoint>` : URL where is deployed implementation to test
* `<runner>` : Test strategy (one of: `HTTP`, `SOAP`, `SOAP_UI`, `POSTMAN`, `OPEN_API_SCHEMA`)")

The flags:
* `--microcksURL` for the Microcks API endpoint,
* `--waitFor` for the time to wait for test to finish (int + one of: milli, sec, min),
* `--keycloakURL` for the Keycloak Realm API endpoint for Microcks,
* `--keycloakUsername` for the Keycloak Realm ServiceAccount,
* `--keycloakPassword` for the Keycloak Realm Account Password.

Real life exemple command and execution:
```
$ ./microcks-cli test 'Beer Catalog API:0.9' http://localhost:9090/api/ POSTMAN \
        --microcksURL=http://localhost:8080/api/ \
        --keycloakURL=http://localhost:8180/auth/realms/microcks/ \
        --keycloakUsername=microcks-serviceaccount \
        --keycloakPassword=7deb71e8-8c80-4376-95ad-00a399ee3ca1 \
        --waitFor=3sec
[...]
MicrocksClient got status for test "5c1781cf6310d94f8169384e" - success: false, inProgress: true
MicrocksTester waiting for 2 seconds before checking again.
MicrocksClient got status for test "5c1781cf6310d94f8169384e" - success: true, inProgress: false
```