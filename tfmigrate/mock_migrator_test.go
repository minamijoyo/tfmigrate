package tfmigrate

import (
	"context"
	"testing"
)

func TestMockMigratorConfigNewMigrator(t *testing.T) {
	cases := []struct {
		desc   string
		config *MockMigratorConfig
		o      *MigratorOption
		ok     bool
	}{
		{
			desc: "valid",
			config: &MockMigratorConfig{
				PlanError:  true,
				ApplyError: false,
			},
			o: &MigratorOption{
				ExecPath: "direnv exec . terraform",
			},
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
				_ = got.(*MockMigrator)
			}
		})
	}
}

func TestMockMigratorPlan(t *testing.T) {
	cases := []struct {
		desc string
		m    *MockMigrator
		ok   bool
	}{
		{
			desc: "no error",
			m: &MockMigrator{
				planError:  false,
				applyError: false,
			},
			ok: true,
		},
		{
			desc: "plan error",
			m: &MockMigrator{
				planError:  true,
				applyError: false,
			},
			ok: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.m.Plan(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}

func TestMockMigratorApply(t *testing.T) {
	cases := []struct {
		desc string
		m    *MockMigrator
		ok   bool
	}{
		{
			desc: "no error",
			m: &MockMigrator{
				planError:  false,
				applyError: false,
			},
			ok: true,
		},
		{
			desc: "plan error",
			m: &MockMigrator{
				planError:  true,
				applyError: false,
			},
			ok: false,
		},
		{
			desc: "apply error",
			m: &MockMigrator{
				planError:  false,
				applyError: true,
			},
			ok: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.m.Apply(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
		})
	}
}
