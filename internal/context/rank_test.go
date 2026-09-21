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

// TestRank_TextMatchOrderingFollowsBM25NotTermCount covers
// 036-text-search-ranking spec FR-005: among same-tier text_match
// candidates, the one with the more relevant (more negative) bm25()
// TextRank must rank first — regardless of how many literal term
// occurrences its own Content happens to contain.
func TestRank_TextMatchOrderingFollowsBM25NotTermCount(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/knowledge/KNOW-001.md", StartLine: 1, EndLine: 3,
			Content:  "rotation rotation rotation rotation rotation", // many raw term hits
			TextRank: -1.0,                                           // but a weak bm25 match
			Reasons:  []Reason{{Tier: TierText, Relation: "text_match"}},
		},
		{
			Path: "ai/knowledge/KNOW-002.md", StartLine: 1, EndLine: 3,
			Content:  "rotation", // only one raw term hit
			TextRank: -20.0,      // yet a strong bm25 match
			Reasons:  []Reason{{Tier: TierText, Relation: "text_match"}},
		},
	}}

	ranked := Rank(cs, Request{Query: "rotation"})

	if len(ranked) != 2 {
		t.Fatalf("Rank returned %d candidates, want 2", len(ranked))
	}
	if ranked[0].Path != "ai/knowledge/KNOW-002.md" {
		t.Errorf("ranked[0].Path = %q, want KNOW-002 (stronger bm25 TextRank must win over raw term-occurrence count — spec FR-005)", ranked[0].Path)
	}
}

// TestRank_StrongTextMatchNeverOutranksHigherTier covers spec FR-006,
// FR-012: an extremely strong bm25 TextRank on a TierText candidate
// must never let it sort ahead of a TierStructural candidate.
func TestRank_StrongTextMatchNeverOutranksHigherTier(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/knowledge/KNOW-999.md", StartLine: 1, EndLine: 3,
			Content: "extremely relevant text match", TextRank: -1000000.0,
			Reasons: []Reason{{Tier: TierText, Relation: "text_match"}},
		},
		{
			Path: "ai/specs/SPEC-011.md", StartLine: 1, EndLine: 3,
			Content: "irrelevant structural text",
			Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}},
		},
	}}

	ranked := Rank(cs, Request{Query: "extremely relevant text match"})

	if len(ranked) != 2 {
		t.Fatalf("Rank returned %d candidates, want 2", len(ranked))
	}
	if minTier(ranked[0].Reasons) != TierStructural {
		t.Errorf("ranked[0] tier = %v, want TierStructural — an extreme TextRank must never cross the tier boundary", minTier(ranked[0].Reasons))
	}
}

// TestRank_EqualBM25ScoresStillDeterministicByPathThenLine covers spec
// FR-007/SC-003: when two same-tier text_match candidates carry the
// exact same TextRank, the existing Path-then-StartLine tie-break still
// applies.
func TestRank_EqualBM25ScoresStillDeterministicByPathThenLine(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{Path: "ai/knowledge/KNOW-002.md", StartLine: 1, EndLine: 3, Content: "match", TextRank: -5.0, Reasons: []Reason{{Tier: TierText, Relation: "text_match"}}},
		{Path: "ai/knowledge/KNOW-001.md", StartLine: 1, EndLine: 3, Content: "match", TextRank: -5.0, Reasons: []Reason{{Tier: TierText, Relation: "text_match"}}},
	}}

	for i := 0; i < 5; i++ {
		ranked := Rank(cs, Request{Query: "match"})
		if len(ranked) != 2 {
			t.Fatalf("Rank returned %d candidates, want 2", len(ranked))
		}
		if ranked[0].Path != "ai/knowledge/KNOW-001.md" {
			t.Errorf("run %d: ranked[0].Path = %q, want KNOW-001 (equal TextRank must still sort deterministically by Path)", i, ranked[0].Path)
		}
	}
}

func TestRank_PreferSectionFalseIsByteIdenticalToDefault(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/specs/SPEC-001.md", StartLine: 1, EndLine: 3,
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink", SourceSection: "Requirements"}},
		},
		{
			Path: "ai/specs/SPEC-002.md", StartLine: 1, EndLine: 3,
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink", SourceSection: "Notes"}},
		},
	}}

	withoutFlag := Rank(cs, Request{})
	explicitFalse := Rank(cs, Request{PreferSection: false})

	if withoutFlag[0].Score != explicitFalse[0].Score || withoutFlag[1].Score != explicitFalse[1].Score {
		t.Fatalf("PreferSection's zero value must be byte-identical to omitting it: %+v vs %+v", withoutFlag, explicitFalse)
	}
	for _, sc := range withoutFlag {
		if sc.Components.SectionPreference != 0 {
			t.Errorf("SectionPreference = %d, want 0 when PreferSection is false (spec FR-006)", sc.Components.SectionPreference)
		}
	}
}

func TestRank_PreferSectionTrueRanksRequirementsSectionHigher(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/specs/SPEC-002.md", StartLine: 1, EndLine: 3,
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink", SourceSection: "Notes"}},
		},
		{
			Path: "ai/specs/SPEC-001.md", StartLine: 1, EndLine: 3,
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink", SourceSection: "Requirements"}},
		},
	}}

	ranked := Rank(cs, Request{PreferSection: true})

	if ranked[0].Path != "ai/specs/SPEC-001.md" {
		t.Fatalf("ranked[0] = %q, want the Requirements-sourced candidate to rank first with PreferSection true: %+v", ranked[0].Path, ranked)
	}
	if ranked[0].Components.SectionPreference != sectionPreferenceBonus {
		t.Errorf("ranked[0].Components.SectionPreference = %d, want %d", ranked[0].Components.SectionPreference, sectionPreferenceBonus)
	}
	if ranked[1].Components.SectionPreference != 0 {
		t.Errorf("ranked[1].Components.SectionPreference = %d, want 0 (unrelated section)", ranked[1].Components.SectionPreference)
	}
}

func TestRank_PreferSectionMatchIsCaseInsensitive(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/specs/SPEC-001.md", StartLine: 1, EndLine: 3,
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink", SourceSection: "requirements"}},
		},
	}}

	ranked := Rank(cs, Request{PreferSection: true})

	if ranked[0].Components.SectionPreference != sectionPreferenceBonus {
		t.Errorf("SectionPreference = %d, want %d (case-insensitive match)", ranked[0].Components.SectionPreference, sectionPreferenceBonus)
	}
}

func TestRank_PreferSectionMatchesFunctionalRequirementsHeading(t *testing.T) {
	cs := CandidateSet{Candidates: []Candidate{
		{
			Path: "ai/specs/SPEC-001.md", StartLine: 1, EndLine: 3,
			Reasons: []Reason{{Tier: TierSemantic, Relation: "wikilink", SourceSection: "Functional Requirements"}},
		},
	}}

	ranked := Rank(cs, Request{PreferSection: true})

	if ranked[0].Components.SectionPreference != sectionPreferenceBonus {
		t.Errorf("SectionPreference = %d, want %d", ranked[0].Components.SectionPreference, sectionPreferenceBonus)
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
