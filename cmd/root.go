package cmd

import (
  "fmt"
  "os"
  "github.com/spf13/cobra"
)

//rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "microcks-cli",
  Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
    os.Exit(1)
	}
}
