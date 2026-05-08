package cmd

import (
	"fmt"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewStopCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop microcks instance",
		Long:  "stop microcks instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := globalClientOpts.ConfigPath
			localConfig, err := config.ReadLocalConfig(configFile)
			errors.CheckError(err)

			if localConfig == nil {
				fmt.Println("Config not found, nothing to stop")
				return nil
			}

			ctx, err := localConfig.ResolveContext("")
			errors.CheckError(err)
			instance := ctx.Instance

			if instance.Name == "" {
				return fmt.Errorf("no instance is associated with this context")
			}

			containerClient, err := connectors.NewContainerClient(instance.Driver)
			errors.CheckError(err)
			defer containerClient.CloseClient()

			if err := containerClient.StopContainer(instance.ContainerID); err != nil {
				return fmt.Errorf("failed to stop container: %w", err)
			}

			fmt.Printf("Instance %s stopped successfully\n", instance.Name)

			if instance.AutoRemove {
				_, ok := localConfig.RemoveContext(ctx.Name)
				if !ok {
					return fmt.Errorf("context %s does not exist", ctx.Name)
				}
				_ = localConfig.RemoveServer(ctx.Server.Server)
				_ = localConfig.RemoveUser(ctx.User.Name)
				_ = localConfig.RemoveAuth(ctx.Server.Server)
				_ = localConfig.RemoveInstance(instance.Name)

				localConfig.CurrentContext = ""
				fmt.Printf("Instance %s removed successfully\n", instance.Name)
			} else {
				instance.Status = "Exited"
				localConfig.UpsertInstance(instance)
				fmt.Printf("Instance %s status updated to Exited\n", instance.Name)
			}

			if err := config.WriteLocalConfig(*localConfig, configFile); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}
			return nil
		},
	}

	return stopCmd
}
