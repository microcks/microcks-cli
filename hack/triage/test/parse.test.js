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

test('parseCommands reads commands at the start of lines', () => {
  assert.deepEqual(parseCommands('Looks good.\n/triage accepted\n  /kind bug\n/assign @dave'), [
    { name: 'triage', args: ['accepted'] },
    { name: 'kind', args: ['bug'] },
    { name: 'assign', args: ['dave'] },
  ]);
});

test('parseCommands ignores code blocks, quotes, inline slashes and paths', () => {
  const body = [
    '```', '/close', '```',
    '> /kind bug',
    'please run `/assign`',
    'see cmd/test.go',
    '/usr/local/bin/microcks',
  ].join('\n');
  assert.deepEqual(parseCommands(body), []);
});

test('parseCommands ignores commands hidden in HTML comments', () => {
  assert.deepEqual(parseCommands('Nice!\n<!--\n/lgtm\n-->\n/kind bug\n<!-- /hold cancel -->\n<!--\n/approve'), [
    { name: 'kind', args: ['bug'] },
  ]);
});

test('parseCommands caps the number of commands', () => {
  assert.equal(parseCommands(Array(50).fill('/kind bug').join('\n')).length, 10);
});

