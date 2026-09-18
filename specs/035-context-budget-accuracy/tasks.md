---

description: "Task list for Orçamento sobre a saída efetiva de contexto (PROP-05)"
---

# Tasks: Orçamento sobre a saída efetiva de contexto

**Input**: Design documents from `/specs/035-context-budget-accuracy/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/budget-and-estimator-contract.md, quickstart.md

**Tests**: Required for every user story per Constitution Principle V (Test-First, NON-NEGOTIABLE) — write each test, confirm it fails (`go test`/`go vet`), then implement until green.

**Organization**: Tasks are grouped by user story (US1/US2/US3, matching spec.md's priorities P1/P1/P2) so each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1, US2, or US3 — omitted for Setup/Foundational/Polish tasks

## Path Conventions

Single Go module. All paths are relative to the repository root (`/home/marciovcm/workspace/golang/misterspec`).

---

## Phase 1: Setup

**Purpose**: Confirm the branch starts from a green baseline before any change.

- [X] T001 Run `go build ./... && go vet ./... && go test ./...` on branch `035-context-budget-accuracy` and confirm it is green before making any change

**Checkpoint**: Baseline confirmed green.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Thread an injectable `artifacts.Estimator` through `ApplyBudget`/`PayloadTokens` — every user story's diagnostics and behavior build on this plumbing existing first (research.md Decision 1).

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 [P] Write a failing test in `internal/artifacts/tokens_test.go` asserting `artifacts.Estimator` requires a `Name() string` method and that `DefaultEstimator{}.Name()` returns the non-empty string `"default"` (data-model.md Estimator validation rule: `Name()` MUST be non-empty)
- [X] T003 Add `Name() string` to the `Estimator` interface in `internal/artifacts/tokens.go`; implement `func (DefaultEstimator) Name() string { return "default" }` (makes T002 pass)
- [X] T004 Write a failing test in `internal/context/budget_test.go` asserting `ApplyBudget` accepts an `artifacts.Estimator` parameter and uses it (not `artifacts.EstimateTokens` directly) to compute every `ResultItem.Tokens` and `Diagnostics.TokensAvailable`/`TokensSelected`/`TokensExcluded`
- [X] T005 Change `ApplyBudget(ranked []ScoredCandidate, req Request) Result` to `ApplyBudget(ranked []ScoredCandidate, req Request, estimator artifacts.Estimator) Result` in `internal/context/budget.go`, replacing the direct `artifacts.EstimateTokens(c.Content)` call at line 79 with `estimator.Estimate(c.Content)` (makes T004 pass)
- [X] T006 [P] Change `PayloadTokens(contentTokens int, packageItems []PackageItem, rendered string) int` in `internal/context/pack.go` to take an injected `artifacts.Estimator` parameter in place of its own direct `artifacts.EstimateTokens` calls (lines 92, 100), matching T005's pattern
- [X] T007 Update every existing call site of `ApplyBudget`/`PayloadTokens` in `internal/context/*.go` and `internal/cli/internalcmd/context.go` to construct and pass `artifacts.DefaultEstimator{}` (compile fixups only — no behavior or diagnostics change yet)
- [X] T008 Fix every existing test in `internal/context/*_test.go` and `internal/cli/internalcmd/context_test.go` broken by the new `Estimator` parameter (pass `artifacts.DefaultEstimator{}`), and run `go build ./... && go vet ./... && go test ./...` to confirm green

**Checkpoint**: `Estimator` is threaded through the whole pipeline, but nothing yet reports it or changes behavior. Foundation ready — user stories can now proceed.

---

## Phase 3: User Story 1 - Saber exatamente o que o orçamento está medindo (Priority: P1) 🎯 MVP

**Goal**: Every Context Pack response explicitly identifies which estimator produced its token counts, and those counts correspond to the text actually delivered (spec FR-001, FR-003, FR-004).

**Independent Test**: Request a Context Pack; confirm the response names its estimator (`"default"` today) and that the reported content-token count matches the real size of the delivered text.

### Tests for User Story 1

- [X] T009 [P] [US1] Write a failing test in `internal/context/budget_test.go` asserting `ApplyBudget(...).Diagnostics.Estimator == "default"` when called with `artifacts.DefaultEstimator{}` (spec FR-001)
- [X] T010 [P] [US1] Write a failing test in `internal/context/pack_test.go` asserting `PayloadTokens` computes its result using the injected `Estimator.Estimate` applied to the actually-rendered payload, not a disconnected approximation (spec FR-004)
- [X] T011 [P] [US1] Write a failing test in `internal/cli/internalcmd/context_test.go` asserting `diagnostics.estimator == "default"` is present in the JSON response for all three output modes (`--mode manifest`, `--mode package`, plain Markdown's diagnostics companion)

### Implementation for User Story 1

- [X] T012 [US1] Add `Estimator string` field to `Diagnostics` in `internal/context/budget.go`; populate it from `estimator.Name()` inside `ApplyBudget` (makes T009 pass)
- [X] T013 [US1] Verify/adjust `PayloadTokens` in `internal/context/pack.go` so every token count it returns is computed exclusively through the injected `Estimator.Estimate` — no residual direct `artifacts.EstimateTokens` call remains anywhere in `internal/context` (makes T010 pass; spec FR-004)
- [X] T014 [US1] Surface `diagnostics.estimator` in `internal/cli/internalcmd/context.go`'s JSON rendering (`renderContextItems`/`renderPackageItems` and the manifest-mode path) for all three output modes (makes T011 pass)
- [X] T015 [US1] Run `go build ./... && go vet ./... && go test ./...` and confirm green

**Checkpoint**: User Story 1 is independently functional — every response names its estimator and reports correspondence-verified counts.

---

## Phase 4: User Story 2 - Comportamento previsível quando o conteúdo obrigatório não cabe (Priority: P1)

**Goal**: A separate, explicit hard limit governs whether mandatory content is reported as "exceeded" — decoupled from the soft budget that only shapes optional content (spec FR-005–FR-007, FR-006 default; research.md Decision 2/3).

**Independent Test**: Request a Context Pack with a small soft budget but mandatory content under the hard limit — confirm no exceeded flag. Then request with a hard limit below mandatory content — confirm an explicit diagnostic, with mandatory content still delivered whole.

### Tests for User Story 2

- [X] T016 [P] [US2] Write a failing test in `internal/context/request_test.go` asserting a new `Request.HardLimit *int` field exists and, when `nil`, resolves to `DefaultHardLimit` (spec FR-006)
- [X] T017 [P] [US2] Write a failing test in `internal/context/budget_test.go` asserting `DefaultHardLimit == 2 * DefaultBudget` (`12000`) (research.md Decision 2)
- [X] T018 [P] [US2] Write a failing test in `internal/context/budget_test.go` asserting: mandatory content greater than the soft budget but less than the resolved hard limit → `Result.BudgetExceeded == false` and every mandatory item still appears whole in `Result.Items` (spec Acceptance Scenario 1; SC-002)
- [X] T019 [P] [US2] Write a failing test in `internal/context/budget_test.go` asserting: mandatory content greater than the resolved hard limit → `Result.BudgetExceeded == true`, `Result.Overage == mandatoryTokens - hardLimit`, and every mandatory item still appears whole in `Result.Items` — never truncated (spec FR-007, Acceptance Scenario 2; SC-003)
- [X] T020 [P] [US2] Write a failing test in `internal/context/request_test.go` (or `budget_test.go`) asserting a `HardLimit` smaller than the resolved `Budget` is accepted, not rejected, and produces the expected tight-ceiling behavior (data-model.md Request validation rule; spec Edge Cases)
- [X] T021 [P] [US2] Write a failing test in `internal/cli/internalcmd/context_test.go` asserting a `--hard-limit N` flag parses into `Request.HardLimit`, and that omitting it reports `diagnostics.hard_limit == 12000` (spec FR-006; contracts §1)
- [X] T022 [P] [US2] Write a failing test in `internal/cli/internalcmd/context_test.go` asserting `schema_version == 2` is reported in every `internal context` response, replacing 033's `1` (contracts §2)

### Implementation for User Story 2

- [X] T023 [US2] Add `HardLimit *int` field to `Request` in `internal/context/request.go`, documented the same way `Budget` already is (makes T016 pass)
- [X] T024 [US2] Add `const DefaultHardLimit = 2 * DefaultBudget` to `internal/context/budget.go`, and a `resolveHardLimit(req Request) int` helper mirroring `resolveBudget` (makes T017 pass)
- [X] T025 [US2] Add `HardLimit int` field to `Diagnostics` in `internal/context/budget.go`, always populated with the resolved value from `resolveHardLimit` (contracts §2.2)
- [X] T026 [US2] Change `ApplyBudget`'s `exceeded`/`overage` computation in `internal/context/budget.go` to compare `mandatoryTokens` against the resolved hard limit instead of `budget` (the soft budget) — `exceeded := mandatoryTokens > hardLimit`, `overage := mandatoryTokens - hardLimit` when exceeded (makes T018/T019/T020 pass; research.md Decision 3)
- [X] T027 [US2] Add a `--hard-limit` int flag to the Cobra command in `internal/cli/internalcmd/context.go`, wired into `Request.HardLimit`; reject a non-integer with `invalid_argument` (makes T021 pass; contracts §1)
- [X] T028 [US2] Bump the `contextSchemaVersion` constant from `1` to `2` in `internal/cli/internalcmd/context.go` (makes T022 pass; contracts §2)
- [X] T029 [US2] Run `go build ./... && go vet ./... && go test ./...` and confirm green, including every pre-existing test asserting the old `schema_version: 1`/soft-budget-based `budget_exceeded` behavior updated to the new contract

**Checkpoint**: User Stories 1 AND 2 both work independently — estimator identification plus decoupled soft/hard limits.

---

## Phase 5: User Story 3 - Aproveitar melhor o espaço disponível sem violar prioridades (Priority: P2)

**Goal**: Backfill keeps trying smaller same-tier optional candidates instead of stopping at the first miss, and an oversized single candidate falls back to a coherent-unit prefix instead of being excluded outright — both fully deterministic, both never violating tier order (spec FR-008–FR-013; research.md Decisions 4–6).

**Independent Test**: Build a Context Pack where the first optional candidate after mandatory content doesn't fit but a smaller same-tier one later does — confirm the smaller one is included and the larger one is recorded as excluded. Separately, build one with an oversized single section and confirm a coherent unit prefix is considered instead of an all-or-nothing exclusion.

### Tests for User Story 3

- [X] T030 [P] [US3] Write a failing test in `internal/artifacts/split_test.go` asserting `SplitIntoCoherentUnits(content string) []string` never splits inside a fenced code block (a fence's opening/closing pair and everything between stay in one unit), and that concatenating all returned units, in order, reproduces `content` exactly (data-model.md CoherentUnit validation rules; spec FR-011)
- [X] T031 [P] [US3] Write a failing test in `internal/context/budget_test.go` asserting: when an earlier optional candidate doesn't fit the remaining soft-budget space, a smaller later candidate of the *same* tier is still evaluated and included, and the skipped larger candidate appears in `Result.Diagnostics.Exclusions` with `Reason == "did_not_fit_remaining_budget"` (spec FR-008, FR-010, Acceptance Scenario 1; SC-004)
- [X] T032 [P] [US3] Write a failing test in `internal/context/budget_test.go` asserting tier order is never violated by backfill — no candidate from a lower-priority tier appears in `Result.Items` while any higher-priority-tier candidate remains undecided (spec FR-009, Acceptance Scenario 2)
- [X] T033 [P] [US3] Write a failing test in `internal/context/budget_test.go` asserting an oversized single optional candidate whose first coherent unit fits is included as the longest fitting prefix of its own coherent units (not excluded whole) (spec FR-011, Acceptance Scenario 3)
- [X] T034 [P] [US3] Write a failing test in `internal/context/budget_test.go` asserting an oversized single optional candidate whose *first* coherent unit does not even fit is excluded whole, recorded in `Result.Diagnostics.Exclusions` with `Reason == "no_coherent_unit_fit"` — never included partially in a way that breaks its own meaning (spec FR-012, Acceptance Scenario 4)
- [X] T035 [P] [US3] Write a failing test in `internal/context/budget_test.go` asserting `ApplyBudget` is deterministic — two calls with identical `ranked`/`req`/`estimator` arguments produce byte-identical `Result` values (spec FR-013; SC-005)

### Implementation for User Story 3

- [X] T036 [US3] Implement `SplitIntoCoherentUnits(content string) []string` in new file `internal/artifacts/split.go`, reusing the existing unexported `fencedLines` helper from `internal/artifacts/wikilink.go` to keep a fenced block's opening/closing pair and everything between in one unit; paragraph-bounded elsewhere (makes T030 pass; research.md Decision 5)
- [X] T037 [US3] Add `ExclusionRecord{Path, Heading, Reason string}` type and `Exclusions []ExclusionRecord` field to `Diagnostics` in `internal/context/budget.go` (data-model.md ExclusionRecord)
- [X] T038 [US3] Remove the `stopped` bool from `ApplyBudget`'s optional-content loop in `internal/context/budget.go`; when a candidate doesn't fit the remaining budget, append an `ExclusionRecord{Reason: "did_not_fit_remaining_budget"}` and `continue` to the next candidate in `ranked`'s existing tier-ascending/score-descending order instead of breaking (makes T031/T032 pass; research.md Decision 4)
- [X] T039 [US3] Add the coherent-split fallback inside `ApplyBudget`'s optional-content loop in `internal/context/budget.go`: when a whole candidate doesn't fit, call `artifacts.SplitIntoCoherentUnits(c.Content)` and include the longest fitting prefix of units as a `ResultItem`; if even the first unit doesn't fit, append an `ExclusionRecord{Reason: "no_coherent_unit_fit"}` and exclude the candidate whole (makes T033/T034 pass; research.md Decision 5)
- [X] T040 [US3] Verify `ApplyBudget`'s new backfill/split logic introduces no non-determinism (no map iteration, no time/randomness) — confirm T035 passes as-is or fix any ordering issue found
- [X] T041 [US3] Surface `diagnostics.exclusions` in `internal/cli/internalcmd/context.go`'s JSON rendering for all three output modes (contracts §2.2)
- [X] T042 [US3] Run `go build ./... && go vet ./... && go test ./...` and confirm green

**Checkpoint**: All three user stories are independently functional — estimator identification, decoupled limits, and backfill/coherent-splitting all behave per spec.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Whole-feature validation and documentation, spanning all three user stories.

- [X] T043 Manually run every scenario in `specs/035-context-budget-accuracy/quickstart.md` against a built binary (`go build -o /tmp/ms-035-bin ./cmd/...` or the project's existing build target) in a disposable temp project fixture, then clean up
- [X] T044 [P] Update `site/commands.html` with the new `--hard-limit` flag, the new diagnostics fields (`estimator`, `hard_limit`, `exclusions`), and the `schema_version: 2` bump (matching the pattern established for specs 031–034)
- [X] T045 Run the whole-repo gate one final time: `go build ./... && go vet ./... && go test ./...`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (the `Estimator` parameter signature change in `ApplyBudget`/`PayloadTokens` is shared by every later test).
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on Foundational only — touches the same `ApplyBudget` function as US1 but different fields (`HardLimit`/`BudgetExceeded`/`Overage` vs `Diagnostics.Estimator`), so implement sequentially to avoid the same-file conflict, though the two stories' own test/behavior scope is independent.
- **User Story 3 (Phase 5)**: Depends on Foundational only in principle, but shares `ApplyBudget`'s loop body with US2's hard-limit computation (same function, same file) — implement after US2 to avoid rebasing the loop twice.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### Within Each User Story

- Tests are written first and confirmed to fail before implementation (Constitution Principle V).
- Implementation tasks within a story generally touch the same few files sequentially (`budget.go`'s loop is shared); tasks marked `[P]` are on genuinely different files (e.g., `split.go` vs `budget.go`) or independent test files.

### Parallel Opportunities

- T002/T004/T006 (foundational tests, different files) can run in parallel.
- T009/T010/T011 (US1 tests, different files) can run in parallel.
- T016–T022 (US2 tests, different files) can run in parallel.
- T030–T035 (US3 tests, different files) can run in parallel.
- T043/T044 (Polish) can run in parallel.

---

## Parallel Example: User Story 3 tests

```bash
# Launch all US3 tests together (different files, no shared state):
Task: "Write failing test in internal/artifacts/split_test.go for fence-safe splitting"
Task: "Write failing test in internal/context/budget_test.go for backfill inclusion"
Task: "Write failing test in internal/context/budget_test.go for tier-order preservation"
Task: "Write failing test in internal/context/budget_test.go for coherent-unit fallback inclusion"
Task: "Write failing test in internal/context/budget_test.go for whole-exclusion when no unit fits"
Task: "Write failing test in internal/context/budget_test.go for determinism"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (blocks everything)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: run `go test ./internal/context/... ./internal/cli/internalcmd/...` and confirm every response names its estimator
5. Demo if ready — this alone already resolves the "unverifiable budget promise" problem

### Incremental Delivery

1. Setup + Foundational → estimator threaded, nothing yet behaves differently
2. Add User Story 1 → estimator identification lands (MVP)
3. Add User Story 2 → soft/hard limit split lands, `schema_version: 2`
4. Add User Story 3 → backfill + coherent splitting land
5. Each story adds value without breaking the previous one — `ApplyBudget`'s public shape only ever grows additively (aside from the one deliberate `BudgetExceeded`/`Overage` redefinition in US2, versioned)

---

## Notes

- `[P]` tasks = different files, no dependencies
- `[Story]` label maps a task to US1/US2/US3 for traceability
- Verify each test fails before implementing (Constitution Principle V, NON-NEGOTIABLE)
- Run `go build ./... && go vet ./... && go test ./...` after every phase, not just at the end
- Avoid: vague tasks, same-file conflicts within a story, cross-story dependencies that break independence
