# Phase 1 Contracts: Validation and Status Package APIs

Like the prior features, no HTTP/CLI surface yet. This feature's contract
is the exported Go API of one new package (`internal/validation`) and one
new file in the existing `internal/operations` package (`status.go`).

**Layering note (see research.md)**: `internal/validation` depends only on
`internal/ids`, `internal/artifacts`, `internal/project` — never on
`internal/operations`, to avoid an import cycle with `operations/status.go`
and because validation's "collect every anomaly as data" philosophy is a
poor fit for `Resolve`/`Inspect`'s fail-fast error semantics anyway.

**Reconciled against the actual implementation (T014)**: two `Finding`
codes were added during implementation that the original draft below
didn't anticipate — `frontmatter_malformed` and `required_field_missing`,
covering the case where `checkEntity` calls `artifacts.ParseMetadata` and
it fails outright (distinct from the entity parsing successfully but
having an invalid status/parent/dependency). Both map to `invalid_metadata`
in the Finding-code table below, consistent with how `artifacts`'s own
identically-named errors were already mapped in 001-core-foundation's
contracts.

## `internal/validation` — User Story 1 & 2

```go
package validation

type Severity int

const SeverityError Severity = iota

type Finding struct {
    Code     string
    Severity Severity
    Path     string // relative to root
    Message  string
}

// ValidateEntity checks rawID's entity against every applicable rule
// (FR-001, FR-002), returning every Finding — never stopping at the
// first. error is reserved for a genuinely unusable request: malformed
// rawID, or a type this feature does not validate standalone (Task,
// Plan, Tasks, Validation, Constitution). Every structural anomaly about
// the entity itself (not found, ambiguous, missing/wrong-type parent,
// invalid lifecycle state, unresolved dependency, an ID that doesn't
// match its own canonical location) is a Finding, never an error.
func ValidateEntity(root string, cfg project.Configuration, rawID string) ([]Finding, error)

// ValidateProject runs ValidateEntity's checks across every discovered
// Program/Feature/Spec/Knowledge/Learning, plus duplicate-ID detection
// for those types and Task, aggregating every Finding from every entity
// (FR-003, FR-004, FR-010). error is reserved for a filesystem-level
// failure, never for a structural problem.
func ValidateProject(root string, cfg project.Configuration) ([]Finding, error)
```

**Guarantees**:
- Both functions return every applicable `Finding` in one call — never a
  partial result requiring repeated calls to see everything wrong
  (FR-004, SC-001).
- An entity or project with zero problems returns an empty (possibly nil)
  slice and a nil error — never an error used to signal "found problems"
  (FR-011, SC-002). This is the concrete form of Constitution Principle
  IX's `ok` vs `valid` distinction: a completed validation run with
  Findings is still `ok`.
- Every duplicate ID present in the project is caught during
  `ValidateProject`, reusing `ids.Scan`'s own `Duplicates` — zero false
  negatives by construction (FR-010, SC-003).

## `internal/operations` — User Story 3: `Status`

```go
package operations

type StatusSummary struct {
    Counts           map[ids.EntityType]int
    SpecsByState     map[string]int
    StructuralErrors int
}

// Status computes a fresh snapshot of the project's current state —
// entity counts by type, Spec counts by lifecycle state, and the total
// structural-finding count (FR-012, FR-013). Never cached, never a
// stored value — recomputed from the live filesystem on every call.
func Status(root string, cfg project.Configuration) (StatusSummary, error)
```

**Guarantees**:
- `Status`'s `Counts` and `SpecsByState` exactly match a direct count of
  the same project's artifacts, every time (FR-013, SC-004).
- `Status`'s `StructuralErrors` exactly equals
  `len(validation.ValidateProject(root, cfg))` for the same project state
  — computed by calling it, not by an independent, potentially-drifting
  implementation (SC-003 extended to `Status`).
- Every function in this feature is read-only: none creates, modifies, or
  deletes any file (FR-014, SC-005).

## Cross-cutting: Finding code → future JSON error code mapping

Extends the running table from 001/002/003. Unlike prior features'
sentinel errors (one condition, one Go error), a single validation run
can report many `Finding`s of different `Code`s in one call — the mapping
below is for each `Code` value, not a Go `error`:

| `Finding.Code` | Future JSON `error.code` (§7) |
|---|---|
| `not_found` | `entity_not_found` |
| `duplicate_id` | `duplicate_id` |
| `missing_parent` | `invalid_parent` |
| `invalid_parent_type` | `invalid_parent` |
| `invalid_status` | `invalid_metadata` |
| `id_location_mismatch` | `invalid_metadata` |
| `frontmatter_malformed` *(added during implementation)* | `invalid_metadata` |
| `required_field_missing` *(added during implementation)* | `invalid_metadata` |
| `unresolved_dependency` | `invalid_reference` |
