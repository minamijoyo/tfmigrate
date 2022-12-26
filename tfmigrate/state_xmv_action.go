package tfmigrate

import (
	"context"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// StateXMvAction implements the StateAction interface.
// StateXMvAction moves a resource from source address to destination address in
// the same tfstate file.
type StateXMvAction struct {
	// source is a address of resource or module to be moved which can contain wildcards.
	source string
	// destination is a new address of resource or module to move which can contain placeholders.
	destination string
}

var _ StateAction = (*StateXMvAction)(nil)

// NewStateXMvAction returns a new StateXMvAction instance.
func NewStateXMvAction(source string, destination string) *StateXMvAction {
	return &StateXMvAction{
		source:      source,
		destination: destination,
	}
}

// StateUpdate updates a given state and returns a new state.
// Source resources have wildcards which should be matched against the tf state.
// Each occurrence will generate a move command.
func (a *StateXMvAction) StateUpdate(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) (*tfexec.State, error) {
	stateMvActions, err := a.generateMvActions(ctx, tf, state)
	if err != nil {
		return nil, err
	}

	for _, action := range stateMvActions {
		state, err = action.StateUpdate(ctx, tf, state)
		if err != nil {
			return nil, err
		}
	}
	return state, err
}

// Use an xmv and use the state to determine the corresponding mv actions.
func (a *StateXMvAction) generateMvActions(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) ([]*StateMvAction, error) {
	stateList, err := tf.StateList(ctx, state, nil)
	if err != nil {
		return nil, err
	}

	e := newXMvExpander(a)
	return e.expand(stateList)
}
