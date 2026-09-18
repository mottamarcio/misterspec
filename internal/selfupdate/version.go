// Package selfupdate implements misterspec's own version reporting and
// self-update mechanics (specs/030-cli-version-update): reading the
// build-time-embedded version, checking GitHub for a newer release,
// verifying its checksum, and atomically replacing the running binary.
package selfupdate

// Version is this build's misterspec version. It is empty by default and
// set at build time via
// "-ldflags -X github.com/mottamarcio/misterspec/internal/selfupdate.Version=$TAG"
// (release.yml) — never a committed file, so it can never drift from the
// actual tag a binary was built from (research.md).
var Version string

// Report returns Version, or a plain "development build" message when no
// version was embedded — never a blank or misleading value (FR-002).
func Report() string {
	if Version == "" {
		return "development build"
	}
	return Version
}
