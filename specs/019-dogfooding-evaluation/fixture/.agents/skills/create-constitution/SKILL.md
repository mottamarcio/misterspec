---
name: create-constitution
description: Distill the project's Knowledge base into a concise, durable Constitution of non-negotiable invariants.
---

## Purpose

Distill the project's Knowledge base into a Constitution: a short,
durable statement of the invariants the project must never violate —
distinct from the larger body of ordinary facts Knowledge already
records.

## Invocation

`/create-constitution`

Takes no required argument. Operates on the project's existing
Knowledge base.

## Responsibility

Separate durable, non-negotiable rules from ordinary facts, and record
only the former. A Constitution that repeats Knowledge verbatim, or
that states preferences rather than invariants, has failed its purpose.

## Inputs

The project's full Knowledge base — every existing `KNOW-NNN` artifact.

## Outputs

The project's Constitution (`ai/memory/constitution.md`), structured per
its required schema: `Product Invariants`, `Architecture Invariants`,
`Security Invariants`, `Data Invariants`, `Integration Invariants`,
`Quality Requirements`, `Compatibility Requirements`.

`Quality Requirements` always includes this fixed baseline, verbatim,
regardless of what Knowledge does or doesn't say — it is written, not
distilled:

```markdown
## Quality Requirements

- Code MUST follow SOLID, DRY, KISS, and YAGNI: no speculative
  abstraction, no duplicated logic, no unnecessary complexity ahead of
  a demonstrated need.
- Every change to behavior MUST be covered by tests. Prefer writing
  the test first (TDD) and specifying behavior through concrete
  scenarios before implementation (BDD).
```

Any project-specific quality facts Knowledge does support are appended
to this section after the baseline, not in place of it.

## Preconditions

At least one Knowledge artifact must exist. A Constitution distilled
from no Knowledge is not grounded in anything real.

## Required Context

Every Knowledge artifact's `Known Facts` and `Constraints` sections —
the two sections most likely to contain material that rises to the
level of an invariant.

## Optional Context

Knowledge artifacts' `Unknowns` and `Conflicts` sections, which may
reveal a contradiction the Constitution itself must resolve or
explicitly flag rather than silently pick a side.

## Conditional Context

If the Constitution already exists (this Skill is being re-run), its
current content is required context too, so this invocation amends it
rather than starting over.

## Unnecessary Context

Raw source documents directly — Knowledge has already extracted what
matters from them; re-reading raw sources here would risk introducing
facts Knowledge itself doesn't yet record.

## Authority

This Skill decides which facts are durable invariants versus ordinary,
mutable facts. It does not decide project scope or priorities (that is
`/create-program`'s responsibility) and does not invent new facts
Knowledge doesn't already support.

## Allowed Reads

Every Knowledge artifact; the existing Constitution, if one exists.

## Allowed Creates

The Constitution file itself, if it does not yet exist.

## Allowed Modifications

The Constitution file's own content, when amending it.

## Forbidden Mutations

Modifying any Knowledge artifact. Creating a Program, Feature, Spec, or
any entity with an independent ID — the Constitution itself carries no
ID (§22).

## Deterministic Operations

Use misterspec operations for every mechanical repository step.

Required operations:

- `internal inventory knowledge` — discover the current Knowledge base.
- `internal resolve KNOW-###` — confirm a Knowledge artifact's canonical
  location before citing it, rather than guessing its path.
- `internal validate` — confirm the project's overall structure remains
  valid after this invocation.

## Procedure

1. Run `internal inventory knowledge` to discover every Knowledge
   artifact.
2. For each one, resolve it via `internal resolve KNOW-###` and read its
   `Known Facts`/`Constraints`.
3. Identify which facts are durable invariants — true for the life of
   the project, not situational — and classify each into exactly one of
   the seven required Constitution sections.
4. Write or amend `ai/memory/constitution.md` directly (the Constitution
   has no independent entity ID and no `internal create` operation of
   its own — §22). Ensure `Quality Requirements` contains the fixed
   baseline from Outputs above, unconditionally — this step does not
   depend on step 3 finding anything; append any Knowledge-derived
   quality facts after it.
5. Run `internal validate` once, at the end.
6. Report completion per the Completion Contract below.

## Decision Rules

- A fact stated as a preference, a current implementation detail, or
  something reasonably expected to change is not an invariant — leave
  it in Knowledge, do not promote it.
- A genuine contradiction between two Knowledge artifacts is
  surfaced explicitly in the Constitution's own text (or left out
  pending resolution), never silently resolved by picking one side.
- Keep the Constitution concise — a long Constitution that restates
  ordinary facts has failed its purpose as durable, load-bearing memory.
- The `Quality Requirements` baseline (Outputs) is the one exception to
  "distill only what Knowledge supports" — it is written every time,
  independent of Knowledge content.

## Interaction Rules

When Knowledge conflicts on a point that would otherwise become an
invariant, do not guess — omit that invariant and name the conflict in
the completion summary's Attention section instead.

## Validation Rules

The Constitution must contain all seven required sections. It must not
contain an entity ID or duplicate any specific fact verbatim from
Knowledge — only the distilled invariant. `Quality Requirements` must
contain the fixed baseline from Outputs.

## Failure Conditions

- No Knowledge artifacts exist yet: stop and recommend
  `/create-knowledge-base` first.
- The project is not initialized (`project_not_initialized`): stop and
  report that `misterspec init` must run first.

## Stop Conditions

Stop once every current Knowledge artifact has been reviewed for
invariant-worthy content.

## Success Criteria

The Constitution exists, contains all seven required sections, its
`Quality Requirements` section carries the fixed baseline, and every
other invariant it states traces back to specific Knowledge content.

## Postconditions

The project has a durable Constitution. `/create-program` can now
proceed with it as a known constraint.

## Idempotency

Re-running this Skill after Knowledge changes amends the existing
Constitution rather than replacing it wholesale — an invariant already
recorded and still supported by Knowledge is left as-is. If the
`Quality Requirements` baseline is already present, it is left
untouched rather than duplicated or rewritten.

## Resume Behavior

If interrupted partway through, re-running the Skill re-reviews the
full Knowledge base; already-recorded invariants still supported by
Knowledge are unaffected.

## Completion Contract

Every invocation ends with a concise operational summary naming:

- **Outcome** — succeeded, amended, or found nothing to change.
- **Artifacts** — the Constitution file, created or amended.
- **Important findings** — which invariants were newly recorded.
- **Attention** — any Knowledge conflict left unresolved rather than
  silently decided.
- **Recommended next step** — naming the exact next Skill invocation.

## Recommended Next Step

Once the Constitution exists:

```text
/create-program
```

If no Knowledge exists yet, recommend `/create-knowledge-base` instead
and do not proceed.

## Related Skills

`/create-knowledge-base` — the Skill this one depends on directly.
`/create-program` — the next Skill in the pipeline.
