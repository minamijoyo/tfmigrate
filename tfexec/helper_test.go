package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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
	// store called args to pass runFunc callback.
	cmd.calledArgs = args

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

// AppendEnv appends an environment variable.
func (e *mockExecutor) AppendEnv(key string, value string) {
	// no op.
}

// mockRunFunc is a type for callback of mockCommand.Run() to allow us to cause side effects.
type mockRunFunc func(args ...string) error

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
	// calledArgs stores arguments actually called to pass runFunc.
	calledArgs []string
	// runFunc is a callback of Run() to allow us to cause side effects.
	runFunc mockRunFunc
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
	if c.runFunc != nil {
		err := c.runFunc(c.calledArgs...)
		if err != nil {
			return err
		}
	}

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

// SetupTestAcc is a common setup helper for acceptance tests.
func SetupTestAcc(t *testing.T, source string) Executor {
	workDir, err := setupTestWorkDir(source)
	if err != nil {
		t.Fatalf("failed to setup work dir: %s", err)
	}
	t.Cleanup(func() { os.RemoveAll(workDir) })

	e := NewExecutor(workDir, os.Environ())
	if err := setupTestPluginCacheDir(e); err != nil {
		t.Fatalf("failed to set plugin cache dir: %s", err)
	}

	return e
}

// SkipUnlessAcceptanceTestEnabled skips acceptance tests unless TEST_ACC is set to 1.
func SkipUnlessAcceptanceTestEnabled(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("skip acceptance tests")
	}
}

// setupTestWorkDir creates temporary working directory with a given source for testing.
func setupTestWorkDir(source string) (string, error) {
	workDir, err := ioutil.TempDir("", "workDir")
	if err != nil {
		return "", fmt.Errorf("failed to create work dir: %s", err)
	}

	if err := ioutil.WriteFile(filepath.Join(workDir, "main.tf"), []byte(source), 0644); err != nil {
		os.RemoveAll(workDir)
		return "", fmt.Errorf("failed to create main.tf: %s", err)
	}
	return workDir, nil
}

// setupTestPluginCacheDir sets TF_PLUGIN_CACHE_DIR to a given executor.
func setupTestPluginCacheDir(e Executor) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current dir: %s", err)
	}
	e.AppendEnv("TF_PLUGIN_CACHE_DIR", filepath.Join(pwd, "tmp/plugin-cache"))
	return nil
}
