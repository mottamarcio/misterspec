package operations_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func setupReferencesProgram(t *testing.T, root string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md", "---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n")
}

func writeReferencesSubject(t *testing.T, root, frontmatterExtra, body string) {
	t.Helper()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md", ""+
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\n"+frontmatterExtra+"---\n"+body)
}

func TestReferences_FormalOnly(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "depends_on:\n  - SPEC-002\nsupersedes: []\n", "")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Semantic) != 0 {
		t.Errorf("Semantic = %v, want none", result.Semantic)
	}
	if len(result.Formal) != 2 {
		t.Fatalf("Formal = %+v, want 2 entries (parent, depends_on)", result.Formal)
	}
	if result.Formal[0].Relation != "parent" || result.Formal[0].Target.String() != "FEAT-001" {
		t.Errorf("Formal[0] = %+v, want relation=parent target=FEAT-001", result.Formal[0])
	}
	if result.Formal[1].Relation != "depends_on" || result.Formal[1].Target.String() != "SPEC-002" {
		t.Errorf("Formal[1] = %+v, want relation=depends_on target=SPEC-002", result.Formal[1])
	}
}

func TestReferences_SemanticOnly(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	// No parent line here on purpose would break other checks — keep parent,
	// just no depends_on/supersedes, to isolate the semantic-only case.
	writeReferencesSubject(t, root, "", "See [[SPEC-002]] for details.\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Formal) != 1 || result.Formal[0].Relation != "parent" {
		t.Fatalf("Formal = %+v, want exactly [parent FEAT-001]", result.Formal)
	}
	if len(result.Semantic) != 1 {
		t.Fatalf("Semantic = %+v, want 1 entry", result.Semantic)
	}
	if result.Semantic[0].Relation != "wikilink" || result.Semantic[0].Target.String() != "SPEC-002" {
		t.Errorf("Semantic[0] = %+v, want relation=wikilink target=SPEC-002", result.Semantic[0])
	}
}

func TestReferences_FormalAndSemanticCombined(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "depends_on:\n  - SPEC-002\nsupersedes: []\n", "See [[SPEC-002]].\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Formal) != 2 {
		t.Errorf("Formal = %+v, want 2", result.Formal)
	}
	if len(result.Semantic) != 1 {
		t.Errorf("Semantic = %+v, want 1", result.Semantic)
	}
}

func TestReferences_NoRelationshipsIsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n# Nothing here.\n")

	result, err := operations.References(root, cfg, "KNOW-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Formal) != 0 || len(result.Semantic) != 0 {
		t.Errorf("result = %+v, want both empty", result)
	}
}

func TestReferences_SelfReferenceIsReported(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "depends_on:\n  - SPEC-001\nsupersedes: []\n", "See [[SPEC-001]].\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	var sawFormalSelf, sawSemanticSelf bool
	for _, f := range result.Formal {
		if f.Relation == "depends_on" && f.Target.String() == "SPEC-001" {
			sawFormalSelf = true
		}
	}
	for _, s := range result.Semantic {
		if s.Target.String() == "SPEC-001" {
			sawSemanticSelf = true
		}
	}
	if !sawFormalSelf {
		t.Errorf("Formal = %+v, want a self-referencing depends_on entry", result.Formal)
	}
	if !sawSemanticSelf {
		t.Errorf("Semantic = %+v, want a self-referencing wikilink", result.Semantic)
	}
}

func TestReferences_DuplicateWikilinkNotCollapsed(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "", "[[SPEC-002]] and again [[SPEC-002]].\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Semantic) != 2 {
		t.Fatalf("Semantic = %+v, want 2 entries — each occurrence preserved", result.Semantic)
	}
}

func TestReferences_BrokenMalformedAmbiguousWikilinksExcluded(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-a.md", "---\nid: KNOW-005\ntype: knowledge\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-005-b.md", "---\nid: KNOW-005\ntype: knowledge\nstatus: active\n---\n")
	writeReferencesSubject(t, root, "", ""+
		"[[SPEC-999]] broken\n"+
		"[[nodash]] malformed\n"+
		"[[KNOW-005]] ambiguous\n"+
		"[[SPEC-002]] valid\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Semantic) != 1 {
		t.Fatalf("Semantic = %+v, want exactly 1 (only SPEC-002)", result.Semantic)
	}
	if result.Semantic[0].Target.String() != "SPEC-002" {
		t.Errorf("Semantic[0] = %+v, want target=SPEC-002", result.Semantic[0])
	}
}

// TestReferences_OnlyBrokenLinksStillSucceeds directly proves spec.md's
// own User Story 3, Acceptance Scenario 2: a query against an artifact
// whose body contains nothing but unresolvable wikilinks must still
// succeed (ok, no error) — the integrity problem is simply absent from
// the answer, never a query-level failure.
func TestReferences_OnlyBrokenLinksStillSucceeds(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "", "[[SPEC-999]] broken\n[[nodash]] malformed\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error for a body with only broken/malformed links: %v", err)
	}
	if len(result.Semantic) != 0 {
		t.Errorf("Semantic = %+v, want none", result.Semantic)
	}
}

func TestReferences_OutOfScopeTargetIsInvalid(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")

	_, err := operations.References(root, cfg, "TASK-001")
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Fatalf("References() error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}

// wantSubjectPath is SPEC-001's own canonical path, as every
// writeReferencesSubject-based test in this file already writes it.
const wantSubjectPath = "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md"

func TestReferences_SemanticEntryHasSourceOccurrence(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "", "## Related Specs\n\nSee [[SPEC-002]] for details.\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Semantic) != 1 {
		t.Fatalf("Semantic = %+v, want 1 entry", result.Semantic)
	}
	entry := result.Semantic[0]
	if entry.SourcePath != wantSubjectPath {
		t.Errorf("SourcePath = %q, want %q", entry.SourcePath, wantSubjectPath)
	}
	if entry.SourceSection != "Related Specs" {
		t.Errorf("SourceSection = %q, want %q", entry.SourceSection, "Related Specs")
	}
	if entry.SourceLine <= 0 {
		t.Errorf("SourceLine = %d, want a positive file-absolute line", entry.SourceLine)
	}
}

func TestReferences_FormalEntryHasSourcePathButNoSection(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "depends_on:\n  - SPEC-002\nsupersedes: []\n", "")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	for _, entry := range result.Formal {
		if entry.SourcePath != wantSubjectPath {
			t.Errorf("Formal entry %+v: SourcePath = %q, want %q", entry, entry.SourcePath, wantSubjectPath)
		}
		if entry.SourceSection != "" || entry.SourceLine != 0 {
			t.Errorf("Formal entry %+v: want SourceSection empty and SourceLine 0 (spec FR-009 — no line-level wikilink origin)", entry)
		}
	}
}

func TestReferences_TwoOccurrencesFromDifferentSectionsStayDistinct(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupReferencesProgram(t, root)
	writeReferencesSubject(t, root, "", ""+
		"## Requirements\n\nSee [[SPEC-002]] here.\n\n"+
		"## Notes\n\nAlso mentioned here: [[SPEC-002]].\n")

	result, err := operations.References(root, cfg, "SPEC-001")
	if err != nil {
		t.Fatalf("References() unexpected error: %v", err)
	}
	if len(result.Semantic) != 2 {
		t.Fatalf("Semantic = %+v, want 2 distinct occurrences (spec FR-004)", result.Semantic)
	}
	sections := map[string]bool{result.Semantic[0].SourceSection: true, result.Semantic[1].SourceSection: true}
	if !sections["Requirements"] || !sections["Notes"] {
		t.Errorf("Semantic sections = %v, want both %q and %q represented", sections, "Requirements", "Notes")
	}
	if result.Semantic[0].SourceLine == result.Semantic[1].SourceLine {
		t.Errorf("both occurrences report the same SourceLine %d, want two distinct lines", result.Semantic[0].SourceLine)
	}
}
