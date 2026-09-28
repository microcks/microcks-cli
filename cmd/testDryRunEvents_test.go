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
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDryRunEventWriterEmitsOneJSONDocumentPerLine(t *testing.T) {
	var buffer bytes.Buffer
	events := newDryRunEventWriter(&buffer)
	if err := events.emit(dryRunWatchEvent{Type: "ready", Endpoint: "http://localhost:1234"}); err != nil {
		t.Fatalf("emit returned error: %v", err)
	}
	if err := events.emit(dryRunWatchEvent{Type: "waiting"}); err != nil {
		t.Fatalf("emit returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), buffer.String())
	}
	for _, line := range lines {
		var event dryRunWatchEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("line is not JSON: %v", err)
		}
		if event.Timestamp == "" {
			t.Fatal("event timestamp is empty")
		}
	}
}
