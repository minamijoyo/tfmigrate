package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateRmAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "aws_security_group" "foo" {}
resource "aws_security_group" "bar" {}
resource "aws_security_group" "baz" {}
resource "aws_security_group" "qux" {}
`
	tf := tfexec.SetupTestAccWithApply(t, backend+source)
	ctx := context.Background()

	updatedSource := `
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
		NewStateRmAction([]string{"aws_security_group.foo", "aws_security_group.bar"}),
		NewStateRmAction([]string{"aws_security_group.qux"}),
	}

	m := NewStateMigrator(tf.Dir(), actions, nil)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}
}
