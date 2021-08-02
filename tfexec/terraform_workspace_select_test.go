package tfexec

import (
	"context"
	"testing"
)

func TestTerraformCLIWorkspaceSelect(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		workspace    string
		ok           bool
	}{
		{
			desc: "no workspace",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "workspace", "select"},
					exitCode: 1,
				},
			},
			ok: false,
		},
		{
			desc: "with existing workspace",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "workspace", "select", "foo"},
					exitCode: 0,
				},
			},
			workspace: "foo",
			ok:        true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.WorkspaceSelect(context.Background(), tc.workspace)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestAccTerraformCLIWorkspaceSelect(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := ``
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "", "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = terraformCLI.WorkspaceNew(context.Background(), "myworkspace")
	if err != nil {
		t.Fatalf("failed to create a new workspace: %s", err)
	}

	err = terraformCLI.WorkspaceSelect(context.Background(), "default")
	if err != nil {
		t.Fatalf("failed to switch back to default workspace: %s", err)
	}

	got, err := terraformCLI.WorkspaceShow(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform workspace show: %s", err)
	}

	if got != "default" {
		t.Error("The current workspace doesn't match the workspace that was just selected")
	}
}
