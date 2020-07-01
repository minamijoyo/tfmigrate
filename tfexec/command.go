package tfexec

import (
	"bytes"
	"os/exec"
)

// Command is an interface for wrapping os/exec.Cmd.
// Since the os/exec.Cmd is a struct, not an interface,
// we cannot mock it for testing, so we wrap it here.
// We implement only functions we need.
type Command interface {
	// Run executes an arbitrary command.
	Run() error
	// Stdout returns outputs of stdout.
	Stdout() string
	// Stderr returns outputs of stderr.
	Stderr() string
	// Args returns args of the command.
	Args() []string
}

// command implements the Command interface.
type command struct {
	// osExecCmd is a underlying object.
	osExecCmd *exec.Cmd
	// stdout is a buffer for stdout.
	stdout *bytes.Buffer
	// stderr is a buffer for stderr.
	stderr *bytes.Buffer
}

var _ Command = (*command)(nil)

// Run executes an arbitrary command.
func (c *command) Run() error {
	return c.osExecCmd.Run()
}

// Stdout returns outputs of stdout.
func (c *command) Stdout() string {
	return c.stdout.String()
}

// Stderr returns outputs of stderr.
func (c *command) Stderr() string {
	return c.stderr.String()
}

// Args returns args of the command.
func (c *command) Args() []string {
	return c.osExecCmd.Args
}
