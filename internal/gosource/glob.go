package gosource

import (
	"regexp"
	"strings"
)

// MatchGlob reports whether path matches pattern, where "**" matches
// any sequence of characters (including "/") and "*" matches any
// sequence of characters except "/" — a small, self-contained
// translation to regexp, shared by internal/context/index (path
// exclusions) and internal/architecture (rule From/To patterns) rather
// than duplicated in each (044-architecture-code-context-rules
// research.md #4, Constitution Principle VI).
func MatchGlob(pattern, path string) bool {
	var b strings.Builder
	b.WriteString("^")
	runes := []rune(pattern)
	i := 0
	for i < len(runes) {
		switch {
		case runes[i] == '*' && i+1 < len(runes) && runes[i+1] == '*':
			b.WriteString(".*")
			i += 2
		case runes[i] == '*':
			b.WriteString("[^/]*")
			i++
		case runes[i] == '?':
			b.WriteString(".")
			i++
		default:
			b.WriteString(regexp.QuoteMeta(string(runes[i])))
			i++
		}
	}
	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return false
	}
	return re.MatchString(path)
}

// MatchAnyGlob reports whether path matches any of patterns.
func MatchAnyGlob(patterns []string, path string) bool {
	for _, p := range patterns {
		if MatchGlob(p, path) {
			return true
		}
	}
	return false
}
