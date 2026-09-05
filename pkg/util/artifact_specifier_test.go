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

package util

import "testing"

func TestParseImportFileSpecifier_SuffixBool(t *testing.T) {
	in := "./specs/openapi.yaml:false"
	path, main := ParseImportFileSpecifier(in)
	if path != "./specs/openapi.yaml" {
		t.Fatalf("path mismatch: got %q", path)
	}
	if main != false {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
}

func TestParseImportFileSpecifier_NoSuffix_Unchanged(t *testing.T) {
	in := "./specs/openapi.yaml"
	path, main := ParseImportFileSpecifier(in)
	if path != in {
		t.Fatalf("path mismatch: got %q", path)
	}
	if main != true {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
}
