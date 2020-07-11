package tfexec

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// mockExecutor impolements the Executor interface for testing.
type mockExecutor struct {
	// mockCommands is a sequece of mocked commands.
	mockCommands []*mockCommand
	// newCommnadContextCalls counts the NewCommandContext method calls.
	newCommnadContextCalls int
	// runCalls counts the Run method calls.
	runCalls int
}

var _ Executor = (*mockExecutor)(nil)

// NewMockExecutor returns a mock executor for testing.
func NewMockExecutor(mockCommands []*mockCommand) Executor {
	return &mockExecutor{
		mockCommands: mockCommands,
	}
}

// NewCommandContext builds and returns an instance of Command.
func (e *mockExecutor) NewCommandContext(ctx context.Context, name string, args ...string) (Command, error) {
	cmd := e.mockCommands[e.newCommnadContextCalls]
	e.newCommnadContextCalls++

	// check if the command call order is expected.
	got := name + " " + strings.Join(args, " ")
	if cmd.argsRe != nil {
		// check with a regex pattern match
		if !cmd.argsRe.MatchString(got) {
			return nil, fmt.Errorf("unexpected NewCommandContext call. got = %s, want = %s", got, cmd.argsRe)
		}
	} else {
		// check with an exact match
		want := strings.Join(cmd.args, " ")
		if got != want {
			return nil, fmt.Errorf("unexpected NewCommandContext call. got = %s, want = %s", got, want)
		}
	}
	return cmd, nil
}

// Run executes a command.
func (e *mockExecutor) Run(cmd Command) error {
	e.runCalls++
	return cmd.Run()
}

// mockCommand implements the Command interface for testing.
type mockCommand struct {
	// args is arguments of the command.
	// Note that args[0] is a name of the command.
	args []string
	// argsRe is an expected regex pattern for a string of args (including
	// command name). It is intended to test args with a regex pattern match
	// insted of an exact match if the args contain a variable such as a path of
	// temporary file.
	argsRe *regexp.Regexp
	// mockStdout is a mocked string for stdout.
	stdout string
	// mockStderr is a mocked string for stderr.
	stderr string
	// mockExitCode is a mocked exit code.
	exitCode int
}

var _ Command = (*mockCommand)(nil)

// Run executes an arbitrary command.
func (c *mockCommand) Run() error {
	if c.exitCode != 0 {
		return &mockExitError{
			exitCode: c.exitCode,
			cmd:      c,
		}
	}
	return nil
}

// Stdout returns outputs of stdout.
func (c *mockCommand) Stdout() string {
	return c.stdout
}

// Stderr returns outputs of stderr.
func (c *mockCommand) Stderr() string {
	return c.stderr
}

// Args returns args of the command.
func (c *mockCommand) Args() []string {
	return c.args
}

// mockExitError implements the ExitError interface for testing.
type mockExitError struct {
	// exitCode is a mocked exit code.
	exitCode int
	// cmd is a executed command.
	cmd Command
}

var _ ExitError = (*exitError)(nil)

// String returns a string represention of the error.
func (e *mockExitError) String() string {
	code := e.ExitCode()
	args := strings.Join(e.cmd.Args(), " ")
	return fmt.Sprintf("mockExitError: exitCode = %d, args = %s", code, args)
}

// Error returns a string useful for displaying error messages.
func (e *mockExitError) Error() string {
	code := e.ExitCode()
	// args[0] contains the command name.
	args := strings.Join(e.cmd.Args(), " ")
	stdout := e.cmd.Stdout()
	stderr := e.cmd.Stderr()
	return fmt.Sprintf(
		"failed to run command (exited %d): %s\nstdout:\n%s\nstderr:\n%s", code, args, stdout, stderr,
	)
}

// ExitCode returns a exit status code of the command.
func (e *mockExitError) ExitCode() int {
	return e.exitCode
}
