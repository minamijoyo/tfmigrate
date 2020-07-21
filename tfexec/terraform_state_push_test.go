package tfexec

import (
	"context"
	"regexp"
	"testing"
)

func TestTerraformCLIStatePush(t *testing.T) {
	state := NewState([]byte(`{
  "version": 4,
  "terraform_version": "0.12.28",
  "serial": 0,
  "lineage": "3d2cf549-8051-c117-aaa7-f93cda2674e8",
  "outputs": {},
  "resources": []
}
`))
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		state        *State
		ok           bool
	}{
		{
			desc: "state push from memory",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "push", "/path/to/tempfile"},
					argsRe:   regexp.MustCompile(`^terraform state push .+$`),
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
					argsRe:   regexp.MustCompile(`^terraform state push .+$`),
					exitCode: 1,
				},
			},
			state: state,
			ok:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.StatePush(context.Background(), tc.state)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error")
			}
		})
	}
}
