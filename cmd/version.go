package cmd

import "fmt"

type versionCommand struct {
}

// NewVersionCommand build a new VersionCommand implementation
func NewVersionCommand() Command {
	return new(versionCommand)
}

// Execute implementation on versionCommand structure
func (c *versionCommand) Execute() {
	fmt.Println("0.2.0")
}
