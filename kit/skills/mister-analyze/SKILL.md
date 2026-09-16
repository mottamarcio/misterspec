---
name: mister-analyze
description: Compare a Spec's required behavior against its actual implementation and record the verdict.
---

## Purpose

Compare a Spec's required behavior against what was actually
implemented, and record a Validation artifact stating, per requirement,
whether it is genuinely satisfied — the pipeline's own final check.

## Invocation

`/mister-analyze SPEC-###`

Requires the Spec ID to verify.

## Responsibility

Judge each requirement's satisfaction against real evidence — Plan
coverage, Task coverage, code evidence, test evidence — and report a
specific, named verdict per requirement. Never a generic pass/fail with
no explanation, and never a judgment made without evidence.

## Inputs

The Spec's own requirements; its Plan's `Requirement Coverage`; its
Tasks' own completion state and recorded evidence; the actual repository
code and test results.

## Outputs

The Spec's Validation artifact (`validation.md`), with a `Summary`, one
`### R#` subsection per requirement (`Plan coverage`, `Task coverage`,
`Code evidence`, `Test evidence`, `Result`), `Unplanned Implementation`,
`Findings`, and `Recommended Corrections`.

## Preconditions

The named Spec must have a Plan and Tasks already, with at least one
Task attempted.

## Required Context

The Spec's own requirements; the Plan's requirement coverage; the
Tasks' completion state and evidence; the actual current repository
code and test results for what each Task claims to have verified.

## Optional Context

The Spec's `depends_on` metadata (via `internal inspect`), when a
requirement's satisfaction depends on another Spec's own state.

The Context Pack returned by `internal context` (below) is a starting
point, not a boundary — this Skill remains free to read further
repository files or run further deterministic operations whenever the
pack alone is insufficient.

## Conditional Context

If a Task's recorded evidence is stale (repository code has changed
since), re-running that Task's own verification method becomes required
context before judging the requirement it serves.

## Unnecessary Context

Requirements or Tasks belonging to a different Spec entirely.

## Authority

This Skill judges pass/fail per requirement based on real evidence. It
does not decide how to fix a gap it finds — it names which artifact
layer (Spec, Plan, or implementation) is responsible, and recommends
returning to that layer's own Skill.

## Allowed Reads

The Spec, its Plan, its Tasks, and the full repository source and test
results.

## Allowed Creates

The Spec's Validation artifact, via `internal create-artifact
validation --for SPEC-###`.

## Allowed Modifications

The Validation artifact's own content, on a re-run — and, only on the
user's explicit, per-finding confirmation, and only for a finding whose
responsible layer is "implementation incomplete," appending exactly one
new `## TASK-NNN` entry to the Spec's existing Tasks artifact.

## Forbidden Mutations

Modifying the Spec, its Plan, its Tasks, or any repository source file
— this Skill only judges and records; it does not fix. The one narrow,
confirmed exception in Allowed Modifications above is an append, never a
modification, and never automatic — every other Task, and the Spec and
Plan themselves, remain forbidden to touch.

## Deterministic Operations

Use misterspec operations for every mechanical repository step.

Required operations:

- `internal resolve SPEC-###` — confirm the Spec exists and locate its
  canonical file.
- `internal inspect SPEC-###` — read the Spec's requirements and
  metadata, including `depends_on`/`supersedes` (in place of a
  dedicated cross-reference lookup, which does not exist as a separate
  operation).
- `internal context SPEC-### --intent validation` — request a budgeted
  Context Pack before broader exploration. If this fails, proceed
  using this Skill's own Optional Context above instead — it is never
  a Failure Condition.
- `internal create-artifact validation --for SPEC-###` — scaffold the
  Validation artifact at its fixed canonical location.
- `internal validate SPEC-###` — confirm the project's own structure
  (distinct from this Skill's own semantic judgment of requirement
  satisfaction) is valid.

## Procedure

1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###`.
2. Run `internal context SPEC-### --intent validation` and begin from
   its returned items. If the request fails, proceed using this
   Skill's own Optional Context above instead.
3. Read the Spec's own Plan (`Requirement Coverage`) and Tasks (their
   completion checkboxes and recorded evidence).
4. For each requirement, compare what the Plan claims, what Tasks claim
   to have done, and what the actual repository code and test results
   show. Record `Plan coverage`, `Task coverage`, `Code evidence`, `Test
   evidence`, and a `Result` (pass or fail) per requirement.
5. Note any implementation found that isn't traceable to a requirement
   (`Unplanned Implementation`).
6. Run `internal create-artifact validation --for SPEC-###`, write the
   full body, then `internal validate SPEC-###`.
7. For each requirement whose `Result` is `fail` with "implementation
   incomplete" as the responsible layer: present the gap (the
   requirement, its recorded evidence, and why it fails), then ask
   whether to append a tracking Task for it — presenting the choice with
   a recommended option and a one-sentence reason (e.g. "(Recommended)
   Yes — track this so `/mister-implement` picks it up next" vs. "No — leave it
   as a reported finding only"). On explicit confirmation, append exactly
   one new `## TASK-NNN` entry directly to the Spec's existing Tasks
   artifact (scanning existing `TASK-NNN` headings for the next number,
   the same convention `/mister-tasks` itself uses), naming the specific
   failing requirement (`Serves: SPEC-###:R#`), the verification method
   already recorded against it, and an `Origin: /mister-analyze finding
   (implementation incomplete)` line with a one-line evidence summary. On
   decline, or for a fail whose responsible layer is the Spec or Plan
   itself, make no Task-related offer or write — the finding stays a
   prose recommendation only.
8. Report completion per the Completion Contract below.

## Decision Rules

- A requirement's `Result` is `pass` only when real code and test
  evidence support it — a Task marked complete with no actual
  verification evidence is not sufficient on its own.
- When a requirement fails, identify which artifact layer is
  responsible: the Spec itself (its requirement was wrong or
  incomplete), the Plan (the strategy cannot satisfy the requirement as
  designed), or the implementation (the Plan is sound but execution is
  incomplete) — §53's branching rule; never recommend a fixed "run the
  next Skill" regardless of which layer actually needs attention.

## Interaction Rules

Report every requirement's own verdict individually — never collapse
several requirements' distinct results into one summary judgment.

## Validation Rules

Every requirement must have a recorded `Result`. A `pass` result must
cite actual code and test evidence, not merely a completed Task
checkbox.

## Failure Conditions

- The Spec has no Plan or Tasks yet: stop and recommend the missing
  earlier Skill.
- No Task has been attempted yet: stop and recommend `/mister-implement`
  first.

## Stop Conditions

Stop once every requirement has a recorded `Result` with supporting
evidence.

## Success Criteria

Every requirement has a specific, evidence-backed `Result`; any gap
found names the specific responsible artifact layer, not a generic
failure.

## Postconditions

The Spec's Validation artifact reflects its true current state — fully
satisfied, or specifically not, with the gap's own responsible layer
named.

## Idempotency

Re-running this Skill re-judges every requirement against current
evidence and replaces the Validation artifact's own content — it does
not append a second, conflicting judgment.

## Resume Behavior

If interrupted partway through, re-running the Skill re-evaluates every
requirement from scratch — this Skill's own judgment is cheap to redo
and must reflect current, not partial, evidence.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — pass (every requirement satisfied) or fail (at least
  one requirement unmet, named specifically).
- **Artifacts** — the Validation file.
- **Important findings** — any unplanned implementation, any notable
  evidence gap.
- **Attention** — the specific requirement(s) that failed and which
  artifact layer is responsible.
- **Recommended next step** — naming the exact next Skill invocation,
  or the earlier-stage Skill to return to.

## Recommended Next Step

If every requirement passes:

```text
The Spec's requirements are satisfied. No further action required
for this unit of work.
```

If a requirement fails because implementation is incomplete:

```text
/mister-implement SPEC-###
```

If a requirement fails because the Plan cannot satisfy it as designed:

```text
/mister-plan SPEC-###
```

If a requirement fails because the Spec's own requirement was wrong or
incomplete:

```text
/mister-specify FEAT-###
```

Never recommend a fixed, always-the-same next command — the
recommendation must match which artifact layer this invocation actually
found responsible (§53).

## Related Skills

`/mister-implement` — the Skill this one depends on directly.
`/mister-plan`, `/mister-specify` — the Skills a failed requirement may
recommend returning to, depending on which layer is responsible.
