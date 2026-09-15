---
name: create-specs
description: Write one or more testable Specs for a Feature, with clear requirements and acceptance scenarios.
---

## Purpose

Write one or more Specs for a Feature — each a testable unit of
required behavior, with explicit requirements, acceptance scenarios,
and dependencies, ready to be planned and implemented.

## Invocation

`/create-specs FEAT-###`

Requires the Feature ID to write Specs for.

## Responsibility

Identify the required behaviors within the named Feature and write one
Spec per coherent behavior boundary, each with testable requirements —
never a Spec whose requirements are too vague to verify.

## Inputs

The named Feature's own content (capability, user value, scope); every
Spec already created under it; relevant Knowledge.

## Outputs

One or more new Spec artifacts (`SPEC-NNN`), each with `Intent`,
`Requirements` (numbered `R1`, `R2`, ...), `Acceptance Scenarios`,
`Edge Cases`, `Constraints`, and `Non-Goals` sections populated,
parented to the named Feature.

## Preconditions

The named Feature must exist and resolve successfully.

## Required Context

The Feature's own capability and scope; every Spec already created
under it (to avoid overlapping requirement boundaries).

## Optional Context

Relevant Knowledge artifacts bearing on specific required behaviors.

## Conditional Context

If the Feature's scope suggests dependencies between candidate Specs,
each dependency becomes required context for correctly recording
`depends_on`.

## Unnecessary Context

Implementation or technical-design detail — that belongs to
`/create-plan`, one stage later.

## Authority

This Skill decides requirement boundaries and acceptance criteria
within an already-defined Feature. It does not decide implementation
strategy — that is `/create-plan`'s responsibility.

## Allowed Reads

The named Feature; every Spec already under it; relevant Knowledge.

## Allowed Creates

One or more Spec artifacts, via `internal create spec --parent
FEAT-###`.

## Allowed Modifications

None — this Skill only creates.

## Forbidden Mutations

Creating a Plan, Tasks, or Validation artifact (those belong to later
Skills, once implementation is in scope). Modifying the parent Feature.
Modifying an existing Spec.

## Deterministic Operations

Use misterspec operations for every mechanical repository step.

Required operations:

- `internal resolve FEAT-###` — confirm the parent Feature exists and
  locate its canonical file.
- `internal children FEAT-### --type spec` — discover Specs already
  written under it, to avoid redundant or overlapping requirements.
- `internal create spec --parent FEAT-###` — allocate the next Spec ID
  and scaffold its canonical file, correctly parented; never construct
  a `SPEC-NNN` identifier or its nested path by hand.
- `internal validate SPEC-###` — confirm each newly created Spec is
  structurally valid.

## Procedure

1. Run `internal resolve FEAT-###` to confirm the Feature exists.
2. Run `internal children FEAT-### --type spec` to see what's already
   specified.
3. Identify one or more coherent required-behavior boundaries not yet
   covered.
4. For each, run `internal create spec --parent FEAT-###`, then write
   its body: intent, numbered requirements, acceptance scenarios (Given/
   When/Then), edge cases, constraints, non-goals, and any dependency on
   another Spec (`depends_on`).
5. Run `internal validate SPEC-###` for each newly created Spec.
6. Report completion per the Completion Contract below.

## Decision Rules

- Every requirement must be phrased so it can be verified true or
  false against an actual implementation — a requirement that cannot be
  tested is not yet ready to record as one.
- A behavior already covered by an existing Spec under the same Feature
  is not re-specified; a genuine dependency between two Specs is
  recorded via `depends_on`, not merged into one oversized Spec.

## Interaction Rules

If a requirement's exact boundary is ambiguous, prefer a narrower,
clearly testable requirement plus a named open question over a broad,
untestable one.

## Validation Rules

Every created Spec must pass `internal validate SPEC-###` with zero
findings, must resolve `internal parent SPEC-###` back to the correct
Feature, and every `depends_on` entry must name a Spec that actually
exists.

## Failure Conditions

- The named Feature does not exist (`entity_not_found`): stop and
  report it plainly.
- A described behavior substantially duplicates an existing Spec: stop,
  name the existing Spec, and recommend extending it instead.

## Stop Conditions

Stop once every required-behavior boundary the invocation was asked to
specify has a corresponding Spec, passing validation.

## Success Criteria

Every new Spec is correctly parented, structurally valid, and every one
of its requirements is independently testable.

## Postconditions

The Feature now has one or more Specs ready for planning.

## Idempotency

Re-running this Skill for the same Feature only creates Specs for
behavior boundaries not already covered.

## Resume Behavior

If interrupted after `internal create spec` but before a Spec's body is
complete, re-run `internal inspect SPEC-###` to see its partial state,
complete the body, and re-validate — never create a second Spec for the
same behavior boundary.

## Completion Contract

Every invocation ends with a concise operational summary naming:

- **Outcome** — one or more Specs created, or none (already covered).
- **Artifacts** — every Spec ID created, with a one-line intent summary
  each.
- **Important findings** — any requirement-boundary judgment worth
  surfacing, any dependency recorded between Specs.
- **Attention** — any open question left unresolved.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

For the primary Spec just created:

```text
/create-plan SPEC-###
```

If the Feature still has unspecified behavior, note as a secondary
option:

```text
/create-specs FEAT-###
```

## Related Skills

`/create-feature` — the Skill this one depends on directly.
`/create-plan` — the next Skill in the pipeline.
