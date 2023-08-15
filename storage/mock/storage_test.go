package mock

import (
	"context"
	"testing"
)

func TestStorageWrite(t *testing.T) {
	cases := []struct {
		desc     string
		config   *Config
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			config: &Config{
				Data:       "",
				WriteError: false,
				ReadError:  false,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "write error",
			config: &Config{
				Data:       "",
				WriteError: true,
				ReadError:  false,
			},
			contents: []byte("foo"),
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewStorage(tc.config)
			if err != nil {
				t.Fatalf("failed to NewStorage: %s", err)
			}
			err = s.Write(context.Background(), tc.contents)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				got := []byte(s.data)
				if err != nil {
					t.Fatalf("failed to read contents: %s", err)
				}
				if string(got) != string(tc.contents) {
					t.Errorf("got: %s, want: %s", string(got), string(tc.contents))
				}
			}
		})
	}
}

func TestStorageRead(t *testing.T) {
	cases := []struct {
		desc     string
		config   *Config
		contents []byte
		ok       bool
	}{
		{
			desc: "simple",
			config: &Config{
				Data:       "foo",
				WriteError: false,
				ReadError:  false,
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "read error",
			config: &Config{
				Data:       "foo",
				WriteError: false,
				ReadError:  true,
			},
			contents: nil,
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := NewStorage(tc.config)
			if err != nil {
				t.Fatalf("failed to NewStorage: %s", err)
			}
			got, err := s.Read(context.Background())
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %#v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				if string(got) != string(tc.contents) {
					t.Errorf("got: %s, want: %s", string(got), string(tc.contents))
				}
			}
		})
	}
}
