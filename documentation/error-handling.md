# Error handling & exit codes

How the `microcks` CLI reports failures — for users scripting it in CI, and for
contributors adding code.

## Exit codes

The CLI returns a distinct exit code per outcome so a pipeline can branch on *why*
a run failed — for example, retry a flaky infrastructure problem but fail the
build on a genuinely broken contract.

| Code | Meaning |
| ---- | ------- |
| 0  | Success — the operation completed, or the contract test conformed |
| 1  | Contract test failed — a clean run whose result did **not** conform |
| 2  | Usage — bad arguments or flags |
| 11 | Connection — could not reach the Microcks or Keycloak endpoint |
| 12 | API — a server rejected the request or returned an unusable response |
| 13 | Not found — a requested remote resource does not exist |
| 14 | Environment — a local precondition failed (container runtime, image, readiness) |
| 20 | Generic — an unclassified failure |

Note that `1` means "the tool ran fine and the API violated its contract" — not
"the tool errored". This mirrors `kubectl diff` (0 = no diff, 1 = diff found,
>1 = errored) and pytest (1 = tests failed). It lets CI tell "your API is broken"
apart from "my pipeline is broken".

## Terminology

- **Failure Kind** — the category of *why* an operation could not complete
  (connection, API, not-found, usage, environment). The library's vocabulary for
  failure; it never mentions exit codes.
- **Exit Code** — the integer status the process returns, derived from a Failure
  Kind. A CLI-only concept.
- **Conformance / Contract Test Result** — the outcome of a completed test run:
  the API either conforms or does not. A *does-not-conform* result is the tool
  working correctly, **not** a failure.
- **Environment failure** — a local precondition wasn't met (the container runtime
  is down/unreachable, an image can't be pulled, the ephemeral server wasn't ready
  in time). Distinct from a *connection* failure, which is about reaching the
  Microcks server.

## How it works

- **The library (`pkg/*`) never exits or panics on a runtime error.** It returns
  errors classified by Failure Kind via `errors.Wrap(kind, err)`. A consumer that
  embeds the client (e.g. an editor extension) reads the kind with
  `errors.KindOf(err)` — exit codes are not a library concern.
- **Commands use Cobra `RunE`** and return kind-tagged errors instead of exiting.
- **`cmd.Handle`, called only by `main()`, is the single exit point.** It prints
  the error to stderr and maps Failure Kind → exit code (`cmd/exit.go`).
- **A non-conforming test is a Result, not a failure.** It travels as the silent
  `errors.ErrTestFailed` sentinel: the command already rendered the result, so
  `Handle` exits `1` and prints nothing further.

## Adding code — the rule

- In `pkg/*`: `return errors.Wrap(errors.KindConnection, err)` (or the fitting
  kind). Never `os.Exit`, `panic`, or `log.Fatal`.
- In a command: return the error from `RunE`; don't exit yourself.
- Only `main()` / `cmd.Handle` exits the process.

Classify at the point with the most context: transport failures (`httpClient.Do`)
are `KindConnection`; non-2xx responses are `KindAPI` (a `404` for a missing
resource is `KindNotFound`); a missing local input file is `KindUsage`; a failed
container runtime is `KindEnvironment`.
