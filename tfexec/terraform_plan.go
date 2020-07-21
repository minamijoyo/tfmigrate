package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
)

// Plan computes expected changes.
// If a state is given, use it for the input state.
func (c *terraformCLI) Plan(ctx context.Context, state *State, dir string, opts ...string) (*Plan, error) {
	args := []string{"plan"}

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

	// disallow -out option for writing a plan file to a temporary file and load it to memory
	if hasPrefixOptions(opts, "-out=") {
		return nil, fmt.Errorf("failed to build options. The -out= option is not allowed. Read a return value: %v", opts)
	}

	tmpPlan, err := ioutil.TempFile("", "tfplan")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary plan file: %s", err)
	}
	defer os.Remove(tmpPlan.Name())

	if err := tmpPlan.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary plan file: %s", err)
	}
	args = append(args, "-out="+tmpPlan.Name())

	args = append(args, opts...)

	if len(dir) > 0 {
		args = append(args, dir)
	}

	_, _, err = c.Run(ctx, args...)

	// terraform plan -detailed-exitcode returns 2 if there is a diff.
	// So we intentionally ignore an error of read the plan file and returns the
	// original error of terraform plan command.
	plan, _ := ioutil.ReadFile(tmpPlan.Name())
	return NewPlan(plan), err
}
