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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockMicrocksClient struct {
	Uploaded    []string
	FailedFiles map[string]error
	UploadCalls int
}

func (m *MockMicrocksClient) UploadArtifact(file string, main bool) (string, error) {
	m.UploadCalls++
	m.Uploaded = append(m.Uploaded, file)

	if err, exists := m.FailedFiles[file]; exists {
		return "", err
	}
	return fmt.Sprintf("mock-id-%d", m.UploadCalls), nil
}

type MockFileSystem struct {
	Files      map[string]bool // path -> isDir
	StatErrors map[string]error
	WalkErrors map[string]error
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if err, exists := m.StatErrors[path]; exists {
		return nil, err
	}

	isDir, exists := m.Files[path]
	if !exists {
		return nil, os.ErrNotExist
	}

	return &MockFileInfo{name: filepath.Base(path), isDir: isDir}, nil
}

func (m *MockFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	if err, exists := m.WalkErrors[root]; exists {
		return err
	}

	for path, isDir := range m.Files {
		if path == root || strings.HasPrefix(path, root+string(filepath.Separator)) {
			info := &MockFileInfo{name: filepath.Base(path), isDir: isDir}
			if err := walkFn(path, info, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockFileSystem) ReadDir(name string) ([]os.DirEntry, error) {
	// Not used in our current implementation, but required by interface
	return nil, nil
}

type MockFileInfo struct {
	name  string
	isDir bool
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }

func TestImportDirectory(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]bool // path -> isDir
		failedFiles    map[string]error
		config         ImportConfig
		expectedResult ImportResult
		expectError    bool
	}{
		{
			name: "successful import of all files",
			files: map[string]bool{
				"/test":              true, // Directory must exist
				"/test/openapi.yaml": false,
				"/test/postman.json": false,
			},
			config: ImportConfig{Recursive: false, Pattern: ""},
			expectedResult: ImportResult{
				TotalFiles:   2,
				SuccessCount: 2,
				FailedCount:  0,
				SuccessFiles: []string{"/test/openapi.yaml", "/test/postman.json"},
				FailedFiles:  []string{},
				Errors:       []string{},
			},
		},
		{
			name: "partial failure",
			files: map[string]bool{
				"/test":              true, // Directory must exist
				"/test/openapi.yaml": false,
				"/test/postman.json": false,
			},
			failedFiles: map[string]error{
				"/test/postman.json": fmt.Errorf("upload failed"),
			},
			config: ImportConfig{Recursive: false, Pattern: ""},
			expectedResult: ImportResult{
				TotalFiles:   2,
				SuccessCount: 1,
				FailedCount:  1,
				SuccessFiles: []string{"/test/openapi.yaml"},
				FailedFiles:  []string{"/test/postman.json"},
				Errors:       []string{"error importing /test/postman.json: upload failed"},
			},
		},
		{
			name: "recursive scan",
			files: map[string]bool{
				"/test":                 true, // Directory must exist
				"/test/openapi.yaml":    false,
				"/test/subdir/spec.yml": false,
				"/test/subdir":          true,
			},
			config: ImportConfig{Recursive: true, Pattern: ""},
			expectedResult: ImportResult{
				TotalFiles:   2,
				SuccessCount: 2,
				FailedCount:  0,
				SuccessFiles: []string{"/test/openapi.yaml", "/test/subdir/spec.yml"},
				FailedFiles:  []string{},
				Errors:       []string{},
			},
		},
		{
			name: "pattern filtering",
			files: map[string]bool{
				"/test":              true, // Directory must exist
				"/test/openapi.yaml": false,
				"/test/postman.json": false,
			},
			config: ImportConfig{Recursive: false, Pattern: "*.yaml"},
			expectedResult: ImportResult{
				TotalFiles:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SuccessFiles: []string{"/test/openapi.yaml"},
				FailedFiles:  []string{},
				Errors:       []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := &MockMicrocksClient{
				FailedFiles: tt.failedFiles,
			}
			mockFS := &MockFileSystem{
				Files: tt.files,
			}

			// Execute
			result, err := ImportDirectory(mockClient, mockFS, "/test", tt.config)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.TotalFiles, result.TotalFiles)
			assert.Equal(t, tt.expectedResult.SuccessCount, result.SuccessCount)
			assert.Equal(t, tt.expectedResult.FailedCount, result.FailedCount)
			assert.ElementsMatch(t, tt.expectedResult.SuccessFiles, result.SuccessFiles)
			assert.ElementsMatch(t, tt.expectedResult.FailedFiles, result.FailedFiles)
		})
	}
}

// TestValidateDirectory tests directory validation
func TestValidateDirectory(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]bool
		statErrors  map[string]error
		expectError bool
		errorType   string
	}{
		{
			name: "valid directory",
			files: map[string]bool{
				"/test": true,
			},
			expectError: false,
		},
		{
			name: "directory does not exist",
			statErrors: map[string]error{
				"/test": os.ErrNotExist,
			},
			expectError: true,
			errorType:   "ValidationError",
		},
		{
			name: "path is not a directory",
			files: map[string]bool{
				"/test": false,
			},
			expectError: true,
			errorType:   "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &MockFileSystem{
				Files:      tt.files,
				StatErrors: tt.statErrors,
			}

			err := validateDirectory(mockFS, "/test")

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType == "ValidationError" {
					_, ok := err.(*ValidationError)
					assert.True(t, ok, "Expected ValidationError")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFindSpecificationFilesWithFS tests file discovery with mock filesystem
func TestFindSpecificationFilesWithFS(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]bool
		recursive bool
		pattern   string
		expected  []string
	}{
		{
			name: "non-recursive scan",
			files: map[string]bool{
				"/test":              true, // Directory must exist
				"/test/openapi.yaml": false,
				"/test/postman.json": false,
				"/test/ignore.txt":   false,
			},
			recursive: false,
			expected:  []string{"/test/openapi.yaml", "/test/postman.json"},
		},
		{
			name: "recursive scan",
			files: map[string]bool{
				"/test":                 true, // Directory must exist
				"/test/openapi.yaml":    false,
				"/test/subdir":          true,
				"/test/subdir/spec.yml": false,
			},
			recursive: true,
			expected:  []string{"/test/openapi.yaml", "/test/subdir/spec.yml"},
		},
		{
			name: "pattern filtering",
			files: map[string]bool{
				"/test":              true, // Directory must exist
				"/test/openapi.yaml": false,
				"/test/postman.json": false,
			},
			recursive: false,
			pattern:   "*.yaml",
			expected:  []string{"/test/openapi.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &MockFileSystem{Files: tt.files}

			files, err := findSpecificationFiles(mockFS, "/test", tt.recursive, tt.pattern)

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, files)
		})
	}
}

// TestDetectFileTypeWithLogic tests file type detection
func TestDetectFileTypeWithLogic(t *testing.T) {
	tests := []struct {
		filePath string
		expected FileType
	}{
		{
			filePath: "openapi.yaml",
			expected: FileType{Extension: ".yaml", IsPrimary: true},
		},
		{
			filePath: "swagger.json",
			expected: FileType{Extension: ".json", IsPrimary: true},
		},
		{
			filePath: "postman-collection.json",
			expected: FileType{Extension: ".json", IsPrimary: false},
		},
		{
			filePath: "my-collection.json",
			expected: FileType{Extension: ".json", IsPrimary: false},
		},
		{
			filePath: "api-spec.yml",
			expected: FileType{Extension: ".yml", IsPrimary: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := detectFileType(tt.filePath)
			assert.Equal(t, tt.expected.Extension, result.Extension)
			assert.Equal(t, tt.expected.IsPrimary, result.IsPrimary)
		})
	}
}

// TestNewImportDirCommand tests command creation
func TestNewImportDirCommand(t *testing.T) {
	clientOpts := &connectors.ClientOptions{}
	cmd := NewImportDirCommand(clientOpts)

	// Test command properties
	assert.Equal(t, "import-dir", cmd.Use)
	assert.Equal(t, "Import API artifacts from a directory", cmd.Short)
	assert.Contains(t, cmd.Long, "Supported file types")

	// Test that flags are properly defined
	recursiveFlag := cmd.Flags().Lookup("recursive")
	assert.NotNil(t, recursiveFlag)
	assert.Equal(t, "bool", recursiveFlag.Value.Type())

	patternFlag := cmd.Flags().Lookup("pattern")
	assert.NotNil(t, patternFlag)
	assert.Equal(t, "string", patternFlag.Value.Type())

	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "bool", verboseFlag.Value.Type())

	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "string", outputFlag.Value.Type())
	assert.Equal(t, "text", outputFlag.DefValue)
}

func TestImportDirCommandRejectsUnsupportedOutput(t *testing.T) {
	_, err := executeCLIForTest(t, "import-dir", t.TempDir(), "--output", "yaml")
	require.Error(t, err)
	assert.Equal(t, errors.KindUsage, errors.KindOf(err))
}

func TestImportDirCommandOutputsJSON(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	postmanPath := filepath.Join(dir, "postman.json")
	require.NoError(t, os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0o600))
	require.NoError(t, os.WriteFile(postmanPath, []byte("{}"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/artifact/upload", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		_, writeErr := w.Write([]byte("Petstore:1.0.0"))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "import-dir", dir, "--microcksURL", server.URL, "--output", "json")
	require.NoError(t, err)

	var result ImportResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.TotalFiles)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.Equal(t, []string{specPath, postmanPath}, result.SuccessFiles)
	assert.Empty(t, result.FailedFiles)
	assert.Empty(t, result.Errors)
}

func TestImportDirCommandOutputsJSONOnPartialFailure(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	postmanPath := filepath.Join(dir, "postman.json")
	require.NoError(t, os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0o600))
	require.NoError(t, os.WriteFile(postmanPath, []byte("{}"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(1<<20))
		if r.FormValue("mainArtifact") == "false" {
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr := w.Write([]byte("upload failed"))
			require.NoError(t, writeErr)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, writeErr := w.Write([]byte("Petstore:1.0.0"))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	out, err := executeCLIForTest(t, "import-dir", dir, "--microcksURL", server.URL, "--output", "json")
	require.Error(t, err)
	assert.Equal(t, errors.KindAPI, errors.KindOf(err))

	var result ImportResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 2, result.TotalFiles)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Equal(t, []string{specPath}, result.SuccessFiles)
	assert.Equal(t, []string{postmanPath}, result.FailedFiles)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], postmanPath)
}

// TestImportResult tests the ImportResult struct
func TestImportResult(t *testing.T) {
	result := ImportResult{
		TotalFiles:   5,
		SuccessCount: 3,
		FailedCount:  2,
		SuccessFiles: []string{"a.yaml", "b.json", "c.yml"},
		FailedFiles:  []string{"d.txt", "e.pdf"},
		Errors:       []string{"error 1", "error 2"},
	}

	assert.Equal(t, 5, result.TotalFiles)
	assert.Equal(t, 3, result.SuccessCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Len(t, result.SuccessFiles, 3)
	assert.Len(t, result.FailedFiles, 2)
	assert.Len(t, result.Errors, 2)
}

// Benchmark tests for performance
func BenchmarkImportDirectory(b *testing.B) {
	// Create a mock with many files
	mockClient := &MockMicrocksClient{}
	mockFS := &MockFileSystem{
		Files: make(map[string]bool),
	}

	// Add 100 test files
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/test/spec_%d.yaml", i)
		mockFS.Files[path] = false
	}

	config := ImportConfig{Recursive: false, Pattern: ""}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ImportDirectory(mockClient, mockFS, "/test", config)
		require.NoError(b, err)
	}
}
