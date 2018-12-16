package main

import (
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/cmd"
)

func main() {
	fmt.Println("Args: " + os.Args[1])
	var c cmd.Command

	switch os.Args[1] {
	case "version":
		c = cmd.NewVersionCommand()
	case "help":
		c = cmd.NewHelpCommand()
	case "test":
		c = cmd.NewTestCommand()
	default:
		cmd.NewHelpCommand().Execute()
		os.Exit(1)
	}

	c.Execute()
	return
}
