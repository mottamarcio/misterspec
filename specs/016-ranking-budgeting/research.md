# Phase 0 Research: Ranking and Budgeting

All unknowns spec.md's own Assumptions section deliberately deferred to
planning are resolved below. No `NEEDS CLARIFICATION` markers remain.

## 1. Package placement

**Decision**: Extend `internal/context` (`package contextengine`, 015's
own package) with two new files — `rank.go` (US1) and `budget.go`
(US2/US3/US4) — rather than a new subpackage.

**Rationale**: This feature scores and trims 015's own `CandidateSet`
in memory; it introduces no new external dependency, no new I/O
surface, and no concept that doesn't already belong next to `Collect`,
`Request`, `Candidate`. A new subpackage would be structure without
demonstrated need (Principle IV).

**Alternatives considered**: A new `internal/context/rank` subpackage,
mirroring `internal/context/index`'s own separation. Rejected — `index`
is a separate subpackage because it owns a genuinely distinct concern
(SQLite, a real store abstraction with multiple implementations
conceivable). Ranking/budgeting has no such independent concern; it is
pure, dependency-free computation over types `contextengine` already
owns.

## 2. Two functions, not one

**Decision**: Expose two separate functions — `Rank(cs CandidateSet,
req Request) []ScoredCandidate` and `ApplyBudget(ranked
[]ScoredCandidate, req Request) Result` — rather than one combined
`RankAndBudget`.

**Rationale**: Maps 1:1 onto this feature's own two primary,
independently-testable user stories (US1 ranking, US2/US3 budgeting)
and mirrors `docs/context-engine-implementation.md` §16's own pipeline,
which lists "10. Rank candidates" and "11. Apply token budget" as
distinct steps. Each function is independently unit-testable without
constructing the other's preconditions by hand (US1's Independent Test
needs no budget at all; US2/US3's Independent Test needs no live
`CandidateSet`, only a hand-built ranked slice).

**Alternatives considered**: A single `Rank(cs, req) Result` that does
both. Rejected — it would force every ranking-only test to also reason
about budget defaults, and every budgeting-only test to first produce a
plausible ranked slice through `Rank` itself rather than constructing
one directly, weakening test isolation for no benefit (Principle V).

## 3. The scoring formula

**Decision**: A candidate's `Score` is `relationWeight + intentWeight +
textRelevance`, computed purely for **within-tier** ordering.
`Tier` (already carried by 015's own `Reason.Tier`) is the sole,
absolute primary sort key — `Score` is never summed with, or allowed to
outweigh, a `Tier` difference (FR-002, FR-003). This directly encodes
§17.1's "important ranking invariant": a strong text match can never
outrank mandatory or structural content, because tier and score never
compete in the same comparison.

Weights (deliberately example values, not immutable — matching §17's
own "these values are examples" guidance, and this project's own
established test discipline of asserting *ordering*, not magic
numbers):

```text
relationWeight:
  "constitution", "target"                 100
  "parent", "depends_on", "supersedes"       90
  "wikilink"                                 80
  "backlink"                                 65
  "text_match"                                0  (scored via textRelevance instead)
  anything else (e.g. a second-hop relation)  45

textRelevance: min(4, occurrences of each query/task term in
  Heading+Content) * 10   → range 0..40, mirroring §17's own
  "BM25 contribution 0..40" flavor without depending on a real BM25
  value (see #6 below).

intentWeight: +5 when at least one of the candidate's own Reason
  relations is "preferred" for the request's Intent (§15's own
  informal per-intent preference lists, mapped onto this project's
  actual, narrower relation vocabulary), else 0.
```

When a merged `Candidate` carries more than one `Reason` (found through
more than one path — FR-008/FR-009 already require every Reason to
survive the merge), `relationWeight`/`intentWeight` use the strongest
(maximum) signal among all of that candidate's Reasons — "the best
known reason wins," never double-counted or averaged down.

**Rationale**: Simple, additive, fully deterministic, and requires no
new dependency or learned model, matching §17's own explicit
preference ("a simple initial model is preferable to a complex learned
model").

**Alternatives considered**: A single combined numeric score across all
tiers (e.g. `tierWeight*1000 + relationWeight + ...`), relying on tier
weight magnitude to dominate the sum. Rejected — encoding the invariant
as arithmetic dominance is fragile (a future weight tweak could
silently break it) and untestable-by-construction; keeping `Tier` as a
wholly separate, structural sort key makes the invariant impossible to
violate by accident, not just unlikely.

## 4. Intent-to-relation preference mapping

**Decision**: A small, fixed map from each `Intent` constant to the
relation strings this codebase actually has (not the source document's
more abstract, hypothetical categories like "acceptance scenarios" or
"evidence," which have no corresponding relation kind yet):

```text
IntentPlanning       → "parent", "depends_on"
IntentTasks          → "depends_on", "parent"
IntentImplementation → "wikilink", "backlink", "text_match"
IntentValidation     → "backlink", "text_match"
IntentAnalysis       → "backlink", "text_match"
```

**Rationale**: §15 is explicitly "conceptual" guidance describing an
agent workflow richer than this project's current relation vocabulary
("Constitution", "parent Feature/Program", "dependencies", "explicit
Knowledge/Learning links", etc.); this project's own five relation
strings (`parent`, `depends_on`, `supersedes`, `wikilink`, `backlink`,
plus `text_match`/`constitution`/`target`) are what `operations.
References`/`Backlinks` and `Collect` already produce, so the mapping
translates §15's intent into the vocabulary that actually exists today
rather than inventing new relation kinds this feature has no mandate to
add (Principle IV).

## 5. Text relevance is recomputed, not reused from FTS5's bm25()

**Decision**: `textRelevance` is a fresh, simple term-overlap count
against `Candidate.Heading`/`Content`, using the same query-resolution
rule Tier 4 already uses (`Request.Query`, falling back to
`Request.Task` — verbatim, never fabricated). It does **not** reuse
`index.SearchResult.Rank` (SQLite's own `bm25()` value).

**Rationale**: 015's `Collect` intentionally does not carry
`SearchResult.Rank` forward onto `Candidate` — `Reason{Tier: TierText,
Relation: "text_match"}` records *that* a chunk matched, not *how well*.
By the time a `CandidateSet` reaches this feature, `mergeAndSort` has
already re-sorted every Tier 4 candidate by `(Path, StartLine)`, so the
original bm25 ordering is gone regardless. Recomputing a small,
self-contained heuristic here avoids reopening 015's already-merged,
already-tested `collector.go`/`result.go` (no regression risk to a
shipped feature) and keeps this feature fully decoupled from
`internal/context/index`'s own storage detail — `Rank`/`ApplyBudget`
take a `CandidateSet` and a `Request`, nothing else, and remain provably
read-only (FR-012) since they touch no store, no filesystem, at all.

**Alternatives considered**: Adding a `TextRank float64` field to
`Candidate` in 015's `result.go` so the true bm25 value survives the
merge. Rejected — it would touch a shipped, merged feature's own type
for a benefit (marginally more faithful text ordering) this project's
own test discipline (assert ordering, not exact scores) doesn't need,
and would reopen `mergeAndSort` behavior other features may already
depend on.

## 6. Optional budget representation

**Decision**: `Request` (015's own type) gains one new field,
`Budget *int`. `nil` means "not specified — use `DefaultBudget`"; a
non-nil pointer, even to `0` or a negative number, is a real, explicit
budget request.

**Rationale**: spec.md's own Edge Cases require distinguishing "no
budget was requested" (→ default) from "a budget of exactly zero (or
negative) was requested" (→ an explicit, extremely small budget,
**not** silently promoted to the default). A plain `int` field cannot
carry that distinction, since `0` would be ambiguous between the two.
A pointer is the minimal Go idiom that resolves the ambiguity without
inventing a new wrapper type or sentinel constant (Principle IV);
`Chunk.Tokens` (013) and `Request.Budget` (015) were both deliberately
left unset until a real consumer needed them — this feature is that
consumer for `Budget`, so it is added now, not before.

**Alternatives considered**: A sentinel value (e.g. `-1` meaning
"use default"). Rejected — the edge cases explicitly require a negative
value to be treated as *itself* an explicit tiny budget, not as
"unspecified," so no integer sentinel is actually available.

## 7. Default budget constant

**Decision**: `const DefaultBudget = 6000`, exported from
`internal/context` (`contextengine.DefaultBudget`).

**Rationale**: `docs/context-engine-implementation.md` §19's own
illustrative constant, explicitly offered as "easy to change after
measurement" — adopted verbatim as this project's actual starting
default, per spec.md's own Assumption that the exact figure is
planning's call, not a new user-facing configuration knob (Principle
IV — no `.misterspec/config.yaml` field is added for this).

## 8. Mandatory-content detection

**Decision**: A `ScoredCandidate` counts as mandatory when its own
lowest `Reason.Tier` is `TierMandatory` — computed via 015's own
unexported `minTier` helper (same package, already fully tested by
015's own `result_test.go`), not a reimplementation.

**Rationale**: DRY (Principle VI) — `minTier` already exists, already
correctly handles a merged candidate carrying Reasons from more than
one Tier (it never happens for `TierMandatory` specifically, since
Tier 0/1 candidates are only ever discovered as Tier 0/1, but reusing
the same helper for every tier check keeps exactly one definition of
"a candidate's own effective Tier" in the whole package).

## 9. Budget-fill algorithm: strict tier-ordered greedy accumulation

**Decision**: `ApplyBudget` walks `ranked` (already fully,
deterministically ordered by `Rank`) in order, always including every
mandatory item in full regardless of remaining budget (FR-006), then
for optional items accumulating tokens while they fit — and, at the
very first optional item that would *not* fit, stopping entirely rather
than skipping it and trying a later, possibly-smaller item.

**Rationale**: §19.1 frames this literally as "budget by priority
tiers," not as bin-packing for maximum utilization; the source
document's own explicit warning — "do not simply sort every chunk by
one score and truncate blindly" — is about *cross-tier* truncation, but
the simplest, fully deterministic reading that also avoids a
knapsack-style optimization problem (with its own tie-breaking
subtleties) is a hard stop at the first miss. This keeps the result
trivially predictable and explainable ("everything up to here fit, the
rest didn't") — directly serving US4's own "Explain Every Decision"
goal — at the cost of not always producing the byte-for-byte *smallest
possible* excess capacity, a tradeoff §17's "simple model over complex
model" preference already endorses.

**Alternatives considered**: Skip-and-continue (take a later smaller
item if an earlier larger one doesn't fit). Rejected as unnecessary
complexity (Principle IV) with no requirement demanding it — FR-005
requires *fitting within budget*, not *optimal packing*.

## 10. `BudgetExceeded`/overage scope

**Decision**: `Result.BudgetExceeded` and `Result.Overage` describe
exactly one condition — mandatory content's own combined token cost
exceeding the requested budget (FR-007) — never ordinary optional-tier
trimming, which is US2's normal, expected behavior, not a failure
state.

**Rationale**: Matches §19.2's own framing ("if mandatory context alone
exceeds the budget...") and spec.md's own User Story 3, which is
scoped specifically to the mandatory-overage scenario. A single boolean
that also fired on routine optional trimming would make "was anything
unusual here?" unanswerable from the flag alone.

## 11. Result/Diagnostics shape

**Decision**:

```go
type ScoredCandidate struct {
    Candidate
    Score int
}

type ResultItem struct {
    ScoredCandidate
    Tokens int
}

type Diagnostics struct {
    CandidatesConsidered int
    ItemsSelected        int
    TokensAvailable      int
    TokensSelected       int
    TokensExcluded       int
    ReductionPercent     float64
}

type Result struct {
    Items          []ResultItem
    BudgetExceeded bool
    Overage        int
    Diagnostics    Diagnostics
}
```

**Rationale**: Follows §20's own suggested `ContextItem`/`Result` shape
closely, adapted to types this package already has (`Candidate`
embedding rather than duplicating `Path`/`Heading`/`Content`/
`StartLine`/`EndLine`/`Reasons`) and to what this feature actually
needs — no `ArtifactID` field (015 never introduced one; `Path` is this
whole subsystem's identity per 013's own precedent), no `Section`
string duplicate of `Heading`, no separate `Intent`/`Target` echo
fields on `Result` (the caller already has its own `Request`). No JSON
tags yet — Phase 8's own CLI rendering is where a wire format is
decided (Principle IV, matching 015's own "no rendering here" scoping).

## 12. Token estimation reused verbatim

**Decision**: `ResultItem.Tokens` is `artifacts.EstimateTokens(sc.
Content)` — 013's own estimator, called directly, no new estimation
logic.

**Rationale**: Spec's own Assumption; DRY (Principle VI); consistent
with 014's own chunk `token_estimate` column, which uses the same
function.

## Summary of Go footprint

New files only: `internal/context/rank.go`, `internal/context/
rank_test.go`, `internal/context/budget.go`, `internal/context/
budget_test.go`. One field added to the already-merged `internal/
context/request.go` (`Budget *int`). No new package, no new external
dependency, no CLI change (Phase 8's own later scope), no change to
`internal/context/index`, `internal/artifacts`, `internal/operations`,
or `internal/validation`.
