// This file compiles and runs specs/014-sqlite-index/quickstart.md's
// end-to-end flow (initial build, search, an incremental sync after a
// targeted edit, a deletion, an Outgoing/Incoming lookup matching
// operations.References/Backlinks, a full rebuild, and an artifact that
// fails to parse not halting the run) against internal/context/index,
// on top of the fixture project the other quickstart tests in this
// package use (plan.md Phase 6, T022).
package example

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestSQLiteIndexQuickstart_EndToEnd(t *testing.T) {
	root := testutil.Project(t)
	proj, err := project.Detect(root)
	if err != nil {
		t.Fatalf("project.Detect() unexpected error: %v", err)
	}
	cfg := proj.Config

	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nRefresh token rotation.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n## Intent\n\nAnother spec.\n")

	store, err := index.Open(filepath.Join(root, ".misterspec/cache/context.db"))
	if err != nil {
		t.Fatalf("index.Open() unexpected error: %v", err)
	}
	defer store.Close()

	// 1. Build the index (quickstart.md §1).
	report, err := store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}
	if report.Indexed != 4 {
		t.Fatalf("Sync() report.Indexed = %d, want 4", report.Indexed)
	}

	// 2. Search it (quickstart.md §2).
	results, err := store.Search("rotation", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Path != "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md" {
		t.Fatalf("Search(\"rotation\") = %+v, unexpected", results)
	}

	// 3. Change one artifact, synchronize incrementally (quickstart.md §3).
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nSomething else now.\n")
	report, err = store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() (incremental) unexpected error: %v", err)
	}
	if report.Updated != 1 || report.Skipped != 3 {
		t.Fatalf("Sync() (incremental) report = %+v, want Updated=1 Skipped=3", report)
	}

	// 4. Delete an artifact, synchronize (quickstart.md §4).
	if err := os.Remove(filepath.Join(root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md")); err != nil {
		t.Fatalf("removing SPEC-002: %v", err)
	}
	report, err = store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() (after deletion) unexpected error: %v", err)
	}
	if report.Removed != 1 {
		t.Fatalf("Sync() (after deletion) report.Removed = %d, want 1", report.Removed)
	}

	// 5. Look up a relationship from the index, matching 012's own
	// direct computation (quickstart.md §5) — recreate the fixture's
	// FEAT-001 -> PRG-001 parent edge, since SPEC-002 (the only other
	// relationship source) was just removed.
	want, err := operations.References(root, cfg, "FEAT-001")
	if err != nil {
		t.Fatalf("operations.References() unexpected error: %v", err)
	}
	got, err := store.Outgoing("FEAT-001")
	if err != nil {
		t.Fatalf("Outgoing() unexpected error: %v", err)
	}
	if len(got) != len(want.Formal)+len(want.Semantic) {
		t.Fatalf("Outgoing(FEAT-001) = %+v, want len %d to match operations.References", got, len(want.Formal)+len(want.Semantic))
	}

	// 6. Discard and rebuild from scratch (quickstart.md §6).
	report, err = store.Rebuild(root, cfg)
	if err != nil {
		t.Fatalf("Rebuild() unexpected error: %v", err)
	}
	if report.Indexed != 3 {
		t.Fatalf("Rebuild() report.Indexed = %d, want 3 (PRG-001, FEAT-001, SPEC-001)", report.Indexed)
	}
	afterRebuild, err := store.Search("Something else", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(afterRebuild) != 1 {
		t.Fatalf("Search() after rebuild = %+v, want exactly 1", afterRebuild)
	}

	// 7. An artifact that fails to parse doesn't stop the run
	// (quickstart.md §7).
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-broken.md", "no frontmatter at all\n")
	report, err = store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() (with a broken artifact) unexpected error: %v", err)
	}
	if len(report.Errors) != 1 {
		t.Fatalf("Sync() report.Errors = %+v, want exactly 1", report.Errors)
	}
	if report.Errors[0].Path != "ai/knowledge/KNOW-001-broken.md" {
		t.Errorf("report.Errors[0].Path = %q, unexpected", report.Errors[0].Path)
	}
}
