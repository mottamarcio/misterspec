# Implementation Plan: Ranking and Budgeting

**Branch**: `016-ranking-budgeting` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/016-ranking-budgeting/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 7 ("Ranking
and Budgeting") exactly: `internal/context` (015's own `contextengine`
package) gains two new pure, read-only functions on top of 015's own
`CandidateSet` — `Rank`, which assigns every candidate a deterministic
`Score` used only to order candidates already known to share the same
priority Tier (Tier itself remains the absolute, never-crossed primary
sort key — §17.1's own "important ranking invariant"), and
`ApplyBudget`, which fits a ranked list into a token budget by filling
tier-by-tier, always preserving mandatory (Constitution + target)
content in full, flagging and reporting the overage when mandatory
content alone exceeds the budget, and reporting measurable diagnostics
(candidates considered/selected, tokens available/selected/excluded,
reduction percentage). `Request` gains one additive field, `Budget
*int` (nil = use the fixed `DefaultBudget = 6000`). No new package, no
new dependency, no CLI surface (Phase 8's own later scope), no change
to any existing 011-015 file's shape beyond that one field.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: Go standard library only — no new dependency.
**Storage**: None — `Rank`/`ApplyBudget` are pure in-memory functions over values already in hand (`CandidateSet`, `[]ScoredCandidate`, `Request`); neither touches the filesystem or `internal/context/index` at all (FR-012, research.md #5).
**Testing**: `go test` — new `internal/context` tests covering `docs/context-engine-implementation.md` §29.7's own Budget test list (result below budget, optional content removed at the budget boundary, mandatory content retained, mandatory content exceeding budget, deterministic output, token estimates included) plus this feature's own ranking-invariant tests (mandatory/structural never outranked by a strong text match regardless of score, same-tier ordering driven by score, intent influencing only within a tier, stable deterministic order for identical-relevance candidates); 011's, 012's, 013's, 014's, and 015's own full suites re-run unmodified as the named regression gate. Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. `internal/context` gains two new files (`rank.go`, `budget.go`) and one additive field on its existing `Request` type; no other package touched, no CLI command added.
**Performance Goals**: `Rank` is O(n log n) in candidate count (one sort); `ApplyBudget` is O(n) over the already-ranked slice — negligible next to the per-artifact filesystem/index costs `Collect` itself already pays upstream.
**Constraints**: `Rank`/`ApplyBudget` are strictly read-only (FR-012) — no mutation of any project artifact, the reference graph, or the search index; neither function performs any I/O of its own. A candidate's Tier MUST always dominate its Score in ordering (FR-002); Intent MAY only reorder candidates already sharing a Tier (FR-003). Mandatory (Tier 0/1) content MUST always be returned in full regardless of budget (FR-006). Determinism is required end-to-end for identical inputs (FR-011).
**Scale/Scope**: `internal/context/{rank.go, rank_test.go, budget.go, budget_test.go}` (new); `internal/context/request.go` gains one field (`Budget *int`). No existing package modified, no CLI change, no new dependency.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Scoring and budgeting are purely mechanical (a fixed additive formula, a greedy tier-ordered fill) — no judgment about *what the request means* happens here; that boundary was already drawn by 015's own `Collect` and stays untouched. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** Zero mutation — `Rank`/`ApplyBudget` read only their own in-memory arguments; no filesystem or index access at all. |
| III. Filesystem Is Single Source of Truth | **Pass (N/A, trivially).** Neither function touches the filesystem or the disposable index — the strongest possible form of "the index is never a correctness dependency," since this feature depends on nothing external at all. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised repeatedly.** No new subpackage (research.md #1); two small functions mapped directly onto two user stories rather than one speculative combined API (research.md #2); the default budget stays an internal constant, not a new `.misterspec/config.yaml` field (research.md #7); the budget-fill algorithm is a simple hard-stop greedy walk, not a knapsack optimizer, because nothing requires optimal packing (research.md #9); `Request.Budget` uses a plain `*int` rather than a new wrapper type (research.md #6). |
| V. Test-First Discipline | **Gate carried into tasks.** §29.7's own Budget test list is the explicit floor; 011-015's full suites are named regression gates. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `ApplyBudget` reuses 015's own unexported `minTier` helper directly rather than reimplementing tier detection (research.md #8); `ResultItem.Tokens` reuses 013's `artifacts.EstimateTokens` verbatim (research.md #12); `ScoredCandidate`/`ResultItem` embed `Candidate`/`ScoredCandidate` rather than duplicating their fields. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes nothing at all — the fifth consecutive read-only capability in this project's own reference/retrieval line (012, 014, 015, now 016), and the first that doesn't even read anything beyond its own arguments. |
| VIII. Safety by Construction | **Pass (N/A).** No write path exists in this feature to make safe or unsafe. |
| IX. Transparent, Machine-Readable Contracts | **Pass (N/A this feature).** No CLI command is added — Phase 8's own "internal context" command is the eventual, later machine-readable surface `Rank`/`ApplyBudget` feed into. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/016-ranking-budgeting/
├── plan.md                      # This file (/speckit-plan command output)
├── research.md                  # Phase 0 output (/speckit-plan command)
├── data-model.md                # Phase 1 output (/speckit-plan command)
├── quickstart.md                # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── ranking-budgeting.md     # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md          # /speckit-specify quality checklist
└── tasks.md                     # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # unchanged — no new dependency
├── cmd/misterspec/, internal/cli/         # unchanged — no CLI surface added
├── internal/
│   ├── project/, ids/, testutil/, example/   # example/ extended (+1 quickstart test)
│   ├── artifacts/, validation/, operations/    # unchanged — reused, not modified
│   └── context/
│       ├── index/                                # unchanged (014) — untouched by this feature
│       ├── request.go                              # MODIFIED — +Budget *int field
│       ├── request_test.go                          # MODIFIED — +coverage for the new field's nil/explicit semantics
│       ├── result.go, collector.go                    # unchanged (015) — reused, not modified
│       ├── rank.go                                       # NEW — ScoredCandidate, Rank, DefaultBudget
│       ├── rank_test.go                                   # NEW
│       ├── budget.go                                        # NEW — ResultItem, Diagnostics, Result, ApplyBudget
│       └── budget_test.go                                    # NEW
└── kit/                                    # unchanged
```

**Structure Decision**: `internal/context` gains two new files and one
additive field on an already-merged type — no new package, the
smallest possible Go footprint for a feature that composes entirely
in-memory over 015's own output, with zero I/O of its own.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
