package main

import (
	"os"

	"github.com/microcks/microcks-cli/cmd"
)

func main() {
	var c cmd.Command

	if len(os.Args) == 1 {
		cmd.NewHelpCommand().Execute()
		os.Exit(1)
	}
 //  start and stop command using cobra by preventing overwriting of the previous command
  if os.Args[1] == "start" || os.Args[1] == "stop" {
    cmd.Execute()
    return
  }

	switch os.Args[1] {
	case "version":
		c = cmd.NewVersionCommand()
	case "help":
		c = cmd.NewHelpCommand()
	case "test":
		c = cmd.NewTestCommand()
	case "import":
		c = cmd.NewImportCommand()
	default:
		cmd.NewHelpCommand().Execute()
		os.Exit(1)
	}

	c.Execute()
	return
}
