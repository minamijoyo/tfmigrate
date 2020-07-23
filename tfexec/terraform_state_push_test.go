package tfexec

import (
	"context"
	"regexp"
	"testing"
)

func TestTerraformCLIStatePush(t *testing.T) {
	state := NewState([]byte("dummy state"))
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		opts         []string
		ok           bool
	}{
		{
			desc: "state push from memory",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "push", "/path/to/tempfile"},
					argsRe:   regexp.MustCompile(`^terraform state push \S+$`),
					exitCode: 0,
				},
			},
			state: state,
			ok:    true,
		},
		{
			desc: "failed to run terraform state push",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "push", "/path/to/tempfile"},
					argsRe:   regexp.MustCompile(`^terraform state push \S+$`),
					exitCode: 1,
				},
			},
			state: state,
			ok:    false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "push", "-force", "-lock=false", "/path/to/tempfile"},
					argsRe:   regexp.MustCompile(`^terraform state push -force -lock=false \S+$`),
					exitCode: 0,
				},
			},
			state: state,
			opts:  []string{"-force", "-lock=false"},
			ok:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.StatePush(context.Background(), tc.state, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error")
			}
		})
	}
}
