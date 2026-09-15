# Phase 7 Contracts: Ranking and Budgeting

**Reconciled against the actual implementation (T013)** — zero drift.
Every signature below (`Request.Budget`, `DefaultBudget`,
`ScoredCandidate`, `Rank`, `ResultItem`, `Diagnostics`, `Result`,
`ApplyBudget`) matches what shipped, verbatim. Unexported helpers
(`resolveBudget`, `relationWeight`, `intentWeight`, `textRelevance`,
`scoreCandidate`, `queryTerms`, `preferredRelationsByIntent`) are
internal decomposition detail, not part of this feature's exported
surface.

No HTTP/CLI surface is added by this feature (research.md #1) — its
contract is the new Go API added to `internal/context`, on top of
015's own `Candidate`/`CandidateSet`/`Request`.

## `internal/context` (extended)

```go
package contextengine

// Request (015, extended): one new field.
type Request struct {
    Target string
    Task   string
    Intent Intent
    Query  string
    // Budget optionally overrides DefaultBudget. nil means "use the
    // default"; a non-nil zero or negative value is a real, explicit
    // budget request (research.md #6).
    Budget *int
}

// DefaultBudget is the fixed, documented budget used whenever
// Request.Budget is nil (FR-004, research.md #7).
const DefaultBudget = 6000

// ScoredCandidate is one Candidate carrying a deterministic relevance
// Score, meaningful only among candidates already known to share the
// same Tier (FR-001, FR-002, FR-003).
type ScoredCandidate struct {
    Candidate
    Score int
}

// Rank assigns every Candidate in cs a Score and returns them in the
// feature's one required order: ascending Tier first — absolutely,
// never crossed by Score (FR-002, FR-003) — then descending Score,
// then Path, then StartLine, matching 015's own tie-break convention.
// Deterministic for the same inputs (FR-011). Pure — no I/O.
func Rank(cs CandidateSet, req Request) []ScoredCandidate

// ResultItem is one ScoredCandidate that survived budgeting, carrying
// its own estimated token cost (FR-009).
type ResultItem struct {
    ScoredCandidate
    Tokens int
}

// Diagnostics reports the measurable counts and token estimates for
// one ApplyBudget call (FR-010).
type Diagnostics struct {
    CandidatesConsidered int
    ItemsSelected        int
    TokensAvailable      int
    TokensSelected       int
    TokensExcluded       int
    ReductionPercent     float64
}

// Result is ApplyBudget's own output.
type Result struct {
    Items          []ResultItem
    BudgetExceeded bool
    Overage        int
    Diagnostics    Diagnostics
}

// ApplyBudget resolves req.Budget (DefaultBudget when nil), always
// includes every mandatory (Tier 0/1) candidate from ranked in full
// (FR-006), then includes optional candidates in ranked order while
// they fit, stopping entirely at the first one that does not
// (FR-005, research.md #9). When mandatory content alone exceeds the
// resolved budget, Result.BudgetExceeded is true and Result.Overage
// reports the estimated excess (FR-007) — optional content remains
// free to be omitted in that case too (FR-008). Deterministic for the
// same inputs (FR-011). Pure — no I/O, no mutation (FR-012).
func ApplyBudget(ranked []ScoredCandidate, req Request) Result
```

**Guarantees**:
- `Rank`/`ApplyBudget` never modify any project artifact, the
  reference graph, or the search index — both take only in-memory
  values and return in-memory values (FR-012).
- A candidate from a higher-priority Tier always ranks, and is always
  considered for budgeting, before any candidate from a lower-priority
  Tier, regardless of Score (FR-002). `Request.Intent` can only ever
  reorder candidates that already share one Tier (FR-003).
- Every mandatory candidate `Rank` produces always appears in
  `ApplyBudget`'s own `Result.Items`, in full, regardless of the
  resolved budget (FR-006, US3).
- `Result.Diagnostics` is always internally consistent:
  `TokensAvailable - TokensExcluded == TokensSelected`, and
  `ItemsSelected == len(Result.Items)` (SC-004).
- Calling `Rank` then `ApplyBudget` twice with the same `CandidateSet`
  and `Request`, against unchanged project content, produces
  byte-for-byte identical output every time (FR-011, SC-005).
- Neither function reads `Request.Target`/calls into `operations` or
  `internal/context/index` at all — both operate purely on the
  `CandidateSet`/`[]ScoredCandidate` and `Request` values already
  passed in (research.md #5).

## Cross-cutting: no change to any existing contract

`internal/artifacts`, `internal/validation`, `internal/operations`,
`internal/context/index`, `internal/cli`, and 015's own `Collect`,
`CandidateSet`, `Candidate`, `Reason`, `Tier`, `Intent` are completely
untouched in shape — only `Request` gains the new `Budget` field
(additive, zero-value-compatible: `Budget: nil` behaves exactly as
every existing 015 caller/test already expects). No new CLI command
(research.md's own "Summary of Go footprint") — Phase 8's own "internal
context" command remains the later, actual CLI surface this capability
feeds into.
