package command

import (
	"context"
	"testing"

	"github.com/minamijoyo/tfmigrate/config"
	"github.com/minamijoyo/tfmigrate/history"
	"github.com/minamijoyo/tfmigrate/storage/mock"
)

func TestListMigrations(t *testing.T) {
	migrations := map[string]string{
		"20201109000001_test1.hcl": `
migration "mock" "test1" {
	plan_error  = false
	apply_error = false
}
`,
		"20201109000002_test2.hcl": `
migration "mock" "test2" {
	plan_error  = false
	apply_error = false
}
`,
		"20201109000003_test3.hcl": `
migration "mock" "test3" {
	plan_error  = false
	apply_error = false
}
`,
		"20201109000004_test4.hcl": `
migration "mock" "test4" {
	plan_error  = false
	apply_error = false
}
`,
	}
	historyFile := `{
    "version": 1,
    "records": {
        "20201109000001_test1.hcl": {
            "type": "mock",
            "name": "test1",
            "applied_at": "2020-11-10T00:00:01Z"
        },
        "20201109000002_test2.hcl": {
            "type": "mock",
            "name": "test2",
            "applied_at": "2020-11-10T00:00:02Z"
        }
    }
}`

	cases := []struct {
		desc        string
		status      string
		migrations  map[string]string
		historyFile string
		want        string
		ok          bool
	}{
		{
			desc:        "all",
			status:      "all",
			migrations:  migrations,
			historyFile: historyFile,
			want: `20201109000001_test1.hcl
20201109000002_test2.hcl
20201109000003_test3.hcl
20201109000004_test4.hcl`,
			ok: true,
		},
		{
			desc:        "unapplied",
			status:      "unapplied",
			migrations:  migrations,
			historyFile: historyFile,
			want: `20201109000003_test3.hcl
20201109000004_test4.hcl`,
			ok: true,
		},
		{
			desc:        "unknown status",
			status:      "foo",
			migrations:  migrations,
			historyFile: historyFile,
			want:        "",
			ok:          false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			migrationDir := setupMigrationDir(t, tc.migrations)
			storage := &mock.Config{
				Data:       tc.historyFile,
				WriteError: false,
				ReadError:  false,
			}
			config := &config.TfmigrateConfig{
				MigrationDir: migrationDir,
				History: &history.Config{
					Storage: storage,
				},
			}
			got, err := listMigrations(context.Background(), config, tc.status)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}
			if got != tc.want {
				t.Errorf("got = %#v, want = %#v", got, tc.want)
			}
		})
	}
}
