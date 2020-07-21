package tfexec

import (
	"context"
	"testing"
)

func TestTerraformCLIApply(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		dirOrPlan    string
		opts         []string
		ok           bool
	}{
		{
			desc: "no dir and no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply"},
					exitCode: 0,
				},
			},
			ok: true,
		},
		{
			desc: "failed to run terraform apply",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "foo"},
					exitCode: 0,
				},
			},
			dirOrPlan: "foo",
			ok:        true,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "-input=false", "-no-color"},
					exitCode: 0,
				},
			},
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
		{
			desc: "with dir and opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "-input=false", "-no-color", "foo"},
					exitCode: 0,
				},
			},
			dirOrPlan: "foo",
			opts:      []string{"-input=false", "-no-color"},
			ok:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.Apply(context.Background(), tc.dirOrPlan, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}
