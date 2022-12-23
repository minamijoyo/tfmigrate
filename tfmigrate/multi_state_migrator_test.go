package tfmigrate

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func TestMultiStateMigratorConfigNewMigrator(t *testing.T) {
	cases := []struct {
		desc   string
		config *MultiStateMigratorConfig
		o      *MigratorOption
		ok     bool
	}{
		{
			desc: "valid and default workspace",
			config: &MultiStateMigratorConfig{
				FromDir: "dir1",
				ToDir:   "dir2",
				Actions: []string{
					"mv null_resource.foo null_resource.foo2",
					"mv null_resource.bar null_resource.bar2",
				},
			},
			o: &MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
			ok: true,
		},
		{
			desc: "valid and custom workspace",
			config: &MultiStateMigratorConfig{
				FromDir:       "dir1",
				ToDir:         "dir2",
				FromWorkspace: "work1",
				ToWorkspace:   "work2",
				Actions: []string{
					"mv null_resource.foo null_resource.foo2",
					"mv null_resource.bar null_resource.bar2",
				},
			},
			o: &MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
			ok: true,
		},
		{
			desc: "invalid action",
			config: &MultiStateMigratorConfig{
				FromDir:       "dir1",
				ToDir:         "dir2",
				FromWorkspace: "work1",
				ToWorkspace:   "work2",
				Actions: []string{
					"mv null_resource.foo",
				},
			},
			o:  nil,
			ok: false,
		},
		{
			desc: "no actions",
			config: &MultiStateMigratorConfig{
				FromDir:       "dir1",
				ToDir:         "dir2",
				FromWorkspace: "work1",
				ToWorkspace:   "work2",
				Actions:       []string{},
			},
			o:  nil,
			ok: false,
		},
		{
			desc: "force true",
			config: &MultiStateMigratorConfig{
				FromDir:       "dir1",
				ToDir:         "dir2",
				FromWorkspace: "work1",
				ToWorkspace:   "work2",
				Actions: []string{
					"mv null_resource.foo null_resource.foo2",
					"mv null_resource.bar null_resource.bar2",
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
				_ = got.(*MultiStateMigrator)
			}
		})
	}
}

func TestAccMultiStateMigratorApplySimple(t *testing.T) {
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
	m := NewMultiStateMigrator(fromTf.Dir(), toTf.Dir(), fromWorkspace, toWorkspace, actions, o, force)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}

	// verify state migration results
	fromGot, err := fromTf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list in fromDir: %s", err)
	}
	fromWant := []string{
		"null_resource.baz",
	}
	sort.Strings(fromGot)
	sort.Strings(fromWant)
	if !reflect.DeepEqual(fromGot, fromWant) {
		t.Errorf("got state: %v, want state: %v in fromDir", fromGot, fromWant)
	}

	toGot, err := toTf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list in toDir: %s", err)
	}
	toWant := []string{
		"null_resource.foo",
		"null_resource.bar2",
		"null_resource.qux",
	}
	sort.Strings(toGot)
	sort.Strings(toWant)
	if !reflect.DeepEqual(toGot, toWant) {
		t.Errorf("got state: %v, want state: %v in toDir", toGot, toWant)
	}

	fromChanged, err = fromTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
	}
	if fromChanged {
		t.Error("expect not to have changes in fromDir")
	}

	toChanged, err = toTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
	}
	if toChanged {
		t.Error("expect not to have changes in toDir")
	}
}

func TestAccMultiStateMigratorApplyWithWorkspace(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)
	ctx := context.Background()

	// setup the initial files and states
	fromBackend := tfexec.GetTestAccBackendS3Config(t.Name() + "/fromDir")
	fromSource := `
resource "null_resource" "foo" {}
resource "null_resource" "bar" {}
resource "null_resource" "baz" {}
`
	fromWorkspace := "work1"
	fromTf := tfexec.SetupTestAccWithApply(t, fromWorkspace, fromBackend+fromSource)

	toBackend := tfexec.GetTestAccBackendS3Config(t.Name() + "/toDir")
	toSource := `
resource "null_resource" "qux" {}
`
	toWorkspace := "work2"
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
	m := NewMultiStateMigrator(fromTf.Dir(), toTf.Dir(), fromWorkspace, toWorkspace, actions, o, force)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}

	// verify state migration results
	fromGot, err := fromTf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list in fromDir: %s", err)
	}
	fromWant := []string{
		"null_resource.baz",
	}
	sort.Strings(fromGot)
	sort.Strings(fromWant)
	if !reflect.DeepEqual(fromGot, fromWant) {
		t.Errorf("got state: %v, want state: %v in fromDir", fromGot, fromWant)
	}

	toGot, err := toTf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list in toDir: %s", err)
	}
	toWant := []string{
		"null_resource.foo",
		"null_resource.bar2",
		"null_resource.qux",
	}
	sort.Strings(toGot)
	sort.Strings(toWant)
	if !reflect.DeepEqual(toGot, toWant) {
		t.Errorf("got state: %v, want state: %v in toDir", toGot, toWant)
	}

	fromChanged, err = fromTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
	}
	if fromChanged {
		t.Error("expect not to have changes in fromDir")
	}

	toChanged, err = toTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
	}
	if toChanged {
		t.Error("expect not to have changes in toDir")
	}
}

func TestAccMultiStateMigratorApplyWithForce(t *testing.T) {
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
	// The expect not to have changes
	fromUpdatedSource := `
resource "null_resource" "baz" {}
`
	tfexec.UpdateTestAccSource(t, fromTf, fromBackend+fromUpdatedSource)

	// The expect to have changes
	// Note that null_resource.qux2 will be added
	toUpdatedSource := `
resource "null_resource" "foo" {}
resource "null_resource" "bar2" {}
resource "null_resource" "qux" {}
resource "null_resource" "qux2" {}
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
	o.PlanOut = "foo.tfplan"
	force := true
	m := NewMultiStateMigrator(fromTf.Dir(), toTf.Dir(), fromWorkspace, toWorkspace, actions, o, force)
	err = m.Plan(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator plan: %s", err)
	}

	err = m.Apply(ctx)
	if err != nil {
		t.Fatalf("failed to run migrator apply: %s", err)
	}

	// verify state migration results
	fromGot, err := fromTf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list in fromDir: %s", err)
	}
	fromWant := []string{
		"null_resource.baz",
	}
	sort.Strings(fromGot)
	sort.Strings(fromWant)
	if !reflect.DeepEqual(fromGot, fromWant) {
		t.Errorf("got state: %v, want state: %v in fromDir", fromGot, fromWant)
	}

	toGot, err := toTf.StateList(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to run terraform state list in toDir: %s", err)
	}
	toWant := []string{
		"null_resource.foo",
		"null_resource.bar2",
		"null_resource.qux",
	}
	sort.Strings(toGot)
	sort.Strings(toWant)
	if !reflect.DeepEqual(toGot, toWant) {
		t.Errorf("got state: %v, want state: %v in toDir", toGot, toWant)
	}

	fromChanged, err = fromTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
	}
	if fromChanged {
		t.Error("expect not to have changes in fromDir")
	}

	toChanged, err = toTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
	}
	if !toChanged {
		t.Fatalf("expect to have changes in toDir")
	}

	// Note that the saved plan file is not applicable in Terraform 1.1+.
	// https://github.com/minamijoyo/tfmigrate/pull/63
	// It's intended to use only for static analysis.
	// https://github.com/minamijoyo/tfmigrate/issues/106
	fromTfVersionMatched, err := tfexec.MatchTerraformVersion(ctx, fromTf, ">= 1.1.0")
	if err != nil {
		t.Fatalf("failed to check terraform version constraints in fromDir: %s", err)
	}
	if fromTfVersionMatched {
		t.Skip("skip the following test because the saved plan can't apply in Terraform v1.1+")
	}

	toTfVersionMatched, err := tfexec.MatchTerraformVersion(ctx, toTf, ">= 1.1.0")
	if err != nil {
		t.Fatalf("failed to check terraform version constraints in toDir: %s", err)
	}
	if toTfVersionMatched {
		t.Skip("skip the following test because the saved plan can't apply in Terraform v1.1+")
	}

	// apply the saved plan files
	fromPlan, err := os.ReadFile(filepath.Join(fromTf.Dir(), o.PlanOut))
	if err != nil {
		t.Fatalf("failed to read a saved plan file in fromDir: %s", err)
	}
	err = fromTf.Apply(ctx, tfexec.NewPlan(fromPlan), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to apply the saved plan file in fromDir: %s", err)
	}

	toPlan, err := os.ReadFile(filepath.Join(toTf.Dir(), o.PlanOut))
	if err != nil {
		t.Fatalf("failed to read a saved plan file in toDir: %s", err)
	}
	err = toTf.Apply(ctx, tfexec.NewPlan(toPlan), "-input=false", "-no-color")
	if err != nil {
		t.Fatalf("failed to apply the saved plan file in toDir: %s", err)
	}

	// Terraform >= v0.12.25 and < v0.13 has a bug for state push -force
	// https://github.com/hashicorp/terraform/issues/25761
	fromTfVersionMatched, err = tfexec.MatchTerraformVersion(ctx, fromTf, ">= 0.12.25, < 0.13")
	if err != nil {
		t.Fatalf("failed to check terraform version constraints in fromDir: %s", err)
	}
	if fromTfVersionMatched {
		t.Skip("skip the following test due to a bug in Terraform v0.12")
	}

	toTfVersionMatched, err = tfexec.MatchTerraformVersion(ctx, toTf, ">= 0.12.25, < 0.13")
	if err != nil {
		t.Fatalf("failed to check terraform version constraints in toDir: %s", err)
	}
	if toTfVersionMatched {
		t.Skip("skip the following test due to a bug in Terraform v0.12")
	}

	// Note that applying the plan file only affects a local state,
	// make sure to force push it to remote after terraform apply.
	// The -force flag is required here because the lineage of the state was changed.
	fromState, err := os.ReadFile(filepath.Join(fromTf.Dir(), "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to read a local state file in fromDir: %s", err)
	}
	err = fromTf.StatePush(ctx, tfexec.NewState(fromState), "-force")
	if err != nil {
		t.Fatalf("failed to force push the local state in fromDir: %s", err)
	}

	toState, err := os.ReadFile(filepath.Join(toTf.Dir(), "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to read a local state file in toDir: %s", err)
	}
	err = toTf.StatePush(ctx, tfexec.NewState(toState), "-force")
	if err != nil {
		t.Fatalf("failed to force push the local state in toDir: %s", err)
	}

	// confirm no changes
	fromChanged, err = fromTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
	}
	if fromChanged {
		t.Error("expect not to have changes in fromDir")
	}
	toChanged, err = toTf.PlanHasChange(ctx, nil)
	if err != nil {
		t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
	}
	if toChanged {
		t.Error("expect not to have changes in toDir")
	}
}
