package tfmigrate

import (
	"context"
	"errors"
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
	// Force option controls behaviour in case of unexpected diff in plan.
	// When set forces applying even if plan shows diff.
	Force bool `hcl:"force,optional"`
	// SkipPlan controls whether or not to run and analyze Terraform plan.
	SkipPlan bool `hcl:"skip_plan,optional"`
	// Workspace is the state workspace which the migration works with.
	Workspace string `hcl:"workspace,optional"`
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
		return nil, fmt.Errorf("failed to NewMigrator with no actions")
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

	//use default workspace if not specified by user
	if len(c.Workspace) == 0 {
		c.Workspace = "default"
	}

	return NewStateMigrator(dir, c.Workspace, actions, o, c.Force, c.SkipPlan), nil
}

// StateMigrator implements the Migrator interface.
type StateMigrator struct {
	// tf is an instance of TerraformCLI.
	tf tfexec.TerraformCLI
	// actions is a list of state migration operations.
	actions []StateAction
	// o is an option for migrator.
	// It is used for shared settings across Migrator instances.
	o *MigratorOption
	// skipPlan controls whether or not to run and analyze Terraform plan.
	skipPlan bool
	// force operation in case of unexpected diff
	force bool
	// workspace is the state workspace which the migration works with.
	workspace string
}

var _ Migrator = (*StateMigrator)(nil)

// NewStateMigrator returns a new StateMigrator instance.
func NewStateMigrator(dir string, workspace string, actions []StateAction,
	o *MigratorOption, force bool, skipPlan bool) *StateMigrator {
	e := tfexec.NewExecutor(dir, os.Environ())
	tf := tfexec.NewTerraformCLI(e)
	if o != nil && len(o.ExecPath) > 0 {
		// While NewTerraformCLI reads the environment variable TFMIGRATE_EXEC_PATH
		// at initialization, the MigratorOption takes precedence over it.
		tf.SetExecPath(o.ExecPath)
	}

	return &StateMigrator{
		tf:        tf,
		actions:   actions,
		o:         o,
		force:     force,
		skipPlan:  skipPlan,
		workspace: workspace,
	}
}

// plan computes a new state by applying state migration operations to a temporary state.
// It will fail if terraform plan detects any diffs with the new state.
// We intentionally keep this method private as to not expose internal states and unify
// the Migrator interface between a single and multi state migrator.
func (m *StateMigrator) plan(ctx context.Context) (currentState *tfexec.State, err error) {
	ignoreLegacyStateInitErr := false
	for _, action := range m.actions {
		// When invoking `state replace-provider`, it's necessary to first
		// invoke `terraform init`. However, when using a non-legacy Terraform CLI
		// against a legacy Terraform state, `terraform init` returns an error.
		// In this scenario, we can (and must) safely ignore the error.
		if _, ok := action.(*StateReplaceProviderAction); ok {
			ignoreLegacyStateInitErr = true
			break
		}
	}

	// setup work dir.
	currentState, switchBackToRemoteFunc, err := setupWorkDir(ctx, m.tf, m.workspace, m.o.IsBackendTerraformCloud, m.o.BackendConfig, ignoreLegacyStateInitErr)
	if err != nil {
		return nil, err
	}

	// switch back it to remote on exit.
	defer func() {
		err = errors.Join(err, switchBackToRemoteFunc())
	}()

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

	// build plan options
	planOpts := []string{"-input=false", "-no-color", "-detailed-exitcode"}
	if m.o.PlanOut != "" {
		planOpts = append(planOpts, "-out="+m.o.PlanOut)
	}

	if m.skipPlan {
		log.Printf("[INFO] [migrator@%s] skipping check diffs\n", m.tf.Dir())
	} else {
		log.Printf("[INFO] [migrator@%s] check diffs\n", m.tf.Dir())
		_, err = m.tf.Plan(ctx, currentState, planOpts...)
		if err != nil {
			if exitErr, ok := err.(tfexec.ExitError); ok && exitErr.ExitCode() == 2 {
				if !m.force {
					log.Printf("[ERROR] [migrator@%s] unexpected diffs\n", m.tf.Dir())
					return nil, fmt.Errorf("terraform plan command returns unexpected diffs: %s", err)
				}
				log.Printf("[INFO] [migrator@%s] unexpected diffs, ignoring as force option is true: %s", m.tf.Dir(), err)
				// reset err to nil to intentionally ignore unexpected diffs.
				err = nil
			} else {
				return nil, err
			}
		}
	}

	return currentState, err
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
