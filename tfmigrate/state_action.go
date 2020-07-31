package tfmigrate

import (
	"context"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// StateAction abstracts state migration operations.
type StateAction interface {
	// StateUpdate updates a given state and returns a new state.
	StateUpdate(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) (*tfexec.State, error)
}

// StateMvAction implements the StateAction interface.
// StateMvAction moves a resource from source address to destination address in
// the same tfstate file.
type StateMvAction struct {
	// source is a address of resource or module to be moved.
	source string
	// // destination is a new address of resource or module to move.
	destination string
}

var _ StateAction = (*StateMvAction)(nil)

// NewStateMvAction returns a new StateMvAction instance.
func NewStateMvAction(source string, destination string) *StateMvAction {
	return &StateMvAction{
		source:      source,
		destination: destination,
	}
}

// StateUpdate updates a given state and returns a new state.
// It moves a resource from source address to destination address in the same tfstate file.
func (a *StateMvAction) StateUpdate(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) (*tfexec.State, error) {
	newState, _, err := tf.StateMv(ctx, state, nil, a.source, a.destination)
	return newState, err
}
