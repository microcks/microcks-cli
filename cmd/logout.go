package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

func NewLogoutCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	logoutCmd := &cobra.Command{
		Use:   "logout CONTEXT",
		Short: "Log out from Microcks",
		Long:  "Log out from Microcks",
		Example: `# Log out from a Microcks server URL
microcks logout http://localhost:8080

# Log out from a named context
microcks logout dev-context`,

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			target := args[0]
			err := logoutContext(target, globalClientOpts.ConfigPath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Logged out from '%s'\n", target)
		},
	}

	return logoutCmd
}

func logoutContext(target, configPath string) error {
	localCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return err
	}
	if localCfg == nil {
		return fmt.Errorf("Nothing to logout from")
	}

	userName := target
	if ctx, err := localCfg.ResolveContext(target); err == nil {
		userName = ctx.User.Name
	}

	if ok := localCfg.RemoveToken(userName); !ok {
		return fmt.Errorf("Context %s does not exist", target)
	}

	err = config.ValidateLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("Error in loging out: %s", err)
	}

	return config.WriteLocalConfig(*localCfg, configPath)
}
