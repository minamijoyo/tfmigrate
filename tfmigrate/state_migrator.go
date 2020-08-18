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
// It will fail if terraform plan detects any diffs with the new state.
// We intentional private this method not to expose internal states and unify
// the Migrator interface between a single and multi state migrator.
func (m *StateMigrator) plan(ctx context.Context) (*tfexec.State, error) {
	// setup work dir.
	currentState, switchBackToRemotekFunc, err := setupWorkDir(ctx, m.tf)
	if err != nil {
		return nil, err
	}
	// switch back it to remote on exit.
	defer switchBackToRemotekFunc()

	// computes a new state by applying state migration operations to a temporary state.
	log.Printf("[INFO] [migrator@%s] compute a new state\n", m.tf.Dir())
	var newState *tfexec.State
	for _, action := range m.actions {
		newState, err = action.StateUpdate(ctx, m.tf, currentState)
		if err != nil {
			return nil, err
		}
		currentState = tfexec.NewState(newState.Bytes())
	}

	log.Printf("[INFO] [migrator@%s] check diffs\n", m.tf.Dir())
	_, err = m.tf.Plan(ctx, currentState, "", "-input=false", "-no-color", "-detailed-exitcode")
	if err != nil {
		if exitErr, ok := err.(tfexec.ExitError); ok && exitErr.ExitCode() == 2 {
			log.Printf("[ERROR] [migrator@%s] unexpected diffs\n", m.tf.Dir())
			return nil, fmt.Errorf("terraform plan command returns unexpected diffs: %s", err)
		}
		return nil, err
	}

	return currentState, nil
}

// Plan computes a new state by applying state migration operations to a temporary state.
// It will fail if terraform plan detects any diffs with the new state.
func (m *StateMigrator) Plan(ctx context.Context) error {
	log.Printf("[INFO] [migrator] start state migrator plan\n")
	_, err := m.plan(ctx)
	if err != nil {
		return err
	}
	log.Printf("[INFO] [migrator] state migrator plan success!\n")
	return nil
}

// Apply computes a new state and pushes it to remote state.
// It will fail if terraform plan detects any diffs with the new state.
// We are intended to this is used for state refactoring.
// Any state migration operations should not break any real resources.
func (m *StateMigrator) Apply(ctx context.Context) error {
	// Check if a new state does not have any diffs compared to real resources
	// before push a new state to remote.
	log.Printf("[INFO] [migrator] start state migrator plan phase for apply\n")
	state, err := m.plan(ctx)
	if err != nil {
		return err
	}

	// push the new state to remote.
	log.Printf("[INFO] [migrator] start state migrator apply phase\n")
	log.Printf("[INFO] [migrator] push the new state to remote\n")
	err = m.tf.StatePush(ctx, state)
	if err != nil {
		return err
	}
	log.Printf("[INFO] [migrator] state migrator apply success!\n")
	return nil
}
