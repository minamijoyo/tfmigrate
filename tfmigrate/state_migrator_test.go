package tfmigrate

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateMigratorApply(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t)

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
resource "aws_iam_user" "piyo" {
	name = "piyo"
}
`
	tf := tfexec.SetupTestAccWithApply(t, backend+source)
	ctx := context.Background()

	updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "baz" {}
resource "aws_iam_user" "piyo" {
	name = "piyo"
}
`

	tfexec.UpdateTestAccSource(t, tf, backend+updatedSource)

	_, err := tf.StateRm(ctx, nil, []string{"aws_iam_user.piyo"})
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
		NewStateImportAction("aws_iam_user.piyo", "piyo"),
	}

	m := NewStateMigrator(tf.Dir(), actions, nil)
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
		"aws_iam_user.piyo",
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
