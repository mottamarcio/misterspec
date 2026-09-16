---
name: mister-features
description: Decompose a Program into one or more Features, each a coherent capability boundary.
---

## Purpose

Decompose a Program into one or more Features — coherent capability
boundaries within it, each independently specifiable and eventually
independently deliverable.

## Invocation

`/mister-features PRG-###`

Requires the Program ID to decompose.

## Responsibility

Identify genuine capability boundaries within the named Program and
create one Feature per boundary — never one Feature per arbitrary task,
and never a Feature that actually spans two unrelated capabilities.

## Inputs

The named Program's own content (problem, outcome, scope); every
Feature already created under it.

## Outputs

One or more new Feature artifacts (`FEAT-NNN`), each with `Capability`,
`User Value`, `Scope`, `Non-Goals`, `Constraints`, `Relevant Knowledge`,
and `Open Questions` sections populated, parented to the named Program.

## Preconditions

The named Program must exist and resolve successfully.

## Required Context

The Program's own scope and outcome; every Feature already created
under it (to avoid overlapping or redundant decomposition).

## Optional Context

Relevant Knowledge artifacts bearing on how the Program's capabilities
naturally divide.

## Conditional Context

If the Program's scope is broad, the user's own guidance on which
capability to decompose first becomes required context.

## Unnecessary Context

Spec-level requirement detail — that belongs to `/mister-specify`, one
level down.

## Authority

This Skill decides capability boundaries and cohesion within an
already-defined Program. It does not redefine the Program's own scope,
and does not decide individual requirements — that is `/mister-specify`'s
responsibility.

## Allowed Reads

The named Program; every Feature already under it; relevant Knowledge.

## Allowed Creates

One or more Feature artifacts, via `internal create feature --parent
PRG-###`.

## Allowed Modifications

None — this Skill only creates.

## Forbidden Mutations

Creating a Spec directly (that is `/mister-specify`'s responsibility, one
level down). Modifying the parent Program. Modifying an existing
Feature.

## Deterministic Operations

Use misterspec operations for every mechanical repository step.

Required operations:

- `internal resolve PRG-###` — confirm the parent Program exists and
  locate its canonical file before decomposing it.
- `internal children PRG-### --type feature` — discover Features
  already created under it, to avoid redundant decomposition.
- `internal create feature --parent PRG-###` — allocate the next
  Feature ID and scaffold its canonical file, correctly parented; never
  construct a `FEAT-NNN` identifier or its nested path by hand.
- `internal validate FEAT-###` — confirm each newly created Feature is
  structurally valid.

## Procedure

1. Run `internal resolve PRG-###` to confirm the Program exists.
2. Run `internal children PRG-### --type feature` to see what's already
   decomposed.
3. Identify one or more genuine capability boundaries not yet covered.
4. For each, run `internal create feature --parent PRG-###`, then write
   its body: capability, user value, scope, non-goals, constraints,
   relevant Knowledge, open questions.
5. Run `internal validate FEAT-###` for each newly created Feature.
6. Report completion per the Completion Contract below.

## Decision Rules

- A Feature should be independently specifiable — if two candidate
  Features cannot be described without constantly referencing each
  other's internals, they are likely one Feature, not two.
- A capability already covered by an existing Feature under the same
  Program is not re-created.

## Interaction Rules

If the Program's scope suggests many possible Features, decompose the
one(s) the user actually asked about first, and note in the completion
summary that further decomposition remains possible.

## Validation Rules

Every created Feature must pass `internal validate FEAT-###` with zero
findings, and must resolve `internal parent FEAT-###` back to the
correct Program.

## Failure Conditions

- The named Program does not exist (`entity_not_found`): stop and
  report it plainly.
- A described capability substantially duplicates an existing Feature:
  stop, name the existing Feature, and recommend extending it instead.

## Stop Conditions

Stop once every capability boundary the invocation was asked to
decompose has a corresponding Feature, passing validation.

## Success Criteria

Every new Feature is correctly parented, structurally valid, and
represents a genuine, non-overlapping capability boundary.

## Postconditions

The Program now has one or more Features ready for Spec decomposition.

## Idempotency

Re-running this Skill for the same Program only creates Features for
capability boundaries not already covered.

## Resume Behavior

If interrupted after `internal create feature` but before a Feature's
body is complete, re-run `internal inspect FEAT-###` to see its
partial state, complete the body, and re-validate — never create a
second Feature for the same boundary.

## Completion Contract

Every invocation ends with a concise operational summary naming:

Render this summary using structured formatting, not prose paragraphs: present **Artifacts** as a Markdown table when more than one artifact is involved (columns matching what's relevant — ID, path/type, and status or a one-line summary), or a single bullet when there is exactly one; present **Important findings** and **Attention** as bullet lists. This applies equally to a failure/stop report.

- **Outcome** — one or more Features created, or none (already
  covered).
- **Artifacts** — every Feature ID created, with a one-line capability
  summary each.
- **Important findings** — any capability boundary judgment worth
  surfacing.
- **Attention** — any open question left unresolved.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

For the primary Feature just created:

```text
/mister-specify FEAT-###
```

If the Program still has undecomposed capability, note as a secondary
option:

```text
/mister-features PRG-###
```

## Related Skills

`/mister-program` — the Skill this one depends on directly.
`/mister-specify` — the next Skill in the pipeline.
