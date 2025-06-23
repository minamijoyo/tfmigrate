package tfmigrate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// MultiStateMigratorConfig is a config for MultiStateMigrator.
type MultiStateMigratorConfig struct {
	// FromDir is a working directory where states of resources move from.
	FromDir string `hcl:"from_dir"`
	// FromSkipPlan controls whether or not to run and analyze Terraform plan
	// within the from_dir.
	FromSkipPlan bool `hcl:"from_skip_plan,optional"`
	// ToDir is a working directory where states of resources move to.
	ToDir string `hcl:"to_dir"`
	// ToSkipPlan controls whether or not to run and analyze Terraform plan
	// within the to_dir.
	ToSkipPlan bool `hcl:"to_skip_plan,optional"`
	// FromWorkspace is a workspace within FromDir
	FromWorkspace string `hcl:"from_workspace,optional"`
	// ToWorkspace is a workspace within ToDir
	ToWorkspace string `hcl:"to_workspace,optional"`
	// Actions is a list of multi state action.
	// Each action is a plain text for state operation.
	// Valid formats are the following.
	// "mv <source> <destination>"
	Actions []string `hcl:"actions"`
	// Force option controls behaviour in case of unexpected diff in plan.
	// When set forces applying even if plan shows diff.
	Force bool `hcl:"force,optional"`
	// FromTfTarget specifies the target parameter for the from_tf plan.
	FromTfTarget string `hcl:"from_tf_target,optional"`
}

// MultiStateMigratorConfig implements a MigratorConfig.
var _ MigratorConfig = (*MultiStateMigratorConfig)(nil)

// NewMigrator returns a new instance of MultiStateMigrator.
func (c *MultiStateMigratorConfig) NewMigrator(o *MigratorOption) (Migrator, error) {
	if len(c.Actions) == 0 {
		return nil, fmt.Errorf("failed to NewMigrator with no actions")
	}

	// build actions from config.
	actions := []MultiStateAction{}
	for _, cmdStr := range c.Actions {
		action, err := NewMultiStateActionFromString(cmdStr)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	// use default workspace if not specified by user
	if len(c.FromWorkspace) == 0 {
		c.FromWorkspace = "default"
	}
	if len(c.ToWorkspace) == 0 {
		c.ToWorkspace = "default"
	}

	// Pass the FromTfTarget to the migrator instance
	return NewMultiStateMigrator(c.FromDir, c.ToDir, c.FromWorkspace, c.ToWorkspace, actions, o, c.Force, c.FromSkipPlan, c.ToSkipPlan, c.FromTfTarget), nil
}

// MultiStateMigrator implements the Migrator interface.
type MultiStateMigrator struct {
	// fromTf is an instance of TerraformCLI which executes terraform command in a fromDir.
	fromTf tfexec.TerraformCLI
	// fromSkipPlan disables the running of Terraform plan in fromDir.
	fromSkipPlan bool
	// fromTf is an instance of TerraformCLI which executes terraform command in a toDir.
	toTf tfexec.TerraformCLI
	// toSkipPlan disables the running of Terraform plan in toDir.
	toSkipPlan bool
	//fromWorkspace is the workspace from which the resource will be migrated
	fromWorkspace string
	//toWorkspace is the workspace to which the resource will be migrated
	toWorkspace string
	// actions is a list of multi state migration operations.
	actions []MultiStateAction
	// o is an option for migrator.
	// It is used for shared settings across Migrator instances.
	o *MigratorOption
	// force operation in case of unexpected diff
	force bool
	// Add FromTfTarget to the MultiStateMigrator struct
	fromTfTarget string
}

var _ Migrator = (*MultiStateMigrator)(nil)

// NewMultiStateMigrator returns a new MultiStateMigrator instance.
func NewMultiStateMigrator(fromDir string, toDir string, fromWorkspace string, toWorkspace string,
	actions []MultiStateAction, o *MigratorOption, force bool, fromSkipPlan bool, toSkipPlan bool, fromTfTarget string) *MultiStateMigrator {
	fromTf := tfexec.NewTerraformCLI(tfexec.NewExecutor(fromDir, os.Environ()))
	toTf := tfexec.NewTerraformCLI(tfexec.NewExecutor(toDir, os.Environ()))
	if o != nil {
		// Set the exec paths based on the options provided
		if len(o.SourceExecPath) > 0 {
			// If source exec path is specified, use it for the from directory
			fromTf.SetExecPath(o.SourceExecPath)
		} else if len(o.ExecPath) > 0 {
			// Otherwise, fall back to the common exec path if provided
			fromTf.SetExecPath(o.ExecPath)
		}

		if len(o.DestinationExecPath) > 0 {
			// If destination exec path is specified, use it for the to directory
			toTf.SetExecPath(o.DestinationExecPath)
		} else if len(o.ExecPath) > 0 {
			// Otherwise, fall back to the common exec path if provided
			toTf.SetExecPath(o.ExecPath)
		}
	}

	return &MultiStateMigrator{
		fromTf:        fromTf,
		fromSkipPlan:  fromSkipPlan,
		toTf:          toTf,
		toSkipPlan:    toSkipPlan,
		fromWorkspace: fromWorkspace,
		toWorkspace:   toWorkspace,
		actions:       actions,
		o:             o,
		force:         force,
		fromTfTarget:  fromTfTarget,
	}
}

// plan computes new states by applying multi state migration operations to temporary states.
// It will fail if terraform plan detects any diffs with at least one new state.
// We intentionally make this method private to avoid exposing internal states and unify
// the Migrator interface between a single and multi state migrator.
func (m *MultiStateMigrator) plan(ctx context.Context) (fromCurrentState *tfexec.State, toCurrentState *tfexec.State, err error) {
	// setup fromDir.
	fromCurrentState, fromSwitchBackToRemoteFunc, err := setupWorkDir(ctx, m.fromTf, m.fromWorkspace, m.o.IsBackendTerraformCloud, m.o.BackendConfig, false)
	if err != nil {
		return nil, nil, err
	}
	// switch back it to remote on exit.
	defer func() {
		err = errors.Join(err, fromSwitchBackToRemoteFunc())
	}()

	// setup toDir.
	toCurrentState, toSwitchBackToRemoteFunc, err := setupWorkDir(ctx, m.toTf, m.toWorkspace, m.o.IsBackendTerraformCloud, m.o.BackendConfig, false)
	if err != nil {
		return nil, nil, err
	}
	// switch back it to remote on exit.
	defer func() {
		err = errors.Join(err, toSwitchBackToRemoteFunc())
	}()

	// computes new states by applying state migration operations to temporary states.
	log.Printf("[INFO] [migrator] compute new states (%s => %s)\n", m.fromTf.Dir(), m.toTf.Dir())
	var fromNewState, toNewState *tfexec.State
	for _, action := range m.actions {
		fromNewState, toNewState, err = action.MultiStateUpdate(ctx, m.fromTf, m.toTf, fromCurrentState, toCurrentState)
		if err != nil {
			return nil, nil, err
		}
		fromCurrentState = tfexec.NewState(fromNewState.Bytes())
		toCurrentState = tfexec.NewState(toNewState.Bytes())
	}

	// build plan options
	planOpts := []string{"-input=false", "-no-color", "-detailed-exitcode"}
	if m.o.PlanOut != "" {
		planOpts = append(planOpts, "-out="+m.o.PlanOut)
	}
	if m.fromTfTarget != "" {
		planOpts = append(planOpts, "-target="+m.fromTfTarget)
	}
	if m.fromSkipPlan {
		log.Printf("[INFO] [migrator@%s] skipping check diffs\n", m.fromTf.Dir())
	} else {
		// check if a plan in fromDir has no changes.
		log.Printf("[INFO] [migrator@%s] check diffs\n", m.fromTf.Dir())
		plan, err := m.fromTf.Plan(ctx, fromCurrentState, planOpts...)
		clean, _ := checkPlan(plan, m.fromTf, err)
		if !clean {
			return nil, nil, fmt.Errorf("terraform plan command returns unexpected diffs in %s from_dir: %s", m.fromTf.Dir(), err)
		}
	}

	if m.toSkipPlan {
		log.Printf("[INFO] [migrator@%s] skipping check diffs\n", m.toTf.Dir())
	} else {
		// check if a plan in toDir has no changes.
		log.Printf("[INFO] [migrator@%s] check diffs\n", m.toTf.Dir())
		plan, err := m.toTf.Plan(ctx, toCurrentState, planOpts...)

		clean, _ := checkPlan(plan, m.toTf, err)
		if !clean {
			if m.force {
				log.Printf("[INFO] [migrator@%s] plan has unexpected diffs, but force option is true, ignoring: %s", m.toTf.Dir(), err)
			} else {
				return nil, nil, fmt.Errorf("terraform plan command returns unexpected diffs in %s to_dir: %s", m.toTf.Dir(), err)
			}
		}
	}

	return fromCurrentState, toCurrentState, err
}

func checkPlan(plan *tfexec.Plan, tf tfexec.TerraformCLI, er error) (bool, error) {
	if er != nil {

		if exitErr, ok := er.(tfexec.ExitError); ok && exitErr.ExitCode() == 2 {
			planJSON, jsonerr := tf.ConvertPlanToJson(plan)
			if jsonerr != nil {
				log.Printf("[ERROR] [migrator] failed to parse plan JSON: %s\n", jsonerr)
				return false, jsonerr
			}

			if !planJSON.HasChanges() {
				log.Printf("[INFO] [migrator] plan has only output changes and irrelvant output changes")
				planJSON.LogResourceChanges()
				return true, nil
			}

			log.Printf("[ERROR] [migrator] unexpected diffs\n")
			planJSON.LogResourceChanges()
			return false, nil
		}
		log.Printf("[ERROR] [migrator] unexpected error: %s\n", er)
		return false, er
	}
	return true, nil
}

// Plan computes new states by applying multi state migration operations to temporary states.
// It will fail if terraform plan detects any diffs with at least one new state.
func (m *MultiStateMigrator) Plan(ctx context.Context) error {
	log.Printf("[INFO] [migrator] multi start state migrator plan\n")
	_, _, err := m.plan(ctx)
	if err != nil {
		return err
	}
	log.Printf("[INFO] [migrator] multi state migrator plan success!\n")
	return nil
}

// Apply computes new states and pushes them to remote states.
// It will fail if terraform plan detects any diffs with at least one new state.
// We are intended to this is used for state refactoring.
// Any state migration operations should not break any real resources.
func (m *MultiStateMigrator) Apply(ctx context.Context) error {
	// Check if new states don't have any diffs compared to real resources
	// before push new states to remote.
	log.Printf("[INFO] [migrator] start multi state migrator plan phase for apply\n")
	fromState, toState, err := m.plan(ctx)
	if err != nil {
		return err
	}

	// push the new states to remote.
	// We push toState before fromState, because when moving resources across
	// states, write them to new state first and then remove them from old one.
	log.Printf("[INFO] [migrator] start multi state migrator apply phase\n")
	log.Printf("[INFO] [migrator@%s] push the new state to remote\n", m.toTf.Dir())
	err = m.toTf.StatePush(ctx, toState)
	if err != nil {
		return err
	}
	log.Printf("[INFO] [migrator@%s] push the new state to remote\n", m.fromTf.Dir())
	err = m.fromTf.StatePush(ctx, fromState)
	if err != nil {
		return err
	}
	log.Printf("[INFO] [migrator] multi state migrator apply success!\n")
	return nil
}
