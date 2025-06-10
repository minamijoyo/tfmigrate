package tfexec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Plan computes expected changes.
// If a state is given, use it for the input state.
func (c *terraformCLI) Plan(ctx context.Context, state *State, opts ...string) (*Plan, error) {
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

	// To return a plan file as a return value, we always use an -out option and load it to memory.
	// if the option exists just use it else create a temporary file.
	planOut := ""
	if hasPrefixOptions(opts, "-out=") {
		planOut = getOptionValue(opts, "-out=")
	} else {
		tmpPlan, err := os.CreateTemp("", "tfplan")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary plan file: %s", err)
		}
		planOut = tmpPlan.Name()
		defer os.Remove(planOut)

		if err := tmpPlan.Close(); err != nil {
			return nil, fmt.Errorf("failed to close temporary plan file: %s", err)
		}
		args = append(args, "-out="+planOut)
	}

	args = append(args, opts...)

	_, _, err := c.Run(ctx, args...)

	// terraform plan -detailed-exitcode returns 2 if there is a diff.
	// So we intentionally ignore an error of read the plan file and returns the
	// original error of terraform plan command.
	plan, _ := os.ReadFile(planOut)
	return NewPlan(plan), err
}

func (c *terraformCLI) ConvertPlanToJson(plan *Plan) (*TerraformPlanJSON, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan is nil")
	}

	if len(plan.Bytes()) == 0 {
		return nil, fmt.Errorf("plan is empty")
	}
	tmpPlan, _ := os.CreateTemp("", "tfplan")
	defer os.Remove(tmpPlan.Name())

	err := os.WriteFile(tmpPlan.Name(), plan.Bytes(), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write plan to temporary file: %w", err)
	}
	args := []string{"show", "-json", tmpPlan.Name()}

	output, _, err := c.Run(context.Background(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run terraform show -json: %w", err)
	}
	var planJSON TerraformPlanJSON

	if err := json.Unmarshal([]byte(output), &planJSON); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}
	return &planJSON, nil

}
