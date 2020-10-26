package history

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLoadMigrationFileNames(t *testing.T) {
	cases := []struct {
		desc  string
		files []string
		want  []string
		ok    bool
	}{
		{
			desc: "hcl",
			files: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			ok: true,
		},
		{
			desc: "json",
			files: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.json",
				"20201012020202_foo.txt",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.json",
			},
			ok: true,
		},
		{
			desc: "unsorted",
			files: []string{
				"20201012020202_foo.hcl",
				"20201012010101_foo.hcl",
			},
			want: []string{
				"20201012010101_foo.hcl",
				"20201012020202_foo.hcl",
			},
			ok: true,
		},
		{
			desc:  "empty",
			files: []string{},
			want:  []string{},
			ok:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			migrationDir, err := ioutil.TempDir("", "migrationDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(migrationDir) })

			for _, filename := range tc.files {
				err = ioutil.WriteFile(filepath.Join(migrationDir, filename), []byte{}, 0644)
				if err != nil {
					t.Fatalf("failed to write dummy migration file: %s", err)
				}
			}

			got, err := loadMigrationFileNames(migrationDir)

			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %#v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				if diff := cmp.Diff(got, tc.want); diff != "" {
					t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
				}
			}
		})
	}
}

func TestLoadHistory(t *testing.T) {
	cases := []struct {
		desc     string
		path     string
		contents []byte
		want     *History
		ok       bool
	}{
		{
			desc: "simple",
			path: "history.json",
			contents: []byte(`{
    "version": 1,
    "records": {
        "20201012010101_foo.hcl": {
            "type": "state",
            "name": "foo",
            "applied_at": "2020-10-13T01:02:03Z"
        },
        "20201012020202_foo.hcl": {
            "type": "state",
            "name": "bar",
            "applied_at": "2020-10-13T04:05:06Z"
        }
    }
}`),
			want: &History{
				records: map[string]Record{
					"20201012010101_foo.hcl": Record{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": Record{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			ok: true,
		},
		{
			desc:     "file does not exist",
			path:     "not_exist.json",
			contents: []byte{},
			want:     newEmptyHistory(),
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

			config := &LocalStorageConfig{
				Path: filepath.Join(localDir, tc.path),
			}

			got, err := loadHistory(context.Background(), config)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %#v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected to return an error, but no error")
			}

			if tc.ok {
				if diff := cmp.Diff(*got, *tc.want, cmp.AllowUnexported(*got)); diff != "" {
					t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
				}
			}
		})
	}
}

func TestControllerSave(t *testing.T) {
	cases := []struct {
		desc string
		h    *History
		want []byte
		ok   bool
	}{
		{
			desc: "simple",
			h:    newEmptyHistory(),
			want: []byte(`{
    "version": 1,
    "records": {}
}`),
			ok: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			localDir, err := ioutil.TempDir("", "localDir")
			if err != nil {
				t.Fatalf("failed to craete temp dir: %s", err)
			}
			t.Cleanup(func() { os.RemoveAll(localDir) })

			path := filepath.Join(localDir, "history.json")
			c := &Controller{
				history: *tc.h,
				config: Config{
					Storage: &LocalStorageConfig{
						Path: path,
					},
				},
			}
			err = c.Save(context.Background())
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
				if string(got) != string(tc.want) {
					t.Errorf("got: %s, want: %s", string(got), string(tc.want))
				}
			}
		})
	}
}
