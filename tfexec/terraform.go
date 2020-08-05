package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mattn/go-shellwords"
)

// State is a named type for tfstate.
// We don't parse contents of tfstate to avoid depending on internal details,
// but we define it as a named type to clarify interface.
type State []byte

// Bytes returns raw contents of tfstate as []byte.
func (s *State) Bytes() []byte {
	return []byte(*s)
}

// NewState returns a new State instance with a given content of tfstate.
func NewState(b []byte) *State {
	s := State(b)
	return &s
}

// Plan is a named type for tfplan.
// We don't parse contents of tfplan to avoid depending on internal details,
// but we define it as a named type to clarify interface.
type Plan []byte

// Bytes returns raw contents of tfstate as []byte.
func (p *Plan) Bytes() []byte {
	return []byte(*p)
}

// NewPlan returns a new Plan instance with a given content of tfplan.
func NewPlan(b []byte) *Plan {
	p := Plan(b)
	return &p
}

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
	// Verison returns a version number of Terraform.
	Version(ctx context.Context) (string, error)

	// Init initializes a given work directory.
	Init(ctx context.Context, dir string, opts ...string) error

	// Plan computes expected changes.
	// If a state is given, use it for the input state.
	Plan(ctx context.Context, state *State, dir string, opts ...string) (*Plan, error)

	// Apply applies changes.
	// If a plan is given, use it for the input plan.
	Apply(ctx context.Context, plan *Plan, dir string, opts ...string) error

	// Destroy destroys resources.
	Destroy(ctx context.Context, dir string, opts ...string) error

	// Import imports an existing resource to state.
	// If a state is given, use it for the input state.
	Import(ctx context.Context, state *State, address string, id string, opts ...string) (*State, error)

	// StateList shows a list of resources.
	// If a state is given, use it for the input state.
	StateList(ctx context.Context, state *State, addresses []string, opts ...string) ([]string, error)

	// StatePull returns the current tfstate from remote.
	StatePull(ctx context.Context, opts ...string) (*State, error)

	// StateMv moves resources from source to destination address.
	// If a state argument is given, use it for the input state.
	// If a stateOut argument is given, move resources from state to stateOut.
	// It returns updated the given state and the stateOut.
	StateMv(ctx context.Context, state *State, stateOut *State, source string, destination string, opts ...string) (*State, *State, error)

	// StateRm removes resources from state.
	// If a state is given, use it for the input state and return a new state.
	// Note that if the input state is not given, always return nil state,
	// becasuse the terraform state rm command doesn't have -state-out option.
	StateRm(ctx context.Context, state *State, addresses []string, opts ...string) (*State, error)

	// StatePush pushs a given State to remote.
	StatePush(ctx context.Context, state *State, opts ...string) error

	// Run is a low-level generic method for running an arbitrary terraform comamnd.
	Run(ctx context.Context, args ...string) (string, string, error)

	// dir returns a working directory where terraform command is executed.
	Dir() string

	// SetExecPath customizes how the terraform command is executed. Default to terraform.
	// It's intended to inject a wrapper command such as direnv.
	SetExecPath(execPath string)
}

// terraformCLI implements the TerraformCLI interface.
type terraformCLI struct {
	// Executor is an interface which executes an arbitrary command.
	Executor

	// execPath is a string which executes the terraform command.
	// If empty, default to terraform.
	execPath string
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
	// The default binary path is `terraform`.
	name := "terraform"
	// If execPath is customized
	if len(c.execPath) > 0 {
		// execPath may contain spaces and environment variables, so we parse it.
		// e.g.) "direnv exec . terraform" => ["direnv", "exec", ".", "terraform"]
		parts, err := shellwords.Parse(c.execPath)
		if err != nil {
			return "", "", err
		}
		// The first part is a binary path.
		name = parts[0]
		// if execPath contains spaces, insert remains before original arguments.
		if len(parts) > 1 {
			args = append(parts[1:], args...)
		}
	}

	cmd, err := c.Executor.NewCommandContext(ctx, name, args...)
	if err != nil {
		return "", "", err
	}

	err = c.Executor.Run(cmd)

	return cmd.Stdout(), cmd.Stderr(), err
}

// dir returns a working directory where terraform command is executed.
func (c *terraformCLI) Dir() string {
	return c.Executor.Dir()
}

// SetExecPath customizes how the terraform command is executed. Default to terraform.
// It's intended to inject a wrapper command such as direnv.
func (c *terraformCLI) SetExecPath(execPath string) {
	c.execPath = execPath
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
