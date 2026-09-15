# Data Model: Spec-Kit Tooling Improvements

No application data model applies — this feature edits instruction text and
YAML configuration, not Go structs or persisted entities. This documents the
conceptual shapes that change instead.

## Hook (existing, `.specify/extensions.yml`)

Unchanged schema (`extension`, `command`, `enabled`, `optional`, `prompt`,
`description`, `condition`). Two new keys added under `hooks`:

| Key | Shape | Notes |
|---|---|---|
| `before_converge` | same shape as every other `before_*` key | Populated with the same optional, `enabled: true`, `speckit.git.commit` entry every other `before_*` key already has. |
| `after_converge` | same shape as every other `after_*` key | Same pattern as every other `after_*` key. |

Every existing hook consumer's own execution contract changes (not its
schema): a mandatory hook must now be *actually invoked*, not merely
described (FR-001), and a hook-config parse failure must be *reported*, not
silently swallowed (FR-002).

## Checklist (existing, `checklists/*.md`)

Unchanged file format (`- [ ]`/`- [x]` Markdown checkboxes under headings).
Gains an explicit, documented contract:

| Checklist | Owner (who may set `[x]`) | Who may write new items | Who must never write to it |
|---|---|---|---|
| `checklists/requirements.md` | `speckit-specify` (initial), `speckit-clarify` (re-validation) | `speckit-specify` | `speckit-implement`, `speckit-checklist` |
| Any custom checklist from `speckit-checklist` | The human reviewer | `speckit-checklist` (always unchecked) | `speckit-implement`, `speckit-checklist` itself (never marks its own generated items `[x]`) |

## Convergence Finding (new, produced by `speckit-converge`, never persisted as its own file)

An in-session-only record, one per detected gap, before being rendered as a
`tasks.md` task line:

| Field | Type | Notes |
|---|---|---|
| `id` | string | Stable within one run (`F1`, `F2`, ...) — not carried across runs. |
| `source_ref` | string | What it traces to: `FR-###`, `SC-###`, `US#/AC#`, `plan: <decision>`, or a Constitution principle name. |
| `gap_type` | enum | One of `missing`, `partial`, `contradicts`, `unrequested`. |
| `severity` | enum | One of `CRITICAL`, `HIGH`, `MEDIUM`, `LOW` — a Constitution violation is always `CRITICAL`. |
| `description` | string | Human-readable, includes the observed evidence (file/area). |

Rendered, per finding, as one new task line appended to `tasks.md`:

```markdown
- [ ] T042 <imperative description> per <source-ref> (<gap-type>)
```

## Task ID matching pattern (used by both `speckit-taskstoissues` and, implicitly, `speckit-converge`'s own numbering)

`\bT\d{3,}\b` — a `T` followed by three or more digits, on a word boundary
at both ends. Matches `T001`, `T042`, `T1000`; does not match `ST001` or a
digit run embedded inside a longer number.
