# Quickstart: Validating requirement coverage and dependency cycles

This guide proves the feature works end-to-end once implemented, against a disposable temp project. It mirrors what the Go integration tests for this feature assert.

## Prerequisites

- A built `misterspec` binary reflecting this feature's changes (`go build ./cmd/misterspec`).
- A scratch project directory with `.misterspec/config.yaml` (see 031-canonical-task-identity's own `quickstart.md` for the exact minimal config shape).

## Scenario 1 — Uncovered requirement and unknown reference (User Story 1)

```sh
# SPEC-001/spec.md declares R1 and R2.
# SPEC-001/tasks.md's only Task declares "Serves: SPEC-001:R1" (R2 left uncovered).
misterspec internal validate SPEC-001
```

**Pass condition**: the findings include `uncovered_requirement` naming `R2`, and no false positive for `R1`.

```sh
# Add a second Task with "Serves: SPEC-001:R9" (R9 does not exist).
misterspec internal validate SPEC-001
```

**Pass condition**: findings now also include `unknown_requirement_reference` naming the Task and `R9`.

## Scenario 2 — Cross-Spec reference rejected (User Story 1)

```sh
# SPEC-001/tasks.md's Task declares "Serves: SPEC-002:R1" instead of its own Spec's requirement.
misterspec internal validate SPEC-001
```

**Pass condition**: findings include `cross_spec_requirement_reference` — this is never treated as valid coverage of `SPEC-002:R1`, and `SPEC-002:R1` still shows as uncovered from `SPEC-002`'s own perspective.

## Scenario 3 — Task without any Serves reference (User Story 1)

```sh
# Add a Task with no "Serves:" line at all.
misterspec internal validate SPEC-001
```

**Pass condition**: findings include `task_without_requirement` naming that Task.

## Scenario 4 — Clean coverage produces no findings (User Story 1)

```sh
# Every requirement covered by exactly one Task, every Task's Serves reference valid and same-Spec.
misterspec internal validate SPEC-001
```

**Pass condition**: no coverage-related Finding codes appear.

## Scenario 5 — Dependency cycle detected with full path (User Story 2)

```sh
# SPEC-001 depends_on SPEC-002; SPEC-002 depends_on SPEC-003; SPEC-003 depends_on SPEC-001.
misterspec internal validate
```

**Pass condition**: exactly one `dependency_cycle` finding, whose message lists all three Specs in cycle order, and every other unrelated Spec in the project still validates normally (the cycle does not halt the rest of the run).

## Scenario 6 — Self-referencing dependency (User Story 2)

```sh
# SPEC-004 depends_on SPEC-004.
misterspec internal validate SPEC-004
```

**Pass condition**: a `dependency_cycle` finding is reported for the self-reference, distinct in wording from a missing-dependency finding.

## Scenario 7 — No cycle, no false positive (User Story 2)

```sh
# A deep but acyclic dependency chain across many Specs.
misterspec internal validate
```

**Pass condition**: no `dependency_cycle` finding.

## Scenario 8 — Draft Spec is exempt from the phase gate (User Story 3)

```sh
# SPEC-005: status: draft, no tasks.md yet at all.
misterspec internal validate SPEC-005
```

**Pass condition**: no `phase_gate_blocked` finding (and no crash/error from a missing `tasks.md`).

## Scenario 9 — Non-draft Spec is blocked on incomplete coverage (User Story 3)

```sh
# Same SPEC-005, now status: ready, still no tasks.md.
misterspec internal validate SPEC-005
```

**Pass condition**: `phase_gate_blocked` finding is present, alongside the underlying `uncovered_requirement`/`task_without_requirement` findings it escalates.

## Scenario 10 — Single-Spec and project-wide validation agree (spec FR-013)

```sh
misterspec internal validate SPEC-001
misterspec internal validate | jq '.findings[] | select(.path | startswith("ai/.../SPEC-001"))'
```

**Pass condition**: the two outputs contain the same set of coverage/dependency-related findings for `SPEC-001`.

## Cleanup

```sh
rm -rf /tmp/ms-quickstart-032
```
