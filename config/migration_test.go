package config

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
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

func TestParseMigrationFileWithState(t *testing.T) {
	const source = `
	migration "state" "test" {
	  dir = "dir1"
		actions = [
			"mv aws_security_group.foo aws_security_group.foo2",
			"mv aws_security_group.bar aws_security_group.bar2",
			"rm aws_security_group.baz",
			"import aws_security_group.qux qux",
		]
	}
	`

	config, err := ParseMigrationFile("test.hcl", []byte(source))
	if err != nil {
		t.Fatalf("failed to ParseMigration: %s", err)
	}
	spew.Dump(config)

	o := &tfmigrate.MigratorOption{
		ExecPath: "direnv exec . terraform",
	}
	m, err := config.NewMigrator(o)
	if err != nil {
		t.Fatalf("failed to NewMigrator: %s", err)
	}
	spew.Dump(m)
}

func TestParseMigrationFileWithMultiState(t *testing.T) {
	const source = `
	migration "multi_state" "mv_dir1_dir2" {
	  from_dir = "dir1"
	  to_dir   = "dir2"
		actions = [
			"mv aws_security_group.foo aws_security_group.foo2",
			"mv aws_security_group.bar aws_security_group.bar2",
		]
	}
	`

	config, err := ParseMigrationFile("test.hcl", []byte(source))
	if err != nil {
		t.Fatalf("failed to ParseMigration: %s", err)
	}
	spew.Dump(config)

	o := &tfmigrate.MigratorOption{
		ExecPath: "direnv exec . terraform",
	}
	m, err := config.NewMigrator(o)
	if err != nil {
		t.Fatalf("failed to NewMigrator: %s", err)
	}
	spew.Dump(m)
}
