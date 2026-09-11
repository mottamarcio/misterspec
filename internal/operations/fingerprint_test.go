package operations_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
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

func TestFingerprint_RejectsTraversalOutsideRoot(t *testing.T) {
	root := testutil.Project(t)

	_, err := operations.Fingerprint(root, "../../etc/passwd")
	if !errors.Is(err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("Fingerprint() error = %v, want errors.Is(err, artifacts.ErrPathOutsideProject)", err)
	}
}
