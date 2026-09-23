---

description: "Task list template for feature implementation"
---

# Tasks: Modelo canônico de tarefas e identidade composta

**Input**: Design documents from `/specs/031-canonical-task-identity/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/task-identity-resolution.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story and MUST be written and confirmed failing before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths are included in every description

## Path Conventions

Single Go module/CLI project (plan.md "Structure Decision") — all paths are relative to the repository root, inside the existing `internal/ids`, `internal/operations`, `internal/validation`, `internal/cli/internalcmd` packages. No new top-level directory.

---

## Phase 1: Setup

**Purpose**: Establish a clean, verified starting point. No new dependencies, no new project structure — this feature only modifies existing packages (plan.md "Structure Decision").

- [X] T001 Run `go build ./...` and `go test ./...` from the repository root and confirm a clean baseline (no pre-existing failures) before making any change, so any later red test is attributable to this feature's work.

**Checkpoint**: Baseline confirmed green — safe to start Foundational work.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The composite `TaskID` type, its parser, and the rescoped per-Spec/cross-Spec scan data (research.md Decisions 1–2, data-model.md `TaskID`/`TaskScanEntry`) that every user story below depends on.

**⚠️ CRITICAL**: No user story task may start until this phase is complete.

- [X] T002 [P] Write failing unit tests for the new `ids.TaskID` composite type (`Spec ids.EntityID`, `Local ids.EntityID`, `String()` renders `"SPEC-###:TASK-###"`, equality requires both halves equal per data-model.md "TaskID (composite identity)") in `internal/ids/ids_test.go`.
- [X] T003 [P] Write failing unit tests for `ids.ParseTaskRef` covering: valid composite (`"SPEC-014:TASK-003"`), valid bare (`"TASK-003"`, Spec half absent), malformed composite (`"SPEC-014:"`, `":TASK-003"`, `"SPEC-014-TASK-003"`), wrong ID width on either half, and wrong entity type in either slot (e.g. `"TASK-003:TASK-004"`) — all MUST return `ids.ErrInvalidIDSyntax` per contracts §1's table — in `internal/ids/ids_test.go`.
- [X] T004 Add the `TaskID` type to `internal/ids/types.go` (fields, `String()`, equality) to make T002 pass.
- [X] T005 Implement `ids.ParseTaskRef(raw string, cfg project.Configuration) (TaskID, error)` in `internal/ids/ids.go` per contracts §1, to make T003 pass (depends on T004).
- [X] T006 [P] Write failing unit tests in `internal/ids/scan_test.go` for the rescoped Task scan: two Specs each claiming `TASK-001` produce two separate `TaskScanEntry{Spec, Task, Path}` entries (not one collision); two `## TASK-001` headings inside the *same* Spec's `tasks.md` group into one same-Spec entry with two `Path`s (data-model.md "TaskScanEntry" and its "Derived views" §1–2).
- [X] T007 Rescope `scanTaskHeadings` in `internal/ids/scan.go` to key claims by `(specNumber, taskNumber)` instead of bare `taskNumber` — extracting `specNumber` from each `tasks.md`'s parent directory (`SPEC-###`) via the existing `idDirPattern`/path-glob convention already used by `scanDirs` — and expose the two derived views from research.md Decision 2 (same-Spec duplicates; cross-Spec by-number index), to make T006 pass (depends on T006).

**Checkpoint**: `ids.TaskID`, `ids.ParseTaskRef`, and the rescoped scan are implemented and unit-tested — User Stories 1–3 can now proceed in priority order.

---

## Phase 3: User Story 1 - Resolver uma tarefa sem ambiguidade entre Specs (Priority: P1) 🎯 MVP

**Goal**: A composite reference (`SPEC-###:TASK-###`) always resolves to exactly one task; a bare `TASK-###` resolves unambiguously when only one Spec claims that number, and fails with an actionable, structured error (never an arbitrary pick) when more than one Spec claims it.

**Independent Test**: Two Specs each with a `TASK-001`; `SPEC-A:TASK-001` and `SPEC-B:TASK-001` each resolve to their own task only; bare `TASK-001` fails with both candidates named (quickstart.md Scenarios 1–2).

### Tests for User Story 1 ⚠️

- [X] T008 [P] [US1] Write failing integration test in `internal/operations/resolve_test.go`: given two Specs each with `TASK-001` (temp repo fixture), `ResolveTask(root, cfg, "SPEC-A:TASK-001", nil)` and `ResolveTask(root, cfg, "SPEC-B:TASK-001", nil)` each resolve to their own Spec's task and never mention the other (spec Acceptance Scenarios 1–2, contracts §2).
- [X] T009 [P] [US1] Write failing integration test in `internal/operations/resolve_test.go`: bare `"TASK-001"` with no `specCtx`, claimed by two Specs, returns `ErrSpecContextRequired` wrapped in a `SpecContextRequiredError` naming both candidate Specs — not `ErrEntityAmbiguous` (spec Acceptance Scenario 3, contracts §2's table).
- [X] T010 [P] [US1] Write failing integration test in `internal/operations/resolve_test.go`: bare `"TASK-001"` with no `specCtx`, claimed by exactly one Spec project-wide, resolves successfully (today's behavior preserved — spec FR-009, contracts §2's table).
- [X] T011 [P] [US1] Write failing integration test in `internal/operations/inspect_test.go`: `Inspect(root, cfg, "SPEC-A:TASK-001")` returns `InspectResult` scoped to Spec A's task, consuming the new composite path (data-model.md "Task (existing entity, identity clarified)").
- [X] T012 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/inspect_test.go` (or a sibling test file matching existing naming): `misterspec internal inspect SPEC-A:TASK-001` succeeds with JSON naming only Spec A's task; `misterspec internal inspect TASK-001` (ambiguous case) exits non-zero with a structured JSON error naming both candidate Specs (contracts §5, quickstart.md Scenario 2).

### Implementation for User Story 1

- [X] T013 [US1] Add `ErrSpecContextRequired` sentinel and `SpecContextRequiredError{TaskNumber int, Candidates []ids.EntityID}` (with `Error()`/`Unwrap()`, mirroring the existing `AmbiguousIDError` shape) to `internal/operations/resolve.go`, to make T009 pass (depends on T007).
- [X] T014 [US1] Implement `operations.ResolveTask(root string, cfg project.Configuration, raw string, specCtx *ids.EntityID) (ResolvedLocation, error)` in `internal/operations/resolve.go` per contracts §2 — composite input resolves directly; bare input with no `specCtx` consults the cross-Spec index from T007 (one claimant → resolve, zero → `ErrEntityNotFound`, many → `ErrSpecContextRequired`); a `specCtx` that disagrees with a composite `raw`'s Spec half is rejected as invalid input — to make T008, T009, T010 pass (depends on T013).
- [X] T015 [US1] Update `operations.Resolve` (`internal/operations/resolve.go`) so its existing `ids.Task` path delegates to `ResolveTask`'s bare-form handling (no `specCtx`), preserving every other entity type's behavior unchanged (depends on T014).
- [X] T016 [US1] Update `Inspect`/`inspectTask` in `internal/operations/inspect.go` to accept and pass through a composite Task reference via `ResolveTask`, to make T011 pass (depends on T014).
- [X] T017 [US1] Extend `misterspec internal inspect <id>` in `internal/cli/internalcmd/inspect.go` to accept the composite `SPEC-###:TASK-###` form for its existing single positional `<id>` argument, and to render `SpecContextRequiredError` as a structured JSON error including the candidate Specs (contracts §5), to make T012 pass (depends on T016).

**Checkpoint**: User Story 1 is independently functional and testable — composite resolution works, and ambiguous bare references fail safely instead of guessing.

---

## Phase 4: User Story 2 - Detectar duplicatas apenas dentro da mesma Spec (Priority: P1)

**Goal**: `misterspec internal validate` (project-wide) and `validate SPEC-###` (single-Spec) stop false-flagging matching `TASK-NNN` numbers across different Specs as duplicates, while still correctly flagging two identical `TASK-NNN` headings inside the same Spec's `tasks.md`.

**Independent Test**: Two Specs each with a legitimate `TASK-001` produce zero duplicate findings; a single Spec with two `## TASK-001` headings produces exactly one finding naming that Spec and both locations (quickstart.md Scenarios 3–4).

### Tests for User Story 2 ⚠️

- [X] T018 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: two Specs each with a legitimate `TASK-001` in their own `tasks.md` → `ValidateProject` returns zero `CodeDuplicateID` findings for Task number 1 (spec Acceptance Scenario, User Story 2 §1; contracts §3's table, row 1).
- [X] T019 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: one Spec's `tasks.md` with two `## TASK-001` headings → `ValidateProject` returns exactly one `CodeDuplicateID` finding whose message names that Spec and both heading locations (spec Acceptance Scenario, User Story 2 §2; contracts §3's table, row 2 and its "Message wording" note).
- [X] T020 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: `ValidateEntity(root, cfg, "SPEC-###")` on the Spec from T019 returns the same `CodeDuplicateID` finding as `ValidateProject` filtered to that Spec (spec FR-011).

### Implementation for User Story 2

- [X] T021 [US2] Rescope the Task duplicate-detection loop in `validation.ValidateProject` (`internal/validation/validator.go`) to iterate the Specs already discovered via `projectEntityTypes`, and for each Spec consult only that Spec's same-Spec `TaskScanEntry` view (from T007) — replacing the current single project-wide `ids.Scan(root, cfg, ids.Task)` duplicate loop — to make T018 and T019 pass (depends on T007).
- [X] T022 [US2] Update the `CodeDuplicateID` `Finding.Message` produced in T021 to explicitly name the owning Spec (e.g. `"SPEC-001: 2 Task headings claim the same number 1: [...]"`), per contracts §3.
- [X] T023 [US2] Wire the same per-Spec Task duplicate check into `validation.ValidateEntity` (`internal/validation/validator.go`) for a `Spec`-typed target, so single-Spec validation includes it, to make T020 pass (spec FR-011; depends on T021).

**Checkpoint**: User Stories 1 AND 2 both work independently — resolution and validation now agree on what counts as a duplicate.

---

## Phase 5: User Story 3 - Migrar projetos existentes sem renumeração automática (Priority: P2)

**Goal**: A read-only diagnostic lists every Task number claimed by more than one Spec project-wide (the cases that collided under the old global-scan behavior), without renumbering or otherwise modifying any artifact.

**Independent Test**: Running the diagnostic against a project with a real cross-Spec collision lists it; `git status` shows zero changes afterward (quickstart.md Scenario 5).

### Tests for User Story 3 ⚠️

- [X] T024 [P] [US3] Write failing unit test in `internal/operations` (new `migration_test.go` or alongside `resolve_test.go`, matching existing package conventions) building `MigrationDiagnosticEntry` values from a fixture with a genuine cross-Spec collision, asserting `TaskNumber`, `Specs`, and `Paths` per data-model.md "MigrationDiagnosticEntry" (depends on T007).
- [X] T025 [P] [US3] Write failing CLI test in `internal/cli/internalcmd` (new test file matching the new subcommand, e.g. `migration_check_test.go`) asserting the JSON shape from contracts §4 (`ok`, `collisions[].task_number/specs/paths`), that the command exits `0` on a genuine collision (informational, not an error), and that a project with no collisions returns an empty `collisions` array.
- [X] T026 [US3] Write failing integration test asserting the diagnostic command performs no filesystem writes: run it against a temp repo fixture with a collision, then assert every file's mtime/content is unchanged (spec FR-007/FR-008, quickstart.md Scenario 5's `git status --short` check) — co-located with T025's test file.

### Implementation for User Story 3

- [X] T027 [US3] Implement `MigrationDiagnosticEntry` computation from the cross-Spec index (T007) in `internal/operations` (e.g. `internal/operations/migration.go`), per data-model.md "MigrationDiagnosticEntry", to make T024 pass.
- [X] T028 [US3] Add the `misterspec internal migration-check tasks` subcommand in `internal/cli/internalcmd/` (new file), wired into the internal command tree, rendering the JSON shape from contracts §4 and never writing to the filesystem, to make T025 and T026 pass (depends on T027).

**Checkpoint**: All three user stories are independently functional — composite resolution, correctly-scoped duplicate detection, and a safe migration diagnostic.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the whole feature end-to-end and confirm the previously-identified doc/code mismatch is fully closed.

- [X] T029 [P] Re-read `kit/skills/mister-implement/SKILL.md`, `kit/skills/mister-tasks/SKILL.md`, and `kit/skills/mister-analyze/SKILL.md` and confirm their existing per-Spec task-numbering wording now matches the implemented behavior exactly (plan.md Project Structure note); update wording only if a genuine mismatch is found — do not rewrite sections that already match.
- [X] T030 [P] Review `docs/architecture-specification.md` §29 (or wherever Task ID scanning/resolution is described) for wording that still describes the old global-scan behavior, and correct it to describe per-Spec-scoped duplicate detection and composite resolution.
- [X] T031 Manually execute every scenario in `quickstart.md` (1–6) against a disposable temp project built from the finished binary, and confirm each documented "Pass condition" holds.
- [X] T032 Run `go build ./...`, `go test ./...`, and `go vet ./...` from the repository root and confirm a fully green result across the whole module, not just the packages touched by this feature.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup (T001). BLOCKS every user story (T008–T028).
- **User Story 1 (Phase 3)**: Depends on Foundational (T002–T007) only.
- **User Story 2 (Phase 4)**: Depends on Foundational (specifically T007's rescoped scan). Independent of User Story 1's `operations` changes — touches `internal/validation`, not `internal/operations`.
- **User Story 3 (Phase 5)**: Depends on Foundational (specifically T007's cross-Spec index). Independent of User Stories 1 and 2's own changes.
- **Polish (Phase 6)**: Depends on all three user stories being complete (T029–T030 verify wording against final behavior; T031 exercises all scenarios; T032 is a final whole-repo gate).

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on US2 or US3.
- **User Story 2 (P1)**: No dependency on US1 or US3 — shares only the Phase 2 scan data.
- **User Story 3 (P2)**: No dependency on US1 or US2 — shares only the Phase 2 scan data.

All three user stories touch different files (`internal/operations/*` for US1, `internal/validation/validator.go` for US2, a new `internal/operations/migration.go` + new CLI subcommand for US3) once Phase 2 lands, so they can proceed in parallel if staffed, or sequentially in priority order (US1 → US2 → US3) otherwise.

### Within Each User Story

- Tests are written and confirmed failing before their corresponding implementation task (Principle V).
- `ResolveTask`/type work before the CLI surface that depends on it (US1: T013–T016 before T017).
- Scan rescoping (US2: T021) before Finding-message wording (T022) and before wiring into `ValidateEntity` (T023).
- Diagnostic computation (US3: T027) before the CLI subcommand that renders it (T028).

### Parallel Opportunities

- T002 and T003 (Foundational, different assertions in the same file but independent of each other's subject) — mark `[P]` only if written as separate test functions with no shared fixture state; otherwise run sequentially in the same file.
- T008, T009, T010, T011, T012 (US1 tests) — different files or independent fixtures within `resolve_test.go`; run in parallel.
- T018, T019, T020 (US2 tests) — independent fixtures within `validator_test.go`; run in parallel.
- T024, T025 (US3 tests) — different files; run in parallel.
- Once Phase 2 (Foundational) is complete, US1, US2, and US3 implementation work can proceed in parallel by different contributors, since they touch disjoint files.

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together once Phase 2 is complete:
Task: "Integration test: SPEC-A:TASK-001 and SPEC-B:TASK-001 resolve independently in internal/operations/resolve_test.go"
Task: "Integration test: bare TASK-001 ambiguous across Specs returns SpecContextRequiredError in internal/operations/resolve_test.go"
Task: "Integration test: bare TASK-001 unique project-wide still resolves in internal/operations/resolve_test.go"
Task: "Integration test: Inspect(SPEC-A:TASK-001) scoped correctly in internal/operations/inspect_test.go"
Task: "CLI test: inspect SPEC-A:TASK-001 and ambiguous inspect TASK-001 in internal/cli/internalcmd/inspect_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001).
2. Complete Phase 2: Foundational (T002–T007) — CRITICAL, blocks all stories.
3. Complete Phase 3: User Story 1 (T008–T017).
4. **STOP and VALIDATE**: run quickstart.md Scenarios 1–2 against a temp project.
5. This alone already fixes the most user-visible symptom (wrong/ambiguous task resolution) and is independently shippable.

### Incremental Delivery

1. Setup + Foundational → composite identity and rescoped scan ready.
2. Add User Story 1 → unambiguous resolution → validate independently (MVP).
3. Add User Story 2 → validation stops false-flagging cross-Spec matches → validate independently.
4. Add User Story 3 → migration diagnostic available → validate independently.
5. Polish → confirm docs match code, run full quickstart, whole-repo test gate.

### Parallel Team Strategy

1. One contributor completes Setup + Foundational (T001–T007) — this is the shared dependency every story needs.
2. Once Phase 2 lands:
   - Contributor A: User Story 1 (`internal/operations`, `internal/cli/internalcmd/inspect.go`).
   - Contributor B: User Story 2 (`internal/validation/validator.go`).
   - Contributor C: User Story 3 (new `internal/operations/migration.go` + new CLI subcommand).
3. Stories integrate independently since they touch disjoint files; Polish (Phase 6) runs once all three land.

---

## Notes

- `[P]` tasks touch different files or independent fixtures within a shared test file — no shared mutable state.
- `[Story]` labels map every Phase 3+ task to its spec.md user story for traceability.
- Every user story is independently completable, testable, and shippable on its own.
- Confirm each test fails before writing its implementation (Principle V, NON-NEGOTIABLE).
- Avoid: vague tasks, same-file conflicts inside a `[P]` group, and cross-story dependencies that would break US1/US2/US3 independence.
