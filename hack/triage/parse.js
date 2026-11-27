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

// Reads slash commands from comment text.

const MAX_COMMANDS = 10;

/**
 * @typedef {object} Command
 * @property {string} name command name without the slash, lowercased
 * @property {string[]} args arguments, with a leading "@" removed
 */

/**
 * Extracts slash commands from a comment body. Commands must start a line;
 * lines inside fenced code blocks and quoted lines ("> /kind bug") are
 * ignored so that quoting a previous comment does not re-run it.
 *
 * @param {string | null | undefined} body
 * @returns {Command[]}
 */
function parseCommands(body) {
  const commands = [];
  let inFence = false;
  // GitHub doesn't render HTML comments, so commands inside them would run
  // without reviewers seeing them. An unclosed comment hides the rest.
  const visible = String(body || '').replace(/<!--[\s\S]*?(-->|$)/g, '');
  for (const raw of visible.split(/\r?\n/)) {
    const line = raw.trim();
    if (line.startsWith('```') || line.startsWith('~~~')) {
      inFence = !inFence;
      continue;
    }
    if (inFence || !line.startsWith('/')) {
      continue;
    }
    const [head, ...args] = line.split(/\s+/);
    const name = head.slice(1).toLowerCase();
    if (!/^[a-z][a-z-]*$/.test(name)) {
      continue;
    }
    commands.push({ name, args: args.map((a) => a.replace(/^@/, '')).filter(Boolean) });
    if (commands.length >= MAX_COMMANDS) {
      break;
    }
  }
  return commands;
}

module.exports = { parseCommands };
