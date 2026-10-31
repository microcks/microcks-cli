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

// Run with: node --test hack/triage/test/*.test.js

const test = require('node:test');
const assert = require('node:assert/strict');
const { parseCommands } = require('../parse');
const { plan } = require('../plan');

const maintainer = { commenter: 'alice', author: 'bob', isMaintainer: true, isContributor: true, isPullRequest: false, labels: [] };
// A newcomer without commits in the repository.
const contributor = { ...maintainer, commenter: 'carol', isMaintainer: false, isContributor: false };
// Someone who already has commits in the repository.
const pastContributor = { ...contributor, commenter: 'dan', isContributor: true };

test('/triage accepted replaces needs-triage and other triage labels', () => {
  const r = plan(parseCommands('/triage accepted'), { ...maintainer, labels: ['needs-triage', 'triage/needs-information'] });
  assert.deepEqual(r.addLabels, ['triage/accepted']);
  assert.deepEqual(r.removeLabels.sort(), ['needs-triage', 'triage/needs-information']);
  assert.deepEqual(r.errors, []);
});

test('/remove-triage puts needs-triage back', () => {
  const r = plan(parseCommands('/remove-triage'), { ...maintainer, labels: ['triage/accepted'] });
  assert.deepEqual(r.removeLabels, ['triage/accepted']);
  assert.deepEqual(r.addLabels, ['needs-triage']);
});

test('label commands map to existing repository labels', () => {
  const r = plan(parseCommands('/kind bug\n/area docs\n/platform windows'), maintainer);
  assert.deepEqual(r.addLabels, ['kind/bug', 'component/documentation', 'platform/windows']);
});

test('/remove-kind removes only labels that are present', () => {
  const r = plan(parseCommands('/remove-kind bug\n/remove-kind task'), { ...maintainer, labels: ['kind/bug'] });
  assert.deepEqual(r.removeLabels, ['kind/bug']);
  assert.deepEqual(r.errors, []);
});

test('unknown label values are reported', () => {
  const r = plan(parseCommands('/kind nonsense\n/kind'), maintainer);
  assert.equal(r.addLabels.length, 0);
  assert.equal(r.errors.length, 2);
  assert.match(r.errors[0], /expects one of/);
});

test('newcomers cannot triage or label', () => {
  const r = plan(parseCommands('/triage accepted\n/kind bug\n/keep-open\n/good-first-issue'), contributor);
  assert.equal(r.addLabels.length, 0);
  assert.equal(r.errors.length, 4);
  assert.match(r.errors[0], /people who have contributed/);
  assert.match(r.errors[2], /only be used by maintainers/);
});

test('past contributors can triage and label other people\'s issues', () => {
  const r = plan(parseCommands('/triage accepted\n/kind bug\n/area cli\n/good-first-issue'), { ...pastContributor, labels: ['needs-triage'] });
  assert.deepEqual(r.addLabels, ['triage/accepted', 'kind/bug', 'component/cli', 'good first issue']);
  assert.deepEqual(r.removeLabels, ['needs-triage']);
  assert.deepEqual(r.errors, []);
});

test('past contributors cannot triage their own issue', () => {
  const r = plan(parseCommands('/triage accepted\n/kind bug'), { ...pastContributor, author: 'dan' });
  assert.equal(r.addLabels.length, 0);
  assert.equal(r.errors.length, 2);
  assert.match(r.errors[0], /can't be used on your own issue/);
});

test('maintainers can triage their own issue', () => {
  const r = plan(parseCommands('/triage accepted'), { ...maintainer, author: 'alice' });
  assert.deepEqual(r.addLabels, ['triage/accepted']);
});

test('only maintainers can remove labels or use /keep-open', () => {
  const labels = ['triage/accepted', 'kind/bug', 'good first issue'];
  const r = plan(parseCommands('/remove-triage\n/remove-kind bug\n/remove-good-first-issue\n/keep-open'), { ...pastContributor, labels });
  assert.deepEqual(r.removeLabels, []);
  assert.deepEqual(r.addLabels, []);
  assert.equal(r.errors.length, 4);
  r.errors.forEach((e) => assert.match(e, /only be used by maintainers/));
});

test('/good-first-issue and /help need triage/accepted on issues', () => {
  const before = plan(parseCommands('/good-first-issue\n/help'), maintainer);
  assert.equal(before.addLabels.length, 0);
  assert.equal(before.errors.length, 2);

  const after = plan(parseCommands('/good-first-issue\n/help'), { ...maintainer, labels: ['triage/accepted'] });
  assert.deepEqual(after.addLabels, ['good first issue', 'help wanted']);
});

test('triaging and flagging in the same comment works', () => {
  const r = plan(parseCommands('/triage accepted\n/good-first-issue'), { ...maintainer, labels: ['needs-triage'] });
  assert.deepEqual(r.addLabels, ['triage/accepted', 'good first issue']);
  assert.deepEqual(r.errors, []);
});

test('contributors can self-assign only accepted issues', () => {
  const pending = plan(parseCommands('/assign'), contributor);
  assert.deepEqual(pending.assign, []);
  assert.match(pending.errors[0], /not been accepted yet/);

  const accepted = plan(parseCommands('/assign'), { ...contributor, labels: ['triage/accepted'] });
  assert.deepEqual(accepted.assign, ['carol']);
  assert.deepEqual(accepted.errors, []);
});

test('contributors cannot assign others or pull requests', () => {
  const others = plan(parseCommands('/assign @dave'), { ...contributor, labels: ['triage/accepted'] });
  assert.deepEqual(others.assign, []);
  assert.match(others.errors[0], /Only maintainers can assign other people/);

  const pr = plan(parseCommands('/assign'), { ...contributor, isPullRequest: true });
  assert.deepEqual(pr.assign, []);
  assert.match(pr.errors[0], /pull requests/);
});

test('maintainers can assign anyone, before triage', () => {
  const r = plan(parseCommands('/assign @dave @erin'), maintainer);
  assert.deepEqual(r.assign, ['dave', 'erin']);
});

test('anyone can unassign themselves but not others', () => {
  assert.deepEqual(plan(parseCommands('/unassign'), contributor).unassign, ['carol']);
  const r = plan(parseCommands('/unassign @dave'), contributor);
  assert.deepEqual(r.unassign, []);
  assert.equal(r.errors.length, 1);
});

test('/close and /reopen work for maintainers and the author only', () => {
  assert.equal(plan(parseCommands('/close'), maintainer).state, 'closed');
  assert.equal(plan(parseCommands('/reopen'), { ...contributor, commenter: 'bob' }).state, 'open');
  const r = plan(parseCommands('/close'), contributor);
  assert.equal(r.state, null);
  assert.equal(r.errors.length, 1);
});

test('authors can reopen only what they closed themselves', () => {
  const author = { ...contributor, commenter: 'bob' };
  assert.equal(plan(parseCommands('/reopen'), { ...author, closedBy: 'bob' }).state, 'open');

  const r = plan(parseCommands('/reopen'), { ...author, closedBy: 'alice' });
  assert.equal(r.state, null);
  assert.match(r.errors[0], /a maintainer closed/);

  assert.equal(plan(parseCommands('/reopen'), { ...maintainer, closedBy: 'erin' }).state, 'open');
});

test('unknown commands are ignored without errors', () => {
  const r = plan(parseCommands('/test all\n/meow'), contributor);
  assert.deepEqual(r, {
    addLabels: [], removeLabels: [], assign: [], unassign: [], reviewers: [], unreviewers: [], retest: false, state: null, stateReason: null, milestone: null, errors: [],
  });
});

const prMaintainer = { ...maintainer, isPullRequest: true };
const prPastContributor = { ...pastContributor, isPullRequest: true };
const prNewcomer = { ...contributor, isPullRequest: true };

test('pull request commands are rejected on issues', () => {
  const r = plan(parseCommands('/lgtm\n/approve\n/hold\n/retest\n/cc @dave'), maintainer);
  assert.deepEqual(r.addLabels, []);
  assert.equal(r.retest, false);
  assert.equal(r.errors.length, 5);
  r.errors.forEach((e) => assert.match(e, /only be used on pull requests/));
});

test('/lgtm is for maintainers and past contributors, never on their own pull request', () => {
  assert.deepEqual(plan(parseCommands('/lgtm'), prPastContributor).addLabels, ['lgtm']);
  assert.deepEqual(plan(parseCommands('/lgtm'), prMaintainer).addLabels, ['lgtm']);

  const newcomer = plan(parseCommands('/lgtm'), prNewcomer);
  assert.deepEqual(newcomer.addLabels, []);
  assert.match(newcomer.errors[0], /people who have contributed/);

  const own = plan(parseCommands('/lgtm'), { ...prMaintainer, author: 'alice' });
  assert.deepEqual(own.addLabels, []);
  assert.match(own.errors[0], /your own pull request/);
});

test('/approve is maintainer-only, never on their own pull request', () => {
  assert.deepEqual(plan(parseCommands('/approve'), prMaintainer).addLabels, ['approved']);

  const own = plan(parseCommands('/approve'), { ...prMaintainer, author: 'alice' });
  assert.deepEqual(own.addLabels, []);
  assert.match(own.errors[0], /your own pull request/);

  const r = plan(parseCommands('/approve'), prPastContributor);
  assert.deepEqual(r.addLabels, []);
  assert.match(r.errors[0], /only be used by maintainers/);
});

test('/remove-lgtm, /remove-approve and the cancel spelling are maintainer-only', () => {
  const labels = ['lgtm', 'approved'];
  const r = plan(parseCommands('/lgtm cancel\n/remove-approve'), { ...prMaintainer, labels });
  assert.deepEqual(r.removeLabels, ['lgtm', 'approved']);
  assert.deepEqual(r.addLabels, []);

  const denied = plan(parseCommands('/remove-lgtm\n/approve cancel'), { ...prPastContributor, labels });
  assert.deepEqual(denied.removeLabels, []);
  assert.equal(denied.errors.length, 2);
});

test('/hold works for participants and the author, /unhold only for maintainers', () => {
  assert.deepEqual(plan(parseCommands('/hold'), prPastContributor).addLabels, ['do-not-merge/hold']);
  assert.deepEqual(plan(parseCommands('/hold'), { ...prNewcomer, commenter: 'bob' }).addLabels, ['do-not-merge/hold']);
  assert.equal(plan(parseCommands('/hold'), prNewcomer).errors.length, 1);

  const held = ['do-not-merge/hold'];
  assert.deepEqual(plan(parseCommands('/hold cancel'), { ...prMaintainer, labels: held }).removeLabels, held);
  assert.deepEqual(plan(parseCommands('/unhold'), { ...prMaintainer, labels: held }).removeLabels, held);
  for (const who of [prPastContributor, { ...prNewcomer, commenter: 'bob' }]) {
    const r = plan(parseCommands('/unhold\n/hold cancel'), { ...who, labels: held });
    assert.deepEqual(r.removeLabels, []);
    assert.equal(r.errors.length, 2);
  }
});

test('/priority is maintainer-only and keeps a single priority', () => {
  const r = plan(parseCommands('/priority important-soon'), { ...maintainer, labels: ['priority/backlog'] });
  assert.deepEqual(r.addLabels, ['priority/important-soon']);
  assert.deepEqual(r.removeLabels, ['priority/backlog']);

  assert.deepEqual(plan(parseCommands('/remove-priority'), { ...maintainer, labels: ['priority/backlog'] }).removeLabels, ['priority/backlog']);
  assert.match(plan(parseCommands('/priority urgent'), maintainer).errors[0], /expects one of/);

  const denied = plan(parseCommands('/priority critical-urgent\n/remove-priority'), { ...pastContributor, labels: ['priority/backlog'] });
  assert.deepEqual(denied.addLabels, []);
  assert.deepEqual(denied.removeLabels, []);
  assert.equal(denied.errors.length, 2);
});

test('/close with a reason labels the item and is maintainer-only', () => {
  const r = plan(parseCommands('/close wontfix'), maintainer);
  assert.equal(r.state, 'closed');
  assert.equal(r.stateReason, 'not_planned');
  assert.deepEqual(r.addLabels, ['status/wontfix']);

  // Pull requests have no state_reason.
  const pr = plan(parseCommands('/close duplicate'), prMaintainer);
  assert.deepEqual(pr.addLabels, ['status/duplicate']);
  assert.equal(pr.stateReason, null);

  const author = plan(parseCommands('/close invalid'), { ...contributor, commenter: 'bob' });
  assert.equal(author.state, null);
  assert.match(author.errors[0], /with a reason can only be used by maintainers/);

  const typo = plan(parseCommands('/close dupe'), maintainer);
  assert.equal(typo.state, null);
  assert.match(typo.errors[0], /expects nothing or one of/);
});

test('/milestone is maintainer-only and can be cleared', () => {
  assert.deepEqual(plan(parseCommands('/milestone 1.0.3'), maintainer).milestone, { title: '1.0.3' });
  assert.deepEqual(plan(parseCommands('/milestone Next Release'), maintainer).milestone, { title: 'Next Release' });
  assert.deepEqual(plan(parseCommands('/milestone clear'), maintainer).milestone, { title: null });
  assert.deepEqual(plan(parseCommands('/remove-milestone'), maintainer).milestone, { title: null });
  assert.match(plan(parseCommands('/milestone'), maintainer).errors[0], /expects a milestone title/);

  const denied = plan(parseCommands('/milestone 1.0.3\n/remove-milestone'), pastContributor);
  assert.equal(denied.milestone, null);
  assert.equal(denied.errors.length, 2);
});

test('/reopen removes the close reason', () => {
  const r = plan(parseCommands('/reopen'), { ...maintainer, labels: ['status/duplicate', 'kind/bug'] });
  assert.equal(r.state, 'open');
  assert.deepEqual(r.removeLabels, ['status/duplicate']);
});

test('/retest works for maintainers and the author only', () => {
  assert.equal(plan(parseCommands('/retest'), prMaintainer).retest, true);
  assert.equal(plan(parseCommands('/retest'), { ...prNewcomer, commenter: 'bob' }).retest, true);
  const r = plan(parseCommands('/retest'), prPastContributor);
  assert.equal(r.retest, false);
  assert.match(r.errors[0], /maintainers or the author/);
});

test('/cc and /uncc request and remove reviewers', () => {
  const r = plan(parseCommands('/cc @dave @erin\n/uncc @frank'), prPastContributor);
  assert.deepEqual(r.reviewers, ['dave', 'erin']);
  assert.deepEqual(r.unreviewers, ['frank']);

  assert.match(plan(parseCommands('/cc'), prMaintainer).errors[0], /expects one or more users/);
  assert.equal(plan(parseCommands('/cc @dave'), prNewcomer).errors.length, 1);
});

