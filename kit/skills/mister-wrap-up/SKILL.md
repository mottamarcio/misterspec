---
name: mister-wrap-up
description: Generate a Spec's own downstream documentation page, for a future official docs site or GitBook.
---

## Purpose

Synthesize a Spec's own artifacts, its real commit history, the
Knowledge base, the Constitution, and related Learnings into one
readable Markdown document — raw material for a future official
documentation site or GitBook, not itself a canonical, validated
project artifact.

## Invocation

`/mister-wrap-up SPEC-###`

Requires the Spec ID to document. Optional at any point after the Spec
has a Plan — independent of whether `/mister-analyze` has been run.

## Responsibility

Compose synthesized prose — the Spec's purpose, what was actually
implemented, and its outcome — never a raw concatenation or restatement
of its source files, and never fabricated content for a source that has
nothing relevant to offer.

## Inputs

The Spec's own artifacts (`spec.md`, `plan.md`, `tasks.md`,
`validation.md` if it exists); the commit range `internal
commits-since-file` returns for the Spec's own `plan.md`; the project's
Knowledge base; the Constitution; related Learning artifacts.

## Outputs

One Markdown file, `cortex/SPEC-###-<slug>.md`, at the project root —
written fresh (created or fully replaced) on every invocation.

## Preconditions

The named Spec must have a Plan already — the commit range and the
document's own account of the Spec's strategy both depend on it
existing.

## Required Context

The Spec's own `spec.md` and `plan.md`; the commit range for that
Spec's own `plan.md`.

## Optional Context

The Spec's own `tasks.md` and `validation.md`, when they exist, for a
fuller account of what was implemented and its verified outcome.

The Context Pack returned by `internal context` (below) is a starting
point, not a boundary — this Skill remains free to read further
repository files or run further deterministic operations whenever the
pack alone is insufficient.

## Conditional Context

When a Task's own recorded evidence or the Spec's own Validation
artifact names a specific behavior worth explaining, the underlying
code that satisfies it becomes useful context for describing what was
actually built, not just what was intended.

## Unnecessary Context

Another Spec's own artifacts, commits, or Tasks — this document is
scoped to exactly the one named Spec.

## Authority

This Skill decides how to present a Spec's own story — it does not
decide what the Spec required, planned, or verified. If something it
reads seems wrong or incomplete, that belongs to the artifact it came
from to fix, not something this Skill silently corrects while writing
about it.

## Allowed Reads

The Spec, its Plan, its Tasks, its Validation artifact (if any); the
Knowledge base; the Constitution; Learning artifacts; the repository's
own Git history for the Spec's own `plan.md`.

## Allowed Creates

Its own `cortex/SPEC-###-<slug>.md`, creating the `cortex/` directory
at the project root if it does not yet exist.

## Allowed Modifications

Its own `cortex/SPEC-###-<slug>.md` — replaced wholesale on every
invocation (see Idempotency), never amended in place the way every
other canonical Skill's own output is.

## Forbidden Mutations

Modifying the Spec, its Plan, its Tasks, its Validation artifact, any
Knowledge or Learning artifact, the Constitution, or any repository
source file. Writing to any path other than its own
`cortex/SPEC-###-<slug>.md` — never another Spec's own wrap-up
document.

## Deterministic Operations

Use misterspec operations for every mechanical repository step this
Skill has one for.

Required operations:

- `internal resolve SPEC-###` — confirm the Spec exists and locate its
  canonical file.
- `internal inspect SPEC-###` — read the Spec's own requirements and
  metadata.
- `internal inventory knowledge` — discover the current Knowledge base
  to judge which artifacts are relevant to this Spec.
- `internal commits-since-file --path <the Spec's own plan.md path>` —
  the commit range from the commit that first added `plan.md` (that
  Spec's own real implementation history) through the current state,
  correct regardless of what any commit message says. Returns
  `available: false` (never an error) when the project isn't a Git
  repository or `plan.md` was never committed — this is never a
  Failure Condition, only a note the document itself must carry (see
  Procedure).
- `internal context SPEC-### --intent wrap-up` — request a budgeted
  Context Pack before broader exploration. If this fails, proceed
  using this Skill's own Required Context above instead — it is never
  a Failure Condition. Record the occurrence as
  `context_fallbacks` in this task's own RunRecord once one exists,
  rather than silently absorbing the extra read
  (037-eval-quality-efficiency).

## Procedure

1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###`. If
   the Spec has no Plan yet, stop (Failure Conditions).
2. Run `internal context SPEC-### --intent wrap-up` and begin from its
   returned items. If the request fails, proceed using this Skill's
   own Required Context above instead.
3. Read the Spec's own `spec.md`, `plan.md`, and — if they exist —
   `tasks.md` and `validation.md`.
4. Run `internal commits-since-file --path <the Spec's own plan.md
   path>`. If `available: false`, note in the document that commit
   history could not be included, and continue with every other
   source — this is never a reason to stop.
5. Run `internal inventory knowledge`, and read `ai/memory/constitution.md`
   and any `ai/memory/learnings/*.md`, judging which content is
   genuinely relevant to this Spec (a semantic call — Learnings in
   particular carry no structured link to a Spec of their own).
6. Compose one synthesized Markdown document: the Spec's purpose (from
   `spec.md`), what was implemented and how (from `plan.md`/`tasks.md`
   and the commit range), its verified outcome (from `validation.md`,
   if it exists), and any relevant Knowledge/Constitution/Learnings
   context — prose, not a restatement of any one source verbatim. If
   any Task is not yet attempted, state plainly that implementation is
   incomplete and name what remains (never present an unfinished Spec
   as finished).
7. Create `cortex/` at the project root if it does not exist. Write
   (or fully replace) `cortex/SPEC-###-<slug>.md`.
8. Report completion per the Completion Contract below.

## Decision Rules

- A source with nothing relevant to this Spec is noted as "none
  found" in the document — never fabricated to fill a section.
- The document's own account of implementation completeness must match
  the Tasks artifact's actual current state — never rounded up to
  "done" when Tasks remain outstanding.

## Interaction Rules

Write for a reader who was not present during this Spec's own
development — the document must stand alone, understandable without
also having `spec.md`/`plan.md`/`tasks.md` open beside it.

## Validation Rules

The document must state plainly whether implementation is complete or
not, matching the Tasks artifact's own current state. It must note,
rather than silently omit, any source (commit history, Knowledge,
Constitution, Learnings) that had nothing relevant to offer.

## Failure Conditions

- The named Spec has no Plan yet: stop and recommend `/mister-plan`
  first — there is no meaningful commit range or implementation
  strategy to document without one.

## Stop Conditions

Stop once `cortex/SPEC-###-<slug>.md` has been written (or replaced)
reflecting the Spec's current state.

## Success Criteria

`cortex/SPEC-###-<slug>.md` exists, reads as synthesized prose (not a
raw dump of its sources), and accurately reflects the Spec's actual
implementation completeness.

## Postconditions

The Spec now has a standalone documentation page under `cortex/`,
ready to be compiled into a future official docs site or GitBook.

## Idempotency

Re-running this Skill for the same Spec always regenerates
`cortex/SPEC-###-<slug>.md` wholesale from current state — it never
appends to, amends, or duplicates a prior run's own output.

## Resume Behavior

If interrupted partway through, re-running the Skill starts over from
its own Procedure step 1 — there is no partial state to resume from,
since nothing is written until the whole document is composed.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — document created or replaced, or blocked.
- **Artifacts** — the `cortex/SPEC-###-<slug>.md` file.
- **Important findings** — which sources contributed content, and
  which had nothing relevant (Knowledge, Constitution, Learnings,
  commit history).
- **Attention** — whether implementation is complete, and, if not,
  what remains outstanding.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

If every Task is complete and `/mister-analyze` hasn't run yet:

```text
/mister-analyze SPEC-###
```

Otherwise, this Skill can simply be re-run later to refresh the
document as the Spec's own state changes:

```text
/mister-wrap-up SPEC-###
```

## Related Skills

`/mister-plan` — the Skill this one depends on directly (a Plan must
exist first).
`/mister-tasks`, `/mister-implement`, `/mister-analyze` — the Skills
whose own output this one draws on, when they exist.
