package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

func TestParseMigrationFileWithNativeSyntax(t *testing.T) {
	cases := []struct {
		desc   string
		env    map[string]string
		source string
		want   *tfmigrate.MigrationConfig
		ok     bool
	}{
		{
			desc: "mock",
			source: `
migration "mock" "test" {
	plan_error  = true
	apply_error = false
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "mock",
				Name: "test",
				Migrator: &tfmigrate.MockMigratorConfig{
					PlanError:  true,
					ApplyError: false,
				},
			},
			ok: true,
		},
		{
			desc: "state with dir",
			source: `
migration "state" "test" {
	dir = "dir1"
	actions = [
		"mv null_resource.foo null_resource.foo2",
		"mv null_resource.bar null_resource.bar2",
		"rm time_static.baz",
		"import time_static.qux 2006-01-02T15:04:05Z",
	]
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Dir: "dir1",
					Actions: []string{
						"mv null_resource.foo null_resource.foo2",
						"mv null_resource.bar null_resource.bar2",
						"rm time_static.baz",
						"import time_static.qux 2006-01-02T15:04:05Z",
					},
				},
			},
			ok: true,
		},
		{
			desc: "state without dir",
			source: `
migration "state" "test" {
	actions = [
		"mv null_resource.foo null_resource.foo2",
	]
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Dir: "",
					Actions: []string{
						"mv null_resource.foo null_resource.foo2",
					},
				},
			},
			ok: true,
		},
		{
			desc: "state with a simple wildcard action",
			source: `
migration "state" "test" {
	actions = [
		"xmv null_resource.* null_resource.new_$1",
	]
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Dir: "",
					Actions: []string{
						"xmv null_resource.* null_resource.new_$1",
					},
				},
			},
			ok: true,
		},
		{
			desc: "state with a escaped wildcard action",
			source: `
migration "state" "test" {
	actions = [
		"xmv null_resource.* null_resource.$${1}2",
	]
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Dir: "",
					Actions: []string{
						"xmv null_resource.* null_resource.${1}2",
					},
				},
			},
			ok: true,
		},
		{
			desc: "state without actions",
			source: `
migration "state" "test" {
	dir = ""
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc:   "no migration block",
			source: "",
			want:   nil,
			ok:     false,
		},
		{
			desc: "state with force",
			source: `
migration "state" "test" {
	dir = "dir1"
    force = true
	actions = [
		"mv null_resource.foo null_resource.foo2",
		"mv null_resource.bar null_resource.bar2",
		"rm time_static.baz",
		"import time_static.qux 2006-01-02T15:04:05Z",
	]
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Dir: "dir1",
					Actions: []string{
						"mv null_resource.foo null_resource.foo2",
						"mv null_resource.bar null_resource.bar2",
						"rm time_static.baz",
						"import time_static.qux 2006-01-02T15:04:05Z",
					},
					Force: true,
				},
			},
			ok: true,
		},
		{
			desc: "multi state with from_dir and to_dir",
			source: `
migration "multi_state" "mv_dir1_dir2" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv null_resource.foo null_resource.foo2",
		"mv null_resource.bar null_resource.bar2",
	]
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "multi_state",
				Name: "mv_dir1_dir2",
				Migrator: &tfmigrate.MultiStateMigratorConfig{
					FromDir: "dir1",
					ToDir:   "dir2",
					Actions: []string{
						"mv null_resource.foo null_resource.foo2",
						"mv null_resource.bar null_resource.bar2",
					},
				},
			},
			ok: true,
		},
		{
			desc: "multi state without from_dir",
			source: `
migration "multi_state" "mv_dir1_dir2" {
	to_dir   = "dir2"
	actions = [
		"mv null_resource.foo null_resource.foo2",
		"mv null_resource.bar null_resource.bar2",
	]
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "multi state without to_dir",
			source: `
migration "multi_state" "mv_dir1_dir2" {
	from_dir = "dir1"
	actions = [
		"mv null_resource.foo null_resource.foo2",
		"mv null_resource.bar null_resource.bar2",
	]
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "multi state without actions",
			source: `
migration "multi_state" "mv_dir1_dir2" {
	from_dir = "dir1"
	to_dir   = "dir2"
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "multi state with force",
			source: `
migration "multi_state" "mv_dir1_dir2" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv null_resource.foo null_resource.foo2",
		"mv null_resource.bar null_resource.bar2",
	]
    force    = true
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "multi_state",
				Name: "mv_dir1_dir2",
				Migrator: &tfmigrate.MultiStateMigratorConfig{
					FromDir: "dir1",
					ToDir:   "dir2",
					Actions: []string{
						"mv null_resource.foo null_resource.foo2",
						"mv null_resource.bar null_resource.bar2",
					},
					Force: true,
				},
			},
			ok: true,
		},
		{
			desc: "unknown migration type",
			source: `
migration "foo" "test" {
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "duplicated state migration blocks",
			source: `
migration "state" "foo" {
	actions = [
		"mv null_resource.foo null_resource.foo2",
	]
}
migration "state" "bar" {
	actions = [
		"mv null_resource.bar null_resource.bar2",
	]
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "duplicated multi_state migration blocks",
			source: `
migration "multi_state" "foo" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv null_resource.foo null_resource.foo2",
	]
}
migration "multi_state" "bar" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv null_resource.bar null_resource.bar2",
	]
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "duplicated migration blocks (state and multi_state mixed)",
			source: `
migration "state" "foo" {
	actions = [
		"mv null_resource.foo null_resource.foo2",
	]
}
migration "multi_state" "bar" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv null_resource.bar null_resource.bar2",
	]
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "unknown block type",
			source: `
foo "bar" "baz" {}
`,
			want: nil,
			ok:   false,
		},
		{
			desc:   "empty file",
			source: ``,
			want:   nil,
			ok:     false,
		},
		{
			desc: "envirionment variable",
			env:  map[string]string{"TFMIGRATE_TEST_WORKSPACE": "test-workspace"},
			source: `
migration "state" "test" {
	workspace = env.TFMIGRATE_TEST_WORKSPACE
	actions = []
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Actions:   []string{},
					Workspace: "test-workspace",
				},
			},
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			got, err := ParseMigrationFile("test.hcl", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}

func TestParseMigrationFileWithJsonSyntax(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   *tfmigrate.MigrationConfig
		ok     bool
	}{
		{
			desc: "state with dir",
			source: `
{
  "migration": {
    "state": {
      "test": {
        "dir": "dir1",
        "actions": [
          "mv null_resource.foo null_resource.foo2",
          "mv null_resource.bar null_resource.bar2",
          "rm time_static.baz",
          "import time_static.qux 2006-01-02T15:04:05Z"
        ]
      }
    }
  }
}
`,
			want: &tfmigrate.MigrationConfig{
				Type: "state",
				Name: "test",
				Migrator: &tfmigrate.StateMigratorConfig{
					Dir: "dir1",
					Actions: []string{
						"mv null_resource.foo null_resource.foo2",
						"mv null_resource.bar null_resource.bar2",
						"rm time_static.baz",
						"import time_static.qux 2006-01-02T15:04:05Z",
					},
				},
			},
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := ParseMigrationFile("test.json", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}
