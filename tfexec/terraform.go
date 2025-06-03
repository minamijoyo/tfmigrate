package tfexec

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
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

// Bytes returns raw contents of tfplan as []byte.
func (p *Plan) Bytes() []byte {
	return []byte(*p)
}

// NewPlan returns a new Plan instance with a given content of tfplan.
func NewPlan(b []byte) *Plan {
	p := Plan(b)
	return &p
}

// TerraformCLI is an interface for executing the terraform command.
// The main results of the terraform command are generally side effects, and the
// stdout functionality may not be useful. In addition, the interfaces of state
// subcommands are inconsistent, and if a state file is required for the
// argument, we need a temporary file. However, it's hard to clean up the
// temporary file when an error occurs in the middle of a series of commands.
// This means implementing the exactly same interface for the terraform command
// doesn't make sense for us. So we wrap the terraform command and provider a
// high-level and easy-to-use interface which can be used in memory as much as
// possible.
// As a result, the interface is opinionated and less flexible. For running arbitrary terraform commands
// you can use Run(), which is a low-level generic method.
type TerraformCLI interface {
	// Version returns the Terraform execType and version number.
	// The execType can be either terraform or opentofu.
	Version(ctx context.Context) (string, *version.Version, error)

	// Init initializes the current work directory.
	Init(ctx context.Context, opts ...string) error

	// Plan computes expected changes.
	// If a state is given, use it for the input state.
	Plan(ctx context.Context, state *State, opts ...string) (*Plan, error)

	// Apply applies changes.
	// If a plan is given, use it for the input plan.
	Apply(ctx context.Context, plan *Plan, opts ...string) error

	// Destroy destroys resources.
	Destroy(ctx context.Context, opts ...string) error

	// Import imports an existing resource to state.
	// If a state is given, use it for the input state.
	Import(ctx context.Context, state *State, address string, id string, opts ...string) (*State, error)

	// Providers shows a tree of modules in the referenced configuration annotated with
	// their provider requirements.
	Providers(ctx context.Context) (string, error)

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
	// because the terraform state rm command doesn't have -state-out option.
	StateRm(ctx context.Context, state *State, addresses []string, opts ...string) (*State, error)

	// StateReplaceProvider replaces a provider from source to destination address.
	// If a state argument is given, use it for the input state.
	// It returns the given state.
	// Unlike other state subcommands, the terraform state replace-provider
	// command doesn't support a -state-out option; it only supports the -state option.
	StateReplaceProvider(ctx context.Context, state *State, source string, destination string, opts ...string) (*State, error)

	// StatePush pushes a given State to remote.
	StatePush(ctx context.Context, state *State, opts ...string) error

	// WorkspaceNew creates a new workspace with name "workspace".
	WorkspaceNew(ctx context.Context, workspace string, opts ...string) error

	// WorkspaceShow returns the current selected workspace.
	WorkspaceShow(ctx context.Context) (string, error)

	// WorkspaceSelect switches to the workspace with name "workspace". This workspace should already exist.
	WorkspaceSelect(ctx context.Context, workspace string) error

	// Run is a low-level generic method for running an arbitrary terraform command.
	Run(ctx context.Context, args ...string) (string, string, error)

	// Dir returns a working directory where terraform command is executed.
	Dir() string

	// SetExecPath customizes how the terraform command is executed. Default to terraform.
	// It's intended to inject a wrapper command such as direnv.
	SetExecPath(execPath string)

	// ExecPath returns the current execution path for the terraform CLI.
	// This is primarily used for testing.
	ExecPath() string

	// OverrideBackendToLocal switches the backend to local and returns a function
	// to switch it back to remote with defer.
	// The -state flag for terraform command is not valid for remote state,
	// so we need to switch the backend to local for temporary state operations.
	// The filename argument must meet constraints for override file.
	// (e.g.) _tfexec_override.tf
	OverrideBackendToLocal(ctx context.Context, filename string, workspace string, isBackendTerraformCloud bool, backendConfig []string, supportsStateReplaceProvider bool) (func() error, error)

	// PlanHasChange is a helper method which runs plan and return true if the plan has change.
	PlanHasChange(ctx context.Context, state *State, opts ...string) (bool, error)

	// SupportsStateReplaceProvider is a helper method used to determine whether or
	// not the terraform version supports `state replace-provider`.
	SupportsStateReplaceProvider(ctx context.Context) (bool, version.Constraints, error)
}

// terraformCLI implements the TerraformCLI interface.
type terraformCLI struct {
	// Executor is an interface which executes an arbitrary command.
	Executor

	// execPath is a string which executes the terraform command.
	// Default to terraform. To use OpenTofu, set this to `tofu`.
	execPath string
}

var _ TerraformCLI = (*terraformCLI)(nil)

// NewTerraformCLI returns an implementation of the TerraformCLI interface.
// This function reads the environment variable TFMIGRATE_EXEC_PATH and sets it
// to execPath.
func NewTerraformCLI(e Executor) TerraformCLI {
	execPath := os.Getenv("TFMIGRATE_EXEC_PATH")
	if len(execPath) == 0 {
		// The default binary path is `terraform`.
		execPath = "terraform"
	}

	return &terraformCLI{
		Executor: e,
		execPath: execPath,
	}
}

// Run is a low-level generic method for running an arbitrary terraform command.
func (c *terraformCLI) Run(ctx context.Context, args ...string) (string, string, error) {
	name := c.execPath
	// If execPath is customized
	if name != "terraform" {
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

// Dir returns a working directory where terraform command is executed.
func (c *terraformCLI) Dir() string {
	return c.Executor.Dir()
}

// SetExecPath customizes how the terraform command is executed. Default to terraform.
// It's intended to inject a wrapper command such as direnv.
func (c *terraformCLI) SetExecPath(execPath string) {
	c.execPath = execPath
}

// ExecPath returns the current execution path for the terraform CLI.
// This is primarily used for testing.
func (c *terraformCLI) ExecPath() string {
	return c.execPath
}

// OverrideBackendToLocal switches the backend to local and returns a function
// that will switch it back to remote with defer.
// The -state flag for terraform command is not valid for remote state,
// so we need to switch the backend to local for temporary state operations.
// The filename argument must meet constraints in order to override the file.
// (e.g.) _tfexec_override.tf
func (c *terraformCLI) OverrideBackendToLocal(ctx context.Context, filename string,
	workspace string, isBackendTerraformCloud bool, backendConfig []string, supportsStateReplaceProvider bool) (func() error, error) {
	// create local backend override file.
	path := filepath.Join(c.Dir(), filename)
	contents := `
terraform {
  backend "local" {
  }
}
`
	log.Printf("[INFO] [executor@%s] create an override file\n", c.Dir())
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		return nil, fmt.Errorf("failed to create override file: %s", err)
	}

	// create local workspace state directory
	workspaceStatePath := filepath.Join(c.Dir(), "terraform.tfstate.d", workspace)
	workspacePath := filepath.Join(c.Dir(), "terraform.tfstate.d")
	log.Printf("[INFO] [migrator@%s] creating local workspace folder in: %s\n", c.Dir(), workspaceStatePath)
	if err := os.MkdirAll(workspaceStatePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create local workspace state directory: %s", err)
	}

	log.Printf("[INFO] [executor@%s] switch backend to local\n", c.Dir())
	err := c.Init(ctx, "-input=false", "-no-color", "-reconfigure")
	if err != nil {
		// remove the override file before return an error.
		os.Remove(path)
		os.Remove(workspaceStatePath)
		os.Remove(workspacePath)
		return nil, fmt.Errorf("failed to switch backend to local: %s", err)
	}

	switchBackToRemoteFunc := func() error {
		log.Printf("[INFO] [executor@%s] remove the override file\n", c.Dir())
		err := os.Remove(path)
		if err != nil {
			log.Printf("[ERROR] [executor@%s] failed to remove the override file: %s\n", c.Dir(), err)
			log.Printf("[ERROR] [executor@%s] please remove the override file(%s) and re-run terraform init -reconfigure\n", c.Dir(), path)
			return err
		}
		// cleanup the local workspace directly used for local state
		log.Printf("[INFO] [executor@%s] remove the workspace state folder\n", c.Dir())
		err = os.Remove(workspaceStatePath)
		if err != nil {
			log.Printf("[ERROR] [executor@%s] failed to remove local workspace state directory: %s\n", c.Dir(), err)
			log.Printf("[ERROR] [executor@%s] please remove the local workspace state directory(%s) and re-run terraform init -reconfigure\n", c.Dir(), workspaceStatePath)
			return err
		}
		err = os.Remove(workspacePath)
		if err != nil {
			log.Printf("[ERROR] [executor@%s] failed to remove local workspace directory: %s\n", c.Dir(), err)
			log.Printf("[ERROR] [executor@%s] please remove the local workspace directory(%s) and re-run terraform init -reconfigure\n", c.Dir(), workspacePath)
			return err
		}
		log.Printf("[INFO] [executor@%s] switch back to remote\n", c.Dir())

		var args = []string{"-input=false", "-no-color"}
		for _, b := range backendConfig {
			args = append(args, fmt.Sprintf("-backend-config=%s", b))
		}
		// Run the correct init command depending on whether the remote backend is Terraform Cloud
		if !isBackendTerraformCloud {
			args = append(args, "-reconfigure")
		}

		err = c.Init(ctx, args...)
		if err != nil {
			if supportsStateReplaceProvider && strings.Contains(err.Error(), AcceptableLegacyStateInitError) {
				log.Printf("[INFO] [migrator@%s] ignoring error '%s'; the error is expected when using Terraform with a legacy Terraform state\n", c.Dir(), AcceptableLegacyStateInitError)
			} else {
				log.Printf("[ERROR] [executor@%s] failed to switch back to remote: %s\n", c.Dir(), err)
				log.Printf("[ERROR] [executor@%s] please re-run terraform init -reconfigure\n", c.Dir())
				return err
			}
		}

		return nil
	}

	return switchBackToRemoteFunc, nil
}

// PlanHasChange is a helper method which runs plan and return true only if the plan has change.
func (c *terraformCLI) PlanHasChange(ctx context.Context, state *State, opts ...string) (bool, error) {

	merged := mergeOptions(opts, []string{"-input=false", "-no-color", "-detailed-exitcode"})
	_, err := c.Plan(ctx, state, merged...)
	if err != nil {
		if exitErr, ok := err.(ExitError); ok && exitErr.ExitCode() == 2 {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// mergeOptions merges two lists and returns a new uniq list.
func mergeOptions(a []string, b []string) []string {
	input := append(a, b...)
	m := make(map[string]struct{})
	uniq := make([]string, 0)
	for _, e := range input {
		if _, ok := m[e]; !ok {
			m[e] = struct{}{}
			uniq = append(uniq, e)
		}
	}

	return uniq
}

// writeTempFile writes content to a temporary file and return its file.
func writeTempFile(content []byte) (*os.File, error) {
	tmpfile, err := os.CreateTemp("", "tmp")
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

// getOptionValue returns a value if any element in a list of options has a given prefix.
func getOptionValue(opts []string, prefix string) string {
	for _, opt := range opts {
		if strings.HasPrefix(opt, prefix) {
			return opt[len(prefix):]
		}
	}
	return ""
}
