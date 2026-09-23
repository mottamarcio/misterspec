---
name: mister-tasks
description: Decompose a Spec's Plan into ordered, independently verifiable Tasks.
---

## Purpose

Decompose a Spec's Plan into ordered, independently verifiable Tasks —
the concrete units of work `/mister-implement` will pick up one at a time.

## Invocation

`/mister-tasks SPEC-###`

Requires the Spec ID whose Plan is being decomposed.

## Responsibility

Break the Plan's implementation sequence into Tasks small enough to
verify independently, each mapped to the requirement(s) it serves and
ordered by real dependency — never a single monolithic Task, and never
Tasks whose ordering ignores an actual dependency between them.

## Inputs

The Spec's own Plan (`Requirement Coverage`, `Implementation Sequence`,
`Components Affected`).

## Outputs

The Spec's Tasks artifact (`tasks.md`), with one `## TASK-NNN —
<title>` section per task, each carrying a completion checkbox,
requirement references, dependency references, file/component scope,
and a verification method.

## Preconditions

The named Spec must have a Plan already.

## Required Context

The Plan's own `Requirement Coverage` and `Implementation Sequence`.

## Optional Context

The Spec's own requirements directly, for precise requirement
references on each Task.

The Context Pack returned by `internal context` (below) is a starting
point, not a boundary — this Skill remains free to read further
repository files or run further deterministic operations whenever the
pack alone is insufficient.

## Conditional Context

If the Plan identifies components with real ordering constraints
(e.g. a data model change before the code that depends on it), that
ordering becomes required context for Task sequencing.

## Unnecessary Context

The actual repository source code's current content — Tasks describe
*what* to do, not the resulting diff; that is `/mister-implement`'s job.

## Authority

This Skill decides task decomposition and ordering for an
already-written Plan. It does not change the Plan's own strategy — if
the Plan cannot be decomposed into verifiable Tasks as written, that is
surfaced, not silently worked around.

## Allowed Reads

The named Spec and its Plan.

## Allowed Creates

The Spec's Tasks artifact, via `internal create-artifact tasks --for
SPEC-###`.

## Allowed Modifications

The Tasks artifact's own body directly (to add `## TASK-NNN` entries —
no independent Task-ID allocator exists; Tasks are authored directly
into this one file, the same human/agent-authored boundary this
project's own entity-creation capability already draws for Task).

## Forbidden Mutations

Modifying the Spec's requirements or its Plan. Writing or modifying any
repository source file.

## Deterministic Operations

Use misterspec operations for every mechanical repository step this
Skill has one for.

Required operations:

- `internal resolve SPEC-###` — confirm the Spec exists and locate its
  canonical file.
- `internal inspect SPEC-###` — read the Spec's requirements, to
  reference precisely in each Task.
- `internal context SPEC-### --intent tasks` — request a budgeted
  Context Pack before broader exploration. If this fails, proceed
  using this Skill's own Optional Context above instead — it is never
  a Failure Condition. Record the occurrence as
  `context_fallbacks` in this task's own RunRecord once one exists,
  rather than silently absorbing the extra read
  (037-eval-quality-efficiency).
- `internal create-artifact tasks --for SPEC-###` — scaffold the Tasks
  artifact at its fixed canonical location.
- `internal validate SPEC-###` — confirm the Spec (now with Tasks) is
  still structurally valid, including catching duplicate Task numbers.

No allocator operation exists for Task IDs — `## TASK-NNN` headings are
authored directly into `tasks.md`, sequentially numbered by hand within
this one file; `internal validate` is the safety net that catches a
duplicate number, not a pre-check this Skill performs itself.

## Procedure

1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###` to
   read the Spec's requirements.
2. Run `internal context SPEC-### --intent tasks` and begin from its
   returned items. If the request fails, proceed using this Skill's
   own Optional Context above instead.
3. Read the Spec's own Plan for its `Implementation Sequence` and
   `Requirement Coverage`.
4. Run `internal create-artifact tasks --for SPEC-###` if the Tasks
   artifact does not exist yet.
5. Decompose the Plan into Tasks, each with: a title; an unchecked
   completion checkbox; a `Serves: SPEC-###:R#` line naming the
   requirement(s) it serves; a `Depends on: TASK-NNN` line naming
   another Task in this same Spec (or `Depends on: none` — or omit the
   line entirely, which means the same thing); a `Scope:` line naming
   its file/component scope; and a `Verify:` line naming how it will be
   verified (e.g. a specific `go test` invocation) — these four labels
   are what `internal prepare` and `internal validate` both parse
   mechanically (034-task-oriented-context-preparation), so use them
   verbatim rather than free prose. Number each `## TASK-NNN`
   sequentially, checking existing Tasks first so numbers are never
   reused.
6. Run `internal validate SPEC-###` to confirm no duplicate Task
   numbers, no Task referencing a nonexistent requirement, and no
   invalid or cyclic `Depends on:` declaration.
7. Report completion per the Completion Contract below.

## Decision Rules

- A Task should be independently verifiable — if verifying it requires
  another Task to already be done, that dependency must be named
  explicitly, not left implicit.
- Every requirement the Plan covers must be served by at least one
  Task; a requirement with no Task behind it is a gap, not an
  oversight to silently accept.
- When a Requirement states an explicit constraint (a specific limit,
  required format, or measurable threshold), quote that constraint's
  exact text in the Task description, in addition to the `SPEC-###:R#`
  reference — not instead of it, so it is not left to implementation-time
  discretion. A Requirement with no explicit constraint keeps the ID
  reference alone.

## Interaction Rules

If the Plan's implementation sequence is coarse, prefer decomposing it
into more, smaller Tasks over fewer, larger ones — smaller Tasks are
easier for `/mister-implement` to pick up and verify one at a time.

## Validation Rules

`internal validate SPEC-###` must report zero new findings after Tasks
are written — no duplicate Task IDs, no Task referencing a nonexistent
requirement.

## Failure Conditions

- The named Spec has no Plan yet: stop and recommend `/mister-plan`
  first.
- The Spec does not exist (`entity_not_found`): stop and report it
  plainly.

## Stop Conditions

Stop once every part of the Plan's implementation sequence maps to at
least one Task, and the Tasks artifact passes validation.

## Success Criteria

Every requirement the Plan covers is served by at least one Task; every
Task is independently verifiable; `internal validate SPEC-###` reports
zero new findings.

## Postconditions

The Spec now has an ordered, verifiable Task list ready for
`/mister-implement`.

## Idempotency

Re-running this Skill for the same Spec adds Tasks for any part of the
Plan not yet decomposed — it does not renumber or duplicate existing
Tasks.

## Resume Behavior

If interrupted partway through decomposition, re-running the Skill
reads the Tasks artifact's current content first, continuing numbering
from the highest existing `TASK-NNN` rather than restarting.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — Tasks created or extended.
- **Artifacts** — the Tasks file, with the count of Tasks added.
- **Important findings** — how the Plan was decomposed, and why; plus,
  covering the Tasks file's current full state (not only Tasks added in
  this run): every dependency relationship between Tasks (e.g. "Task
  TASK-004 depends on TASK-002"), and every group of Tasks with no
  dependency between them, named explicitly as safe to implement in
  parallel — useful to a team splitting work across people. If there is
  nothing to report (a single Task, or every Task in one strict
  sequential chain with no parallel opportunity), state that plainly
  rather than omitting the topic.
- **Attention** — any requirement not yet served by a Task.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

```text
/mister-implement SPEC-###
```

## Related Skills

`/mister-plan` — the Skill this one depends on directly.
`/mister-implement` — the next Skill in the pipeline.
