---
name: mister-program
description: Define a new Program capturing one initiative's problem, outcome, and scope.
---

## Purpose

Define a new Program: the top-level unit of work capturing one
initiative's problem, desired outcome, and scope, grounded in the
project's existing Knowledge and Constitution.

## Invocation

`/mister-program`

Takes no required argument beyond the initiative to describe, provided
by the user in conversation.

## Responsibility

Turn a described initiative into a well-formed Program artifact with a
clear problem statement, outcome, scope, and success criteria — and
detect when it would duplicate an existing Program's already-covered
problem space instead of silently creating a redundant one.

## Inputs

The user's description of the initiative; the project's existing
Knowledge base and Constitution; every currently existing Program (to
check for overlap).

## Outputs

One new Program artifact (`PRG-NNN`), with `Problem`, `Users and
Stakeholders`, `Desired Outcome`, `Scope`, `Non-Goals`, `Constraints`,
`Success Criteria`, `Relevant Knowledge`, and `Open Questions` sections
populated.

## Preconditions

The project must be initialized. A Constitution should exist (not
strictly required, but a Program defined before one exists risks
missing a known constraint).

## Required Context

The user's own description of what the initiative is meant to
accomplish; the current count and identity of existing Programs
(`internal status`).

## Optional Context

The project's Constitution and relevant Knowledge artifacts, to ground
scope and constraints in what's already known rather than assumptions.

## Conditional Context

If an existing Program appears to cover an overlapping problem space,
that Program's own content becomes required context for the overlap
decision.

## Unnecessary Context

Feature- or Spec-level detail — a Program's scope is deliberately
coarser than either.

## Authority

This Skill decides problem framing, outcome, and scope boundaries. It
does not decide how the Program is decomposed into Features — that is
`/mister-features`'s responsibility.

## Allowed Reads

The project's Knowledge base, Constitution, and every existing Program.

## Allowed Creates

One Program artifact, via `internal create program`.

## Allowed Modifications

None — this Skill only creates; amending an existing Program's scope
after the fact is a separate, deliberate action, not an implicit side
effect of this Skill.

## Forbidden Mutations

Creating a Feature, Spec, or any other entity type. Modifying an
existing Program.

## Deterministic Operations

Use misterspec operations for every mechanical repository step.

Required operations:

- `internal status` — see how many Programs already exist and get a
  structural overview before deciding whether this is genuinely new.
- `internal create program` — allocate the next Program ID and scaffold
  its canonical file; never construct a `PRG-NNN` identifier by hand.
- `internal validate PRG-###` — confirm the newly created Program is
  structurally valid.

## Procedure

1. Run `internal status` to see the project's current structural
   overview.
2. Review existing Programs for problem-space overlap with the
   initiative being described.
3. If no genuine overlap exists, run `internal create program` to
   allocate the new Program's ID and canonical file.
4. Write its body: problem, users/stakeholders, desired outcome, scope,
   non-goals, constraints, success criteria, relevant Knowledge, and any
   open questions.
5. Run `internal validate PRG-###` to confirm structural validity.
6. Report completion per the Completion Contract below.

## Decision Rules

- If an existing Program's problem statement already substantially
  covers the described initiative, do not create a new one — surface
  the overlap instead (Failure Conditions).
- Scope should name what's explicitly excluded (`Non-Goals`) as
  clearly as what's included.

## Interaction Rules

When the initiative's boundaries are ambiguous from the user's own
description, prefer asking a clarifying question over guessing scope —
a Program's scope shapes every Feature and Spec decomposed from it.

## Validation Rules

The created Program must pass `internal validate PRG-###` with zero
findings. Every required section must be present and non-empty.

## Failure Conditions

- An existing Program already substantially covers the same problem
  space: stop, name the existing Program, and recommend extending or
  reusing it instead of creating a duplicate.
- The project is not initialized (`project_not_initialized`): stop and
  report that `misterspec init` must run first.

## Stop Conditions

Stop once the new Program's body is complete and passes validation, or
once a genuine overlap with an existing Program is identified.

## Success Criteria

Exactly one new Program exists, fully populated, passing
`internal validate PRG-###` cleanly, with no undetected overlap against
an existing Program.

## Postconditions

The Program exists and is ready for decomposition into Features.

## Idempotency

Invoking this Skill again for a genuinely different initiative creates
a second, independent Program — invoking it again for the same
initiative should be caught by the overlap check and refused.

## Resume Behavior

If interrupted after `internal create program` but before the body is
fully written, re-running `internal inspect PRG-###` shows the partial
state; complete the body and re-run `internal validate PRG-###` rather
than creating a second Program.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — created, or refused due to overlap.
- **Artifacts** — the new Program's ID, or the existing one it
  overlaps with.
- **Important findings** — the problem, outcome, and scope recorded.
- **Attention** — any open question left unresolved.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

```text
/mister-features PRG-###
```

## Related Skills

`/mister-constitution` — should generally precede this Skill.
`/mister-features` — the next Skill in the pipeline.
