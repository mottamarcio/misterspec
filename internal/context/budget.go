package contextengine

import (
	"strings"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

// DefaultBudget is the fixed, documented token budget ApplyBudget uses
// whenever a Request's own Budget field is nil (FR-004,
// docs/context-engine-implementation.md §19, 016-ranking-budgeting/
// research.md #7). Deliberately an internal constant, not a new
// project configuration field — easy to change later once real usage
// provides evidence for a better default (research.md #7).
const DefaultBudget = 6000

// resolveBudget returns req.Budget's own pointed-to value when it is
// non-nil — including zero or a negative number, both real, explicit
// budget requests, never treated as "unspecified" (research.md #6) —
// or DefaultBudget when req.Budget is nil.
func resolveBudget(req Request) int {
	if req.Budget != nil {
		return *req.Budget
	}
	return DefaultBudget
}

// DefaultHardLimit is the fixed, documented ceiling above which
// mandatory content is reported as exceeded, whenever a Request's own
// HardLimit field is nil (035-context-budget-accuracy FR-006,
// research.md Decision 2) — a fixed multiple of DefaultBudget, always
// finite; never treated as "no limit."
const DefaultHardLimit = 2 * DefaultBudget

// resolveHardLimit returns req.HardLimit's own pointed-to value when
// non-nil — including zero or negative, both real, explicit requests —
// or DefaultHardLimit when req.HardLimit is nil (035-context-budget-
// accuracy FR-006).
func resolveHardLimit(req Request) int {
	if req.HardLimit != nil {
		return *req.HardLimit
	}
	return DefaultHardLimit
}

// ResultItem is one ScoredCandidate that survived budgeting, carrying
// its own estimated token cost (FR-009).
type ResultItem struct {
	ScoredCandidate
	Tokens int
}

// Diagnostics reports the measurable counts and token estimates for
// one ApplyBudget call (FR-010, docs/context-engine-implementation.md
// §22).
type Diagnostics struct {
	CandidatesConsidered int
	ItemsSelected        int
	TokensAvailable      int
	TokensSelected       int
	TokensExcluded       int
	ReductionPercent     float64
	// Estimator is the Name() of the Estimator ApplyBudget was called
	// with — the same value that produced every token count in this
	// Diagnostics (035-context-budget-accuracy FR-001).
	Estimator string
	// HardLimit is the resolved ceiling actually applied — always
	// present, even when the caller's own Request.HardLimit was nil
	// (035-context-budget-accuracy FR-006).
	HardLimit int
	// Exclusions names every candidate, or partial-unit fallback, left
	// out of Items and why (035-context-budget-accuracy FR-010).
	Exclusions []ExclusionRecord
}

// ExclusionRecord is one Diagnostics.Exclusions entry — a candidate
// that could have contributed but was left out
// (035-context-budget-accuracy data-model.md "ExclusionRecord").
type ExclusionRecord struct {
	Path    string
	Heading string
	// Reason is one of a closed, stable set of labels:
	// "did_not_fit_remaining_budget" or "no_coherent_unit_fit".
	Reason string
}

const (
	// ExclusionDidNotFitRemainingBudget marks a whole optional
	// candidate skipped because it did not fit in the remaining
	// soft-budget space — backfill continues past it to try smaller,
	// later same-tier candidates (research.md Decision 4).
	ExclusionDidNotFitRemainingBudget = "did_not_fit_remaining_budget"
	// ExclusionNoCoherentUnitFit marks a whole optional candidate
	// excluded because not even its own first coherent unit fit
	// (research.md Decision 5).
	ExclusionNoCoherentUnitFit = "no_coherent_unit_fit"
)

// Result is ApplyBudget's own output — the final, budgeted answer to
// one request.
type Result struct {
	Items []ResultItem
	// BudgetExceeded is true exactly when mandatory content alone
	// (Tier 0/1, per 015's own guarantee) exceeds the resolved hard
	// limit (035-context-budget-accuracy FR-007, research.md Decision
	// 3 — a deliberate redefinition: previously compared against the
	// soft budget) — never merely because some optional content was
	// trimmed, which is this feature's own normal, expected behavior
	// (research.md #10).
	BudgetExceeded bool
	// Overage is the estimated token amount by which mandatory content
	// alone exceeded the resolved hard limit; 0 when BudgetExceeded is
	// false (FR-007, research.md Decision 3).
	Overage     int
	Diagnostics Diagnostics
}

// ApplyBudget resolves req's own budget (DefaultBudget when
// req.Budget is nil), always includes every mandatory (Tier 0/1)
// candidate from ranked in full regardless of budget (FR-006), then
// includes optional candidates in ranked's own order while they fit,
// stopping entirely — never skipping ahead to a later, possibly
// smaller one — at the first optional candidate that does not
// (FR-005, research.md #9). When mandatory content alone exceeds the
// resolved budget, Result.BudgetExceeded is true and Result.Overage
// reports the estimated excess (FR-007); optional content remains
// free to be omitted in that case too (FR-008). Deterministic for the
// same inputs (FR-011). Pure — no I/O, no mutation (FR-012).
func ApplyBudget(ranked []ScoredCandidate, req Request, estimator artifacts.Estimator) Result {
	budget := resolveBudget(req)
	hardLimit := resolveHardLimit(req)

	items := make([]ResultItem, 0, len(ranked))
	var exclusions []ExclusionRecord
	var totalAvailable, selected, mandatoryTokens int

	for _, c := range ranked {
		tokens := estimator.Estimate(c.Content)
		totalAvailable += tokens

		if minTier(c.Reasons) == TierMandatory {
			mandatoryTokens += tokens
			items = append(items, ResultItem{ScoredCandidate: c, Tokens: tokens})
			selected += tokens
			continue
		}

		if selected+tokens <= budget {
			items = append(items, ResultItem{ScoredCandidate: c, Tokens: tokens})
			selected += tokens
			continue
		}

		// c does not fit whole in the remaining soft-budget space.
		// Backfill (research.md Decision 4): record the exclusion and
		// keep going — never abandon the rest of ranked's own order,
		// which already guarantees no lower-tier candidate is ever
		// reached before every higher-tier candidate has been decided
		// (FR-008, FR-009). Before giving up on c entirely, try its
		// longest fitting prefix of coherent units (research.md
		// Decision 5, FR-011).
		remaining := budget - selected
		units := artifacts.SplitIntoCoherentUnits(c.Content)

		var prefix strings.Builder
		prefixTokens := 0
		for _, u := range units {
			uTokens := estimator.Estimate(u)
			if prefixTokens+uTokens > remaining {
				break
			}
			prefix.WriteString(u)
			prefixTokens += uTokens
		}

		if prefixTokens > 0 {
			partial := c
			partial.Content = prefix.String()
			items = append(items, ResultItem{ScoredCandidate: partial, Tokens: prefixTokens})
			selected += prefixTokens
			continue
		}

		reason := ExclusionDidNotFitRemainingBudget
		if len(units) > 1 {
			// c had genuine internal structure to split on, but not
			// even its own first, most-preferred unit fit (FR-012) —
			// distinct from an atomic candidate that simply has no
			// smaller piece to offer.
			reason = ExclusionNoCoherentUnitFit
		}
		exclusions = append(exclusions, ExclusionRecord{Path: c.Path, Heading: c.Heading, Reason: reason})
	}

	excluded := totalAvailable - selected
	exceeded := mandatoryTokens > hardLimit
	overage := 0
	if exceeded {
		overage = mandatoryTokens - hardLimit
	}

	reduction := 0.0
	if totalAvailable > 0 {
		reduction = float64(excluded) / float64(totalAvailable) * 100
	}

	return Result{
		Items:          items,
		BudgetExceeded: exceeded,
		Overage:        overage,
		Diagnostics: Diagnostics{
			CandidatesConsidered: len(ranked),
			ItemsSelected:        len(items),
			TokensAvailable:      totalAvailable,
			TokensSelected:       selected,
			TokensExcluded:       excluded,
			ReductionPercent:     reduction,
			Estimator:            estimator.Name(),
			HardLimit:            hardLimit,
			Exclusions:           exclusions,
		},
	}
}
