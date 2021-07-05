package tfmigrate

import (
	"context"
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
			got, err := tc.config.NewMigrator(tc.o)
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

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
resource "aws_iam_user" "qux" {
	name = "qux"
}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source)
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

	changed, err := tf.PlanHasChange(ctx, nil, "")
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

	m := NewStateMigrator(tf.Dir(), actions, nil, false)
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

	changed, err = tf.PlanHasChange(ctx, nil, "")
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if changed {
		t.Fatalf("expect not to have changes")
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

	changed, err := tf.PlanHasChange(ctx, nil, "")
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}

	actions := []StateAction{
		NewStateMvAction("aws_security_group.foo", "aws_security_group.foo2"),
	}

	m := NewStateMigrator(tf.Dir(), actions, nil, true)
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

	changed, err = tf.PlanHasChange(ctx, nil, "")
	if err != nil {
		t.Fatalf("failed to run PlanHasChange: %s", err)
	}
	if !changed {
		t.Fatalf("expect to have changes")
	}
}
