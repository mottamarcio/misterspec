# Phase 1 Data Model: Ranking and Budgeting

All types live in `internal/context` (`package contextengine`), 015's
own package. No new package, no persisted schema — everything here is
an in-memory, read-only computation over a `CandidateSet` (research.md
#1).

## Request (extended)

015's own `Request` gains exactly one new field:

```go
type Request struct {
    Target string
    Task   string
    Intent Intent
    Query  string
    // Budget optionally overrides DefaultBudget (FR-004). nil means
    // "not specified" — DefaultBudget is used. A non-nil pointer, even
    // to zero or a negative number, is a real, explicit budget request
    // (research.md #6) — never silently promoted to the default.
    Budget *int
}
```

Validation: none needed beyond what already exists — a `nil` or
explicit `Budget` are both valid; `ApplyBudget` treats any resolved
value (including zero or negative) uniformly (research.md #9).

## ScoredCandidate

```go
// ScoredCandidate is one 015 Candidate carrying a deterministic
// relevance Score, used only to order candidates already known to
// share the same Tier (FR-001, FR-002, FR-003) — Score is never
// compared, or summed with anything, across a Tier boundary.
type ScoredCandidate struct {
    Candidate
    Score int
}
```

Fields inherited from `Candidate` (unchanged): `Path`, `Heading`,
`Content`, `StartLine`, `EndLine`, `Reasons []Reason`.

**Derivation** (`Rank`, research.md #3):

```text
Score = relationWeight(best Reason) + intentWeight(Intent, Reasons) + textRelevance(Content, Heading, query terms)
```

Where "best Reason" is the Reason among the candidate's own (possibly
multiple, post-merge) Reasons with the highest `relationWeight`.

## ResultItem

```go
// ResultItem is one ScoredCandidate that survived budgeting, carrying
// its own estimated token cost (FR-009).
type ResultItem struct {
    ScoredCandidate
    Tokens int // artifacts.EstimateTokens(Content)
}
```

## Diagnostics

```go
// Diagnostics reports the measurable counts and token estimates for
// one ApplyBudget call (FR-010), consistent with §22's own suggested
// retrieval-metrics shape.
type Diagnostics struct {
    CandidatesConsidered int     // len(ranked) — every candidate Rank produced
    ItemsSelected        int     // len(Result.Items)
    TokensAvailable      int     // sum of EstimateTokens(Content) across every ranked candidate
    TokensSelected       int     // sum of EstimateTokens(Content) across Result.Items
    TokensExcluded       int     // TokensAvailable - TokensSelected
    ReductionPercent     float64 // TokensExcluded / TokensAvailable * 100, or 0 when TokensAvailable is 0
}
```

Invariant (SC-004): `TokensAvailable - TokensExcluded == TokensSelected`
always holds exactly (no rounding involved — all three are integer sums
of the same per-candidate `EstimateTokens` values).

## Result

```go
// Result is ApplyBudget's own output — the final, budgeted answer to
// one request (FR-005 through FR-010).
type Result struct {
    Items []ResultItem
    // BudgetExceeded is true exactly when mandatory content alone
    // (Tier 0 + Tier 1, per 015's own guarantee) exceeds the resolved
    // budget (FR-007, research.md #10) — never merely because some
    // optional content was trimmed, which is Us2's own normal,
    // expected behavior.
    BudgetExceeded bool
    // Overage is the estimated token amount by which mandatory content
    // alone exceeded the resolved budget; 0 when BudgetExceeded is
    // false (FR-007).
    Overage     int
    Diagnostics Diagnostics
}
```

**State/derivation rules**:
- Every mandatory `ScoredCandidate` (its own `minTier(Reasons) ==
  TierMandatory`) always appears in `Items`, in full, regardless of
  budget (FR-006).
- Optional (`Tier > TierMandatory`) candidates are appended, in
  `ranked` order, only while doing so keeps the running total at or
  under the resolved budget; the first optional candidate that would
  not fit stops further inclusion entirely (research.md #9) — no lower
  or equal-priority candidate after it is considered either.
- `Items` preserves `ranked`'s own order exactly (already
  tier-then-score-then-path-then-line total order from `Rank`) — no
  re-sorting inside `ApplyBudget`.

## Relationships to existing types

```text
CandidateSet (015)  --Rank-->  []ScoredCandidate  --ApplyBudget-->  Result
     |                              |                                   |
  Candidate                    Candidate (embedded)              ScoredCandidate (embedded)
  Reason{Tier, Relation}       + Score                            + Tokens
                                                                   Diagnostics
```

No existing 015 type (`Intent`, `Tier`, `Reason`, `Candidate`,
`CandidateSet`) changes shape — `Request` gains one field (`Budget`),
everything else is additive, new types only.
