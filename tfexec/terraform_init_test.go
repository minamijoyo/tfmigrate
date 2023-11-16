package tfexec

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTerraformCLIInit(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		opts         []string
		ok           bool
	}{
		{
			desc: "no opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init"},
					exitCode: 0,
				},
			},
			ok: true,
		},
		{
			desc: "failed to run terraform init",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with opts",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "init", "-input=false", "-no-color"},
					exitCode: 0,
				},
			},
			opts: []string{"-input=false", "-no-color"},
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			terraformCLI.SetExecPath("terraform")
			err := terraformCLI.Init(context.Background(), tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestAccTerraformCLIInit(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	if _, err := os.Stat(filepath.Join(e.Dir(), ".terraform")); os.IsNotExist(err) {
		t.Fatalf("failed to find .terraform directory: %s", err)
	}
}
