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
	"strings"
	"text/tabwriter"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewServiceCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "List and inspect Microcks services",
	}

	serviceCmd.AddCommand(newServiceListCommand(globalClientOpts))
	serviceCmd.AddCommand(newServiceGetCommand(globalClientOpts))

	return serviceCmd
}

func newServiceListCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		page         int
		size         int
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Microcks services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}
			if page < 0 {
				return errors.Wrapf(errors.KindUsage, "--page must be greater than or equal to 0")
			}
			if size <= 0 {
				return errors.Wrapf(errors.KindUsage, "--size must be greater than 0")
			}

			mc, _, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}

			services, err := mc.ListServices(page, size)
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				return output.WriteJSON(os.Stdout, services)
			}
			return printServices(services)
		},
	}
	cmd.Flags().IntVar(&page, "page", 0, "Page index to fetch")
	cmd.Flags().IntVar(&size, "size", 50, "Number of services to fetch")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return cmd
}

func newServiceGetCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get <id-or-name-version>",
		Short: "Get Microcks service details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}

			mc, _, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}

			service, err := mc.GetService(args[0])
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				return output.WriteJSON(os.Stdout, service)
			}
			return printServiceDetail(service)
		},
	}
	cmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return cmd
}

func printServices(services []connectors.Service) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()
	if _, err := fmt.Fprintln(w, "ID\tNAME\tVERSION\tTYPE"); err != nil {
		return err
	}
	for _, service := range services {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", service.ID, service.Name, service.Version, service.Type); err != nil {
			return err
		}
	}
	return nil
}

func printServiceDetail(detail *connectors.ServiceDetail) error {
	service := detail.Service
	if _, err := fmt.Printf("%s:%s %s\n", service.Name, service.Version, service.Type); err != nil {
		return err
	}
	if len(service.Operations) == 0 {
		return nil
	}
	for _, operation := range service.Operations {
		parts := []string{operation.Name}
		if operation.Method != "" {
			parts = append(parts, operation.Method)
		}
		if len(operation.ResourcePaths) > 0 {
			parts = append(parts, strings.Join(operation.ResourcePaths, ","))
		}
		if _, err := fmt.Printf("- %s\n", strings.Join(parts, " ")); err != nil {
			return err
		}
	}
	return nil
}
