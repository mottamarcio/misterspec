package index

import "github.com/mottamarcio/misterspec/internal/project"

// Link is one relationship edge, in operations' own vocabulary (012) —
// Relation is "parent", "depends_on", "supersedes", or "wikilink".
type Link struct {
	Relation string
	Source   string // entity ID string
	Target   string // entity ID string
}

// SearchResult is one full-text match, identifying exactly which
// indexed chunk it came from (013's own Chunk provenance).
type SearchResult struct {
	Path      string
	Heading   string
	Content   string
	StartLine int
	EndLine   int
	// Rank is FTS5's own bm25() score — lower is more relevant, per
	// FTS5's own convention.
	Rank float64
}

// SyncError names one artifact that failed to be read or structured
// during a Sync/Rebuild run — recorded, never fatal to the rest of the
// run (FR-010).
type SyncError struct {
	Path string
	Err  error
}

// SyncReport summarizes one Sync or Rebuild run.
type SyncReport struct {
	Indexed int
	Updated int
	Skipped int
	Removed int
	Errors  []SyncError
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

	// Search runs a free-text query over indexed chunk content,
	// returning at most limit results ordered by relevance (FR-003).
	// query is treated entirely as literal text — never FTS5 query
	// syntax (036-text-search-ranking spec FR-001, FR-003).
	Search(query string, limit int) ([]SearchResult, error)

	// SearchAdvanced runs query against FTS5's own MATCH grammar
	// unmodified — phrase, prefix, boolean, and NEAR syntax are honored
	// (036-text-search-ranking spec FR-002). A malformed expression
	// returns an error wrapping ErrQuerySyntax (spec FR-004).
	SearchAdvanced(query string, limit int) ([]SearchResult, error)

	// Outgoing/Incoming report id's already-indexed relationships in
	// each direction (FR-008) — consistent with operations.References/
	// Backlinks for the five entity types those already cover (FR-009,
	// research.md #4).
	Outgoing(id string) ([]Link, error)
	Incoming(id string) ([]Link, error)

	// Close releases the underlying database handle.
	Close() error
}
