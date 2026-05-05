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
			errors.CheckError(err)

			if localConfig == nil {
				fmt.Println("Config not found, nothing to stop")
				return nil
			}

			ctx, err := localConfig.ResolveContext("")
			errors.CheckError(err)
			instance := ctx.Instance

			if instance.Name == "" {
				fmt.Println("No instance is associated with this context")
				return nil
			}

			containerClient, err := connectors.NewContainerClient(instance.Driver)
			errors.CheckError(err)
			defer containerClient.CloseClient()

			err = containerClient.StopContainer(instance.ContainerID)
			if err != nil {
				return fmt.Errorf("failed to stop a container: %w", err)
			}
			fmt.Println("")
			log.Printf("Instance %s stopped successfully", instance.Name)

			if instance.AutoRemove {
				// Update config after removal.
				_, ok := localConfig.RemoveContext(ctx.Name)
				if !ok {
					return fmt.Errorf("context %s does not exist", ctx.Name)
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
			err = config.WriteLocalConfig(*localConfig, configFile)
			errors.CheckError(err)
			return nil
		},
	}

	return stopCmd
}
