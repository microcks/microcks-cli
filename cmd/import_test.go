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

import "testing"

func TestParseImportFileArg(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantFile         string
		wantMainArtifact bool
	}{
		{
			name:             "Windows absolute path",
			input:            `C:\Temp\api.yaml`,
			wantFile:         `C:\Temp\api.yaml`,
			wantMainArtifact: true,
		},
		{
			name:             "Windows absolute path with false suffix",
			input:            `C:\Temp\api.yaml:false`,
			wantFile:         `C:\Temp\api.yaml`,
			wantMainArtifact: false,
		},
		{
			name:             "relative path defaults to main artifact",
			input:            "./api.yaml",
			wantFile:         "./api.yaml",
			wantMainArtifact: true,
		},
		{
			name:             "relative path with true suffix",
			input:            "./api.yaml:true",
			wantFile:         "./api.yaml",
			wantMainArtifact: true,
		},
		{
			name:             "relative path with false suffix",
			input:            "./api.yaml:false",
			wantFile:         "./api.yaml",
			wantMainArtifact: false,
		},
		{
			name:             "colon-containing filename",
			input:            "path:with:colon.yaml",
			wantFile:         "path:with:colon.yaml",
			wantMainArtifact: true,
		},
		{
			name:             "colon-containing filename with false suffix",
			input:            "path:with:colon.yaml:false",
			wantFile:         "path:with:colon.yaml",
			wantMainArtifact: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotMainArtifact := parseImportFileArg(tt.input)
			if gotFile != tt.wantFile {
				t.Errorf("file = %q, want %q", gotFile, tt.wantFile)
			}
			if gotMainArtifact != tt.wantMainArtifact {
				t.Errorf("mainArtifact = %t, want %t", gotMainArtifact, tt.wantMainArtifact)
			}
		})
	}
}
