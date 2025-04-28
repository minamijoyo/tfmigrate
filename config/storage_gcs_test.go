package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/storage"
	"github.com/minamijoyo/tfmigrate/storage/gcs"
)

func TestParseGCSStorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		env    map[string]string
		source string
		want   storage.Config
		ok     bool
	}{
		{
			desc: "valid (required)",
			source: `
tfmigrate {
  history {
    storage "gcs" {
      bucket = "tfmigrate-test"
      name   = "tfmigrate/history.json"
    }
  }
}
`,
			want: &gcs.Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/history.json",
			},
			ok: true,
		},
		{
			desc: "env vars",
			env: map[string]string{
				"VAR_NAME": "env1",
			},
			source: `
tfmigrate {
  history {
    storage "gcs" {
      bucket = "tfmigrate-test"
      name   = "tfmigrate/${env.VAR_NAME}/history.json"
    }
  }
}
`,
			want: &gcs.Config{
				Bucket: "tfmigrate-test",
				Name:   "tfmigrate/env1/history.json",
			},
			ok: true,
		},
		{
			desc: "missing required attribute (bucket)",
			source: `
tfmigrate {
  history {
    storage "gcs" {
      name = "tfmigrate/history.json"
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "missing required attribute (name)",
			source: `
tfmigrate {
  history {
    storage "gcs" {
      bucket = "tfmigrate-test"
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			config, err := ParseConfigurationFile("test.hcl", []byte(tc.source))
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", config)
			}
			if tc.ok {
				got := config.History.Storage
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got: %#v, want: %#v", got, tc.want)
				}
			}
		})
	}
}
