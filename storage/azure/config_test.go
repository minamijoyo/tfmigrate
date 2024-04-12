package azure

import "testing"

func TestConfigNewStorage(t *testing.T) {
	cases := []struct {
		desc   string
		config *Config
		ok     bool
	}{
		{
			desc: "valid",
			config: &Config{
				AccountName:   "tfmigrate-test",
				ContainerName: "tfmigrate",
				BlobName:      "history.json",
			},
			ok: true,
		},
		{
			desc: "valid",
			config: &Config{
				AccessKey:     "ZHVtbXkK", // expected to be a base64-encoded string
				AccountName:   "tfmigrate-test",
				ContainerName: "tfmigrate",
				BlobName:      "history.json",
			},
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.config.NewStorage()
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				_ = got.(*Storage)
			}
		})
	}
}
