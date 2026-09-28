---

description: "Task list for Numeração Global de Tasks"
---

# Tasks: Numeração Global de Tasks

**Input**: Design documents from `/specs/045-global-task-numbering/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/next-task-id-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic scanning/computation logic, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US3)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. No new package. One new file in `internal/ids` (`next_task_id.go`), one new file in `internal/cli/internalcmd` (`next_task_id.go`), one registration edit (`internal/cli/internal.go`), one Skill doc edit (`kit/skills/mister-tasks/SKILL.md`), one docs edit (`site/commands.html`).

---

## Phase 1: Setup

**Purpose**: N/A — this feature adds no new package (plan.md "Project Structure"); it extends `internal/ids` and `internal/cli/internalcmd`, both of which already exist. No setup tasks are required.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The one primitive every user story depends on — a project-wide "next Task number" computation.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundational

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T001 [P] Add `internal/ids/next_task_id_test.go` (write first): `NextTaskID(root, cfg)` returns `TASK-001` when no Task exists anywhere in the project (spec FR-007); returns `TASK-(N+1)` when the highest existing Task number found in *any* Spec is N, even when the "current" Spec has no Tasks of its own yet (spec US1 Acceptance Scenario 1, data-model.md); returns `TASK-052` when the current Spec already has TASK-050/TASK-051 and no other Spec has a higher number (spec US1 Acceptance Scenario 3); the returned `EntityID.Width` equals `cfg.IDWidth`; and two consecutive calls with no intervening write return the same value (contracts/next-task-id-contract.md §3 read-only guarantee). Expected to FAIL until T002.

### Implementation for Foundational

- [X] T002 Implement `ids.NextTaskID(root string, cfg project.Configuration) (EntityID, error)` in new file `internal/ids/next_task_id.go`, composed exactly as `Scan(root, cfg, Task)` + `NextID(result.IDs, Task, cfg.IDWidth)` — no new scanning logic, pure composition of the two already-exported primitives (research.md Decision 1, data-model.md "NextTaskID") (depends on T001).

**Checkpoint**: `ids.NextTaskID` available and correct; nothing downstream consumes it yet.

---

## Phase 3: User Story 1 - Próximo número de Task nunca reinicia (Priority: P1) 🎯 MVP

**Goal**: Give the Task-decomposition Skill a deterministic, project-wide "next Task number" it can call before authoring each new `## TASK-NNN` heading.

**Independent Test**: In a project where the highest existing Task number anywhere is N, decomposing a brand-new, empty Spec's Tasks produces a first Task numbered N+1, never 1 (spec.md US1 Independent Test).

### Tests for User Story 1

- [X] T003 [P] [US1] Add `internal/cli/internalcmd/next_task_id_test.go` (write first), mirroring `migration_check_tasks_test.go`'s existing shape: an empty project returns `{"ok": true, "next_task_id": "TASK-001"}` (contracts §2); the highest existing Task number found in a *different* Spec than the one being decomposed still yields that number + 1 (spec US1 Acceptance Scenario 1); a Spec that already has TASK-050/TASK-051, with no higher number elsewhere in the project, yields `TASK-052` (spec US1 Acceptance Scenario 3); and the command never writes to the filesystem — an existing `tasks.md`'s content and mtime are unchanged after the call (mirrors `TestMigrationCheckTasksCmd_NeverWritesToTheFilesystem`, contracts §3). Expected to FAIL until T004.

### Implementation for User Story 1

- [X] T004 [US1] Implement `internalcmd.NewNextTaskIDCmd()` in new file `internal/cli/internalcmd/next_task_id.go`: `misterspec internal next-task-id [--dir <path>]`, calling `ids.NextTaskID` and rendering `{"next_task_id": <EntityID.String()>}` via `WriteSuccess` (contracts §1/§2) (depends on T002, T003).
- [X] T005 [US1] Register the new command in `internal/cli/internal.go`, mirroring exactly how `internalcmd.NewMigrationCheckTasksCmd()` is already registered (depends on T004).
- [X] T006 [US1] Update `kit/skills/mister-tasks/SKILL.md`: add `internal next-task-id` to the **Deterministic Operations** required-operations list; rewrite **Procedure** step 5 to call it before writing each new `## TASK-NNN` heading (incrementing locally for any additional Tasks authored within the same run, per research.md Decision 3); correct the now-inaccurate "No allocator operation exists for Task IDs" sentence in **Allowed Modifications** to reflect that a read-only, non-mutating operation now exists, without changing that section's actual boundary — Task authorship remains direct hand/agent editing of `tasks.md` (depends on T005).

**Checkpoint**: User Story 1 is independently complete and testable — `misterspec internal next-task-id` returns the correct project-wide number, and the Skill that authors Tasks now calls it before numbering each new one.

---

## Phase 4: User Story 2 - Referência isolada a uma Task permanece inequívoca (Priority: P2)

**Goal**: Confirm that both the bare `TASK-NNN` form and the composite `SPEC-###:TASK-NNN` form keep resolving exactly as they do today — this feature does not touch reference parsing/resolution code at all (spec FR-006).

**Independent Test**: After adopting this feature, resolving a bare `TASK-NNN` reference for any Task created post-adoption returns exactly one match project-wide (spec.md US2 Independent Test).

### Tests for User Story 2

- [X] T007 [P] [US2] Add a regression case to `internal/operations/resolve_test.go` (existing `ResolveTask` tests) confirming both the bare `TASK-NNN` form and the composite `SPEC-###:TASK-NNN` form still resolve successfully after this feature's changes — no behavior change to `ParseTaskRef`/`ResolveTask` (spec FR-006, US2 Acceptance Scenario 2). This is a regression guard, not new behavior, so it is expected to already PASS against the code as of T006 — its purpose is to fail loudly if a future change accidentally breaks either form.
- [X] T008 [US2] Add an integration test (new case in `internal/cli/internalcmd/next_task_id_test.go`) asserting that two Tasks created (in fixture `tasks.md` files) using consecutive values returned by `next-task-id` for two different Specs each resolve unambiguously by their bare `TASK-NNN` number alone via `operations.ResolveTask` with no Spec context (spec US2 Acceptance Scenario 1, SC-004) (depends on T004).

**Checkpoint**: User Story 2 confirmed — bare and composite Task references are both still unambiguous, and a newly-numbered Task resolves by its bare number alone.

---

## Phase 5: User Story 3 - Projetos existentes continuam válidos sem migração (Priority: P2)

**Goal**: Confirm this feature requires zero migration/renumbering of existing Specs, and that existing cross-Spec collision tolerance is unchanged.

**Independent Test**: Running the existing collision diagnostic against the project's pre-existing Specs reports the same historical collisions, unchanged (spec.md US3 Independent Test).

### Tests for User Story 3

- [X] T009 [P] [US3] Add a regression case to `internal/cli/internalcmd/migration_check_tasks_test.go` confirming that, on a project fixture with a pre-existing cross-Spec Task-number collision, `internal migration-check-tasks` still reports that same collision unchanged, **and** `internal next-task-id` on that same fixture correctly returns the colliding number's own value + 1 (not confused or blocked by the collision) (spec FR-004/FR-005, US3 Acceptance Scenario 1) (depends on T004).
- [X] T010 [US3] Confirm the existing per-Spec duplicate-Task-number check in `internal/validation/validator_test.go` still flags a true duplicate within one Spec's own `tasks.md`, and still does **not** flag a cross-Spec collision as an error, after this feature's changes (spec FR-004, US3 Acceptance Scenario 2). Resolution: `TestValidateProject_SameTaskNumberAcrossDifferentSpecsIsNotADuplicate`, `TestValidateProject_DuplicateTaskIDWithinSameSpec`, `TestValidateEntity_SpecScopedTaskDuplicate`, and `TestValidateEntity_SpecScopedTaskDuplicate_DifferentSpecUnaffected` already assert exactly this — re-run and confirmed passing unchanged; no new test added here, since a near-duplicate case would be pure redundancy (Constitution Principle IV/VI, DRY/YAGNI). `internal/validation/validator_test.go` has zero diff in this feature.

**Checkpoint**: User Story 3 confirmed — adopting this feature is safe on a project (including MisterSpec's own repository) with pre-existing Task-number collisions; nothing is renumbered or newly flagged as invalid.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Formatting/vet cleanliness, end-to-end manual verification, and documentation.

- [X] T011 [P] Run `gofmt -l` and `go vet ./...` across every new/changed file (`internal/ids/next_task_id.go`, `internal/ids/next_task_id_test.go`, `internal/cli/internalcmd/next_task_id.go`, `internal/cli/internalcmd/next_task_id_test.go`, `internal/cli/internal.go`) and confirm both are clean.
- [X] T012 Dogfood: run the built `misterspec internal next-task-id` against MisterSpec's own real repository root (which already has known cross-Spec Task-number collisions across Specs 001-044) and confirm the returned number is exactly one greater than the true project-wide maximum — record the result in the completion report (spec.md dogfooding-adjacent verification, quickstart.md Cenário 4).
- [X] T013 Manually verify quickstart.md Cenários 1-5 end-to-end against a real built binary, including inspecting the updated `kit/skills/mister-tasks/SKILL.md` for the three specific edits named in Cenário 5.
- [X] T014 [P] Update `site/commands.html` with a new `internal next-task-id` section, mirroring the existing `internal migration-check-tasks` entry's structure and level (`<h2 id="internal-next-task-id">`).

---

## Dependencies & Execution Order

- **Phase 2 (Foundational)** blocks every user story — `ids.NextTaskID` (T002) is what the CLI command (T004) wraps.
- **User Story 1 (Phase 3)** is the MVP — independently testable and deliverable on its own once Phase 2 is done.
- **User Story 2 (Phase 4)** depends on T004 (T008 needs the real CLI command) but is otherwise independent of User Story 3.
- **User Story 3 (Phase 5)** depends on T004 (T009 needs the real CLI command) but is otherwise independent of User Story 2.
- **Polish (Phase 6)** runs after all user stories are complete.

## Parallel Execution Examples

- T001 has no sibling in its phase — Foundational is a single test-then-implementation pair.
- Within Phase 3: T003 (test) can be written in parallel with nothing else in that phase (T004/T005/T006 depend on it sequentially).
- Once T004 lands, **T007, T008, T009** (from User Story 2 and User Story 3) can all run in parallel with each other — they touch different files (`resolve_test.go`, `next_task_id_test.go`, `migration_check_tasks_test.go`) and neither story depends on the other.
- T010 (User Story 3) touches `internal/validation/validator_test.go` and can run in parallel with T007/T008/T009.
- In Polish, T011 and T014 can run in parallel; T012 and T013 are manual verification steps best run sequentially after T011 passes.

## Implementation Strategy

**MVP first**: Complete Phase 2 (Foundational) + Phase 3 (User Story 1) — this alone resolves the reported problem (the Skill no longer restarts Task numbering per Spec). Phases 4 and 5 add regression confidence that nothing else broke; they can be delivered incrementally after the MVP if desired, though for a feature this small there is no real benefit to splitting the rollout.
