package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
	"github.com/minamijoyo/tfmigrate/storage/local"
)

func TestParseHistoryBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   *history.Config
		ok     bool
	}{
		{
			desc: "valid",
			source: `
tfmigrate {
  migration_dir = "tfmigrate"
  history {
    storage "local" {
      path = "tmp/history.json"
    }
  }
}
`,
			want: &history.Config{
				Storage: &local.Config{
					Path: "tmp/history.json",
				},
			},
			ok: true,
		},
		{
			desc: "missing block (storage)",
			source: `
tfmigrate {
  migration_dir = "tfmigrate"
  history {
  }
}
`,
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			config, err := ParseConfigurationFile("test.hcl", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", config)
			}
			if tc.ok {
				got := config.History
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}
