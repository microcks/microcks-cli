## `microcks capabilities` - List CLI capabilities

Lists stable capability identifiers that integrations can use to detect
whether a Microcks CLI release supports the commands they require.
Capability identifiers describe public workflows and machine-readable
contracts; they are not a copy of every CLI flag.
The command runs locally and does not require a Microcks server or a configured
context.

```sh
microcks capabilities --output json
```

Example output:

```json
{
  "schemaVersion": "v1",
  "cliVersion": "1.0.3",
  "capabilities": [
    "auth.login",
    "auth.login.sso",
    "auth.logout",
    "context.list",
    "context.list.json",
    "context.use",
    "context.use.json",
    "context.delete",
    "context.delete.json",
    "instance.start",
    "instance.start.json",
    "instance.stop",
    "artifact.import.file",
    "artifact.import.file.json",
    "artifact.import.file.watch",
    "artifact.import.directory",
    "artifact.import.url",
    "service.list.json",
    "service.get.json",
    "test.run",
    "test.run.output.json",
    "test.run.output.yaml",
    "test.run.output.github-actions",
    "test.dry-run",
    "test.dry-run.watch",
    "test.dry-run.watch.events.json",
    "test.list.json",
    "test.get.json"
  ]
}
```

### Capability identifiers

| Capability | Available workflow or contract |
| --- | --- |
| `auth.login` | Log in with username and password |
| `auth.login.sso` | Log in through the browser-based SSO flow |
| `auth.logout` | Remove authentication for a context |
| `context.list` | List configured contexts as text |
| `context.list.json` | List configured contexts using the stable JSON contract |
| `context.use` | Select the current context |
| `context.use.json` | Select a context and return the selection as JSON |
| `context.delete` | Delete a configured context |
| `context.delete.json` | Delete a context and return the result as JSON |
| `instance.start` | Start a local Microcks container |
| `instance.start.json` | Start an instance and return its server/context as JSON |
| `instance.stop` | Stop a local Microcks container |
| `artifact.import.file` | Import one or more local artifact files |
| `artifact.import.file.json` | Import local artifacts and return their identifiers as JSON |
| `artifact.import.file.watch` | Re-import local artifacts when files change |
| `artifact.import.directory` | Import artifacts discovered in a directory |
| `artifact.import.url` | Import artifacts from remote URLs |
| `service.list.json` | List services using the stable JSON contract |
| `service.get.json` | Retrieve service details using the stable JSON contract |
| `test.run` | Run a test against a target endpoint |
| `test.run.output.json` | Render a test result as JSON |
| `test.run.output.yaml` | Render a test result as YAML |
| `test.run.output.github-actions` | Render annotations for GitHub Actions |
| `test.dry-run` | Run a test with an ephemeral Microcks container |
| `test.dry-run.watch` | Re-run an ephemeral test when its artifact changes |
| `test.dry-run.watch.events.json` | Emit NDJSON lifecycle and result events while watching |
| `test.list.json` | List test results using the stable JSON contract |
| `test.get.json` | Retrieve a test result using the stable JSON contract |

`test.dry-run.watch` describes the interactive text workflow.
`test.dry-run.watch.events.json` guarantees newline-delimited `ready`,
`imported`, `test-result`, `waiting`, `error`, and `stopped` events.

Capabilities describe the behavior of the installed CLI binary. They do not
report optional features enabled by a particular Microcks server.

### Options

| Flag | Description |
| --- | --- |
| `--output` | Output format: `text` or `json` (default: `text`) |
