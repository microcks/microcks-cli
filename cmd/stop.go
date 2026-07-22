package cmd

import (
	"fmt"
	"log"

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
			if err != nil {
				return err
			}

			if localConfig == nil {
				fmt.Println("Config not found, nothing to stop")
				return nil
			}

			ctx, err := localConfig.ResolveContext("")
			if err != nil {
				return err
			}
			instance := ctx.Instance

			if instance.Name == "" {
				fmt.Println("No instance is associated with this context")
				return nil
			}

			containerClient, err := connectors.NewContainerClient(instance.Driver)
			if err != nil {
				return errors.Wrap(errors.KindEnvironment, err)
			}
			defer containerClient.CloseClient()

			if err := containerClient.StopContainer(instance.ContainerID); err != nil {
				return errors.Wrap(errors.KindEnvironment, fmt.Errorf("failed to stop container: %w", err))
			}
			fmt.Println("")
			log.Printf("Instance %s stopped successfully", instance.Name)

			// update configs

			if instance.AutoRemove {
				_, ok := localConfig.RemoveContext(ctx.Name)
				if !ok {
					return errors.Wrapf(errors.KindNotFound, "context %q does not exist", ctx.Name)
				}
				_ = localConfig.RemoveServer(ctx.Server.Server)
				_ = localConfig.RemoveUser(ctx.User.Name)
				_ = localConfig.RemoveAuth(ctx.Server.Server)
				_ = localConfig.RemoveInstance(instance.Name)

				localConfig.CurrentContext = ""
				log.Printf("Instance %s removed successfully", instance.Name)
			} else {
				instance.Status = "Exited"
				localConfig.UpsertInstance(instance)
				log.Printf("Instance %s status updated to Exited", instance.Name)
			}
			return config.WriteLocalConfig(*localConfig, configFile)
		},
	}

	return stopCmd
}
