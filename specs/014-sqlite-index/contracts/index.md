# Phase 4 Contracts: Disposable SQLite Index

**Reconciled against the actual implementation (T023)** — zero drift on
the public surface. Every signature below (`Link`, `SearchResult`,
`SyncError`, `SyncReport`, `Store`, `Open`) matches what shipped,
verbatim. One cleanup made during reconciliation: an unexported
transitional placeholder, `ErrNotImplemented` (used only while `Search`/
`Outgoing`/`Incoming` were still stubs mid-implementation, per tasks.md's
own incremental build order), was removed once every `Store` method had
a real implementation and nothing referenced it anymore — dead code, not
part of this feature's contract either before or after. Every other
unexported helper (`indexOneArtifact`, `indexAllFound`,
`walkEligibleArtifacts`, `deleteDocument`, `loadExistingDocuments`,
`insertChunk`, `insertLink`, `indexLinks`, `titleFor`, `ensureSchema`,
`schemaStatements`/`dropStatements`/`clearStatements`) is internal
decomposition detail, not part of this feature's exported surface.

No HTTP/CLI surface is added by this feature (research.md) — its
contract is the new Go API in `internal/context/index`.

## `internal/context/index` (new package)

```go
package index

// Link is one relationship edge, in operations' own vocabulary (012) —
// Relation is "parent" | "depends_on" | "supersedes" | "wikilink".
type Link struct {
    Relation string
    Source   string // entity ID string
    Target   string // entity ID string
}

// SearchResult is one full-text match, identifying exactly which
// indexed chunk it came from.
type SearchResult struct {
    Path      string
    Heading   string
    Content   string
    StartLine int
    EndLine   int
    Rank      float64 // FTS5 bm25() — lower is more relevant
}

// SyncError names one artifact that failed to be read or structured
// during a Sync/Rebuild run — recorded, never fatal to the rest (FR-010).
type SyncError struct {
    Path string
    Err  error
}

// SyncReport summarizes one Sync or Rebuild run.
type SyncReport struct {
    Indexed, Updated, Skipped, Removed int
    Errors                              []SyncError
}

// Store is this package's own abstraction over its storage — no SQL
// type crosses this interface (docs/context-engine-implementation.md
// §10.4).
type Store interface {
    // Sync incrementally reconciles the index with root's current
    // state: new/changed/deleted artifacts are (re)processed; unchanged
    // ones are left untouched (FR-004, FR-005). One transaction per
    // call (FR-007).
    Sync(root string, cfg project.Configuration) (SyncReport, error)

    // Rebuild discards every derived row first, then performs the same
    // walk Sync does, treating every artifact as new (FR-006).
    Rebuild(root string, cfg project.Configuration) (SyncReport, error)

    // Search runs an FTS5 full-text query over indexed chunk content,
    // returning at most limit results ordered by relevance (FR-003).
    Search(query string, limit int) ([]SearchResult, error)

    // Outgoing/Incoming report id's already-indexed relationships in
    // each direction (FR-008) — consistent with operations.References/
    // Backlinks for the five entity types those already cover (FR-009,
    // research.md #4).
    Outgoing(id string) ([]Link, error)
    Incoming(id string) ([]Link, error)

    // Close releases the underlying database handle.
    Close() error
}

// Open opens (creating if necessary) the SQLite database at path,
// creating or recreating its schema as needed (research.md #7).
func Open(path string) (Store, error)
```

**Guarantees**:
- Every operation is read-only with respect to the project's own
  authoritative files (FR-011) — `Store` only ever reads artifacts via
  `internal/artifacts`/`internal/operations`'s own existing, already
  read-only functions; nothing in this package ever writes to a project
  artifact.
- The database itself is fully disposable (FR-002, FR-012): deleting
  it and calling `Open` again produces a store with an empty schema,
  ready for a fresh `Sync`/`Rebuild` — no project information is ever
  lost, since none was ever held there exclusively.
- `Outgoing`/`Incoming` results for the five entity types
  `operations.References`/`Backlinks` already cover are sourced
  directly from those functions during `Sync` (research.md #4) —
  consistency (FR-009) holds by construction, not by parallel
  reimplementation.
- A `Sync`/`Rebuild` call either fully commits or leaves the database
  completely unchanged from before the call (FR-007) — no partially-
  applied state is ever observable.
- One bad artifact (unreadable, malformed) never stops the rest of a
  `Sync`/`Rebuild` run (FR-010) — it is recorded in `SyncReport.Errors`
  and the walk continues.

## Cross-cutting: no change to any existing contract

`internal/artifacts`, `internal/validation`, `internal/operations`, and
`internal/cli`'s exported surfaces are completely untouched — this
feature only calls their already-existing, already-tested functions
(`ClassifyPath`, `ParseMetadata`, `ReadBody`, `ParseDocument`, `Chunks`,
`Fingerprint`, `References`). No new CLI command (research.md's own
"Summary of Go footprint") — Phase 8's own "internal context" command
is the later, actual CLI surface this capability feeds into.
