package secrets

import "testing"

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2", "1.2", 0},
		{"1.2", "1.2.3", -1},
		{"1.2.3", "1.2", 1},
		{"1.2.0", "1.2", 0},
		{"1.10", "1.2", 1},
		{"1.2", "1.10", -1},
	}

	for _, tc := range cases {
		if got := CompareVersions(tc.v1, tc.v2); got != tc.expected {
			t.Fatalf("CompareVersions(%q, %q) = %d, want %d", tc.v1, tc.v2, got, tc.expected)
		}
	}
}
