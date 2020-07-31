package tfmigrate

import (
	"context"
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
