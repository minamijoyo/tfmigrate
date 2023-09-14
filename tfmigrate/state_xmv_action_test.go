package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccStateXmvAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)

	backend := tfexec.GetTestAccBackendS3Config(t.Name())

	source := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
`

	workspace := "default"
	tf := tfexec.SetupTestAccWithApply(t, workspace, backend+source)
	ctx := context.Background()

	updatedSource := `
resource "null_resource" "foo2" {}
resource "null_resource" "bar2" {}
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
		NewStateXmvAction("null_resource.*", "null_resource.${1}2"),
	}

	m := NewStateMigrator(tf.Dir(), workspace, actions, &MigratorOption{}, false, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}
}
