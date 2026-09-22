package impact

import (
	"testing"
	"time"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

func timeoutAfterSeconds(s int) <-chan time.Time {
	return time.After(time.Duration(s) * time.Second)
}

// propagateFixture builds a real fixture repository with:
//   - SPEC-014, which SPEC-020 formally depends_on
//   - SPEC-014's own R1, served by TASK-001 in SPEC-014's tasks.md
//
// mirroring the layout internal/impact/change_set_test.go's fixtures
// already establish for entityIDForDiffPath (042-impact-analysis-review
// T014, Foundational reuse).
func propagateFixture(t *testing.T) (dir string, cfg project.Configuration) {
	t.Helper()
	dir = initFixtureRepo(t)
	cfg = fixtureConfig()

	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"
	dependentSpecDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-020"

	writeFixtureFile(t, dir, specDir+"/spec.md", specFrontmatter+"### R1 — First\n\nOriginal body.\n")
	writeFixtureFile(t, dir, specDir+"/tasks.md", "---\nfor: SPEC-014\ntype: tasks\n---\n## TASK-001 — Do the thing\n\n- [ ] Complete\n\nServes: SPEC-014:R1\n")
	writeFixtureFile(t, dir, dependentSpecDir+"/spec.md", "---\nid: SPEC-020\ntype: spec\nstatus: draft\ndepends_on: [SPEC-014]\n---\n### R1 — Depends\n\nBody.\n")
	commitFixture(t, dir, "base")
	return dir, cfg
}

func specElem(number int, path string, reqNumber *int) ChangedElement {
	return ChangedElement{
		ID:                ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: number, Width: 3},
		Path:              path,
		Status:            Modified,
		RequirementNumber: reqNumber,
	}
}

func intPtr(n int) *int { return &n }

// TestWalkPropagation_FormalBacklink is 042-impact-analysis-review T014
// (US1): a whole-artifact ChangedElement for a Spec with a one-hop
// depends_on backlink produces one PropagationHop{Relation:
// "depends_on"} and one affected item for the dependent Spec.
func TestWalkPropagation_FormalBacklink(t *testing.T) {
	dir, cfg := propagateFixture(t)
	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", nil)

	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}

	paths, ok := result.PathsByItem["SPEC-020"]
	if !ok || len(paths) != 1 {
		t.Fatalf("PathsByItem[SPEC-020] = %+v (ok=%v), want exactly 1 path", paths, ok)
	}
	hops := paths[0].Hops
	if len(hops) != 1 || hops[0].Relation != RelationDependsOn || hops[0].FromID != "SPEC-014" || hops[0].ToID != "SPEC-020" {
		t.Errorf("hops = %+v, want one depends_on hop SPEC-014 -> SPEC-020", hops)
	}
	if !result.HadAnyHop[elem.ID] {
		t.Error("HadAnyHop[SPEC-014] = false, want true")
	}
}

// TestWalkPropagation_CoverageBacklink is 042-impact-analysis-review
// T014 (US1): a Requirement-level ChangedElement with a Task covering
// it produces one PropagationHop{Relation: "coverage"} and one
// affected item for that Task's own composite ID.
func TestWalkPropagation_CoverageBacklink(t *testing.T) {
	dir, cfg := propagateFixture(t)
	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", intPtr(1))

	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}

	paths, ok := result.PathsByItem["SPEC-014:TASK-001"]
	if !ok || len(paths) != 1 {
		t.Fatalf("PathsByItem[SPEC-014:TASK-001] = %+v (ok=%v), want exactly 1 path", paths, ok)
	}
	hops := paths[0].Hops
	if len(hops) != 1 || hops[0].Relation != RelationCoverage || hops[0].ToID != "SPEC-014:TASK-001" {
		t.Errorf("hops = %+v, want one coverage hop to SPEC-014:TASK-001", hops)
	}
}

// TestWalkPropagation_WikilinkBacklinkIsSuggestion is 042-impact-
// analysis-review T020 (US2, spec FR-003/FR-004): a wikilink-only
// backlink to a changed element produces a "wikilink" PropagationHop,
// classified suggested_review — never deterministic_invalidation,
// regardless of how many wikilinks point to it.
func TestWalkPropagation_WikilinkBacklinkIsSuggestion(t *testing.T) {
	dir, cfg := propagateFixture(t)
	knowDir := "ai/knowledge"
	writeFixtureFile(t, dir, knowDir+"/KNOW-001-mentions-spec.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\nSee [[SPEC-014]] for details.\n")
	commitFixture(t, dir, "add wikilink mention")

	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", nil)
	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}

	paths, ok := result.PathsByItem["KNOW-001"]
	if !ok || len(paths) == 0 {
		t.Fatalf("PathsByItem[KNOW-001] = %+v (ok=%v), want at least one path", paths, ok)
	}
	lastHop := paths[0].Hops[len(paths[0].Hops)-1]
	if lastHop.Relation != RelationWikilink {
		t.Fatalf("last hop relation = %q, want %q", lastHop.Relation, RelationWikilink)
	}
	if ClassifyRelation(lastHop.Relation) != SuggestedReview {
		t.Errorf("ClassifyRelation(wikilink hop) = %q, want %q", ClassifyRelation(lastHop.Relation), SuggestedReview)
	}

	// The same run's formal depends_on hop (SPEC-020) is unaffected and
	// still classifies deterministic — the two relation kinds never
	// blend into one classification for the run as a whole.
	depPaths, ok := result.PathsByItem["SPEC-020"]
	if !ok || len(depPaths) == 0 {
		t.Fatalf("PathsByItem[SPEC-020] missing — the depends_on relation should be unaffected by adding wikilink handling")
	}
	if ClassifyRelation(depPaths[0].Hops[len(depPaths[0].Hops)-1].Relation) != DeterministicInvalidation {
		t.Error("SPEC-020's own depends_on path classification changed after adding wikilink support, want unaffected")
	}
}

// TestWalkPropagation_TransitiveChain is 042-impact-analysis-review
// T023 (US3, spec FR-005): a two-hop dependency chain (C depends_on B
// depends_on A, A changed) produces one PropagationPath for C whose
// first hop originates at A and whose last hop reaches C, with exactly
// 2 hops — not just the nearest (B) hop.
func TestWalkPropagation_TransitiveChain(t *testing.T) {
	dir, cfg := propagateFixture(t)
	// propagateFixture already gives us SPEC-014 (A) and SPEC-020 (B,
	// depends_on SPEC-014). Add SPEC-030 (C), depends_on SPEC-020.
	writeFixtureFile(t, dir, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-030/spec.md",
		"---\nid: SPEC-030\ntype: spec\nstatus: draft\ndepends_on: [SPEC-020]\n---\n### R1 — Third\n\nBody.\n")
	commitFixture(t, dir, "add SPEC-030 depending on SPEC-020")

	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", nil)
	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}

	paths, ok := result.PathsByItem["SPEC-030"]
	if !ok || len(paths) == 0 {
		t.Fatalf("PathsByItem[SPEC-030] = %+v (ok=%v), want at least one path", paths, ok)
	}
	hops := paths[0].Hops
	if len(hops) != 2 {
		t.Fatalf("hops = %+v, want exactly 2 (SPEC-014->SPEC-020->SPEC-030)", hops)
	}
	if hops[0].FromID != "SPEC-014" || hops[0].ToID != "SPEC-020" {
		t.Errorf("hops[0] = %+v, want SPEC-014 -> SPEC-020", hops[0])
	}
	if hops[1].FromID != "SPEC-020" || hops[1].ToID != "SPEC-030" {
		t.Errorf("hops[1] = %+v, want SPEC-020 -> SPEC-030", hops[1])
	}
}

// TestWalkPropagation_MultiplePathsToSameItemKeptSeparate is
// 042-impact-analysis-review T023 (US3, spec FR-005): the same target
// reached by two distinct relation types keeps two distinct
// PropagationPaths, never collapsed into one.
func TestWalkPropagation_MultiplePathsToSameItemKeptSeparate(t *testing.T) {
	dir, cfg := propagateFixture(t)
	// SPEC-020 already depends_on SPEC-014 (formal). Add a second,
	// independent wikilink relation from SPEC-020 to SPEC-014 too.
	writeFixtureFile(t, dir, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-020/spec.md",
		"---\nid: SPEC-020\ntype: spec\nstatus: draft\ndepends_on: [SPEC-014]\n---\n### R1 — Depends\n\nAlso see [[SPEC-014]] directly.\n")
	commitFixture(t, dir, "add a second relation SPEC-020 -> SPEC-014")

	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", nil)
	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}

	paths := result.PathsByItem["SPEC-020"]
	if len(paths) != 2 {
		t.Fatalf("PathsByItem[SPEC-020] has %d paths, want exactly 2 (depends_on and wikilink kept separate): %+v", len(paths), paths)
	}
}

// TestWalkPropagation_CycleTerminates is 042-impact-analysis-review
// T023 (US3, spec FR-006): two artifacts wikilinking each other
// terminate the walk without expanding either more than once, and the
// originally-changed element is never itself reported as affected.
func TestWalkPropagation_CycleTerminates(t *testing.T) {
	dir, cfg := propagateFixture(t)
	knowDir := "ai/knowledge"
	writeFixtureFile(t, dir, knowDir+"/KNOW-001-a.md", "---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\nSee [[SPEC-014]].\n")
	commitFixture(t, dir, "add KNOW-001 wikilinking SPEC-014")

	// SPEC-014's own spec.md does not link back to KNOW-001 in this
	// fixture, but the walk must still terminate cleanly regardless —
	// exercised together with TestAnalyzeImpact_CycleTerminates, which
	// constructs a genuine mutual A<->B cycle end-to-end.
	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", nil)

	done := make(chan struct{})
	var result propagationResult
	var err error
	go func() {
		result, err = walkPropagation(dir, cfg, []ChangedElement{elem})
		close(done)
	}()
	select {
	case <-done:
	case <-timeoutAfterSeconds(5):
		t.Fatal("walkPropagation() did not terminate within 5s")
	}
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}
	if _, ok := result.PathsByItem["SPEC-014"]; ok {
		t.Error("SPEC-014 (the originally changed element) appears in its own PathsByItem, want it excluded")
	}
}

// TestWalkPropagation_CoverageHopUsesEvidenceRelationWhenTaskAlreadyStale
// is a code-review finding fix (042-impact-analysis-review research.md
// #10): a Task covering a changed Requirement, whose own 041 evidence
// is already Stale against its own current content (for reasons
// unrelated to this run's change), is reached via a RelationEvidence
// hop — not RelationCoverage — carrying its own TaskEvidenceState, so
// classify.go's SeverityHigh applies and the reason can name the
// distinct problem.
func TestWalkPropagation_CoverageHopUsesEvidenceRelationWhenTaskAlreadyStale(t *testing.T) {
	dir, cfg := propagateFixture(t)
	specDir := "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014"

	// Give TASK-001 a checked box with a stale Evidence-Fingerprint (it
	// won't match the Task's own current content), so its own 041
	// EvidenceState is Stale — independent of the R1 change this test
	// then applies.
	writeFixtureFile(t, dir, specDir+"/tasks.md",
		"---\nfor: SPEC-014\ntype: tasks\n---\n"+
			"## TASK-001 — Do the thing\n\n"+
			"- [x] Complete\n\n"+
			"Serves: SPEC-014:R1\n\n"+
			"Evidence-Result: pass\n"+
			"Evidence-Fingerprint: sha256:0000000000000000000000000000000000000000000000000000000000000000\n")
	commitFixture(t, dir, "make TASK-001 evidence stale")

	elem := specElem(14, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-014/spec.md", intPtr(1))
	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}

	paths, ok := result.PathsByItem["SPEC-014:TASK-001"]
	if !ok || len(paths) == 0 {
		t.Fatalf("PathsByItem[SPEC-014:TASK-001] missing")
	}
	lastHop := paths[0].Hops[len(paths[0].Hops)-1]
	if lastHop.Relation != RelationEvidence {
		t.Errorf("Relation = %q, want %q", lastHop.Relation, RelationEvidence)
	}
	if lastHop.TaskEvidenceState == "" {
		t.Error("TaskEvidenceState is empty, want the Task's own non-Verified EvidenceState")
	}
	if ClassifyRelation(lastHop.Relation) != DeterministicInvalidation {
		t.Errorf("ClassifyRelation(evidence) = %q, want %q", ClassifyRelation(lastHop.Relation), DeterministicInvalidation)
	}
	if SeverityFor(DeterministicInvalidation, RelationEvidence) != SeverityHigh {
		t.Errorf("SeverityFor(deterministic, evidence) = %q, want %q", SeverityFor(DeterministicInvalidation, RelationEvidence), SeverityHigh)
	}
}

// TestWalkPropagation_NoKnownRelation is 042-impact-analysis-review
// T014 (US1, spec FR-008): a ChangedElement with zero backlinks/
// coverage produces zero affected items and HadAnyHop == false for it.
func TestWalkPropagation_NoKnownRelation(t *testing.T) {
	dir, cfg := propagateFixture(t)
	// SPEC-020 has no dependents and its own R1 is not served by any
	// Task (no tasks.md written for it in the fixture).
	elem := specElem(20, "ai/programs/PROG-001/features/FEAT-002/specs/SPEC-020/spec.md", nil)

	result, err := walkPropagation(dir, cfg, []ChangedElement{elem})
	if err != nil {
		t.Fatalf("walkPropagation() unexpected error: %v", err)
	}
	if len(result.PathsByItem) != 0 {
		t.Errorf("PathsByItem = %+v, want none", result.PathsByItem)
	}
	if result.HadAnyHop[elem.ID] {
		t.Error("HadAnyHop[SPEC-020] = true, want false — SPEC-020 has no known relation in this fixture")
	}
}
