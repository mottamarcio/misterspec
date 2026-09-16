# Phase 1 Data Model: Dual-Mode Implement

This feature does not introduce a new persisted entity or storage format. It reads the existing `tasks.md` structure produced by `kit/skills/create-tasks/SKILL.md`. Documented here for the two entities this Skill reasons about.

## Spec

| Field | Description |
|---|---|
| ID | `SPEC-###`, resolved via `internal resolve` |
| Owns | Exactly one `tasks.md` file (Tasks artifact) |

## Task

| Field | Description |
|---|---|
| ID | `TASK-NNN` — a `## TASK-NNN — <title>` heading, unique **only within its owning Spec's `tasks.md`**; no cross-Spec uniqueness and no independent allocator (authored by hand by `/create-tasks`) |
| Completion state | Checkbox: incomplete or complete, plus recorded verification evidence once complete |
| Requirement references | `SPEC-###:R#` — which requirement(s) the task serves |
| Dependencies | Reference(s) to other Task(s) within the *same* Spec that must be complete first, or "none" |
| Parallel marker | Optional `[P]` — describes what could run concurrently in a multi-task run; does not change this Skill's own sequencing or single-task scoping |
| Verification method | The task's own stated check (e.g. a specific `go test` invocation) that must actually run and pass before the task is marked complete |

## Derived concept: Executable Task

Not a stored field — computed at Skill-invocation time by reading current Task states: a Task is **executable** when every Task it declares as a dependency is currently marked complete. Both invocation modes use this same definition:

- **All-tasks mode** (`SPEC-### only`): repeatedly selects *an* executable, not-yet-complete Task, implements and verifies it, then re-derives the executable set (now-unblocked Tasks may appear) and continues, until no executable Task remains or a Task is named explicitly.
- **Named-task mode** (`SPEC-### TASK-NNN`): checks only whether the *named* Task is executable; if not, reports the specific blocking dependency(ies) and stops without touching any other Task.

## State Transitions (per Task, either mode)

```text
not started ──(dependencies satisfied AND selected for this run)──> implementing
implementing ──(verification passes)──> complete (checkbox + evidence recorded)
implementing ──(verification fails)──> not started (run stops; reported as failure, no partial mark)
```

No Task transitions directly from "not started" to "complete" without passing through a verified "implementing" step, in either mode.
