package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// sqliteStore is the concrete, modernc.org/sqlite-backed Store.
type sqliteStore struct {
	db *sql.DB
}

// Open opens (creating if necessary) the SQLite database at path,
// creating path's parent directory if missing, and creating or
// recreating its schema as needed (research.md #7). The returned Store
// must be Close()d when no longer needed.
func Open(path string) (Store, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("index: creating %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("index: opening %s: %w", path, err)
	}

	if err := ensureSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &sqliteStore{db: db}, nil
}

// Close releases the underlying database handle.
func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// Outgoing reports id's already-indexed outgoing relationships (FR-008),
// sourced from links rows operations.References itself populated during
// Sync/Rebuild (research.md #4) — consistent with that direct
// computation by construction (FR-009).
func (s *sqliteStore) Outgoing(id string) ([]Link, error) {
	return s.queryLinks("SELECT relation, source_artifact_id, target_artifact_id FROM links WHERE source_artifact_id = ?", id)
}

// Incoming reports id's already-indexed incoming relationships (FR-008),
// the inverse of Outgoing.
func (s *sqliteStore) Incoming(id string) ([]Link, error) {
	return s.queryLinks("SELECT relation, source_artifact_id, target_artifact_id FROM links WHERE target_artifact_id = ?", id)
}

func (s *sqliteStore) queryLinks(query, id string) ([]Link, error) {
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("index: querying links for %s: %w", id, err)
	}
	defer rows.Close()

	links := make([]Link, 0)
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.Relation, &l.Source, &l.Target); err != nil {
			return nil, fmt.Errorf("index: reading link row: %w", err)
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("index: reading link rows: %w", err)
	}
	return links, nil
}
