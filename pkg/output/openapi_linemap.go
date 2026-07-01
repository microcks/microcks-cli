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
	"os"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

// operationLine returns the 1-based line of an operation (e.g. "GET /products")
// within an OpenAPI spec file, or 0 if it can't be determined (non-YAML spec,
// parse error, or operation not found). yaml.v3 nodes carry line numbers, which
// yaml.v2 does not expose.
func operationLine(specPath, operationName string) int {
	method, path, ok := splitOperation(operationName)
	if !ok {
		return 0
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		return 0
	}

	var doc yamlv3.Node
	if err := yamlv3.Unmarshal(data, &doc); err != nil {
		return 0
	}
	root := &doc
	if root.Kind == yamlv3.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	_, pathsNode := mappingEntry(root, "paths")
	if pathsNode == nil {
		return 0
	}
	_, pathNode := mappingEntry(pathsNode, path)
	if pathNode == nil {
		return 0
	}
	line, _ := mappingEntry(pathNode, method)
	return line
}

// splitOperation parses "GET /products" into ("get", "/products", true).
func splitOperation(operationName string) (method, path string, ok bool) {
	parts := strings.SplitN(strings.TrimSpace(operationName), " ", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return strings.ToLower(parts[0]), parts[1], true
}

// mappingEntry looks up key in a mapping node and returns the key node's line
// (where "key:" appears) and its value node.
func mappingEntry(node *yamlv3.Node, key string) (int, *yamlv3.Node) {
	if node == nil || node.Kind != yamlv3.MappingNode {
		return 0, nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i].Line, node.Content[i+1]
		}
	}
	return 0, nil
}
