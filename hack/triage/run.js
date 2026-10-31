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

// Entry points called from actions/github-script. They read the event,
// ask plan() what to do and apply it through the GitHub API.

const { parseCommands } = require('./parse');
const { plan, NEEDS_TRIAGE, HOLD } = require('./plan');

/**
 * The parts of actions/github-script's arguments that are used here. The
 * Octokit client is left untyped so no type packages need to be installed.
 *
 * @typedef {{ rest: any, paginate: any }} GitHub
 * @typedef {{ repo: { owner: string, repo: string }, payload: any }} Context
 * @typedef {{ warning: (message: string) => void }} Core
 */

const DOCS_URL = 'https://github.com/microcks/microcks-cli/blob/master/documentation/triage-commands.md';
const HOLD_WORKFLOW = 'hold.yml';

const MAINTAINER_PERMISSIONS = ['admin', 'maintain', 'write'];
const CONTRIBUTOR_ASSOCIATIONS = ['CONTRIBUTOR', 'COLLABORATOR', 'MEMBER', 'OWNER'];

/**
 * @param {GitHub} github
 * @param {string} owner
 * @param {string} repo
 * @param {string} username
 * @returns {Promise<boolean>}
 */
async function isMaintainer(github, owner, repo, username) {
  try {
    const { data } = await github.rest.repos.getCollaboratorPermissionLevel({ owner, repo, username });
    return MAINTAINER_PERMISSIONS.includes(data.permission);
  } catch (err) {
    // 404 means the user is not a collaborator.
    if (err.status === 404) return false;
    throw err;
  }
}

/**
 * Applies the slash commands of an issue_comment event.
 *
 * @param {{ github: GitHub, context: Context, core: Core }} args
 */
async function run({ github, context, core }) {
  const { comment, issue } = context.payload;
  if (!comment || !issue || comment.user.type === 'Bot') return;

  const commands = parseCommands(comment.body);
  if (commands.length === 0) return;

  const { owner, repo } = context.repo;
  const issue_number = issue.number;
  const commenter = comment.user.login;

  const maintainer = await isMaintainer(github, owner, repo, commenter);
  // The event payload is a snapshot from when the comment was posted; an
  // earlier comment's run may have changed labels since, so read them again.
  const { data: current } = await github.rest.issues.get({ owner, repo, issue_number });
  const result = plan(commands, {
    commenter,
    author: issue.user && issue.user.login,
    isMaintainer: maintainer,
    isContributor: maintainer || CONTRIBUTOR_ASSOCIATIONS.includes(comment.author_association),
    isPullRequest: Boolean(issue.pull_request),
    labels: (current.labels || []).map((l) => (typeof l === 'string' ? l : l.name)),
    closedBy: current.closed_by ? current.closed_by.login : undefined,
  });

  const failures = [...result.errors];
  const attempt = async (what, fn) => {
    try {
      await fn();
    } catch (err) {
      core.warning(`${what}: ${err.message}`);
      failures.push(`Could not ${what}: ${err.message}`);
    }
  };

  for (const name of result.removeLabels) {
    await attempt(`remove label \`${name}\``, () => github.rest.issues.removeLabel({ owner, repo, issue_number, name }));
  }
  if (result.addLabels.length) {
    await attempt(`add ${result.addLabels.map((l) => `\`${l}\``).join(', ')}`, () =>
      github.rest.issues.addLabels({ owner, repo, issue_number, labels: result.addLabels }));
  }
  if (result.unassign.length) {
    await attempt('unassign', () => github.rest.issues.removeAssignees({ owner, repo, issue_number, assignees: result.unassign }));
  }
  if (result.assign.length) {
    await attempt('assign', async () => {
      const { data } = await github.rest.issues.addAssignees({ owner, repo, issue_number, assignees: result.assign });
      const assigned = new Set((data.assignees || []).map((a) => a.login.toLowerCase()));
      const missing = result.assign.filter((u) => !assigned.has(u.toLowerCase()));
      if (missing.length) {
        throw new Error(`${missing.map((u) => `@${u}`).join(', ')} cannot be assigned. They need to comment on this issue first.`);
      }
    });
  }
  if (result.unreviewers.length) {
    await attempt('remove review requests', () =>
      github.rest.pulls.removeRequestedReviewers({ owner, repo, pull_number: issue_number, reviewers: result.unreviewers }));
  }
  if (result.reviewers.length) {
    await attempt('request reviews', () =>
      github.rest.pulls.requestReviewers({ owner, repo, pull_number: issue_number, reviewers: result.reviewers }));
  }
  if (result.retest) {
    await attempt('re-run failed checks', async () => {
      const { data: pr } = await github.rest.pulls.get({ owner, repo, pull_number: issue_number });
      const { data } = await github.rest.actions.listWorkflowRunsForRepo({ owner, repo, head_sha: pr.head.sha, per_page: 100 });
      const failed = data.workflow_runs.filter((r) => ['failure', 'cancelled', 'timed_out'].includes(r.conclusion));
      if (failed.length === 0) {
        throw new Error('there are no failed checks on the latest commit.');
      }
      for (const r of failed) {
        await github.rest.actions.reRunWorkflowFailedJobs({ owner, repo, run_id: r.id });
      }
    });
  }
  if (result.state) {
    await attempt(result.state === 'closed' ? 'close' : 'reopen', () =>
      github.rest.issues.update({
        owner, repo, issue_number, state: result.state, ...(result.stateReason ? { state_reason: result.stateReason } : {}),
      }));
  }
  // Labels set with GITHUB_TOKEN don't start workflows, so re-run the hold
  // check by hand for it to see the new label.
  if ([...result.addLabels, ...result.removeLabels].includes(HOLD)) {
    await attempt('update the hold check', async () => {
      const { data: pr } = await github.rest.pulls.get({ owner, repo, pull_number: issue_number });
      const { data } = await github.rest.actions.listWorkflowRuns({
        owner, repo, workflow_id: HOLD_WORKFLOW, head_sha: pr.head.sha, per_page: 1,
      });
      const [latest] = data.workflow_runs;
      if (latest) {
        await github.rest.actions.reRunWorkflow({ owner, repo, run_id: latest.id });
      }
    });
  }

  if (result.milestone) {
    const { title } = result.milestone;
    await attempt(title === null ? 'clear the milestone' : 'set the milestone', async () => {
      let number = null;
      if (title !== null) {
        const open = await github.paginate(github.rest.issues.listMilestones, { owner, repo, state: 'open', per_page: 100 });
        const match = open.find((m) => m.title.toLowerCase() === title.toLowerCase());
        if (!match) {
          // The title is not echoed back, so a comment can't inject text into the reply.
          throw new Error(`no open milestone has that title. Open milestones: ${open.map((m) => `\`${m.title}\``).join(', ') || 'none'}.`);
        }
        number = match.number;
      }
      await github.rest.issues.update({ owner, repo, issue_number, milestone: number });
    });
  }

  const changed = result.addLabels.length || result.removeLabels.length || result.assign.length
    || result.unassign.length || result.reviewers.length || result.unreviewers.length || result.retest || result.state
    || result.milestone;
  if (changed && failures.length === 0) {
    await github.rest.reactions.createForIssueComment({ owner, repo, comment_id: comment.id, content: '+1' });
  }
  if (failures.length) {
    await github.rest.reactions.createForIssueComment({ owner, repo, comment_id: comment.id, content: 'confused' });
    const greeting = `@${commenter} thanks! Some commands were not applied`;
    // Reply once per person and issue, so repeated commands can't make the
    // bot spam the thread; later failures only get the reaction.
    if (!maintainer && await alreadyReplied(github, owner, repo, issue_number, greeting)) return;
    const lines = failures.map((f) => `- ${f}`).join('\n');
    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number,
      body: `${greeting}:\n\n${lines}\n\nSee the [triage commands](${DOCS_URL}) for details.`,
    });
  }
}

/**
 * @param {GitHub} github
 * @param {string} owner
 * @param {string} repo
 * @param {number} issue_number
 * @param {string} greeting start of the bot's reply to this commenter
 * @returns {Promise<boolean>}
 */
async function alreadyReplied(github, owner, repo, issue_number, greeting) {
  const comments = await github.paginate(github.rest.issues.listComments, { owner, repo, issue_number, per_page: 100 });
  return comments.some((c) => c.user && c.user.type === 'Bot' && String(c.body).startsWith(greeting));
}

/**
 * Adds needs-triage to newly opened issues, unless a maintainer already
 * triaged them through a label.
 *
 * @param {{ github: GitHub, context: Context }} args
 */
async function labelNewIssue({ github, context }) {
  const { issue } = context.payload;
  if (!issue) return;
  const labels = (issue.labels || []).map((l) => (typeof l === 'string' ? l : l.name));
  if (labels.some((l) => l.startsWith('triage/'))) return;
  await github.rest.issues.addLabels({
    ...context.repo,
    issue_number: issue.number,
    labels: [NEEDS_TRIAGE],
  });
}

module.exports = { run, labelNewIssue };
