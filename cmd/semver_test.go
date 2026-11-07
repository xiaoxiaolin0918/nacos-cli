package cmd

import "testing"

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		inst, other string
		want        VersionRelation
	}{
		{"1.2.3", "1.2.3", RelEqual},
		{"v1.2.3", "1.2.3", RelEqual},
		{"0.3.3", "v0.3.4", RelOutdated},
		{"0.3.5", "0.3.4", RelAhead},
		{"1.0", "1.0.0", RelEqual},
		{"1", "1.0.0", RelEqual},
		{"1.2.3-beta", "1.2.3", RelInvalid},
	}
	for _, tc := range cases {
		got := CompareSemver(tc.inst, tc.other)
		if got != tc.want {
			t.Fatalf("CompareSemver(%q,%q)=%q want %q", tc.inst, tc.other, got, tc.want)
		}
	}
}
