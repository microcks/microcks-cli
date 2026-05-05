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
	"strconv"
	"strings"
)

// parseImportURLSpecifier parses an import-url argument of the form:
//   <url>[:<mainArtifactBool>[:<secret>]]
//
// It is intentionally parsed from the right side so that normal URLs containing
// additional ':' characters (scheme, ports, etc.) are preserved unchanged unless
// they end with the supported suffixes.
func parseImportURLSpecifier(spec string) (url string, mainArtifact bool, secret string) {
	mainArtifact = true

	lastColon := strings.LastIndex(spec, ":")
	if lastColon == -1 {
		return spec, mainArtifact, ""
	}

	tail := spec[lastColon+1:]
	if b, err := strconv.ParseBool(tail); err == nil {
		return spec[:lastColon], b, ""
	}

	// Might be <url>:<bool>:<secret> — only treat it as such if the second-to-last
	// segment parses as bool.
	secretCandidate := tail
	rest := spec[:lastColon]
	secondColon := strings.LastIndex(rest, ":")
	if secondColon == -1 {
		return spec, mainArtifact, ""
	}
	boolCandidate := rest[secondColon+1:]
	if b, err := strconv.ParseBool(boolCandidate); err == nil {
		return rest[:secondColon], b, secretCandidate
	}

	return spec, mainArtifact, ""
}

// parseImportFileSpecifier parses an import argument of the form:
//   <path>[:<mainArtifactBool>]
//
// Like parseImportURLSpecifier, it parses from the right to avoid breaking
// paths that may contain ':' characters.
func parseImportFileSpecifier(spec string) (path string, mainArtifact bool) {
	mainArtifact = true

	lastColon := strings.LastIndex(spec, ":")
	if lastColon == -1 {
		return spec, mainArtifact
	}

	tail := spec[lastColon+1:]
	if b, err := strconv.ParseBool(tail); err == nil {
		return spec[:lastColon], b
	}

	return spec, mainArtifact
}

