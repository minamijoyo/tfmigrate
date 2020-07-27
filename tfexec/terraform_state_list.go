package tfexec

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// StateList shows a list of resources.
// If a state is given, use it for the input state.
func (c *terraformCLI) StateList(ctx context.Context, state *State, addresses []string, opts ...string) ([]string, error) {
	args := []string{"state", "list"}

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

	args = append(args, opts...)

	if len(addresses) > 0 {
		args = append(args, addresses...)
	}

	stdout, _, err := c.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	// we want to split stdout by '\n', but strings.Split returns []string{""} if stdout is empty.
	// we should remove empty strings from the list so that its length to be 0.
	resources := strings.FieldsFunc(
		strings.TrimRight(stdout, "\n"),
		func(c rune) bool {
			return c == '\n'
		},
	)
	return resources, nil
}
