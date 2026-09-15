# Phase 1 Data Model: Atomic Entity Creation

Entities extracted from `spec.md` § Key Entities, expanded with fields and
validation rules. This feature adds two new packages
(`internal/lock`, `internal/templates`) and two new files in
`internal/operations` (`create.go`, `create_artifact.go`).

## `internal/lock`

### Lock

An opaque handle to a held project lock.

| Field | Type | Notes |
|---|---|---|
| *(unexported)* `path` | string | The lock file's absolute path. |

**Validation rules** (FR-005, FR-006):
- A `Lock` only exists once successfully acquired; there is no
  partially-acquired value.
- `Release()` removes the lock file; calling it more than once is safe
  (a missing file on the second call is not an error).

**Errors**:
- `ErrLockTimeout` — `Acquire` could not obtain the lock within its
  timeout (the lock is legitimately held by another in-progress
  creation, not stale).

## `internal/templates`

### Kind

Which artifact/entity template to render.

| Value | Renders |
|---|---|
| `Program` | `program.md` (§25) |
| `Feature` | `feature.md` (§26) |
| `Spec` | `spec.md` (§27) |
| `Knowledge` | `KNOW-*.md` (§23) |
| `Learning` | `LRN-*.md` (§31) |
| `Plan` | `plan.md` (§28) |
| `Tasks` | `tasks.md` (§29) |
| `Validation` | `validation.md` (§30) |

### Per-kind render data

Each `Kind` has its own typed data struct — no shared "everything" struct,
so a caller cannot accidentally supply irrelevant fields for a type
(Constitution Principle VI):

| Kind | Data fields |
|---|---|
| Program | `ID string` |
| Feature | `ID, Parent string` |
| Spec | `ID, Parent string` |
| Knowledge | `ID string` |
| Learning | `ID string` |
| Plan | `For string` |
| Tasks | `For string` |
| Validation | `For string` |

**Validation rules**: `Render(kind, data)` returns an error if `data` is
not that `Kind`'s expected type — a compile-time-adjacent guard against
mismatched template/data pairs, checked once at the single call site in
`create.go`/`create_artifact.go`.

## `internal/operations` (extended)

### CreateRequest / CreateResult (User Story 1)

| Field | Type | Notes |
|---|---|---|
| `Type` | `ids.EntityType` | One of Program, Feature, Spec, Knowledge, Learning (FR-009). |
| `Parent` | string | Raw parent ID; required for Feature (a Program) and Spec (a Feature); ignored for Program/Knowledge/Learning. |
| `Slug` | string | Required for Knowledge/Learning; validated non-empty and filesystem-safe; ignored otherwise. |

| `CreateResult` field | Type | Notes |
|---|---|---|
| `ID` | `ids.EntityID` | The newly allocated ID. |
| `Path` | string | The new artifact's canonical path, relative to root. |

**Validation rules** (FR-001 through FR-009):
- `Create` either returns a fully populated `CreateResult` with a
  corresponding, immediately-inspectable artifact on disk, or returns an
  error and writes nothing — never a partial result.
- ID allocation (via `ids.Scan` + `ids.NextID`), directory creation, and
  the artifact write all happen inside one held `Lock`'s critical section.

### CreateArtifactRequest / CreateArtifactResult (User Story 2)

| Field | Type | Notes |
|---|---|---|
| `Kind` | `artifacts.ArtifactType` | Must be `TypePlan`, `TypeTasks`, or `TypeValidation` (FR-010). |
| `For` | string | Raw Spec ID this artifact belongs to. |

| `CreateArtifactResult` field | Type | Notes |
|---|---|---|
| `Path` | string | The new artifact's canonical path, relative to root. |

**Validation rules** (FR-010 through FR-012): Same all-or-nothing
guarantee as `Create`; no `ids.EntityID` is allocated since these types
have none.

## Error vocabulary (new sentinels in `internal/operations`, plus `lock.ErrLockTimeout`)

| Sentinel | Condition | FR |
|---|---|---|
| `ErrInvalidParent` | Declared parent doesn't exist, is ambiguous, or is the wrong type. | FR-004, FR-011 |
| `ErrAlreadyExists` | Target canonical path already has an artifact. | FR-008, FR-012 |
| `ErrUnsupportedType` | An entity/artifact type this feature does not create was requested. | FR-014 |
| `lock.ErrLockTimeout` | The lock could not be acquired within its timeout. | FR-014 |

`Create`/`CreateArtifact` reuse `artifacts.ErrPathOutsideProject` directly
for FR-008's traversal case, the same vocabulary-reuse discipline
`002-read-operations` established.

## State / Flow Summary

```text
Create(type, parent?, slug?)                                    [US1]
  lock.Acquire(root, staleAfter, timeout) → *Lock | ErrLockTimeout
  defer lock.Release()
  if type needs a parent:
    operations.Resolve(parent) → validates existence + correct type
                                → ErrInvalidParent on failure
  ids.Scan(root, cfg, type) → ScanResult
  ids.NextID(result.IDs, type, cfg.IDWidth) → new EntityID
  artifacts.ResolvePath(...) → canonical directory + file
  check target does not already exist → ErrAlreadyExists
  templates.Render(kind, data) → initial content
  write temp file in target directory → fsync → os.Rename → final path
  → CreateResult{ID, Path}

CreateArtifact(kind, forSpecID)                                  [US2]
  lock.Acquire(...) / defer Release()
  operations.Resolve(forSpecID) → validates Spec exists
                                → ErrInvalidParent on failure
  compute fixed path (same directory as the Spec, kind's filename)
  check target does not already exist → ErrAlreadyExists
  templates.Render(kind, data{For: forSpecID}) → initial content
  write temp file → fsync → os.Rename → final path
  → CreateArtifactResult{Path}

Concurrency (US3)
  Two Create/CreateArtifact calls for the same type/target serialize on
  the same lock.Acquire — the second call's Scan/NextID (or existence
  check) only runs after the first call's Release, so it always sees the
  first call's result and allocates/rejects correctly.
  A lock file older than staleAfter, found on Acquire, is removed and
  acquisition retried — recovering automatically from an interrupted
  prior attempt with no manual step.
```
