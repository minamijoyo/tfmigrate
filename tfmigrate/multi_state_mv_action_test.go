package tfmigrate

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestAccMultiStateMvAction(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)
	ctx := context.Background()

	// setup the initial files and states
	fromBackend := tfexec.GetTestAccBackendS3Config(t.Name() + "/fromDir")
	fromSource := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
resource "null_resource" "baz" {}
`
	fromWorkspace := "default"
	fromTf := tfexec.SetupTestAccWithApply(t, fromWorkspace, fromBackend+fromSource)

	toBackend := tfexec.GetTestAccBackendS3Config(t.Name() + "/toDir")
	toSource := `
resource "null_resource" "qux" {}
`
	toWorkspace := "default"
	toTf := tfexec.SetupTestAccWithApply(t, toWorkspace, toBackend+toSource)

	// update terraform resource files for migration
	fromUpdatedSource := `
resource "null_resource" "baz" {}
`
	tfexec.UpdateTestAccSource(t, fromTf, fromBackend+fromUpdatedSource)

	toUpdatedSource := `
resource "null_resource" "foo" {}
resource "null_resource" "bar2" {}
resource "null_resource" "qux" {}
`
	tfexec.UpdateTestAccSource(t, toTf, toBackend+toUpdatedSource)

	fromChanged, err := fromTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
	}
	if !fromChanged {
		t.Fatalf("expect to have changes in fromDir")
	}

	toChanged, err := toTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
	}
	if !toChanged {
		t.Fatalf("expect to have changes in toDir")
	}

	// perform state migration
	actions := []MultiStateAction{
		NewMultiStateMvAction("null_resource.foo", "null_resource.foo"),
		NewMultiStateMvAction("null_resource.bar", "null_resource.bar2"),
	}
	o := &MigratorOption{}
	force := false
	m := NewMultiStateMigrator(fromTf.Dir(), toTf.Dir(), fromWorkspace, toWorkspace, actions, o, force, false, false, false, false)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}
}
