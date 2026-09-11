package operations

import (
	"io"
	"os"
	"path/filepath"
)

// writeAtomic writes content to targetAbs following the atomic-write
// recipe (docs/architecture-specification.md §57): create a temp file in
// the same directory, write the complete content, fsync, then
// os.Rename into place. A reader never observes a partially written
// file, and a failure partway through never leaves a truncated or
// corrupted file at targetAbs (FR-007). Shared by Create and
// CreateArtifact rather than duplicated (Constitution Principle VI).
func writeAtomic(targetAbs, content string) error {
	dir := filepath.Dir(targetAbs)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	// Best-effort cleanup: a no-op once the rename below succeeds, since
	// the file no longer exists at tmpPath at that point.
	defer os.Remove(tmpPath)

	if _, err := io.WriteString(tmp, content); err != nil {
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
