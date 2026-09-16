# Feature Specification: Dual-Mode Implement — All Tasks or One Named Task

**Feature Branch**: `026-implement-single-task`
**Created**: 2026-09-16
**Status**: Draft
**Input**: User description: "Change `/implement` so it supports two invocation forms: `SPEC-00X` alone keeps implementing all executable tasks sequentially in one invocation (the behavior observed in Antigravity), while `SPEC-00X TASK-00X` implements only that specific task and stops (the behavior observed in Claude today). (Refined twice: first from 'TASK-00X replaces SPEC-00X' after discovering Task IDs are numbered per-Spec, not globally unique — requiring the Spec identifier to stay in the invocation; then from 'single-task only' to this dual-mode design per the user's follow-up decision to keep both behaviors available.)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run all executable tasks in one invocation (Priority: P1)

A developer wants to hand a Spec's entire remaining task list to the agent and have it work through every executable task sequentially, without pausing to ask for the same command again after each one.

**Why this priority**: This restores the "Antigravity" behavior the user explicitly wants to keep available — running `/implement SPEC-###` alone should advance the whole Spec in one sitting, not stop after a single task and wait to be re-invoked.

**Independent Test**: Can be fully tested by running the command with only a Spec ID against a `tasks.md` with several not-yet-started, dependency-satisfiable tasks, and confirming every executable task gets implemented, verified, and marked complete in one invocation.

**Acceptance Scenarios**:

1. **Given** a Spec `SPEC-012` whose `tasks.md` has three tasks with no unmet dependencies, **When** the user runs the implement command with only `SPEC-012`, **Then** the system implements, verifies, and marks complete all three tasks in that single invocation, in dependency order, without stopping to ask the user to re-run the command between tasks.
2. **Given** a Spec whose remaining tasks include one with an unmet dependency on a task outside this Spec (or on an upstream artifact that's missing), **When** the user runs the implement command with only the Spec ID, **Then** the system implements every task it can, then stops and reports which task(s) it could not reach and why, rather than silently skipping them without explanation.
3. **Given** one task in the sequence fails verification, **When** running in this all-tasks mode, **Then** the system stops the sequential run at that point, reports the failure with evidence, and does not mark that task or any tasks depending on it as complete, while still reporting which earlier tasks in the same run did complete successfully.

---

### User Story 2 - Run exactly one named task (Priority: P1)

A developer working through a Spec's task list wants to execute exactly one task by naming both the Spec and the task, without the command implementing any other task in the same run.

**Why this priority**: This is the other half of the dual-mode design and preserves the "Claude" behavior the user wants to keep available for cases where they want tighter control — e.g. reviewing each task's diff before letting the agent continue, or picking a specific out-of-order task to run next.

**Independent Test**: Can be fully tested by running the command with an existing Spec ID and a specific task ID from that Spec's `tasks.md`, and confirming only that task is executed, implemented, and marked complete — no other eligible task in the same Spec is touched.

**Acceptance Scenarios**:

1. **Given** a Spec `SPEC-012` whose `tasks.md` contains `TASK-003` in a not-started state alongside other eligible tasks, **When** the user runs the implement command with `SPEC-012 TASK-003`, **Then** the system implements only `TASK-003`, marks it complete in `SPEC-012`'s `tasks.md`, leaves every other task untouched, and reports a completion summary for that task alone.
2. **Given** the user names a task that has unmet dependencies (a prerequisite task in the same Spec is not yet marked complete), **When** running in this named-task mode, **Then** the system stops before implementing the task, tells the user which prerequisite task(s) must be completed first, and does not mark the requested task complete.
3. **Given** the user has just completed `SPEC-012 TASK-003` via the named-task form, **When** the completion summary is shown, **Then** it names the next eligible task ID(s) within that same Spec (dependencies satisfied, not yet started), and reminds the user they can also re-run with only the Spec ID to implement all remaining tasks sequentially.

---

### User Story 3 - Reject unknown or malformed identifiers (Priority: P2)

A developer mistypes a Spec ID, a task ID, or references a task ID that doesn't exist within the named Spec's `tasks.md`.

**Why this priority**: Prevents silent no-ops or the command falling back to unexpected behavior (such as implementing an arbitrary task, the wrong Spec, or defaulting to all-tasks mode on a typo) when given bad input. Since task IDs are only unique within one Spec, correctly rejecting a mismatched or unknown pair matters for both invocation forms.

**Independent Test**: Can be fully tested by invoking the command with a task ID absent from the named Spec's `tasks.md`, or with a Spec ID that doesn't exist, and confirming the system reports the error and takes no implementation action.

**Acceptance Scenarios**:

1. **Given** `SPEC-012`'s `tasks.md` has no task with ID `TASK-099`, **When** the user runs the implement command with `SPEC-012 TASK-099`, **Then** the system reports that the task ID was not found within that Spec and lists the valid task IDs for `SPEC-012`, without modifying any files and without falling back to all-tasks mode.
2. **Given** the user runs the implement command with a Spec ID that does not exist (with or without a task ID), **When** the command executes, **Then** the system reports that the Spec was not found, without attempting to resolve or run any task.

---

### User Story 4 - Task already completed (Priority: P3)

A developer re-runs the implement command naming a task that is already marked complete, either by mistake or to double check.

**Why this priority**: A lower-frequency but plausible mistake; handling it cleanly avoids duplicate work or confusing output, but it doesn't block either core invocation form the way User Stories 1–3 do.

**Independent Test**: Can be fully tested by running the named-task form against a task already marked complete in its Spec's `tasks.md` and confirming the system reports it as already complete instead of re-implementing it; and by confirming the all-tasks form simply skips already-complete tasks without comment beyond the summary.

**Acceptance Scenarios**:

1. **Given** `TASK-002` in `SPEC-012` is already marked complete, **When** the user runs the implement command with `SPEC-012 TASK-002`, **Then** the system reports the task is already complete and takes no further action, unless the user explicitly confirms they want it redone.
2. **Given** some tasks in `SPEC-012` are already complete, **When** the user runs the implement command with only `SPEC-012` (all-tasks mode), **Then** the system silently skips already-complete tasks and implements only the remaining executable ones, noting the skipped count in the completion summary.

### Edge Cases

- What happens when the named Spec has no `tasks.md` yet (Tasks were never generated)? System must direct the user to run the task-generation command first, without attempting to guess or fabricate tasks, in either invocation form.
- What happens when the named task ID exists, but in a *different* Spec than the one given? System must treat this as "task not found in the named Spec" (per User Story 3), never silently implementing the same-numbered task from another Spec.
- What happens when, in all-tasks mode, no task is currently executable at all (every remaining task has an unmet dependency)? System must stop immediately and report which dependency is blocking, the same failure reporting used today for a single blocked task.
- What happens when the task named in single-task mode is marked `[P]` (parallelizable) alongside other not-yet-started tasks? System still implements only the single named task; parallel markers only describe what *could* run concurrently, not a change to this mode's one-task scope.
- What happens when required upstream artifacts (plan.md, data-model.md, etc.) referenced by a task are missing? System must halt and report what is missing, consistent with current pre-flight checks, rather than partially implementing the task, in either mode.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The implement command MUST support two invocation forms: a Spec identifier alone (e.g. `SPEC-012`), and a Spec identifier plus a task identifier together (e.g. `SPEC-012 TASK-003`).
- **FR-002**: When invoked with only a Spec identifier, the system MUST implement, verify, and mark complete every currently executable task in that Spec sequentially within the same invocation, continuing automatically from one task to the next without pausing to ask the user to re-invoke the command.
- **FR-003**: When invoked with a Spec identifier and a task identifier together, the system MUST implement only that single named task and MUST NOT proceed to any other task in the same invocation, even if another task becomes executable afterward.
- **FR-004**: The system MUST resolve the given Spec identifier first, then — when a task identifier is also given — verify that task ID exists within that Spec's own `tasks.md`, before doing any implementation work.
- **FR-005**: The system MUST verify that a task's declared dependencies (prerequisite tasks within the same Spec) are marked complete before implementing it, in both invocation forms; in all-tasks mode this determines execution order, and in named-task mode an unmet dependency halts before implementing the named task.
- **FR-006**: In all-tasks mode, if a task fails verification, the system MUST stop the sequential run at that point, report the failure with evidence, leave that task and anything depending on it incomplete, and still report which earlier tasks in the same run succeeded.
- **FR-007**: In all-tasks mode, if no remaining task is currently executable (all have unmet dependencies), the system MUST stop and report which dependency is blocking, without silently doing nothing.
- **FR-008**: Upon successful completion of a task (in either mode), the system MUST mark it complete in the Spec's `tasks.md` and record its verification evidence, exactly as the current implementation flow does today.
- **FR-009**: Upon completion of a named-task invocation, the system MUST report a summary of the work done for that task and MUST list the task ID(s) within the same Spec that become eligible to run next, plus a reminder that re-running with only the Spec ID implements all remaining tasks sequentially.
- **FR-010**: Upon completion of an all-tasks invocation, the system MUST report a summary listing every task implemented in that run (and any skipped because already complete), in addition to any stop/failure condition encountered.
- **FR-011**: If a given task ID does not exist within the named Spec's `tasks.md` (including when it exists only under a different Spec), the system MUST report the error and list the valid task IDs for the named Spec, without modifying any files and without falling back to all-tasks mode.
- **FR-012**: If the given Spec identifier does not resolve to an existing Spec, the system MUST report that the Spec was not found and MUST NOT attempt to resolve or implement any task, in either invocation form.
- **FR-013**: If a named task is already marked complete, the system MUST report that it is already done and MUST NOT re-implement it unless the user explicitly confirms they want it redone; in all-tasks mode, already-complete tasks are silently skipped and counted in the summary.
- **FR-014**: All existing pre-flight checks that apply today (Spec/task resolution, required upstream artifacts, structural validation, verification-before-completion) MUST continue to run in both invocation forms, scoped to whichever task(s) are actually being implemented in that run.

### Key Entities

- **Spec**: A feature specification identified by a Spec ID (e.g. `SPEC-012`), owning exactly one `tasks.md` file.
- **Task**: A single unit of work identified by a task ID (e.g. `TASK-003`) that is unique only within its owning Spec's `tasks.md`, with a description, a completion state, an optional parallel marker, and zero or more dependencies on other tasks within the same Spec.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can implement an entire Spec's remaining task list in one invocation by supplying only the Spec ID, with no pause between tasks beyond what a failure or blocking dependency requires.
- **SC-002**: A user can implement exactly one named task, and only that task, by supplying the Spec ID and task ID together — no other task in the Spec is touched in that invocation.
- **SC-003**: After each named-task run, 100% of completion summaries correctly identify the next eligible task ID(s) within the same Spec, verified against that Spec's `tasks.md` dependency data.
- **SC-004**: Attempting to implement a task with unmet dependencies is blocked before any file changes occur, in 100% of cases, in both invocation forms.
- **SC-005**: A verification failure partway through an all-tasks run never leaves the failed task, or any task depending on it, marked complete.

## Assumptions

- `tasks.md` already encodes task IDs (unique within that one Spec), completion state, parallel markers, and enough structure (dependency references) to determine prerequisite relationships and a valid execution order; this feature reads that structure but does not change how `tasks.md` is generated or how Task IDs are numbered.
- Task IDs are **not** globally unique across Specs — whenever a task identifier is given, the Spec identifier must be given alongside it; a bare task identifier alone (with no Spec) is not a supported invocation form.
- The existing pre-flight and verification behavior of the implement command (resolution, dependency checks, upstream-artifact checks, evidence recording) is still desired in both invocation forms, not removed.
- Task ID formatting (`TASK-NNN`) follows whatever convention the task-generation command already produces; this feature does not standardize or change that format.
- "Sequential" in all-tasks mode means task-by-task with individual verification between tasks (never batching multiple tasks into one unverified change), consistent with the existing one-task-verified-at-a-time discipline — only the *pausing for re-invocation* behavior changes, not the per-task verification discipline.
