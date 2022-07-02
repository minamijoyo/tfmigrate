package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateImportAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name(), false)

	source := `
resource "aws_iam_user" "foo" {
	name = "foo"
}
resource "aws_iam_user" "bar" {
	name = "bar"
}
resource "aws_iam_user" "baz" {
	name = "baz"
}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source, nil)
	ctx := context.Background()

	_, err := tf.StateRm(ctx, nil, []string{"aws_iam_user.foo", "aws_iam_user.baz"})
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
		NewStateImportAction("aws_iam_user.foo", "foo"),
		NewStateImportAction("aws_iam_user.baz", "baz"),
	}

	m := NewStateMigrator(tf.Dir(), "default", actions, &MigratorOption{}, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}
}
