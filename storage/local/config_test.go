package local

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
				Path: "tmp/history.json",
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
