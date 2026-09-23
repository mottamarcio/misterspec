# Phase 1 Data Model: Modelo canônico de tarefas e identidade composta

This feature does not add persisted storage (Constitution Principle III — filesystem is the sole source of truth; no database, no counter). "Entities" below are in-memory Go types derived at scan/resolve time from existing Markdown files, extending `internal/ids`'s existing `EntityID` model — not a new storage schema.

## TaskID (composite identity)

Represents a Task unambiguously, wherever it is referenced outside its own Spec's context (spec FR-001, FR-002).

| Field | Type | Description |
|---|---|---|
| `Spec` | `ids.EntityID` (`Type == ids.Spec`) | The owning Spec's identity. Always present and valid — a `TaskID` cannot exist without a Spec. |
| `Local` | `ids.EntityID` (`Type == ids.Task`) | The task's own local identity (`TASK-NNN`), unique only within `Spec`. |

**Derived behavior**:
- `String()` renders the canonical composite form: `"<Spec>:<Local>"`, e.g. `"SPEC-014:TASK-003"`.
- Two `TaskID` values are equal iff both `Spec` and `Local` are equal — matching numbers under different Specs are explicitly *not* the same identity (spec FR-001, User Story 1).

**Validation rules** (spec FR-002, FR-010):
- `Spec` MUST be a syntactically valid `SPEC-###` per the project's configured ID width.
- `Local` MUST be a syntactically valid `TASK-###` per the project's configured ID width.
- A malformed composite string (wrong separator, missing half, invalid syntax on either half) is rejected as `ErrInvalidIDSyntax` (reusing the existing sentinel from `internal/ids/ids.go`), not silently truncated to one half.

## Task (existing entity, identity clarified)

No change to the entity's own attributes (spec, out of scope) — only to how it is identified and located, per the existing `artifacts.Metadata`/`operations.ResolvedLocation` shapes already used by `Inspect`.

| Field | Type | Description |
|---|---|---|
| `ID` | `TaskID` (new) | Replaces the implicit "just a `TASK-NNN` number" identity assumption in code paths that resolve tasks project-wide. Within a single Spec's own `tasks.md`, the short local form (`TASK-003`) remains the valid, sufficient reference (spec FR-009). |
| `Status` | `string` (existing, unchanged) | `"pending"` / `"complete"`, derived from the Markdown checkbox, per `operations.inspectTask`. |
| `Parent` | `*ids.EntityID` (existing, unchanged) | Derived from the owning `tasks.md`'s `for:` frontmatter field — already equivalent to `TaskID.Spec` in practice; unchanged by this feature. |

## TaskScanEntry (internal scan representation)

Not user-facing; the shape `internal/ids/scan.go`'s rescoped Task scanning produces, replacing today's flat `map[int][]string` claims map for the Task entity type.

| Field | Type | Description |
|---|---|---|
| `Spec` | `int` (Spec number) | Which Spec's `tasks.md` this heading was found in. |
| `Task` | `int` (Task number) | The local `TASK-NNN` number. |
| `Path` | `string` | `"<tasks.md path>#TASK-NNN"`, exactly as today's `ScanResult.Paths` values (format unchanged; only the map's key structure changes). |

**Derived views** (research.md Decision 2):
1. **Same-Spec duplicates** — group entries by `(Spec, Task)`; more than one `Path` per group is a genuine duplicate (spec FR-004), fed into `validation.Finding{Code: CodeDuplicateID}` scoped to that one Spec.
2. **Cross-Spec number index** — group entries by `Task` alone, across all Specs; used only when a bare `TASK-NNN` is resolved with no Spec context (spec FR-003):
   - exactly one `Spec` claims that `Task` number → resolves unambiguously (today's convenient case, preserved).
   - more than one `Spec` claims it → resolution fails with an actionable "Spec context required" error, naming the candidate Specs — never an arbitrary pick, never treated as a duplicate Finding.

## MigrationDiagnosticEntry (FR-008, read-only report)

| Field | Type | Description |
|---|---|---|
| `TaskNumber` | `int` | The local Task number that collided under the old global-scan model. |
| `Specs` | `[]ids.EntityID` | Every Spec that independently owns a Task with this number. |
| `Paths` | `[]string` | The corresponding `tasks.md#TASK-NNN` locations, one per `Specs` entry, in the same order. |

Populated from the same Cross-Spec number index above (any group with more than one `Spec`), formatted as a diagnostic rather than a validation error — this is expected, valid state under the new model, not something requiring a fix (spec User Story 3, Assumptions).

## State / Lifecycle

No new lifecycle states are introduced. Task's existing `pending`/`complete` states (from its Markdown checkbox) are untouched; this feature only changes how a Task is *named and located*, not what states it can be in.
