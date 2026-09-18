package installer

import (
	"os"
	"path/filepath"
)

// DefaultFileMode is the file permission every pre-030 caller of
// WriteAtomicFile relied on implicitly (os.CreateTemp's own default),
// preserved explicitly now that WriteAtomicFile takes a mode parameter
// (specs/030-cli-version-update/research.md, decision 6).
const DefaultFileMode = os.FileMode(0o600)

// WriteAtomicFile writes content to targetAbs, with the given file mode,
// following the same atomic-write recipe as every prior feature's writes
// (docs/architecture-specification.md §57): create a temp file in the
// same directory, write the complete content, fsync, chmod, then
// os.Rename into place. A reader never observes a partially written
// file, and a failure partway through never leaves a truncated or
// corrupted file at targetAbs (FR-003, FR-007). The mode parameter lets
// an executable binary replace (specs/030-cli-version-update) request
// 0o755 instead of DefaultFileMode, reusing this recipe rather than
// duplicating it (Constitution Principle VI, DRY).
//
// This is installer's own small implementation, not imported from
// internal/operations — see specs/005-embedded-kit/research.md for why
// duplicating this ~20-line recipe is preferred here over an
// architecturally backwards dependency on a higher-level package.
func WriteAtomicFile(targetAbs string, content []byte, mode os.FileMode) error {
	dir := filepath.Dir(targetAbs)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once the rename below succeeds

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}

	return os.Rename(tmpPath, targetAbs)
}
