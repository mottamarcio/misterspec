# Contract: Task identity parsing, resolution and validation

This documents the machine-readable contract for the `internal/ids` / `internal/operations` / `internal/validation` surfaces this feature changes, and the `misterspec internal inspect` CLI surface built on top of them (Constitution Principle IX — structured, stable outcomes; no decorative prose is a substitute for these).

## 1. `ids.ParseTaskRef(raw string, cfg project.Configuration) (TaskID, error)`

Parses a composite task reference.

**Input**: `raw` — either `"SPEC-###:TASK-###"` (both halves present) or a bare `"TASK-###"` (Spec half absent — see §2 for how a bare form is completed).

**Output** (success): a `TaskID{Spec, Local}` per data-model.md, both halves syntactically valid against `cfg.IDWidth`.

**Errors** (all wrap the existing `ids.ErrInvalidIDSyntax` sentinel — no new sentinel needed for syntax failures):
| Condition | Behavior |
|---|---|
| Malformed composite (e.g. `"SPEC-014:"`, `":TASK-003"`, `"SPEC-014-TASK-003"`) | `ErrInvalidIDSyntax`, message names which half failed |
| Wrong ID width on either half | `ErrInvalidIDSyntax` (mirrors existing `ids.Parse` width enforcement) |
| Right shape but wrong entity type in either slot (e.g. `"TASK-003:TASK-004"`) | `ErrInvalidIDSyntax` |

## 2. `operations.ResolveTask(root string, cfg project.Configuration, raw string, specCtx *ids.EntityID) (ResolvedLocation, error)`

Resolves a task reference to exactly one location. Supersedes calling generic `Resolve` with a Task ID for any caller that can supply Spec context; `Resolve` itself keeps working unchanged for bare `TASK-###` (delegates internally, per Decision 2's cross-Spec index).

**Inputs**:
- `raw` — composite (`SPEC-###:TASK-###`) or bare (`TASK-###`).
- `specCtx` — optional. If `raw` is already composite, `specCtx` (if also given) MUST agree with `raw`'s Spec half or resolution fails as invalid input, never silently prefers one over the other.

**Output** (success): `ResolvedLocation` (existing type, unchanged shape) naming exactly one Task.

**Errors / outcomes** (new stable codes alongside existing `ErrEntityNotFound` / `ErrEntityAmbiguous` / `ErrInvalidTarget`):
| Condition | Behavior |
|---|---|
| Composite reference, Spec or Task doesn't exist | `ErrEntityNotFound` (existing sentinel, reused) |
| Bare reference, unique across all Specs | Resolves successfully (today's behavior, preserved — spec FR-009) |
| Bare reference, no `specCtx`, claimed by >1 Spec | **New**: `ErrSpecContextRequired`, wrapped in a `SpecContextRequiredError{TaskNumber int, Candidates []ids.EntityID}` naming every candidate Spec — never `ErrEntityAmbiguous` (that sentinel stays reserved for a true same-identity collision, e.g. duplicate directories) |
| Composite or spec-qualified reference, Spec exists but has no such Task | `ErrEntityNotFound` |
| Two `## TASK-NNN` headings with the same number inside the *same* Spec's `tasks.md` | `ErrEntityAmbiguous` (existing sentinel — this genuinely is the same identity claimed twice, spec FR-004) |

## 3. `validation.ValidateProject` — duplicate-Task finding, rescoped

**Behavior change**: the existing `Finding{Code: CodeDuplicateID}` emitted for Task headings is now computed **per Spec** (grouped via the Cross-Spec/Same-Spec split in data-model.md), not across the whole project.

| Scenario | Before this feature | After this feature |
|---|---|---|
| `SPEC-001` and `SPEC-002` each have a `TASK-001` | `CodeDuplicateID` Finding (false positive) | No Finding |
| `SPEC-001`'s `tasks.md` has two `## TASK-001` headings | `CodeDuplicateID` Finding | `CodeDuplicateID` Finding (unchanged — still correctly flagged) |

`Finding.Message` wording for the still-flagged case is updated to name the owning Spec explicitly (e.g. `"SPEC-001: 2 Task headings claim the same number 1: [...]"`) so the scope is legible without cross-referencing the path.

`validation.ValidateEntity(root, cfg, "SPEC-###")` (single-Spec validation) MUST produce the same per-Spec duplicate Findings for that Spec's own tasks as `ValidateProject` would for it — this is spec FR-011 and was already implicitly true once §3's scoping fix lands, since both now consume the same per-Spec view.

## 4. Migration diagnostic — `misterspec internal migration-check tasks` (or equivalent internal subcommand name, finalized during `/speckit-tasks`)

**Behavior**: Structured JSON report, one entry per `MigrationDiagnosticEntry` (data-model.md §"MigrationDiagnosticEntry") — every Task number claimed by more than one Spec project-wide. Exit code `0` always (this is informational, not a failure signal — spec FR-008 says it MUST NOT alter artifacts and is not itself a validation error).

```json
{
  "ok": true,
  "collisions": [
    {
      "task_number": 1,
      "specs": ["SPEC-001", "SPEC-002"],
      "paths": [
        "programs/.../SPEC-001/tasks.md#TASK-001",
        "programs/.../SPEC-002/tasks.md#TASK-001"
      ]
    }
  ]
}
```

An empty `collisions` array is a valid, common outcome (most projects will have none), distinguished from `ok: false` (a genuine command failure, e.g. unreadable project root) per Constitution Principle IX's `ok` vs. `valid`/`result` distinction.

## 5. `misterspec internal inspect <id>` CLI surface

**Behavior change**: the existing `Use: "inspect <id>"` / `Args: cobra.ExactArgs(1)` contract (`internal/cli/internalcmd/inspect.go`) is extended, not replaced:
- `<id>` MAY now be the composite form (`SPEC-014:TASK-003`) for a Task — resolved via §2.
- `<id>` remaining a bare form for any entity type (including `TASK-003`) keeps working exactly as today for every non-ambiguous case.
- When a bare Task reference is ambiguous (§2's `ErrSpecContextRequired`), the command's JSON error output MUST include the candidate Specs (Constitution Principle IX — Skills reason about error codes and structured fields, not prose), so a calling Skill can immediately retry with the composite form instead of parsing free text.

No new required flag/argument is introduced — composite disambiguation happens through the existing single positional argument, per research.md Decision 3.
