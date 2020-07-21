package tfexec

import (
	"context"
	"regexp"
	"testing"
)

func TestTerraformCLIApply(t *testing.T) {
	plan := NewPlan([]byte("dummy plan"))

	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		plan         *Plan
		dir          string
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
			dir: "foo",
			ok:  true,
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
			dir:  "foo",
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
		{
			desc: "with plan",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "-input=false", "-no-color", "/path/to/planfile"},
					argsRe:   regexp.MustCompile(`^terraform apply -input=false -no-color \S+$`),
					exitCode: 0,
				},
			},
			plan: plan,
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
		{
			desc: "with plan and dir (conflict error)",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "apply", "-input=false", "-no-color", "/path/to/planfile", "foo"},
					argsRe:   regexp.MustCompile(`^terraform apply -input=false -no-color \S+ foo$`),
					exitCode: 1,
				},
			},
			plan: plan,
			dir:  "foo",
			opts: []string{"-input=false", "-state=foo.tfstate"},
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.Apply(context.Background(), tc.plan, tc.dir, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}
