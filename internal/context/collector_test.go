package contextengine

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func hasReason(candidates []Candidate, tier Tier, relation string) bool {
	for _, c := range candidates {
		for _, r := range c.Reasons {
			if r.Tier == tier && r.Relation == relation {
				return true
			}
		}
	}
	return false
}

func TestCollect_TargetWithNoConnectionsStillReturnsMandatoryBaseline(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/memory/constitution.md",
		"---\ntype: constitution\n---\n## Principles\n\nFilesystem is the source of truth.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome facts.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	if !hasReason(set.Candidates, TierMandatory, "target") {
		t.Errorf("Candidates = %+v, want a TierMandatory/target entry", set.Candidates)
	}
	if !hasReason(set.Candidates, TierMandatory, "constitution") {
		t.Errorf("Candidates = %+v, want a TierMandatory/constitution entry", set.Candidates)
	}
}

// TestCollect_CandidateLinesAreFileAbsoluteNotBodyRelative covers
// 033-context-pack-output-contract research.md Decision 1: chunkArtifact
// must add the frontmatter's own line offset to every Chunk's
// StartLine/EndLine before building a Candidate — verified against a
// fixture's own real, file-absolute line numbers, not the numbers
// artifacts.ParseDocument would report relative to the post-frontmatter
// body alone.
func TestCollect_CandidateLinesAreFileAbsoluteNotBodyRelative(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", ""+
		"---\n"+ // file line 1
		"id: KNOW-001\n"+ // file line 2
		"type: knowledge\n"+ // file line 3
		"status: active\n"+ // file line 4
		"---\n"+ // file line 5
		"## Summary\n"+ // file line 6 — body-relative line 1
		"\n"+
		"Some facts.\n"+ // file line 8
		"\n"+
		"## Details\n"+ // file line 10 — body-relative line 5
		"\n"+
		"More facts.\n") // file line 12
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var summary, details *Candidate
	for i := range set.Candidates {
		switch set.Candidates[i].Heading {
		case "Summary":
			summary = &set.Candidates[i]
		case "Details":
			details = &set.Candidates[i]
		}
	}
	if summary == nil || details == nil {
		t.Fatalf("Candidates = %+v, want both a Summary and a Details heading", set.Candidates)
	}
	if summary.StartLine != 6 {
		t.Errorf("Summary.StartLine = %d, want 6 (file-absolute) — got body-relative %d instead if this reads 1", summary.StartLine, summary.StartLine)
	}
	if details.StartLine != 10 {
		t.Errorf("Details.StartLine = %d, want 10 (file-absolute) — got body-relative %d instead if this reads 5", details.StartLine, details.StartLine)
	}
}

func TestCollect_NoConstitutionIsNotAnError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome facts.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if !hasReason(set.Candidates, TierMandatory, "target") {
		t.Errorf("Candidates = %+v, want a TierMandatory/target entry even with no Constitution", set.Candidates)
	}
	if hasReason(set.Candidates, TierMandatory, "constitution") {
		t.Errorf("Candidates = %+v, want no constitution entry when none exists", set.Candidates)
	}
}

func TestCollect_NonexistentTargetIsRejected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	store := openSyncedStore(t, root, cfg)

	_, err := Collect(root, cfg, store, Request{Target: "KNOW-999"})
	if !errors.Is(err, operations.ErrEntityNotFound) {
		t.Fatalf("Collect() error = %v, want errors.Is(err, ErrEntityNotFound)", err)
	}
}

func TestCollect_OutOfScopeTargetIsRejected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/tasks.md",
		"---\ntype: tasks\nfor: SPEC-001\n---\n# Tasks\n\n## TASK-001 — Do it\n\n- [ ] Complete\n")
	store := openSyncedStore(t, root, cfg)

	_, err := Collect(root, cfg, store, Request{Target: "TASK-001"})
	if !errors.Is(err, operations.ErrInvalidTarget) {
		t.Fatalf("Collect() error = %v, want errors.Is(err, ErrInvalidTarget)", err)
	}
}

func TestCollect_StructuralAndSemanticConnectionsLabeledCorrectly(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThe dependency target.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nBackground facts.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n## Intent\n\nSee [[KNOW-001]] for background.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-002"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	if !hasReason(set.Candidates, TierStructural, "parent") {
		t.Errorf("Candidates = %+v, want a TierStructural/parent entry (FEAT-001)", set.Candidates)
	}
	if !hasReason(set.Candidates, TierStructural, "depends_on") {
		t.Errorf("Candidates = %+v, want a TierStructural/depends_on entry (SPEC-001)", set.Candidates)
	}
	if !hasReason(set.Candidates, TierSemantic, "wikilink") {
		t.Errorf("Candidates = %+v, want a TierSemantic/wikilink entry (KNOW-001)", set.Candidates)
	}
}

func TestCollect_IncomingReferenceLabeledAsBacklink(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThe backlink target.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-001.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if !hasReason(set.Candidates, TierSemantic, "backlink") {
		t.Errorf("Candidates = %+v, want a TierSemantic/backlink entry (SPEC-002)", set.Candidates)
	}
}

func TestCollect_NoStructuralConnectionsContributesNothingBeyondBaseline(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome facts.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if hasReason(set.Candidates, TierStructural, "parent") || hasReason(set.Candidates, TierSemantic, "wikilink") || hasReason(set.Candidates, TierSemantic, "backlink") {
		t.Errorf("Candidates = %+v, want no structural/semantic entries for an unconnected target", set.Candidates)
	}
}

func TestCollect_TextMatchLabeledCorrectly(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nThe target's own content.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nRefresh token rotation details.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001", Query: "rotation"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if !hasReason(set.Candidates, TierText, "text_match") {
		t.Errorf("Candidates = %+v, want a TierText/text_match entry (KNOW-002)", set.Candidates)
	}
}

func TestCollect_TaskUsedAsQueryWhenQueryEmpty(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nThe target's own content.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nRefresh token rotation details.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001", Task: "rotation"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if !hasReason(set.Candidates, TierText, "text_match") {
		t.Errorf("Candidates = %+v, want a TierText/text_match entry derived from Task", set.Candidates)
	}
}

func TestCollect_NoQueryNoTaskContributesNoTextTier(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome facts.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if hasReason(set.Candidates, TierText, "text_match") {
		t.Errorf("Candidates = %+v, want no text_match entries with no Query and no Task", set.Candidates)
	}
}

func TestCollect_QueryMatchingNothingIsNotAnError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome facts.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001", Query: "xyzzyunmatchable"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if hasReason(set.Candidates, TierText, "text_match") {
		t.Errorf("Candidates = %+v, want no text_match entries for a query matching nothing", set.Candidates)
	}
}

// TestCollect_TextMatchCandidateCarriesBM25TextRank covers
// 036-text-search-ranking spec FR-005: Tier 4's free-text branch must
// populate each resulting Candidate's TextRank from the SearchResult's
// own bm25() Rank, not leave it at the zero value.
func TestCollect_TextMatchCandidateCarriesBM25TextRank(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nThe target's own content.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nRefresh token rotation details.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001", Query: "rotation"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	found := false
	for _, c := range set.Candidates {
		for _, r := range c.Reasons {
			if r.Tier == TierText && r.Relation == "text_match" {
				found = true
				if c.TextRank == 0 {
					t.Errorf("Candidate %+v has TextRank == 0, want the real bm25() value from SearchResult.Rank", c)
				}
			}
		}
	}
	if !found {
		t.Fatalf("Candidates = %+v, want a TierText/text_match entry to check TextRank on", set.Candidates)
	}
}

// TestCollect_DeduplicatesAcrossStructuralAndTextSignals directly
// proves spec.md's own User Story 4, Acceptance Scenario 1: the same
// chunk discovered as both a structural connection and a text match
// appears exactly once, carrying both reasons.
func TestCollect_DeduplicatesAcrossStructuralAndTextSignals(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nRefresh token rotation details.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-001.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-002", Query: "rotation"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var matches int
	var reasons []Reason
	for _, c := range set.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md" {
			matches++
			reasons = c.Reasons
		}
	}
	if matches != 1 {
		t.Fatalf("SPEC-001's own chunk appears %d times, want exactly 1: %+v", matches, set.Candidates)
	}
	hasStructural, hasText := false, false
	for _, r := range reasons {
		if r.Tier == TierStructural && r.Relation == "depends_on" {
			hasStructural = true
		}
		if r.Tier == TierText && r.Relation == "text_match" {
			hasText = true
		}
	}
	if !hasStructural || !hasText {
		t.Errorf("reasons = %+v, want both a TierStructural/depends_on and a TierText/text_match reason", reasons)
	}
}

// TestCollect_DeduplicatesAcrossTwoStructuralRelationships directly
// proves spec.md's own User Story 4, Acceptance Scenario 2: the same
// source artifact found via two distinct structural relationships (a
// direct dependency, and its own semantic backlink to the target)
// appears exactly once.
func TestCollect_DeduplicatesAcrossTwoStructuralRelationships(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-002.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nSee [[SPEC-001]] for context.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var matches int
	var reasons []Reason
	for _, c := range set.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md" {
			matches++
			reasons = c.Reasons
		}
	}
	if matches != 1 {
		t.Fatalf("SPEC-002's own chunk appears %d times, want exactly 1: %+v", matches, set.Candidates)
	}
	hasDependsOn, hasBacklink := false, false
	for _, r := range reasons {
		if r.Tier == TierStructural && r.Relation == "depends_on" {
			hasDependsOn = true
		}
		if r.Tier == TierSemantic && r.Relation == "backlink" {
			hasBacklink = true
		}
	}
	if !hasDependsOn || !hasBacklink {
		t.Errorf("reasons = %+v, want both a TierStructural/depends_on and a TierSemantic/backlink reason", reasons)
	}
}

func TestCollect_SecondHopSurfacesFurtherDependency(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-002.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-003\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-003.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md",
		"---\nid: SPEC-003\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-004\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-004 (a third hop away from SPEC-001).\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-004/spec.md",
		"---\nid: SPEC-004\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThird hop — must never appear.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	if !hasReason(set.Candidates, TierSecondHop, "depends_on") {
		t.Errorf("Candidates = %+v, want a TierSecondHop/depends_on entry (SPEC-003)", set.Candidates)
	}
	for _, c := range set.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-004/spec.md" {
			t.Errorf("Candidates unexpectedly includes SPEC-004, a third hop from the target: %+v", set.Candidates)
		}
	}
}

func TestCollect_SecondHopDuplicateOfDirectStaysClassifiedAsDirect(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	// SPEC-001 depends on both SPEC-002 (first hop) and SPEC-003
	// (directly) — SPEC-002 also depends on SPEC-003, so SPEC-003 is
	// reachable both directly and via second-hop expansion.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\n  - SPEC-003\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-002 and SPEC-003.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-003\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-003.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md",
		"---\nid: SPEC-003\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nBoth direct and second-hop.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var matches int
	var reasons []Reason
	for _, c := range set.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-003/spec.md" {
			matches++
			reasons = c.Reasons
		}
	}
	if matches != 1 {
		t.Fatalf("SPEC-003's own chunk appears %d times, want exactly 1: %+v", matches, set.Candidates)
	}
	for _, r := range reasons {
		if r.Tier == TierSecondHop {
			t.Errorf("reasons = %+v, want no TierSecondHop reason once a direct one exists for the same chunk", reasons)
		}
	}
}

func TestCollect_UnrecognizedIntentIsRejected(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSome facts.\n")
	store := openSyncedStore(t, root, cfg)

	_, err := Collect(root, cfg, store, Request{Target: "KNOW-001", Intent: Intent("not-a-real-intent")})
	if err == nil {
		t.Fatal("Collect() expected an error for an unrecognized intent, got nil")
	}
}
