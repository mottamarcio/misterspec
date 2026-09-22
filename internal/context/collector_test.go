package contextengine

import (
	"errors"
	"fmt"
	"testing"
	"time"

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

func candidateReasons(set CandidateSet, path string) []Reason {
	for _, c := range set.Candidates {
		if c.Path == path {
			return c.Reasons
		}
	}
	return nil
}

func TestCollect_WikilinkReasonCarriesSourceOccurrence(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Related Specs\n\nSee [[KNOW-002]].\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nUnrelated facts.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	reasons := candidateReasons(set, "ai/knowledge/KNOW-002-y.md")
	if len(reasons) == 0 {
		t.Fatalf("no candidate found for KNOW-002; set = %+v", set.Candidates)
	}
	var found bool
	for _, r := range reasons {
		if r.Relation != "wikilink" {
			continue
		}
		found = true
		if r.SourcePath != "ai/knowledge/KNOW-001-x.md" {
			t.Errorf("SourcePath = %q, want %q", r.SourcePath, "ai/knowledge/KNOW-001-x.md")
		}
		if r.SourceSection != "Related Specs" {
			t.Errorf("SourceSection = %q, want %q", r.SourceSection, "Related Specs")
		}
		if r.SourceLine <= 0 {
			t.Errorf("SourceLine = %d, want a positive file-absolute line", r.SourceLine)
		}
	}
	if !found {
		t.Fatalf("reasons = %+v, want a wikilink reason", reasons)
	}
}

func TestCollect_MandatoryReasonsCarryNoOccurrenceData(t *testing.T) {
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

	for _, c := range set.Candidates {
		for _, r := range c.Reasons {
			if r.Relation != "target" && r.Relation != "constitution" {
				continue
			}
			if r.SourcePath != "" || r.SourceSection != "" || r.SourceLine != 0 {
				t.Errorf("candidate %+v reason %+v: want SourcePath/SourceSection/SourceLine all empty/zero for %q", c, r, r.Relation)
			}
		}
	}
}

func TestCollect_FormalDependsOnReasonHasSourcePathButNoSection(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md", "---\nid: PRG-001\ntype: program\nstatus: active\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md", "---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThe dependency target.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-001\nsupersedes: []\n---\n## Intent\n\nDepends on SPEC-001.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-002"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	reasons := candidateReasons(set, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md")
	var found bool
	for _, r := range reasons {
		if r.Relation != "depends_on" {
			continue
		}
		found = true
		if r.SourcePath != "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md" {
			t.Errorf("SourcePath = %q, want SPEC-002's own path", r.SourcePath)
		}
		if r.SourceSection != "" || r.SourceLine != 0 {
			t.Errorf("SourceSection/SourceLine = %q/%d, want empty/zero for a formal relation (spec FR-009)", r.SourceSection, r.SourceLine)
		}
	}
	if !found {
		t.Fatalf("reasons = %+v, want a depends_on reason", reasons)
	}
}

// --- User Story 3: retrieval stays bounded and non-redundant ---

func TestCollect_MutualWikilinkCycleCompletesWithoutUnboundedExpansion(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	// KNOW-001 and KNOW-002 wikilink each other — a two-artifact cycle.
	// This feature's own occurrence-threading (T009/T010) must not
	// regress the pre-existing 2-hop cap that already makes this safe.
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-a.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-002]].\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-b.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-001]].\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error for a mutual wikilink cycle: %v", err)
	}

	counts := map[string]int{}
	for _, c := range set.Candidates {
		counts[c.Path]++
	}
	for path, n := range counts {
		if n != 1 {
			t.Errorf("candidate %q appears %d times, want exactly 1 (no unbounded/duplicated expansion from the cycle)", path, n)
		}
	}
	if len(counts) == 0 || len(counts) > 4 {
		t.Errorf("candidate count = %d (%v), want a small, bounded set for a two-artifact cycle", len(counts), counts)
	}
}

func TestCollect_HeavilyReferencedHubStaysBounded(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	// 25 artifacts all wikilink the same hub — Collect for the hub must
	// stay within the existing bounded shape (mandatory + one
	// TierSemantic backlink Candidate per referencing artifact), never
	// growing unboundedly via second-hop expansion (research.md #7:
	// second hop only follows outgoing references, never a backlink's
	// own further backlinks).
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-hub.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nThe hub.\n")
	const n = 25
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("KNOW-%03d", i+100)
		testutil.WriteFile(t, root, fmt.Sprintf("ai/knowledge/%s-x.md", id),
			fmt.Sprintf("---\nid: %s\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-001]].\n", id))
	}
	store := openSyncedStore(t, root, cfg)

	start := time.Now()
	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Collect() unexpected error for a heavily-referenced hub: %v", err)
	}
	if elapsed > 5*time.Second {
		t.Errorf("Collect() took %s for a %d-artifact hub, want it to complete promptly", elapsed, n)
	}

	counts := map[string]int{}
	for _, c := range set.Candidates {
		counts[c.Path]++
	}
	for path, count := range counts {
		if count != 1 {
			t.Errorf("candidate %q appears %d times, want exactly 1", path, count)
		}
	}
	// The hub itself (mandatory) plus exactly one candidate per
	// referencing artifact — no runaway multiplication.
	if len(counts) != n+1 {
		t.Errorf("candidate count = %d, want exactly %d (hub + %d referencing artifacts, each once)", len(counts), n+1, n)
	}
}

func TestCollect_TwoReferencePathsToSameChunkDeduplicateWithOccurrenceDataIntact(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/programs/PRG-001/program.md",
		"---\nid: PRG-001\ntype: program\nstatus: active\n---\n## Overview\n\nThe program.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/feature.md",
		"---\nid: FEAT-001\ntype: feature\nstatus: active\nparent: PRG-001\n---\n## Overview\n\nThe feature.\n")
	// SPEC-001 both depends_on SPEC-002 directly AND wikilinks it from
	// its own Requirements section — two distinct reference paths to
	// the identical chunk.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on:\n  - SPEC-002\nsupersedes: []\n---\n## Requirements\n\nSee [[SPEC-002]] too.\n")
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md",
		"---\nid: SPEC-002\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nThe shared target.\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "SPEC-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	reasons := candidateReasons(set, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md")
	var count int
	for _, c := range set.Candidates {
		if c.Path == "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-002/spec.md" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("SPEC-002's own chunk appears %d times, want exactly 1 (spec FR-007)", count)
	}

	var hasDependsOn, hasWikilinkWithOccurrence bool
	for _, r := range reasons {
		if r.Relation == "depends_on" {
			hasDependsOn = true
		}
		if r.Relation == "wikilink" && r.SourceSection == "Requirements" {
			hasWikilinkWithOccurrence = true
		}
	}
	if !hasDependsOn || !hasWikilinkWithOccurrence {
		t.Errorf("reasons = %+v, want both a depends_on reason and a wikilink reason with SourceSection=Requirements — dedup must not drop either reason's own occurrence data", reasons)
	}
}

// --- Spec 040: stable section anchors ---

// TestCollect_AnchorQualifiedReferenceReturnsOnlyThatSection proves spec
// 040 FR-006/data-model.md "Candidate / Reason (extended)": a wikilink
// referencing a specific anchor produces exactly one Candidate for that
// Section — not one Candidate per Chunk of the whole target artifact,
// which multiple, unrelated sections here would otherwise contribute.
func TestCollect_AnchorQualifiedReferenceReturnsOnlyThatSection(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n"+
			"## Networking Policies\n\nGeneral networking notes.\n\n"+
			"### Client Retry Policy {#retry-policy}\n\nBackoff details.\n\n"+
			"## Unrelated Section\n\nSomething else entirely.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-003#retry-policy]].\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var matches int
	var found Candidate
	for _, c := range set.Candidates {
		if c.Path == "ai/knowledge/KNOW-003-x.md" {
			matches++
			found = c
		}
	}
	if matches != 1 {
		t.Fatalf("KNOW-003 contributed %d candidates, want exactly 1 (the anchored section alone): %+v", matches, set.Candidates)
	}
	if found.Heading != "Client Retry Policy" {
		t.Errorf("Heading = %q, want %q", found.Heading, "Client Retry Policy")
	}
	if len(found.HeadingPath) != 1 || found.HeadingPath[0] != "Networking Policies" {
		t.Errorf("HeadingPath = %+v, want [\"Networking Policies\"]", found.HeadingPath)
	}
}

// TestCollect_UnknownAnchorReferenceIsOmittedNotFabricated proves the
// code-review finding on chunkArtifactAnchor: operations.References
// populates ReferenceEntry.TargetAnchor straight from the wikilink's
// raw text with no existence check (internal/validation's own separate
// job, CodeUnknownAnchor) — so a reference to an anchor that doesn't
// actually exist on the target IS reachable in normal Collect use.
// Collect must silently omit that reference, never include a
// contentless placeholder Candidate for it (which would previously
// have appeared with a nonzero score and no content, polluting the
// Context Pack) and never fail the whole request over it.
func TestCollect_UnknownAnchorReferenceIsOmittedNotFabricated(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n"+
			"## Networking Policies\n\nGeneral networking notes.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-003#nonexistent]].\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	for _, c := range set.Candidates {
		if c.Path == "ai/knowledge/KNOW-003-x.md" {
			t.Fatalf("expected no candidate for KNOW-003 (its only reference names a nonexistent anchor), got %+v", c)
		}
	}
}

// TestCollect_NonAnchorReferenceStillReturnsEveryChunk proves spec 040
// FR-008: a plain (non-anchor) wikilink to a multi-section target is
// completely unaffected — every Chunk of the target still becomes its
// own Candidate, exactly as before this feature, with no HeadingPath.
func TestCollect_NonAnchorReferenceStillReturnsEveryChunk(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n"+
			"## Networking Policies\n\nGeneral networking notes.\n\n"+
			"### Client Retry Policy {#retry-policy}\n\nBackoff details.\n\n"+
			"## Unrelated Section\n\nSomething else entirely.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-003]].\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var matches int
	for _, c := range set.Candidates {
		if c.Path == "ai/knowledge/KNOW-003-x.md" {
			matches++
			if len(c.HeadingPath) != 0 {
				t.Errorf("candidate %+v: HeadingPath = %+v, want empty for a non-anchor reference", c, c.HeadingPath)
			}
		}
	}
	if matches != 3 {
		t.Fatalf("KNOW-003 contributed %d candidates, want exactly 3 (every one of its own Chunks, unaffected by this feature): %+v", matches, set.Candidates)
	}
}

// TestCollect_AnchorOnEmptyBodySectionStillResolves proves spec 040's
// Edge Cases / data-model.md "Chunk (extended)" note: an anchor
// declared on a heading immediately followed by a subheading (empty
// own Body, so artifacts.Chunks() would produce no Chunk for it at
// all) still resolves successfully to exactly one Candidate with empty
// Content — never treated as not-found.
func TestCollect_AnchorOnEmptyBodySectionStillResolves(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n"+
			"## Networking Policies {#networking}\n"+
			"### Client Retry Policy {#retry-policy}\n\nBackoff details.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-003#networking]].\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var matches int
	var found Candidate
	for _, c := range set.Candidates {
		if c.Path == "ai/knowledge/KNOW-003-x.md" {
			matches++
			found = c
		}
	}
	if matches != 1 {
		t.Fatalf("KNOW-003 contributed %d candidates, want exactly 1 (the empty-body anchored section): %+v", matches, set.Candidates)
	}
	if found.Heading != "Networking Policies" {
		t.Errorf("Heading = %q, want %q", found.Heading, "Networking Policies")
	}
	if found.Content != "" {
		t.Errorf("Content = %q, want empty for an anchor on a Section whose own Body is empty", found.Content)
	}
}

// TestCollect_AnchorSurvivesHeadingTitleRename proves spec 040 User
// Story 2, Acceptance Scenario 1: renaming a Section's own heading
// title text, while keeping its declared "{#anchor}" unchanged, leaves
// an existing anchor-qualified reference resolving correctly — no
// wikilink edit required.
func TestCollect_AnchorSurvivesHeadingTitleRename(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-x.md",
		"---\nid: KNOW-003\ntype: knowledge\nstatus: active\n---\n"+
			"## Backoff and Retry Policy for Clients {#retry-policy}\n\nBackoff details.\n")
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md",
		"---\nid: KNOW-001\ntype: knowledge\nstatus: active\n---\n## Summary\n\nSee [[KNOW-003#retry-policy]].\n")
	store := openSyncedStore(t, root, cfg)

	set, err := Collect(root, cfg, store, Request{Target: "KNOW-001"})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	var found bool
	for _, c := range set.Candidates {
		if c.Path == "ai/knowledge/KNOW-003-x.md" {
			found = true
			if c.Heading != "Backoff and Retry Policy for Clients" {
				t.Errorf("Heading = %q, want the renamed title", c.Heading)
			}
			if c.Content != "\nBackoff details." {
				t.Errorf("Content = %q, want %q", c.Content, "\nBackoff details.")
			}
		}
	}
	if !found {
		t.Fatalf("no candidate found for KNOW-003 after renaming its title but keeping its anchor; set = %+v", set.Candidates)
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
