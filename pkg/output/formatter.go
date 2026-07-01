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

// Package output renders a completed Microcks TestResult in a selectable format
// (text, json, yaml, github-actions) for the `microcks test --output` flag.
package output

import (
	"fmt"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

// OutputFormat is a supported value of the --output flag.
type OutputFormat string

const (
	FormatText          OutputFormat = "text"
	FormatJSON          OutputFormat = "json"
	FormatYAML          OutputFormat = "yaml"
	FormatGitHubActions OutputFormat = "github-actions"
)

// Formatter renders a completed TestResult for a chosen output target. The
// returned string is written to stdout by the caller; formatters that also have
// side effects (e.g. github-actions writing the job step summary) perform them
// during Format.
type Formatter interface {
	Format(result *connectors.TestResult) (string, error)
}

// Option configures a Formatter.
type Option func(*config)

type config struct {
	artifactPath string
}

func WithArtifactPath(path string) Option {
	return func(c *config) { c.artifactPath = path }
}

// NewFormatter returns the Formatter for the given format.
func NewFormatter(format OutputFormat, opts ...Option) (Formatter, error) {
	c := &config{}
	for _, o := range opts {
		o(c)
	}
	switch format {
	case FormatText:
		return &TextFormatter{}, nil
	case FormatJSON:
		return &JSONFormatter{}, nil
	case FormatYAML:
		return &YAMLFormatter{}, nil
	case FormatGitHubActions:
		return &GitHubActionsFormatter{artifactPath: c.artifactPath}, nil
	default:
		return nil, fmt.Errorf("unsupported output format %q (use: text, json, yaml, github-actions)", format)
	}
}

// IsValid reports whether s is a supported output format.
func IsValid(s string) bool {
	switch OutputFormat(s) {
	case FormatText, FormatJSON, FormatYAML, FormatGitHubActions:
		return true
	}
	return false
}
