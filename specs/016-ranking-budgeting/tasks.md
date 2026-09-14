---

description: "Task list template for feature implementation"
---

# Tasks: Ranking and Budgeting

**Input**: Design documents from `/specs/016-ranking-budgeting/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ranking-budgeting.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 011-wikilink-foundation's, 012-references-backlinks's, 013-document-model-chunking's, 014-sqlite-index's, and 015-context-collector's full test suites are explicit, named regression gates (Polish), since this feature adds no call site into any of them requiring a change beyond `Request`'s own additive `Budget` field.

**Organization**: Tasks are grouped by user story. This feature's story chain is **strictly sequential**, following spec.md's own explicit "Why this priority" dependency chain: **User Story 1** (`Rank`) must exist and be trustworthy before **User Story 2** (`ApplyBudget`'s core fitting) can decide what to cut based on that order; **User Story 3** (oversized-mandatory flagging) and **User Story 4** (diagnostics) are both, like 011's own User Story 3 and 015's own User Story 4, *proving/regression guarantees* about behavior `ApplyBudget`'s one coherent algorithm already delivers once User Story 2 is correctly implemented — `BudgetExceeded`/`Overage` and full `Diagnostics` are inseparable parts of the same selection loop (unlike 015's genuinely separable per-tier collection steps), so splitting them into artificial half-working intermediate implementations would misrepresent the actual code shape. Each has a dedicated proving-test task but no separate implementation task of its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US4)
- Every task names its exact file path

## Path Conventions

```text
internal/context/            # existing package (015's own contextengine) — gains rank.go, budget.go here
internal/example/              # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package and no new dependency (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Stand up the one piece of shared plumbing every later story needs but no single story "owns" as its own value proposition: resolving an optional `Request.Budget` into a concrete token ceiling, distinguishing "not specified" from an explicit zero or negative value (research.md #6, #7). `Rank` (User Story 1) does not consume this at all; it is prerequisite specifically for User Story 2 onward.

**⚠️ CRITICAL**: No User Story 2+ implementation may begin until this phase's regression checkpoint (T003) passes. User Story 1 has no dependency on this phase's own output but still waits for it, per this project's established sequencing convention.

- [X] T001 [P] Unit tests for the budget-resolution helper in new file `internal/context/budget_test.go`: `resolveBudget` returns `DefaultBudget` (6000) when `Request.Budget` is `nil`; returns the exact pointed-to value unchanged — including zero and a negative number — when it is non-nil (research.md #6, #7, edge cases).
- [X] T002 [P] Implement the `Budget *int` field on `Request` in `internal/context/request.go`, and `DefaultBudget`/`resolveBudget` in new file `internal/context/budget.go` (data-model.md, contracts/ranking-budgeting.md). Depends on T001.
- [X] T003 Regression checkpoint: `go build ./...` succeeds with the new field/file in place; `go vet ./internal/context/...` clean; 015's own full suite (`go test ./internal/context/...`) still green (the new `Request` field is additive and zero-value-compatible with every existing caller/test). Depends on T002.

**Checkpoint**: Foundation ready — `Request.Budget` exists with proven resolution semantics. User Story 1 implementation can begin; User Story 2 implementation can begin once User Story 1 also completes.

---

## Phase 3: User Story 1 - Rank Structural Truth Above Textual Coincidence (Priority: P1) 🎯 MVP

**Goal**: Every candidate gets a deterministic `Score`, but ordering is always `Tier` first, absolutely, then `Score` — a strong textual coincidence can never outrank mandatory or structural content, and Intent can only reorder candidates already sharing a Tier.

**Independent Test**: Build a candidate set containing mandatory content, a structural connection, and a strong but purely textual match; confirm mandatory and structural both rank above the text match regardless of its own score, and that two same-tier candidates order by relevance, deterministically.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T005.

- [X] T004 [P] [US1] Unit tests for `Rank` in new file `internal/context/rank_test.go`: a mandatory candidate ranks above a structural candidate, which ranks above a candidate with only a strong textual match, regardless of the text match's own score magnitude (§17.1's "important ranking invariant," FR-002); within one shared Tier, a candidate with a stronger relation/text signal ranks first (Acceptance Scenario 2); two candidates with genuinely identical relevance still produce a stable, deterministic order via `Path`/`StartLine`, never an arbitrary one (edge case, FR-011); an `Intent` whose preferred relation exists only in a lower Tier never lets that Tier outrank a higher one — Intent only reorders within a Tier (FR-003, Acceptance Scenario 3).

### Implementation for User Story 1

- [X] T005 [US1] Implement `ScoredCandidate` and `Rank`, plus the unexported `relationWeight`, `textRelevance`, `intentWeight`, `scoreCandidate`, and `queryTerms` helpers, in new file `internal/context/rank.go` (research.md #3-#5, data-model.md, contracts/ranking-budgeting.md). Depends on T004, T003 (Foundational).

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/context/... -run TestRank` passes on its own, with zero dependency on budgeting.

---

## Phase 4: User Story 2 - Fit Within a Token Budget Without Losing What Matters (Priority: P2)

**Goal**: `ApplyBudget` returns the smallest ranked set of candidates whose combined estimated token cost fits within the resolved budget, filling tier by tier — mandatory content always included in full, optional content trimmed from the lowest priority tier first, a hard stop (not skip-ahead) at the first optional item that does not fit (research.md #9).

**Independent Test**: Build a candidate set larger than a given budget, spanning several priority tiers; confirm the result fits within the budget, with lower-priority tiers trimmed before any higher-priority content is touched.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T007.

- [X] T006 [P] [US2] Unit tests for `ApplyBudget`'s core fitting behavior in new file `internal/context/budget_test.go` (alongside T001's own tests): a ranked set larger than the budget returns a result whose combined estimated token cost fits within it, with content from a lower-priority tier trimmed before any higher-priority tier's own content is touched (Acceptance Scenario 2, FR-005); a set that already fits comfortably within the budget has nothing removed at all (Acceptance Scenario 3); no `Budget` specified (`nil`) resolves to `DefaultBudget` (Acceptance Scenario 4, FR-004); `Result.Items` preserves `ranked`'s own order exactly.

### Implementation for User Story 2

- [X] T007 [US2] Implement `ResultItem`, `Diagnostics`, `Result`, and `ApplyBudget` in `internal/context/budget.go` (extending T002's own file): resolve the budget via `resolveBudget`, always include every mandatory (`minTier(Reasons) == TierMandatory`, reusing 015's own helper) candidate in full, then include optional candidates in `ranked` order while they fit, stopping entirely at the first one that does not (research.md #9); compute `Diagnostics` (candidates considered, items selected, tokens available/selected/excluded, reduction percent) and `BudgetExceeded`/`Overage` (mandatory tokens vs. the resolved budget) as part of the same pass (data-model.md, contracts/ranking-budgeting.md). Depends on T006, T005 (needs User Story 1's own `Rank` output).

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/context/... -run TestApplyBudget_Fits` passes on its own. (Mandatory-exceeds-budget and full-diagnostics scenarios are proven next, in User Story 3/4 — the implementation above already handles them correctly as one coherent algorithm.)

---

## Phase 5: User Story 3 - Never Silently Drop Mandatory Context (Priority: P3)

**Goal**: Prove that when mandatory content alone exceeds the requested budget, `ApplyBudget` (already fully implemented in User Story 2) still returns all of it in full, flags the result, reports the exact overage, and remains free to omit every optional item.

**Independent Test**: Construct mandatory content whose own estimated cost alone exceeds a deliberately small budget; confirm the full mandatory content is still returned, the result is flagged as exceeded, and the estimated overage is reported.

### Tests for User Story 3

> This story is a regression/correctness guarantee about User Story 2's own already-implemented `ApplyBudget` (mirroring 015's own User Story 4 precedent) — no new implementation task of its own.

- [X] T008 [US3] Dedicated tests in `internal/context/budget_test.go` directly proving spec.md's own User Story 3 acceptance scenarios end-to-end through `ApplyBudget`: mandatory content whose own estimated cost exceeds the requested budget is still returned in full (Acceptance Scenario 1, FR-007); the result is flagged `BudgetExceeded == true` with `Overage` reporting the exact estimated excess (Acceptance Scenario 2, FR-007); every optional item remains eligible to be omitted in that scenario — none is force-included (Acceptance Scenario 3, FR-008); a requested budget of zero, or a negative value, is treated as an explicit, extremely small budget — mandatory content is still returned in full and flagged as exceeding it, never silently promoted to `DefaultBudget` (edge case, research.md #6). Depends on T007.

**Checkpoint**: User Story 3 is independently complete and testable — `go test ./internal/context/... -run TestApplyBudget_Mandatory` passes on its own, proving a guarantee that already held once User Story 2 existed.

---

## Phase 6: User Story 4 - Explain Every Decision (Priority: P4)

**Goal**: Prove that `ApplyBudget`'s own `Diagnostics` (already computed in User Story 2's implementation) are verifiably accurate against the actual result, for any request.

**Independent Test**: Run a request against a known candidate set and a known budget; confirm the reported counts and token estimates are verifiably accurate against the actual result produced.

### Tests for User Story 4

> Like User Story 3, this story is a regression/correctness guarantee about User Story 2's own already-implemented `ApplyBudget` — no new implementation task of its own.

- [X] T009 [US4] Dedicated tests in `internal/context/budget_test.go` directly proving spec.md's own User Story 4 acceptance scenarios end-to-end through `ApplyBudget`: `Diagnostics.CandidatesConsidered`/`ItemsSelected` match `len(ranked)`/`len(Result.Items)` exactly (Acceptance Scenario 1, FR-010); `TokensAvailable - TokensExcluded == TokensSelected` exactly, and `ReductionPercent` is consistent with those three figures, including the `TokensAvailable == 0` edge case (0%, never a division panic) (Acceptance Scenario 2, FR-010, SC-004); every item in `Result.Items` still carries its own non-empty `Reasons` (inherited from `Candidate`, per FR-009) and its own `Tokens` (Acceptance Scenario 3); a candidate set consisting only of mandatory content, well within an ample budget, reports zero exclusion and zero reduction (edge case). Depends on T007.

**Checkpoint**: User Story 4 is independently complete and testable — `go test ./internal/context/... -run TestApplyBudget_Diagnostics` passes on its own, proving a guarantee that already held once User Story 2 existed.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all four stories, no new capability.

- [X] T010 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 016-ranking-budgeting's packages included) and fix any findings.
- [X] T011 [P] Verify/extend `internal/context`'s own package-level doc comment, cross-checked against `specs/016-ranking-budgeting/contracts/ranking-budgeting.md`.
- [X] T012 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: ranking a candidate set (mandatory/structural/semantic never outranked by a strong text match), fitting it into an explicit budget, using the default budget when none is specified, an oversized-mandatory scenario correctly flagged with its overage, and a determinism check (two `ApplyBudget` calls on the same inputs produce identical results) — each matching `quickstart.md`'s own documented result exactly.
- [X] T013 Reconcile `specs/016-ranking-budgeting/contracts/ranking-budgeting.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T014 Full regression run: `go test ./...` across the entire module (001 through 016) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/context/... -race` clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS User Story 2 onward (budget resolution semantics). Does not block User Story 1.
- **User Story 1 (Phase 3)**: Depends on Foundational only by this project's sequencing convention, not by actual code dependency.
- **User Story 2 (Phase 4)**: Depends on User Story 1's own `Rank` output and on Foundational's `resolveBudget`.
- **User Story 3 (Phase 5)**: Depends on User Story 2's own `ApplyBudget` already existing — nothing to prove exceeded until budgeting itself works.
- **User Story 4 (Phase 6)**: Depends on User Story 2's own `ApplyBudget` already existing — nothing meaningful to report before ranking and budgeting actually happen.
- **Polish (Phase 7)**: Depends on all four user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational (no real code dependency on it).
- **User Story 2 (P2)**: Depends on User Story 1 (ranking must be trustworthy before budgeting decides what to cut) and Foundational.
- **User Story 3 (P3)**: Depends on User Story 2 (proving-only, no new implementation).
- **User Story 4 (P4)**: Depends on User Story 2 (proving-only, no new implementation) — independent of User Story 3.

This feature's chain (Foundational → US1 → US2 → [US3 proving] /
[US4 proving]) is strictly sequential through US2, then fans out into
two independent proving phases — the same shape 011's own User Story 3
and 015's own User Story 4 took: a dedicated test suite proving a
guarantee the prior story's own implementation already delivers, not a
new capability of its own.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V) — except User Story 3 and User Story 4, regression/correctness guarantees about already-implemented behavior, matching 011's and 015's own precedent.
- Foundational's regression checkpoint (T003) passes before User Story 2's implementation begins.

### Parallel Opportunities

- T001 (Foundational's test) can be written in parallel with T004 (User Story 1's test) — different files, no dependency between them.
- T008 and T009 (User Story 3's and User Story 4's own proving-test tasks) can be written in parallel with each other once T007 exists — both depend only on T007, not on one another.
- Within Polish: T010 and T011 in parallel.

---

## Parallel Example: Foundational + User Story 1

```bash
# These two can be authored together — no shared file, no dependency:
Task: "Unit tests for the budget-resolution helper in internal/context/budget_test.go"
Task: "Unit tests for Rank in internal/context/rank_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (`Request.Budget`, `DefaultBudget`, `resolveBudget`, proven).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/context/... -run TestRank` green, independently.
5. This alone already proves the one guarantee nothing else in this feature may ever violate: mandatory and structural content can never be outranked by a lucky text match.

### Incremental Delivery

1. Setup (nothing) + Foundational (budget resolution, proven).
2. Add User Story 1 → validate independently → trustworthy ordering usable (MVP).
3. Add User Story 2 (extends on US1's output) → validate independently → budget-fitting usable.
4. Add User Story 3 (proves US2's own oversized-mandatory behavior) → validate independently → the "never silently drop mandatory" guarantee formally proven.
5. Add User Story 4 (proves US2's own diagnostics) → validate independently → measurable token-reduction reporting formally proven.
6. Polish (Phase 7), including the full-module `-race` regression run (T014).

### Team Strategy

Like 011's and 015's own chains, User Story 3 and User Story 4 are
independent of each other once User Story 2 exists — two developers
could write their proving tests in parallel — but the chain up through
User Story 2 is strictly sequential, the natural shape for a single
developer working through it in order.

---

## Notes

- [P] tasks touch different files, or the same file with no dependency on an incomplete task's own content.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement — except User Story 3/4's own proving tests, which pass immediately once User Story 2's implementation is correct.
- `Rank`/`ApplyBudget` are strictly read-only (FR-012) — no mutation of any project artifact, the reference graph, or the search index, anywhere in this feature; neither performs any I/O of its own.
- `Tier` always dominates `Score` in `Rank`'s own ordering (FR-002) — this must never be violated by a future weight tweak; T004's own tests are the permanent guard against that.
- No CLI-layer file is touched anywhere in this feature (research.md) — Phase 8's own "internal context" command is the later, actual CLI surface `Rank`/`ApplyBudget` feed into.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
