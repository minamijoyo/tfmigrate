package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateMvAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name(), false)

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
`
	tf := tfexec.SetupTestAccWithApply(t, "default", backend+source, nil)
	ctx := context.Background()

	updatedSource := `
resource "aws_security_group" "foo2" {}
resource "aws_security_group" "bar2" {}
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
		NewStateMvAction("aws_security_group.bar", "aws_security_group.bar2"),
	}

	m := NewStateMigrator(tf.Dir(), "default", actions, &MigratorOption{}, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}
}
