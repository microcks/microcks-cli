package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewLogoutCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	logoutCmd := &cobra.Command{
		Use:   "logout CONTEXT",
		Short: "Log out from Microcks",
		Long:  "Log out from Microcks",
		Example: `# To log out of Microcks
$ microcks logout`,

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			context := args[0]
			localCfg, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
			errors.CheckError(err)
			if localCfg == nil {
				log.Fatalf("Nothing to logout from")
			}

			// Remove authToken
			ok := localCfg.RemoveToken(context)
			if !ok {
				log.Fatalf("Context %s does not exist", context)
			}

			err = config.ValidateLocalConfig(*localCfg)
			if err != nil {
				log.Fatalf("Error in loging out: %s", err)
			}
			err = config.WriteLocalConfig(*localCfg, globalClientOpts.ConfigPath)
			errors.CheckError(err)

			fmt.Printf("Logged out from '%s'\n", context)
		},
	}

	return logoutCmd
}
