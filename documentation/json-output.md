# JSON output contracts

How to consume the `microcks` CLI from another program — an editor extension, a CI
job, a script. This page is the reference for the machine-readable surface: which
commands produce structured output, what lands on which stream, the schema of the
dry-run event stream, and how to detect support before depending on any of it.

For the human-facing description of each command, see [documentation/cmd/](cmd/).

## Streams

Commands render human progress on `stdout` in the default `text` mode. As soon as
a machine-readable `--output` value is selected, progress and diagnostics move to
`stderr` and `stdout` carries the payload alone:

```bash
microcks test "Pastry API:1.0.0" http://localhost:8080/api OPEN_API_SCHEMA \
  --microcksURL=http://localhost:8585/api --output=json > result.json
```

`result.json` holds the test result and nothing else — no status lines to strip, no
partial JSON. The same split applies to the dry-run path, so container startup and
teardown messages stay out of the parsed stream.

Diagnostics on stderr are written for people to read and are not a contract. Do not
parse them.

## Supported formats per command

| Command | `text` | `json` | `yaml` | `github-actions` |
| --- | :---: | :---: | :---: | :---: |
| `test` | ✓ | ✓ | ✓ | ✓ |
| `test list` | ✓ | ✓ | | |
| `test get` | ✓ | ✓ | | |
| `service list` | ✓ | ✓ | | |
| `service get` | ✓ | ✓ | | |
| `context` | ✓ | ✓ | | |
| `start` | ✓ | ✓ | | |
| `import` | ✓ | ✓ | | |
| `capabilities` | ✓ | ✓ | | |

`test` is the only command that renders a full `TestResult`, so it is the only one
offering `yaml` and `github-actions`. Everything else exposes `text` or `json`.

`import --watch` cannot be combined with `--output json`: watching is an open-ended
interactive loop with no single result document to emit.

`import-dir` and `import-url` have no `--output` flag yet. Do not expect JSON from
them.

## Payload shapes

| Command | stdout on `--output json` |
| --- | --- |
| `test`, `test get` | One complete `TestResult` object, including `testCaseResults` when present |
| `test list` | Array of test-result summaries |
| `service list` | Array of service summaries — each with `id`, `name`, `version`, `type`, and possibly `operations` |
| `service get` | Object with `service` and, when available, `messagesMap` |
| `context` | Array of `{name, server, current}`; `[]` when no local config exists, which is a valid disconnected state |
| `start` | `{name, server, context, status}` for the started instance |
| `import` | Array of `{file, id, primary, action}`, one entry per imported artifact |
| `capabilities` | `{schemaVersion, cliVersion, capabilities[]}` |

## Dry-run watch events

`microcks test --dry-run --watch --output json` switches to a newline-delimited
event stream instead of rendering results. All three flags are required — with
`--watch` but without `--output json` the run stays in human mode, and without
`--watch` a single rendered result is written instead.

One JSON object per line, so a consumer can read incrementally:

```bash
microcks test --dry-run --watch --output json \
  --artifact ./openapi.yaml "Pastry API:1.0.0" http://localhost:3000 OPEN_API_SCHEMA \
  | jq -c '{type, success: .result.success}'
```

### Event fields

| Field | Present on | Meaning |
| --- | --- | --- |
| `type` | every event | One of the types below |
| `timestamp` | every event | RFC 3339 with nanoseconds, UTC, stamped at emit |
| `endpoint` | `ready` | HTTP endpoint of the ephemeral Microcks server |
| `artifact` | `imported` | Path of the artifact that was imported |
| `service` | `imported` | Service reference the artifact was imported as |
| `testResultId` | `test-result` | Id of the completed test result |
| `result` | `test-result` | The complete `TestResult` object |
| `message` | `error` | Text of the underlying failure |

Fields not relevant to an event type are omitted, not sent as `null`.

### Event types

| Type | Emitted when |
| --- | --- |
| `ready` | The ephemeral container is running and its endpoint is resolved |
| `imported` | The artifact has been imported — once at startup, then after each change |
| `test-result` | A test run completed; carries the full result |
| `waiting` | The CLI is idle, watching for the next artifact change |
| `error` | An import or test run failed |
| `stopped` | The CLI is exiting |

A normal session emits `ready`, `imported`, `test-result`, `waiting`, then a further
`imported` / `test-result` / `waiting` triple per saved change, and `stopped` on exit.

Two properties matter when writing a consumer:

- **`error` is not terminal.** A spec that is invalid while you are still editing it
  is expected. The CLI reports the failure and keeps watching; the next valid save
  produces `imported` and `test-result` as usual. Treat `error` as a status to
  display, not a reason to end your session.
- **`stopped` is always emitted**, including on `Ctrl+C` and on the error paths, so it
  is a reliable end-of-stream marker.

## Detecting support

Do not infer support from a version number. Query the CLI instead:

```bash
microcks capabilities --output json
```

```json
{
  "schemaVersion": "v1",
  "cliVersion": "1.0.3",
  "capabilities": ["service.list.json", "test.dry-run.watch.events.json", "..."]
}
```

Check for the identifier covering the contract you are about to use — for example
`test.dry-run.watch.events.json` before subscribing to the event stream, or
`service.get.json` before parsing service details — and degrade gracefully when it is
absent. The full list is in [capabilities](cmd/capabilities.md).

Capabilities describe the installed binary. They say nothing about optional features
of the Microcks server it talks to.

## Exit codes

A structured payload on `stdout` does not replace the exit code; both are part of the
contract. In particular, exit code `1` means a clean run whose contract test did not
conform, which is the tool working correctly. Codes of `2` and above mean the tool
itself could not complete. See [error handling and exit codes](error-handling.md).
