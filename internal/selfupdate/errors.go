package selfupdate

import "errors"

// ErrReleaseCheckFailed is returned when the GitHub Releases API call
// itself fails (network, outage, rate limit) — contracts/version-update.md's
// "release_check_failed" (FR-012).
var ErrReleaseCheckFailed = errors.New("selfupdate: release check failed")

// ErrNoCompatibleAsset is returned when the latest release has no asset
// matching the current OS/architecture — contracts/version-update.md's
// "no_compatible_asset" (FR-012).
var ErrNoCompatibleAsset = errors.New("selfupdate: no compatible asset for this platform")

// ErrChecksumMismatch is returned when a downloaded binary's own SHA-256
// does not match the published SHA256SUMS entry — contracts/version-update.md's
// "checksum_mismatch" (FR-007, FR-008).
var ErrChecksumMismatch = errors.New("selfupdate: checksum mismatch")
