# Phase 1 Data Model: Read-Only Deterministic Operations

Entities extracted from `spec.md` § Key Entities, expanded with fields and
validation rules drawn from the Functional Requirements. This feature adds
new types in `internal/operations`, plus the small, additive extensions
to `internal/ids` and `internal/artifacts` recorded in `research.md`.

## Extensions to 001-core-foundation types

### `ids.ScanResult` (extended)

| Field | Type | Notes |
|---|---|---|
| `IDs` | `[]EntityID` | Unchanged from 001-core-foundation. |
| `Duplicates` | `[]DuplicateID` | Unchanged. |
| `Paths` *(new)* | `map[int][]string` | Every number found, mapped to every claiming path (relative to root) — length 1 for an unambiguous ID, length >1 for a duplicate (the same data `Duplicates` derives from). |

### `artifacts.Metadata` (extended)

| Field | Type | Notes |
|---|---|---|
| *(all 001-core-foundation fields)* | | Unchanged. |
| `For` *(new)* | `*ids.EntityID` | Populated from a `for:` frontmatter key (Plan/Tasks/Validation schemas); `nil` when absent. |

## New: `internal/operations` types

### ResolvedLocation

The outcome of successfully resolving exactly one ID (FR-001).

| Field | Type | Notes |
|---|---|---|
| `ID` | `ids.EntityID` | The resolved entity's ID. |
| `Type` | `artifacts.ArtifactType` | Classified from its canonical location. |
| `Path` | string | Relative to project root. For a Task, this is `"<tasks.md path>#TASK-NNN"` — no independent file of its own (FR-016). |

**Validation rules**: A `ResolvedLocation` is only ever produced for an ID
that matched exactly one artifact. Zero matches or more than one match
never produce a `ResolvedLocation` — see the error types below.

### AmbiguousIDError

Wraps `ErrEntityAmbiguous`; carries what a bare sentinel would lose
(FR-003).

| Field | Type | Notes |
|---|---|---|
| `ID` | string | The raw ID string that was ambiguous. |
| `Locations` | `[]string` | Every claiming path, relative to root. |

### InspectResult

An entity's structured metadata, plus where it lives (FR-004, FR-005).

| Field | Type | Notes |
|---|---|---|
| `Location` | `ResolvedLocation` | Where this entity is. |
| `Metadata` | `artifacts.Metadata` | Its parsed frontmatter — for Task, a narrowed subset per research.md's Task metadata scope decision (`ID`, `Status`, `Parent` only). |

### ParentResult

The outcome of a parent query — deliberately not an error for the
legitimate "no parent" case (FR-006, FR-007).

| Field | Type | Notes |
|---|---|---|
| `HasParent` | bool | `false` for a Program (or any type with no parent concept). |
| `Parent` | `*ResolvedLocation` | `nil` when `HasParent` is `false`; otherwise the resolved parent. |

### FileEntry

One file found while inventorying a directory (FR-010).

| Field | Type | Notes |
|---|---|---|
| `Path` | string | Relative to project root. |
| `Extension` | string | Including the leading dot (e.g. `.md`), or `""` for an extensionless file. |
| `Size` | int64 | Bytes. |

### FileFingerprint

A file's algorithm-tagged content digest (FR-012). Named `FileFingerprint`
rather than `Fingerprint` — the `Fingerprint` function and a same-named
type would collide as two package-level declarations, a naming conflict
discovered and fixed during implementation (see `contracts/operations.md`).

| Field | Type | Notes |
|---|---|---|
| `Algorithm` | string | Always `"sha256"` for this feature. |
| `Digest` | string | Lowercase hex-encoded SHA-256, e.g. rendered as `sha256:<hex>` by `String()`. |

## Error vocabulary (new sentinels in `internal/operations`)

| Sentinel | Condition | FR |
|---|---|---|
| `ErrEntityNotFound` | No artifact anywhere matches the requested ID. | FR-002 |
| `ErrEntityAmbiguous` | More than one artifact claims the same ID (wrapped by `*AmbiguousIDError`). | FR-003 |
| `ErrInvalidTarget` | A syntactically invalid ID was given to resolve/inspect/parent/children (reuses `ids.ErrInvalidIDSyntax` under the hood, re-exposed at this layer for a consistent operations-level vocabulary). | edge case |

`Inventory`/`Fingerprint` reuse `artifacts.ErrPathOutsideProject` and
`artifacts.ErrArtifactNotFound` directly rather than redefining them
(FR-013) — one vocabulary, no duplicate sentinels for the same condition.

## State / Flow Summary

```text
Resolve(rawID)                                    [US1]
  ids.ParseAny(rawID) → EntityID | ErrInvalidTarget
  ids.Scan(root, cfg, id.Type) → ScanResult{Paths}
  Paths[id.Number]:
    0 paths  → ErrEntityNotFound
    1 path   → ResolvedLocation
    >1 paths → *AmbiguousIDError (wraps ErrEntityAmbiguous)

Inspect(rawID)                                    [US1, depends on Resolve]
  Resolve(rawID) → ResolvedLocation | error
  artifacts.ParseMetadata(location.Path)            (file-based types)
  OR: read owning tasks.md + checkbox + `for:`       (Task — narrowed scope)
  → InspectResult

Parent(rawID)                                     [US2, depends on Inspect]
  Inspect(rawID) → InspectResult | error
  no parent concept (Program) → ParentResult{HasParent: false}
  Metadata.Parent (or, for Task, Metadata.For) → Resolve(that ID) → ParentResult

Children(parentRawID, filterType?)                [US2, depends on Resolve + Scan]
  Resolve(parentRawID) → parent ResolvedLocation
  ids.Scan(root, cfg, filterType or each nestable child type) → ScanResult{Paths}
  keep only paths nested under parent.Path           → []ResolvedLocation

Inventory(dir)                                    [US3, independent]
  artifacts.RelativeWithinRoot(root, dir) → rejects traversal
  list files directly under dir                      → []FileEntry

Fingerprint(path)                                 [US3, independent]
  artifacts.RelativeWithinRoot(root, path) → rejects traversal
  os.Open + io.Copy → sha256                          → Fingerprint
```
