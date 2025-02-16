package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Microcks server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping Microcks server...")

		stopCmd := exec.Command("docker", "stop", "microcks-server")

		stopCmd.Stdout = cmd.OutOrStdout()
		stopCmd.Stderr = cmd.OutOrStderr()

		if err := stopCmd.Run(); err != nil {
			fmt.Printf("Error stopping Microcks: %v\n", err)
		} else {
			fmt.Println("Microcks server stopped.")
		}
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
}
