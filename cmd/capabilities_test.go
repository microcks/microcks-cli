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
package cmd

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestCapabilitiesCommandOutputsJSON(t *testing.T) {
	out, err := executeCLIForTest(t, "capabilities", "--output", "json")
	if err != nil {
		t.Fatalf("command returned error: %v", err)
	}

	var document capabilitiesDocument
	if err := json.Unmarshal([]byte(out), &document); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if document.SchemaVersion != capabilitiesSchemaVersion {
		t.Fatalf("unexpected schema version: %s", document.SchemaVersion)
	}
	if document.CLIVersion == "" {
		t.Fatal("expected a CLI version")
	}
	expectedCapabilities := []string{
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
		"test.get.json",
	}
	if !slices.Equal(document.Capabilities, expectedCapabilities) {
		t.Fatalf("unexpected capabilities:\n got: %#v\nwant: %#v", document.Capabilities, expectedCapabilities)
	}

	seen := make(map[string]struct{}, len(document.Capabilities))
	for _, capability := range document.Capabilities {
		if _, duplicate := seen[capability]; duplicate {
			t.Errorf("duplicate capability %q", capability)
		}
		seen[capability] = struct{}{}
	}
}

func TestCapabilitiesCommandRejectsUnsupportedOutput(t *testing.T) {
	_, err := executeCLIForTest(t, "capabilities", "--output", "yaml")
	if err == nil {
		t.Fatal("expected unsupported output format to fail")
	}
}
