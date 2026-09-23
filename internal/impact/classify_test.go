package impact

import "testing"

// TestClassifyRelation_FixedTable is 042-impact-analysis-review T011
// (Foundational): every RelationKind maps to exactly one Classification
// per the fixed table (research.md #7) — wikilink always
// SuggestedReview, unconditionally; every other relation
// DeterministicInvalidation.
func TestClassifyRelation_FixedTable(t *testing.T) {
	cases := []struct {
		relation RelationKind
		want     Classification
	}{
		{RelationDependsOn, DeterministicInvalidation},
		{RelationParent, DeterministicInvalidation},
		{RelationSupersedes, DeterministicInvalidation},
		{RelationCoverage, DeterministicInvalidation},
		{RelationEvidence, DeterministicInvalidation},
		{RelationWikilink, SuggestedReview},
	}
	for _, tc := range cases {
		if got := ClassifyRelation(tc.relation); got != tc.want {
			t.Errorf("ClassifyRelation(%q) = %q, want %q", tc.relation, got, tc.want)
		}
	}
}

// TestClassifyRelation_WikilinkNeverInvalidation guards spec FR-003/
// FR-004 directly: no repeated call, and no other relation value ever
// mutates RelationWikilink's own mapping.
func TestClassifyRelation_WikilinkNeverInvalidation(t *testing.T) {
	for i := 0; i < 5; i++ {
		if got := ClassifyRelation(RelationWikilink); got != SuggestedReview {
			t.Fatalf("ClassifyRelation(wikilink) call #%d = %q, want %q", i, got, SuggestedReview)
		}
	}
}

// TestSeverityFor_Deterministic is 042-impact-analysis-review T011:
// calling SeverityFor twice with the same input always returns the
// same result (spec FR-010).
func TestSeverityFor_Deterministic(t *testing.T) {
	cases := []struct {
		classification Classification
		relation       RelationKind
	}{
		{DeterministicInvalidation, RelationEvidence},
		{DeterministicInvalidation, RelationDependsOn},
		{DeterministicInvalidation, RelationCoverage},
		{SuggestedReview, RelationWikilink},
	}
	for _, tc := range cases {
		first := SeverityFor(tc.classification, tc.relation)
		second := SeverityFor(tc.classification, tc.relation)
		if first != second {
			t.Errorf("SeverityFor(%q, %q) = %q then %q, want identical results", tc.classification, tc.relation, first, second)
		}
		if first == "" {
			t.Errorf("SeverityFor(%q, %q) = \"\", want a non-empty Severity", tc.classification, tc.relation)
		}
	}
}
