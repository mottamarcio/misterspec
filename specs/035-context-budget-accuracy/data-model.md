# Phase 1 Data Model: Orçamento sobre a saída efetiva de contexto

No persisted storage is added (Constitution Principle III). This extends `internal/artifacts`'s existing `Estimator`/`DefaultEstimator` and `internal/context`'s existing `Request`/`Result`/`Diagnostics` — additive fields plus one deliberate, versioned redefinition (research.md Decision 3).

## Estimator (extended, `internal/artifacts`)

| Member | Type | Description |
|---|---|---|
| `Estimate(text string) int` | method (existing) | Unchanged — the cost of one piece of text. |
| `Name() string` | method (**new**) | A short, stable identifier for this estimator (spec FR-001). `DefaultEstimator.Name()` returns `"default"`. |

**Validation rule**: `Name()` MUST be non-empty for any implementation — a response that can't say *which* estimator produced its numbers doesn't satisfy FR-001.

## Request (extended, `internal/context`)

| Field | Type | Description |
|---|---|---|
| `Budget` | `*int` (existing) | The soft limit — the optional-content target. Unchanged meaning (research.md Decision 2). |
| `HardLimit` | `*int` (**new**) | The hard limit — the ceiling above which mandatory content is reported as exceeded. `nil` resolves to `DefaultHardLimit`. |

**Validation rule** (spec Edge Cases): a `HardLimit` smaller than the resolved soft `Budget` is not rejected as invalid — the two are independent knobs; a caller setting a smaller hard limit than soft budget gets exactly what that combination means (a tight mandatory-content ceiling, regardless of how generous the optional-content target is).

## DefaultHardLimit (`internal/context`, new constant)

`DefaultHardLimit = 2 * DefaultBudget` (`12000`) — a fixed multiple of the existing `DefaultBudget`, always finite (research.md Decision 2, spec FR-006).

## Diagnostics (extended, `internal/context`)

| Field | Type | Description |
|---|---|---|
| `CandidatesConsidered`, `ItemsSelected`, `TokensAvailable`, `TokensSelected`, `TokensExcluded`, `ReductionPercent`, `PayloadTokens` | (existing, from 033) | Unchanged meaning. |
| `Estimator` | `string` (**new**) | The `Name()` of the `Estimator` used to produce every token count in this response (spec FR-001). |
| `HardLimit` | `int` (**new**) | The resolved hard limit actually applied (spec.md "Hard Limit" — always present, even when the caller didn't set one, per FR-006). |
| `Exclusions` | `[]ExclusionRecord` (**new**) | Every candidate or partial-unit-fallback that could have contributed but was left out, and why (spec FR-010). |

## Result (extended, `internal/context`)

| Field | Type | Description |
|---|---|---|
| `Items` | `[]ResultItem` (existing) | Unchanged shape. A backfilled smaller candidate, or a partial-unit fallback (Decision 5), is just another `ResultItem` — no new item shape. |
| `BudgetExceeded` | `bool` (existing name, **redefined**) | Now `mandatoryTokens > resolvedHardLimit` — previously compared against the soft budget (research.md Decision 3). |
| `Overage` | `int` (existing name, **redefined**) | Now `mandatoryTokens - resolvedHardLimit` when `BudgetExceeded`, else `0`. |
| `Diagnostics` | `Diagnostics` (extended above) | |

**Validation rule** (spec FR-007): `BudgetExceeded == true` never removes anything from `Items` — every mandatory candidate is always present in full, regardless of `Overage`'s size.

## ExclusionRecord (`internal/context`, new type)

One entry in `Diagnostics.Exclusions` (spec.md "Exclusion Record").

| Field | Type | Description |
|---|---|---|
| `Path` | `string` | The excluded candidate's own path. |
| `Heading` | `string` | Its own heading, if any. |
| `Reason` | `string` | A short, stable label — `"did_not_fit_remaining_budget"` (whole candidate skipped, backfill continued past it) or `"no_coherent_unit_fit"` (not even the first coherent unit of an oversized candidate fit — FR-012). |

## CoherentUnit (`internal/artifacts`, conceptual — `SplitIntoCoherentUnits` returns `[]string`)

A contiguous, paragraph-bounded run of one Chunk's own body text that can stand alone without breaking a fenced code block in the middle (spec.md "Coherent Unit", research.md Decision 5).

**Validation rules**:
- A fenced code block's opening and closing lines, and everything between them, always belong to the same unit — never split across two units (spec FR-011).
- Concatenating every returned unit, in order, reproduces the original content exactly (a pure partition, no content dropped or duplicated by the split itself — only by the caller's own choice of which prefix to include).

## State / Lifecycle

No new lifecycle or persisted state. Every value here (resolved hard limit, exclusions, coherent units) is recomputed fresh from the same in-memory `Result`/`Candidate` data `ApplyBudget` already holds, on every call.
