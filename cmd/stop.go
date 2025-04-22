package cmd

import (
	"context"
	"fmt"
	"log"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop microcks instance",
		Long:  "stop microcks instance",
		Run: func(cmd *cobra.Command, args []string) {

			cfg, err := config.LoadConfig(config.ConfigPath)
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			cli, err := createClient(cfg.Instance.Driver)

			if err != nil {
				fmt.Println(err)
				return
			}

			stopContainer(cfg.Instance.ContainerID, cli)

			cfg.Instance.Status = "Stopped"

			if cfg.Instance.AutoRemove {
				cfg.Instance = struct {
					Name        string "yaml:\"name\""
					Image       string "yaml:\"image\""
					Status      string "yaml:\"status\""
					Port        string "yaml:\"port\""
					ContainerID string "yaml:\"containerID\""
					AutoRemove  bool   "yaml:\"autoRemove\""
					Driver      string "yaml:\"driver\""
				}{}
			}

			err = config.SaveConfig(config.ConfigPath, cfg)

			if err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}

			fmt.Println("Microcks stopped successfully...")
		},
	}

	return stopCmd
}

func stopContainer(containerId string, cli *client.Client) {
	ctx := context.Background()

	fmt.Print("Stopping container ", containerId, "... ")
	noWaitTimeout := 0 // to not wait for the container to exit gracefully
	if err := cli.ContainerStop(ctx, containerId, containertypes.StopOptions{Timeout: &noWaitTimeout}); err != nil {
		panic(err)
	}
	fmt.Println("Success")
}
