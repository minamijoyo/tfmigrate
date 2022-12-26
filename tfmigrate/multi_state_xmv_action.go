package tfmigrate

import (
	"context"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

// MultiStateXmvAction implements the MultiStateAction interface.
// MultiStateXmvAction is an extended version of MultiStateMvAction.
// It allows you to move multiple resouces with wildcard matching.
type MultiStateXmvAction struct {
	// source is a address of resource or module to be moved which can contain wildcards.
	source string
	// destination is a new address of resource or module to move which can contain placeholders.
	destination string
}

var _ MultiStateAction = (*MultiStateXmvAction)(nil)

// NewMultiStateXmvAction returns a new MultiStateXmvAction instance.
func NewMultiStateXmvAction(source string, destination string) *MultiStateXmvAction {
	return &MultiStateXmvAction{
		source:      source,
		destination: destination,
	}
}

// MultiStateUpdate updates given two states and returns new two states.
// It moves a resource from a dir to another.
// It also can rename an address of resource.
func (a *MultiStateXmvAction) MultiStateUpdate(ctx context.Context, fromTf tfexec.TerraformCLI, toTf tfexec.TerraformCLI, fromState *tfexec.State, toState *tfexec.State) (*tfexec.State, *tfexec.State, error) {
	multiStateMvActions, err := a.generateMvActions(ctx, fromTf, fromState)
	if err != nil {
		return nil, nil, err
	}

	for _, action := range multiStateMvActions {
		fromState, toState, err = action.MultiStateUpdate(ctx, fromTf, toTf, fromState, toState)
		if err != nil {
			return nil, nil, err
		}
	}
	return fromState, toState, nil
}

// generateMvActions uses an xmv and use the state to determine the corresponding mv actions.
func (a *MultiStateXmvAction) generateMvActions(ctx context.Context, fromTf tfexec.TerraformCLI, fromState *tfexec.State) ([]*MultiStateMvAction, error) {
	stateList, err := fromTf.StateList(ctx, fromState, nil)
	if err != nil {
		return nil, err
	}

	// create a temporary single state mv actions.
	// It may look a bit strange as a type.
	// This is only because sharing the logic while maintaining consistency.
	stateXmv := NewStateXmvAction(a.source, a.destination)

	e := newXmvExpander(stateXmv)
	stateMvActions, err := e.expand(stateList)
	if err != nil {
		return nil, err
	}

	// convert StateMvAction to MultiStateMvAction.
	multiStateMvActions := []*MultiStateMvAction{}
	for _, action := range stateMvActions {
		multiStateMvActions = append(multiStateMvActions, NewMultiStateMvAction(action.source, action.destination))
	}

	return multiStateMvActions, nil
}
