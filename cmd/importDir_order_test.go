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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderedImportClient struct {
	MockMicrocksClient
	primary []bool
}

func (c *orderedImportClient) UploadArtifact(file string, primary bool) (string, error) {
	c.primary = append(c.primary, primary)
	return c.MockMicrocksClient.UploadArtifact(file, primary)
}

func TestImportDirectoryPrimaryOrder(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		config      ImportConfig
		wantOrder   []string
		wantPrimary []bool
		failedFile  string
	}{
		{
			name:        "primary artifacts precede secondary artifacts with stable order",
			files:       []string{"a-examples.yaml", "b-openapi.yaml", "c-collection.json", "d-swagger.json", "e-metadata.yaml"},
			wantOrder:   []string{"b-openapi.yaml", "d-swagger.json", "a-examples.yaml", "c-collection.json", "e-metadata.yaml"},
			wantPrimary: []bool{true, true, false, false, false},
		},
		{
			name:        "primary artifacts in later directories are imported first",
			files:       []string{"a/examples.yaml", "z/openapi.yaml"},
			config:      ImportConfig{Recursive: true},
			wantOrder:   []string{"z/openapi.yaml", "a/examples.yaml"},
			wantPrimary: []bool{true, false},
		},
		{
			name:        "primary-only order is preserved",
			files:       []string{"a-openapi.yaml", "b-swagger.json"},
			wantOrder:   []string{"a-openapi.yaml", "b-swagger.json"},
			wantPrimary: []bool{true, true},
		},
		{
			name:        "secondary-only order is preserved",
			files:       []string{"a-examples.yaml", "b-collection.json"},
			wantOrder:   []string{"a-examples.yaml", "b-collection.json"},
			wantPrimary: []bool{false, false},
		},
		{
			name:        "pattern and nonrecursive filters are preserved",
			files:       []string{"a-examples.yaml", "b-openapi.yaml", "c-collection.json", "nested/swagger.yaml"},
			config:      ImportConfig{Pattern: "*.yaml"},
			wantOrder:   []string{"b-openapi.yaml", "a-examples.yaml"},
			wantPrimary: []bool{true, false},
		},
		{
			name:        "upload failure preserves accounting and continues in order",
			files:       []string{"a-examples.yaml", "b-openapi.yaml", "c-swagger.json"},
			wantOrder:   []string{"b-openapi.yaml", "c-swagger.json", "a-examples.yaml"},
			wantPrimary: []bool{true, true, false},
			failedFile:  "b-openapi.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range tt.files {
				path := filepath.Join(dir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
				require.NoError(t, os.WriteFile(path, []byte("fixture"), 0600))
			}
			client := &orderedImportClient{}
			var wantSuccess, wantFailed, wantErrors []string
			wantUploads := make([]string, 0, len(tt.wantOrder))
			for _, name := range tt.wantOrder {
				path := filepath.Join(dir, name)
				wantUploads = append(wantUploads, path)
				if name == tt.failedFile {
					client.FailedFiles = map[string]error{path: fmt.Errorf("upload failed")}
					wantFailed = append(wantFailed, path)
					wantErrors = append(wantErrors, fmt.Sprintf("error importing %s: upload failed", path))
				} else {
					wantSuccess = append(wantSuccess, path)
				}
			}
			result, err := ImportDirectory(client, &RealFileSystem{}, dir, tt.config)
			require.NoError(t, err)
			assert.Equal(t, wantUploads, client.Uploaded)
			assert.Equal(t, tt.wantPrimary, client.primary)
			assert.Equal(t, len(wantUploads), result.TotalFiles)
			assert.Equal(t, len(wantSuccess), result.SuccessCount)
			assert.Equal(t, len(wantFailed), result.FailedCount)
			assert.Equal(t, wantSuccess, result.SuccessFiles)
			assert.Equal(t, append([]string{}, wantFailed...), result.FailedFiles)
			assert.Equal(t, append([]string{}, wantErrors...), result.Errors)
		})
	}
}
