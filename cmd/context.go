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
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewContextCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var delete bool
	var outputFormat string
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
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}
			configPath := globalClientOpts.ConfigPath
			localCfg, err := config.ReadLocalConfig(configPath)
			if err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}
			if delete {
				if len(args) == 0 {
					return errors.Wrapf(errors.KindUsage, "context --delete requires a CONTEXT argument")
				}
				if err := deleteContext(args[0], configPath); err != nil {
					return err
				}
				if outputFormat == "json" {
					return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, contextMutationResult{
						Name:   args[0],
						Action: "deleted",
					}))
				}
				_, err = fmt.Printf("Context '%s' deleted\n", args[0])
				return errors.Wrap(errors.KindEnvironment, err)
			}

			if len(args) == 0 {
				contexts, err := listMicrocksContexts(configPath)
				if err != nil {
					return err
				}
				if outputFormat == "json" {
					return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, contexts))
				}
				if len(contexts) == 0 {
					return errors.Wrapf(errors.KindUsage, "no contexts defined in %s", configPath)
				}
				return printMicrocksContexts(contexts)
			}

			ctxName := args[0]
			if localCfg == nil {
				return errors.Wrapf(errors.KindUsage, "no contexts defined in %s", configPath)
			}
			if localCfg.CurrentContext == ctxName {
				return writeContextSelection(outputFormat, localCfg, ctxName, "unchanged")
			}
			if _, err = localCfg.ResolveContext(ctxName); err != nil {
				return errors.Wrap(errors.KindNotFound, err)
			}
			localCfg.CurrentContext = ctxName
			if err := config.WriteLocalConfig(*localCfg, configPath); err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}
			return writeContextSelection(outputFormat, localCfg, ctxName, "selected")
		},
	}

	ctxCmd.Flags().BoolVarP(&delete, "delete", "d", false, "Delete a context")
	ctxCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")

	return ctxCmd
}

func deleteContext(context, configPath string) error {
	localCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}
	if localCfg == nil {
		return errors.Wrapf(errors.KindUsage, "nothing to delete")
	}
	contextIndex := slices.IndexFunc(localCfg.Contexts, func(ref config.ContextRef) bool {
		return ref.Name == context
	})
	if contextIndex < 0 {
		return errors.Wrapf(errors.KindNotFound, "context %q does not exist", context)
	}
	resolved, err := localCfg.ResolveContext(context)
	if err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}
	serverName, ok := localCfg.RemoveContext(context)
	if !ok {
		return errors.Wrapf(errors.KindAPI, "context %q disappeared while deleting it", context)
	}
	userStillReferenced := slices.ContainsFunc(localCfg.Contexts, func(ref config.ContextRef) bool {
		return ref.User == resolved.User.Name
	})
	if !userStillReferenced && !localCfg.RemoveUser(resolved.User.Name) {
		return errors.Wrapf(errors.KindAPI, "user %q referenced by context %q does not exist", resolved.User.Name, context)
	}
	serverStillReferenced := slices.ContainsFunc(localCfg.Contexts, func(ref config.ContextRef) bool {
		return ref.Server == serverName
	})
	if !serverStillReferenced && !localCfg.RemoveServer(serverName) {
		return errors.Wrapf(errors.KindAPI, "server %q referenced by context %q does not exist", serverName, context)
	}

	if localCfg.IsEmpty() {
		if err := localCfg.DeleteLocalConfig(configPath); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
	} else {
		if localCfg.CurrentContext == context {
			localCfg.CurrentContext = ""
		}
		if err := config.ValidateLocalConfig(*localCfg); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
		if err := config.WriteLocalConfig(*localCfg, configPath); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
	}
	return nil
}

type contextSummary struct {
	Name    string `json:"name"`
	Server  string `json:"server"`
	Current bool   `json:"current"`
}

type contextMutationResult struct {
	Name   string `json:"name"`
	Server string `json:"server,omitempty"`
	Action string `json:"action"`
}

func listMicrocksContexts(configPath string) ([]contextSummary, error) {
	localCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return nil, errors.Wrap(errors.KindEnvironment, err)
	}
	if localCfg == nil {
		return []contextSummary{}, nil
	}
	contexts := make([]contextSummary, 0, len(localCfg.Contexts))
	for _, contextRef := range localCfg.Contexts {
		resolved, err := localCfg.ResolveContext(contextRef.Name)
		if err != nil {
			return nil, errors.Wrap(errors.KindEnvironment, fmt.Errorf("resolving context %q: %w", contextRef.Name, err))
		}
		contexts = append(contexts, contextSummary{
			Name:    resolved.Name,
			Server:  resolved.Server.Server,
			Current: localCfg.CurrentContext == resolved.Name,
		})
	}
	return contexts, nil
}

func printMicrocksContexts(contexts []contextSummary) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	columnNames := []string{"CURRENT", "NAME", "SERVER"}
	if _, err := fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}
	for _, context := range contexts {
		prefix := " "
		if context.Current {
			prefix = "*"
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", prefix, context.Name, context.Server); err != nil {
			return errors.Wrap(errors.KindEnvironment, err)
		}
	}
	return errors.Wrap(errors.KindEnvironment, w.Flush())
}

func writeContextSelection(outputFormat string, localCfg *config.LocalConfig, name, action string) error {
	resolved, err := localCfg.ResolveContext(name)
	if err != nil {
		return errors.Wrap(errors.KindNotFound, err)
	}
	if outputFormat == "json" {
		return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, contextMutationResult{
			Name:   resolved.Name,
			Server: resolved.Server.Server,
			Action: action,
		}))
	}
	if action == "unchanged" {
		_, err = fmt.Printf("Already at context '%s'\n", name)
	} else {
		_, err = fmt.Printf("Switched to context '%s'\n", name)
	}
	return errors.Wrap(errors.KindEnvironment, err)
}
