package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// tfVersionRe is a pattern to parse outputs from terraform version.
var tfVersionRe = regexp.MustCompile(`^Terraform v(.+)\s*$`)

// State is a named type for tfstate.
// We doesn't need to parse contents of tfstate,
// but we define it as a name type to clarify TerraformCLI interface.
type State string

// TerraformCLI is an interface for executing the terraform command.
// The main features of the terraform command are many of side effects, and the
// most of stdout may not be useful. In addition, the interfaces of state
// subcommands are inconsistent, and if a state file is required for the
// argument, we need a temporary file. However, It's hard to clean up the
// temporary file when an error occurs in the middle of a series of commands.
// This means implementing the exactly same interface for the terraform command
// doesn't make sense for us. So we wrap the terraform command and provider a
// high-level and easy-to-use interface which can be used in memory as much as
// possible.
// The interface is an opinionated, if it doesn't match you need, you can use
// Run(), which is a low-level generic method for running an arbitrary
// terraform command.
type TerraformCLI interface {
	// Run is a low-level generic method for running an arbitrary terraform comamnd.
	Run(ctx context.Context, args ...string) (string, string, error)

	// Verison returns a version number of Terraform.
	Version(ctx context.Context) (string, error)

	// Init initializes a given work directory.
	Init(ctx context.Context, dir string, opts ...string) error

	// Apply applies changes.
	Apply(ctx context.Context, dirOrPlan string, opts ...string) error

	// Destroy destroys resources.
	Destroy(ctx context.Context, dir string, opts ...string) error

	// StateList shows a list of resources.
	// If a state is given, use it for the input state.
	StateList(ctx context.Context, state *State, addresses []string, opts ...string) ([]string, error)

	// StatePull returns the current tfstate from remote.
	StatePull(ctx context.Context) (State, error)

	// StatePush pushs a given State to remote.
	StatePush(ctx context.Context, state State) error
}

// terraformCLI implements the TerraformCLI interface.
type terraformCLI struct {
	// Executor is an interface which executes an arbitrary command.
	Executor
}

var _ TerraformCLI = (*terraformCLI)(nil)

// NewTerraformCLI returns an implementation of the TerraformCLI interface.
func NewTerraformCLI(e Executor) TerraformCLI {
	return &terraformCLI{
		Executor: e,
	}
}

// Run is a low-level generic method for running an arbitrary terraform comamnd.
func (c *terraformCLI) Run(ctx context.Context, args ...string) (string, string, error) {
	cmd, err := c.Executor.NewCommandContext(ctx, "terraform", args...)
	if err != nil {
		return "", "", err
	}

	err = c.Executor.Run(cmd)

	return cmd.Stdout(), cmd.Stderr(), err
}

// Verison returns a version number of Terraform.
func (c *terraformCLI) Version(ctx context.Context) (string, error) {
	stdout, _, err := c.Run(ctx, "version")
	if err != nil {
		return "", err
	}

	matched := tfVersionRe.FindStringSubmatch(stdout)
	if len(matched) != 2 {
		return "", fmt.Errorf("failed to parse terraform version: %s", stdout)
	}
	version := matched[1]
	return version, nil
}

// Init initializes a given work directory.
func (c *terraformCLI) Init(ctx context.Context, dir string, opts ...string) error {
	args := []string{"init"}
	args = append(args, opts...)
	if len(dir) > 0 {
		args = append(args, dir)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}

// Apply applies changes.
func (c *terraformCLI) Apply(ctx context.Context, dirOrPlan string, opts ...string) error {
	args := []string{"apply"}
	args = append(args, opts...)
	if len(dirOrPlan) > 0 {
		args = append(args, dirOrPlan)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}

// Destroy destroys resources.
func (c *terraformCLI) Destroy(ctx context.Context, dir string, opts ...string) error {
	args := []string{"destroy"}
	args = append(args, opts...)
	if len(dir) > 0 {
		args = append(args, dir)
	}
	_, _, err := c.Run(ctx, args...)
	return err
}

// StateList shows a list of resources.
// If a state is given, use it for the input state.
func (c *terraformCLI) StateList(ctx context.Context, state *State, addresses []string, opts ...string) ([]string, error) {
	args := []string{"state", "list"}

	if state != nil {
		if hasPrefixOptions(opts, "-state=") {
			return nil, fmt.Errorf("failed to build `terraform state list` options. The state argument (!= nil) and the -state= option cannot be set at the same time: state=%v, opts=%v", state, opts)
		}
		tmpfile, err := writeTempFile([]byte(*state))
		defer os.Remove(tmpfile.Name())
		if err != nil {
			return nil, err
		}
		args = append(args, "-state="+tmpfile.Name())
	}

	args = append(args, opts...)

	if len(addresses) > 0 {
		args = append(args, addresses...)
	}

	stdout, _, err := c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimRight(stdout, "\n"), "\n"), nil
}

// StatePull returns the current tfstate from remote.
func (c *terraformCLI) StatePull(ctx context.Context) (State, error) {
	stdout, _, err := c.Run(ctx, "state", "pull")
	if err != nil {
		return "", err
	}

	return State(stdout), nil
}

// StatePush pushs a given State to remote.
func (c *terraformCLI) StatePush(ctx context.Context, state State) error {
	tmpfile, err := writeTempFile([]byte(state))
	defer os.Remove(tmpfile.Name())
	if err != nil {
		return err
	}

	_, _, err = c.Run(ctx, "state", "push", tmpfile.Name())
	return err
}

// writeTempFile writes content to a temporary file and return its file.
func writeTempFile(content []byte) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return tmpfile, fmt.Errorf("failed to create temporary file: %s", err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		return tmpfile, fmt.Errorf("failed to write temporary file: %s", err)
	}

	if err := tmpfile.Close(); err != nil {
		return tmpfile, fmt.Errorf("failed to close temporary file: %s", err)
	}

	return tmpfile, nil
}

// hasPrefixOptions returns true if any element in a list of string has a given prefix.
func hasPrefixOptions(opts []string, prefix string) bool {
	for _, opt := range opts {
		if strings.HasPrefix(opt, prefix) {
			return true
		}
	}
	return false
}
