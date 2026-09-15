# Data Model: Interactive Clarification and Richer Output for Product Skills

No new persisted entity, field, or template section is introduced — every
existing artifact schema (Feature, Spec, Plan, Tasks, Validation) is
unchanged. This documents the conceptual shapes involved.

## Clarification Question (new concept — in-session only, never persisted directly)

| Field | Notes |
|---|---|
| Question text | Full interrogative sentence, ends in `?` |
| Options | 2-4 concrete answers |
| Recommended | Exactly one option, plus a one-sentence reason |
| Resolution | Either integrated directly into the requirement/decision text, or — if unasked due to quota — recorded under the Spec's existing `## Unresolved Questions` section (Feature/Plan use their own existing equivalent fields) |

Not a new entity type; nothing about `ids`/`artifacts` path resolution
changes. Applies to `create-specs` (Spec-level requirement ambiguity) and
`create-plan` (Plan-level strategic forks).

## `analyze`'s widened (narrowly) mutation boundary

| Section (in `analyze/SKILL.md`) | Before | After |
|---|---|---|
| `Allowed Modifications` | "The Validation artifact's own content, on a re-run." | Adds: "…, and — only on the user's explicit, per-finding confirmation, and only for a finding whose responsible layer is 'implementation incomplete' — appending exactly one new `## TASK-NNN` entry to the Spec's existing Tasks artifact." |
| `Forbidden Mutations` | "Modifying the Spec, its Plan, its Tasks, or any repository source file…" | Clarifies: "...(the one narrow, confirmed exception in Allowed Modifications above is an append, never a modification, and never automatic)." |

No new `internal` command. The append reuses the exact same
direct-file-authoring convention `create-tasks/SKILL.md` already documents
for the same artifact (`## TASK-NNN` headings, next sequential number
scanned from existing headings — no independent Task-ID allocator exists).

## Task entry appended by `analyze` (shape — matches `create-tasks`'s own existing format, no new fields)

```markdown
## TASK-NNN

- [ ] <imperative description of the specific remaining work>
- Serves: SPEC-###:R# (the requirement analyze found failing)
- Depends on: none (or a named prior Task, if evident from the finding)
- Verification: <the same verification method analyze already recorded evidence against>
- Origin: /analyze finding (implementation incomplete) — <one-line evidence summary>
```

`Origin` is the only new line vs. a normally-authored Task, and exists so
a later reader can distinguish an `analyze`-appended Task from one written
during ordinary `/create-tasks` planning — informational only, not a new
formal field enforced by `internal validate`.

## Completion Contract rendering (all 9 Skills)

No schema change to the `Completion Contract` section's own five named
buckets (`Outcome`, `Artifacts`, `Important findings`, `Attention`,
`Recommended next step`). Only the *rendering* changes:

| Bucket | Rendering when it applies |
|---|---|
| `Artifacts` | Markdown table (columns: ID, path/type, status/summary) when 2+ artifacts; a single bullet when exactly 1 |
| `Important findings` | Bullet list |
| `Attention` | Bullet list |
| `Outcome`, `Recommended next step` | Unchanged — short, single-line statements; not tabular by nature |
