---

description: "Task list for feature implementation"
---

# Tasks: Dual-Mode Implement — All Tasks or One Named Task

**Input**: Design documents from `/specs/026-implement-single-task/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/implement-invocation.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First, NON-NEGOTIABLE), a test task is included and comes first (Phase 2) rather than per-story. Justification: every User Story here edits sections of the same single Markdown file (`kit/skills/implement/SKILL.md`), whose only mechanically testable property is structural/content conformance, already covered by one shared Go test suite (`internal/example/skills_content_test.go`). A per-story unit test would just re-run the same assertion; instead, one new assertion is written first (and must fail against the current file) before any `SKILL.md` edit lands, and each story's own acceptance scenarios are validated behaviorally via `quickstart.md` (Polish phase), since runtime behavior here is agent-interpreted prompt content, not unit-testable code.

**Organization**: Tasks are grouped by user story to enable independent implementation and validation of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Single Go project with an embedded content kit. All Skill-content edits are under `kit/skills/`; the one Go test touched is under `internal/example/`. No `src/`/`backend/`/`frontend/` split applies (see plan.md Project Structure).

---

## Phase 1: Setup

**Purpose**: Confirm the working baseline before editing shared Skill content.

- [X] T001 Read the current `kit/skills/implement/SKILL.md` in full and note every section that will be touched (`Invocation`, `Procedure`, `Decision Rules`, `Interaction Rules`, `Failure Conditions`, `Completion Contract`, `Recommended Next Step`) so later edits preserve every other section verbatim, per `requiredSkillHeadings` order in `internal/example/skills_content_test.go`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Lock in the shared contract (both invocation forms) as a machine-checked assertion, and update the `Invocation` section that both User Story 1 and User Story 2 depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T002 Add a new test function in `internal/example/skills_content_test.go` (near `TestSkillsContent_PlanTasksImplementAnalyze`) that reads `kit/skills/implement/SKILL.md`'s `Invocation` section (via the existing `extractSection` helper) and asserts it contains both `SPEC-###` and `SPEC-### TASK-NNN` as documented invocation forms. Run `go test ./internal/example/... -run TestSkillsContent` and confirm this new assertion **fails** against the current file content (test-first, Constitution Principle V).
- [X] T003 Update the `## Invocation` section of `kit/skills/implement/SKILL.md` per `contracts/implement-invocation.md`'s "After this feature" section: document `/implement SPEC-###` (all currently executable Tasks, sequentially) and `/implement SPEC-### TASK-NNN` (only that Task) as the two supported forms. Run `go test ./internal/example/... -run TestSkillsContent` and confirm T002's assertion now **passes**.

**Checkpoint**: Foundation ready — the Invocation contract is documented and machine-checked; user story implementation can now begin.

---

## Phase 3: User Story 1 - Run all executable tasks in one invocation (Priority: P1) 🎯 MVP

**Goal**: `/implement SPEC-###` alone implements, verifies, and marks complete every currently executable Task in the Spec sequentially, in one invocation, without pausing to ask the user to re-invoke it between tasks.

**Independent Test**: Follow `quickstart.md` step 2 — run `/implement SPEC-0XX` against a Spec with 3+ executable Tasks and confirm all are implemented, verified, and marked complete in one invocation, with a summary enumerating each.

### Implementation for User Story 1

- [X] T004 [US1] In `kit/skills/implement/SKILL.md`'s `## Procedure` section, rewrite the task-selection step to branch on whether a Task ID was given: when only `SPEC-###` is given, loop — select an executable, not-yet-complete Task, implement it, verify it, mark it complete, then re-derive the executable set from `tasks.md`'s current state and repeat — continuing automatically until no executable Task remains, without stopping to request re-invocation.
- [X] T005 [US1] In `kit/skills/implement/SKILL.md`'s `## Decision Rules` and `## Interaction Rules` sections, state explicitly that all-tasks mode remains one-Task-verified-at-a-time — never batching multiple Tasks into one unverified change — quoting spec.md's exact constraint: "never a single unverified multi-Task change" (Assumptions, spec.md).
- [X] T006 [US1] In `kit/skills/implement/SKILL.md`'s `## Failure Conditions` section, add: (a) if no remaining Task is currently executable (all have unmet dependencies), stop and report which dependency is blocking; (b) if a Task fails verification mid-run, stop the sequential run at that point, report the failure with evidence, leave that Task and anything depending on it incomplete, and still report which earlier Tasks in the same run succeeded (FR-006, FR-007, spec.md).
- [X] T007 [US1] In `kit/skills/implement/SKILL.md`'s `## Completion Contract` section, update the **Outcome**/**Artifacts** guidance so an all-tasks-mode run's summary enumerates every Task implemented in that run, and any Task skipped because it was already complete (FR-010, spec.md).

**Checkpoint**: At this point, User Story 1 (all-tasks mode) should be fully usable and independently testable via `quickstart.md` step 2.

---

## Phase 4: User Story 2 - Run exactly one named task (Priority: P1)

**Goal**: `/implement SPEC-### TASK-NNN` implements and verifies only that one named Task and stops; every other Task in the Spec is left untouched.

**Independent Test**: Follow `quickstart.md` step 3 — run `/implement SPEC-0XX TASK-002` and confirm only that Task is implemented and marked complete, with every other Task's checkbox untouched.

### Implementation for User Story 2

- [X] T008 [US2] In `kit/skills/implement/SKILL.md`'s `## Procedure` section (same branch point as T004), add the named-task path: when a Task ID is also given, resolve the Spec first, then verify that Task ID exists within that Spec's own `tasks.md`, implement and verify only that Task, and stop — do not proceed to any other Task even if one becomes executable afterward (FR-003, FR-004, spec.md).
- [X] T009 [US2] In `kit/skills/implement/SKILL.md`'s `## Failure Conditions` section, add: if the named Task's dependencies aren't yet satisfied, stop before implementing it and report which prerequisite Task(s) are outstanding (FR-005, spec.md).
- [X] T010 [US2] In `kit/skills/implement/SKILL.md`'s `## Completion Contract` section, update the named-task-mode summary to name the next eligible Task ID(s) within the same Spec, and add a reminder that re-running with only the Spec ID (`/implement SPEC-###`) implements all remaining Tasks sequentially (FR-009, spec.md).

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently, exactly as documented in `contracts/implement-invocation.md`.

---

## Phase 5: User Story 3 - Reject unknown or malformed identifiers (Priority: P2)

**Goal**: A nonexistent Spec ID, a Task ID not present in the named Spec (including one that exists only under a *different* Spec), or a Spec ID given without a required Task ID in named-task context, are all rejected clearly with no file changes and no silent fallback to a different mode.

**Independent Test**: Follow `quickstart.md` step 4 — run `/implement SPEC-0XX TASK-999` and confirm the system reports the Task wasn't found in that Spec, lists valid Task IDs, and makes no file changes.

### Implementation for User Story 3

- [X] T011 [US3] In `kit/skills/implement/SKILL.md`'s `## Failure Conditions` section, add: if the given Spec identifier does not resolve to an existing Spec (in either invocation form), report that the Spec was not found and do not attempt to resolve or implement any Task (FR-012, spec.md).
- [X] T012 [US3] In `kit/skills/implement/SKILL.md`'s `## Failure Conditions` section, add: if a given Task ID does not exist within the named Spec's `tasks.md` — including when it exists only under a different Spec — report the error and list the valid Task IDs for the named Spec, without modifying any files and without falling back to all-tasks mode (FR-011, spec.md Edge Cases).

**Checkpoint**: All error paths for both invocation forms now report clearly with zero file changes, per `contracts/implement-invocation.md`'s invariant section.

---

## Phase 6: User Story 4 - Task already completed (Priority: P3)

**Goal**: Re-running the command against an already-complete Task reports that instead of re-implementing it — cleanly in named-task mode, and silently (with a summary count) in all-tasks mode.

**Independent Test**: Follow `quickstart.md`-style manual check — run named-task mode against an already-complete Task and confirm it reports "already complete" and takes no action; run all-tasks mode against a Spec with some already-complete Tasks and confirm they're skipped and counted in the summary.

### Implementation for User Story 4

- [X] T013 [US4] In `kit/skills/implement/SKILL.md`'s `## Failure Conditions` (or `## Decision Rules`, whichever the existing structure best fits) section, add: if the named Task is already marked complete, report that and take no further action unless the user explicitly confirms they want it redone (FR-013, spec.md).
- [X] T014 [US4] In `kit/skills/implement/SKILL.md`'s `## Procedure` (all-tasks loop from T004), add: already-complete Tasks are silently skipped during the executable-set derivation, and the skipped count is included in the Completion Contract's summary (FR-013, spec.md).

**Checkpoint**: All four user stories are now independently functional and documented consistently within `kit/skills/implement/SKILL.md`.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Keep sibling Skills consistent and validate the finished feature end-to-end.

- [X] T015 [P] Review `kit/skills/create-tasks/SKILL.md`'s `/implement SPEC-###` example (Recommended Next Step) and confirm its wording still reads correctly under the dual-mode design; adjust only if it implies single-task-only behavior.
- [X] T016 [P] Review `kit/skills/analyze/SKILL.md`'s `/implement SPEC-###` example and confirm the same as T015.
- [X] T017 Run `go test ./internal/example/...` in full and confirm no regressions in cross-agent rendering (`multi_agent_skill_integration_quickstart_test.go`) or other `TestSkillsContent_*` assertions.
- [X] T018 Execute `quickstart.md`'s manual validation steps 2–5 end-to-end against a scratch Spec and confirm every acceptance scenario in `spec.md` (User Stories 1–4) passes as described. NOTE: performed as a structural trace against a scratch `tasks.md` (3-task dependency chain) walked through the updated `Procedure`/`Failure Conditions`/`Completion Contract` text for every acceptance scenario in User Stories 1–4 — all traced correctly with no gaps found. A live multi-turn `/implement` invocation against a real scratch project (true end-to-end) requires a separate agent session and was not performed here.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories (both invocation forms share the same `Invocation` section and machine-checked contract).
- **User Stories (Phase 3–6)**: All depend on Foundational phase completion.
  - US1 and US2 (both P1) touch the same `Procedure` branch point and the same file — implement sequentially, not in parallel, even though both are P1.
  - US3 and US4 build on the `Failure Conditions`/`Procedure` structure US1/US2 establish — implement after US1/US2.
- **Polish (Phase 7)**: Depends on all user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2). Shares `kit/skills/implement/SKILL.md` with US2 — not file-parallel with it, but conceptually independent (all-tasks mode never needs the named-task branch).
- **User Story 2 (P1)**: Can start after Foundational (Phase 2). Same file as US1 — sequence after US1 to avoid edit conflicts, even though both are P1.
- **User Story 3 (P2)**: Extends the `Failure Conditions` section both US1 and US2 also touch — sequence after both.
- **User Story 4 (P3)**: Extends both the `Procedure` (all-tasks skip logic) and Failure/Decision handling — sequence after US1–US3.

### Parallel Opportunities

- T015 and T016 (Phase 7) touch different files (`create-tasks/SKILL.md`, `analyze/SKILL.md`) and can run in parallel.
- No other tasks are file-parallel: every other task in this feature edits `kit/skills/implement/SKILL.md`, `internal/example/skills_content_test.go` (T002, ahead of T003), or is a validation-only step — all sequential by file-conflict, not by design choice.

---

## Parallel Example: Phase 7 (Polish)

```bash
# Launch both sibling-Skill wording reviews together (different files):
Task: "Review kit/skills/create-tasks/SKILL.md's /implement SPEC-### example"
Task: "Review kit/skills/analyze/SKILL.md's /implement SPEC-### example"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — the shared Invocation contract and its test).
3. Complete Phase 3: User Story 1 (all-tasks mode).
4. **STOP and VALIDATE**: Run `quickstart.md` step 2 against a scratch Spec.
5. This alone restores the "Antigravity" behavior the user asked to keep — a usable MVP even before User Story 2 lands.

### Incremental Delivery

1. Setup + Foundational → shared Invocation contract locked in and machine-checked.
2. Add User Story 1 → validate independently → all-tasks mode usable (MVP).
3. Add User Story 2 → validate independently → named-task mode usable.
4. Add User Story 3 → validate independently → error paths hardened.
5. Add User Story 4 → validate independently → idempotent re-runs handled.
6. Polish → sibling-Skill consistency + full regression pass.

### Note on "Parallel Team Strategy"

Not applicable in the usual multi-developer sense here: because every core task (T002–T014) edits one of two shared files (`kit/skills/implement/SKILL.md`, `internal/example/skills_content_test.go`), this feature is realistically a single-owner, sequential effort through Phase 6, with only Phase 7's T015/T016 offering genuine parallelism.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- T002 must fail before T003 makes it pass (test-first, Constitution Principle V).
- Every task after T003 edits `kit/skills/implement/SKILL.md` — commit after each user-story phase (T007, T010, T012, T014) rather than after every single task, to keep the file's diff reviewable per phase.
- Avoid: editing `Invocation`, `Procedure`, or `Failure Conditions` out of the `requiredSkillHeadings` order `internal/example/skills_content_test.go` enforces.
