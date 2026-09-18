# Quickstart: Validating task-oriented preparation

This guide proves the feature works end-to-end once implemented, against a disposable temp project. It mirrors what the Go tests for this feature assert.

## Prerequisites

- A built `misterspec` binary reflecting this feature's changes.
- A scratch project with `.misterspec/config.yaml`, a Spec with two Requirements (`R1`, `R2`), a `plan.md`, and a `tasks.md` with at least two Tasks — one serving `R1` with no dependency, one serving `R2` declaring `Depends on: TASK-001`.

## Scenario 1 — One call returns everything (User Story 1)

```sh
misterspec internal prepare SPEC-014 --task TASK-001
```

**Pass condition**: the response includes the Task, its `requirements` (with `R1`'s own text), `scope`, `verify`, and `plan_sections` — no follow-up call to `resolve`, `inspect`, or `context` is needed to obtain any of it.

```sh
misterspec internal prepare SPEC-014 --task TASK-002
```

**Pass condition**: this response's `requirements`/`scope`/`plan_sections` genuinely differ from TASK-001's own (different Requirement, different declared scope) — never the same Spec-wide bundle repeated.

## Scenario 2 — A blocked Task is never presented as ready (User Story 2)

```sh
# TASK-002 declares "Depends on: TASK-001"; TASK-001 is still unchecked.
misterspec internal prepare SPEC-014 --task TASK-002
```

**Pass condition**: `ready: false`, `blockers: ["SPEC-014:TASK-001"]`, exit code `10`.

```sh
# Mark TASK-001's checkbox complete.
misterspec internal prepare SPEC-014 --task TASK-002
```

**Pass condition**: `ready: true`, exit code `0`, full context populated.

## Scenario 3 — Task with no dependency is ready by default (User Story 2, Acceptance Scenario 3)

```sh
misterspec internal prepare SPEC-014 --task TASK-001
```

**Pass condition**: `ready: true` even though TASK-001 has no `Depends on:` line at all.

## Scenario 4 — A dependency cycle is reported, not an infinite loop (User Story 2, Acceptance Scenario 4)

```sh
# TASK-003 declares "Depends on: TASK-004"; TASK-004 declares "Depends on: TASK-003".
misterspec internal validate SPEC-014
```

**Pass condition**: exactly one `task_dependency_cycle` finding, naming both Tasks in order.

```sh
misterspec internal prepare SPEC-014 --task TASK-003
```

**Pass condition**: returns promptly (no hang), `ready: false`, blockers reflect the cycle.

## Scenario 5 — Automatic selection with a stable tie-break (User Story 3)

```sh
misterspec internal prepare SPEC-014
```

**Pass condition**: the lowest-numbered Ready Task is selected — repeatable across calls with no state change in between.

```sh
# Mark that Task complete.
misterspec internal prepare SPEC-014
```

**Pass condition**: a different Task is now selected, reflecting the new state — never the previous response reused.

## Scenario 6 — No Task is ready (User Story 3, Acceptance Scenario 3)

```sh
# Every Task in the Spec is either complete or blocked.
misterspec internal prepare SPEC-014
```

**Pass condition**: `preparation: null` with a clear, explicit message — never an empty result with no explanation, never an arbitrary pick.

## Cleanup

```sh
rm -rf /tmp/ms-quickstart-034
```
