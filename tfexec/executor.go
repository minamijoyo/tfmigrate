package tfexec

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/davecgh/go-spew/spew"
)

// Executor abstracts the os command execution layer.
type Executor interface {
	// NewCommandContext builds and returns an instance of Command.
	NewCommandContext(ctx context.Context, name string, args ...string) (Command, error)
	// Run executes a command.
	Run(cmd Command) error
}

// executor impolements the Executor interface.
type executor struct {
	// outStream is the stdout stream.
	outStream io.Writer
	// errStream is the stderr stream.
	errStream io.Writer

	// a working directory where a command is executed.
	dir string
	// envirommental variables passed to a command.
	env []string
}

var _ Executor = (*executor)(nil)

// NewExecutor returns a default executor for real environments.
func NewExecutor(dir string, env []string) Executor {
	return &executor{
		outStream: os.Stdout,
		errStream: os.Stderr,
		dir:       dir,
		env:       env,
	}
}

// NewCommandContext builds and returns an instance of Command.
func (e *executor) NewCommandContext(ctx context.Context, name string, args ...string) (Command, error) {
	osExecCmd := exec.CommandContext(ctx, name, args...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	osExecCmd.Stdout = stdout
	osExecCmd.Stderr = stderr
	osExecCmd.Dir = e.dir
	osExecCmd.Env = e.env

	return &command{
		osExecCmd: osExecCmd,
		stdout:    stdout,
		stderr:    stderr,
	}, nil
}

// Run executes a command.
func (e *executor) Run(cmd Command) error {
	err := cmd.Run()
	log.Printf("[DEBUG] run command: %s", spew.Sdump(cmd))
	if err != nil {
		log.Printf("[DEBUG] failed to run command: %s", spew.Sdump(err))
		if osExecErr, ok := err.(*exec.ExitError); ok {
			return &exitError{
				osExecErr: osExecErr,
				cmd:       cmd,
			}
		}
		return err
	}
	return nil
}
