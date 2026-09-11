# Phase 1 Data Model: Embedded Kit and Resource Installer

Entities extracted from `spec.md` § Key Entities. This feature adds a new
top-level package (`kit`), a new package (`internal/installer`), and
modifies `internal/templates` to read from `kit` instead of its own
private embed (research.md).

## `kit`

| Symbol | Type | Notes |
|---|---|---|
| `TemplatesFS` | `embed.FS` | `//go:embed templates/*.tmpl`. The single source of truth for template content (FR-008) — both `internal/templates` (rendering, for `operations.Create`/`CreateArtifact`) and `internal/installer` (raw materialization) read from this same `embed.FS`. |

**Validation rules**: `kit/templates/` holds exactly the 8 files already
proven correct in 003-entity-creation (`program.md.tmpl`,
`feature.md.tmpl`, `spec.md.tmpl`, `knowledge.md.tmpl`,
`learning.md.tmpl`, `plan.md.tmpl`, `tasks.md.tmpl`, `validation.md.tmpl`)
— byte-for-byte identical to their pre-relocation content (SC-005).

## `internal/installer`

### Resource (User Story 1)

One embedded item available for installation (FR-002).

| Field | Type | Notes |
|---|---|---|
| `Name` | string | The embedded file's own name (e.g. `program.md.tmpl`). |
| `Kind` | string | Which category of resource this is — `"template"` for everything this feature embeds; a string (not a closed enum) so a later feature can add `"skill"`/`"integration"` kinds without a breaking change to this type. |
| `ArtifactType` | string | For a template resource: the entity/artifact type it belongs to (e.g. `"program"`), derived from its filename. |

**Validation rules**: `List()` derives every `Resource` directly from
`kit.TemplatesFS`'s own contents (FR-009) — there is no separately
maintained list that could drift from what's actually embedded.

### Outcome (User Story 2)

The per-resource result of one `Install` call (FR-005).

| Field | Type | Notes |
|---|---|---|
| `Resource` | `Resource` | Which resource this outcome is for. |
| `Status` | `OutcomeStatus` | `Installed`, `Skipped`, or `Failed`. |
| `Path` | string | The resource's destination, relative to the target directory. |
| `Err` | error | Set only when `Status == Failed`; the specific reason. |

**Validation rules** (FR-003 through FR-007):
- `Install` returns exactly one `Outcome` per `Resource` `List()` would
  report — never fewer, never a single aggregate pass/fail for the whole
  call.
- `Status == Skipped` occurs only when the target file already exists and
  overwrite was not requested — the file itself is untouched (FR-004,
  SC-002).
- `Status == Installed` guarantees the target file is now byte-for-byte
  identical to the embedded source (SC-003), written via the atomic
  temp-file-then-rename recipe (FR-003, FR-007).
- A resource whose destination would resolve outside the target
  directory never reaches `Installed` or `Skipped` — it is `Failed` with
  `Err` wrapping the containment rejection, and nothing is written for it
  (FR-006, SC-004).

## State / Flow Summary

```text
List()                                                    [US1]
  fs.WalkDir(kit.TemplatesFS, "templates") → for each file:
    Resource{Name, Kind: "template", ArtifactType: derived from filename}
  → []Resource (no filesystem access outside the embedded FS itself)

Install(targetDir, overwrite)                             [US2]
  for each Resource from List():
    dest := filepath.Join("templates", resource.Name)     # under targetDir
    artifacts.RelativeWithinRoot(targetDir, dest)          # FR-006
      → escapes root: Outcome{Failed, Err: containment rejection}; continue
    already exists && !overwrite:
      → Outcome{Skipped}; continue
    read resource bytes from kit.TemplatesFS
    write via temp-file → fsync → os.Rename (filesystem.go's own helper,
      not operations's — research.md)
      → success: Outcome{Installed}
      → failure: Outcome{Failed, Err: the write error}
  → []Outcome, one per Resource

003-entity-creation regression (US3)
  internal/templates.Render(kind, data) now reads its .tmpl content via
    kit.TemplatesFS.ReadFile("templates/"+filename) instead of its own
    private embed.FS — Render's signature, and everything
    operations.Create/CreateArtifact do with it, is unchanged.
  → 003's existing test suite passes unmodified (SC-005)
```
