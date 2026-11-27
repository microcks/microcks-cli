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
	"io"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/output"
	"github.com/spf13/cobra"
)

type instanceStopResult struct {
	Name    string `json:"name,omitempty"`
	Context string `json:"context,omitempty"`
	Status  string `json:"status"`
	Removed bool   `json:"removed,omitempty"`
}

func NewStopCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var outputFormat string

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop microcks instance",
		Long:  "stop microcks instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !output.IsTextOrJSON(outputFormat) {
				return errors.Wrapf(errors.KindUsage, "--output must be one of: text, json")
			}

			return runStop(globalClientOpts.ConfigPath, globalClientOpts.Context, outputFormat, connectors.NewContainerClient)
		},
	}

	stopCmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text or json")
	return stopCmd
}

func runStop(configFile, contextName, outputFormat string, newContainerClient func(string) (connectors.ContainerClient, error)) error {
	localConfig, err := config.ReadLocalConfig(configFile)
	if err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}

	if localConfig == nil {
		return writeStopResult(outputFormat, instanceStopResult{
			Status: "not_found",
		}, "Config not found, nothing to stop")
	}

	ctx, err := localConfig.ResolveContext(contextName)
	if err != nil {
		return errors.Wrap(errors.KindNotFound, err)
	}
	instance := ctx.Instance

	if instance.Name == "" {
		return writeStopResult(outputFormat, instanceStopResult{
			Context: ctx.Name,
			Status:  "not_found",
		}, "No instance is associated with this context")
	}

	containerClient, err := newContainerClient(instance.Driver)
	if err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}

	exists, err := containerClient.ContainerExists(instance.ContainerID)
	if err != nil {
		if closeErr := containerClient.CloseClient(); closeErr != nil {
			return errors.Wrapf(errors.KindEnvironment, "checking container existence: %v; closing container client: %v", err, closeErr)
		}
		return errors.Wrap(errors.KindEnvironment, err)
	}

	status := "stopped"
	var textMsg string
	if !exists {
		status = "missing"
		textMsg = fmt.Sprintf("Container for instance %s no longer exists\nRun 'microcks start --name %s' to bring the container back", instance.Name, instance.Name)
	} else {
		progress := progressWriter(outputFormat)
		var stopErr error
		if stopper, ok := containerClient.(interface {
			StopContainerWithOutput(string, io.Writer) error
		}); ok {
			stopErr = stopper.StopContainerWithOutput(instance.ContainerID, progress)
		} else {
			stopErr = containerClient.StopContainer(instance.ContainerID)
		}
		if stopErr != nil {
			if closeErr := containerClient.CloseClient(); closeErr != nil {
				return errors.Wrapf(errors.KindEnvironment, "failed to stop container: %v; closing container client: %v", stopErr, closeErr)
			}
			return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to stop container: %w", stopErr))
		}
		textMsg = fmt.Sprintf("Instance %s stopped successfully", instance.Name)
	}

	if closeErr := containerClient.CloseClient(); closeErr != nil {
		return errors.Wrap(errors.KindEnvironment, fmt.Errorf("closing container client: %w", closeErr))
	}

	if instance.AutoRemove {
		_, ok := localConfig.RemoveContext(ctx.Name)
		if !ok {
			return errors.Wrapf(errors.KindNotFound, "context %q does not exist", ctx.Name)
		}
		localConfig.RemoveServer(ctx.Server.Server)
		localConfig.RemoveUser(ctx.User.Name)
		localConfig.RemoveAuth(ctx.Server.Server)
		localConfig.RemoveInstance(instance.ContainerID)

		if localConfig.CurrentContext == ctx.Name {
			localConfig.CurrentContext = ""
		}
		status = "removed"
		if exists {
			textMsg = fmt.Sprintf("%s\nInstance %s removed successfully", textMsg, instance.Name)
		} else {
			textMsg = fmt.Sprintf("Container for instance %s no longer exists\nInstance %s removed successfully", instance.Name, instance.Name)
		}
	} else {
		instance.Status = "Exited"
		localConfig.UpsertInstance(instance)
		textMsg = fmt.Sprintf("%s\nInstance %s status updated to Exited", textMsg, instance.Name)
	}

	if err := config.WriteLocalConfig(*localConfig, configFile); err != nil {
		return errors.Wrap(errors.KindEnvironment, err)
	}

	return writeStopResult(outputFormat, instanceStopResult{
		Name:    instance.Name,
		Context: ctx.Name,
		Status:  status,
		Removed: instance.AutoRemove,
	}, textMsg)
}

func writeStopResult(outputFormat string, result instanceStopResult, textMsg string) error {
	if outputFormat == "json" {
		return errors.Wrap(errors.KindEnvironment, output.WriteJSON(os.Stdout, result))
	}
	_, err := fmt.Fprintln(os.Stdout, textMsg)
	return errors.Wrap(errors.KindEnvironment, err)
}
