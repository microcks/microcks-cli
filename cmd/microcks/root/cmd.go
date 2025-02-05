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
package root 

import "github.com/spf13/cobra"


func RootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "microcks",
		Short: "Microcks CLI",
		Long:  `Microcks CLI is a command line interface for Microcks.`,
		Example: `
  # Start the Microcks CLI
  microcks`,
  	}

	root.AddCommand(
		importCmd,
		testCmd,
	)
	

	return root

}