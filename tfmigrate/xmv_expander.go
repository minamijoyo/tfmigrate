package tfmigrate

import (
	"fmt"
	"regexp"
	"strings"
)

// xmvExpander is a helper object for implementing wildcard expansion for xmv actions.
type xmvExpander struct {
	// xmv action to be expanded
	action *StateXmvAction
}

// newXMvExpander returns a new xmvExpander instance.
func newXMvExpander(action *StateXmvAction) *xmvExpander {
	return &xmvExpander{
		action: action,
	}
}

// A wildcardChar will greedy match with any character in the resource path.
const matchWildcardRegex = "(.*)"
const wildcardChar = "*"

// makeSourceMatchPattern returns regex pattern that matches the wildcard
// source and make sure characters are not treated as special meta characters.
func makeSourceMatchPattern(s string) string {
	safeString := regexp.QuoteMeta(s)
	quotedWildCardChar := regexp.QuoteMeta(wildcardChar)
	return strings.ReplaceAll(safeString, quotedWildCardChar, matchWildcardRegex)
}

// makeSrcRegex returns a regex that will do matching based on the wildcard
// source that was given.
func makeSrcRegex(source string) (*regexp.Regexp, error) {
	regPattern := makeSourceMatchPattern(source)
	regExpression, err := regexp.Compile(regPattern)
	if err != nil {
		return nil, fmt.Errorf("could not make pattern out of %s (%s) due to %s", source, regPattern, err)
	}
	return regExpression, nil
}

// expand returns actions matching wildcard move actions based on the list of resources.
func (e *xmvExpander) expand(stateList []string) ([]*StateMvAction, error) {
	if e.nrOfWildcards() == 0 {
		staticActionAsList := make([]*StateMvAction, 1)
		staticActionAsList[0] = NewStateMvAction(e.action.source, e.action.destination)
		return staticActionAsList, nil
	}
	matchingSources, err := e.getMatchingSourcesFromState(stateList)
	if err != nil {
		return nil, err
	}
	matchingActions := make([]*StateMvAction, len(matchingSources))
	for i, matchingSource := range matchingSources {
		destination, e2 := e.getDestinationForStateSrc(matchingSource)
		if e2 != nil {
			return nil, e2
		}
		matchingActions[i] = NewStateMvAction(matchingSource, destination)
	}
	return matchingActions, nil
}

// nrOfWildcards counts a number of wildcard characters.
func (e *xmvExpander) nrOfWildcards() int {
	return strings.Count(e.action.source, wildcardChar)
}

// getMatchingSourcesFromState looks into the state and find sources that match
// pattern with wildcards.
func (e *xmvExpander) getMatchingSourcesFromState(stateList []string) ([]string, error) {
	re, err := makeSrcRegex(e.action.source)
	if err != nil {
		return nil, err
	}

	var matchingStateSources []string

	for _, s := range stateList {
		match := re.FindString(s)
		if match != "" {
			matchingStateSources = append(matchingStateSources, match)
		}
	}
	return matchingStateSources, err
}

// getDestinationForStateSrc returns the destination for a source.
func (e *xmvExpander) getDestinationForStateSrc(stateSource string) (string, error) {
	re, err := makeSrcRegex(e.action.source)
	if err != nil {
		return "", err
	}
	destination := re.ReplaceAllString(stateSource, e.action.destination)
	return destination, err
}
