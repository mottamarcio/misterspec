---
name: implement
description: Implement one executable Task against the real codebase, then verify it.
---

## Purpose

Implement one executable Task from a Spec's Tasks artifact against the
real codebase, then verify it — the one Skill in the pipeline that
actually writes code, using the agent's own normal coding capabilities.

## Invocation

`/implement SPEC-###`

Requires the Spec ID whose Tasks are being implemented. Operates on
whichever Task(s) are next executable.

## Responsibility

Select an executable Task (its own dependencies already satisfied),
load the context it needs, implement it, verify it by its own stated
method, and mark it complete only once that evidence exists — never
mark a Task complete on the basis of intent alone.

## Inputs

The Spec's own Tasks artifact; the requirement(s) each Task serves; the
real repository source code.

## Outputs

A real code change satisfying the selected Task(s); the Task's
completion checkbox checked and its own evidence recorded, once
verified.

## Preconditions

The named Spec must have Tasks already. At least one Task must be
currently executable (its own dependencies, if any, already complete).

## Required Context

The selected Task's own scope, requirement references, and verification
method; the real repository source code it touches.

The Context Pack returned by `internal context` (below) is a starting
point, not a boundary — this Skill remains free to read further
repository files or run further deterministic operations whenever the
pack alone is insufficient.

## Optional Context

The Spec's own requirements and acceptance scenarios directly, for
precise verification against what the Task is actually meant to
satisfy.

## Conditional Context

If the Task depends on another Task, that other Task's own completed
state and evidence become required context before this one begins.

## Unnecessary Context

Tasks not yet executable (their own dependencies unmet) — working ahead
of dependency order risks implementing against an assumption a
dependency's own implementation later invalidates.

## Authority

This Skill implements and verifies — it does not redesign the Plan or
renumber Tasks. If a Task turns out to be unimplementable as written,
that is surfaced (Failure Conditions), not silently worked around by
changing the Task's own scope.

## Allowed Reads

The Spec, its Plan, its Tasks, and the full repository source tree.

## Allowed Creates

New repository source files, when the Task's own scope calls for one.

## Allowed Modifications

Repository source files within the selected Task's own stated scope;
the Tasks artifact's own completion checkbox and evidence for the Task
just completed.

## Forbidden Mutations

Modifying the Spec's requirements, its Plan, or a Task's own scope —
misterspec does not mediate code edits, but a Task whose own stated
scope turns out to be wrong is a Failure Condition to surface, not a
scope change to make silently. Marking a Task complete without
verification evidence.

## Deterministic Operations

Use misterspec operations for every mechanical repository step this
Skill has one for. Code changes themselves are not mediated by
misterspec — this Skill operates directly on repository source using
normal coding capabilities.

Required operations:

- `internal resolve SPEC-###` — confirm the Spec exists and locate its
  canonical file.
- `internal inspect SPEC-###` — read the Spec's requirements the
  selected Task serves.
- `internal context SPEC-### --intent implementation` — request a
  budgeted Context Pack before broader exploration. If this fails,
  proceed using this Skill's own Required Context above instead — it
  is never a Failure Condition.
- `internal validate SPEC-###` — confirm the project is still
  structurally valid after the Task's own changes (this checks
  artifact structure, not the code change's own correctness — that is
  this Skill's own verification step, below).

## Procedure

1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###`.
2. Run `internal context SPEC-### --intent implementation` and begin
   from its returned items. If the request fails, proceed using this
   Skill's own Required Context above instead.
3. Read the Spec's Tasks artifact; select one Task whose own
   dependencies are already complete.
4. Load the context that Task's own scope names.
5. Implement the change directly in the repository.
6. Verify it by the Task's own stated method (e.g. run the named `go
   test` command); record the evidence.
7. If verification succeeds, check the Task's completion checkbox and
   record its evidence in the Tasks artifact directly.
8. Run `internal validate SPEC-###` to confirm the project's structure
   is still valid.
9. Report completion per the Completion Contract below.

## Decision Rules

- A Task is not complete until its own stated verification method has
  actually been run and passed — intent to implement correctly is not
  evidence.
- If implementing a Task reveals the Plan itself cannot satisfy its
  mapped requirement without a change the Plan did not anticipate, stop
  and surface it (Failure Conditions) rather than silently expanding
  the Task's own scope.

## Interaction Rules

Implement and verify one Task at a time — do not batch multiple Tasks
into one unverified change, even when they touch related code.

## Validation Rules

`internal validate SPEC-###` must report zero new structural findings
after a Task's change. The Task's own verification method must have
actually run, with its result recorded as evidence.

## Failure Conditions

- No Task is currently executable (all remaining Tasks have unmet
  dependencies): stop and report which dependency is blocking.
- A Task's own stated scope turns out to be wrong or insufficient once
  attempted: stop, do not silently redefine the Task, and recommend
  revisiting the Plan (`/create-plan`) or the Tasks themselves
  (`/create-tasks`).
- Verification fails: do not mark the Task complete; report the
  failure's evidence.

## Stop Conditions

Stop once the selected Task(s) are verified and marked complete, or
once a Failure Condition is hit.

## Success Criteria

The implemented Task's own verification method passes; its completion
checkbox and evidence are recorded; `internal validate SPEC-###` reports
zero new structural findings.

## Postconditions

The selected Task is complete, with recorded evidence. Its dependents,
if any, are now potentially executable.

## Idempotency

Re-running this Skill selects the next executable Task — an already-
complete Task with recorded evidence is not re-implemented.

## Resume Behavior

If interrupted mid-implementation, re-running the Skill re-reads the
Tasks artifact's current checkbox state and resumes with whichever Task
is still incomplete, rather than restarting from the first Task.

## Completion Contract

Every invocation ends with a concise operational summary naming:

- **Outcome** — Task completed and verified, or blocked.
- **Artifacts** — the code files changed, and the Task's own ID.
- **Important findings** — what changed and why.
- **Attention** — anything discovered that affects a later Task, Plan,
  or the Spec itself.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

If executable Tasks remain:

```text
/implement SPEC-###
```

Once every Task is complete:

```text
/analyze SPEC-###
```

## Related Skills

`/create-tasks` — the Skill this one depends on directly.
`/analyze` — the next Skill in the pipeline, once all Tasks are done.
