package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/history"
)

func TestParseGCSStorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   history.StorageConfig
		ok     bool
	}{
		{
			desc: "valid (required)",
			source: `
tfmigrate {
  history {
    storage "gcs" {
      bucket = "tfmigrate-test"
      prefix = "tfmigrate/history.json"
    }
  }
}
`,
			want: &history.GCSStorageConfig{
				Bucket: "tfmigrate-test",
				Prefix: "tfmigrate/history.json",
			},
			ok: true,
		},
		{
			desc: "valid (with optional)",
			source: `
tfmigrate {
  history {
    storage "gcs" {
      bucket = "tfmigrate-test"
      prefix = "tfmigrate/history.json"

      credentials                           = "~/somePath"
      access_token                          = "dummy"
      impersonate_service_account           = "someAccount"
      impersonate_service_account_delegates = ["delegate1", "delegate2"]
      encryption_key                        = "dummy"
    }
  }
}
`,
			want: &history.GCSStorageConfig{
				Bucket:                             "tfmigrate-test",
				Prefix:                             "tfmigrate/history.json",
				Credentials:                        "~/somePath",
				AccessToken:                        "dummy",
				ImpersonateServiceAccount:          "someAccount",
				ImpersonateServiceAccountDelegates: []string{"delegate1", "delegate2"},
				EncryptionKey:                      "dummy",
			},
			ok: true,
		},
		{
			desc: "missing required attribute (bucket)",
			source: `
tfmigrate {
  history {
    storage "gcs" {
      prefix = "tfmigrate/history.json"
    }
  }
}
`,
			want: nil,
			ok:   false,
		},
		{
			desc: "missing required attribute (prefix)",
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
