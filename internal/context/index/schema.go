package index

import (
	"database/sql"
	"fmt"
)

// schemaVersion is tagged onto every database this package creates via
// PRAGMA user_version (research.md #7). This index is explicitly
// disposable (Constitution Principle III) — a version mismatch (an
// older schema, or a brand-new empty database whose default version is
// 0) always means "recreate from scratch," never "migrate in place."
// Bumped to 2 by 038-wikilink-chunk-provenance: links gains
// source_section/source_line (data-model.md "links table (widened)").
// Bumped to 3 by 040-stable-section-anchors: chunks gains anchor,
// links gains target_anchor (data-model.md "Index Schema (extended)").
// Bumped to 4 by 043-incremental-context-reuse: new packs table
// (data-model.md "StoredPack") — this table is disposable exactly
// like every other table here (Constitution Principle III, spec
// FR-011): losing every stored pack on an unrelated schema bump only
// ever costs one call's worth of reuse, never correctness.
const schemaVersion = 4

// schemaStatements creates this package's own schema (data-model.md) —
// one statement per Exec call, rather than relying on a driver's
// multi-statement support.
var schemaStatements = []string{
	`CREATE TABLE documents (
		id            INTEGER PRIMARY KEY,
		path          TEXT NOT NULL UNIQUE,
		artifact_id   TEXT,
		artifact_type TEXT NOT NULL,
		title         TEXT,
		fingerprint   TEXT NOT NULL,
		indexed_at    INTEGER NOT NULL
	)`,
	`CREATE TABLE chunks (
		id             INTEGER PRIMARY KEY,
		document_id    INTEGER NOT NULL REFERENCES documents(id),
		heading        TEXT,
		anchor         TEXT,
		content        TEXT NOT NULL,
		start_line     INTEGER NOT NULL,
		end_line       INTEGER NOT NULL,
		token_estimate INTEGER NOT NULL
	)`,
	`CREATE VIRTUAL TABLE chunks_fts USING fts5(
		heading,
		content,
		chunk_id UNINDEXED
	)`,
	`CREATE TABLE links (
		id                 INTEGER PRIMARY KEY,
		source_artifact_id TEXT NOT NULL,
		target_artifact_id TEXT NOT NULL,
		relation           TEXT NOT NULL,
		source_section     TEXT,
		source_line        INTEGER,
		target_anchor      TEXT
	)`,
	`CREATE TABLE packs (
		pack_id     TEXT PRIMARY KEY,
		config_hash TEXT NOT NULL,
		target      TEXT NOT NULL,
		created_at  INTEGER NOT NULL,
		items_json  TEXT NOT NULL
	)`,
}

// dropStatements removes every table this package's schema may have
// created, in dependency order (chunks_fts/chunks/links before
// documents), tolerant of a table that doesn't exist (a brand-new
// database, or a stale schema that never created a given table at all).
var dropStatements = []string{
	`DROP TABLE IF EXISTS chunks_fts`,
	`DROP TABLE IF EXISTS chunks`,
	`DROP TABLE IF EXISTS links`,
	`DROP TABLE IF EXISTS packs`,
	`DROP TABLE IF EXISTS documents`,
}

// ensureSchema creates db's schema if it doesn't exist yet, or
// recreates it from scratch if the stored PRAGMA user_version doesn't
// match schemaVersion — this index is never migrated in place
// (research.md #7). A database already at the current version is left
// completely untouched, including any data it holds.
func ensureSchema(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("index: reading schema version: %w", err)
	}
	if version == schemaVersion {
		return nil
	}

	for _, stmt := range dropStatements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("index: dropping stale schema: %w", err)
		}
	}
	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("index: creating schema: %w", err)
		}
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
		return fmt.Errorf("index: setting schema version: %w", err)
	}
	return nil
}
