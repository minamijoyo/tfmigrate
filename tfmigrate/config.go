package tfmigrate

// MigrationConfig is a config for a migration.
type MigrationConfig struct {
	// Type is a type for migration.
	// Valid values are `state` and `multi_state`.
	Type string
	// Name is an arbitrary name for migration.
	Name string
	// Migrator is an interface of factory method for Migrator.
	Migrator MigratorConfig
}

// MigratorConfig is an interface of factory method for Migrator.
type MigratorConfig interface {
	// NewMigrator returns a new instance of Migrator.
	NewMigrator(o *MigratorOption) (Migrator, error)
}

// MigratorOption customizes a behaviror of Migrator.
// It is used for shared settings across Migrator instances.
type MigratorOption struct {
	// ExecPath is a string how terraform command is executed. Default to terraform.
	// It's intended to inject a wrapper command such as direnv.
	// e.g.) direnv exec . terraform
	ExecPath string
}
