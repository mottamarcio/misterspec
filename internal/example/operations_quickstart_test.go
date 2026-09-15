// This file compiles and runs specs/002-read-operations/quickstart.md's
// end-to-end flow (Resolve → Inspect → Parent → Children → Inventory →
// Fingerprint) against internal/operations, on top of the fixture project
// built the same way quickstart_test.go's 001-core-foundation flow was
// (plan.md Phase 6, T021).
package example

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestOperationsQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n# Program\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/feature.md",
		"---\nid: FEAT-004\ntype: feature\nstatus: active\nparent: PRG-001\n---\n# Feature\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md",
		"---\nid: SPEC-014\ntype: spec\nstatus: ready\nparent: FEAT-004\n---\n# Spec\n")
	testutil.WriteFile(t, root, "ai/raw/architecture.md", "# Architecture\n")

	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}

	// 1. Resolve and inspect by ID alone.
	loc, err := operations.Resolve(proj.Root, proj.Config, "SPEC-014")
	if err != nil {
		t.Fatalf("operations.Resolve() unexpected error: %v", err)
	}
	if loc.Path != "ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/spec.md" {
		t.Fatalf("Resolve().Path = %q, unexpected", loc.Path)
	}

	inspected, err := operations.Inspect(proj.Root, proj.Config, "SPEC-014")
	if err != nil {
		t.Fatalf("operations.Inspect() unexpected error: %v", err)
	}
	if inspected.Metadata.Status != "ready" {
		t.Fatalf("Inspect().Metadata.Status = %q, want %q", inspected.Metadata.Status, "ready")
	}

	// 2. Walk structural relationships.
	parent, err := operations.Parent(proj.Root, proj.Config, "SPEC-014")
	if err != nil {
		t.Fatalf("operations.Parent() unexpected error: %v", err)
	}
	if !parent.HasParent || parent.Parent.ID.String() != "FEAT-004" {
		t.Fatalf("Parent() = %+v, want HasParent with FEAT-004", parent)
	}

	featureType := ids.Feature
	children, err := operations.Children(proj.Root, proj.Config, "PRG-001", &featureType)
	if err != nil {
		t.Fatalf("operations.Children() unexpected error: %v", err)
	}
	if len(children) != 1 || children[0].ID.String() != "FEAT-004" {
		t.Fatalf("Children() = %+v, want [FEAT-004]", children)
	}

	// 3. Discover files and verify content.
	files, err := operations.Inventory(proj.Root, proj.Config.RawDir)
	if err != nil {
		t.Fatalf("operations.Inventory() unexpected error: %v", err)
	}
	if len(files) != 1 || files[0].Path != "ai/raw/architecture.md" {
		t.Fatalf("Inventory() = %+v, want [ai/raw/architecture.md]", files)
	}

	fp, err := operations.Fingerprint(proj.Root, "ai/raw/architecture.md")
	if err != nil {
		t.Fatalf("operations.Fingerprint() unexpected error: %v", err)
	}
	if fp.Algorithm != "sha256" || len(fp.Digest) != 64 {
		t.Fatalf("Fingerprint() = %+v, unexpected shape", fp)
	}
}
