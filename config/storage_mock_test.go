package config

import (
	"reflect"
	"testing"

	"github.com/minamijoyo/tfmigrate/storage"
	"github.com/minamijoyo/tfmigrate/storage/mock"
)

func TestParseMockStorageBlock(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   storage.Config
		ok     bool
	}{
		{
			desc: "valid",    source: `
tfmigrate {
  history {
    storage "mock" {
       data        = "foo"
       write_error = true
       read_error  = false
       lock_exists = false
    }
  }
}
`,
            want: &mock.Config{
                Data:       "foo",
                WriteError: true,
                ReadError:  false,
                LockExists: false,
            },
			ok: true,
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
