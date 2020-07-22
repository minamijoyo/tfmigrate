package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
)

// StateMv moves resources from source to destination address.
// If a state is given, use it for the input state.
func (c *terraformCLI) StateMv(ctx context.Context, state *State, source string, destination string, opts ...string) (*State, error) {
	args := []string{"state", "mv"}

	if state != nil {
		if hasPrefixOptions(opts, "-state=") {
			return nil, fmt.Errorf("failed to build options. The state argument (!= nil) and the -state= option cannot be set at the same time: state=%v, opts=%v", state, opts)
		}
		tmpState, err := writeTempFile(state.Bytes())
		defer os.Remove(tmpState.Name())
		if err != nil {
			return nil, err
		}
		args = append(args, "-state="+tmpState.Name())
	}

	// disallow -state-out option for writing a state file to a temporary file and load it to memory
	if hasPrefixOptions(opts, "-state-out=") {
		return nil, fmt.Errorf("failed to build options. The -state-out= option is not allowed. Read a return value: %v", opts)
	}

	tmpStateOut, err := ioutil.TempFile("", "tfstate")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary state out file: %s", err)
	}
	defer os.Remove(tmpStateOut.Name())

	if err := tmpStateOut.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary state out file: %s", err)
	}
	args = append(args, "-state-out="+tmpStateOut.Name())

	args = append(args, opts...)
	args = append(args, source, destination)

	_, _, err = c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	stateOut, err := ioutil.ReadFile(tmpStateOut.Name())
	if err != nil {
		return nil, err
	}
	return NewState(stateOut), nil
}
