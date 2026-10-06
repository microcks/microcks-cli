# Design

This document outlines the core design principles and patterns used in the Microcks CLI codebase.

## Design Principles
- **Simplicity:** The CLI should be easy to use with minimal dependencies, providing a straightforward developer experience.
- **CI/CD Readiness:** Designed to be easily integrated into pipelines (GitHub Actions, Jenkins, Tekton) with machine-readable outputs (JSON, YAML, GitHub Actions).
- **Graceful Error Handling:** Code under `pkg/` and `cmd/` must return errors using the internal `errors` package (e.g., `errors.Wrap(errors.KindConnection, err)`), never exiting or panicking arbitrarily.

## Coding Standards
- **Go Standard:** Strict adherence to idiomatic Go formatting (`go fmt`).
- **Error Types:** Follow the guidelines in `documentation/error-handling.md` and use the specific exit codes (0-2 for standard, 11-20 for Microcks-specific).

## Best Practices
- **Modularity:** Keep command handlers thin and delegate business logic to packages in `pkg/`.
- **Testability:** Provide dry-run options and use ephemeral test environments (Testcontainers) when necessary.
