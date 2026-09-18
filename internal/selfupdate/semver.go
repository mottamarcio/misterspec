package selfupdate

import "golang.org/x/mod/semver"

// Newer reports whether latest is a real semver-newer version than
// current, using golang.org/x/mod/semver's own precedence rules rather
// than a hand-rolled comparison — this project's own real tag history
// mixes suffixed and bare tags (v1.1.5-alpha immediately followed by
// v1.2.0), which naive comparison would order incorrectly (research.md).
// An invalid/unparseable current or latest (e.g. "development build")
// is treated as "no update available" rather than an error here — the
// caller decides how to report that.
func Newer(current, latest string) bool {
	c, l := withVPrefix(current), withVPrefix(latest)
	if !semver.IsValid(c) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(c, l) < 0
}

func withVPrefix(v string) string {
	if v == "" || v[0] == 'v' {
		return v
	}
	return "v" + v
}
