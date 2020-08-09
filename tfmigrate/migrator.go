package tfmigrate

import (
	"context"
	"log"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// Migrator abstracts migration operations.
type Migrator interface {
	// Plan computes a new state by applying state migration operations to a temporary state.
	// It will fail if terraform plan detects any diffs with a new state.
	Plan(ctx context.Context) error

	// Apply computes a new state and push it to remote state.
	// It will fail if terraform plan detects any diffs with a new state.
	// We are intended to this is used for state refactoring.
	// Any state migration operations should not break any real resources.
	Apply(ctx context.Context) error
}

// MigratorOption customizes a behaviror of Migrator.
// It is used for shared settings across Migrator instances.
type MigratorOption struct {
	// ExecPath is a string how terraform command is executed. Default to terraform.
	// It's intended to inject a wrapper command such as direnv.
	ExecPath string
}

// setupWorkDir is a common helper function to setup work dir and returns the
// current state and a swtich back function.
func setupWorkDir(ctx context.Context, tf tfexec.TerraformCLI) (*tfexec.State, func(), error) {
	// check if terraform command is available.
	version, err := tf.Version(ctx)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[DEBUG] [%s] terraform version: %s\n", tf.Dir(), version)

	// initialize work dir.
	err = tf.Init(ctx, "", "-input=false", "-no-color")
	if err != nil {
		return nil, nil, err
	}

	// get the current remote state.
	currentState, err := tf.StatePull(ctx)
	if err != nil {
		return nil, nil, err
	}

	// The -state flag for terraform command is not valid for remote state,
	// so we need to switch the backend to local for temporary state operations.
	switchBackToRemotekFunc, err := tf.OverrideBackendToLocal(ctx, "_tfmigrate_override.tf")
	if err != nil {
		return nil, nil, err
	}

	return currentState, switchBackToRemotekFunc, nil
}
