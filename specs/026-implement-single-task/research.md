# Phase 0 Research: Dual-Mode Implement

No `[NEEDS CLARIFICATION]` markers remained in the Technical Context — every open question was resolved directly with the user during `/speckit-specify` and `/speckit-plan` (see `checklists/requirements.md` Notes). This file records those decisions in research format for traceability.

## Decision: Invocation carries an optional trailing Task ID, not a separate Skill or flag

- **Decision**: `/implement SPEC-###` and `/implement SPEC-### TASK-NNN` are two forms of the same Skill invocation — a required Spec identifier, plus an optional Task identifier.
- **Rationale**: Both forms share every precondition, Deterministic Operation, and completion-contract shape; only "which task(s) to touch, and whether to keep going after one" differs. A second Skill or a `--task` flag would duplicate the whole Procedure/Failure Conditions/Completion Contract for no behavioral gain, violating Constitution Principle IV (YAGNI).
- **Alternatives considered**:
  - *Separate Skill (`/implement-task`)*: rejected — doubles the machine-checked contract surface (`skills_content_test.go`'s per-Skill allowlist and heading set) for a distinction that's really "one optional argument."
  - *A `--single` / `--task=` flag*: rejected — the plain positional `SPEC-### TASK-NNN` form matches every other Skill's own `SPEC-###`-first convention (`create-plan`, `create-tasks`, `analyze` all take a bare Spec ID) and needs no flag-parsing convention this project doesn't otherwise use.

## Decision: Task identifiers are never accepted without their owning Spec identifier

- **Decision**: There is no supported invocation form that takes a bare `TASK-NNN` alone.
- **Rationale**: `kit/skills/create-tasks/SKILL.md` states plainly that Task IDs have "no independent Task-ID allocator" and are "numbered sequentially by hand within this one file" — i.e. per-Spec, not globally. A bare `TASK-003` is ambiguous the instant two Specs each have their own `TASK-003`.
- **Alternatives considered**:
  - *Track an "active Spec" (e.g. from the current git branch or a state file) so a bare Task ID could be resolved*: rejected outright — Constitution Principle III forbids persistent authoritative state beyond the filesystem/Git history itself, and a branch-name-derived "active Spec" would be exactly that kind of implicit secondary state, fragile the moment a user works on a branch not named after the Spec.

## Decision: "Sequential" in all-tasks mode still verifies each task before moving to the next

- **Decision**: All-tasks mode is a loop of `{ pick next executable task → implement → verify → mark complete }`, not "implement everything, then verify at the end."
- **Rationale**: The Skill's existing Interaction Rules already require this ("do not batch multiple Tasks into one unverified change") for the single-task case; the user's request was specifically about removing the *pause-and-ask-to-re-invoke* behavior between tasks, not about relaxing per-task verification. Preserving per-task verification keeps Constitution Principle V's test-first discipline meaningful at each step and keeps a mid-run failure isolated to exactly the task that failed (FR-006).
- **Alternatives considered**:
  - *Implement all tasks first, run all verifications at the end*: rejected — a failure at the end wouldn't identify which task's change caused it as cleanly, and it reintroduces the "unverified multi-task change" pattern the existing Skill explicitly forbids.

## Decision: No new `internal` operation

- **Decision**: Both modes are built entirely from the Skill's existing allowlisted operations (`resolve`, `inspect`, `context`, `validate`).
- **Rationale**: Determining "which tasks are currently executable" and "does this named task exist" both only require reading `tasks.md`'s content, which `internal inspect` / `internal context` already return. `internal validate` already catches structural issues (e.g. duplicate Task numbers) after a change. Nothing here requires ID allocation, path resolution, or fingerprinting.
- **Alternatives considered**: *A new `internal next-task` operation to compute "the next executable task" deterministically in Go*: considered, but rejected for this feature — it would be a legitimate future YAGNI candidate if task-ordering logic grows complex, but today's dependency model (declared prerequisite Task references within one file) is simple enough for the agent to read directly, and Principle IV says not to add a command before the need is demonstrated.
