# Architecture

Microcks CLI is a Go-based command-line tool designed for interacting with Microcks server APIs, launching tests, and importing artifacts.

## Core Components
- **CLI Commands:** Built using a modern Go CLI framework (e.g., Cobra), providing subcommands like `start`, `stop`, `import`, `test`, etc.
- **Microcks API Client:** Communicates with the Microcks backend REST APIs to perform operations.
- **Keycloak Integration:** Handles authentication and token management for secure interactions with the backend.
- **Local Ephemeral Runner:** Employs Testcontainers to spin up ephemeral Microcks instances for local dry-run contract testing (`microcks test --dry-run`).

## Structure
- `cmd/`: Contains the entry points and CLI command definitions.
- `pkg/`: Core packages and business logic.
- `documentation/`: Additional documentation.

## Deployment
Available as a standalone binary for Linux, macOS, and Windows, and as a container image (`quay.io/microcks/microcks-cli`) for CI/CD integration.
