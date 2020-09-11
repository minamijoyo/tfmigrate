package history

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorageWrite(t *testing.T) {
	cases := []struct {
		desc     string
		path     string
		contents []byte
		ok       bool
	}{
		{
			desc:     "simple",
			path:     "history.json",
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc:     "dir does not exist",
			path:     "not_exist/history.json",
			contents: []byte("foo"),
			ok:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			localDir, err := ioutil.TempDir("", "localDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(localDir) })

			path := filepath.Join(localDir, tc.path)
			s := NewLocalStorage(path)
			err = s.Write(context.Background(), tc.contents)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				got, err := ioutil.ReadFile(path)
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

func TestLocalStorageRead(t *testing.T) {
	cases := []struct {
		desc     string
		path     string
		contents []byte
		ok       bool
	}{
		{
			desc:     "simple",
			path:     "history.json",
			contents: []byte("foo"),
			ok:       true,
		},
		{
			desc:     "file does not exist",
			path:     "not_exist.json",
			contents: []byte{},
			ok:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			localDir, err := ioutil.TempDir("", "localDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(localDir) })

			err = ioutil.WriteFile(filepath.Join(localDir, "history.json"), tc.contents, 0644)
			if err != nil {
				t.Fatalf("failed to write contents: %s", err)
			}

			path := filepath.Join(localDir, tc.path)
			s := NewLocalStorage(path)
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
