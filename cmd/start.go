package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Microcks server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Microcks server...")

		startCmd := exec.Command("docker", "run", "-d", "-p", "8585:8080", "--rm", "--name", "microcks-server", "quay.io/microcks/microcks-uber:latest-native")

		startCmd.Stdout = nil
		startCmd.Stderr = os.Stderr

		if err := startCmd.Run(); err != nil {
			fmt.Printf("Error starting Microcks: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
