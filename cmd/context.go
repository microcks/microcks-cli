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

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewContextCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var delete bool
	ctxCmd := &cobra.Command{
		Use:     "context [CONTEXT]",
		Aliases: []string{"ctx"},
		Short:   "switch between contexts",
		Example: `# List Microcks context
microcks context

# Switch Microcks context
microcks context http://localhost:8080 

# Delete Microcks context
microcks context http://localhost:8080 --delete`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := globalClientOpts.ConfigPath
			localCfg, err := config.ReadLocalConfig(configPath)
			if err != nil {
				return err
			}
			if delete {
				if len(args) == 0 {
					return errors.Wrapf(errors.KindUsage, "context --delete requires a CONTEXT argument")
				}
				return deleteContext(args[0], configPath)
			}

			if len(args) == 0 {
				return printMicrocksContexts(configPath)
			}

			ctxName := args[0]
			if localCfg == nil {
				return errors.Wrapf(errors.KindUsage, "no contexts defined in %s", configPath)
			}
			if localCfg.CurrentContext == ctxName {
				fmt.Printf("Already at context '%s'\n", localCfg.CurrentContext)
				return nil
			}
			if _, err = localCfg.ResolveContext(ctxName); err != nil {
				return errors.Wrap(errors.KindNotFound, err)
			}
			localCfg.CurrentContext = ctxName
			if err := config.WriteLocalConfig(*localCfg, configPath); err != nil {
				return err
			}
			fmt.Printf("Switched to context '%s'\n", localCfg.CurrentContext)
			return nil
		},
	}

	ctxCmd.Flags().BoolVarP(&delete, "delete", "d", false, "Delete a context")

	return ctxCmd
}

func deleteContext(context, configPath string) error {
	localCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return err
	}
	if localCfg == nil {
		return errors.Wrapf(errors.KindUsage, "nothing to delete")
	}
	serverName, ok := localCfg.RemoveContext(context)
	if !ok {
		return errors.Wrapf(errors.KindNotFound, "context %q does not exist", context)
	}
	localCfg.RemoveUser(context)
	localCfg.RemoveServer(serverName)

	if localCfg.IsEmpty() {
		if err := localCfg.DeleteLocalConfig(configPath); err != nil {
			return err
		}
	} else {
		if localCfg.CurrentContext == context {
			localCfg.CurrentContext = ""
		}
		if err := config.ValidateLocalConfig(*localCfg); err != nil {
			return err
		}
		if err := config.WriteLocalConfig(*localCfg, configPath); err != nil {
			return err
		}
	}
	fmt.Printf("Context '%s' deleted\n", context)
	return nil
}

func printMicrocksContexts(configPath string) error {
	localCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return err
	}
	if localCfg == nil {
		return errors.Wrapf(errors.KindUsage, "no contexts defined in %s", configPath)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	columnNames := []string{"CURRENT", "NAME", "SERVER"}
	if _, err = fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("writing contexts output: %w", err))
	}

	for _, contextRef := range localCfg.Contexts {
		context, err := localCfg.ResolveContext(contextRef.Name)
		if err != nil {
			return errors.Wrap(errors.KindUsage, fmt.Errorf("context %q is invalid: %w", contextRef.Name, err))
		}
		prefix := " "
		if localCfg.CurrentContext == context.Name {
			prefix = "*"
		}
		if _, err = fmt.Fprintf(w, "%s\t%s\t%s\n", prefix, context.Name, context.Server.Server); err != nil {
			return errors.Wrap(errors.KindEnvironment, fmt.Errorf("writing contexts output: %w", err))
		}
	}
	if err := w.Flush(); err != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("writing contexts output: %w", err))
	}
	return nil
}
