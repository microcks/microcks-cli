package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
  "github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Microcks server with your desired port or default port",
	Run: func(cmd *cobra.Command, args []string) {
    // Set the port for the server
    port := viper.GetInt("port")
		fmt.Printf("Starting Microcks server on port %d...\n", port)

		startCmd := exec.Command("docker", "run", "-d", "-p", fmt.Sprintf("%d:8080", port), "--rm", "--name", "microcks-server", "quay.io/microcks/microcks-uber:latest-native")

		startCmd.Stdout = nil
		startCmd.Stderr = os.Stderr

		if err := startCmd.Run(); err != nil {
			fmt.Printf("Error starting Microcks: %v\n", err)
			os.Exit(1)
		}
    fmt.Printf("Microcks server is running on localhost:%d\n", port)
	},
}

func init() {
	// Add the start command to the root command
	rootCmd.AddCommand(startCmd)

	// Set up Viper to read from flags
	startCmd.Flags().Int("port", 8585, "Port for Microcks server")
	viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
}
