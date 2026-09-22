package operations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
)

// FileFingerprint is a file's algorithm-tagged content digest (FR-012).
//
// Named FileFingerprint rather than Fingerprint (contracts/operations.md's
// original draft used the same identifier for both the type and the
// function below, which Go does not allow — corrected here during
// implementation; see specs/002-read-operations/tasks.md T022).
type FileFingerprint struct {
	// Algorithm is always "sha256" for this feature.
	Algorithm string
	// Digest is the lowercase hex-encoded SHA-256 digest.
	Digest string
}

// String renders f as "sha256:<hex>".
func (f FileFingerprint) String() string {
	return fmt.Sprintf("%s:%s", f.Algorithm, f.Digest)
}

// Fingerprint computes path's content digest, streamed via io.Copy rather
// than fully buffered, without interpreting path's contents in any way —
// it supports any file type (FR-012). It rejects a path that would
// resolve outside root (FR-013, via artifacts.RelativeWithinRoot) and
// reports artifacts.ErrArtifactNotFound for one that does not exist,
// reusing artifacts's vocabulary rather than a parallel sentinel for the
// same condition.
func Fingerprint(root, path string) (FileFingerprint, error) {
	rel, err := artifacts.RelativeWithinRoot(root, path)
	if err != nil {
		return FileFingerprint{}, err
	}
	full := filepath.Join(root, rel)

	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return FileFingerprint{}, fmt.Errorf("%w: %s", artifacts.ErrArtifactNotFound, rel)
		}
		return FileFingerprint{}, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return FileFingerprint{}, err
	}

	return FileFingerprint{
		Algorithm: "sha256",
		Digest:    hex.EncodeToString(h.Sum(nil)),
	}, nil
}

// digestBytes computes data's own SHA-256 digest directly, with no I/O
// — the core Fingerprint already performed via io.Copy, factored out
// so TaskContentFingerprint (041-task-evidence-fingerprint) can reuse
// it without a fake file to stream from (research.md #6).
func digestBytes(data []byte) FileFingerprint {
	sum := sha256.Sum256(data)
	return FileFingerprint{
		Algorithm: "sha256",
		Digest:    hex.EncodeToString(sum[:]),
	}
}

// TaskContentFingerprint hashes body (a Task's own already-extracted
// Section.Body) using the same SHA-256 algorithm as Fingerprint, via
// digestBytes — pure, no I/O, unlike Fingerprint itself
// (041-task-evidence-fingerprint contracts §2, data-model.md
// "FileFingerprint (extended)"). task is accepted for API symmetry
// with the rest of this feature's Task-scoped operations, even though
// the digest itself only depends on body.
//
// body is stripped of any existing "Evidence-*:" lines first
// (evidence.StripEvidenceLines) — correction found during
// implementation: without this, a Task's own previously-recorded
// Evidence-Fingerprint: line would be part of what gets fingerprinted,
// making Verified unreachable the instant evidence is written (writing
// the fingerprint would change the very content it describes).
func TaskContentFingerprint(task ids.EntityID, body []byte) FileFingerprint {
	return digestBytes(evidence.StripEvidenceLines(body))
}
