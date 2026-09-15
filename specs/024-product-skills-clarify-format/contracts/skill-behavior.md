# Contract: Modified Product Skill Behavior

Reconciled against the actual implementation (024-product-skills-clarify-
format) — every guarantee below was checked against the final wording in
each `kit/skills/*/SKILL.md` file. These are behavioral guarantees a coding
agent following `kit/skills/` must satisfy — not an HTTP/CLI schema.

## `create-specs` (User Story 1)

- Given a requirement boundary with multiple reasonable interpretations
  and materially different implications, the Skill MUST present a
  question (full sentence, ends in `?`), 2-4 concrete options, and exactly
  one marked "(Recommended)" with a one-sentence reason, and MUST wait for
  the user's answer before finalizing that requirement.
- Given a low-stakes ambiguity with an obvious default, behavior is
  unchanged: the Skill applies the default and records the assumption
  without interrupting the user.
- Given the interactive question quota is reached with a real ambiguity
  still unresolved, the Skill MUST record it under the new Spec's own
  `## Unresolved Questions` section.

## `create-plan` (User Story 2)

- Given a Plan decision with more than one reasonable strategy and
  materially different tradeoffs, the Skill MUST present the same
  question/options/recommended structure as `create-specs` and wait for
  the user's answer.
- Given only one reasonable strategy, behavior is unchanged.

## `create-tasks` (User Story 3)

- Given a Requirement with an explicit constraint (limit, format, or
  measurable threshold), a Task touching that Requirement MUST quote the
  constraint's exact text in its own description, in addition to the
  `SPEC-###:R#` reference.
- Given a Requirement with no explicit constraint, behavior is unchanged
  (ID reference only).

## `analyze` (User Story 4)

- Given a `Result: fail` whose responsible layer is "implementation
  incomplete," the Skill MUST ask the user whether to append a tracking
  Task, presenting the choice with a recommended option, before any write.
- The Skill MUST append a Task only on the user's explicit confirmation
  for that specific finding — never automatically, and never for more
  findings than the user confirmed.
- Any append MUST be the Skill's only write for that finding: no existing
  Task, the Spec, or the Plan may be modified.
- Given a `Result: fail` whose responsible layer is the Spec or Plan
  itself, the Skill MUST NOT offer to append a Task — it continues to
  recommend the responsible Skill in prose, exactly as before this
  feature.

## All 9 Skills (User Story 5)

- Every completion report (success, failure, or stop) MUST render
  `Artifacts` as a Markdown table when 2 or more artifacts are involved,
  or a single bullet when exactly 1, and MUST render `Important findings`
  and `Attention` as bullet lists — never as an undifferentiated prose
  paragraph.

## Behavioral guarantees carried over (not re-tested here, only extended)

- Every Skill's existing `Allowed Reads/Creates/Modifications`,
  `Forbidden Mutations`, `Preconditions`, and `Failure Conditions` remain
  in force, with `analyze`'s own `Allowed Modifications`/`Forbidden
  Mutations` narrowly extended exactly as documented in data-model.md —
  no other Skill's mutation boundary changes at all.
- `internal validate` still runs the same structural checks after any
  artifact write, including a Task `analyze` appends — no exemption is
  introduced for `analyze`-originated Tasks.
