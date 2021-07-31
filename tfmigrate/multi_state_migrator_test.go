package tfmigrate

import (
	"context"
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
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
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
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
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
					"mv aws_security_group.foo",
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
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
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

func TestAccMultiStateMigratorApply(t *testing.T) {
	tfexec.SkipUnlessAcceptanceTestEnabled(t)
	cases := []struct {
		desc                  string
		fromWorkspace         string
		fromSource            string
		fromUpdatedSource     string
		fromUpdatedState      []string
		fromStateExpectChange bool
		toWorkspace           string
		toSource              string
		toUpdatedSource       string
		toUpdatedState        []string
		toStateExpectChange   bool
		actions               []string
		force                 bool
	}{
		{
			desc:          "multi-state migration between default workspaces",
			fromWorkspace: "default",
			fromSource: `
			resource "aws_security_group" "foo" {}
			resource "aws_security_group" "bar" {}
			resource "aws_security_group" "baz" {}
			`,
			fromUpdatedSource: `
			resource "aws_security_group" "baz" {}
			`,
			fromUpdatedState: []string{
				"aws_security_group.baz",
			},
			fromStateExpectChange: false,
			toWorkspace:           "default",
			toSource: `
			resource "aws_security_group" "qux" {}
			`,
			toUpdatedSource: `
			resource "aws_security_group" "foo" {}
			resource "aws_security_group" "bar2" {}
			resource "aws_security_group" "qux" {}
			`,
			toUpdatedState: []string{
				"aws_security_group.foo",
				"aws_security_group.bar2",
				"aws_security_group.qux",
			},
			toStateExpectChange: false,
			actions: []string{
				"mv aws_security_group.foo aws_security_group.foo",
				"mv aws_security_group.bar aws_security_group.bar2",
			},
			force: false,
		},
		{
			desc:          "multi-state migration between default workspaces with force == true",
			fromWorkspace: "default",
			fromSource: `
			resource "aws_security_group" "foo" {}
			resource "aws_security_group" "bar" {}
			resource "aws_security_group" "baz" {}
			`,
			fromUpdatedSource: `
			resource "aws_security_group" "baz" {}
			`,
			fromUpdatedState: []string{
				"aws_security_group.baz",
			},
			fromStateExpectChange: false,
			toWorkspace:           "default",
			toSource: `
			resource "aws_security_group" "qux" {}
			`,
			toUpdatedSource: `
			resource "aws_security_group" "foo" {}
			resource "aws_security_group" "bar2" {}
			resource "aws_security_group" "qux" {}
			resource "aws_security_group" "qux2" {}
			`,
			toUpdatedState: []string{
				"aws_security_group.foo",
				"aws_security_group.bar2",
				"aws_security_group.qux",
			},
			toStateExpectChange: true,
			actions: []string{
				"mv aws_security_group.foo aws_security_group.foo",
				"mv aws_security_group.bar aws_security_group.bar2",
			},
			force: true,
		},
		{
			desc:          "multi-state migration between user-defined workspaces",
			fromWorkspace: "work1",
			fromSource: `
			resource "aws_security_group" "foo" {}
			resource "aws_security_group" "bar" {}
			resource "aws_security_group" "baz" {}
			`,
			fromUpdatedSource: `
			resource "aws_security_group" "baz" {}
			`,
			fromUpdatedState: []string{
				"aws_security_group.baz",
			},
			fromStateExpectChange: false,
			toWorkspace:           "work2",
			toSource: `
			resource "aws_security_group" "qux" {}
			`,
			toUpdatedSource: `
			resource "aws_security_group" "foo" {}
			resource "aws_security_group" "bar2" {}
			resource "aws_security_group" "qux" {}
			`,
			toUpdatedState: []string{
				"aws_security_group.foo",
				"aws_security_group.bar2",
				"aws_security_group.qux",
			},
			toStateExpectChange: false,
			actions: []string{
				"mv aws_security_group.foo aws_security_group.foo",
				"mv aws_security_group.bar aws_security_group.bar2",
			},
			force: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			//setup the initial files and states
			fromBackend := tfexec.GetTestAccBackendS3Config(t.Name() + "/fromDir")
			fromTf := tfexec.SetupTestAccWithApply(t, tc.fromWorkspace, fromBackend+tc.fromSource)
			toBackend := tfexec.GetTestAccBackendS3Config(t.Name() + "/toDir")
			toTf := tfexec.SetupTestAccWithApply(t, tc.toWorkspace, toBackend+tc.toSource)
			//update terraform resource files for migration
			tfexec.UpdateTestAccSource(t, fromTf, fromBackend+tc.fromUpdatedSource)
			tfexec.UpdateTestAccSource(t, toTf, toBackend+tc.toUpdatedSource)
			changed, err := fromTf.PlanHasChange(ctx, nil, "")
			if err != nil {
				t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
			}
			if !changed {
				t.Fatalf("expect to have changes in fromDir")
			}
			changed, err = toTf.PlanHasChange(ctx, nil, "")
			if err != nil {
				t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
			}
			if !changed {
				t.Fatalf("expect to have changes in toDir")
			}
			//perform state migration
			actions := []MultiStateAction{}
			for _, cmdStr := range tc.actions {
				action, err := NewMultiStateActionFromString(cmdStr)
				if err != nil {
					t.Fatalf("unable to parse migration action")
				}
				actions = append(actions, action)
			}
			m := NewMultiStateMigrator(fromTf.Dir(), toTf.Dir(), tc.fromWorkspace, tc.toWorkspace, actions, nil, tc.force)
			err = m.Plan(ctx)
			if err != nil {
				t.Fatalf("failed to run migrator plan: %s", err)
			}

			err = m.Apply(ctx)
			if err != nil {
				t.Fatalf("failed to run migrator apply: %s", err)
			}
			//verify state migration results
			fromGot, err := fromTf.StateList(ctx, nil, nil)
			if err != nil {
				t.Fatalf("failed to run terraform state list in fromDir: %s", err)
			}
			sort.Strings(fromGot)
			sort.Strings(tc.fromUpdatedState)
			if !reflect.DeepEqual(fromGot, tc.fromUpdatedState) {
				t.Errorf("got state: %v, want state: %v in fromDir", fromGot, tc.fromUpdatedState)
			}
			toGot, err := toTf.StateList(ctx, nil, nil)
			if err != nil {
				t.Fatalf("failed to run terraform state list in toDir: %s", err)
			}
			sort.Strings(toGot)
			sort.Strings(tc.toUpdatedState)
			if !reflect.DeepEqual(toGot, tc.toUpdatedState) {
				t.Errorf("got state: %v, want state: %v in toDir", toGot, tc.toUpdatedState)
			}
			changed, err = fromTf.PlanHasChange(ctx, nil, "")
			if err != nil {
				t.Fatalf("failed to run PlanHasChange in fromDir: %s", err)
			}
			if changed != tc.fromStateExpectChange {
				t.Fatalf("expected change in fromDir is %t but actual value is %t", tc.fromStateExpectChange, changed)
			}
			changed, err = toTf.PlanHasChange(ctx, nil, "")
			if err != nil {
				t.Fatalf("failed to run PlanHasChange in toDir: %s", err)
			}
			if changed != tc.toStateExpectChange {
				t.Fatalf("expected change in toDir is %t but actual value is %t", tc.toStateExpectChange, changed)
			}
		})
	}
}
