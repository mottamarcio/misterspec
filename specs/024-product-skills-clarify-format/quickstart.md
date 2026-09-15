# Quickstart: Interactive Clarification and Richer Output for Product Skills

Every scenario below is manual verification against `kit/skills/` — there is
no unit-test framework for instruction text. Run each in a disposable test
project (`misterspec init` into a scratch directory), never against a real
in-progress project.

## 1. `create-specs` asks instead of silently resolving

```text
/create-feature PRG-001 "a feature with a deliberately ambiguous boundary"
/create-specs FEAT-001
```

Expect: a question presented as a full sentence, 2-4 options, one marked
"(Recommended)" with a reason. Answer it; confirm the written Requirement
reflects your choice directly (not a hedge across options).

## 2. `create-plan` asks instead of silently resolving

```text
/create-plan SPEC-001    # against a Spec whose implementation has a genuine strategic fork
```

Expect: same question/options/recommended structure, for the planning
decision instead of a requirement boundary.

## 3. Low-stakes ambiguity is still handled silently (regression check)

```text
/create-specs FEAT-001   # against a Feature with only minor, low-stakes ambiguity
```

Expect: no interruption — the Skill applies a default and records the
assumption, exactly as before this feature.

## 4. `create-tasks` quotes the constraint verbatim

```text
# Ensure SPEC-### has a requirement with an explicit constraint, e.g.
# "R3 — the export MUST complete in under 5 seconds"
/create-tasks SPEC-001
grep -A2 "SPEC-001:R3" specs/.../tasks.md
```

Expect: the Task's own description contains "under 5 seconds" verbatim, not
just the `SPEC-001:R3` reference.

## 5. `analyze` asks before appending a remediation Task

```text
# Against a Spec with a deliberately incomplete implementation:
/analyze SPEC-001
```

Expect: the failing requirement is reported, then a question asking
whether to append a tracking Task, with a recommended answer. Decline once
— confirm no Task is added and no file besides the Validation artifact
changed. Run again and accept — confirm exactly one new `## TASK-NNN` is
appended, and the Spec/Plan/every pre-existing Task are byte-for-byte
unchanged.

## 6. `analyze` never offers this for a Spec/Plan-layer gap

```text
# Against a Spec whose requirement is fundamentally unsatisfiable as
# planned (Plan-layer gap, not implementation):
/analyze SPEC-002
```

Expect: the gap is reported with the Plan named as the responsible layer,
and no offer to append a Task is presented.

## 7. Every Skill's completion report is scannable

```text
/create-program "..."
/create-feature PRG-001 "..."
/create-specs FEAT-001
/create-plan SPEC-001
/create-tasks SPEC-001
/implement TASK-001
/analyze SPEC-001
/create-constitution
/create-knowledge-base
```

For each: confirm the completion report presents `Artifacts` as a table
(2+ items) or single bullet (1 item), and `Important findings`/`Attention`
as bullet lists — never a single prose paragraph.
