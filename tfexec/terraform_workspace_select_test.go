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
		dir          string
		ok           bool
	}{
		{
			desc: "no workspace and no dir",
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
		{
			desc: "with workspace and dir",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "workspace", "select", "foo", "bar"},
					exitCode: 0,
				},
			},
			workspace: "foo",
			dir:       "bar",
			ok:        true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			err := terraformCLI.WorkspaceSelect(context.Background(), tc.workspace, tc.dir)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}
