package tfmigrate

import (
	"context"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// StateXmvAction implements the StateAction interface.
// StateXmvAction is an extended version of StateMvAction.
// It allows you to move multiple resouces with a wildcard match.
type StateXmvAction struct {
	// source is a address of resource or module to be moved which can contain wildcards.
	source string
	// destination is a new address of resource or module to move which can contain placeholders.
	destination string
}

var _ StateAction = (*StateXmvAction)(nil)

// NewStateXmvAction returns a new StateXmvAction instance.
func NewStateXmvAction(source string, destination string) *StateXmvAction {
	return &StateXmvAction{
		source:      source,
		destination: destination,
	}
}

// StateUpdate updates a given state and returns a new state.
// Source resources have wildcards which should be matched against the tf state.
// Each occurrence will generate a move command.
func (a *StateXmvAction) StateUpdate(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) (*tfexec.State, error) {
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

// generateMvActions uses an xmv and use the state to determine the corresponding mv actions.
func (a *StateXmvAction) generateMvActions(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) ([]*StateMvAction, error) {
	stateList, err := tf.StateList(ctx, state, nil)
	if err != nil {
		return nil, err
	}

	e := newXMvExpander(a)
	return e.expand(stateList)
}
