package tfexec

import (
	"fmt"
	"os/exec"
	"strings"
)

// ExitError is an interface for wrapping os/exec.ExitError.
// We want to add helper methods we need.
type ExitError interface {
	// String returns a string representation of the error.
	String() string
	// Error returns a string useful for displaying error messages.
	Error() string
	// ExitCode returns an exit status code of the command.
	ExitCode() int
}

// exitError implements the ExitError interface.
type exitError struct {
	// osExecErr is an underlying object.
	osExecErr *exec.ExitError
	// cmd is an executed command.
	cmd Command
}

var _ ExitError = (*exitError)(nil)

// String returns a string representation of the error.
func (e *exitError) String() string {
	return e.osExecErr.String()
}

// Error returns a string useful for displaying error messages.
func (e *exitError) Error() string {
	code := e.ExitCode()
	// args[0] contains the command name.
	args := strings.Join(e.cmd.Args(), " ")
	stdout := e.cmd.Stdout()
	stderr := e.cmd.Stderr()
	return fmt.Sprintf(
		"failed to run command (exited %d): %s\nstdout:\n%s\nstderr:\n%s", code, args, stdout, stderr,
	)
}

// ExitCode returns an exit status code of the command.
func (e *exitError) ExitCode() int {
	return e.osExecErr.ExitCode()
}
