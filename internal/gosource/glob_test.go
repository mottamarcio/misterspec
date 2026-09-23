package gosource

import "testing"

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"vendor/**", "vendor/foo/bar.go", true},
		{"vendor/**", "internal/foo/bar.go", false},
		{"**/*_generated.go", "internal/foo/bar_generated.go", true},
		{"**/*_generated.go", "internal/foo/bar.go", false},
		{"internal/cli/internalcmd", "internal/cli/internalcmd", true},
		{"internal/cli/internalcmd", "internal/cli/internalcmd/context.go", false},
		// Code-review regression: a multi-byte rune in the pattern must
		// not be split across bytes and quoted independently — that
		// previously produced a mismatched pattern that silently never
		// matched (byte-indexed pattern[i] on a UTF-8 string).
		{"pastá/**", "pastá/foo.go", true},
		{"pastá/**", "pasta/foo.go", false},
	}
	for _, tc := range cases {
		if got := MatchGlob(tc.pattern, tc.path); got != tc.want {
			t.Errorf("MatchGlob(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}
