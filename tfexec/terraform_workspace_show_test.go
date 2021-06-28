package tfexec

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestTerraformCLIWorkspaceShow(t *testing.T) {
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		want         string
		ok           bool
	}{
		{
			desc: "parse output of terraform workspace show",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "workspace", "show"},
					stdout:   "default",
					exitCode: 0,
				},
			},
			want: "default",
			ok:   true,
		},
		{
			desc: "with existing workspace",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "workspace", "show"},
					exitCode: 1,
				},
			},
			want: "",
			ok:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.WorkspaceShow(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if tc.ok && got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}

func TestAccTerraformCLIWorkspaceShow(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	e := NewExecutor("", os.Environ())
	terraformCLI := NewTerraformCLI(e)
	got, err := terraformCLI.WorkspaceShow(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform workspace show: %s", err)
	}
	if got != "default" {
		t.Error("terraform workspace show should return the default workspace")
	}
	fmt.Printf("got = %s\n", got)
}
