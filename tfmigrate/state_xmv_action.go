package tfmigrate

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

// NewStateMvAction returns a new StateXMvAction instance.
func NewStateXMvAction(source string, destination string) *StateXMvAction {
	return &StateXMvAction{
		source:      source,
		destination: destination,
	}
}

// StateUpdate updates a given state and returns a new state.
// Source resources have wildcards which should be matched against the tf state. Each occurrence will generate
// a move command.
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

func (a *StateXMvAction) generateMvActions(ctx context.Context, tf tfexec.TerraformCLI, state *tfexec.State) (response []*StateMvAction, err error) {
	stateList, err := tf.StateList(ctx, state, nil)
	if err != nil {
		return nil, err
	}
	return a.getStateMvActionsForStateList(stateList)
}

// When a wildcardChar is used in a path it should only match a single part of the path
// It can therefore not contain a dot(.), whitespace nor square brackets
const matchWildcardRegex = "(.*)"
const wildcardChar = "*"

func (a *StateXMvAction) nrOfWildcards() int {
	return strings.Count(a.source, wildcardChar)
}

// Return regex pattern that matches the wildcard source and make sure characters are not treated as
// special meta characters.
func makeSourceMatchPattern(s string) string {
	safeString := regexp.QuoteMeta(s)
	quotedWildCardChar := regexp.QuoteMeta(wildcardChar)
	return strings.ReplaceAll(safeString, quotedWildCardChar, matchWildcardRegex)
}

// Get a regex that will do matching based on the wildcard source that was given.
func makeSrcRegex(source string) (*regexp.Regexp, error) {
	regPattern := makeSourceMatchPattern(source)
	regExpression, err := regexp.Compile(regPattern)
	if err != nil {
		return nil, fmt.Errorf("could not make pattern out of %s (%s) due to %s", source, regPattern, err)
	}
	return regExpression, nil
}

// Look into the state and find sources that match pattern with wild cards.
func (a *StateXMvAction) getMatchingSourcesFromState(stateList []string) (matchingStateSources []string, err error) {
	r, err := makeSrcRegex(a.source)
	if err != nil {
		return nil, err
	}

	for _, s := range stateList {
		match := r.FindString(s)
		if match != "" {
			matchingStateSources = append(matchingStateSources, match)
		}
	}
	return matchingStateSources, err
}

// When you have the stateXMvAction with wildcards get the destination for a source
func (a *StateXMvAction) getDestinationForStateSrc(stateSource string) (destination string, err error) {
	r, err := makeSrcRegex(a.source)
	if err != nil {
		return "", err
	}
	destination = r.ReplaceAllString(stateSource, a.destination)
	return destination, err
}

// Get actions matching wildcard move actions based on the list of resources.
func (a *StateXMvAction) getStateMvActionsForStateList(stateList []string) (response []*StateMvAction, err error) {
	if a.nrOfWildcards() == 0 {
		response = make([]*StateMvAction, 1)
		response[0] = NewStateMvAction(a.source, a.destination)
		return response, nil
	}
	matchingSources, err := a.getMatchingSourcesFromState(stateList)
	if err != nil {
		return nil, err
	}
	response = make([]*StateMvAction, len(matchingSources))
	for i, matchingSource := range matchingSources {
		destination, e2 := a.getDestinationForStateSrc(matchingSource)
		if e2 != nil {
			return nil, e2
		}
		response[i] = NewStateMvAction(matchingSource, destination)
	}
	return response, nil
}
