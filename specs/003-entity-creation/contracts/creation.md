# Phase 1 Contracts: Creation Package APIs

Like the prior two features, no HTTP/CLI surface yet. This feature's
contract is the exported Go API of two new packages
(`internal/lock`, `internal/templates`) and two new files in the existing
`internal/operations` package.

**Reconciled against the actual implementation (T015)**: two things
discovered during implementation, not present in the original draft
below:

- A fourth `operations` error, `ErrInvalidSlug`, was added — the draft's
  `ErrInvalidParent`/`ErrAlreadyExists`/`ErrUnsupportedType` had no sentinel
  for "the supplied Knowledge/Learning slug is empty or not
  filesystem-safe," which is a distinct condition from an invalid parent.
- `Create`'s "target already exists" check (`ErrAlreadyExists`) is, for
  the five auto-allocated types (Program/Feature/Spec/Knowledge/Learning),
  mathematically unreachable through normal black-box testing: `NextID` is
  always computed from what `ids.Scan` already found, so a freshly
  allocated ID can never collide with an existing, `Scan`-visible artifact
  by construction. The check remains in the code as a defense-in-depth
  safety net (guarding against a genuine lock-correctness bug or
  filesystem corruption), but its real, naturally reachable test coverage
  is `CreateArtifact`'s fixed-path types (Plan/Tasks/Validation), where a
  second request for the same Spec genuinely does collide — see
  `internal/operations/create_test.go`'s comment on this.

## `internal/lock` — prerequisite for User Story 1, 2 & 3

```go
package lock

var ErrLockTimeout error

type Lock struct { /* opaque */ }

// Acquire obtains root's project lock (.misterspec/.lock), recovering
// automatically from a lock file older than staleAfter, retrying up to
// timeout before returning ErrLockTimeout (FR-005, FR-006).
func Acquire(root string, staleAfter, timeout time.Duration) (*Lock, error)

// Release removes the lock file. Safe to call more than once.
func (l *Lock) Release() error

// Sensible defaults for callers that don't need to tune them.
const (
    DefaultStaleAfter = 10 * time.Second
    DefaultTimeout    = 5 * time.Second
)
```

**Guarantees**:
- Two overlapping `Acquire` calls for the same `root` never both return a
  non-nil `*Lock` at the same time — exactly one succeeds until the first
  `Release`.
- A lock file older than `staleAfter` is recovered automatically on the
  next `Acquire` — no manual cleanup required (FR-006, SC-003).

## `internal/templates` — prerequisite for User Story 1 & 2

```go
package templates

type Kind int

const (
    Program Kind = iota
    Feature
    Spec
    Knowledge
    Learning
    Plan
    Tasks
    Validation
)

type ProgramData struct{ ID string }
type FeatureData struct{ ID, Parent string }
type SpecData struct{ ID, Parent string }
type KnowledgeData struct{ ID string }
type LearningData struct{ ID string }
type PlanData struct{ For string }
type TasksData struct{ For string }
type ValidationData struct{ For string }

// Render renders kind's embedded template with data (one of the *Data
// structs above, matching kind), returning the complete initial file
// content — frontmatter and body — per
// docs/architecture-specification.md §22-31's fixed schemas.
func Render(kind Kind, data any) (string, error)
```

**Guarantees**: `Render` returns an error — never partially rendered
content — if `data`'s type does not match `kind`.

## `internal/operations` — User Story 1: `Create`

```go
package operations

var (
    ErrInvalidParent   error
    ErrAlreadyExists   error
    ErrUnsupportedType error
    ErrInvalidSlug     error // added during implementation — see note above
)

type CreateRequest struct {
    Type   ids.EntityType // Program, Feature, Spec, Knowledge, Learning
    Parent string         // raw ID; required for Feature/Spec
    Slug   string         // required for Knowledge/Learning
}

type CreateResult struct {
    ID   ids.EntityID
    Path string // relative to root
}

// Create atomically allocates the next ID for req.Type, creates its
// canonical directory, and writes its initial artifact from that type's
// template — or does none of that, returning an error (FR-001..FR-009).
func Create(root string, cfg project.Configuration, req CreateRequest) (CreateResult, error)
```

**Guarantees**:
- `Create` never allocates an ID without also writing its artifact, and
  never writes a partial artifact (FR-001, FR-007, SC-002).
- Two `Create` calls for the same `req.Type`, issued concurrently, always
  receive different, sequential IDs (FR-005, SC-001).
- A newly created artifact parses via `operations.Inspect` without error
  immediately after `Create` returns (FR-013, SC-004).

## `internal/operations` — User Story 2: `CreateArtifact`

```go
package operations

type CreateArtifactRequest struct {
    Kind artifacts.ArtifactType // must be TypePlan, TypeTasks, or TypeValidation
    For  string                 // raw Spec ID
}

type CreateArtifactResult struct {
    Path string // relative to root
}

// CreateArtifact atomically writes req.Kind's artifact at its fixed
// location under the Spec named by req.For — or does nothing, returning
// an error (FR-010..FR-012).
func CreateArtifact(root string, cfg project.Configuration, req CreateArtifactRequest) (CreateArtifactResult, error)
```

**Guarantees**: Same atomicity and immediate-inspectability guarantees as
`Create`; rejects (`ErrInvalidParent`) a `For` Spec ID that doesn't exist,
and rejects (`ErrAlreadyExists`) a target that is already present, per
User Story 2's acceptance scenarios.

## Cross-cutting: error → future JSON error code mapping

Extends the running table from 001/002:

| Go sentinel error | Future JSON `error.code` |
|---|---|
| `operations.ErrInvalidParent` | `invalid_parent` |
| `operations.ErrAlreadyExists` | `already_exists` |
| `operations.ErrUnsupportedType` | `unsupported_type` |
| `operations.ErrInvalidSlug` | `invalid_argument` |
| `lock.ErrLockTimeout` | `validation_failed` (closest existing §7 code for a transient operational failure; revisit if a later feature needs a more specific one) |
