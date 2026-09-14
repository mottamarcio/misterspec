// Package index implements misterspec's disposable local search index
// (docs/context-engine-implementation.md §10): a SQLite database
// (modernc.org/sqlite — pure Go, no CGO, preserving this project's own
// single-static-binary property) holding documents, chunks, an FTS5
// full-text index, and a copy of the reference graph 012 already
// computes, all fully derived from the project's own files.
//
// Per the project constitution (Principle III, "Filesystem Is the
// Single Source of Truth"), this package's own database is never
// authoritative over any project fact — it can be deleted at any time
// and rebuilt from scratch (Rebuild) with zero loss of real project
// information. Every read into the project itself reuses
// internal/artifacts', internal/operations', and internal/ids' already-
// published, already read-only functions (ClassifyPath, ParseMetadata,
// ReadBody, ParseDocument, Chunks, Fingerprint, References) — this
// package adds no new way of reading a project artifact.
//
// See specs/014-sqlite-index/contracts/index.md for this package's
// exported contract and specs/014-sqlite-index/data-model.md for its
// schema and entity definitions.
package index
