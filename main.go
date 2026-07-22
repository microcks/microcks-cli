package main

import (
	"github.com/microcks/microcks-cli/cmd"
)

func main() {
	// cmd.Handle is the single exit point: it prints the error to stderr and
	// maps its Failure Kind to an exit code. Nothing else in the tree exits.
	command, err := cmd.NewCommand()
	if err != nil {
		cmd.Handle(err)
	}
	cmd.Handle(command.Execute())
}
