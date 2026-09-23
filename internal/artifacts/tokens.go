package artifacts

// Estimator estimates the token cost of a piece of text. Kept as an
// interface so a real provider-specific tokenizer can be substituted
// later without changing any caller
// (docs/context-engine-implementation.md §9).
type Estimator interface {
	Estimate(text string) int
	// Name identifies which estimator produced a set of estimates —
	// always non-empty (035-context-budget-accuracy FR-001), so a
	// response can say which method computed its own numbers instead
	// of presenting an approximation as an unattributed exact count.
	Name() string
}

// DefaultEstimator is this project's own deterministic approximation —
// see EstimateTokens for the exact method.
type DefaultEstimator struct{}

// Estimate implements Estimator by calling EstimateTokens.
func (DefaultEstimator) Estimate(text string) int {
	return EstimateTokens(text)
}

// Name implements Estimator, identifying this as the built-in
// approximation — never presented as a real tokenizer's exact count
// (035-context-budget-accuracy FR-003).
func (DefaultEstimator) Name() string {
	return "default"
}

// EstimateTokens returns a deterministic, approximate token count for
// text: ceil(rune count / 4), a widely used rough approximation for
// common language-model tokenizers on English prose (FR-009). Not a
// real provider tokenizer — the source document explicitly allows a
// documented, consistently applied approximation for the MVP
// (docs/context-engine-implementation.md §9, 013-document-model-
// chunking/research.md #7). EstimateTokens("") is always 0.
func EstimateTokens(text string) int {
	n := len([]rune(text))
	if n == 0 {
		return 0
	}
	return (n + 3) / 4
}
