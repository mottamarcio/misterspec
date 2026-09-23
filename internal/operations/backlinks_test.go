package operations_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func setupBacklinksProgram(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", "---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")
}

func TestBacklinks_FormalOnly(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Semantic) != 0 {
		t.Errorf("Semantic = %v, want none", result.Semantic)
	}
	if len(result.Formal) != 1 {
		t.Fatalf("Formal = %+v, want 1 entry", result.Formal)
	}
	if result.Formal[0].Relation != "depends_on" || result.Formal[0].Source.String() != "SPEC-002" {
		t.Errorf("Formal[0] = %+v, want relation=depends_on source=SPEC-002", result.Formal[0])
	}
}

func TestBacklinks_SemanticOnly(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\nSee [[SPEC-001]].\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Formal) != 0 {
		t.Errorf("Formal = %+v, want none", result.Formal)
	}
	if len(result.Semantic) != 1 || result.Semantic[0].Relation != "wikilink" || result.Semantic[0].Source.String() != "SPEC-002" {
		t.Errorf("Semantic = %+v, want [{wikilink SPEC-002}]", result.Semantic)
	}
}

func TestBacklinks_FormalAndSemanticCombined(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\nSee [[SPEC-001]].\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Formal) != 1 {
		t.Errorf("Formal = %+v, want 1", result.Formal)
	}
	if len(result.Semantic) != 1 {
		t.Errorf("Semantic = %+v, want 1", result.Semantic)
	}
}

func TestBacklinks_NothingReferencesItIsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-999-unused.md", "---\nid: KNOW-999\ntype: knowledge\nstatus: active\n---\n")

	result, err := operations.Backlinks(root, cfg, "KNOW-999")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Formal) != 0 || len(result.Semantic) != 0 {
		t.Errorf("result = %+v, want both empty", result)
	}
}

func TestBacklinks_SelfReferenceAppearsAsOwnBacklink(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\nSee [[FEAT-001]].\n")

	result, err := operations.Backlinks(root, cfg, "FEAT-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Semantic) != 1 || result.Semantic[0].Source.String() != "FEAT-001" {
		t.Errorf("Semantic = %+v, want a self-referencing wikilink from FEAT-001 itself", result.Semantic)
	}
}

func TestBacklinks_MultipleSourcesOrderedDeterministically(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md",
		"---\nid: SPEC-003\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Formal) != 2 {
		t.Fatalf("Formal = %+v, want 2 entries", result.Formal)
	}
	if result.Formal[0].Source.String() != "SPEC-002" || result.Formal[1].Source.String() != "SPEC-003" {
		t.Errorf("Formal = %+v, want [SPEC-002, SPEC-003] in ascending order", result.Formal)
	}
}

func TestBacklinks_BrokenMalformedAmbiguousWikilinksExcluded(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	// KNOW-005 is ambiguous (claimed twice) — a source linking to it must
	// not register as a backlink of SPEC-001 even though SPEC-001 isn't
	// involved at all; this fixture just proves the source's own
	// unrelated ambiguous link doesn't interfere with unrelated queries.
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-a.md", "---\nid: KNOW-005\ntype: knowledge\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-b.md", "---\nid: KNOW-005\ntype: knowledge\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n[[SPEC-999]] broken\n[[nodash]] malformed\n[[KNOW-005]] ambiguous\n[[SPEC-001]] valid\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Semantic) != 1 || result.Semantic[0].Source.String() != "SPEC-002" {
		t.Fatalf("Semantic = %+v, want exactly one backlink from SPEC-002 (the valid [[SPEC-001]] link)", result.Semantic)
	}
}

// TestBacklinks_ProjectWithOnlyBrokenLinksStillSucceeds directly proves
// spec.md's own User Story 3, Acceptance Scenario 2, from the incoming
// side: a project where every wikilink anywhere is unresolvable must
// still let a backlinks query succeed (ok, no error) rather than fail.
func TestBacklinks_ProjectWithOnlyBrokenLinksStillSucceeds(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n[[SPEC-999]] broken\n[[nodash]] malformed\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error for a project with only broken/malformed links: %v", err)
	}
	if len(result.Semantic) != 0 {
		t.Errorf("Semantic = %+v, want none", result.Semantic)
	}
}

func TestBacklinks_OutOfScopeTargetIsInvalid(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")

	_, err := operations.Backlinks(root, cfg, "TASK-001")
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Fatalf("Backlinks() error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}

func TestBacklinks_SemanticEntryHasSourceOccurrence(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	wantPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md"
	testutil.WriteFile(t, root, wantPath,
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\n## Requirements\n\nSee [[SPEC-001]].\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Semantic) != 1 {
		t.Fatalf("Semantic = %+v, want 1 entry", result.Semantic)
	}
	entry := result.Semantic[0]
	if entry.SourcePath != wantPath {
		t.Errorf("SourcePath = %q, want %q", entry.SourcePath, wantPath)
	}
	if entry.SourceSection != "Requirements" {
		t.Errorf("SourceSection = %q, want %q", entry.SourceSection, "Requirements")
	}
	if entry.SourceLine <= 0 {
		t.Errorf("SourceLine = %d, want a positive file-absolute line", entry.SourceLine)
	}
}

// TestBacklinks_AnchorQualifiedWikilinkHasTargetAnchor proves spec 040
// data-model.md "ReferenceEntry / BacklinkEntry (extended)": an
// anchor-qualified wikilink's BacklinkEntry carries TargetAnchor; a
// plain wikilink to the same target stays "".
func TestBacklinks_AnchorQualifiedWikilinkHasTargetAnchor(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\n---\nSee [[SPEC-001#retry-policy]] and also [[SPEC-001]].\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Semantic) != 2 {
		t.Fatalf("Semantic = %+v, want 2 entries", result.Semantic)
	}
	if result.Semantic[0].TargetAnchor != "retry-policy" {
		t.Errorf("Semantic[0].TargetAnchor = %q, want %q", result.Semantic[0].TargetAnchor, "retry-policy")
	}
	if result.Semantic[1].TargetAnchor != "" {
		t.Errorf("Semantic[1].TargetAnchor = %q, want \"\" for a plain wikilink", result.Semantic[1].TargetAnchor)
	}
}

func TestBacklinks_FormalEntryHasSourcePathButNoSection(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupBacklinksProgram(t, root)
	wantPath := "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md"
	testutil.WriteFile(t, root, wantPath,
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n")

	result, err := operations.Backlinks(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("Backlinks() unexpected error: %v", err)
	}
	if len(result.Formal) != 1 {
		t.Fatalf("Formal = %+v, want 1 entry", result.Formal)
	}
	entry := result.Formal[0]
	if entry.SourcePath != wantPath {
		t.Errorf("SourcePath = %q, want %q", entry.SourcePath, wantPath)
	}
	if entry.SourceSection != "" || entry.SourceLine != 0 {
		t.Errorf("entry %+v: want SourceSection empty and SourceLine 0 (spec FR-009)", entry)
	}
}
