package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCompletionCommand(t *testing.T) {
	cmd := NewCompletionCommand()

	assert.Equal(t, "completion [bash|zsh|fish|powershell]", cmd.Use)
	assert.Equal(t, "Generate shell completion scripts", cmd.Short)
}

func TestCompletionCommandRegistered(t *testing.T) {
	cmd := NewCommad()

	completionCmd, _, err := cmd.Find([]string{"completion"})

	assert.NoError(t, err)
	assert.NotNil(t, completionCmd)
	assert.Equal(t, "completion", completionCmd.Name())
}

func TestCompletionCommandGeneratesBash(t *testing.T) {
	cmd := NewCommad()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "bash"})

	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "completion for microcks")
}

func TestCompletionCommandRejectsUnsupportedShell(t *testing.T) {
	cmd := NewCommad()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "nushell"})

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}
