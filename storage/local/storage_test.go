package local

import (
	"context"
	"os"
	"path/filepath"
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
				Path: "history.json",
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "dir does not exist",
			config: &Config{
				Path: "not_exist/history.json",
			},
			contents: []byte("foo"),
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			localDir, err := os.MkdirTemp("", "localDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(localDir) })

			tc.config.Path = filepath.Join(localDir, tc.config.Path)
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
				got, err := os.ReadFile(tc.config.Path)
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
				Path: "history.json",
			},
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc: "file does not exist",
			config: &Config{
				Path: "not_exist.json",
			},
			contents: []byte{},
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			localDir, err := os.MkdirTemp("", "localDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(localDir) })

			err = os.WriteFile(filepath.Join(localDir, "history.json"), tc.contents, 0600)
			if err != nil {
				t.Fatalf("failed to write contents: %s", err)
			}

			tc.config.Path = filepath.Join(localDir, tc.config.Path)
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
