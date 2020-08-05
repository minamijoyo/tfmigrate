package tfmigrate

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// StateMigrator implements the Migrator interface.
type StateMigrator struct {
	// tf is an instance of TerraformCLI.
	tf tfexec.TerraformCLI
	// actions is a list of state migration operations.
	actions []StateAction
}

var _ Migrator = (*StateMigrator)(nil)

// NewStateMigrator returns a new StateMigrator instance.
func NewStateMigrator(dir string, actions []StateAction, o *MigratorOption) *StateMigrator {
	e := tfexec.NewExecutor(dir, os.Environ())
	tf := tfexec.NewTerraformCLI(e)
	if o != nil && len(o.ExecPath) > 0 {
		tf.SetExecPath(o.ExecPath)
	}

	return &StateMigrator{
		tf:      tf,
		actions: actions,
	}
}

// plan computes a new state by applying state migration operations to a temporary state.
// It will fail if terraform plan detects any diffs with a new state.
// We intentional private this method not to expose an internal state type to
// outside of this package.
func (m *StateMigrator) plan(ctx context.Context) (*tfexec.State, error) {
	// check if terraform command is available.
	version, err := m.tf.Version(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] terraform version: %s\n", version)

	// initialize work dir.
	err = m.tf.Init(ctx, "", "-input=false", "-no-color")
	if err != nil {
		return nil, err
	}

	// get the current remote state.
	currentState, err := m.tf.StatePull(ctx)
	if err != nil {
		return nil, err
	}

	// The -state flag for terraform command is not valid for remote state,
	// so we need to switch the backend to local for temporary state operations.
	switchBackToRemotekFunc, err := m.tf.OverrideBackendToRemote(ctx, "_tfmigrate_override.tf")
	if err != nil {
		return nil, err
	}

	// switch back it to remote on exit.
	defer switchBackToRemotekFunc()

	// computes a new state by applying state migration operations to a temporary state.
	var newState *tfexec.State
	for _, action := range m.actions {
		newState, err = action.StateUpdate(ctx, m.tf, currentState)
		if err != nil {
			return nil, err
		}
		currentState = tfexec.NewState(newState.Bytes())
	}

	_, err = m.tf.Plan(ctx, currentState, "", "-input=false", "-no-color", "-detailed-exitcode")
	if err != nil {
		if exitErr, ok := err.(tfexec.ExitError); ok && exitErr.ExitCode() == 2 {
			return nil, fmt.Errorf("terraform plan command returns unexpected diffs: %s", err)
		}
		return nil, err
	}

	return currentState, nil
}

// Plan computes a new state by applying state migration operations to a temporary state.
// It will fail if terraform plan detects any diffs with a new state.
func (m *StateMigrator) Plan(ctx context.Context) error {
	_, err := m.plan(ctx)
	return err
}

// Apply computes a new state and push it to remote state.
// It will fail if terraform plan detects any diffs with a new state.
// We are intended to this is used for state refactoring.
// Any state migration operations should not break any real resources.
func (m *StateMigrator) Apply(ctx context.Context) error {
	// Check if a new state does not have any diffs compared to real resources
	// before push a new state to remote.
	state, err := m.plan(ctx)
	if err != nil {
		return err
	}

	// push the new state to remote.
	return m.tf.StatePush(ctx, state)
}
