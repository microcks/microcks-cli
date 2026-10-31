# Triage commands

Issues and pull requests in this repository can be triaged with Prow-style slash commands, similar to Kubernetes. Write the command at the start of a line in a comment. You can put several commands in one comment, one per line. The bot reacts with 👍 when the commands are applied, and replies if one could not be applied.

## Triage flow

1. Every new issue gets the `needs-triage` label.
2. A maintainer or contributor checks it and comments `/triage accepted`, or `/triage needs-information` if more details are needed from the reporter.
3. Once an issue has `triage/accepted`, contributors can pick it up with `/assign` and open a pull request that references it.

Please wait for `triage/accepted` before starting work. It avoids spending time on a fix or feature that the maintainers would not merge.

## Commands for everyone

| Command | Effect |
| --- | --- |
| `/assign` | Assigns yourself. On issues, only once the issue has `triage/accepted`. |
| `/unassign` | Unassigns yourself. |
| `/close`, `/reopen` | Closes or reopens an issue or pull request you opened. You can only reopen it if you closed it yourself; if a maintainer closed it, ask them. |

## Commands for contributors and maintainers

Contributors are people who already have commits in this repository. They can help triage issues and pull requests opened by others, but not their own.

| Command | Effect |
| --- | --- |
| `/triage accepted` | Adds `triage/accepted` and removes `needs-triage`. |
| `/triage needs-information` | Adds `triage/needs-information` and removes `needs-triage`. |
| `/kind bug\|feature\|enhancement\|task\|question` | Adds the matching `kind/*` label. |
| `/area cli\|connectors\|build\|docs\|packaging` | Adds the matching `component/*` label. |
| `/platform windows\|linux\|macos` | Adds the matching `platform/*` label. |
| `/good-first-issue` | Adds `good first issue`. On issues, needs `triage/accepted` first. |
| `/help` | Adds `help wanted`. On issues, needs `triage/accepted` first. |

## Commands for maintainers

Maintainers are people with write access to this repository. They can use every command above, including on their own issues, and the following ones to correct or manage triage.

| Command | Effect |
| --- | --- |
| `/remove-triage` | Removes the `triage/*` labels and adds `needs-triage` back. |
| `/remove-kind`, `/remove-area`, `/remove-platform` | Removes the matching label, for example `/remove-kind bug`. |
| `/remove-good-first-issue`, `/remove-help` | Removes `good first issue` or `help wanted`. |
| `/keep-open` | Adds `keep-open`, so the stale bot skips the item. `/remove-keep-open` removes it. |
| `/assign @user`, `/unassign @user` | Assigns or unassigns someone else. GitHub only lets you assign people who have commented on the issue or have access to the repository. |
| `/close`, `/reopen` | Closes or reopens any issue or pull request. `/reopen` also removes the `status/*` label. |
| `/close duplicate\|wontfix\|invalid` | Closes with a reason and adds the matching `status/*` label. On issues it also sets GitHub's close reason (duplicate or not planned). |
| `/priority critical-urgent\|important-soon\|important-longterm\|backlog` | Sets the `priority/*` label, replacing any other priority. `/remove-priority` removes it. |
| `/milestone <title>` | Sets an open GitHub milestone, for example `/milestone 1.0.3`. `/milestone clear` or `/remove-milestone` removes it. |

## Pull request commands

These commands only work in pull request comments. `lgtm`, `approved` and `do-not-merge/hold` are review signals for maintainers: merging still needs an approving GitHub review from a code owner.

| Command | Who | Effect |
| --- | --- | --- |
| `/lgtm` | Contributors and maintainers, not on their own pull request | Adds `lgtm`. |
| `/approve` | Maintainers, not on their own pull request | Adds `approved`. |
| `/remove-lgtm` (`/lgtm cancel`), `/remove-approve` (`/approve cancel`) | Maintainers | Removes `lgtm` or `approved`. |
| `/hold` | Contributors, maintainers and the author | Adds `do-not-merge/hold`. The `hold` check fails while the label is set. |
| `/unhold` (`/hold cancel`) | Maintainers | Removes `do-not-merge/hold`. |
| `/retest` | Maintainers and the author | Re-runs the failed checks on the latest commit. |
| `/cc @user`, `/uncc @user` | Contributors, maintainers and the author | Requests or removes a review from someone. |

Commands inside code blocks, quoted lines (`> /kind bug`) and HTML comments are ignored, and unknown commands are ignored silently. If a command can't be applied, the bot reacts with 😕 and explains why, once per person and issue.

## Example

```text
/triage accepted
/kind bug
/area connectors
/good-first-issue
```
