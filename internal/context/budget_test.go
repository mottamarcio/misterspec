package contextengine

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestResolveBudget_NilUsesDefault(t *testing.T) {
	got := resolveBudget(Request{})
	if got != DefaultBudget {
		t.Errorf("resolveBudget(Request{}) = %d, want DefaultBudget (%d)", got, DefaultBudget)
	}
}

func TestResolveBudget_ExplicitValueIsUsedVerbatim(t *testing.T) {
	cases := []int{0, -5, 1, 12345}
	for _, want := range cases {
		want := want
		req := Request{Budget: &want}
		if got := resolveBudget(req); got != want {
			t.Errorf("resolveBudget(Request{Budget: &%d}) = %d, want %d", want, got, want)
		}
	}
}

// sc builds a minimal ScoredCandidate for ApplyBudget tests, with a
// content string of exactly n runes so its own EstimateTokens cost is
// easy to reason about.
func sc(path string, tier Tier, relation string, n int) ScoredCandidate {
	return ScoredCandidate{Candidate: Candidate{
		Path:    path,
		Content: strings.Repeat("a", n),
		Reasons: []Reason{{Tier: tier, Relation: relation}},
	}}
}

func TestApplyBudget_TrimsLowerTierBeforeHigherTier(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 40) // 10 tokens
	structural := sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 40)      // 10 tokens
	semantic := sc("ai/knowledge/KNOW-001.md", TierSemantic, "wikilink", 40)        // 10 tokens
	ranked := []ScoredCandidate{mandatory, structural, semantic}

	budget := 25 // fits mandatory+structural (20) but not +semantic (30)
	result := ApplyBudget(ranked, Request{Budget: &budget})

	if len(result.Items) != 2 {
		t.Fatalf("len(result.Items) = %d, want 2 (semantic must be trimmed before touching mandatory/structural)", len(result.Items))
	}
	if result.Items[0].Path != mandatory.Path || result.Items[1].Path != structural.Path {
		t.Errorf("result.Items = %+v, want mandatory then structural, in that order", result.Items)
	}
	wantAvailable := artifacts.EstimateTokens(mandatory.Content) + artifacts.EstimateTokens(structural.Content) + artifacts.EstimateTokens(semantic.Content)
	wantSelected := artifacts.EstimateTokens(mandatory.Content) + artifacts.EstimateTokens(structural.Content)
	if result.Diagnostics.TokensAvailable != wantAvailable {
		t.Errorf("TokensAvailable = %d, want %d", result.Diagnostics.TokensAvailable, wantAvailable)
	}
	if result.Diagnostics.TokensSelected != wantSelected {
		t.Errorf("TokensSelected = %d, want %d", result.Diagnostics.TokensSelected, wantSelected)
	}
	if result.Diagnostics.TokensExcluded != wantAvailable-wantSelected {
		t.Errorf("TokensExcluded = %d, want %d", result.Diagnostics.TokensExcluded, wantAvailable-wantSelected)
	}
}

func TestApplyBudget_NothingRemovedWhenEverythingFits(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 20),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 20),
	}
	budget := 1000
	result := ApplyBudget(ranked, Request{Budget: &budget})

	if len(result.Items) != len(ranked) {
		t.Fatalf("len(result.Items) = %d, want %d (nothing should be removed)", len(result.Items), len(ranked))
	}
	if result.Diagnostics.TokensExcluded != 0 {
		t.Errorf("TokensExcluded = %d, want 0", result.Diagnostics.TokensExcluded)
	}
	if result.Diagnostics.ReductionPercent != 0 {
		t.Errorf("ReductionPercent = %v, want 0", result.Diagnostics.ReductionPercent)
	}
}

func TestApplyBudget_NoBudgetSpecifiedUsesDefault(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 40),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 40),
	}
	result := ApplyBudget(ranked, Request{}) // no Budget set

	if len(result.Items) != len(ranked) {
		t.Fatalf("len(result.Items) = %d, want %d (DefaultBudget is ample for this tiny set)", len(result.Items), len(ranked))
	}
}

func TestApplyBudget_PreservesRankedOrderExactly(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 10),
		sc("ai/specs/SPEC-002.md", TierStructural, "depends_on", 10),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 10),
	}
	budget := 1000
	result := ApplyBudget(ranked, Request{Budget: &budget})

	for i, item := range result.Items {
		if item.Path != ranked[i].Path {
			t.Errorf("result.Items[%d].Path = %q, want %q (ApplyBudget must preserve ranked's own order)", i, item.Path, ranked[i].Path)
		}
	}
}

// --- User Story 3: Never Silently Drop Mandatory Context ---
// Proving tests only (research.md, tasks.md) — ApplyBudget's own
// oversized-mandatory handling is already implemented as part of its
// one coherent algorithm (T007); nothing new to build here.

func TestApplyBudget_MandatoryAloneExceedingBudgetIsStillReturnedInFull(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 400) // 100 tokens
	ranked := []ScoredCandidate{mandatory}

	tiny := 10
	result := ApplyBudget(ranked, Request{Budget: &tiny})

	if len(result.Items) != 1 || result.Items[0].Path != mandatory.Path {
		t.Fatalf("result.Items = %+v, want the full mandatory candidate, never omitted or shortened", result.Items)
	}
	wantTokens := artifacts.EstimateTokens(mandatory.Content)
	if result.Items[0].Tokens != wantTokens {
		t.Errorf("result.Items[0].Tokens = %d, want %d (mandatory content must never be shortened to force a fit)", result.Items[0].Tokens, wantTokens)
	}
}

func TestApplyBudget_MandatoryOverageIsFlaggedAndReportedExactly(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 400) // 100 tokens
	ranked := []ScoredCandidate{mandatory}

	tiny := 10
	result := ApplyBudget(ranked, Request{Budget: &tiny})

	if !result.BudgetExceeded {
		t.Fatal("result.BudgetExceeded = false, want true")
	}
	wantOverage := artifacts.EstimateTokens(mandatory.Content) - tiny
	if result.Overage != wantOverage {
		t.Errorf("result.Overage = %d, want %d", result.Overage, wantOverage)
	}
}

func TestApplyBudget_OptionalContentStaysOmittedWhenMandatoryAloneExceedsBudget(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 400) // 100 tokens
	optional := sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 4)          // 1 token
	ranked := []ScoredCandidate{mandatory, optional}

	tiny := 10
	result := ApplyBudget(ranked, Request{Budget: &tiny})

	for _, item := range result.Items {
		if item.Path == optional.Path {
			t.Errorf("result.Items contains the optional candidate %q; it must remain free to be omitted once mandatory alone already exceeds budget", optional.Path)
		}
	}
}

func TestApplyBudget_ZeroOrNegativeBudgetIsExplicitNotDefault(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 40) // 10 tokens

	for _, budget := range []int{0, -1} {
		budget := budget
		result := ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &budget})

		if len(result.Items) != 1 {
			t.Fatalf("budget %d: len(result.Items) = %d, want 1 (mandatory always returned in full)", budget, len(result.Items))
		}
		if !result.BudgetExceeded {
			t.Errorf("budget %d: result.BudgetExceeded = false, want true — a zero/negative budget must not be silently promoted to DefaultBudget", budget)
		}
	}
}

// --- User Story 4: Explain Every Decision ---
// Proving tests only (research.md, tasks.md) — Diagnostics is already
// computed as part of ApplyBudget's own one coherent algorithm (T007);
// nothing new to build here.

func TestApplyBudget_DiagnosticsCountsMatchTheActualResult(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 40),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 40),
		sc("ai/knowledge/KNOW-001.md", TierSemantic, "wikilink", 40),
	}
	budget := 15 // fits only the mandatory item (10 tokens)
	result := ApplyBudget(ranked, Request{Budget: &budget})

	if result.Diagnostics.CandidatesConsidered != len(ranked) {
		t.Errorf("CandidatesConsidered = %d, want %d", result.Diagnostics.CandidatesConsidered, len(ranked))
	}
	if result.Diagnostics.ItemsSelected != len(result.Items) {
		t.Errorf("ItemsSelected = %d, want %d (len(result.Items))", result.Diagnostics.ItemsSelected, len(result.Items))
	}
}

func TestApplyBudget_DiagnosticsTokenArithmeticIsConsistent(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 40),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 40),
		sc("ai/knowledge/KNOW-001.md", TierSemantic, "wikilink", 40),
	}
	budget := 15
	result := ApplyBudget(ranked, Request{Budget: &budget})

	d := result.Diagnostics
	if d.TokensAvailable-d.TokensExcluded != d.TokensSelected {
		t.Errorf("TokensAvailable(%d) - TokensExcluded(%d) = %d, want TokensSelected(%d)",
			d.TokensAvailable, d.TokensExcluded, d.TokensAvailable-d.TokensExcluded, d.TokensSelected)
	}
	wantReduction := float64(d.TokensExcluded) / float64(d.TokensAvailable) * 100
	if d.ReductionPercent != wantReduction {
		t.Errorf("ReductionPercent = %v, want %v", d.ReductionPercent, wantReduction)
	}
}

func TestApplyBudget_ReductionPercentIsZeroWhenNothingIsAvailable(t *testing.T) {
	result := ApplyBudget(nil, Request{})
	if result.Diagnostics.TokensAvailable != 0 {
		t.Fatalf("TokensAvailable = %d, want 0", result.Diagnostics.TokensAvailable)
	}
	if result.Diagnostics.ReductionPercent != 0 {
		t.Errorf("ReductionPercent = %v, want 0 (no division panic on an empty candidate set)", result.Diagnostics.ReductionPercent)
	}
}

func TestApplyBudget_EveryKeptItemCarriesReasonsAndTokens(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 40),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 40),
	}
	budget := 1000
	result := ApplyBudget(ranked, Request{Budget: &budget})

	for _, item := range result.Items {
		if len(item.Reasons) == 0 {
			t.Errorf("item %q has no Reasons — every kept item must explain why it was included (FR-009)", item.Path)
		}
		if item.Tokens != artifacts.EstimateTokens(item.Content) {
			t.Errorf("item %q Tokens = %d, want %d", item.Path, item.Tokens, artifacts.EstimateTokens(item.Content))
		}
	}
}

func TestApplyBudget_MandatoryOnlyWellWithinBudgetReportsZeroExclusion(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 40),
		sc("ai/specs/SPEC-014.md", TierMandatory, "target", 40),
	}
	budget := 1000
	result := ApplyBudget(ranked, Request{Budget: &budget})

	if result.Diagnostics.TokensExcluded != 0 {
		t.Errorf("TokensExcluded = %d, want 0", result.Diagnostics.TokensExcluded)
	}
	if result.Diagnostics.ReductionPercent != 0 {
		t.Errorf("ReductionPercent = %v, want 0", result.Diagnostics.ReductionPercent)
	}
}
