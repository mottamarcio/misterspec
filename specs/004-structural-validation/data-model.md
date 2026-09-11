# Phase 1 Data Model: Structural Validation and Project Status

Entities extracted from `spec.md` § Key Entities, expanded with fields and
validation rules. This feature adds a new package (`internal/validation`)
and one new file in the existing `internal/operations` package
(`status.go`).

## `internal/validation`

### Severity

| Value | Meaning |
|---|---|
| `SeverityError` | A genuine structural problem (the only severity this feature emits; the type exists so a future feature can add warnings without a breaking change). |

### Finding

One detected structural problem (FR-005).

| Field | Type | Notes |
|---|---|---|
| `Code` | string | Stable, machine-matchable (e.g. `"missing_parent"`, `"duplicate_id"`, `"invalid_status"`, `"id_location_mismatch"`, `"unresolved_dependency"`) — chosen to match `docs/architecture-specification.md` §7's reserved JSON error codes where one already exists, so no renaming is needed when the CLI/JSON layer is built. |
| `Severity` | `Severity` | Always `SeverityError` in this feature. |
| `Path` | string | The affected artifact's path, relative to root. |
| `Message` | string | Specific, human-readable — never generic. |

**Validation rules**: A `Finding` is only ever constructed by the checks
inside `validation`; there is no way to build an unvalidated one from
outside the package.

### ValidateEntity / ValidateProject

| Function | Signature | Notes |
|---|---|---|
| `ValidateEntity` | `(root string, cfg project.Configuration, rawID string) ([]Finding, error)` | Single entity (US1). `error` is reserved for a genuinely unusable request (malformed `rawID`, or a type this feature doesn't validate standalone — Task, Plan, Tasks, Validation, Constitution); every structural anomaly about the entity itself (not found, ambiguous, missing/wrong-type parent, invalid state, unresolved dependency, ID/location mismatch) is a `Finding`, never an error (research.md). |
| `ValidateProject` | `(root string, cfg project.Configuration) ([]Finding, error)` | Whole project (US2): every applicable single-entity check across every discovered Program/Feature/Spec/Knowledge/Learning, plus duplicate-ID detection across those types and Task (research.md's Task scope decision). `error` is reserved for a filesystem-level failure (e.g. an unreadable directory), never for a structural problem — those are always `Finding`s (FR-011). |

**Validation rules** (FR-001 through FR-011): Both functions return every
applicable `Finding`, never stopping at the first (FR-004); an entity or
project with no problems returns an empty, non-nil-error slice (FR-011).

### Fixed reference tables (not configuration — see research.md)

| Table | Shape |
|---|---|
| Allowed lifecycle states | `map[ids.EntityType][]string` — Program/Feature: draft/active/done/cancelled; Spec: draft/ready/in_progress/validated/blocked/superseded/cancelled; Learning: candidate/promoted/dismissed; Knowledge: active. |
| Required parent type | `map[ids.EntityType]ids.EntityType` (with a presence check) — Feature→Program, Spec→Feature; Program/Knowledge/Learning have no entry (no parent concept). |

## `internal/operations` (extended)

### StatusSummary (User Story 3)

| Field | Type | Notes |
|---|---|---|
| `Counts` | `map[ids.EntityType]int` | Entity counts by type (Program, Feature, Spec, Knowledge, Learning — and Task, from the same free duplicate-detection scan). |
| `SpecsByState` | `map[string]int` | Count of Specs per declared lifecycle state. |
| `StructuralErrors` | int | `len(validation.ValidateProject(root, cfg))`. |

**Validation rules** (FR-012, FR-013): `Status` is computed fresh from the
current filesystem on every call — no caching, no stored state, the same
guarantee `ids.Scan`/`Resolve` already make elsewhere in this project.

## State / Flow Summary

```text
ValidateEntity(rawID)                                          [US1]
  ids.ParseAny(rawID) → type, or a Go error (malformed input)
  type must be one of Program/Feature/Spec/Knowledge/Learning,
    else a Go error (unsupported target)
  ids.Parse(type, rawID, cfg.IDWidth) → strict width check,
    else a Go error (malformed input — mirrors operations.Resolve)
  ids.Scan(root, cfg, type) → Paths[number]:
    0 paths  → Finding "not_found"
    >1 paths → Finding "duplicate_id" (or "ambiguous", same underlying data)
    1 path   → proceed:
      artifacts.ParseMetadata(path) → Metadata, or a Finding per
        distinct artifacts error (malformed frontmatter, missing field)
      meta.ID vs the ID implied by the scanned path → Finding "id_location_mismatch" if they differ
      meta.Status vs the fixed allowed-states table → Finding "invalid_status" if not allowed
      meta.Parent (if type requires one) → same 0/>1/1 lookup against
        the required parent type → Finding "missing_parent" /
        "invalid_parent_type" as appropriate
      (Spec only) meta.DependsOn/Supersedes → same lookup per entry
        against Spec → Finding "unresolved_dependency" per unresolved entry
  → []Finding (possibly empty)

ValidateProject()                                               [US2]
  for each type in {Program, Feature, Spec, Knowledge, Learning}:
    ids.Scan(root, cfg, type) → for every found number, run the same
      per-entity checks ValidateEntity runs after its own Scan step
    append type's Duplicates → Finding "duplicate_id" per duplicated number
  ids.Scan(root, cfg, Task) → append Duplicates only (no per-entity checks)
  → []Finding (aggregated, possibly empty)

Status()                                                        [US3]
  ids.Scan(root, cfg, t) for each type → Counts[t] = len(IDs)
  for each found Spec: operations.Inspect(specID) → SpecsByState[status]++
  validation.ValidateProject(root, cfg) → StructuralErrors = len(findings)
  → StatusSummary
```
