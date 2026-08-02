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
	"text/tabwriter"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newTestListCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		serviceID    string
		page         int
		size         int
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Microcks test results",
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

			tests, err := mc.ListTestResults(serviceID, page, size)
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				return output.WriteJSON(os.Stdout, tests)
			}
			return printTestResults(tests)
		},
	}
	cmd.Flags().StringVar(&serviceID, "serviceId", "", "Service id to filter tests")
	cmd.Flags().IntVar(&page, "page", 0, "Page index to fetch")
	cmd.Flags().IntVar(&size, "size", 50, "Number of test results to fetch")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return cmd
}

func newTestGetCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get <testResultId>",
		Short: "Get a Microcks test result",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}

			mc, _, err := newCommandClient(globalClientOpts)
			if err != nil {
				return err
			}

			result, err := mc.GetFullTestResult(args[0])
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				return output.WriteJSON(os.Stdout, result)
			}
			return printTestResults([]connectors.TestResultSummary{{
				ID:             result.ID,
				Version:        result.Version,
				TestNumber:     result.TestNumber,
				TestDate:       result.TestDate,
				TestedEndpoint: result.TestedEndpoint,
				ServiceID:      result.ServiceID,
				ElapsedTime:    result.ElapsedTime,
				Success:        result.Success,
				InProgress:     result.InProgress,
			}})
		},
	}
	cmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return cmd
}

func printTestResults(results []connectors.TestResultSummary) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()
	if _, err := fmt.Fprintln(w, "ID\tSERVICE ID\tSUCCESS\tIN PROGRESS\tELAPSED"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%dms\n", result.ID, result.ServiceID, result.Success, result.InProgress, result.ElapsedTime); err != nil {
			return err
		}
	}
	return nil
}
