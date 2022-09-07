package cmd

import (
	"fmt"

	"github.com/microcks/microcks-cli/version"
)

type versionCommand struct {
}

// NewVersionCommand build a new VersionCommand implementation
func NewVersionCommand() Command {
	return new(versionCommand)
}

// Execute implementation on versionCommand structure
func (c *versionCommand) Execute() {
	fmt.Println(version.Version)
}
