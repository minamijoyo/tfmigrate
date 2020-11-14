package history

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseHistoryFile(t *testing.T) {
	cases := []struct {
		desc string
		b    []byte
		want *History
		ok   bool
	}{
		{
			desc: "v1",
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
			desc: "unknown version",
			b: []byte(`{
    "version": 99,
    "foo": "bar"
}`),
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := ParseHistoryFile(tc.b)
			if tc.ok && err != nil {
				t.Fatalf("unexpected err: %s", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, got: %#v", got)
			}
			if tc.ok {
				if diff := cmp.Diff(*got, *tc.want, cmp.AllowUnexported(*got)); diff != "" {
					t.Errorf("got = %#v, want = %#v, diff = %s", got, tc.want, diff)
				}
			}
		})
	}
}
