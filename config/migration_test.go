package config

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

func TestParseMigrationFileWithState(t *testing.T) {
	const source = `
	migration "state" "test" {
	  dir = "dir1"
		actions = [
			"mv aws_security_group.foo aws_security_group.foo2",
			"mv aws_security_group.bar aws_security_group.bar2",
			"rm aws_security_group.bar",
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
