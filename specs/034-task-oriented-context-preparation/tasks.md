---

description: "Task list template for feature implementation"
---

# Tasks: Recuperação orientada à tarefa e preparação de execução

**Input**: Design documents from `/specs/034-task-oriented-context-preparation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/prepare-command.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story and MUST be written and confirmed failing before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths are included in every description

## Path Conventions

Single Go module/CLI project (plan.md "Structure Decision") — all paths are relative to the repository root, inside the existing `internal/validation` package (one new file) and a new `internal/prepare` package, plus `internal/cli/internalcmd`.

---

## Phase 1: Setup

**Purpose**: Establish a clean, verified starting point. No new dependencies.

- [X] T001 Run `go build ./...` and `go test ./...` from the repository root and confirm a clean baseline (no pre-existing failures) before making any change, so any later red test is attributable to this feature's work.

**Checkpoint**: Baseline confirmed green — safe to start Foundational work.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: `Depends on:` parsing and basic Task readiness computation — the two things every user story below needs (US1's `ready` field, US2's blocking logic, US3's selection).

**⚠️ CRITICAL**: No user story task may start until this phase is complete.

- [X] T002 [P] Write failing unit tests in `internal/validation/task_dependencies_test.go` for `Depends on:` parsing (data-model.md "TaskDependency", research.md Decision 1): a single `Depends on: TASK-002` yields one entry; `Depends on: TASK-002, TASK-003` yields two; `Depends on: none` and an absent line both yield an empty `DependsOn`; a malformed entry (not a well-formed `TASK-###`) lands in `MalformedDependsOn`, not silently dropped.
- [X] T003 Implement `TaskDependency` and `parseTaskDependsOn` in `internal/validation/task_dependencies.go`, mirroring `parseTaskCoverage`'s existing `Serves:` grammar (`requirements.go:131-160`), to make T002 pass.
- [X] T004 [P] Add the two new Finding code constants to `internal/validation/findings.go`: `CodeTaskDependencyCycle = "task_dependency_cycle"`, `CodeInvalidTaskDependency = "invalid_task_dependency"` (contracts §3) — no dedicated test needed for bare constants, matching this file's existing convention.
- [X] T005 [P] Write failing unit tests in `internal/prepare/readiness_test.go` for `TaskReadiness` computation (data-model.md "TaskReadiness"): a Task with an empty `DependsOn` is `Ready: true`; a Task whose every dependency has Status `"complete"` is `Ready: true`; a Task with at least one incomplete dependency is `Ready: false` with `Blockers` naming exactly the incomplete ones.
- [X] T006 Create the `internal/prepare` package and implement `TaskReadiness`/`computeTaskReadiness` in `internal/prepare/readiness.go`, consuming `validation.TaskDependency` and each dependency's own Status (via the existing per-Task checkbox convention), to make T005 pass (depends on T003).

**Checkpoint**: `Depends on:` parsing and basic readiness computation exist — User Stories 1, 2, and 3 can now proceed in priority order.

---

## Phase 3: User Story 1 - Preparar a execução de uma tarefa em uma única resposta (Priority: P1) 🎯 MVP

**Goal**: `internal prepare SPEC-### --task TASK-###` returns the Task's identity, readiness, served requirements' own text, scope, verification method, and associated Plan sections — all in one read-only response.

**Independent Test**: Prepare a dependency-free Task and confirm every field is already populated without any follow-up call to `resolve`, `inspect`, or `context` (quickstart.md Scenario 1).

### Tests for User Story 1 ⚠️

- [X] T007 [P] [US1] Write failing unit tests in `internal/prepare/fields_test.go` for `Verify:`/`Scope:` line parsing (data-model.md "TaskFields", research.md Decision 1): each line's own text is captured verbatim; either line's absence yields an empty string, never an error.
- [X] T008 [P] [US1] Write failing unit tests in `internal/prepare/context_test.go` for `RequirementText` lookup (data-model.md "RequirementText"): given a Task's own `Serves:` references and its owning Spec's `spec.md` body, each reference's own `### R<N>` section content is returned verbatim, alongside a `contextengine.Fingerprint` of that content.
- [X] T009 [P] [US1] Write failing unit tests in `internal/prepare/context_test.go` for `PlanSectionAssociation` (data-model.md "PlanSectionAssociation", research.md Decision 4): a Plan section whose body mentions a served Requirement (bare `R<N>` or composite `SPEC-###:R#` form) is associated, with `MatchedRequirements` naming which one(s); a section mentioning none is excluded, not returned empty.
- [X] T010 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: `internal prepare SPEC-### --task TASK-###` for a dependency-free Task returns `task`, `heading`, `ready: true`, `requirements`, `scope`, `verify`, and `plan_sections` all populated in one call (spec Acceptance Scenario 1, contracts §1).
- [X] T011 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: two different Tasks of the same Spec, serving different Requirements and declaring different `Scope:`, receive genuinely different `requirements`/`scope`/`plan_sections` — never the same bundle repeated (spec Acceptance Scenario 2).
- [X] T012 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: running `prepare` leaves every project file byte-identical before and after (mirrors 031's own "never modifies artifacts" test pattern) — confirming no command execution occurred (spec Acceptance Scenario 3, FR-002).

### Implementation for User Story 1

- [X] T013 [US1] Implement `TaskFields` and its `Verify:`/`Scope:` parsing in `internal/prepare/fields.go`, to make T007 pass.
- [X] T014 [US1] Implement `RequirementText` lookup in `internal/prepare/context.go` — parsing the owning Spec's `spec.md` body via `artifacts.ParseDocument` for each served Requirement's own `### R<N>` section, reusing `contextengine.Fingerprint` — to make T008 pass.
- [X] T015 [US1] Implement `PlanSectionAssociation` in `internal/prepare/context.go` — parsing `plan.md`'s own sections via `artifacts.ParseDocument` and matching each section's body against the Task's served Requirement references in both the bare and composite forms — to make T009 pass (depends on T014).
- [X] T016 [US1] Implement `TaskPreparation` assembly in `internal/prepare/context.go`, combining the Task's identity, `TaskReadiness` (T006), `RequirementText` list (T014), `TaskFields` (T013), and `PlanSectionAssociation` list (T015) into the one response bundle (data-model.md "TaskPreparation").
- [X] T017 [US1] Implement `internal prepare SPEC-### --task TASK-###` in `internal/cli/internalcmd/prepare.go`, rendering the JSON shape from contracts §1 from T016's assembly, to make T010–T012 pass (depends on T016).
- [X] T018 [US1] Register `internalcmd.NewPrepareCmd()` in `internal/cli/internal.go`'s existing `AddCommand(...)` list.

**Checkpoint**: User Story 1 is independently functional and testable — one call now returns everything a Task needs to start.

---

## Phase 4: User Story 2 - Nunca apresentar uma tarefa bloqueada como pronta (Priority: P1)

**Goal**: A Task with an incomplete dependency is reported as blocked, naming the specific missing dependency; a dependency cycle is detected (by `internal validate` and by `prepare` alike) and never causes a hang; an invalid dependency declaration (nonexistent Task, or a Task in a different Spec) is rejected.

**Independent Test**: Prepare a Task declaring a dependency on an incomplete Task; confirm `ready: false` naming that exact dependency, exit code `10` (quickstart.md Scenario 2).

### Tests for User Story 2 ⚠️

- [X] T019 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: a Task declaring `Depends on:` an incomplete Task → `ready: false`, `blockers` naming exactly that Task, exit code `10` (spec Acceptance Scenario 1, contracts §1's blocked shape).
- [X] T020 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: marking that dependency's checkbox complete and re-preparing → `ready: true`, exit code `0`, full context populated (spec Acceptance Scenario 2).
- [X] T021 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: a Task with no `Depends on:` line at all is `ready: true` by default (spec Acceptance Scenario 3).
- [X] T022 [P] [US2] Write failing unit test in `internal/validation/task_dependencies_test.go` for Task-level cycle detection, reusing `dependency_graph.go`'s existing `DependencyGraph`/`detectCycles` at Task-number granularity (research.md Decision 2): two Tasks depending on each other produce exactly one `task_dependency_cycle` Finding naming both, via `ValidateProject`; a Task depending on itself is the same algorithm's degenerate case; an acyclic Task graph produces none.
- [X] T023 [P] [US2] Write failing unit test in `internal/validation/task_dependencies_test.go`: a `Depends on:` entry naming a nonexistent Task number, or (via the composite form) a Task in a different Spec, produces an `invalid_task_dependency` Finding (data-model.md "TaskDependency" validation rules).
- [X] T024 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: preparing a Task that participates in a dependency cycle returns promptly (bounded time, no hang) with `ready: false` and `blockers` reflecting the cycle (spec Acceptance Scenario 4).

### Implementation for User Story 2

- [X] T025 [US2] Implement per-Spec Task-number `DependencyGraph` construction and cycle/invalid-reference detection in `internal/validation/task_dependencies.go`, reusing the existing `detectCycles`/`DependencyGraph` types from `dependency_graph.go` unchanged, to make T022/T023 pass (depends on T003).
- [X] T026 [US2] Wire `CodeTaskDependencyCycle`/`CodeInvalidTaskDependency` Findings into `validation.ValidateProject` and `ValidateEntity` (`internal/validation/validator.go`), mirroring 032's own Spec-level `dependency_cycle` wiring, to make T022/T023 pass end-to-end (depends on T025).
- [X] T027 [US2] Update `computeTaskReadiness` (`internal/prepare/readiness.go`) to treat a Task inside a detected cycle as blocked without infinite-looping (reusing T025's cycle data), and update `internal/cli/internalcmd/prepare.go` to render `ready: false`/`blockers` with exit code `10` when not ready, to make T019–T021 and T024 pass (depends on T006, T025).

**Checkpoint**: User Stories 1 AND 2 both work independently — full context is delivered, and a blocked or cyclic Task is never presented as ready.

---

## Phase 5: User Story 3 - Selecionar automaticamente a próxima tarefa pronta (Priority: P2)

**Goal**: Omitting `--task` selects the lowest-numbered Ready Task in the Spec, recomputed fresh on every call; a Spec with no Ready Task says so explicitly.

**Independent Test**: In a Spec with one blocked and two Ready Tasks, prepare with no `--task` and confirm the lowest-numbered Ready one is selected; complete it and confirm the next call reflects the new state (quickstart.md Scenario 5).

### Tests for User Story 3 ⚠️

- [X] T028 [P] [US3] Write failing unit test in `internal/prepare/selection_test.go` for `SelectTask` (research.md Decision 6): given multiple Ready Tasks, the lowest-numbered one is chosen; given one blocked and two Ready Tasks, the lowest-numbered *Ready* one is chosen, not the lowest-numbered overall.
- [X] T029 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: `internal prepare SPEC-###` with no `--task` selects the lowest-numbered Ready Task, matching T028's fixture (spec Acceptance Scenario 1).
- [X] T030 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: completing that Task and re-running the same command selects a different Task, reflecting the new state — never reusing the previous response (spec Acceptance Scenario 2, FR-013).
- [X] T031 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/prepare_test.go`: a Spec where every Task is complete or blocked returns `preparation: null` with a clear, explicit message, exit code `0` (spec Acceptance Scenario 3, FR-012).

### Implementation for User Story 3

- [X] T032 [US3] Implement `SelectTask` (lowest-numbered Ready Task) in `internal/prepare/selection.go`, to make T028 pass (depends on T006).
- [X] T033 [US3] Wire automatic selection into `internal/cli/internalcmd/prepare.go` when `--task` is omitted, including the explicit "no Task is ready" `preparation: null` response, to make T029–T031 pass (depends on T032, T017).

**Checkpoint**: All three user stories are independently functional — one-call preparation, safe blocking/cycle handling, and deterministic automatic selection.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the whole feature end-to-end and keep published docs in sync.

- [X] T034 [P] Review `kit/skills/mister-implement/SKILL.md` (Procedure steps 1–3, lines 133-147) and `kit/skills/mister-tasks/SKILL.md` (Procedure step 5, lines 129-135) against the now-implemented `Depends on:`/`Verify:`/`Scope:` convention and the new `internal prepare` operation; update wording only where it genuinely still describes the old manual resolve→inspect→context→dependency-check chain as the only option, without mandating a full Skill rewrite as part of this feature.
- [X] T035 [P] Update `site/commands.html` with the new `internal prepare` command (flags, example, blocked-response shape, exit code `10`) and the two new validation Finding codes, mirroring 031/032/033's own precedent of keeping the published command reference in sync.
- [X] T036 Manually execute every scenario in `quickstart.md` (1–6) against a disposable temp project built from the finished binary, and confirm each documented "Pass condition" holds.
- [X] T037 Run `go build ./...`, `go test ./...`, and `go vet ./...` from the repository root and confirm a fully green result across the whole module, not just the packages touched by this feature.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup (T001). BLOCKS every user story (T007–T033).
- **User Story 1 (Phase 3)**: Depends on Foundational (T002–T006) only.
- **User Story 2 (Phase 4)**: Depends on Foundational directly (T003 for parsing, T006 for readiness); its CLI-level tests (T019–T021, T024) exercise the same `prepare` command User Story 1 builds, so T027's implementation task is sequenced after User Story 1's T017 lands, even though the underlying validation logic (T025–T026) does not depend on US1 at all.
- **User Story 3 (Phase 5)**: Depends on Foundational (T006) and on User Story 1's `prepare` command (T017) existing to wire automatic selection into.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on US2 or US3.
- **User Story 2 (P1)**: Its validation half (`internal/validation/task_dependencies.go`'s cycle/invalid-reference detection, T025–T026) is independent of US1; its `prepare`-response half (T027) is sequenced after US1's T017.
- **User Story 3 (P2)**: Depends on User Story 1's `prepare` command existing (T017) to attach automatic selection to.

### Within Each User Story

- Tests are written and confirmed failing before their corresponding implementation task (Principle V).
- `TaskFields` (T013) and `RequirementText` (T014) before `PlanSectionAssociation` (T015, needs T014's Requirement text) before `TaskPreparation` assembly (T016) before the CLI wiring (T017) before command registration (T018).
- Task-graph construction (T025) before Finding wiring (T026) before readiness/CLI updates that consume it (T027).
- `SelectTask` (T032) before its CLI wiring (T033).

### Parallel Opportunities

- T002, T005 (Foundational tests, different files) — run in parallel.
- T007, T008, T009 (US1 unit tests, independent files/fixtures) — run in parallel.
- T010, T011, T012 (US1 CLI tests, independent fixtures within `prepare_test.go`) — run in parallel.
- T019–T024 (US2 tests) — independent fixtures; run in parallel with each other.
- T028–T031 (US3 tests) — independent fixtures; run in parallel with each other.
- T034 and T035 (Polish doc updates) — different files; run in parallel.

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together once Phase 2 is complete:
Task: "Unit tests for Verify:/Scope: parsing in internal/prepare/fields_test.go"
Task: "Unit tests for RequirementText lookup in internal/prepare/context_test.go"
Task: "Unit tests for PlanSectionAssociation in internal/prepare/context_test.go"
Task: "CLI test: one-call preparation for a dependency-free Task in internal/cli/internalcmd/prepare_test.go"
Task: "CLI test: two Tasks receive genuinely different context in internal/cli/internalcmd/prepare_test.go"
Task: "CLI test: prepare never modifies project files in internal/cli/internalcmd/prepare_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001).
2. Complete Phase 2: Foundational (T002–T006) — CRITICAL, blocks all stories.
3. Complete Phase 3: User Story 1 (T007–T018).
4. **STOP and VALIDATE**: run quickstart.md Scenario 1 against a temp project.
5. This alone already collapses the manual resolve→inspect→context chain into one call and is independently shippable — a Task with no dependency already gets `ready: true` correctly from Foundational's basic readiness check, even before User Story 2's cycle detection lands.

### Incremental Delivery

1. Setup + Foundational → dependency parsing and basic readiness ready.
2. Add User Story 1 → one-call preparation → validate independently (MVP).
3. Add User Story 2 → blocked/cyclic Tasks never presented as ready → validate independently.
4. Add User Story 3 → deterministic automatic selection → validate independently.
5. Polish → confirm Skills/docs match code, run full quickstart, whole-repo test gate.

### Parallel Team Strategy

1. One contributor completes Setup + Foundational (T001–T006) — the shared dependency every story needs.
2. Once Phase 2 lands:
   - Contributor A: User Story 1 (`internal/prepare/fields.go`, `context.go`, `internal/cli/internalcmd/prepare.go`).
   - Contributor B: User Story 2's validation half (`internal/validation/task_dependencies.go`'s cycle detection, T025–T026) can start immediately; its `prepare`-response half (T027) waits on Contributor A's T017.
   - Contributor C starts User Story 3 (T032) once T006 lands, then wires it (T033) once Contributor A's T017 lands.
3. Polish (Phase 6) runs once all three land.

---

## Notes

- `[P]` tasks touch different files or independent fixtures within a shared test file — no shared mutable state.
- `[Story]` labels map every Phase 3+ task to its spec.md user story for traceability.
- Confirm each test fails before writing its implementation (Principle V, NON-NEGOTIABLE).
- Avoid: vague tasks, same-file conflicts inside a `[P]` group, and any change to `internal/context/collector.go`'s existing five-entity-type restriction (research.md Decision 3 — out of scope for this feature).
