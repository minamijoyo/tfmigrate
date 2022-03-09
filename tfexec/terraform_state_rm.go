package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
)

// StateRm removes resources from state.
// If a state is given, use it for the input state and return a new state.
// Note that if the input state is not given, always return nil state,
// because the terraform state rm command doesn't have -state-out option.
func (c *terraformCLI) StateRm(ctx context.Context, state *State, addresses []string, opts ...string) (*State, error) {
	args := []string{"state", "rm"}

	var tmpState *os.File
	var err error
	if state != nil {
		if hasPrefixOptions(opts, "-state=") {
			return nil, fmt.Errorf("failed to build options. The state argument (!= nil) and the -state= option cannot be set at the same time: state=%v, opts=%v", state, opts)
		}
		tmpState, err = writeTempFile(state.Bytes())
		defer os.Remove(tmpState.Name())
		if err != nil {
			return nil, err
		}
		args = append(args, "-state="+tmpState.Name())
	}

	args = append(args, opts...)

	if len(addresses) > 0 {
		args = append(args, addresses...)
	}

	_, _, err = c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	// The terraform state rm command doesn't have -state-out option,
	// It updates the input state in-place, so we read the updated contents and return it.
	// The interface is a bit inconsistency against the terraform command,
	// but we prefer returning a new state to updating the argument in-place.
	if state != nil {
		stateOut, err := ioutil.ReadFile(tmpState.Name())
		if err != nil {
			return nil, err
		}
		return NewState(stateOut), nil
	}
	// If state == nil, it updates the current default state,
	// we can read it with calling the state pull command,
	// but we avoid invoking it implicitly and just return nil.
	return nil, nil
}
