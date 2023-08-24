package tfmigrate

import (
	"context"
	"fmt"
	"log"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// MockMigratorConfig is a config for MockMigrator.
type MockMigratorConfig struct {
	// PlanError is a flag to return an error on Plan().
	PlanError bool `hcl:"plan_error"`
	// ApplyError is a flag to return an error on Apply().
	ApplyError bool `hcl:"apply_error"`
}

// MockMigratorConfig implements a MigratorConfig.
var _ MigratorConfig = (*MockMigratorConfig)(nil)

// NewMigrator returns a new instance of MockMigrator.
func (c *MockMigratorConfig) NewMigrator(_ *MigratorOption) (Migrator, error) {
	return NewMockMigrator(c.PlanError, c.ApplyError), nil
}

// MockMigrator implements the Migrator interface for testing.
// It does nothing, but can return an error.
type MockMigrator struct {
	// planError is a flag to return an error on Plan().
	planError bool
	// applyError is a flag to return an error on Apply().
	applyError bool
}

var _ Migrator = (*MockMigrator)(nil)

// NewMockMigrator returns a new MockMigrator instance.
func NewMockMigrator(planError bool, applyError bool) *MockMigrator {
	return &MockMigrator{
		planError:  planError,
		applyError: applyError,
	}
}

// plan computes a new state by applying state migration operations to a temporary state.
// It does nothing, but can return an error.
func (m *MockMigrator) plan(_ context.Context) (*tfexec.State, error) {
	if m.planError {
		return nil, fmt.Errorf("failed to plan mock migrator: planError = %t", m.planError)
	}
	return nil, nil
}

// Plan computes a new state by applying state migration operations to a temporary state.
// It does nothing, but can return an error.
func (m *MockMigrator) Plan(ctx context.Context) error {
	log.Printf("[INFO] [migrator] start state migrator plan\n")
	_, err := m.plan(ctx)
	if err != nil {
		return err
	}
	log.Printf("[INFO] [migrator] state migrator plan success!\n")
	return nil
}

// Apply computes a new state and pushes it to remote state.
// It does nothing, but can return an error.
func (m *MockMigrator) Apply(ctx context.Context) error {
	log.Printf("[INFO] [migrator] start state migrator plan phase for apply\n")
	_, err := m.plan(ctx)
	if err != nil {
		return err
	}

	log.Printf("[INFO] [migrator] start state migrator apply phase\n")
	if m.applyError {
		return fmt.Errorf("failed to apply mock migrator: applyError = %t", m.applyError)
	}
	log.Printf("[INFO] [migrator] state migrator apply success!\n")
	return nil
}
