---
name: mister-plan
description: Design a technical implementation plan mapping a Spec's requirements to a concrete strategy.
---

## Purpose

Design an implementation plan for a Spec: a concrete technical strategy
mapping every requirement to how it will actually be built, grounded in
the real repository as it exists today.

## Invocation

`/mister-plan SPEC-###`

Requires the Spec ID to plan.

## Responsibility

Inspect the relevant parts of the repository, choose a technical
strategy, and map every one of the Spec's requirements to how that
strategy satisfies it — never a plan that leaves a requirement
unaddressed or unmapped.

## Inputs

The named Spec's requirements and acceptance scenarios; the Spec's own
dependencies (`depends_on`); the relevant existing repository code.

## Outputs

The Spec's Plan artifact (`plan.md`), with `Summary`, `Repository
Context`, `Requirement Coverage` (explicitly `R1 → ...`, `R2 → ...`,
...), `Architecture`, `Components Affected`, `Data Changes`, `API
Changes`, `Integration Changes`, `Implementation Sequence`, `Test
Strategy`, `Risks`, and `Assumptions`.

## Preconditions

The named Spec must exist, resolve successfully, and have at least one
requirement recorded.

## Required Context

The Spec's own requirements and acceptance scenarios; the actual
repository code the Spec's requirements touch.

## Optional Context

Any Spec the named one depends on (`depends_on`), via its own already-
available metadata — its own requirements and current state, not a
dedicated cross-reference lookup.

The Context Pack returned by `internal context` (below) is a starting
point, not a boundary — this Skill remains free to read further
repository files or run further deterministic operations whenever the
pack alone is insufficient.

## Conditional Context

If the Spec depends on another Spec that is not yet implemented, that
dependency's own status becomes required context for sequencing.

## Unnecessary Context

Task-level decomposition detail — that belongs to `/mister-tasks`, one
stage later.

## Authority

This Skill decides technical strategy and architecture for a
already-written Spec. It does not change the Spec's own requirements —
if the Spec cannot be satisfied without changing them, that is
surfaced, not silently worked around.

## Allowed Reads

The named Spec; any Spec it depends on; the relevant existing repository
source code.

## Allowed Creates

The Spec's Plan artifact, via `internal create-artifact plan --for
SPEC-###`.

## Allowed Modifications

None — repository source code is not touched by this Skill; only
`/mister-implement` writes code.

## Forbidden Mutations

Modifying the Spec's own requirements. Writing or modifying any
repository source file. Creating a Tasks or Validation artifact (later
Skills' responsibility).

## Deterministic Operations

Use misterspec operations for every mechanical repository step.

Required operations:

- `internal resolve SPEC-###` — confirm the Spec exists and locate its
  canonical file.
- `internal inspect SPEC-###` — read the Spec's full metadata, including
  its `depends_on`/`supersedes` fields (in place of a dedicated
  cross-reference lookup, which does not exist as a separate
  operation).
- `internal context SPEC-### --intent planning` — request a budgeted
  Context Pack before broader exploration. If this fails, proceed
  using this Skill's own Optional Context above instead — it is never
  a Failure Condition. Record the occurrence as
  `context_fallbacks` in this task's own RunRecord once one exists,
  rather than silently absorbing the extra read
  (037-eval-quality-efficiency).
- `internal create-artifact plan --for SPEC-###` — scaffold the Plan
  artifact at its fixed canonical location; never guess the path by
  hand.
- `internal validate SPEC-###` — confirm the Spec (now with a Plan) is
  still structurally valid.

## Procedure

1. Run `internal resolve SPEC-###` and `internal inspect SPEC-###` to
   read the Spec's requirements and dependencies in full.
2. Run `internal context SPEC-### --intent planning` and begin from
   its returned items. If the request fails, proceed using this
   Skill's own Optional Context above instead.
3. Inspect the real repository code relevant to those requirements.
4. Choose a technical strategy; map every requirement (`R1`, `R2`, ...)
   to how the strategy satisfies it.
5. Run `internal create-artifact plan --for SPEC-###`, then write the
   Plan's full body.
6. Run `internal validate SPEC-###`.
7. Report completion per the Completion Contract below.

## Decision Rules

- Every requirement must appear in `Requirement Coverage` — a
  requirement with no mapped implementation strategy means the plan is
  not yet complete.
- If a requirement cannot be satisfied without a change the Spec itself
  did not anticipate (e.g. a public API change with wider impact than
  expected), surface this as a risk rather than silently expanding
  scope.

## Interaction Rules

If the chosen strategy has more than one reasonable option with
materially different tradeoffs, do not resolve it silently — ask. Present
a full-sentence question, 2-4 concrete options, and mark exactly one
option "(Recommended)" with a one-sentence reason, then wait for the
user's answer before finalizing the Plan's own strategy. Once answered,
reflect the chosen strategy directly in the Plan — never hedge across
multiple options. If the user's reply doesn't map to any offered option,
ask a quick disambiguation rather than guessing.

If only one reasonable approach exists — no genuine fork with materially
different tradeoffs — proceed without interrupting the user, exactly as
before.

## Validation Rules

The Plan must map every one of the Spec's requirements in
`Requirement Coverage`. `internal validate SPEC-###` must report zero
new findings after the Plan is created.

## Failure Conditions

- The named Spec does not exist (`entity_not_found`): stop and report
  it plainly.
- The Spec has no requirements recorded yet: stop and recommend
  `/mister-specify` first.

## Stop Conditions

Stop once every requirement is mapped in `Requirement Coverage` and the
Plan passes validation.

## Success Criteria

The Plan exists, maps every requirement, and `internal validate
SPEC-###` reports zero new findings.

## Postconditions

The Spec now has an implementation plan ready for task decomposition.

## Idempotency

Re-running this Skill for the same Spec amends the existing Plan
(e.g. after a Spec's requirements change) rather than creating a second
one — `internal create-artifact` writes to the Plan's one fixed
location.

## Resume Behavior

If interrupted after `internal create-artifact plan` but before the
Plan's body is complete, re-run `internal inspect SPEC-###` to confirm
the Plan's existence, complete its body, and re-validate.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — Plan created or amended.
- **Artifacts** — the Plan file, with its path.
- **Important findings** — the chosen strategy's key decisions.
- **Attention** — any risk or assumption worth flagging.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

```text
/mister-tasks SPEC-###
```

## Related Skills

`/mister-specify` — the Skill this one depends on directly.
`/mister-tasks` — the next Skill in the pipeline.
