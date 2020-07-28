package tfexec

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
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

func TestAccTerraformCLIStatePush(t *testing.T) {
	SkipUnlessAcceptanceTestEnabled(t)

	source := `resource "null_resource" "foo" {}`
	e := SetupTestAcc(t, source)
	terraformCLI := NewTerraformCLI(e)

	err := terraformCLI.Init(context.Background(), "", "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to run terraform init: %s", err)
	}

	err = terraformCLI.Apply(context.Background(), nil, "", "-input=false", "-no-color", "-auto-approve")
	if err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	// Normally, state push to remote, but we push to `local` here for testing.
	// So we remove the original local state before push.
	err = os.Remove(filepath.Join(e.Dir(), "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to remove local tfstate before push: %s", err)
	}

	err = terraformCLI.StatePush(context.Background(), state)
	if err != nil {
		t.Fatalf("failed to run terraform state push: %s", err)
	}

	got, err := terraformCLI.StateList(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{"null_resource.foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
