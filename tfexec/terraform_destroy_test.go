package tfexec

import (
	"context"
	"testing"
)

func TestTerraformCLIDestroy(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		dir          string
		opts         []string
		ok           bool
	}{
		{
			desc: "no dir and no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "destroy"},
					exitCode: 0,
				},
			},
			ok: true,
		},
		{
			desc: "failed to run terraform destroy",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "destroy"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "destroy", "foo"},
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
					args:     []string{"terraform", "destroy", "-input=false", "-no-color"},
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
					args:     []string{"terraform", "destroy", "-input=false", "-no-color", "foo"},
					exitCode: 0,
				},
			},
			dir:  "foo",
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.Destroy(context.Background(), tc.dir, tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestAccTerraformCLIDestroy(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = terraformCLI.Apply(context.Background(), nil, "", "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	err = terraformCLI.Destroy(context.Background(), "", "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform destroy: %s", err)
	}

	got, err := terraformCLI.StateList(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	if len(got) != 0 {
		t.Errorf("expected no resources, but got: %v", got)
	}
}
