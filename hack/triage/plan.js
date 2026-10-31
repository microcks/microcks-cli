/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// @ts-check

// Decides what a set of commands does. All permission rules live here; no
// GitHub API calls are made.

const NEEDS_TRIAGE = 'needs-triage';
const TRIAGE_ACCEPTED = 'triage/accepted';
// Pull request review signals. They are labels only: merging still needs an
// approving GitHub review from a code owner, and a maintainer merges by hand.
// TODO: merge automatically once a PR has lgtm and approved, no hold and green
// checks, like Prow's Tide. The master ruleset requires a code owner review,
// so the bot would need a bypass there (or a GitHub App) first; that needs
// maintainer agreement.
const LGTM = 'lgtm';
const APPROVED = 'approved';
const HOLD = 'do-not-merge/hold';
const CANCEL = { lgtm: 'remove-lgtm', approve: 'remove-approve', hold: 'unhold' };

// Command argument -> repository label. Only these labels can be set.
const TRIAGE = {
  accepted: TRIAGE_ACCEPTED,
  'needs-information': 'triage/needs-information',
};
const KIND = {
  bug: 'kind/bug',
  feature: 'kind/feature',
  enhancement: 'kind/enhancement',
  task: 'kind/task',
  question: 'kind/question',
};
const AREA = {
  cli: 'component/cli',
  connectors: 'component/connectors',
  build: 'component/build',
  docs: 'component/documentation',
  packaging: 'component/packaging',
};
const PLATFORM = {
  windows: 'platform/windows',
  linux: 'platform/linux',
  macos: 'platform/macos',
};
const FLAGS = {
  help: 'help wanted',
  'good-first-issue': 'good first issue',
  'keep-open': 'keep-open',
};
const PRIORITY = {
  'critical-urgent': 'priority/critical-urgent',
  'important-soon': 'priority/important-soon',
  'important-longterm': 'priority/important-longterm',
  backlog: 'priority/backlog',
};
const CLOSE_REASON = {
  duplicate: { label: 'status/duplicate', stateReason: 'duplicate' },
  wontfix: { label: 'status/wontfix', stateReason: 'not_planned' },
  invalid: { label: 'status/invalid', stateReason: 'not_planned' },
};

/**
 * @typedef {object} Context
 * @property {string} commenter login of the person who wrote the comment
 * @property {string | undefined} author login of the issue or pull request author
 * @property {boolean} isMaintainer commenter has write access or more
 * @property {boolean} isContributor commenter already has commits in the repository
 * @property {boolean} isPullRequest the comment is on a pull request
 * @property {string[]} labels current labels of the issue or pull request
 * @property {string} [closedBy] login of the person who last closed it, if any
 */

/**
 * @typedef {object} Plan
 * @property {string[]} addLabels
 * @property {string[]} removeLabels
 * @property {string[]} assign
 * @property {string[]} unassign
 * @property {string[]} reviewers review requests to add
 * @property {string[]} unreviewers review requests to remove
 * @property {boolean} retest re-run the failed checks of the pull request
 * @property {'open' | 'closed' | null} state
 * @property {'duplicate' | 'not_planned' | null} stateReason why an issue is closed
 * @property {{ title: string | null } | null} milestone milestone to set (title null clears it), or null for no change
 * @property {string[]} errors explanations for commands that were not applied
 */

/**
 * Turns parsed commands into the changes to apply. Pure function so it can be
 * unit tested without the GitHub API.
 *
 * Maintainers can use every command. Contributors (people with commits in the
 * repository) can add triage and label values, but not on their own issues;
 * removing labels stays with maintainers so they can correct a triage.
 *
 * @param {import('./parse').Command[]} commands
 * @param {Context} ctx
 * @returns {Plan}
 */
function plan(commands, ctx) {
  const labels = new Set(ctx.labels);
  /** @type {Plan} */
  const out = {
    addLabels: [], removeLabels: [], assign: [], unassign: [], reviewers: [], unreviewers: [], retest: false, state: null, stateReason: null, milestone: null, errors: [],
  };

  const add = (label) => {
    if (!labels.has(label)) {
      labels.add(label);
      out.addLabels.push(label);
    }
    out.removeLabels = out.removeLabels.filter((l) => l !== label);
  };
  const remove = (label) => {
    if (labels.has(label)) {
      labels.delete(label);
      out.removeLabels.push(label);
    }
    out.addLabels = out.addLabels.filter((l) => l !== label);
  };
  const isAuthor = Boolean(ctx.author) && ctx.author.toLowerCase() === ctx.commenter.toLowerCase();
  const needMaintainer = (cmd) => {
    if (!ctx.isMaintainer) {
      out.errors.push(`\`/${cmd}\` can only be used by maintainers.`);
      return false;
    }
    return true;
  };
  const needTriager = (cmd) => {
    if (ctx.isMaintainer) return true;
    if (!ctx.isContributor) {
      out.errors.push(`\`/${cmd}\` can be used by maintainers and by people who have contributed to this repository.`);
      return false;
    }
    if (isAuthor) {
      out.errors.push(`\`/${cmd}\` can't be used on your own ${ctx.isPullRequest ? 'pull request' : 'issue'}. Another contributor or a maintainer will triage it.`);
      return false;
    }
    return true;
  };
  const needPullRequest = (cmd) => {
    if (!ctx.isPullRequest) {
      out.errors.push(`\`/${cmd}\` can only be used on pull requests.`);
      return false;
    }
    return true;
  };
  // Maintainers, contributors and the pull request author.
  const needParticipant = (cmd) => {
    if (ctx.isMaintainer || ctx.isContributor || isAuthor) return true;
    out.errors.push(`\`/${cmd}\` can be used by maintainers, by people who have contributed to this repository, and by the author.`);
    return false;
  };
  const mapped = (cmd, table, arg) => {
    const label = table[String(arg || '').toLowerCase()];
    if (!label) {
      out.errors.push(`\`/${cmd}\` expects one of: ${Object.keys(table).map((k) => `\`${k}\``).join(', ')}.`);
    }
    return label;
  };
  const labelCommand = (name, table, args) => {
    const removing = name.startsWith('remove-');
    const base = removing ? name.slice('remove-'.length) : name;
    if (!(removing ? needMaintainer(name) : needTriager(name))) return;
    if (args.length === 0) {
      out.errors.push(`\`/${name}\` expects one of: ${Object.keys(table).map((k) => `\`${k}\``).join(', ')}.`);
      return;
    }
    for (const arg of args) {
      const label = mapped(base, table, arg);
      if (label) (removing ? remove : add)(label);
    }
  };

  for (const command of commands) {
    const { args } = command;
    const name = CANCEL[command.name] && String(args[0]).toLowerCase() === 'cancel' ? CANCEL[command.name] : command.name;
    switch (name) {
      case 'triage': {
        if (!needTriager(name)) break;
        const label = mapped(name, TRIAGE, args[0]);
        if (!label) break;
        for (const l of Object.values(TRIAGE)) if (l !== label) remove(l);
        remove(NEEDS_TRIAGE);
        add(label);
        break;
      }
      case 'remove-triage':
        if (!needMaintainer(name)) break;
        for (const l of Object.values(TRIAGE)) remove(l);
        add(NEEDS_TRIAGE);
        break;
      case 'kind':
      case 'remove-kind':
        labelCommand(name, KIND, args);
        break;
      case 'area':
      case 'remove-area':
        labelCommand(name, AREA, args);
        break;
      case 'platform':
      case 'remove-platform':
        labelCommand(name, PLATFORM, args);
        break;
      case 'help':
      case 'good-first-issue':
      case 'keep-open': {
        // keep-open exempts an item from the stale bot, so it stays with maintainers.
        if (!(name === 'keep-open' ? needMaintainer(name) : needTriager(name))) break;
        if (name !== 'keep-open' && !ctx.isPullRequest && !labels.has(TRIAGE_ACCEPTED)) {
          out.errors.push(`\`/${name}\` needs the issue to be accepted first: use \`/triage accepted\`.`);
          break;
        }
        add(FLAGS[name]);
        break;
      }
      case 'remove-help':
      case 'remove-good-first-issue':
      case 'remove-keep-open':
        if (!needMaintainer(name)) break;
        remove(FLAGS[name.slice('remove-'.length)]);
        break;
      case 'assign': {
        const targets = args.length ? args : [ctx.commenter];
        const others = targets.filter((u) => u.toLowerCase() !== ctx.commenter.toLowerCase());
        if (others.length && !ctx.isMaintainer) {
          out.errors.push('Only maintainers can assign other people. Use `/assign` on its own to assign yourself.');
          break;
        }
        if (!ctx.isMaintainer && !ctx.isPullRequest && !labels.has(TRIAGE_ACCEPTED)) {
          out.errors.push('This issue has not been accepted yet. A maintainer will triage it first; please wait for the `triage/accepted` label before picking it up.');
          break;
        }
        if (!ctx.isMaintainer && ctx.isPullRequest) {
          out.errors.push('Only maintainers can assign pull requests.');
          break;
        }
        out.assign.push(...targets);
        break;
      }
      case 'unassign': {
        const targets = args.length ? args : [ctx.commenter];
        const others = targets.filter((u) => u.toLowerCase() !== ctx.commenter.toLowerCase());
        if (others.length && !ctx.isMaintainer) {
          out.errors.push('Only maintainers can unassign other people. Use `/unassign` on its own to unassign yourself.');
          break;
        }
        out.unassign.push(...targets);
        break;
      }
      case 'close':
      case 'reopen': {
        if (!ctx.isMaintainer && !isAuthor) {
          out.errors.push(`\`/${name}\` can only be used by maintainers or the author.`);
          break;
        }
        // Like GitHub itself, the author can only reopen what they closed, not
        // an item a maintainer closed (for example as out of scope).
        const closedBySomeoneElse = Boolean(ctx.closedBy) && ctx.closedBy.toLowerCase() !== ctx.commenter.toLowerCase();
        if (name === 'reopen' && !ctx.isMaintainer && closedBySomeoneElse) {
          out.errors.push('`/reopen` can\'t reopen an item a maintainer closed. Please ask a maintainer if you think it should be reopened.');
          break;
        }
        if (name === 'reopen') {
          for (const { label } of Object.values(CLOSE_REASON)) remove(label);
          out.state = 'open';
          break;
        }
        if (args.length) {
          const reason = CLOSE_REASON[args[0].toLowerCase()];
          if (!reason) {
            out.errors.push(`\`/close\` expects nothing or one of: ${Object.keys(CLOSE_REASON).map((k) => `\`${k}\``).join(', ')}.`);
            break;
          }
          // A reason is a maintainer's judgement, even on the author's own item.
          if (!ctx.isMaintainer) {
            out.errors.push('`/close` with a reason can only be used by maintainers. Use `/close` on its own.');
            break;
          }
          add(reason.label);
          if (!ctx.isPullRequest) out.stateReason = reason.stateReason;
        }
        out.state = 'closed';
        break;
      }
      case 'priority': {
        if (!needMaintainer(name)) break;
        const label = mapped(name, PRIORITY, args[0]);
        if (!label) break;
        for (const l of Object.values(PRIORITY)) if (l !== label) remove(l);
        add(label);
        break;
      }
      case 'milestone':
      case 'remove-milestone': {
        if (!needMaintainer(name)) break;
        // "/milestone clear" is the Prow spelling of "/remove-milestone".
        const clear = name === 'remove-milestone' || (args.length === 1 && args[0].toLowerCase() === 'clear');
        if (!clear && args.length === 0) {
          out.errors.push('`/milestone` expects a milestone title, for example `/milestone 1.0.3`, or `clear`.');
          break;
        }
        // The title is checked against the repository's open milestones when applied.
        out.milestone = { title: clear ? null : args.join(' ') };
        break;
      }
      case 'remove-priority':
        if (!needMaintainer(name)) break;
        for (const l of Object.values(PRIORITY)) remove(l);
        break;
      case 'lgtm':
        if (!needPullRequest(name)) break;
        if (isAuthor) {
          out.errors.push('`/lgtm` can\'t be used on your own pull request.');
          break;
        }
        if (!needTriager(name)) break;
        add(LGTM);
        break;
      case 'approve':
        if (!needPullRequest(name) || !needMaintainer(name)) break;
        // GitHub doesn't let authors approve their own pull request either.
        if (isAuthor) {
          out.errors.push('`/approve` can\'t be used on your own pull request. Another maintainer will review it.');
          break;
        }
        add(APPROVED);
        break;
      case 'remove-lgtm':
      case 'remove-approve':
        if (!needPullRequest(name) || !needMaintainer(name)) break;
        remove(name === 'remove-lgtm' ? LGTM : APPROVED);
        break;
      case 'hold':
        if (!needPullRequest(name) || !needParticipant(name)) break;
        add(HOLD);
        break;
      case 'unhold':
        // Anyone involved can ask to stop a merge, but only a maintainer lifts
        // it, so a contributor can't undo a maintainer's hold.
        if (!needPullRequest(name) || !needMaintainer(name)) break;
        remove(HOLD);
        break;
      case 'retest':
        if (!needPullRequest(name)) break;
        if (!ctx.isMaintainer && !isAuthor) {
          out.errors.push('`/retest` can only be used by maintainers or the author.');
          break;
        }
        out.retest = true;
        break;
      case 'cc':
      case 'uncc':
        if (!needPullRequest(name) || !needParticipant(name)) break;
        if (args.length === 0) {
          out.errors.push(`\`/${name}\` expects one or more users, for example \`/${name} @user\`.`);
          break;
        }
        (name === 'cc' ? out.reviewers : out.unreviewers).push(...args);
        break;
      default:
        // Unknown commands are ignored, like Prow does, so "/path/to/file"
        // or other text starting with a slash does not trigger a reply.
        break;
    }
  }
  return out;
}

module.exports = { plan, NEEDS_TRIAGE, HOLD };
