package tfexec

import (
	"context"
	"testing"
)

func TestTerraformCLIStatePull(t *testing.T) {
	tfstate := `{
  "version": 4,
  "terraform_version": "0.12.28",
  "serial": 0,
  "lineage": "3d2cf549-8051-c117-aaa7-f93cda2674e8",
  "outputs": {},
  "resources": []
}
`
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		want         State
		ok           bool
	}{
		{
			desc: "print tfstate to stdout",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "pull"},
					stdout:   tfstate,
					exitCode: 0,
				},
			},
			want: State(tfstate),
			ok:   true,
		},
		{
			desc: "failed to run terraform state pull",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "pull"},
					exitCode: 1,
				},
			},
			want: State(""),
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.StatePull(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}
