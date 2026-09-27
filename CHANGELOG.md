# CHANGELOG

## Introduction

This file provides a high-level summary of the changes and updates made to the Microcks project.

## Unreleased

### Bug Fixes

- **Fix: HTTP client timeout was not set, causing CLI to hang indefinitely**
  - The internal `http.Client` instances in `NewClient`, `NewMicrocksClient` (microcks_client.go) and `NewKeycloakClient` (keycloak_client.go) were initialized without an explicit `Timeout` field. In Go, an `http.Client` without a configured timeout defaults to waiting infinitely. If a network connection silently dropped or the Microcks/Keycloak server became unresponsive, the CLI would freeze forever in the terminal with no output and had to be manually killed by the user.
  - **Fix:** All three client constructors now set `Timeout: 30 * time.Second` so the CLI fails fast with a helpful `context deadline exceeded (Client.Timeout exceeded while awaiting headers)` error if the server does not respond within 30 seconds.
  - **Files changed:** `pkg/connectors/microcks_client.go`, `pkg/connectors/keycloak_client.go`
  - **Tests added:** `TestMicrocksClientTimeout`, `TestKeycloakClientTimeout`, `TestNewClientTimeout` in `pkg/connectors/microcks_client_test.go`

## Releases

For a comprehensive list of changes, please visit the [official release page](https://github.com/microcks/microcks/releases) 

## For More Information

For more detailed release notes, please refer to the [Microcks blog page](https://microcks.io/blog/).

## For specific changes on this repo

Refer to the [activity view](https://docs.github.com/en/repositories/viewing-activity-and-data-for-your-repository/using-the-activity-view-to-see-changes-to-a-repository) of the current repository for detailed information on all recent changes.
