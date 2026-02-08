package secrets

import (
	"strings"
	"testing"
)

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

func TestMergeSecretsContent(t *testing.T) {
	local := strings.Join([]string{
		"BAR=\"local\" # ENVSYNC_UPDATED_AT=2025-01-01T00:00:00Z",
		"FOO=\"local\" # ENVSYNC_UPDATED_AT=2025-01-01T00:00:00Z",
		"# comment",
		"",
	}, "\n")
	remote := strings.Join([]string{
		"FOO=\"remote\" # ENVSYNC_UPDATED_AT=2025-01-02T00:00:00Z",
		"BAZ=\"remote\" # ENVSYNC_UPDATED_AT=2025-01-03T00:00:00Z",
		"BAR=\"remote\" # ENVSYNC_UPDATED_AT=2024-12-31T23:59:59Z",
	}, "\n")

	got := MergeSecretsContent(local, remote)
	expected := strings.Join([]string{
		"BAR=\"local\" # ENVSYNC_UPDATED_AT=2025-01-01T00:00:00Z",
		"BAZ=\"remote\" # ENVSYNC_UPDATED_AT=2025-01-03T00:00:00Z",
		"FOO=\"remote\" # ENVSYNC_UPDATED_AT=2025-01-02T00:00:00Z",
	}, "\n")

	if got != expected {
		t.Fatalf("MergeSecretsContent mismatch:\nexpected:\n%s\n\ngot:\n%s", expected, got)
	}
}
