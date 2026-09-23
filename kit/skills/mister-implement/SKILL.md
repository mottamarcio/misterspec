---
name: mister-implement
description: Implement one executable Task against the real codebase, then verify it.
---

## Purpose

Implement executable Task(s) from a Spec's Tasks artifact against the
real codebase, then verify each one — the one Skill in the pipeline
that actually writes code, using the agent's own normal coding
capabilities. Either every currently executable Task in a Spec,
sequentially in one invocation, or exactly one named Task, depending on
which Invocation form is used.

## Invocation

`/mister-implement SPEC-###` or `/mister-implement SPEC-### TASK-NNN`

Both forms require the Spec ID whose Tasks are being implemented. Given
only the Spec ID, this Skill implements every currently executable Task
in that Spec, sequentially, within this one invocation. Given a Task ID
as well, it implements only that one named Task and stops — every other
Task in the Spec is left untouched. `TASK-NNN` is numbered per-Spec (no
global allocator, per `/mister-tasks`), so a Task ID is only ever
resolved within the Spec ID given alongside it.

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

The named Spec must have Tasks already. In all-tasks mode, at least one
Task must be currently executable. In named-task mode, the named Task
must exist within this Spec's own Tasks artifact and be currently
executable (or already complete, which is its own Failure Condition
rather than a precondition failure).

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

- `internal prepare SPEC-### [--task TASK-NNN]` — the single deterministic
  call that now performs resolution, readiness/dependency checking, and
  Task-scoped context assembly (served requirements' own text, declared
  scope, verification method, associated Plan sections) in one response;
  `ready: false` means stop and report the named blockers (Failure
  Conditions) — never proceed as if the Task were ready. Named-Task mode
  passes `--task`; all-tasks mode omits it and lets `prepare` select the
  next ready Task itself (034-task-oriented-context-preparation).
- `internal resolve SPEC-###` / `internal inspect SPEC-###` — still
  useful directly when only the Spec's own identity/metadata is needed
  without a Task in play.
- `internal context SPEC-### --intent implementation` — request a
  budgeted Context Pack before broader exploration when `prepare`'s own
  Task-scoped context (above) is not enough. If this fails,
  proceed using this Skill's own Required Context above instead — it
  is never a Failure Condition. Record the occurrence as
  `context_fallbacks` in this task's own RunRecord once one exists,
  rather than silently absorbing the extra read
  (037-eval-quality-efficiency).
- `internal capture-evidence SPEC-### --task TASK-NNN --origin
  automated|declared --by <who> [--result pass|fail |
  --command <cmd> --args <arg> --exec-dir <dir> --timeout <duration>]`
  — the deterministic call backing this Skill's own "mark it complete
  only once that evidence exists" requirement: computes the Task's own
  content fingerprint and current Git revision/working-tree state, and
  (`--origin automated`) runs exactly the given verification command,
  returning a `result`/`log` this Skill then records as the Task's own
  `Evidence-*:` lines (below) — it never writes those lines itself
  (041-task-evidence-fingerprint).
- `internal validate SPEC-###` — confirm the project is still
  structurally valid after the Task's own changes (this checks
  artifact structure, not the code change's own correctness — that is
  this Skill's own verification step, below).

## Procedure

1. Run `internal prepare SPEC-###` (all-tasks mode) or
   `internal prepare SPEC-### --task TASK-NNN` (named-task mode). If the
   response's `ready` is `false`, stop and report the named `blockers`
   (Failure Conditions) — this already covers "already complete" (a
   completed Task is skipped by all-tasks mode's own selection; a
   completed named Task still resolves, so check its own status before
   proceeding) and "dependencies not all complete." Otherwise its own
   `requirements`, `scope`, `verify`, and `plan_sections` fields are
   this Task's starting context — no separate `resolve`/`inspect`/
   `context` call is required to obtain them.
2. If broader exploration beyond `prepare`'s own Task-scoped context is
   needed, run `internal context SPEC-### --intent implementation` and
   begin from its returned items. If the request fails, proceed using
   this Skill's own Required Context instead.
3. Implement the change directly in the repository.
4. Verify it by the Task's own stated method (e.g. run the named `go
   test` command), then run `internal capture-evidence SPEC-###
   --task TASK-NNN --origin automated --by <this Skill/tool> --command
   <the same command> [--args ...]` to record that verification
   deterministically — never assert a result without this call backing
   it. When the Task's own method is not a runnable command (e.g. a
   manual/reviewed check), instead use `--origin declared --result
   pass|fail` naming who/what performed it.
5. If the call's own `result` is `pass`, check the Task's completion
   checkbox and write its returned fields into the Tasks artifact as
   this Task's own `Evidence-Result:`/`Evidence-Origin:`/`Evidence-By:`/
   `Evidence-CapturedAt:`/`Evidence-Command:`/`Evidence-GitRevision:`/
   `Evidence-WorkingTree:`/`Evidence-Fingerprint:`/`Evidence-Log:` lines
   (only the fields the call actually returned — declared origin omits
   `Evidence-Command:`/`Evidence-Log:`). If `result` is `fail`, stop
   here (Failure Conditions) — do not check the box, and still record
   the failing evidence fields so the failure itself is not silently
   lost.
6. Run `internal validate SPEC-###` to confirm the project's structure
   is still valid.
7. **In named-task mode**: stop and report completion per the
   Completion Contract below — never proceed to another Task in this
   invocation.
   **In all-tasks mode**: repeat from step 1 (`internal prepare` itself
   re-derives readiness from the Tasks artifact's current state, so a
   Task just completed may unblock others) until `prepare` reports no
   Task is ready, then report completion per the Completion Contract
   below.

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
into one unverified change, even when they touch related code. This
holds in all-tasks mode too: "sequential" means one Task implemented,
verified, and marked complete before the next one is even selected —
never a single unverified multi-Task change. Only the pause-and-ask-
to-re-invoke behavior between Tasks changes between the two invocation
forms, not this per-Task verification discipline.

## Validation Rules

`internal validate SPEC-###` must report zero new structural findings
after a Task's change. The Task's own verification method must have
actually run, with its result recorded as evidence.

## Failure Conditions

- The given Spec identifier does not resolve to an existing Spec (in
  either invocation form): report that the Spec was not found and do
  not attempt to resolve or implement any Task.
- A given Task ID does not exist within the named Spec's own Tasks
  artifact — including when it exists only under a *different* Spec:
  report the error and list the valid Task IDs for the named Spec,
  without modifying any files and without falling back to all-tasks
  mode.
- The named Task's own dependencies are not all complete
  (named-task mode): stop before implementing it and report which
  prerequisite Task(s) are outstanding.
- The named Task is already marked complete (named-task mode): report
  that it is already done and take no further action unless the user
  explicitly confirms they want it redone.
- No Task is currently executable (all remaining Tasks have unmet
  dependencies): stop and report which dependency is blocking. In
  all-tasks mode, this also ends the run — report every Task
  successfully implemented earlier in the same run before stopping.
- A Task's own stated scope turns out to be wrong or insufficient once
  attempted: stop, do not silently redefine the Task, and recommend
  revisiting the Plan (`/mister-plan`) or the Tasks themselves
  (`/mister-tasks`).
- Verification fails: do not mark the Task complete; report the
  failure's evidence. In all-tasks mode, stop the sequential run at
  that point — do not continue to any further Task — and still report
  which earlier Tasks in the same run completed successfully.

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

Re-running this Skill in all-tasks mode selects from the currently
executable, not-yet-complete Tasks — an already-complete Task with
recorded evidence is skipped, not re-implemented. Re-running in
named-task mode against an already-complete Task reports it as done
and takes no action, per Failure Conditions.

## Resume Behavior

If interrupted mid-implementation, re-running the Skill re-reads the
Tasks artifact's current checkbox state and resumes with whichever Task
is still incomplete, rather than restarting from the first Task.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — in named-task mode: the one Task completed and
  verified, or blocked. In all-tasks mode: every Task implemented and
  verified in this run, any Task skipped because it was already
  complete, and whether the run ended because no further Task was
  executable, a Task failed verification, or every Task in the Spec is
  now complete.
- **Artifacts** — the code files changed, and the Task ID(s) involved.
- **Important findings** — what changed and why.
- **Attention** — anything discovered that affects a later Task, Plan,
  or the Spec itself.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

Named-task mode, once the named Task is done and other Tasks remain:

```text
/mister-implement SPEC-### TASK-NNN
```

naming the next eligible Task ID — or, to implement every remaining
Task sequentially instead of naming them one at a time:

```text
/mister-implement SPEC-###
```

All-tasks mode, if it stopped early (blocked or failed) with executable
Tasks still remaining once resolved:

```text
/mister-implement SPEC-###
```

Once every Task in the Spec is complete:

```text
/mister-analyze SPEC-###
```

## Related Skills

`/mister-tasks` — the Skill this one depends on directly.
`/mister-analyze` — the next Skill in the pipeline, once all Tasks are done.
