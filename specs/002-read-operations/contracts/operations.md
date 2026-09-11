# Phase 1 Contracts: `internal/operations` Package API

Like `001-core-foundation`, this feature exposes no HTTP/CLI surface yet
(see spec Assumptions). Its contract is the exported Go API of one new
package, `internal/operations`, plus the small additive extensions to
`internal/ids` and `internal/artifacts` recorded in `research.md` and
`data-model.md`.

**Reconciled against the actual implementation (T022)**: one naming
collision in the original draft below was corrected — `Fingerprint` was
used for both the result type and the function, which Go does not allow
two package-level declarations to share. The function keeps the name
`Fingerprint` (matching the operation name convention every other function
in this package follows — `Resolve`, `Inspect`, `Parent`, `Children`,
`Inventory`); the type is `FileFingerprint` instead. Every other signature
below matches the shipped implementation exactly.

## `internal/ids` (extended) — prerequisite for User Story 1 & 2

```go
package ids

// ParseAny parses raw (e.g. "SPEC-014"), deriving both EntityType and
// width from the string itself — the type does not need to be known in
// advance. Wraps ErrInvalidIDSyntax on failure, same as Parse.
func ParseAny(raw string) (EntityID, error)

// ScanResult gains a third field; IDs and Duplicates are unchanged.
type ScanResult struct {
    IDs        []EntityID
    Duplicates []DuplicateID
    Paths      map[int][]string // every number found → every claiming path
}
```

**Guarantees**: `Scan`'s existing guarantees (FR-011, FR-012, FR-014 from
001-core-foundation) are unchanged; `Paths` is populated for every number
`IDs` contains, with exactly the same paths `Duplicates` already surfaces
for a duplicated number.

## `internal/artifacts` (extended) — prerequisite for User Story 1, 2 & 3

```go
package artifacts

// Metadata gains a fourth field for Plan/Tasks/Validation's `for:` key.
type Metadata struct {
    // ... existing fields unchanged ...
    For *ids.EntityID
}

// RelativeWithinRoot is relativeWithinRoot, exported for reuse by
// internal/operations (Inventory, Fingerprint) — same containment
// guarantee ResolvePath/ClassifyPath already rely on internally.
func RelativeWithinRoot(root, path string) (string, error)
```

## `internal/operations` — User Story 1: Resolve & Inspect

```go
package operations

var (
    ErrEntityNotFound  error
    ErrEntityAmbiguous error // wrapped by *AmbiguousIDError
    ErrInvalidTarget   error // wraps ids.ErrInvalidIDSyntax
)

type AmbiguousIDError struct {
    ID        string
    Locations []string
}

type ResolvedLocation struct {
    ID   ids.EntityID
    Type artifacts.ArtifactType
    Path string // relative to root; "<tasks.md>#TASK-NNN" for a Task
}

// Resolve finds the one artifact matching rawID (FR-001, FR-002, FR-003).
func Resolve(root string, cfg project.Configuration, rawID string) (ResolvedLocation, error)

type InspectResult struct {
    Location ResolvedLocation
    Metadata artifacts.Metadata
}

// Inspect resolves rawID, then retrieves its metadata (FR-004, FR-005,
// FR-016). For a Task, Metadata is narrowed to ID/Status/Parent — see
// research.md's Task metadata scope decision.
func Inspect(root string, cfg project.Configuration, rawID string) (InspectResult, error)
```

**Guarantees**:
- `Resolve` never returns a `ResolvedLocation` for zero or multiple
  matches — exactly one of `ResolvedLocation`, `ErrEntityNotFound`, or a
  `*AmbiguousIDError` (`errors.Is(err, ErrEntityAmbiguous)`) is returned.
- `Inspect` returns exactly one of a populated `InspectResult`, or one of
  `ErrEntityNotFound`/`ErrEntityAmbiguous`/`ErrInvalidTarget`/
  `artifacts.ErrFrontmatterMalformed`/`artifacts.ErrRequiredFieldMissing`
  — never a partially populated result alongside a non-nil error.

## `internal/operations` — User Story 2: Parent & Children

```go
package operations

type ParentResult struct {
    HasParent bool
    Parent    *ResolvedLocation // nil when HasParent is false
}

// Parent determines rawID's structural parent (FR-006, FR-007).
func Parent(root string, cfg project.Configuration, rawID string) (ParentResult, error)

// Children enumerates rawID's direct structural children, optionally
// filtered to one child EntityType (FR-008, FR-009). A nil filterType
// enumerates every nestable child type for rawID's type.
func Children(root string, cfg project.Configuration, rawID string, filterType *ids.EntityType) ([]ResolvedLocation, error)
```

**Guarantees**:
- `Parent` returns `ParentResult{HasParent: false}` with a `nil` error for
  an entity type with no parent concept — never an error for that case.
- `Children` returns an empty (possibly nil) slice, not an error, when the
  entity has no children of the requested type.

## `internal/operations` — User Story 3: Inventory & Fingerprint

```go
package operations

type FileEntry struct {
    Path      string // relative to root
    Extension string
    Size      int64
}

// Inventory lists the files directly reachable under dir (relative to
// root, or root itself) — e.g. cfg.RawDir, cfg.KnowledgeDir, or a
// ResolvedLocation's own directory (FR-010, FR-011).
func Inventory(root, dir string) ([]FileEntry, error)

type FileFingerprint struct {
    Algorithm string // "sha256"
    Digest    string // lowercase hex
}

func (f FileFingerprint) String() string // "sha256:<hex>"

// Fingerprint computes path's content digest, streamed rather than fully
// buffered, without interpreting path's contents (FR-012).
func Fingerprint(root, path string) (FileFingerprint, error)
```

**Guarantees**:
- `Inventory` returns an empty slice, not an error, for an empty or
  not-yet-created (but validly located) directory (FR-011).
- Both `Inventory` and `Fingerprint` reject (via
  `artifacts.ErrPathOutsideProject`) a `dir`/`path` that would resolve
  outside `root`, and report `artifacts.ErrArtifactNotFound` for one that
  does not exist (FR-013) — reusing `artifacts`'s vocabulary rather than
  minting parallel sentinels for the same conditions.
- Every function in this package is read-only: none opens any file for
  writing, creates, renames, or deletes anything (FR-014).

## Cross-cutting: error → future JSON error code mapping

Extends 001-core-foundation's table with this feature's new sentinels,
for the same reason — so nothing needs renaming when the CLI/JSON layer
(§6-8) is built:

| Go sentinel error | Future JSON `error.code` |
|---|---|
| `operations.ErrEntityNotFound` | `entity_not_found` |
| `operations.ErrEntityAmbiguous` | `entity_ambiguous` |
| `operations.ErrInvalidTarget` | `invalid_target` |
