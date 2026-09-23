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

// StoredPackItem is one persisted item inside a StoredPack's own
// item list (043-incremental-context-reuse data-model.md "StoredPack").
type StoredPackItem struct {
	Path        string
	StartLine   int
	EndLine     int
	Fingerprint string
	Content     string
}

// StoredPack is one packs table row (043-incremental-context-reuse
// data-model.md "StoredPack") — the persisted, disposable record a
// future --base lookup reads.
type StoredPack struct {
	PackID     string
	ConfigHash string
	Target     string
	CreatedAt  int64
	Items      []StoredPackItem
}

// CodeFile / CodeDeclaration mirror the code_files/code_declarations
// tables (044-architecture-code-context-rules data-model.md
// "CodeFile"/"CodeDeclaration").
type CodeFile struct {
	Path        string
	Fingerprint string
}

// CodeDeclaration is one top-level declaration inside an indexed
// CodeFile (data-model.md "CodeDeclaration"). Body is "" when not
// indexed at full-body tier.
type CodeDeclaration struct {
	Path      string // owning CodeFile's own Path
	Name      string
	Kind      string
	Signature string
	Body      string
	StartLine int
	EndLine   int
	IsTest    bool
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

	// SavePack upserts pack into the packs table, then evicts the
	// oldest rows beyond a fixed cap (043-incremental-context-reuse
	// research.md #3) — never fails the caller's own response if
	// eviction itself has nothing to do (an empty/under-cap table).
	SavePack(pack StoredPack) error

	// LookupPack returns the stored pack for packID, and found ==
	// false (never an error) when no such row exists — the same "bool
	// separate from error" convention CommitsSinceFileAdded/HeadCommit
	// already use elsewhere in this codebase, applied here for "no
	// row" being a perfectly normal outcome (research.md #4), not a
	// failure. The caller is responsible for comparing the returned
	// pack's own ConfigHash against the current call's freshly
	// computed ConfigHash before treating it as diffable — LookupPack
	// itself does no validation, only retrieval.
	LookupPack(packID string) (pack StoredPack, found bool, err error)

	// SyncCode incrementally reconciles code_files/code_declarations
	// with root's current *.go files, excluding any path matching one
	// of exclusions' own glob patterns — the same new/changed/deleted
	// reconciliation shape Sync already applies to Markdown
	// (044-architecture-code-context-rules research.md #6).
	SyncCode(root string, exclusions []string) (SyncReport, error)

	// DeclarationsForFiles returns every CodeDeclaration whose own
	// Path is in paths, plus every declaration from that path's own
	// associated _test.go file(s) in the same directory (spec FR-007).
	// A path with no indexed match contributes no rows, not an error.
	DeclarationsForFiles(paths []string) ([]CodeDeclaration, error)

	// Close releases the underlying database handle.
	Close() error
}
