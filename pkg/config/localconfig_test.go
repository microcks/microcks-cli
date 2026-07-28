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
