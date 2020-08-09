package tfmigrate

import (
	"context"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// MultiStateMvAction implements the MultiStateAction interface.
// MultiStateMvAction moves a resource from a dir to another.
// It also can rename an address of resource.
type MultiStateMvAction struct {
	// source is a address of resource or module to be moved.
	source string
	// // destination is a new address of resource or module to move.
	destination string
}

var _ MultiStateAction = (*MultiStateMvAction)(nil)

// NewMultiStateMvAction returns a new MultiStateMvAction instance.
func NewMultiStateMvAction(source string, destination string) *MultiStateMvAction {
	return &MultiStateMvAction{
		source:      source,
		destination: destination,
	}
}

// MultiStateUpdate updates given two states and returns new two states.
// It moves a resource from a dir to another.
// It also can rename an address of resource.
func (a *MultiStateMvAction) MultiStateUpdate(ctx context.Context, fromTf tfexec.TerraformCLI, toTf tfexec.TerraformCLI, fromState *tfexec.State, toState *tfexec.State) (*tfexec.State, *tfexec.State, error) {
	// move a resource from fromState to a temporary diffState.
	diffState := tfexec.NewState([]byte{})
	fromNewState, diffNewState, err := fromTf.StateMv(ctx, fromState, diffState, a.source, a.source, "-backup=/dev/null")
	if err != nil {
		return nil, nil, err
	}

	// move the resource from the diffState to toState.
	_, toNewState, err := toTf.StateMv(ctx, diffNewState, toState, a.source, a.destination, "-backup=/dev/null")
	if err != nil {
		return nil, nil, err
	}

	return fromNewState, toNewState, nil
}
