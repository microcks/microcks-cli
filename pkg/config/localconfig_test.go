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
package config

import (
	"slices"
	"testing"
)

func TestRemoveInstance_ByContainerID(t *testing.T) {
	localConfig := &LocalConfig{
		Instances: []Instance{
			{Name: "microcks", ContainerID: "old-id-123", Status: "Running", Port: "8585"},
			{Name: "staging", ContainerID: "staging-id-456", Status: "Running", Port: "8586"},
		},
	}

	removed := localConfig.RemoveInstance("old-id-123")

	if !removed {
		t.Error("expected RemoveInstance to return true")
	}
	// staging should still be there
	if len(localConfig.Instances) != 1 {
		t.Errorf("expected 1 instance remaining, got %d", len(localConfig.Instances))
	}
	if localConfig.Instances[0].ContainerID != "staging-id-456" {
		t.Errorf("expected staging instance to remain, got %s", localConfig.Instances[0].ContainerID)
	}
}

func TestRemoveInstance_NoDuplicatesAfterRecreate(t *testing.T) {
	localConfig := &LocalConfig{
		Instances: []Instance{
			{Name: "microcks", ContainerID: "old-id-123", Status: "Running", Port: "8585"},
			{Name: "staging", ContainerID: "staging-id-456", Status: "Running", Port: "8586"},
		},
	}

	localConfig.RemoveInstance("old-id-123")

	localConfig.UpsertInstance(Instance{
		Name:        "microcks",
		ContainerID: "new-id-789",
		Status:      "Running",
		Port:        "8585",
	})

	// staging + recreated microcks = 2, no duplicates
	if len(localConfig.Instances) != 2 {
		t.Errorf("expected 2 instances, got %d — duplicate entries present", len(localConfig.Instances))
	}
	// verify microcks has new ID
	for _, i := range localConfig.Instances {
		if i.Name == "microcks" && i.ContainerID != "new-id-789" {
			t.Errorf("expected new-id-789, got %s", i.ContainerID)
		}
	}
}

func TestWatchConfig_UpsertEntry_RejectsEmptyContext(t *testing.T) {
	w := &WatchConfig{}

	err := w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  nil,
	})
	if err == nil {
		t.Fatal("expected error for nil context, got nil")
	}

	err = w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  []string{},
	})
	if err == nil {
		t.Fatal("expected error for empty context slice, got nil")
	}
}

func TestWatchConfig_UpsertEntry_RejectsBlankContext(t *testing.T) {
	w := &WatchConfig{}

	err := w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  []string{""},
	})
	if err == nil {
		t.Fatal("expected error for empty string context, got nil")
	}

	err = w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  []string{"   "},
	})
	if err == nil {
		t.Fatal("expected error for whitespace-only context, got nil")
	}

	err = w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  []string{"valid", " "},
	})
	if err == nil {
		t.Fatal("expected error when one context is blank, got nil")
	}
}

func TestWatchConfig_UpsertEntry_MergesAndDeduplicatesContexts(t *testing.T) {
	w := &WatchConfig{}

	// Add new entry with duplicate contexts in input
	if err := w.UpsertEntry(WatchEntry{
		FilePath:     "api.yaml",
		Context:      []string{"dev", "dev"},
		MainArtifact: false,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(w.Entries))
	}
	if len(w.Entries[0].Context) != 1 || w.Entries[0].Context[0] != "dev" {
		t.Fatalf("expected [dev], got %v", w.Entries[0].Context)
	}

	// Upsert with new context "staging"
	if err := w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  []string{"staging"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Entries[0].Context) != 2 {
		t.Fatalf("expected 2 contexts, got %v", w.Entries[0].Context)
	}

	// Re-upsert with already existing context "staging" -> must preserve "dev" and not wipe out contexts
	if err := w.UpsertEntry(WatchEntry{
		FilePath:     "api.yaml",
		Context:      []string{"staging"},
		MainArtifact: true,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Entries[0].Context) != 2 {
		t.Fatalf("expected contexts to be preserved, got %v", w.Entries[0].Context)
	}
	if !slices.Contains(w.Entries[0].Context, "dev") || !slices.Contains(w.Entries[0].Context, "staging") {
		t.Fatalf("expected [dev, staging], got %v", w.Entries[0].Context)
	}
	if !w.Entries[0].MainArtifact {
		t.Fatal("expected MainArtifact to be updated to true")
	}

	// Upsert with multiple contexts including existing and new
	if err := w.UpsertEntry(WatchEntry{
		FilePath: "api.yaml",
		Context:  []string{"prod", "dev", "prod"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Entries[0].Context) != 3 {
		t.Fatalf("expected 3 contexts [dev, staging, prod], got %v", w.Entries[0].Context)
	}
	if !slices.Contains(w.Entries[0].Context, "prod") {
		t.Fatalf("expected prod to be added, got %v", w.Entries[0].Context)
	}
}
