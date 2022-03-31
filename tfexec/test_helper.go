package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
)

// mockExecutor implements the Executor interface for testing.
type mockExecutor struct {
	// mockCommands is a sequence of mocked commands.
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

// Dir returns the current working directory.
func (e *mockExecutor) Dir() string {
	// no op.
	return ""
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
	// instead of an exact match if the args contain a variable such as a path of
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

// String returns a string representation of the error.
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

// ExitCode returns an exit status code of the command.
func (e *mockExitError) ExitCode() int {
	return e.exitCode
}

// testAccSourceFileName is a filename of terraform configuration for testing.
var testAccSourceFileName = "main.tf"

// SkipUnlessAcceptanceTestEnabled skips acceptance tests unless TEST_ACC is set to 1.
func SkipUnlessAcceptanceTestEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("skip acceptance tests")
	}
}

// SetupTestAcc is a common setup helper for acceptance tests.
func SetupTestAcc(t *testing.T, source string) Executor {
	t.Helper()
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

// setupTestWorkDir creates temporary working directory with a given source for testing.
func setupTestWorkDir(source string) (string, error) {
	workDir, err := ioutil.TempDir("", "workDir")
	if err != nil {
		return "", fmt.Errorf("failed to create work dir: %s", err)
	}

	if err := ioutil.WriteFile(filepath.Join(workDir, testAccSourceFileName), []byte(source), 0600); err != nil {
		os.RemoveAll(workDir)
		return "", fmt.Errorf("failed to create main.tf: %s", err)
	}
	return workDir, nil
}

// setupTestPluginCacheDir sets TF_PLUGIN_CACHE_DIR to a given executor.
func setupTestPluginCacheDir(e Executor) error {
	dir := os.Getenv("TF_PLUGIN_CACHE_DIR")
	if len(dir) == 0 {
		// default to ../tmp/plugin-cache
		_, filename, _, _ := runtime.Caller(0)
		dir = path.Join(path.Dir(filename), "..", "tmp", "plugin-cache")
	}

	// Terraform v0.13+ doesn't create dir if not exist.
	// So we create it if not exist.
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create plugin cache dir: %s", err)
	}
	e.AppendEnv("TF_PLUGIN_CACHE_DIR", dir)
	return nil
}

// GetTestAccBackendS3Config returns mocked backend s3 config for testing.
// Its endpoint can be set via LOCALSTACK_ENDPOINT environment variable.
// default to "http://localhost:4566"
func GetTestAccBackendS3Config(dir string) string {
	endpoint := "http://localhost:4566"
	localstackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if len(localstackEndpoint) > 0 {
		endpoint = localstackEndpoint
	}

	backendConfig := fmt.Sprintf(`
terraform {
  # https://www.terraform.io/docs/backends/types/s3.html
  backend "s3" {
    region = "ap-northeast-1"
    bucket = "tfstate-test"
    key    = "%s/terraform.tfstate"

    // mock s3 endpoint with localstack
    endpoint                    = "%s"
    access_key                  = "dummy"
    secret_key                  = "dummy"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    force_path_style            = true
  }
}

# https://www.terraform.io/docs/providers/aws/index.html
# https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html#localstack
provider "aws" {
  region = "ap-northeast-1"

  access_key                  = "dummy"
  secret_key                  = "dummy"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true
  s3_force_path_style         = true

  // mock endpoints with localstack
  endpoints {
    s3  = "%s"
    ec2 = "%s"
    iam = "%s"
  }
}
`, dir, endpoint, endpoint, endpoint, endpoint)
	return backendConfig
}

// SetupTestAccWithApply is an acceptance test helper for initializing a
// temporary work directory and applying a given source.
func SetupTestAccWithApply(t *testing.T, workspace string, source string) TerraformCLI {
	t.Helper()

	e := SetupTestAcc(t, source)
	tf := NewTerraformCLI(e)
	ctx := context.Background()

	err := tf.Init(ctx, "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	//default workspace always exists so don't try to create it
	if workspace != "default" {
		err = tf.WorkspaceNew(ctx, workspace)
		if err != nil {
			t.Fatalf("failed to run terraform workspace new %s : %s", workspace, err)
		}
	}

	err = tf.Apply(ctx, nil, "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	// destroy resources after each test not to have any state.
	t.Cleanup(func() {
		err := tf.Destroy(ctx, "-input=false", "-no-color", "-auto-approve")
		if err != nil {
			t.Fatalf("failed to run terraform destroy: %s", err)
		}
	})

	return tf
}

// UpdateTestAccSource updates a terraform configuration file with a given contents.
func UpdateTestAccSource(t *testing.T, tf TerraformCLI, source string) {
	t.Helper()
	if err := ioutil.WriteFile(filepath.Join(tf.Dir(), testAccSourceFileName), []byte(source), 0600); err != nil {
		t.Fatalf("failed to update source: %s", err)
	}
}

// MatchTerraformVersion returns true if terraform version matches a given constraints.
func MatchTerraformVersion(ctx context.Context, tf TerraformCLI, constraints string) (bool, error) {
	tfVersionRaw, err := tf.Version(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get terraform version: %s", err)
	}
	v, err := version.NewVersion(tfVersionRaw)
	if err != nil {
		return false, fmt.Errorf("failed to parse terraform version: %s", err)
	}
	c, err := version.NewConstraint(constraints)
	if err != nil {
		return false, fmt.Errorf("failed to new version constraint: %s", err)
	}
	return c.Check(v), nil
}
