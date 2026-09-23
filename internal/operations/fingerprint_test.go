package operations_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestFingerprint_DeterministicAcrossRepeatedCalls(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/raw/architecture.md", "# Architecture\n\nSome content.\n")

	first, err := operations.Fingerprint(root, "ai/raw/architecture.md")
	if err != nil {
		t.Fatalf("Fingerprint() unexpected error: %v", err)
	}
	second, err := operations.Fingerprint(root, "ai/raw/architecture.md")
	if err != nil {
		t.Fatalf("Fingerprint() unexpected error: %v", err)
	}
	if first != second {
		t.Errorf("Fingerprint() not deterministic: %+v != %+v", first, second)
	}
	if first.Algorithm != "sha256" {
		t.Errorf("Algorithm = %q, want %q", first.Algorithm, "sha256")
	}
	if len(first.Digest) != 64 { // sha256 hex length
		t.Errorf("Digest length = %d, want 64", len(first.Digest))
	}
	if !strings.HasPrefix(first.String(), "sha256:") {
		t.Errorf("String() = %q, want a sha256: prefix", first.String())
	}
}

func TestFingerprint_DifferentContentDifferentDigest(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/raw/a.md", "content A\n")
	testutil.WriteFile(t, root, "ai/raw/b.md", "content B\n")

	a, err := operations.Fingerprint(root, "ai/raw/a.md")
	if err != nil {
		t.Fatalf("Fingerprint(a) unexpected error: %v", err)
	}
	b, err := operations.Fingerprint(root, "ai/raw/b.md")
	if err != nil {
		t.Fatalf("Fingerprint(b) unexpected error: %v", err)
	}
	if a.Digest == b.Digest {
		t.Error("Fingerprint() produced the same digest for different content")
	}
}

func TestFingerprint_NotFound(t *testing.T) {
	root := testutil.Project(t)

	_, err := operations.Fingerprint(root, "ai/raw/missing.md")
	if !errors.Is(err, artifacts.ErrArtifactNotFound) {
		t.Fatalf("Fingerprint() error = %v, want errors.Is(err, artifacts.ErrArtifactNotFound)", err)
	}
}

// TestTaskContentFingerprint_DeterministicAndPure is
// 041-task-evidence-fingerprint T006 (Foundational): TaskContentFingerprint
// is a pure function — no I/O, same body always produces the same
// digest, regardless of how many times it is called.
func TestTaskContentFingerprint_DeterministicAndPure(t *testing.T) {
	task := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3}
	body := []byte("- [x] Complete\nServes: SPEC-014:R1\n")

	first := operations.TaskContentFingerprint(task, body)
	second := operations.TaskContentFingerprint(task, body)
	if first != second {
		t.Errorf("TaskContentFingerprint() not deterministic: %+v != %+v", first, second)
	}
	if first.Algorithm != "sha256" {
		t.Errorf("Algorithm = %q, want %q", first.Algorithm, "sha256")
	}
	if len(first.Digest) != 64 {
		t.Errorf("Digest length = %d, want 64", len(first.Digest))
	}
}

// TestTaskContentFingerprint_DifferentBodyDifferentDigest proves a
// different Task body produces a different fingerprint.
func TestTaskContentFingerprint_DifferentBodyDifferentDigest(t *testing.T) {
	task := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3}

	a := operations.TaskContentFingerprint(task, []byte("- [x] Complete\nScope: internal/foo\n"))
	b := operations.TaskContentFingerprint(task, []byte("- [x] Complete\nScope: internal/bar\n"))
	if a.Digest == b.Digest {
		t.Error("TaskContentFingerprint() produced the same digest for different Task bodies")
	}
}

// TestFingerprint_UnaffectedByTaskContentFingerprintExtraction is a
// regression case: extracting the shared digestBytes helper out of
// Fingerprint's own body must not change Fingerprint's own public
// behavior.
func TestFingerprint_UnaffectedByTaskContentFingerprintExtraction(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/raw/regression.md", "unchanged behavior\n")

	got, err := operations.Fingerprint(root, "ai/raw/regression.md")
	if err != nil {
		t.Fatalf("Fingerprint() unexpected error: %v", err)
	}
	if got.Algorithm != "sha256" || len(got.Digest) != 64 {
		t.Errorf("Fingerprint() = %+v, want a well-formed sha256 fingerprint", got)
	}
}

// TestTaskContentFingerprint_MatchesEvidenceContentFingerprint proves
// operations.TaskContentFingerprint and evidence.ContentFingerprint
// (the second, independent implementation needed because
// internal/validation cannot import internal/operations — see
// internal/evidence/state.go's own doc comment) compute the identical
// digest for the same bytes, so a fingerprint recorded via one and
// compared via the other never spuriously disagrees.
func TestTaskContentFingerprint_MatchesEvidenceContentFingerprint(t *testing.T) {
	task := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3}
	body := []byte("- [x] Complete\nServes: SPEC-014:R1\n")

	fromOperations := operations.TaskContentFingerprint(task, body).String()
	fromEvidence := evidence.ContentFingerprint(body)
	if fromOperations != fromEvidence {
		t.Errorf("operations.TaskContentFingerprint() = %q, evidence.ContentFingerprint() = %q, want identical", fromOperations, fromEvidence)
	}
}

func TestFingerprint_RejectsTraversalOutsideRoot(t *testing.T) {
	root := testutil.Project(t)

	_, err := operations.Fingerprint(root, "../../etc/passwd")
	if !errors.Is(err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("Fingerprint() error = %v, want errors.Is(err, artifacts.ErrPathOutsideProject)", err)
	}
}
