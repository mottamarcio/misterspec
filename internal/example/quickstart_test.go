// Package example compiles and runs specs/001-core-foundation/quickstart.md's
// end-to-end flow (Detect → ResolvePath → ParseMetadata → Scan → NextID)
// against internal/project, internal/artifacts, and internal/ids, so the
// three packages' composition is verified in CI rather than only by
// prose (plan.md Phase 6, T024).
package example

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestQuickstart_EndToEnd(t *testing.T) {
	// A project with one Program, one Feature, and two Specs already in
	// place — mirroring the fixture quickstart.md's examples describe.
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-011/spec.md",
		"---\nid: SPEC-011\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n# First spec\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\ndepends_on:\n  - SPEC-011\n---\n# Second spec\n")

	// 1. Detect the project.
	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	// 2. Resolve a canonical path.
	prgID, err := ids.Parse(ids.Program, "PRG-001", proj.Config.IDWidth)
	if err != nil {
		t.Fatalf("ids.Parse(program) unexpected error: %v", err)
	}
	featID, err := ids.Parse(ids.Feature, "FEAT-004", proj.Config.IDWidth)
	if err != nil {
		t.Fatalf("ids.Parse(feature) unexpected error: %v", err)
	}
	specID, err := ids.Parse(ids.Spec, "SPEC-014", proj.Config.IDWidth)
	if err != nil {
		t.Fatalf("ids.Parse(spec) unexpected error: %v", err)
	}

	path, err := artifacts.ResolvePath(proj.Root, proj.Config, ids.Spec, specID, prgID, featID)
	if err != nil {
		t.Fatalf("artifacts.ResolvePath() unexpected error: %v", err)
	}
	wantFile := "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md"
	if path.File != wantFile {
		t.Fatalf("ResolvePath().File = %q, want %q", path.File, wantFile)
	}

	// 3. Parse the artifact's metadata.
	meta, err := artifacts.ParseMetadata(filepath.Join(proj.Root, path.File))
	if err != nil {
		t.Fatalf("artifacts.ParseMetadata() unexpected error: %v", err)
	}
	if meta.ID == nil || meta.ID.String() != "SPEC-014" {
		t.Fatalf("meta.ID = %v, want SPEC-014", meta.ID)
	}
	if len(meta.DependsOn) != 1 || meta.DependsOn[0].String() != "SPEC-011" {
		t.Fatalf("meta.DependsOn = %v, want [SPEC-011]", meta.DependsOn)
	}

	// 4. Discover existing IDs and compute the next one — purely by
	// scanning, never a stored counter (Constitution Principle III).
	result, err := ids.Scan(proj.Root, proj.Config, ids.Spec)
	if err != nil {
		t.Fatalf("ids.Scan() unexpected error: %v", err)
	}
	if len(result.IDs) != 2 {
		t.Fatalf("ids.Scan() found %d IDs, want 2: %+v", len(result.IDs), result.IDs)
	}

	next := ids.NextID(result.IDs, ids.Spec, proj.Config.IDWidth)
	if next.String() != "SPEC-015" {
		t.Fatalf("ids.NextID() = %v, want SPEC-015", next)
	}

	// Sanity check: detection from a directory outside any project is
	// reported distinctly, not as a generic error (FR-003).
	if _, err := project.Detect(t.TempDir()); !errors.Is(err, project.ErrNotInitialized) {
		t.Fatalf("project.Detect(empty dir) error = %v, want errors.Is(err, ErrNotInitialized)", err)
	}
}
