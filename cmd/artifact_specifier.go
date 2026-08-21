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

// parseImportFileSpecifier parses an import argument of the form:
//
//	<path>[:<mainArtifactBool>]
//
// It parses from the right to avoid breaking paths that may contain ':'
// characters (e.g. Windows absolute paths like C:\...).
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
