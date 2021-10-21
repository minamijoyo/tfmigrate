package tfexec

import (
	"context"
	"reflect"
	"testing"
)

func TestTerraformCLIStatePull(t *testing.T) {
	stdout := "dummy state"
	cases := []struct {
		desc         string
		mockCommands []*mockCommand
		opts         []string
		want         *State
		ok           bool
	}{
		{
			desc: "print tfstate to stdout",
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "pull"},
					stdout:   stdout,
					exitCode: 0,
				},
			},
			want: NewState([]byte(stdout)),
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
			want: nil,
			ok:   false,
		},
		{
			desc: "with opts", // there is no valid option for now, just pass a dummy for testing.
			mockCommands: []*mockCommand{
				{
					args:     []string{"terraform", "state", "pull", "-foo"},
					stdout:   stdout,
					exitCode: 0,
				},
			},
			opts: []string{"-foo"},
			want: NewState([]byte(stdout)),
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := NewMockExecutor(tc.mockCommands)
			terraformCLI := NewTerraformCLI(e)
			got, err := terraformCLI.StatePull(context.Background(), tc.opts...)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got = %s", got)
			}
			if tc.ok && !reflect.DeepEqual(got.Bytes(), tc.want.Bytes()) {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}

func TestAccTerraformCLIStatePull(t *testing.T) {
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

	state, err := terraformCLI.StatePull(context.Background())
	if err != nil {
		t.Fatalf("failed to run terraform state pull: %s", err)
	}

	got, err := terraformCLI.StateList(context.Background(), state, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{"null_resource.foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
