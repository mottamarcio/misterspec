package installer

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/kit"
)

// TestInstallOneFS_RejectsMaliciousResourceName is a white-box test of
// the containment guard installOneFS applies before any write (FR-006).
// It is not reachable through the public Install()/InstallFS() today —
// every real Resource.Name comes from an embedded fs.FS's own entries
// (kit.TemplatesFS, or a test's fixture), which go:embed/fstest both
// guarantee are safe basenames (see installer_test.go's note) — so this
// exercises the guard directly with a contrived, deliberately malicious
// name to prove it actually works, independent of whether today's
// resource set can ever trigger it.
func TestInstallOneFS_RejectsMaliciousResourceName(t *testing.T) {
	target := t.TempDir()
	malicious := Resource{Name: "../../../../etc/passwd", Kind: "template", ArtifactType: "program"}

	outcome := installOneFS(kit.TemplatesFS, templatesDir, target, malicious, false)

	if outcome.Status != Failed {
		t.Fatalf("installOneFS() Status = %v, want Failed", outcome.Status)
	}
	if !errors.Is(outcome.Err, artifacts.ErrPathOutsideProject) {
		t.Fatalf("installOneFS() Err = %v, want errors.Is(err, artifacts.ErrPathOutsideProject)", outcome.Err)
	}
}
