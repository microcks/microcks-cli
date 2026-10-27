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

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewExportCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		outputDir string
		snapshot  bool
		all       bool
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:   "export [serviceRef]",
		Short: "Export API artifacts and snapshots from Microcks server",
		Long:  `Export specification artifacts or complete repository snapshots from a Microcks server to the local filesystem.`,
		Example: `# Export artifacts for a specific service by name and version
microcks export "Petstore API:1.0" --output ./specs/

# Export artifacts for a service by service ID
microcks export 3f67825d-4f1b-4d6d-88b1-3e4b78619e05 --output ./specs/

# Export a service as a snapshot JSON bundle
microcks export "Petstore API:1.0" --snapshot --output ./specs/

# Export the entire Microcks repository snapshot
microcks export --all --output ./backup/`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !all {
				return errors.Wrapf(errors.KindUsage, "missing required argument: <apiName:apiVersion> (e.g. 'Petstore API:1.0'), or specify --all")
			}

			if outputDir == "" {
				outputDir = "."
			}

			mc, _, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}

			// If --all is passed, export full snapshot
			if all {
				return exportSnapshot(mc, nil, outputDir, overwrite)
			}

			serviceRef := args[0]
			if strings.TrimSpace(serviceRef) == "" {
				return errors.Wrapf(errors.KindUsage, "service reference cannot be empty")
			}

			// If snapshot flag is explicitly set, export service snapshot
			if snapshot {
				return exportServiceSnapshot(mc, serviceRef, outputDir, overwrite)
			}

			// Otherwise, export specification artifacts
			return exportServiceArtifacts(mc, serviceRef, outputDir, overwrite)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory or file path for exported artifacts")
	cmd.Flags().BoolVar(&snapshot, "snapshot", false, "Export as Microcks snapshot JSON format")
	cmd.Flags().BoolVar(&all, "all", false, "Export entire repository snapshot")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files on disk")

	return cmd
}

func exportSnapshot(mc connectors.MicrocksClient, serviceIDs []string, destPath string, overwrite bool) error {
	data, err := mc.ExportSnapshot(serviceIDs)
	if err != nil {
		return err
	}

	targetFile := destPath
	stat, err := os.Stat(destPath)
	if (err == nil && stat.IsDir()) || strings.HasSuffix(destPath, "/") || strings.HasSuffix(destPath, "\\") {
		targetFile = filepath.Join(destPath, "microcks-repository.json")
	}

	if err := writeFileSafely(targetFile, data, overwrite); err != nil {
		return err
	}

	fmt.Printf("Successfully exported snapshot to %s (%d bytes)\n", targetFile, len(data))
	return nil
}

func exportServiceSnapshot(mc connectors.MicrocksClient, serviceRef string, destPath string, overwrite bool) error {
	serviceDetail, err := mc.GetService(serviceRef)
	if err != nil {
		return err
	}

	data, err := mc.ExportSnapshot([]string{serviceDetail.Service.ID})
	if err != nil {
		return err
	}

	targetFile := destPath
	stat, err := os.Stat(destPath)
	if (err == nil && stat.IsDir()) || strings.HasSuffix(destPath, "/") || strings.HasSuffix(destPath, "\\") {
		cleanName := sanitizeFilename(fmt.Sprintf("%s-%s-snapshot.json", serviceDetail.Service.Name, serviceDetail.Service.Version))
		targetFile = filepath.Join(destPath, cleanName)
	}

	if err := writeFileSafely(targetFile, data, overwrite); err != nil {
		return err
	}

	fmt.Printf("Successfully exported snapshot for '%s:%s' to %s (%d bytes)\n",
		serviceDetail.Service.Name, serviceDetail.Service.Version, targetFile, len(data))
	return nil
}

func exportServiceArtifacts(mc connectors.MicrocksClient, serviceRef string, destPath string, overwrite bool) error {
	serviceDetail, err := mc.GetService(serviceRef)
	if err != nil {
		return err
	}

	resources, err := mc.GetResources(serviceDetail.Service.ID)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		fmt.Printf("No specification resources found for '%s:%s', falling back to snapshot export...\n",
			serviceDetail.Service.Name, serviceDetail.Service.Version)
		return exportServiceSnapshot(mc, serviceRef, destPath, overwrite)
	}

	destDir := destPath
	stat, err := os.Stat(destPath)
	isFileTarget := false
	if (err == nil && !stat.IsDir()) || (!strings.HasSuffix(destPath, "/") && !strings.HasSuffix(destPath, "\\") && filepath.Ext(destPath) != "") {
		if len(resources) == 1 {
			isFileTarget = true
		} else {
			destDir = filepath.Dir(destPath)
		}
	}

	if !isFileTarget {
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return errors.Wrap(errors.KindEnvironment, fmt.Errorf("creating output directory %q: %w", destDir, err))
		}
	}

	for _, resource := range resources {
		var targetFile string
		if isFileTarget {
			targetFile = destPath
		} else {
			filename := resource.Name
			if filename == "" {
				ext := extensionForResourceType(resource.Type)
				filename = sanitizeFilename(fmt.Sprintf("%s-%s%s", serviceDetail.Service.Name, serviceDetail.Service.Version, ext))
			}
			targetFile = filepath.Join(destDir, filename)
		}

		if err := writeFileSafely(targetFile, []byte(resource.Content), overwrite); err != nil {
			return err
		}
		fmt.Printf("Exported %s artifact: %s (%d bytes)\n", resource.Type, targetFile, len(resource.Content))
	}

	return nil
}

func writeFileSafely(targetFile string, data []byte, overwrite bool) error {
	dir := filepath.Dir(targetFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("creating directory %q: %w", dir, err))
	}

	if !overwrite {
		if _, err := os.Stat(targetFile); err == nil {
			return errors.Wrapf(errors.KindUsage, "target file %q already exists (use --overwrite to replace)", targetFile)
		}
	}

	// #nosec G306 -- artifact file exported explicitly by user
	if err := os.WriteFile(targetFile, data, 0o600); err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("writing file %q: %w", targetFile, err))
	}
	return nil
}

func sanitizeFilename(name string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "-")
	}
	return result
}

func extensionForResourceType(resourceType string) string {
	switch strings.ToUpper(resourceType) {
	case "OPEN_API_SPEC":
		return ".yaml"
	case "ASYNC_API_SPEC":
		return ".yaml"
	case "POSTMAN_COLLECTION":
		return ".json"
	case "PROTOBUF_DESCRIPTOR":
		return ".proto"
	case "GRAPHQL_SCHEMA":
		return ".graphql"
	case "WSDL":
		return ".wsdl"
	default:
		return ".json"
	}
}
