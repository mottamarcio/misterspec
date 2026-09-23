package contextengine

import (
	"reflect"
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

func TestDefaultHardLimit_IsTwiceDefaultBudget(t *testing.T) {
	if DefaultHardLimit != 2*DefaultBudget {
		t.Errorf("DefaultHardLimit = %d, want %d (2 * DefaultBudget)", DefaultHardLimit, 2*DefaultBudget)
	}
}

func TestApplyBudget_MandatoryOverSoftBudgetButUnderHardLimitIsNotFlagged(t *testing.T) {
	// 300 runes -> 75 tokens: over the tiny soft budget, well under the hard limit.
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 300)
	softBudget := 10
	hardLimit := 1000

	result := ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &softBudget, HardLimit: &hardLimit}, artifacts.DefaultEstimator{})

	if result.BudgetExceeded {
		t.Errorf("BudgetExceeded = true, want false (mandatory content is under the hard limit, even though over the soft budget — Acceptance Scenario 1)")
	}
	if len(result.Items) != 1 || result.Items[0].Path != mandatory.Path {
		t.Fatalf("result.Items = %+v, want the full mandatory candidate", result.Items)
	}
}

func TestApplyBudget_MandatoryOverHardLimitIsFlaggedAndStillDeliveredWhole(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 400) // 100 tokens
	softBudget := 10000
	hardLimit := 50

	result := ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &softBudget, HardLimit: &hardLimit}, artifacts.DefaultEstimator{})

	if !result.BudgetExceeded {
		t.Fatal("BudgetExceeded = false, want true (mandatory content exceeds the hard limit — Acceptance Scenario 2)")
	}
	wantOverage := artifacts.EstimateTokens(mandatory.Content) - hardLimit
	if result.Overage != wantOverage {
		t.Errorf("Overage = %d, want %d", result.Overage, wantOverage)
	}
	if len(result.Items) != 1 || result.Items[0].Path != mandatory.Path {
		t.Fatalf("result.Items = %+v, want the full mandatory candidate, never truncated (FR-007)", result.Items)
	}
}

func TestApplyBudget_HardLimitSmallerThanSoftBudgetIsAcceptedNotRejected(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 400) // 100 tokens
	softBudget := 10000
	hardLimit := 50 // smaller than softBudget — a probably-misconfigured but well-defined combination

	result := ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &softBudget, HardLimit: &hardLimit}, artifacts.DefaultEstimator{})

	if !result.BudgetExceeded {
		t.Error("BudgetExceeded = false, want true — a HardLimit smaller than Budget is a real, tight ceiling, not rejected or ignored (Edge Cases)")
	}
}

func TestApplyBudget_BackfillsSmallerLaterCandidateOfTheSameTierAfterAMiss(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 0)
	big := sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 80)   // 20 tokens — will not fit
	small := sc("ai/specs/SPEC-002.md", TierStructural, "depends_on", 8) // 2 tokens — fits after big is skipped
	ranked := []ScoredCandidate{mandatory, big, small}

	budget := 5
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

	foundSmall := false
	for _, item := range result.Items {
		if item.Path == small.Path {
			foundSmall = true
		}
		if item.Path == big.Path {
			t.Errorf("result.Items contains %q, which should not fit the remaining budget", big.Path)
		}
	}
	if !foundSmall {
		t.Errorf("result.Items = %+v, want the smaller later same-tier candidate %q included after the bigger one was skipped (FR-008)", result.Items, small.Path)
	}

	foundExclusion := false
	for _, e := range result.Diagnostics.Exclusions {
		if e.Path == big.Path {
			foundExclusion = true
			if e.Reason != ExclusionDidNotFitRemainingBudget {
				t.Errorf("Exclusions entry for %q has Reason %q, want %q", big.Path, e.Reason, ExclusionDidNotFitRemainingBudget)
			}
		}
	}
	if !foundExclusion {
		t.Errorf("Diagnostics.Exclusions = %+v, want an entry for the skipped candidate %q (FR-010)", result.Diagnostics.Exclusions, big.Path)
	}
}

func TestApplyBudget_BackfillPreservesRankedOrderAcrossTiers(t *testing.T) {
	// research.md Decision 4: removing the early-exit flag alone
	// guarantees FR-009 for free, because ranked's own order already
	// visits every higher-tier candidate before any lower-tier one — a
	// lower-tier candidate is only ever reached after every
	// higher-tier candidate has already been decided (included or
	// excluded) in its own turn. This test locks that Items always
	// mirrors ranked's own relative order, with excluded entries
	// simply missing — never reordered.
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 0)
	bigStructural := sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 80) // 20 tokens, doesn't fit
	smallSemantic := sc("ai/knowledge/KNOW-001.md", TierSemantic, "wikilink", 8)  // 2 tokens, fits after bigStructural is decided
	ranked := []ScoredCandidate{mandatory, bigStructural, smallSemantic}

	budget := 5
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

	var gotOrder []string
	for _, item := range result.Items {
		gotOrder = append(gotOrder, item.Path)
	}
	wantOrder := []string{mandatory.Path, smallSemantic.Path}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Errorf("result.Items order = %+v, want %+v (bigStructural excluded, ranked's own relative order otherwise preserved)", gotOrder, wantOrder)
	}
}

func TestApplyBudget_OversizedCandidateFallsBackToItsLongestFittingCoherentPrefix(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 0)
	oversized := ScoredCandidate{Candidate: Candidate{
		Path:    "ai/specs/SPEC-001.md",
		Content: strings.Repeat("a", 40) + "\n\n" + strings.Repeat("b", 40),
		Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}},
	}}
	ranked := []ScoredCandidate{mandatory, oversized}

	// Enough room for the first coherent unit (~10 tokens) but not both.
	budget := 12
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

	foundPartial := false
	for _, item := range result.Items {
		if item.Path == oversized.Path {
			foundPartial = true
			if item.Content == oversized.Content {
				t.Errorf("item %q has the full oversized content, want a coherent-unit prefix", item.Path)
			}
			if !strings.Contains(oversized.Content, item.Content) {
				t.Errorf("item %q content %q is not a prefix of the original candidate", item.Path, item.Content)
			}
		}
	}
	if !foundPartial {
		t.Errorf("result.Items = %+v, want the oversized candidate included as a coherent-unit prefix (FR-011)", result.Items)
	}
}

func TestApplyBudget_OversizedCandidateExcludedWholeWhenNotEvenFirstUnitFits(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 0)
	// Two coherent units, each still far too large for the tiny budget
	// below — even the first (preferred) unit does not fit.
	oversized := ScoredCandidate{Candidate: Candidate{
		Path:    "ai/specs/SPEC-001.md",
		Content: strings.Repeat("a", 200) + "\n\n" + strings.Repeat("b", 200),
		Reasons: []Reason{{Tier: TierStructural, Relation: "depends_on"}},
	}}
	ranked := []ScoredCandidate{mandatory, oversized}

	budget := 5 // far too small for even the first coherent unit
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

	for _, item := range result.Items {
		if item.Path == oversized.Path {
			t.Errorf("result.Items contains %q, want it excluded whole (FR-012)", oversized.Path)
		}
	}
	found := false
	for _, e := range result.Diagnostics.Exclusions {
		if e.Path == oversized.Path {
			found = true
			if e.Reason != ExclusionNoCoherentUnitFit {
				t.Errorf("Exclusions entry for %q has Reason %q, want %q", oversized.Path, e.Reason, ExclusionNoCoherentUnitFit)
			}
		}
	}
	if !found {
		t.Errorf("Diagnostics.Exclusions = %+v, want an entry for %q with reason %q", result.Diagnostics.Exclusions, oversized.Path, ExclusionNoCoherentUnitFit)
	}
}

func TestApplyBudget_IsDeterministicAcrossRepeatedCalls(t *testing.T) {
	ranked := []ScoredCandidate{
		sc("ai/memory/constitution.md", TierMandatory, "constitution", 40),
		sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 80),
		sc("ai/specs/SPEC-002.md", TierStructural, "depends_on", 8),
		sc("ai/knowledge/KNOW-001.md", TierSemantic, "wikilink", 8),
	}
	budget := 10
	req := Request{Budget: &budget}

	first := ApplyBudget(ranked, req, artifacts.DefaultEstimator{})
	second := ApplyBudget(ranked, req, artifacts.DefaultEstimator{})

	if !reflect.DeepEqual(first, second) {
		t.Errorf("ApplyBudget() is not deterministic:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestApplyBudget_DiagnosticsReportsTheEstimatorsOwnName(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 40)
	budget := 1000

	result := ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &budget}, artifacts.DefaultEstimator{})

	if result.Diagnostics.Estimator != "default" {
		t.Errorf("Diagnostics.Estimator = %q, want %q (FR-001)", result.Diagnostics.Estimator, "default")
	}

	result = ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &budget}, stubEstimator{fixed: 7})
	if result.Diagnostics.Estimator != "stub" {
		t.Errorf("Diagnostics.Estimator = %q, want %q (must reflect whichever Estimator was actually passed)", result.Diagnostics.Estimator, "stub")
	}
}

func TestApplyBudget_UsesInjectedEstimatorNotTheFreeFunctionDirectly(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 40)
	budget := 1000

	result := ApplyBudget([]ScoredCandidate{mandatory}, Request{Budget: &budget}, stubEstimator{fixed: 7})

	if len(result.Items) != 1 || result.Items[0].Tokens != 7 {
		t.Fatalf("result.Items = %+v, want one item with Tokens=7 from the injected estimator, not artifacts.EstimateTokens", result.Items)
	}
	if result.Diagnostics.TokensAvailable != 7 || result.Diagnostics.TokensSelected != 7 {
		t.Errorf("Diagnostics = %+v, want TokensAvailable/TokensSelected = 7 (from the injected estimator)", result.Diagnostics)
	}
}

// stubEstimator lets tests observe that ApplyBudget/PayloadTokens
// actually use the Estimator passed to them, rather than calling
// artifacts.EstimateTokens directly (035-context-budget-accuracy
// research.md Decision 1).
type stubEstimator struct{ fixed int }

func (s stubEstimator) Estimate(string) int { return s.fixed }
func (s stubEstimator) Name() string        { return "stub" }

func TestApplyBudget_TrimsLowerTierBeforeHigherTier(t *testing.T) {
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 40) // 10 tokens
	structural := sc("ai/specs/SPEC-001.md", TierStructural, "depends_on", 40)      // 10 tokens
	semantic := sc("ai/knowledge/KNOW-001.md", TierSemantic, "wikilink", 40)        // 10 tokens
	ranked := []ScoredCandidate{mandatory, structural, semantic}

	budget := 25 // fits mandatory+structural (20) but not +semantic (30)
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(ranked, Request{}, artifacts.DefaultEstimator{}) // no Budget set

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
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(ranked, Request{Budget: &tiny}, artifacts.DefaultEstimator{})

	if len(result.Items) != 1 || result.Items[0].Path != mandatory.Path {
		t.Fatalf("result.Items = %+v, want the full mandatory candidate, never omitted or shortened", result.Items)
	}
	wantTokens := artifacts.EstimateTokens(mandatory.Content)
	if result.Items[0].Tokens != wantTokens {
		t.Errorf("result.Items[0].Tokens = %d, want %d (mandatory content must never be shortened to force a fit)", result.Items[0].Tokens, wantTokens)
	}
}

func TestApplyBudget_MandatoryOverageIsFlaggedAndReportedExactly(t *testing.T) {
	// 035-context-budget-accuracy research.md Decision 3: BudgetExceeded/
	// Overage now compare against the hard limit, not the soft budget —
	// both are set here to the same tiny value so this test still
	// exercises "mandatory content exceeds the ceiling."
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 400) // 100 tokens
	ranked := []ScoredCandidate{mandatory}

	tiny := 10
	result := ApplyBudget(ranked, Request{Budget: &tiny, HardLimit: &tiny}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(ranked, Request{Budget: &tiny}, artifacts.DefaultEstimator{})

	for _, item := range result.Items {
		if item.Path == optional.Path {
			t.Errorf("result.Items contains the optional candidate %q; it must remain free to be omitted once mandatory alone already exceeds budget", optional.Path)
		}
	}
}

func TestApplyBudget_ZeroOrNegativeHardLimitIsExplicitNotDefault(t *testing.T) {
	// 035-context-budget-accuracy: the exceeded/overage decision now
	// hinges on HardLimit, not Budget (research.md Decision 3) — a
	// zero/negative HardLimit must not be silently promoted to
	// DefaultHardLimit (spec Edge Cases).
	mandatory := sc("ai/memory/constitution.md", TierMandatory, "constitution", 40) // 10 tokens

	for _, hardLimit := range []int{0, -1} {
		hardLimit := hardLimit
		result := ApplyBudget([]ScoredCandidate{mandatory}, Request{HardLimit: &hardLimit}, artifacts.DefaultEstimator{})

		if len(result.Items) != 1 {
			t.Fatalf("hardLimit %d: len(result.Items) = %d, want 1 (mandatory always returned in full)", hardLimit, len(result.Items))
		}
		if !result.BudgetExceeded {
			t.Errorf("hardLimit %d: result.BudgetExceeded = false, want true — a zero/negative hard limit must not be silently promoted to DefaultHardLimit", hardLimit)
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
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(nil, Request{}, artifacts.DefaultEstimator{})
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
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

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
	result := ApplyBudget(ranked, Request{Budget: &budget}, artifacts.DefaultEstimator{})

	if result.Diagnostics.TokensExcluded != 0 {
		t.Errorf("TokensExcluded = %d, want 0", result.Diagnostics.TokensExcluded)
	}
	if result.Diagnostics.ReductionPercent != 0 {
		t.Errorf("ReductionPercent = %v, want 0", result.Diagnostics.ReductionPercent)
	}
}
