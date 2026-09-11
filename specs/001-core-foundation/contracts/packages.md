# Phase 1 Contracts: Package APIs

This feature exposes no HTTP/CLI surface (see spec Assumptions — the
`misterspec internal ...` CLI is a later feature). Its "contract" is the
exported Go API of three internal packages that later features (the
deterministic `operations` layer, then the `cli` layer) will code against.

**Reconciled against the actual implementation (T025)**: two signatures
changed from the original draft below, for reasons discovered during
implementation, not arbitrarily:

- `ResolvePath`'s `parent *ids.EntityID` became `parents ...ids.EntityID`
  (ordered outermost to innermost). A Spec's canonical path nests under
  *both* its Program and its Feature
  (`ai/programs/PRG-001/features/FEAT-004/specs/SPEC-014/`) — a single
  immediate parent isn't enough to compute it without ResolvePath reading
  the filesystem to discover the grandparent itself, which would have
  broken the "pure function, no filesystem access" guarantee (SC-003).
- `ClassifyPath` gained a `cfg project.Configuration` parameter — it needs
  to know the project's configured `ProgramsRoot`/`KnowledgeDir`/
  `LearningsDir`/`ConstitutionPath` to recognize canonical locations; the
  original draft omitted it.
- `Metadata`'s `Extra map[string]any` field (in the original data-model.md
  draft) was dropped: nothing in this feature's requirements reads or
  tests it, and speculative unused fields are exactly what Constitution
  Principle IV (YAGNI) rules out. Add it back in a later feature if a
  concrete consumer needs it.

## `internal/project` — User Story 1

```go
package project

// Detect walks upward from startDir looking for a misterspec project
// marker. It never returns a partially valid Project.
func Detect(startDir string) (*Project, error)

// Errors distinguishable via errors.Is:
var ErrNotInitialized      error // no project found in ancestry
var ErrInvalidConfiguration error // wraps a *ConfigError naming the field
```

**Guarantees** (→ FR-001, FR-002, FR-003, FR-004, FR-005):
- `Detect` called from `Root` or any nested descendant of `Root` returns an
  equal `Project.Root` and `Project.Config`.
- `Detect` never panics on a missing or malformed config; it returns one of
  the two sentinel-wrapped errors above.
- `Project.Config` always has every required field populated when `Detect`
  returns a non-nil `*Project`.

## `internal/artifacts` — User Story 2 & part of User Story 3

```go
package artifacts

// Canonical path resolution (FR-007, FR-008). parents is ordered
// outermost to innermost: Feature needs parents[0] = its Program's ID;
// Spec needs parents[0] = its Program's ID, parents[1] = its Feature's ID.
// For Knowledge/Learning, CanonicalPath.File is left "" — their filename
// carries an agent-chosen slug ResolvePath is not given; only Directory
// is deterministic from (root, cfg, type, id) alone.
func ResolvePath(root string, cfg project.Configuration, t ids.EntityType, id ids.EntityID, parents ...ids.EntityID) (CanonicalPath, error)

var ErrPathOutsideProject error // returned instead of any path escaping root

// Type classification (FR-006), by canonical location only — cfg is
// required to know where ProgramsRoot/KnowledgeDir/LearningsDir/
// ConstitutionPath live for this project.
func ClassifyPath(root string, cfg project.Configuration, path string) (ArtifactType, error)

// Frontmatter parsing (FR-009)
func ParseMetadata(path string) (Metadata, error)

var (
    ErrArtifactNotFound     error
    ErrFrontmatterMalformed error
    ErrRequiredFieldMissing error // wraps the specific field name
)
```

**Guarantees**:
- `ResolvePath` is a pure function of its inputs: identical arguments always
  produce an identical `CanonicalPath`, with no filesystem access required
  to compute it (SC-003).
- `ResolvePath` and `ClassifyPath` both reject (via `ErrPathOutsideProject`)
  before ever constructing an absolute path that escapes `root` (SC-004).
- `ParseMetadata` returns exactly one of a populated `Metadata`, or one of
  the three distinct errors above — never a partially populated `Metadata`
  alongside a non-nil error.

## `internal/ids` — User Story 3

```go
package ids

// Parse validates syntax for a given entity type (FR-010).
func Parse(t EntityType, raw string, width int) (EntityID, error)

// TypeForPrefix is Prefix's inverse — added during implementation because
// artifacts.ParseMetadata needs to recover an EntityType from an
// already-written ID string (e.g. "SPEC-014") to validate it, without
// knowing the type up front.
func TypeForPrefix(prefix string) (EntityType, bool)

var ErrInvalidIDSyntax error // wrong prefix, non-numeric suffix, or wrong width

// Scan walks the project tree under the type's canonical root and
// enumerates every syntactically valid existing ID (FR-011, FR-012, FR-014).
// Note: unlike every other entity type, Task IDs are not named by their
// own file or directory — they are Markdown headings
// ("## TASK-001 — ...") inside a Spec's shared tasks.md
// (docs/architecture-specification.md §29). For EntityType Task, Scan
// parses those headings across every tasks.md under cfg.ProgramsRoot
// instead of matching directory/file names.
func Scan(root string, cfg project.Configuration, t EntityType) (ScanResult, error)

type ScanResult struct {
    IDs        []EntityID
    Duplicates []DuplicateID // non-fatal; Scan still returns all other IDs
}

// NextID is a pure function — never reads or writes any persisted counter
// (FR-013, Constitution Principle III).
func NextID(existing []EntityID, t EntityType, width int) EntityID
```

**Guarantees**:
- `Scan` over an entity type with zero existing artifacts returns
  `ScanResult{IDs: nil, Duplicates: nil}, nil` — not an error (FR-014,
  SC-005).
- A duplicate ID encountered during `Scan` is appended to `Duplicates` and
  scanning continues; `Scan` only returns a non-nil `error` for a
  filesystem-level failure (e.g. unreadable directory), never merely
  because a duplicate was found (FR-012).
- `NextID` takes no arguments derived from stored state — only the
  already-scanned `existing` slice — satisfying Constitution Principle III
  ("no persistent ID counter") by construction.

## Cross-cutting: error → future JSON error code mapping

Recorded here (not implemented in this feature) so the vocabulary chosen
now doesn't need renaming when the CLI/JSON layer (§6–8 of the architecture
spec) is built in a later feature:

| Go sentinel error | Future JSON `error.code` |
|---|---|
| `project.ErrNotInitialized` | `project_not_initialized` |
| `project.ErrInvalidConfiguration` | `invalid_metadata` |
| `artifacts.ErrPathOutsideProject` | `path_outside_project` |
| `artifacts.ErrArtifactNotFound` | `entity_not_found` |
| `artifacts.ErrFrontmatterMalformed` | `invalid_metadata` |
| `artifacts.ErrRequiredFieldMissing` | `invalid_metadata` |
| `ids.ErrInvalidIDSyntax` | `invalid_id` |
| `ids.DuplicateID` (in `ScanResult`, not an error) | `duplicate_id` (surfaced by a later `validate` operation) |
