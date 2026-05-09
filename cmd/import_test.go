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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImportEntry(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantPath         string
		wantMainArtifact bool
	}{
		{"relative path, no suffix", "./api.yaml", "./api.yaml", true},
		{"relative path, :true suffix", "./api.yaml:true", "./api.yaml", true},
		{"relative path, :false suffix", "./api.yaml:false", "./api.yaml", false},
		{"relative path, :TRUE suffix (ParseBool)", "./api.yaml:TRUE", "./api.yaml", true},
		{"relative path, :0 suffix (ParseBool)", "./api.yaml:0", "./api.yaml", false},
		{"Windows absolute path, no suffix", `C:\Temp\api.yaml`, `C:\Temp\api.yaml`, true},
		{"Windows absolute path, :false suffix", `C:\Temp\api.yaml:false`, `C:\Temp\api.yaml`, false},
		{"Windows absolute path, :true suffix", `C:\Temp\api.yaml:true`, `C:\Temp\api.yaml`, true},
		{"path with multiple colons, no bool suffix", "path:with:colon.yaml", "path:with:colon.yaml", true},
		{"path with multiple colons, :false suffix", "path:with:colon.yaml:false", "path:with:colon.yaml", false},
		{"non-bool trailing segment is part of path", "./api.yaml:notabool", "./api.yaml:notabool", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotMain := parseImportEntry(tt.input)
			assert.Equal(t, tt.wantPath, gotPath, "path")
			assert.Equal(t, tt.wantMainArtifact, gotMain, "mainArtifact")
		})
	}
}
