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
package output

import (
	"encoding/json"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"gopkg.in/yaml.v2"
)

// YAMLFormatter renders the test result as YAML. It round-trips through JSON so
// the keys match the JSON field names (camelCase) rather than Go field names.
type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(r *connectors.TestResult) (string, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	var generic interface{}
	if err := json.Unmarshal(j, &generic); err != nil {
		return "", err
	}
	b, err := yaml.Marshal(generic)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
