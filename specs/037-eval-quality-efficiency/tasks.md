---

description: "Task list for End-to-End Quality and Efficiency Evaluation (037-eval-quality-efficiency)"
---

# Tasks: End-to-End Quality and Efficiency Evaluation

**Input**: Design documents from `/specs/037-eval-quality-efficiency/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/eval-commands-contract.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story and are written before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (US1 = retrieval evaluation, US2 = agent task-execution evaluation, US3 = baseline comparison) to enable independent implementation and testing of each.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single Go project. New package at `internal/eval/`; new CLI commands at
`internal/cli/internalcmd/`; wiring in `internal/cli/internal.go`;
fixture data at `specs/037-eval-quality-efficiency/fixture/`; recorded
run data at the project-level `eval/` directory.

---

## Phase 1: Setup

**Purpose**: Package scaffolding shared by every story.

- [X] T001 Create `internal/eval/doc.go` describing the package's purpose (deterministic retrieval evaluation + comparison of recorded runs; never drives a live LLM session — plan.md Constitution Check, Principle IV) and its relationship to `internal/context`.
- [X] T002 [P] Create `specs/037-eval-quality-efficiency/fixture/` directory tree: `cases/` (empty, populated in US1), `tasks/` (empty, populated in US2), and a minimal multi-Spec fixture project under `fixture/project/` reusing the shape of `internal/context/fixture_test.go`'s existing test fixture (small set of Specs with `depends_on`/wikilinks, per data-model.md `EvaluationCase.dir`).
- [X] T003 [P] Add `eval/` to the project root with `.gitkeep` files under `eval/baselines/.gitkeep` and `eval/runs/.gitkeep` so the directories are tracked before any run is recorded (data-model.md, research.md #3). The `.gitkeep` placeholders were later removed once T040's own committed baselines/runs (`retrieval-2026-09-init`, `task-exec-2026-09-init`) gave both directories real, tracked content.

**Checkpoint**: Package and fixture scaffolding exist; no behavior yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared types every user story's commands read or write. Per data-model.md, `RunRecord` (with its `kind`/`config`/`results` envelope) is used by US1 (produces retrieval `RunRecord`s), US2 (task-execution `RunRecord`s are hand-authored against this same shape), and US3 (consumes both to compare) — so its core shape belongs here, not duplicated per story.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T004 Define core `RunRecord` envelope in `internal/eval/record.go`: `RunID string`, `Kind string` (data-model.md: `"retrieval"` | `"task_execution"`, no other value valid), `CreatedAt time.Time` (serialized RFC 3339), `Config map[string]string` (data-model.md: `{repo_revision, model, variant, repetition_index?}` — every field is a dimension a comparison must check for equality), and `Results json.RawMessage` (deferred typing — US1 unmarshals into `[]CaseResult`, US2 into `[]TaskResult`) plus JSON (de)serialization to/from a file path.
- [X] T005 [P] Add `RunRecord` validation in `internal/eval/record.go` per data-model.md Validation Rules Summary: `Kind` MUST be exactly `"retrieval"` or `"task_execution"` (reject any other value); return a typed error the CLI layer maps to the `invalid_case`/`invalid_argument` sentinels (contracts/eval-commands-contract.md §3).
- [X] T006 [P] [Foundational] Unit tests for T004/T005 in `internal/eval/record_test.go`: round-trip JSON (de)serialization is lossless; an unknown `Kind` value is rejected; `Config` with a missing `repetition_index` serializes without the field (optional, per data-model.md).
- [X] T007 Register the (currently empty) `internal eval-retrieval` and `internal eval-compare` command stubs in `internal/cli/internal.go`'s `cmd.AddCommand(...)` list (constructors added by T014/T024 respectively) so the wiring compiles incrementally; follow the existing flat, hyphenated naming convention already used by e.g. `migration-check-tasks` (internal/cli/internalcmd/migration_check_tasks.go), not a nested `eval` subcommand group.

**Checkpoint**: `internal/eval` package compiles with a shared `RunRecord` type; CLI tree wiring point exists. User story work can now begin.

---

## Phase 3: User Story 1 - Run a Deterministic Retrieval Evaluation (Priority: P1) 🎯 MVP

**Goal**: A maintainer can run a fixed, versioned set of retrieval cases against a known repository state and get a reproducible, case-by-case pass/fail and score report — entirely offline.

**Independent Test**: Run `misterspec internal eval-retrieval --cases specs/037-eval-quality-efficiency/fixture/cases` twice against the same fixture and confirm byte-identical results (quickstart.md §1); confirm a case with a deliberately-missing required item reports the specific missing identifier (quickstart.md §2).

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T008 [P] [US1] Unit test for `EvaluationCase` loading/validation in `internal/eval/case_test.go`: a case with both `query` and `task` set is rejected (data-model.md: "exactly one of `query`/`task` set"); a case with a duplicate `id` within the loaded set is rejected ("`id` unique within the loaded set"); a case whose `dir` does not resolve is rejected ("`dir` MUST resolve to an existing path"); a case with empty `required` and no `note` field is rejected ("`required` non-empty unless the case is explicitly documented... as a negative/empty-result case").
- [X] T009 [P] [US1] Unit test for the retrieval runner in `internal/eval/retrieval_test.go`: given a fixture case whose `required` identifiers are all present in the collector's returned pack, `CaseResult.Passed` is `true` and `Missing`/`Unexpected` are both empty; given a case with one absent required identifier, `Passed` is `false` and `Missing` names exactly that identifier (data-model.md `CaseResult.Missing`); given a `forbidden` identifier present in the pack, `Unexpected` names it and `Passed` is `false`; `Positions` records the 1-based rank of each found required identifier (data-model.md: "for later ranking-position comparisons").
- [X] T010 [P] [US1] Filesystem integration test in `internal/eval/retrieval_test.go` (or a new `internal/eval/fixture_test.go`, following the `internal/context/fixture_test.go` pattern): running the full case set against `specs/037-eval-quality-efficiency/fixture/project/` twice produces identical `RunRecord.Results` except `RunID`/`CreatedAt` (spec SC-001, quickstart.md §1).
- [X] T011 [P] [US1] Golden/contract test in `internal/cli/internalcmd/eval_retrieval_test.go` (skeleton, will fail until T014): `internal eval-retrieval --cases <missing-dir>` returns `invalid_argument` and exit code 2 (contract §1); a well-formed case set returns `schema_version: 1` and the response shape from contract §1's example.

### Implementation for User Story 1

- [X] T012 [P] [US1] Implement `EvaluationCase` type and case-file loader in `internal/eval/case.go`: parses YAML case files (research.md #2) from a directory into `[]EvaluationCase`, applying every validation rule in data-model.md's Validation Rules Summary; returns a structured `invalid_case` error (not a generic error) identifying the offending file and rule.
- [X] T013 [US1] Implement the retrieval runner in `internal/eval/retrieval.go` (depends on T004, T012): for each `EvaluationCase`, build a `contextengine.Request` (query or task/intent mode per the case, budget override if set) and call the existing `contextengine` collector/ranker in-process (no shelling out — plan.md Summary), producing one `CaseResult` per case (`Passed`, `Missing`, `Unexpected`, `Positions`, `Score` when available) and assembling the full `RunRecord{Kind: "retrieval"}`.
- [X] T014 [US1] Implement `NewEvalRetrievalCmd()` in `internal/cli/internalcmd/eval_retrieval.go` (depends on T013): `Use: "eval-retrieval"`, flags `--cases` (required), `--out` (optional), always-JSON output; maps `EvaluationCase` load/validation errors to `invalid_argument`/`invalid_case` per contract §1's error table; on success returns `schema_version: 1` and the `{run, summary}` result shape from contract §1; when `--out` is set, also writes the full `RunRecord` JSON to that path.
- [X] T015 [US1] Wire `internalcmd.NewEvalRetrievalCmd()` into `internal/cli/internal.go`'s `cmd.AddCommand(...)` list (completes T007's stub).
- [X] T016 [P] [US1] Author the initial fixture case set under `specs/037-eval-quality-efficiency/fixture/cases/`: at least one free-text `query` case (exercising 036's BM25 path), one `task`/`intent` case, one case with a deliberately-forbidden identifier present in the fixture, one case with a tight `budget` override (exercising 016/035 budget pressure), and one large-specification case — satisfying spec FR-010's "larger project shape than a single small fixture."
- [X] T017 [US1] Run `go test ./internal/eval/... ./internal/cli/internalcmd/...` and confirm T008-T011 now pass; run quickstart.md §1 and §2 manually against a built binary and confirm the documented `Expected` outcomes.

**Checkpoint**: `misterspec internal eval-retrieval` is fully functional and independently testable — the MVP for this feature.

---

## Phase 4: User Story 2 - Run an Agent Task-Execution Evaluation (Priority: P2)

**Goal**: A maintainer can define a task-execution scenario, run it through a live agent session following a documented protocol, and record the result as a `RunRecord` the harness can later compare — success judged only by an independent acceptance test, never agent self-report.

**Independent Test**: Follow quickstart.md §5 for one fixture task: run it live, run its `acceptance_test`, hand-author a `RunRecord`, and confirm it validates against the shared `RunRecord` shape from T004/T005 (Phase 2) and loads without error.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T018 [P] [US2] Unit test for `EvaluationTask` loading/validation in `internal/eval/task_test.go`: a task file missing `acceptance_test` is rejected; a task whose `dir` does not resolve is rejected (mirrors data-model.md `EvaluationTask` fields); duplicate `id` within a loaded task set is rejected.
- [X] T019 [P] [US2] Unit test for `Metrics`/`TaskResult` (de)serialization in `internal/eval/record_test.go`: a `Metrics` value with `InputTokens` set but `InputTokensEstimated` unset fails validation (data-model.md Validation Rules Summary: "Every `Metrics` numeric field that is present MUST have its sibling `_estimated` boolean present and explicit"); a `TaskResult.Outcome` value other than `"pass"`/`"fail"`/`"inconclusive"` is rejected.
- [X] T020 [P] [US2] Unit test confirming a hand-authored `RunRecord{Kind: "task_execution"}` JSON file (matching quickstart.md §5's protocol) loads via the shared loader from T004 and its `Results` unmarshal into `[]TaskResult` without error.

### Implementation for User Story 2

- [X] T021 [P] [US2] Implement `EvaluationTask` type and task-file loader in `internal/eval/task.go`, mirroring `case.go`'s (T012) validation-error structure.
- [X] T022 [US2] Implement `TaskResult` and `Metrics` types in `internal/eval/record.go` (extends T004): `Outcome` enum (`pass`/`fail`/`inconclusive`), and `Metrics` with `InputTokens`/`OutputTokens`/`CachedTokens`/`Cost` each paired with an explicit `*Estimated bool` sibling field (data-model.md, research.md #6), plus `Calls`, `ExtraReads`, `Rework`, `LatencySeconds`.
- [X] T023 [P] [US2] Author the initial fixture task set under `specs/037-eval-quality-efficiency/fixture/tasks/`: at least one task with a `go test`-style `acceptance_test` command runnable against `fixture/project/`.
- [X] T024 [P] [US2] Write `docs/eval-task-execution-protocol.md` (or extend quickstart.md if the project prefers no new top-level doc — match existing convention) documenting the manual live-session protocol from quickstart.md §5 as a standalone, linkable procedure: pick a task, run it live, run `acceptance_test`, hand-author the `RunRecord`, save under `eval/runs/`.
- [X] T025 [US2] Run `go test ./internal/eval/...` and confirm T018-T020 now pass; manually execute quickstart.md §5 end-to-end for one fixture task and confirm the resulting `RunRecord` file is well-formed per T004/T005/T022.

**Checkpoint**: The task-execution protocol and its `RunRecord` shape are implemented and documented; User Stories 1 and 2 both work independently.

---

## Phase 5: User Story 3 - Compare a Proposed Change Against the Recorded Baseline (Priority: P3)

**Goal**: A maintainer can record a `RunRecord` as a named baseline and compare a later candidate run against it, seeing exactly which cases/tasks changed outcome, whether the comparison isolates a single configuration dimension, and whether a change is a real effect or normal repetition variance.

**Independent Test**: Follow quickstart.md §3 and §4: record a baseline, produce a candidate run, compare them, and confirm `changed_cases`/`dimension_diff` report correctly; confirm a multi-dimension change is flagged (quickstart.md §4).

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T026 [P] [US3] Unit test for `Baseline` (de)serialization in `internal/eval/record_test.go`: a `Baseline` naming one or more `run_ids` round-trips through JSON; `recorded_at` serializes as RFC 3339.
- [X] T027 [P] [US3] Unit test for `dimension_diff` computation in `internal/eval/compare_test.go`: two `RunRecord.Config` maps that are identical produce an empty `dimension_diff`; a difference in one field (e.g. `variant`) is named in the result (contract §2, "`dimension_diff` semantics"); `repetition_index` differences are excluded from the diff (data-model.md: "except `repetition_index`").
- [X] T028 [P] [US3] Unit test for per-case comparison in `internal/eval/compare_test.go`: a case `passed: false` in baseline and `passed: true` in candidate produces a `ChangedCase{Direction: "improved"}`; the reverse produces `"regressed"`; an unchanged case does not appear in `changed_cases` at all (contract §2 example).
- [X] T029 [P] [US3] Unit test for `RunRecord.Kind` mismatch in `internal/eval/compare_test.go`: comparing a `"retrieval"` baseline against a `"task_execution"` candidate returns an `incompatible_run` error (contract §2 error table) without producing a partial report.
- [X] T030 [P] [US3] Unit test for `variance_note` in `internal/eval/compare_test.go`: when `--repetitions` records disagree on a case's outcome among themselves, that case's `Direction` is `"neutral"` with `variance_note` set, never `"improved"`/`"regressed"` (contract §2, "`variance_note` semantics", FR-013).
- [X] T031 [P] [US3] Golden/contract test in `internal/cli/internalcmd/eval_compare_test.go` (skeleton, will fail until T034): a stale-baseline scenario (candidate `config` reports a newer `ranking_version` than the baseline recorded) returns `ok: true`, `result.stale: true` — not a hard failure (contract §2 error table, spec Edge Cases).

### Implementation for User Story 3

- [X] T032 [P] [US3] Implement `Baseline` type and named-baseline file lookup (`eval/baselines/<name>.json`) in `internal/eval/record.go`.
- [X] T033 [US3] Implement comparison logic in `internal/eval/compare.go` (depends on T004, T022, T032): `dimension_diff` computation (T027), per-case/task diffing into `[]ChangedCase` (T028), `incompatible_run` guard (T029), `aggregate.success_rate`/`aggregate.total_tokens_per_correct_task` computed together per FR-009 (never the ratio alone), and `variance_note` handling against an optional `--repetitions` set (T030).
- [X] T034 [US3] Implement `NewEvalCompareCmd()` in `internal/cli/internalcmd/eval_compare.go` (depends on T033): `Use: "eval-compare"`, flags `--baseline` (required), `--candidate` (required), `--repetitions` (optional); maps errors to `invalid_argument`/`incompatible_run`/`stale_baseline` per contract §2's error table; returns `schema_version: 1` and the `{comparison}` result shape from contract §2.
- [X] T035 [US3] Wire `internalcmd.NewEvalCompareCmd()` into `internal/cli/internal.go`'s `cmd.AddCommand(...)` list (completes T007's stub).
- [X] T036 [US3] Run `go test ./internal/eval/... ./internal/cli/internalcmd/...` and confirm T026-T031 now pass; run quickstart.md §3 and §4 manually against a built binary and confirm the documented `Expected` outcomes.

**Checkpoint**: All three user stories are independently functional; a full baseline → candidate → comparison cycle works end-to-end for the retrieval evaluation, and the same comparison logic accepts hand-authored task-execution `RunRecord`s.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories.

- [X] T037 [P] Update `CLAUDE.md`'s plan pointer is already current (037's own plan.md) — confirm no stale reference remains after this feature's tasks complete; no action needed if T-review finds none. Confirmed: `CLAUDE.md` points to `specs/037-eval-quality-efficiency/plan.md` (set during `/speckit-plan`).
- [X] T038 [P] Add a short "Evaluation Harness" section to `docs/` (or extend an existing architecture doc, matching project convention) cross-referencing `internal eval-retrieval`, `internal eval-compare`, and the task-execution protocol doc (T024), so a future proposal (e.g. PROP-06/08/13) can find this feature without re-reading spec 037 in full. Done as a note under `docs/context-engine-implementation.md`'s own "Phase 9 — Dogfooding and Evaluation" section (the natural existing home for this, since it explicitly supersedes 019's one-off exercise described there).
- [X] T039 Run `go vet ./...` and the project's full test suite (`go test ./...`) to confirm no regressions in `internal/context`, `internal/cli`, or elsewhere from the new `internal/eval` package and CLI wiring. Both clean; every package reports `ok`.
- [X] T040 Run the complete `quickstart.md` end-to-end (§1-§5) against a freshly built binary and confirm every documented `Expected` outcome, satisfying spec SC-006 ("at least one retrieval or ranking change proposal can be evaluated end-to-end through this harness... without any manual, ad hoc measurement step outside the harness") using the fixture case/task set itself as that first proposal. Verified: repeated `eval-retrieval` runs are byte-identical (minus `run_id`/`created_at`); `eval-compare` against the committed `retrieval-2026-09-init` and `task-exec-2026-09-init` baselines both succeed with `success_rate: 1` and no unexpected `dimension_diff`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion (T001-T003) — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) only.
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) only. Does not depend on US1's implementation, though it reuses the shared `RunRecord` type from Phase 2.
- **User Story 3 (Phase 5)**: Depends on Foundational (Phase 2). In practice its Independent Test (quickstart.md §3/§4) is easiest to run once US1 (Phase 3) produces real `RunRecord`s to compare, but the comparison *logic* itself (T026-T034) has no code dependency on US1/US2's command implementations — only on the shared types from Phase 2.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2/US3 — independently testable per quickstart.md §1-§2.
- **US2 (P2)**: No dependency on US1/US3 — independently testable per quickstart.md §5 (though exercising it against real repository history is more informative once US1's fixture project (T002) exists, which it does after Phase 1).
- **US3 (P3)**: Logically depends on having at least one `RunRecord` to compare (naturally produced by US1, or hand-authored per US2's protocol) to demonstrate its Independent Test, but its own code (T026-T035) only depends on Phase 2's shared types.

### Within Each User Story

- Tests (T008-T011, T018-T020, T026-T031) MUST be written and FAIL before their corresponding implementation task.
- Types before runner/logic before CLI command before wiring.
- Story complete (checkpoint) before moving to the next priority, though parallel staffing across US1/US2/US3 is possible once Phase 2 completes (see below).

### Parallel Opportunities

- T002 and T003 (Setup) can run in parallel.
- T005 and T006 (Foundational) can run in parallel with each other, after T004.
- All of T008-T011 (US1 tests) can run in parallel with each other.
- T012 and T016 (US1) can run in parallel with each other; T013 depends on T012.
- Once Phase 2 completes, US1 (Phase 3), US2 (Phase 4), and US3's test-writing (T026-T031, which only need Phase 2's types) can proceed in parallel by different contributors; US3's implementation (T032-T035) is easiest to validate once US1 exists but is not code-blocked by it.
- T037 and T038 (Polish) can run in parallel.

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for EvaluationCase loading/validation in internal/eval/case_test.go"
Task: "Unit test for the retrieval runner in internal/eval/retrieval_test.go"
Task: "Filesystem integration test in internal/eval/fixture_test.go"
Task: "Golden/contract test in internal/cli/internalcmd/eval_retrieval_test.go"

# Launch independent implementation tasks together:
Task: "Implement EvaluationCase type and case-file loader in internal/eval/case.go"
Task: "Author the initial fixture case set under specs/037-eval-quality-efficiency/fixture/cases/"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run quickstart.md §1-§2 independently.
5. `misterspec internal eval-retrieval` is now usable in CI (spec FR-014) even before US2/US3 exist.

### Incremental Delivery

1. Setup + Foundational → shared types and CLI wiring point ready.
2. Add User Story 1 → validate independently → retrieval evaluation usable in CI (MVP).
3. Add User Story 2 → validate independently → task-execution `RunRecord` protocol usable for live sessions.
4. Add User Story 3 → validate independently → baseline recording and comparison usable for gating proposals (spec FR-015).
5. Polish → cross-cutting documentation and full regression pass.

### Parallel Team Strategy

With multiple contributors, after Phase 2 completes: one contributor takes US1 (Phase 3), another takes US2 (Phase 4, including the fixture task set and protocol doc), and a third writes US3's tests (T026-T031) against Phase 2's types, then implements comparison logic (T032-T035) as soon as US1 produces a real `RunRecord` to validate against.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- Constitution Principle IV boundary (plan.md Constitution Check) is load-bearing for scope: no task in this list adds a command that launches, drives, or scrapes a live LLM/agent session — US2's "implementation" is a documented protocol plus a data shape, not automation.
- Every `Metrics` numeric field task (T022) must keep its `_estimated` sibling field mandatory, per data-model.md's Validation Rules Summary and spec FR-008/SC-004 — do not let a future task simplify this into an optional/omittable field.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
