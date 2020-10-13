package history

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewFileV1(t *testing.T) {
	cases := []struct {
		desc string
		h    History
		want *FileV1
	}{
		{
			desc: "simple",
			h: History{
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
			want: &FileV1{
				Version: 1,
				Records: map[string]RecordV1{
					"20201012010101_foo.hcl": RecordV1{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": RecordV1{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := newFileV1(tc.h)
			opt := cmp.AllowUnexported(*got)
			if diff := cmp.Diff(*got, *tc.want, opt); diff != "" {
				t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
			}
		})
	}
}

func TestFileV1Serialize(t *testing.T) {
	cases := []struct {
		desc string
		f    FileV1
		want string
	}{
		{
			desc: "simple",
			f: FileV1{
				Version: 1,
				Records: map[string]RecordV1{
					"20201012010101_foo.hcl": RecordV1{
						Type:      "state",
						Name:      "foo",
						AppliedAt: time.Date(2020, 10, 13, 1, 2, 3, 0, time.UTC),
					},
					"20201012020202_foo.hcl": RecordV1{
						Type:      "state",
						Name:      "bar",
						AppliedAt: time.Date(2020, 10, 13, 4, 5, 6, 0, time.UTC),
					},
				},
			},
			want: `{
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
}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.f.Serialize()
			if err != nil {
				t.Fatalf("failed to serialize: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("got = %s, want = %s", string(got), tc.want)
			}
		})
	}
}

func TestParseHistoryFileV1(t *testing.T) {
	cases := []struct {
		desc string
		b    []byte
		want *History
		ok   bool
	}{
		{
			desc: "valid",
			b: []byte(`{
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
			desc: "invalid (empty)",
			b:    []byte(``),
			want: nil,
			ok:   false,
		},
		{
			desc: "invalid (broken)",
			b:    []byte(`{`),
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := parseHistoryFileV1(tc.b)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				opt := cmp.AllowUnexported(*got)
				if diff := cmp.Diff(*got, *tc.want, opt); diff != "" {
					t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
				}
			}
		})
	}
}
