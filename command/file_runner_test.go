package command

import (
	"context"
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/tfmigrate"
)

func TestLoadMigrationFile(t *testing.T) {
	cases := []struct {
		desc   string
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
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			path := setupMigrationFile(t, tc.source)

			got, err := loadMigrationFile(path)
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

func TestFileRunnerPlan(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   *tfmigrate.MigrationConfig
		ok     bool
	}{
		{
			desc: "no error",
			source: `
migration "mock" "test" {
	plan_error  = false
	apply_error = false
}
`,
			ok: true,
		},
		{
			desc: "plan error",
			source: `
migration "mock" "test" {
	plan_error  = true
	apply_error = false
}
`,
			ok: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			path := setupMigrationFile(t, tc.source)

			r, err := NewFileRunner(path, nil)
			if err != nil {
				t.Fatalf("failed to new file runner: %s", err)
			}

			err = r.Plan(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestFileRunnerApply(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   *tfmigrate.MigrationConfig
		ok     bool
	}{
		{
			desc: "no error",
			source: `
migration "mock" "test" {
	plan_error  = false
	apply_error = false
}
`,
			ok: true,
		},
		{
			desc: "plan error",
			source: `
migration "mock" "test" {
	plan_error  = true
	apply_error = false
}
`,
			ok: false,
		},
		{
			desc: "apply error",
			source: `
migration "mock" "test" {
	plan_error  = false
	apply_error = true
}
`,
			ok: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			path := setupMigrationFile(t, tc.source)

			r, err := NewFileRunner(path, nil)
			if err != nil {
				t.Fatalf("failed to new file runner: %s", err)
			}

			err = r.Apply(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}
