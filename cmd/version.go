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
	"fmt"

	"github.com/microcks/microcks-cli/version"
)

type versionCommand struct {
}

// NewVersionCommand build a new VersionCommand implementation
func NewVersionCommand() Command {
	return new(versionCommand)
}

// Execute implementation on versionCommand structure
func (c *versionCommand) Execute() {
	fmt.Println(version.Version)
}
