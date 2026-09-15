package contextengine

import "github.com/mottamarcio/misterspec/internal/artifacts"

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
}

// Result is ApplyBudget's own output — the final, budgeted answer to
// one request.
type Result struct {
	Items []ResultItem
	// BudgetExceeded is true exactly when mandatory content alone
	// (Tier 0/1, per 015's own guarantee) exceeds the resolved budget
	// (FR-007) — never merely because some optional content was
	// trimmed, which is this feature's own normal, expected behavior
	// (research.md #10).
	BudgetExceeded bool
	// Overage is the estimated token amount by which mandatory content
	// alone exceeded the resolved budget; 0 when BudgetExceeded is
	// false (FR-007).
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
func ApplyBudget(ranked []ScoredCandidate, req Request) Result {
	budget := resolveBudget(req)

	items := make([]ResultItem, 0, len(ranked))
	var totalAvailable, selected, mandatoryTokens int
	stopped := false

	for _, c := range ranked {
		tokens := artifacts.EstimateTokens(c.Content)
		totalAvailable += tokens

		if minTier(c.Reasons) == TierMandatory {
			mandatoryTokens += tokens
			items = append(items, ResultItem{ScoredCandidate: c, Tokens: tokens})
			selected += tokens
			continue
		}

		if stopped {
			continue
		}
		if selected+tokens <= budget {
			items = append(items, ResultItem{ScoredCandidate: c, Tokens: tokens})
			selected += tokens
		} else {
			stopped = true
		}
	}

	excluded := totalAvailable - selected
	exceeded := mandatoryTokens > budget
	overage := 0
	if exceeded {
		overage = mandatoryTokens - budget
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
		},
	}
}
