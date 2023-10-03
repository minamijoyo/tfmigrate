package tfmigrate

import (
	"context"
	"log"
	"strings"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// Migrator abstracts migration operations.
type Migrator interface {
	// Plan computes a new state by applying state migration operations to a temporary state.
	// It will fail if terraform plan detects any diffs with the new state.
	Plan(ctx context.Context) error

	// Apply computes a new state and pushes it to remote state.
	// It will fail if terraform plan detects any diffs with the new state.
	// This is intended for solely state refactoring.
	// Any state migration operations should not break any real resources.
	Apply(ctx context.Context) error
}

// setupWorkDir is a common helper function to set up work dir and returns the
// current state and a switch back function.
func setupWorkDir(ctx context.Context, tf tfexec.TerraformCLI, workspace string, isBackendTerraformCloud bool, backendConfig []string, ignoreLegacyStateInitErr bool, isLocal bool) (*tfexec.State, func() error, error) {
	// check if terraform command is available.
	version, err := tf.Version(ctx)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[INFO] [migrator@%s] terraform version: %s\n", tf.Dir(), version)

	supportsStateReplaceProvider, constraints, err := tf.SupportsStateReplaceProvider(ctx)
	if err != nil {
		return nil, nil, err
	}

	// init folder
	log.Printf("[INFO] [migrator@%s] initialize work dir\n", tf.Dir())
	err = tf.Init(ctx, "-input=false", "-no-color")
	if err != nil {
		if supportsStateReplaceProvider && ignoreLegacyStateInitErr && strings.Contains(err.Error(), tfexec.AcceptableLegacyStateInitError) {
			log.Printf("[INFO] [migrator@%s] ignoring error '%s' initilizing work dir; the error is expected when using Terraform %s with a legacy Terraform state\n", tf.Dir(), tfexec.AcceptableLegacyStateInitError, constraints)
		} else {
			return nil, nil, err
		}
	}

	// check current workspace
	currentWorkspace, err := tf.WorkspaceShow(ctx)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[DEBUG] [migrator@%s] currentWorkspace = %s, workspace = %s\n", tf.Dir(), currentWorkspace, workspace)
	if currentWorkspace != workspace {
		// switch to workspace
		log.Printf("[INFO] [migrator@%s] switch to remote workspace %s\n", tf.Dir(), workspace)
		err = tf.WorkspaceSelect(ctx, workspace)
		if err != nil {
			return nil, nil, err
		}
	}

	// get the current remote state.
	log.Printf("[INFO] [migrator@%s] get the current remote state\n", tf.Dir())
	currentState, err := tf.StatePull(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !isLocal {
		// override backend to local
		log.Printf("[INFO] [migrator@%s] override backend to local\n", tf.Dir())
		switchBackToRemoteFunc, err := tf.OverrideBackendToLocal(ctx, "_tfmigrate_override.tf", workspace, isBackendTerraformCloud, backendConfig, ignoreLegacyStateInitErr)
		if err != nil {
			return nil, nil, err
		}
		return currentState, switchBackToRemoteFunc, nil
	} else {
		switchBackToRemoteFunc := func() error {
			log.Printf("[INFO] [executor@%s] nothing to override\n", tf.Dir())
			return nil
		}
		return currentState, switchBackToRemoteFunc, nil
	}
}
