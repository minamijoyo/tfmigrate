package tfexec

import (
	"context"
	"fmt"
	"os"
)

// StateMv moves resources from source to destination address.
// If a state argument is given, use it for the input state.
// If a stateOut argument is given, move resources from state to stateOut.
// It returns updated the given state and the stateOut.
func (c *terraformCLI) StateMv(ctx context.Context, state *State, stateOut *State, source string, destination string, opts ...string) (*State, *State, error) {
	args := []string{"state", "mv"}

	var tmpState *os.File
	var tmpStateOut *os.File
	var err error

	if state != nil {
		if hasPrefixOptions(opts, "-state=") {
			return nil, nil, fmt.Errorf("failed to build options. The state argument (!= nil) and the -state= option cannot be set at the same time: state=%v, opts=%v", state, opts)
		}
		tmpState, err = writeTempFile(state.Bytes())
		defer os.Remove(tmpState.Name())
		if err != nil {
			return nil, nil, err
		}
		args = append(args, "-state="+tmpState.Name())
	}

	if stateOut != nil {
		if hasPrefixOptions(opts, "-state-out=") {
			return nil, nil, fmt.Errorf("failed to build options. The stateOut argument (!= nil) and the -state-out= option cannot be set at the same time: stateOut=%v, opts=%v", stateOut, opts)
		}
		tmpStateOut, err = writeTempFile(stateOut.Bytes())
		defer os.Remove(tmpStateOut.Name())
		if err != nil {
			return nil, nil, err
		}
		args = append(args, "-state-out="+tmpStateOut.Name())
	}

	args = append(args, opts...)
	args = append(args, source, destination)

	_, _, err = c.Run(ctx, args...)
	if err != nil {
		return nil, nil, err
	}

	// Read updated states
	var updatedState *State
	var updatedStateOut *State

	if state != nil {
		bytes, err := os.ReadFile(tmpState.Name())
		if err != nil {
			return nil, nil, err
		}
		updatedState = NewState(bytes)
	}

	if stateOut != nil {
		bytes, err := os.ReadFile(tmpStateOut.Name())
		if err != nil {
			return nil, nil, err
		}
		updatedStateOut = NewState(bytes)
	}

	return updatedState, updatedStateOut, nil
}
