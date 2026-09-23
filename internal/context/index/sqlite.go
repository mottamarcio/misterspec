package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// packCacheCap bounds the packs table's own row count
// (043-incremental-context-reuse research.md #3) — a fixed, simple
// eviction policy (oldest created_at first) rather than a
// configurable knob, since no evidence yet justifies making this
// tunable (Constitution Principle IV).
const packCacheCap = 50

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

// SavePack upserts pack into the packs table, then evicts the oldest
// rows beyond packCacheCap (043-incremental-context-reuse research.md
// #3).
func (s *sqliteStore) SavePack(pack StoredPack) error {
	itemsJSON, err := json.Marshal(pack.Items)
	if err != nil {
		return fmt.Errorf("index: marshaling pack items for %s: %w", pack.PackID, err)
	}

	_, err = s.db.Exec(
		`INSERT INTO packs (pack_id, config_hash, target, created_at, items_json) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(pack_id) DO UPDATE SET config_hash = excluded.config_hash, target = excluded.target, created_at = excluded.created_at, items_json = excluded.items_json`,
		pack.PackID, pack.ConfigHash, pack.Target, pack.CreatedAt, string(itemsJSON),
	)
	if err != nil {
		return fmt.Errorf("index: saving pack %s: %w", pack.PackID, err)
	}

	// ORDER BY created_at DESC, rowid DESC: created_at alone (Unix
	// seconds) ties within a same-second burst of SavePack calls, and
	// SQLite's own tie resolution for an unordered LIMIT is otherwise
	// unspecified — rowid (the table's own implicit, monotonically
	// assigned column, unaffected by ON CONFLICT DO UPDATE on an
	// existing row) breaks the tie in actual insertion order, so the
	// just-saved pack is never the one evicted (code review finding).
	_, err = s.db.Exec(
		`DELETE FROM packs WHERE pack_id NOT IN (SELECT pack_id FROM packs ORDER BY created_at DESC, rowid DESC LIMIT ?)`,
		packCacheCap,
	)
	if err != nil {
		return fmt.Errorf("index: evicting old packs: %w", err)
	}
	return nil
}

// LookupPack returns the stored pack for packID, and found == false
// (never an error) when no such row exists (043-incremental-context-
// reuse research.md #4).
func (s *sqliteStore) LookupPack(packID string) (pack StoredPack, found bool, err error) {
	var itemsJSON string
	row := s.db.QueryRow(
		`SELECT pack_id, config_hash, target, created_at, items_json FROM packs WHERE pack_id = ?`,
		packID,
	)
	if err := row.Scan(&pack.PackID, &pack.ConfigHash, &pack.Target, &pack.CreatedAt, &itemsJSON); err != nil {
		if err == sql.ErrNoRows {
			return StoredPack{}, false, nil
		}
		return StoredPack{}, false, fmt.Errorf("index: looking up pack %s: %w", packID, err)
	}

	if err := json.Unmarshal([]byte(itemsJSON), &pack.Items); err != nil {
		return StoredPack{}, false, fmt.Errorf("index: unmarshaling pack items for %s: %w", packID, err)
	}
	return pack, true, nil
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
