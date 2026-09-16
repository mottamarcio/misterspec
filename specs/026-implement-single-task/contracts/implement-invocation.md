# Contract: `/implement` Invocation Forms

This documents the user-facing command-invocation contract for `kit/skills/implement/SKILL.md`, before and after this feature.

## Before this feature

```text
/implement SPEC-###
```

- Always auto-selects "the next executable task" in the named Spec.
- Implements and verifies exactly one Task, marks it complete, then stops.
- Recommended Next Step always re-issues the identical `/implement SPEC-###` command to advance further — the user (or agent loop) must decide whether to keep re-invoking it.
- Different agent integrations interpreted "keep re-invoking the Recommended Next Step" differently in practice (some auto-continued, some stopped and asked), which is the inconsistency this feature resolves by making both behaviors explicit invocation forms instead of implicit agent interpretation.

## After this feature

```text
/implement SPEC-###
```

- Implements, verifies, and marks complete **every currently executable Task** in the named Spec, sequentially, one verified Task after another, within this single invocation.
- Stops early only on: no executable Task remaining (reports the blocking dependency), or a Task failing verification (reports the failure and which earlier Tasks in the run did complete).
- Completion summary enumerates every Task implemented (and any skipped because already complete) in the run.

```text
/implement SPEC-### TASK-NNN
```

- Implements and verifies **only** the named Task; every other Task in the Spec is left untouched, regardless of its own eligibility.
- Fails clearly (no file changes) if: the Spec doesn't resolve, the Task ID doesn't exist within that Spec's `tasks.md`, or the named Task's dependencies aren't yet satisfied.
- If the named Task is already complete, reports that and takes no action unless the user explicitly confirms a redo.
- Completion summary names the next eligible Task ID(s) in the same Spec, and reminds the user that `/implement SPEC-###` alone will implement all remaining Tasks sequentially.

## Invariant across both forms

- A Task identifier, when given, is only ever resolved *within* the Spec identifier given alongside it — the same numeral in a different Spec is never treated as a match (Edge Case, spec.md).
- Verification always happens per-Task before that Task is marked complete — never a single unverified multi-Task change (unchanged from today's Interaction Rules).
