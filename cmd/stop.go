package cmd

import (
	"fmt"
	"log"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop microcks instance",
		Long:  "stop microcks instance",
		Run: func(cmd *cobra.Command, args []string) {

			configFile, err := config.DefaultLocalConfigPath()
			errors.CheckError(err)
			localConfig, err := config.ReadLocalConfig(configFile)
			errors.CheckError(err)

			if localConfig == nil {
				fmt.Println("Config not found, nothing to stop")
				return
			}

			ctx, err := localConfig.ResolveContext("")
			errors.CheckError(err)
			instance := ctx.Instance

			if instance.Name == "" {
				fmt.Println("No instance is associated with this context")
				return
			}

			containerClient, err := connectors.NewContainerClient(instance.Driver)
			errors.CheckError(err)
			defer containerClient.CloseClient()

			err = containerClient.StopContainer(instance.ContainerID)
			if err != nil {
				log.Fatalf("Failed to stop a container: %v", err)
				return
			}
			log.Printf("Instance %s stopped successfully", instance.Name)

			// update configs

			if instance.AutoRemove {
				_, ok := localConfig.RemoveContext(ctx.Name)
				if !ok {
					log.Fatalf("Context %s does not exist", ctx.Name)
					return
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
				log.Printf("Instance %s status update to Exited", instance.Name)
			}
			err = config.WriteLocalConfig(*localConfig, configFile)
			errors.CheckError(err)
		},
	}

	return stopCmd
}
