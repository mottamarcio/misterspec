package contextengine

import "testing"

func TestTierString_EachRecognizedTierHasItsOwnLabel(t *testing.T) {
	cases := []struct {
		tier Tier
		want string
	}{
		{TierMandatory, "mandatory"},
		{TierStructural, "structural"},
		{TierSemantic, "semantic"},
		{TierText, "text"},
		{TierSecondHop, "second_hop"},
	}
	for _, c := range cases {
		if got := c.tier.String(); got != c.want {
			t.Errorf("Tier(%d).String() = %q, want %q", c.tier, got, c.want)
		}
	}
}

func TestTierString_OutOfRangeValueIsUnknown(t *testing.T) {
	if got := Tier(999).String(); got != "unknown" {
		t.Errorf("Tier(999).String() = %q, want %q", got, "unknown")
	}
}

func TestMergeAndSort_MergesReasonsForSameChunk(t *testing.T) {
	candidates := []Candidate{
		{Path: "a.md", StartLine: 1, EndLine: 2, Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}}},
		{Path: "a.md", StartLine: 1, EndLine: 2, Reasons: []Reason{{Tier: TierText, Relation: "text_match"}}},
	}

	set := mergeAndSort(candidates)

	if len(set.Candidates) != 1 {
		t.Fatalf("len(Candidates) = %d, want 1: %+v", len(set.Candidates), set.Candidates)
	}
	if len(set.Candidates[0].Reasons) != 2 {
		t.Fatalf("Reasons = %+v, want 2 distinct reasons", set.Candidates[0].Reasons)
	}
}

func TestMergeAndSort_DirectAlwaysWinsOverSecondHopForSameChunk(t *testing.T) {
	candidates := []Candidate{
		{Path: "a.md", StartLine: 1, EndLine: 2, Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}}},
		{Path: "a.md", StartLine: 1, EndLine: 2, Reasons: []Reason{{Tier: TierSecondHop, Relation: "depends_on"}}},
	}

	set := mergeAndSort(candidates)

	if len(set.Candidates) != 1 {
		t.Fatalf("len(Candidates) = %d, want 1: %+v", len(set.Candidates), set.Candidates)
	}
	for _, r := range set.Candidates[0].Reasons {
		if r.Tier == TierSecondHop {
			t.Errorf("Reasons = %+v, want no TierSecondHop reason once a direct reason exists for the same chunk", set.Candidates[0].Reasons)
		}
	}
	if len(set.Candidates[0].Reasons) != 1 {
		t.Errorf("Reasons = %+v, want exactly 1 (the direct one, second-hop dropped)", set.Candidates[0].Reasons)
	}
}

func TestMergeAndSort_SecondHopOnlyIsKeptWhenNoDirectReasonExists(t *testing.T) {
	candidates := []Candidate{
		{Path: "a.md", StartLine: 1, EndLine: 2, Reasons: []Reason{{Tier: TierSecondHop, Relation: "depends_on"}}},
	}

	set := mergeAndSort(candidates)

	if len(set.Candidates) != 1 || len(set.Candidates[0].Reasons) != 1 || set.Candidates[0].Reasons[0].Tier != TierSecondHop {
		t.Errorf("set = %+v, want the lone second-hop reason preserved", set)
	}
}

func TestMergeAndSort_OrdersByTierThenPathThenStartLine(t *testing.T) {
	candidates := []Candidate{
		{Path: "z.md", StartLine: 1, EndLine: 1, Reasons: []Reason{{Tier: TierText, Relation: "text_match"}}},
		{Path: "a.md", StartLine: 5, EndLine: 5, Reasons: []Reason{{Tier: TierMandatory, Relation: "target"}}},
		{Path: "a.md", StartLine: 1, EndLine: 1, Reasons: []Reason{{Tier: TierMandatory, Relation: "constitution"}}},
		{Path: "b.md", StartLine: 1, EndLine: 1, Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}}},
	}

	set := mergeAndSort(candidates)

	wantOrder := []struct {
		path      string
		startLine int
	}{
		{"a.md", 1},
		{"a.md", 5},
		{"b.md", 1},
		{"z.md", 1},
	}
	if len(set.Candidates) != len(wantOrder) {
		t.Fatalf("len(Candidates) = %d, want %d: %+v", len(set.Candidates), len(wantOrder), set.Candidates)
	}
	for i, w := range wantOrder {
		if set.Candidates[i].Path != w.path || set.Candidates[i].StartLine != w.startLine {
			t.Errorf("Candidates[%d] = %+v, want Path=%s StartLine=%d", i, set.Candidates[i], w.path, w.startLine)
		}
	}
}

func TestMergeAndSort_EmptyInputProducesEmptySet(t *testing.T) {
	set := mergeAndSort(nil)
	if len(set.Candidates) != 0 {
		t.Errorf("Candidates = %+v, want none", set.Candidates)
	}
}

// TestMergeAndSort_PropagatesTextRankFromLaterTextMatchDuplicate covers
// 036-text-search-ranking data-model.md's Candidate validation rule: a
// duplicate discovered later, not first, still contributes its own
// TextRank to the merged Candidate whenever it carries a text_match
// Reason — the merge must not silently drop the real BM25 value just
// because a non-text-match occurrence of the same chunk was seen first.
func TestMergeAndSort_PropagatesTextRankFromLaterTextMatchDuplicate(t *testing.T) {
	candidates := []Candidate{
		{Path: "a.md", StartLine: 1, EndLine: 2, Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}}},
		{Path: "a.md", StartLine: 1, EndLine: 2, TextRank: -7.5, Reasons: []Reason{{Tier: TierText, Relation: "text_match"}}},
	}

	set := mergeAndSort(candidates)

	if len(set.Candidates) != 1 {
		t.Fatalf("len(Candidates) = %d, want 1: %+v", len(set.Candidates), set.Candidates)
	}
	if set.Candidates[0].TextRank != -7.5 {
		t.Errorf("TextRank = %v, want -7.5 propagated from the later text_match duplicate", set.Candidates[0].TextRank)
	}
}
