package installer

import (
	"os"
	"path/filepath"
)

// WriteAtomicFile writes content to targetAbs following the same
// atomic-write recipe as every prior feature's writes
// (docs/architecture-specification.md §57): create a temp file in the
// same directory, write the complete content, fsync, then os.Rename into
// place. A reader never observes a partially written file, and a failure
// partway through never leaves a truncated or corrupted file at
// targetAbs (FR-003, FR-007).
//
// This is installer's own small implementation, not imported from
// internal/operations — see specs/005-embedded-kit/research.md for why
// duplicating this ~20-line recipe is preferred here over an
// architecturally backwards dependency on a higher-level package.
func WriteAtomicFile(targetAbs string, content []byte) error {
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

	return os.Rename(tmpPath, targetAbs)
}
