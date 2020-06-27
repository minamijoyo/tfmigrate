package tfexec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// osExecCommandFunc is a dependency injection point for the os/exec.Command function.
type osExecCommandFunc func(name string, args ...string) *exec.Cmd

// Executor abstracts the os command execution layer.
type Executor struct {
	// osExecCommandFunc defaults to os/exec.Command, but we can replace it for testing.
	osExecCommandFunc osExecCommandFunc

	// outStream is the stdout stream.
	outStream io.Writer
	// errStream is the stderr stream.
	errStream io.Writer

	// a working directory where a command is executed.
	workDir string
	// envirommental variables passed to a command.
	envVars []string
}

// NewDefaultExecutor returns a default executor for real environments.
func NewDefaultExecutor() *Executor {
	return &Executor{
		osExecCommandFunc: exec.Command,
		outStream:         os.Stdout,
		errStream:         os.Stderr,
		workDir:           ".",
		envVars:           os.Environ(),
	}
}

// Command wraps os/exec.Cmd to mock for testing.
type Command struct {
	osExecCmd *exec.Cmd
	outBuf    *bytes.Buffer
	errBuf    *bytes.Buffer
}

// NewCommand is a helper function for building a command instance.
func (e *Executor) NewCommand(name string, args ...string) *Command {
	osExecCmd := e.osExecCommandFunc(name, args...)

	// set environments
	osExecCmd.Dir = e.workDir
	osExecCmd.Env = append(osExecCmd.Env, e.envVars...)

	// set stdout/stderr
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	osExecCmd.Stdout = &outBuf
	osExecCmd.Stderr = &errBuf

	return &Command{
		osExecCmd: osExecCmd,
		outBuf:    &outBuf,
		errBuf:    &errBuf,
	}
}

// Run executes an arbitrary command.
func (e *Executor) Run(cmd *Command) error {
	err := cmd.osExecCmd.Run()
	if err != nil {
		fmt.Fprintf(e.outStream, "%s", cmd.outBuf)
		fmt.Fprintf(e.errStream, "%s", cmd.errBuf)
		return fmt.Errorf("failed to exec command: %s: %s", strings.Join(cmd.osExecCmd.Args, " "), err)
	}

	fmt.Fprintf(e.outStream, "%s", cmd.outBuf)
	fmt.Fprintf(e.errStream, "%s", cmd.errBuf)
	return nil
}
