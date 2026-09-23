# Phase 0 Research: Orçamento sobre a saída efetiva de contexto

No `[NEEDS CLARIFICATION]` markers were left in the Technical Context. This document records the design decisions the current code's actual behavior forces, verified against `internal/artifacts/tokens.go`, `internal/context/budget.go`/`pack.go`/`request.go`, and `016-ranking-budgeting/research.md` rather than assumed from the backlog document alone.

## Decision 1 — `Estimator` already exists; this feature finally wires it in and names it

**Decision**: Add `Name() string` to the existing `artifacts.Estimator` interface (`DefaultEstimator.Name()` returns `"default"`); change `ApplyBudget` and `PayloadTokens` to accept an `Estimator` parameter instead of calling the free function `artifacts.EstimateTokens` directly; the CLI layer (`internal/cli/internalcmd/context.go`) constructs the estimator (today, always `artifacts.DefaultEstimator{}` — spec Assumptions explicitly do not require a real tokenizer in this feature) and reports its `Name()` in the response.

**Rationale**: Verified `internal/artifacts/tokens.go:3-9` already declares `Estimator` with exactly this stated purpose ("kept as an interface so a real provider-specific tokenizer can be substituted later"), but `budget.go:79` and `pack.go:92,100` call `artifacts.EstimateTokens` directly, never through the interface — the interface has been dead code since it was introduced. Adding one method to a two-line interface with exactly one implementation in the whole codebase is the minimal change that makes it load-bearing.

**Alternatives considered**: A separate `EstimatorName` string threaded alongside the `Estimator` value at every call site — rejected: two values that must always agree (which estimator vs. what to call it) is a needless place for drift; one interface method keeps the name and the behavior on the same value.

## Decision 2 — Soft limit (existing `Request.Budget`) and a new, separate `Request.HardLimit`

**Decision**: `Request.Budget` keeps its existing meaning and field name (the optional-content target — the "soft limit"). A new `Request.HardLimit *int` field is added: the ceiling above which mandatory content is reported as exceeded. When `HardLimit` is nil, it resolves to a new `DefaultHardLimit` constant, defined as a fixed multiple of `DefaultBudget` (`2 * DefaultBudget` = 12000) — large enough to not fire in common cases, but always finite (spec FR-006, Assumptions).

**Rationale**: Verified today's `ApplyBudget` (`budget.go:100-101`) computes `exceeded := mandatoryTokens > budget` — the *same* `budget` value `resolveBudget` returns for the optional-content loop, confirmed by `016-ranking-budgeting/research.md`'s own Decision #9 discussing only one budget concept throughout. Spec User Story 2 explicitly requires decoupling these two roles; a fixed default multiple (rather than "no limit," `nil`-means-unbounded) keeps FR-006's "never treat absence as no limit" requirement mechanically simple — a plain `int`, always compared, never a special-cased nil branch deeper in the budget logic.

**Alternatives considered**: Making the hard limit unbounded by default (only enforced when the caller explicitly sets it) — rejected: spec FR-006 explicitly forbids treating its absence as "no limit," and an always-finite default keeps the comparison logic uniform (no nil-check branch inside `ApplyBudget`).

## Decision 3 — `budget_exceeded`/`overage` change meaning; versioned via `schema_version: 2`

**Decision**: `Result.BudgetExceeded`/`Overage` now compare `mandatoryTokens` against the resolved **hard limit**, not the soft budget. This is a deliberate redefinition of an existing field's meaning (not merely an additive field), the one intentional exception to 033's own "no behavior change to existing fields, only additive ones" discipline. `internal/cli/internalcmd/context.go`'s `schema_version` constant is bumped from `1` to `2`.

**Rationale**: This is precisely what `schema_version` exists for (033/research.md Decision 5: "a consumer branches on... the whole shape") — a real, spec-mandated semantic change to a field name that stays the same. Silently keeping the field's old comparison target while adding a new, separate hard-limit-exceeded flag was considered and rejected below; the field's own name (`budget_exceeded`) already promises "did the budget get exceeded," and after this feature the answer to that question is properly "did the thing that actually matters (mandatory content vs. the real ceiling) get exceeded" — keeping the old, conflated answer under the same name would be the more confusing outcome, not the safer one.

**Alternatives considered**: Keep `budget_exceeded` comparing against the soft budget (old behavior, byte-compatible) and add a second field (e.g. `hard_limit_exceeded`) for the new concept — rejected: this leaves the misleading signal spec User Story 2 exists to fix fully intact under its original name, satisfying compatibility at the cost of leaving the actual reported bug in place; the spec's own Acceptance Scenario 1 requires the *old* false-alarm behavior to stop, not to persist alongside a new correct one.

## Decision 4 — Backfill: delete the "stop at first miss" flag, keep the existing tier-then-score order

**Decision**: `ApplyBudget`'s optional-content loop removes the `stopped` bool entirely. Each optional candidate (still visited in `ranked`'s own existing tier-ascending, score-descending order — no new sort) that doesn't fit in the remaining soft-budget space is recorded as an `Exclusion` and skipped; the loop continues to the *next* candidate in that same order (which may be a smaller item of the same tier) instead of abandoning the rest of the list.

**Rationale**: Verified `budget.go`'s existing loop already visits `ranked` in exactly the order `Rank`/`mergeAndSort` already established (tier ascending, then score descending within a tier, per `016`/`015`'s own research) — removing the early-exit flag, with no other change, already guarantees FR-009 ("never let a lower-tier candidate jump ahead of an undecided higher-tier one") for free: the list itself never reorders, so a lower-tier item is only ever reached after every higher-tier item has already been decided (included or excluded) in its own turn. This is a genuinely minimal change — one flag removed — not a new bin-packing search.

**Alternatives considered**: A true knapsack/best-fit search over remaining optional candidates — rejected exactly as `016-ranking-budgeting/research.md`'s own Decision #9 already reasoned: it introduces its own tie-breaking subtleties and is not what spec FR-008 asks for (which only asks to keep *trying* subsequent same-tier candidates, not to find the provably optimal packing).

## Decision 5 — Coherent-unit splitting lives in `internal/artifacts`, reusing the existing fence tracker

**Decision**: A new function `artifacts.SplitIntoCoherentUnits(content string) []string` splits a Chunk's own body into paragraph-bounded units, using the same unexported `fencedLines` helper `ParseDocument`/`ExtractWikiLinks` already use — a unit never splits in the middle of a fenced code block (a fence's opening/closing pair and everything between travel together as one unit). `ApplyBudget` calls this only as a fallback: when a whole optional Chunk doesn't fit in the remaining soft-budget space, it tries including the longest fitting *prefix* of that Chunk's own coherent units (in order) instead of excluding the Chunk outright; if not even the first unit fits, the whole Chunk is excluded (spec FR-012) — never a partial unit.

**Rationale**: `internal/artifacts` already owns exactly this kind of Markdown-structure-awareness (`ParseDocument`, `Chunks`) and already has the fence-tracking primitive this needs — adding one more small, pure function there is simpler than writing a second fence-tracker inside `internal/context` (Constitution Principle VI, DRY). Scoping this to a *fallback* (only tried when the whole Chunk doesn't fit) keeps the common case (most Chunks fit whole) unaffected — no new behavior for the vast majority of requests.

**Alternatives considered**: Splitting at the sentence level for finer-grained partial inclusion — rejected as unnecessary precision spec.md doesn't ask for (FR-011 only names "a code block or a requirement" as the indivisible units to preserve, both already paragraph-or-larger-sized in practice); paragraph-level splitting already satisfies that bar without inventing a sentence tokenizer.

## Decision 6 — A `Requirement` section is already atomic by construction; no extra logic needed for it specifically

**Decision**: No special-case code treats "a Requirement" as its own indivisible unit distinct from ordinary paragraph/code-fence handling. A Spec's own `### R<N>` section is already exactly one `Chunk` (013's own per-heading chunking) — Decision 5's coherent-unit splitting only ever operates *within* one Chunk's own body when that whole Chunk doesn't fit, and a Requirement section's own body is typically small enough, and internally paragraph-structured, that the same fence-aware paragraph splitting already treats it sensibly (a Requirement's own defining sentence stays in one paragraph-unit, never fragmented mid-sentence by this feature's own logic).

**Rationale**: Verified (031/032 research) that Requirement sections are already chunked one-per-heading; nothing in this feature's own scope (budget/estimation) needs a second, Requirement-specific indivisibility rule beyond what Decision 5 already provides at the paragraph/fence level.

**Alternatives considered**: A dedicated "never split a Requirement Chunk at all, even if it means excluding it whole" rule distinct from the general coherent-unit fallback — rejected as an unjustified special case; spec FR-011/FR-012 describe the same fallback-then-exclude behavior for "a code block or a requirement" uniformly, not two different policies.
