package contextengine

import (
	"strconv"
	"testing"
)

func TestRank_MandatoryOutranksStructuralOutranksTextRegardlessOfScore(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/knowledge/KNOW-999.md", StartLine: 1, EndLine: 5,
			Content: "refresh token rotation refresh token rotation refresh token rotation",
			Reasons: []Reason{{Tier: TierText, Relation: "text_match"}},
		},
		{
			Path: "ai/memory/constitution.md", StartLine: 1, EndLine: 5,
			Content: "irrelevant constitution text",
			Reasons: []Reason{{Tier: TierMandatory, Relation: "constitution"}},
		},
		{
			Path: "ai/specs/SPEC-011.md", StartLine: 1, EndLine: 5,
			Content: "irrelevant structural text",
			Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}},
		},
	}}

	req := Request{Query: "refresh token rotation"}
	ranked := Rank(cs, req)

	if len(ranked) != 3 {
		t.Fatalf("Rank returned %d candidates, want 3", len(ranked))
	}
	if minTier(ranked[0].Reasons) != TierMandatory {
		t.Errorf("ranked[0] tier = %v, want TierMandatory (a strong text match must never outrank it)", minTier(ranked[0].Reasons))
	}
	if minTier(ranked[1].Reasons) != TierStructural {
		t.Errorf("ranked[1] tier = %v, want TierStructural (a strong text match must never outrank it)", minTier(ranked[1].Reasons))
	}
	if minTier(ranked[2].Reasons) != TierText {
		t.Errorf("ranked[2] tier = %v, want TierText", minTier(ranked[2].Reasons))
	}
}

func TestRank_WithinTierStrongerRelationRanksFirst(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/knowledge/KNOW-002.md", StartLine: 1, EndLine: 3,
			Content: "backlink source content",
			Reasons: []Reason{{Tier: TierSemantic, Relation: "backlink"}},
		},
		{
			Path: "ai/knowledge/KNOW-001.md", StartLine: 1, EndLine: 3,
			Content: "wikilink target content",
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink"}},
		},
	}}

	ranked := Rank(cs, Request{})

	if len(ranked) != 2 {
		t.Fatalf("Rank returned %d candidates, want 2", len(ranked))
	}
	if ranked[0].Path != "ai/knowledge/KNOW-001.md" {
		t.Errorf("ranked[0].Path = %q, want the wikilink candidate to rank first (wikilink outweighs backlink within the same Tier)", ranked[0].Path)
	}
}

func TestRank_IdenticalRelevanceIsStillDeterministicByPathThenLine(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{Path: "ai/specs/SPEC-002.md", StartLine: 1, EndLine: 3, Content: "same", Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink"}}},
		{Path: "ai/specs/SPEC-001.md", StartLine: 5, EndLine: 8, Content: "same", Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink"}}},
		{Path: "ai/specs/SPEC-001.md", StartLine: 1, EndLine: 3, Content: "same", Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink"}}},
	}}

	want := []string{"ai/specs/SPEC-001.md:1", "ai/specs/SPEC-001.md:5", "ai/specs/SPEC-002.md:1"}

	for i := 0; i < 5; i++ {
		ranked := Rank(cs, Request{})
		if len(ranked) != 3 {
			t.Fatalf("Rank returned %d candidates, want 3", len(ranked))
		}
		for idx, c := range ranked {
			got := c.Path + ":" + strconv.Itoa(c.StartLine)
			if got != want[idx] {
				t.Errorf("run %d: ranked[%d] = %q, want %q (identical relevance must still sort deterministically)", i, idx, got, want[idx])
			}
		}
	}
}

func TestRank_IntentOnlyReordersWithinATier(t *testing.T) {
	// IntentImplementation prefers "wikilink"/"backlink"/"text_match"
	// (research.md #4) — a TierSecondHop candidate carrying "wikilink"
	// gets an intent bonus, but must never outrank a TierStructural
	// candidate that has no such bonus (FR-003).
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/features/FEAT-005.md", StartLine: 1, EndLine: 3,
			Content: "second hop content",
			Reasons: []Reason{{Tier: TierSecondHop, Relation: "wikilink"}},
		},
		{
			Path: "ai/specs/SPEC-003.md", StartLine: 1, EndLine: 3,
			Content: "structural content",
			Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}},
		},
	}}

	ranked := Rank(cs, Request{Intent: IntentImplementation})

	if len(ranked) != 2 {
		t.Fatalf("Rank returned %d candidates, want 2", len(ranked))
	}
	if minTier(ranked[0].Reasons) != TierStructural {
		t.Errorf("ranked[0] tier = %v, want TierStructural — Intent must never let a lower Tier outrank a higher one", minTier(ranked[0].Reasons))
	}
}
