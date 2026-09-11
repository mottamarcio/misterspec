package installer

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

// TestInstallOne_RejectsMaliciousResourceName is a white-box test of the
// containment guard installOne applies before any write (FR-006). It is
// not reachable through the public Install() today — every real
// Resource.Name comes from kit.TemplatesFS's own entries, which go:embed
// guarantees are safe basenames (see installer_test.go's note) — so this
// exercises the guard directly with a contrived, deliberately malicious
// name to prove it actually works, independent of whether today's
// resource set can ever trigger it.
func TestInstallOne_RejectsMaliciousResourceName(t *testing.T) {
	target := t.TempDir()
	malicious := Resource{Name: "../../../../etc/passwd", Kind: "template", ArtifactType: "program"}

	outcome := installOne(target, malicious, false)

	if outcome.Status != Failed {
		t.Fatalf("installOne() Status = %v, want Failed", outcome.Status)
	}
	if !errors.Is(outcome.Err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("installOne() Err = %v, want errors.Is(err, artifacts.ErrPathOutsideProject)", outcome.Err)
	}
}
