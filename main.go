package main

import (
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/cmd"
)

func main() {
	command := cmd.NewCommad()
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
