package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewContextCommand() *cobra.Command {
	var delete bool
	ctxCmd := &cobra.Command{
		Use:     "context [CONTEXT]",
		Aliases: []string{"ctx"},
		Short:   "switch between contexts",
		Example: `# List Microcks context
microcks context/ctx

#Switch Microcks context
microcks context httP://localhost:8080 

# Delete Microcks context
microcks context httP://localhost:8080 --delete`,
		Run: func(cmd *cobra.Command, args []string) {
			var cfgFile string
			configPath, err := config.DefaultLocalConfigPath()
			errors.CheckError(err)
			cfgFile = configPath
			localCfg, err := config.ReadLocalConfig(cfgFile)
			errors.CheckError(err)
			if delete {
				if len(args) == 0 {
					cmd.HelpFunc()(cmd, args)
					os.Exit(1)
				}
				err := deleteContext(args[0], cfgFile)
				errors.CheckError(err)
				return
			}

			if len(args) == 0 {
				printMicrocksContexts(cfgFile)
				return
			}

			ctxName := args[0]
			if localCfg.CurrentContext == ctxName {
				fmt.Printf("Already at context '%s'\n", localCfg.CurrentContext)
				return
			}
			if _, err = localCfg.ResolveContext(ctxName); err != nil {
				log.Fatal(err)
			}
			localCfg.CurrentContext = ctxName
			err = config.WriteLocalConfig(*localCfg, configPath)
			errors.CheckError(err)
			fmt.Printf("Switched to context '%s'\n", localCfg.CurrentContext)
		},
	}

	ctxCmd.Flags().BoolVarP(&delete, "delete", "d", false, "Delete a context")

	return ctxCmd
}

func deleteContext(context, configPath string) error {
	localCfg, err := config.ReadLocalConfig(configPath)
	errors.CheckError(err)
	if localCfg == nil {
		return fmt.Errorf("Nothing to logout from")
	}
	serverName, ok := localCfg.RemoveContext(context)
	if !ok {
		return fmt.Errorf("Context %s does not exist", context)
	}
	_ = localCfg.RemoveUser(context)
	_ = localCfg.RemoveServer(serverName)

	if localCfg.IsEmpty() {
		err := localCfg.DeleteLocalConfig(configPath)
		errors.CheckError(err)
	} else {
		if localCfg.CurrentContext == context {
			localCfg.CurrentContext = ""
		}
		err = config.ValidateLocalConfig(*localCfg)
		if err != nil {
			return fmt.Errorf("Error in logging out")
		}
		err = config.WriteLocalConfig(*localCfg, configPath)
		errors.CheckError(err)
	}
	fmt.Printf("Context '%s' deleted\n", context)
	return nil
}

func printMicrocksContexts(configPath string) {
	localCfg, err := config.ReadLocalConfig(configPath)
	errors.CheckError(err)
	if localCfg == nil {
		log.Fatalf("No contexts defined in %s", configPath)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()
	columnNames := []string{"CURRENT", "NAME", "SERVER"}
	_, err = fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))
	errors.CheckError(err)

	for _, contextRef := range localCfg.Contexts {
		context, err := localCfg.ResolveContext(contextRef.Name)
		if err != nil {
			log.Printf("Context '%s' had error: %v", contextRef.Name, err)
		}
		prefix := " "
		if localCfg.CurrentContext == context.Name {
			prefix = "*"
		}
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\n", prefix, context.Name, context.Server.Server)
		errors.CheckError(err)
	}
}
