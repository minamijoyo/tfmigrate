package command

import (
	"context"
	"fmt"
	"log"

	"github.com/minamijoyo/tfmigrate/config"
	"github.com/minamijoyo/tfmigrate/history"
	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

// HistoryRunner is a history-aware runner.
// It allows us to apply all unapplied migrations and save them to history.
type HistoryRunner struct {
	// A path to migration file. This is optional.
	// If set, run a single migration and save it to history.
	// If not set, run all unapplied migrations and save them to history.
	filename string
	// A global configuration.
	config *config.TfmigrateConfig
	// A option to share across migrations.
	option *tfmigrate.MigratorOption
	// A controller which manages history.
	hc *history.Controller
}

// NewHistoryRunner returns a new HistoryRunner instance.
func NewHistoryRunner(ctx context.Context, filename string, config *config.TfmigrateConfig, option *tfmigrate.MigratorOption) (*HistoryRunner, error) {
	hc, err := history.NewController(ctx, config.MigrationDir, config.History)
	if err != nil {
		return nil, err
	}

	r := &HistoryRunner{
		filename: filename,
		config:   config,
		option:   option,
		hc:       hc,
	}

	return r, nil
}

// Plan plans migrations with history-aware mode.
// If a filename is set, run a single migration.
// If not set, run all unapplied migrations.
func (r *HistoryRunner) Plan(ctx context.Context) error {
	if len(r.filename) != 0 {
		// file mode
		return r.planFile(ctx, r.filename)
	}

	// directory mode
	return r.planDir(ctx)
}

// planFile plans a single migration.
func (r *HistoryRunner) planFile(ctx context.Context, filename string) error {
	if r.hc.AlreadyApplied(filename) {
		return fmt.Errorf("a migration has already been applied: %s", filename)
	}

	fr, err := NewFileRunner(filename, r.config, r.option)
	if err != nil {
		log.Printf("[ERROR] [runner] failed to plan: %s\n", filename)
		return err
	}

	return fr.Plan(ctx)
}

// planDir plans all unapplied migrations.
func (r *HistoryRunner) planDir(ctx context.Context) error {
	unapplied := r.hc.UnappliedMigrations()

	if len(unapplied) == 0 {
		log.Printf("[INFO] [runner] no unapplied migrations\n")
		return nil
	}
	log.Printf("[INFO] [runner] unapplied migration files: %v\n", unapplied)

	for _, filename := range unapplied {
		err := r.planFile(ctx, filename)
		if err != nil {
			return err
		}
	}

	return nil
}

// Apply applies migrations and save them to history.
// If a filename is set, run a single migration.
// If not set, run all unapplied migrations.
func (r *HistoryRunner) Apply(ctx context.Context) (err error) {
	// save history on exit
	beforeLen := r.hc.HistoryLength()
	defer func() {
		// if the number of records in history doesn't change,
		// we don't want to update a timestamp of history file.
		afterLen := r.hc.HistoryLength()
		log.Printf("[DEBUG] [runner] length of history records: beforeLen = %d, afterLen = %d\n", beforeLen, afterLen)
		
		// Always try to unlock the state, even if no history was changed
		lerr := r.hc.UnlockState()
		if lerr == nil {
			log.Print("[INFO] [runner] state unlocked\n")
		} else {
			log.Printf("[ERROR] [runner] failed to unlock state: %v\n", lerr)
		}
		
		if beforeLen == afterLen {
			return
		}

		// Save history if records were changed
		log.Print("[INFO] [runner] save history\n")
		serr := r.hc.Save(ctx)
		if serr == nil {
			log.Print("[INFO] [runner] history saved\n")
		} else {
			log.Printf("[ERROR] [runner] failed to save history. The history may be inconsistent\n")
		}

		// Update the returned error with appropriate context
		if err == nil && (serr != nil || lerr != nil) {
			if serr != nil {
				err = fmt.Errorf("apply succeed, but failed to save history: %v", serr)
			}
			if lerr != nil {
				if err == nil {
					err = fmt.Errorf("apply succeed, but failed to unlock state: %v", lerr)
				} else {
					err = fmt.Errorf("%v, also failed to unlock state: %v", err, lerr)
				}
			}
			return
		}
		
		if serr != nil {
			err = fmt.Errorf("failed to save history: %v, failed to apply: %v", serr, err)
		}
		if lerr != nil {
			err = fmt.Errorf("%v, also failed to unlock state: %v", err, lerr)
		}
	}()

	err = r.hc.LockState()
	if err != nil {
		log.Printf("[ERROR] [runner] failed to lock state: %v\n", err)
		return fmt.Errorf("failed to lock state: %v", err)
	}

	if len(r.filename) != 0 {
		// file mode
		err = r.applyFile(ctx, r.filename)
		return err
	}

	// directory mode
	err = r.applyDir(ctx)
	return err
}

// applyFile applies a single migration.
func (r *HistoryRunner) applyFile(ctx context.Context, filename string) error {
	if r.hc.AlreadyApplied(filename) {
		return fmt.Errorf("a migration has already been applied: %s", filename)
	}

	fr, err := NewFileRunner(filename, r.config, r.option)
	if err != nil {
		return err
	}

	err = fr.Apply(ctx)
	if err != nil {
		log.Printf("[ERROR] [runner] failed to apply: %s\n", filename)
		return err
	}

	mc := fr.MigrationConfig()
	log.Printf("[INFO] [runner] add a record to history: %s\n", filename)
	r.hc.AddRecord(filename, mc.Type, mc.Name, nil)

	return nil
}

// applyDir applies all unapplied migrations.
func (r *HistoryRunner) applyDir(ctx context.Context) (err error) {
	unapplied := r.hc.UnappliedMigrations()

	if len(unapplied) == 0 {
		log.Printf("[INFO] [runner] no unapplied migrations\n")
		return nil
	}
	log.Printf("[INFO] [runner] unapplied migration files: %v\n", unapplied)

	for _, filename := range unapplied {
		err := r.applyFile(ctx, filename)
		if err != nil {
			return err
		}
	}

	return nil
}
