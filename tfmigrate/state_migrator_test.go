package tfmigrate

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestStateMigratorConfigNewMigrator(t *testing.T) {
	cases := []struct {
		desc   string
		config *StateMigratorConfig
		o      *MigratorOption
		ok     bool
	}{
		{
			desc: "valid (with dir)",
			config: &StateMigratorConfig{
				Dir: "dir1",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
					"rm aws_security_group.baz",
					"import aws_security_group.qux qux",
				},
			},
			o: &MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
			ok: true,
		},
		{
			desc: "valid (without dir)",
			config: &StateMigratorConfig{
				Dir: "",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
					"rm aws_security_group.baz",
					"import aws_security_group.qux qux",
				},
			},
			o: &MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
			ok: true,
		},
		{
			desc: "valid in non-default workspace",
			config: &StateMigratorConfig{
				Dir: "dir1",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
					"rm aws_security_group.baz",
					"import aws_security_group.qux qux",
				},
				Workspace: "workspace1",
			},
			o: &MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
			ok: true,
		},
		{
			desc: "invalid action",
			config: &StateMigratorConfig{
				Dir: "",
				Actions: []string{
					"mv aws_security_group.foo",
				},
			},
			o:  nil,
			ok: false,
		},
		{
			desc: "no actions",
			config: &StateMigratorConfig{
				Dir:     "",
				Actions: []string{},
			},
			o:  nil,
			ok: false,
		},
		{
			desc: "with force true",
			config: &StateMigratorConfig{
				Dir: "dir1",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
					"rm aws_security_group.baz",
					"import aws_security_group.qux qux",
				},
				Force: true,
			},
			o:  nil,
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.config.NewMigrator(tc.o, false)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				_ = got.(*StateMigrator)
			}
		})
	}
}

func TestAccStateMigratorApply(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	cases := []struct {
		desc      string
		workspace string
	}{
		{
			desc:      "default workspace",
			workspace: "default",
		},
		{
			desc:      "non-default workspace",
			workspace: "workspace1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			backend := tfexec.GetTestAccBackendS3Config(t.Name())

			source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
resource "aws_iam_user" "qux" {
	name = "qux"
}
`
			tf := tfexec.SetupTestAccWithApply(t, tc.workspace, backend+source)
			ctx := context.Background()

			updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "baz" {}
resource "aws_iam_user" "qux" {
	name = "qux"
}
`

			tfexec.UpdateTestAccSource(t, tf, backend+updatedSource)

			_, err := tf.StateRm(ctx, nil, []string{"aws_iam_user.qux"})
			if err != nil {
				t.Fatalf("failed to run terraform state rm: %s", err)
			}

			changed, err := tf.PlanHasChange(ctx, nil)
			if err != nil {
				t.Fatalf("failed to run PlanHasChange: %s", err)
			}
			if !changed {
				t.Fatalf("expect to have changes")
			}

			actions := []StateAction{
				NewStateMvAction("aws_security_group.foo", "aws_security_group.foo2"),
				NewStateRmAction([]string{"aws_security_group.bar"}),
				NewStateImportAction("aws_iam_user.qux", "qux"),
			}

			m := NewStateMigrator(tf.Dir(), tc.workspace, actions, &MigratorOption{}, false, false)
			err = m.Plan(ctx)
			if err != nil {
				t.Fatalf("failed to run migrator plan: %s", err)
			}

			err = m.Apply(ctx)
			if err != nil {
				t.Fatalf("failed to run migrator apply: %s", err)
			}

			got, err := tf.StateList(ctx, nil, nil)
			if err != nil {
				t.Fatalf("failed to run terraform state list: %s", err)
			}

			want := []string{
				"aws_security_group.foo2",
				"aws_security_group.baz",
				"aws_iam_user.qux",
			}
			sort.Strings(got)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got state: %v, want state: %v", got, want)
			}

			changed, err = tf.PlanHasChange(ctx, nil)
			if err != nil {
				t.Fatalf("failed to run PlanHasChange: %s", err)
			}
			if changed {
				t.Fatalf("expect not to have changes")
			}
		})
	}
}

func TestAccStateMigratorApplyForce(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source)
	ctx := context.Background()

	updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
`

	tfexec.UpdateTestAccSource(t, tf, backend+updatedSource)

	changed, err := tf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	actions := []StateAction{
		NewStateMvAction("aws_security_group.foo", "aws_security_group.foo2"),
	}

	o := &MigratorOption{}
	o.PlanOut = "foo.tfplan"

	m := NewStateMigrator(tf.Dir(), "default", actions, o, true, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}

	got, err := tf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list: %s", err)
	}

	want := []string{
		"aws_security_group.foo2",
		"aws_security_group.bar",
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got state: %v, want state: %v", got, want)
	}

	changed, err = tf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	// The tfmigrate plan --out=tfplan option was based on a bug prior to Terraform 1.1.
	// Terraform v1.1 now rejects the plan as stale.
	// The tfmigrate plan --out=tfplan option is deprecated without replacement.
	// https://github.com/minamijoyo/tfmigrate/issues/62
	tfVersionMatched, err := tfexec.MatchTerraformVersion(ctx, tf, ">= 1.1.0")
	if err != nil {
		t.Fatalf("failed to check terraform version constraints: %s", err)
	}
	if tfVersionMatched {
		t.Skip("skip the following test because the saved plan can't apply in Terraform v1.1+")
	}

	// apply the saved plan files
	plan, err := ioutil.ReadFile(filepath.Join(tf.Dir(), o.PlanOut))
	if err != nil {
		t.Fatalf("failed to read a saved plan file: %s", err)
	}
	err = tf.Apply(ctx, tfexec.NewPlan(plan), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to apply the saved plan file: %s", err)
	}

	// Terraform >= v0.12.25 and < v0.13 has a bug for state push -force
	// https://github.com/hashicorp/terraform/issues/25761
	tfVersionMatched, err = tfexec.MatchTerraformVersion(ctx, tf, ">= 0.12.25, < 0.13")
	if err != nil {
		t.Fatalf("failed to check terraform version constraints: %s", err)
	}
	if tfVersionMatched {
		t.Skip("skip the following test due to a bug in Terraform v0.12")
	}

	// Note that applying the plan file only affects a local state,
	// make sure to force push it to remote after terraform apply.
	// The -force flag is required here because the lineage of the state was changed.
	state, err := ioutil.ReadFile(filepath.Join(tf.Dir(), "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to read a local state file: %s", err)
	}
	err = tf.StatePush(ctx, tfexec.NewState(state), "-force")
	if err != nil {
		t.Fatalf("failed to force push the local state: %s", err)
	}

	// confirm no changes
	changed, err = tf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if changed {
		t.Fatalf("expect not to have changes")
	}
}
