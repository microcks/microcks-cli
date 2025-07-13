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
	"strings"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

// MicrocksClient interface for dependency injection
type MicrocksClient interface {
	UploadArtifact(file string, main bool) (string, error)
}

type FileType struct {
	Extension string
	IsPrimary bool
}

type ImportResult struct {
	TotalFiles   int
	SuccessCount int
	FailedCount  int
	SuccessFiles []string
	FailedFiles  []string
	Errors       []string
}

type ImportConfig struct {
	Recursive bool
	Pattern   string
	Verbose   bool
}

type FileSystem interface {
	Stat(path string) (os.FileInfo, error)
	Walk(root string, walkFn filepath.WalkFunc) error
	ReadDir(name string) ([]os.DirEntry, error)
}

type RealFileSystem struct{}

func (fs *RealFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (fs *RealFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (fs *RealFileSystem) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

var supportedExtensions = map[string]bool{
	".yaml": true,
	".yml":  true,
	".json": true,
	".xml":  true,
}

type ImportError struct {
	File string
	Err  error
}

type ValidationError struct {
	Message string
}

func (e ImportError) Error() string {
	return fmt.Sprintf("failed to import %s: %v", e.File, e.Err)
}

func (e ValidationError) Error() string {
	return e.Message
}

func NewImportDirCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		recursive bool
		pattern   string
		verbose   bool
	)

	var importDirCmd = &cobra.Command{
		Use:   "import-dir",
		Short: "Import API artifacts from a directory",
		Long: `Import API artifacts from a directory recursively.
		
		This command scans a directory for API specification files and imports them into Microcks.
		Supported file types: .yaml, .yml, .json, .xml

		Examples:
			microcks import-dir ./api-specs
			microcks import-dir ./api-specs --recursive
			microcks import-dir ./api-specs --pattern "*.yaml"
			microcks import-dir ./api-specs --recursive --pattern "openapi.*"`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("import-dir command requires a directory path")
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			dirPath := args[0]

			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
			if err != nil {
				fmt.Println(err)
				return
			}

			if localConfig == nil {
				fmt.Println("Please login to perform operation...")
				return
			}

			if globalClientOpts.Context == "" {
				globalClientOpts.Context = localConfig.CurrentContext
			}

			// Create client
			mc, err := connectors.NewClient(*globalClientOpts)
			if err != nil {
				fmt.Printf("error %v", err)
				return
			}

			// Set up business logic dependencies
			fs := &RealFileSystem{}
			importConfig := ImportConfig{
				Recursive: recursive,
				Pattern:   pattern,
				Verbose:   verbose,
			}

			// Execute business logic
			result, err := ImportDirectory(mc, fs, dirPath, importConfig)
			if err != nil {
				if validationErr, ok := err.(*ValidationError); ok {
					fmt.Println(validationErr.Message)
					return
				}
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Display results
			if verbose {
				fmt.Printf("Found %d specification files to import...\n", result.TotalFiles)
				for i, file := range result.SuccessFiles {
					fmt.Printf("[%d/%d] ✓ Imported: %s\n", i+1, result.TotalFiles, file)
				}
				for i, file := range result.FailedFiles {
					errorMsg := "Unknown error"
					if i < len(result.Errors) {
						errorMsg = result.Errors[i]
					}
					fmt.Printf("✗ Failed: %s - %s\n", file, errorMsg)
				}
			} else {
				for _, file := range result.SuccessFiles {
					fmt.Printf("✓ Imported: %s\n", file)
				}
				for i, file := range result.FailedFiles {
					errorMsg := "Unknown error"
					if i < len(result.Errors) {
						errorMsg = result.Errors[i]
					}
					fmt.Printf("✗ Failed: %s - %s\n", file, errorMsg)
				}
			}

			fmt.Printf("\nImport completed: %d/%d files imported successfully\n", result.SuccessCount, result.TotalFiles)
		},
	}

	importDirCmd.Flags().BoolVar(&recursive, "recursive", false, "Scan subdirectories recursively")
	importDirCmd.Flags().StringVar(&pattern, "pattern", "", "File pattern to match (e.g., '*.yaml', 'openapi.*')")
	importDirCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed progress")

	return importDirCmd
}

func ImportDirectory(client MicrocksClient, fs FileSystem, dirPath string, config ImportConfig) (ImportResult, error) {
	if err := validateDirectory(fs, dirPath); err != nil {
		return ImportResult{}, err
	}

	files, err := findSpecificationFiles(fs, dirPath, config.Recursive, config.Pattern)
	if err != nil {
		return ImportResult{}, fmt.Errorf("error scanning directory: %w", err)
	}

	if len(files) == 0 {
		return ImportResult{}, &ValidationError{Message: fmt.Sprintf("no specification files found in directory: %s", dirPath)}
	}

	result := ImportResult{
		TotalFiles:   len(files),
		SuccessFiles: make([]string, 0, len(files)),
		FailedFiles:  make([]string, 0, len(files)),
		Errors:       make([]string, 0, len(files)),
	}

	for _, file := range files {
		fileType := detectFileType(file)

		_, err := client.UploadArtifact(file, fileType.IsPrimary)
		if err != nil {
			result.FailedCount++
			result.FailedFiles = append(result.FailedFiles, file)
			result.Errors = append(result.Errors, fmt.Sprintf("error importing %s: %v", file, err))
			continue
		}

		result.SuccessCount++
		result.SuccessFiles = append(result.SuccessFiles, file)
	}

	return result, nil
}

// validateDirectory checks if the directory exists and is accessible
func validateDirectory(fs FileSystem, dirPath string) error {
	info, err := fs.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ValidationError{Message: fmt.Sprintf("directory does not exist: %s", dirPath)}
		}
		return fmt.Errorf("error accessing directory %s: %w", dirPath, err)
	}

	if !info.IsDir() {
		return &ValidationError{Message: fmt.Sprintf("path is not a directory: %s", dirPath)}
	}

	return nil
}

func findSpecificationFiles(fs FileSystem, dirPath string, recursive bool, pattern string) ([]string, error) {
	var files []string

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !recursive && path != dirPath {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExtensions[ext] {
			return nil
		}

		if pattern != "" {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		files = append(files, path)
		return nil
	}

	err := fs.Walk(dirPath, walkFunc)
	return files, err
}

func detectFileType(filePath string) FileType {
	fileName := strings.ToLower(filepath.Base(filePath))
	ext := filepath.Ext(filePath)

	// Default to primary for most files
	isPrimary := true

	if strings.Contains(fileName, "postman") || strings.Contains(fileName, "collection") {
		isPrimary = false
	}

	// If there's an OpenAPI file, prefer it as primary
	if strings.Contains(fileName, "openapi") || strings.Contains(fileName, "swagger") {
		isPrimary = true
	}

	return FileType{
		Extension: ext,
		IsPrimary: isPrimary,
	}
}
