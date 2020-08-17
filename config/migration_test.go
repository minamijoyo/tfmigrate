package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

func TestStateMigratorConfigNewMigrator(t *testing.T) {
	cases := []struct {
		desc   string
		config *StateMigratorConfig
		o      *tfmigrate.MigratorOption
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
			o: &tfmigrate.MigratorOption{
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
			o: &tfmigrate.MigratorOption{
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
				_ = got.(*tfmigrate.StateMigrator)
			}
		})
	}
}

func TestMultiStateMigratorConfigNewMigrator(t *testing.T) {
	cases := []struct {
		desc   string
		config *MultiStateMigratorConfig
		o      *tfmigrate.MigratorOption
		ok     bool
	}{
		{
			desc: "valid",
			config: &MultiStateMigratorConfig{
				FromDir: "dir1",
				ToDir:   "dir2",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
				},
			},
			o: &tfmigrate.MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
			ok: true,
		},
		{
			desc: "invalid action",
			config: &MultiStateMigratorConfig{
				FromDir: "dir1",
				ToDir:   "dir2",
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
				FromDir: "dir1",
				ToDir:   "dir2",
				Actions: []string{},
			},
			o:  nil,
			ok: false,
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
				_ = got.(*tfmigrate.MultiStateMigrator)
			}
		})
	}
}

func TestParseMigrationFile(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   MigratorConfig
		ok     bool
	}{
		{
			desc: "state with dir",
			source: `
migration "state" "test" {
	dir = "dir1"
	actions = [
		"mv aws_security_group.foo aws_security_group.foo2",
		"mv aws_security_group.bar aws_security_group.bar2",
		"rm aws_security_group.baz",
		"import aws_security_group.qux qux",
	]
}
`,
			want: &StateMigratorConfig{
				Dir: "dir1",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
					"rm aws_security_group.baz",
					"import aws_security_group.qux qux",
				},
			},
			ok: true,
		},
		{
			desc: "state without dir",
			source: `
migration "state" "test" {
	actions = [
		"mv aws_security_group.foo aws_security_group.foo2",
	]
}
`,
			want: &StateMigratorConfig{
				Dir: "",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
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
			desc: "multi state with from_dir and to_dir",
			source: `
migration "multi_state" "mv_dir1_dir2" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv aws_security_group.foo aws_security_group.foo2",
		"mv aws_security_group.bar aws_security_group.bar2",
	]
}
`,
			want: &MultiStateMigratorConfig{
				FromDir: "dir1",
				ToDir:   "dir2",
				Actions: []string{
					"mv aws_security_group.foo aws_security_group.foo2",
					"mv aws_security_group.bar aws_security_group.bar2",
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
		"mv aws_security_group.foo aws_security_group.foo2",
		"mv aws_security_group.bar aws_security_group.bar2",
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
		"mv aws_security_group.foo aws_security_group.foo2",
		"mv aws_security_group.bar aws_security_group.bar2",
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
		"mv aws_security_group.foo aws_security_group.foo2",
	]
}
migration "state" "bar" {
	actions = [
		"mv aws_security_group.bar aws_security_group.bar2",
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
		"mv aws_security_group.foo aws_security_group.foo2",
	]
}
migration "multi_state" "bar" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv aws_security_group.bar aws_security_group.bar2",
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
		"mv aws_security_group.foo aws_security_group.foo2",
	]
}
migration "multi_state" "bar" {
	from_dir = "dir1"
	to_dir   = "dir2"
	actions = [
		"mv aws_security_group.bar aws_security_group.bar2",
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
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
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
