# Phase 1 Data Model: Constitution Frontmatter Guarantee, and Task Dependency/Parallelism Reporting

## Constitution Frontmatter (extends `internal/artifacts.Metadata`)

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | yes | Must be `"constitution"` for `ai/memory/constitution.md` — already parsed by `Metadata.Type` today. |
| `schema_version` | int (new field) | yes | Required per `docs/architecture-specification.md` §24; not previously parsed anywhere in `internal/artifacts`. |

No `id`, `status`, or `parent` field applies to the Constitution — it is deliberately excluded from `idBearingTypes` and has no lifecycle status (it is amended in place, never has a `draft`/`active` state machine like Program/Feature/Spec).

## Constitution Finding (uses existing `internal/validation.Finding` shape — no new type)

| Condition | Code | Path |
|---|---|---|
| `ai/memory/constitution.md` does not exist | *(no Finding produced)* | — |
| File exists, frontmatter block missing or unparseable | `CodeFrontmatterMalformed` | `ai/memory/constitution.md` |
| File exists, frontmatter parses, `schema_version` absent | `CodeRequiredFieldMissing` | `ai/memory/constitution.md` |
| File exists, frontmatter parses, both fields present | *(no Finding produced)* | — |

## Task Dependency / Parallel-Safe Group (derived, not persisted)

Not a new entity — a presentation-layer derivation from data `tasks.md` already records per Task (per `create-tasks/SKILL.md`'s own existing Outputs):

| Concept | Derived from |
|---|---|
| Dependency relationship | A Task's own recorded "depends on: TASK-NNN" (or "none") |
| Parallel-safe group | The set of Tasks in the current `tasks.md` whose own dependency records name no other Task in that same set — i.e., no edge exists between any two of them |

This is computed fresh each time `/create-tasks` finishes (FR-006: reflects the file's current full state, not only newly added Tasks) — nothing is cached or written back to `tasks.md` itself.
