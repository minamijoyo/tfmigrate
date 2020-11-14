package tfmigrate

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// StateMigratorConfig is a config for StateMigrator.
type StateMigratorConfig struct {
	// Dir is a working directory for executing terraform command.
	// Default to `.` (current directory).
	Dir string `hcl:"dir,optional"`
	// Actions is a list of state action.
	// action is a plain text for state operation.
	// Valid formats are the following.
	// "mv <source> <destination>"
	// "rm <addresses>...
	// "import <address> <id>"
	// We could define strict block schema for action, but intentionally use a
	// schema-less string to allow us to easily copy terraform state command to
	// action.
	Actions []string `hcl:"actions"`
}

// StateMigratorConfig implements a MigratorConfig.
var _ MigratorConfig = (*StateMigratorConfig)(nil)

// NewMigrator returns a new instance of StateMigrator.
func (c *StateMigratorConfig) NewMigrator(o *MigratorOption) (Migrator, error) {
	// default working directory
	dir := "."
	if len(c.Dir) > 0 {
		dir = c.Dir
	}

	if len(c.Actions) == 0 {
		return nil, fmt.Errorf("faild to NewMigrator with no actions")
	}

	// build actions from config.
	actions := []StateAction{}
	for _, cmdStr := range c.Actions {
		action, err := NewStateActionFromString(cmdStr)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return NewStateMigrator(dir, actions, o), nil
}

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
