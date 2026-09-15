# Phase 1 Contracts: Kit and Installer Package APIs

Like the prior features, no HTTP/CLI surface yet. This feature's contract
is: a new top-level package (`kit`), a new package
(`internal/installer`), and a modification to `internal/templates`'s
internals only (its exported API is unchanged — not repeated here).

**Reconciled against the actual implementation (T016)**: every public
signature below matches the shipped implementation exactly — no drift.
One discovery worth recording: a black-box "`Install` rejects a
traversal attempt" test turned out to be unreachable, the same situation
003-entity-creation documented for `Create`'s "already exists" check —
every `Resource.Name` `List()` can ever produce comes from
`kit.TemplatesFS`'s own directory entries, which `go:embed` guarantees
are safe basenames. `Install`'s containment guard (FR-006) is still real,
tested infrastructure — its coverage is a white-box unit test of the
unexported `installOne` helper, called directly with a contrived
malicious resource name (`internal/installer/containment_test.go`),
proving the guard itself works independent of whether today's resource
set can ever trigger it.

## `kit` — prerequisite for User Story 1, 2 & 3

```go
package kit

import "embed"

// TemplatesFS embeds every artifact template
// (docs/architecture-specification.md §22-31's schemas). The single
// source of truth for template content — both internal/templates
// (rendering) and internal/installer (raw materialization) read from
// this same embed.FS (FR-008).
//
//go:embed templates/*.tmpl
var TemplatesFS embed.FS
```

## `internal/installer` — User Story 1: Discovery

```go
package installer

// Resource is one embedded item available for installation (FR-002).
type Resource struct {
    Name         string // the embedded file's own name
    Kind         string // "template" today; a later feature may add more
    ArtifactType string // for a template: the entity/artifact type it belongs to
}

// List enumerates every resource in the embedded kit, derived directly
// from kit.TemplatesFS's own contents — never a separately maintained
// list that could drift from what's actually embedded (FR-002, FR-009).
func List() []Resource
```

**Guarantees**: `List()` performs no filesystem access outside the
embedded FS itself, and returns the identical result on every call
(spec.md US1 acceptance scenario 2).

## `internal/installer` — User Story 2: Installation

```go
package installer

type OutcomeStatus int

const (
    Installed OutcomeStatus = iota
    Skipped
    Failed
)

// Outcome is the per-resource result of one Install call (FR-005).
type Outcome struct {
    Resource Resource
    Status   OutcomeStatus
    Path     string // relative to targetDir
    Err      error  // set only when Status == Failed
}

// Install materializes every resource List() would report into
// targetDir, one atomic write per resource (FR-003, FR-007). A resource
// whose target already exists is Skipped unless overwrite is true
// (FR-004). A resource whose destination would resolve outside
// targetDir is Failed before anything is written (FR-006), reusing
// artifacts.RelativeWithinRoot — not a new containment check.
func Install(targetDir string, overwrite bool) ([]Outcome, error)
```

**Guarantees**:
- `Install` returns exactly one `Outcome` per resource `List()` reports
  — never a single pass/fail flag for the whole call (FR-005).
- Re-running `Install` with `overwrite: false` against a fully-populated
  `targetDir` never modifies any file — every `Outcome.Status` is
  `Skipped` (FR-004, SC-002).
- Every `Outcome.Status == Installed` file is byte-for-byte identical to
  its embedded source (SC-003).
- `Install`'s returned `error` (as opposed to a per-resource `Failed`
  `Outcome`) is reserved for something that prevented the operation from
  running at all (e.g. `targetDir` itself cannot be created) — a single
  resource's own problem is always an `Outcome`, never this error.

## Cross-cutting: `OutcomeStatus` → future JSON status mapping

Extends the running convention from 001-004 — this feature has no
sentinel `error` values of its own (containment/write failures surface as
`Outcome.Err`, wrapping `artifacts.ErrPathOutsideProject` for the
containment case, reusing that existing vocabulary rather than minting a
new one):

| `OutcomeStatus` | Future JSON `status` |
|---|---|
| `Installed` | `"installed"` |
| `Skipped` | `"skipped"` |
| `Failed` | `"failed"` (with `Err`'s message, and `artifacts.ErrPathOutsideProject`'s existing `path_outside_project` code when that's the cause) |
