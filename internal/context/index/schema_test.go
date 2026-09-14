package index

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openRawDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestEnsureSchema_CreatesAllTables(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	for _, table := range []string{"documents", "chunks", "chunks_fts", "links"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type IN ('table') AND name = ?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not created: %v", table, err)
		}
	}
}

func TestEnsureSchema_SetsUserVersion(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("reading PRAGMA user_version: %v", err)
	}
	if version != schemaVersion {
		t.Errorf("user_version = %d, want %d (schemaVersion)", version, schemaVersion)
	}
}

func TestEnsureSchema_RecreatesOnVersionMismatch(t *testing.T) {
	db := openRawDB(t)

	// Simulate a stale, incompatible schema left by an older version of
	// this package.
	if _, err := db.Exec("CREATE TABLE documents (id INTEGER)"); err != nil {
		t.Fatalf("seeding stale schema: %v", err)
	}
	if _, err := db.Exec("PRAGMA user_version = 999"); err != nil {
		t.Fatalf("seeding stale user_version: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	// The real schema's own documents columns must now exist.
	_, err := db.Exec("INSERT INTO documents (path, artifact_type, fingerprint, indexed_at) VALUES ('x', 'spec', 'sha256:x', 0)")
	if err != nil {
		t.Errorf("schema was not recreated with the current shape: %v", err)
	}

	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != schemaVersion {
		t.Errorf("user_version after recreation = %d, want %d", version, schemaVersion)
	}
}

func TestEnsureSchema_IdempotentWhenVersionMatches(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}
	if _, err := db.Exec("INSERT INTO documents (path, artifact_type, fingerprint, indexed_at) VALUES ('x', 'spec', 'sha256:x', 0)"); err != nil {
		t.Fatalf("seeding a row: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() (second call) unexpected error: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&count); err != nil {
		t.Fatalf("counting documents: %v", err)
	}
	if count != 1 {
		t.Errorf("ensureSchema() wiped existing data when the version already matched: rows = %d, want 1", count)
	}
}
