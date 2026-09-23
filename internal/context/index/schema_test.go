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

	for _, table := range []string{"documents", "chunks", "chunks_fts", "links", "packs", "code_files", "code_declarations"} {
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

func TestEnsureSchema_LinksTableHasSourceSectionAndSourceLineColumns(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	_, err := db.Exec(
		`INSERT INTO links (source_artifact_id, target_artifact_id, relation, source_section, source_line) VALUES ('SPEC-001', 'SPEC-002', 'wikilink', 'Requirements', 42)`,
	)
	if err != nil {
		t.Fatalf("links table does not accept source_section/source_line: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO links (source_artifact_id, target_artifact_id, relation, source_section, source_line) VALUES ('SPEC-001', 'FEAT-001', 'parent', NULL, NULL)`,
	)
	if err != nil {
		t.Errorf("links table does not accept NULL source_section/source_line (formal relation): %v", err)
	}
}

func TestEnsureSchema_RecreatesFromPreProvenanceSchemaVersion1(t *testing.T) {
	db := openRawDB(t)

	// Simulate a real pre-038 database: the old links shape, tagged
	// schemaVersion 1 (038-wikilink-chunk-provenance contract §6 — "a
	// caller holding an existing cache database built under
	// schemaVersion: 1 observes no manual step").
	if _, err := db.Exec(`CREATE TABLE links (
		id                 INTEGER PRIMARY KEY,
		source_artifact_id TEXT NOT NULL,
		target_artifact_id TEXT NOT NULL,
		relation           TEXT NOT NULL
	)`); err != nil {
		t.Fatalf("seeding pre-provenance links table: %v", err)
	}
	if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
		t.Fatalf("seeding stale user_version: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	_, err := db.Exec(
		`INSERT INTO links (source_artifact_id, target_artifact_id, relation, source_section, source_line) VALUES ('SPEC-001', 'SPEC-002', 'wikilink', 'Requirements', 42)`,
	)
	if err != nil {
		t.Errorf("schema was not transparently rebuilt from version 1: %v", err)
	}
	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != schemaVersion {
		t.Errorf("user_version after rebuild = %d, want %d", version, schemaVersion)
	}
}

// TestEnsureSchema_ChunksAndLinksHaveAnchorColumns proves spec 040
// data-model.md "Index Schema (extended)": chunks.anchor and
// links.target_anchor exist and accept both a real value and NULL.
func TestEnsureSchema_ChunksAndLinksHaveAnchorColumns(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO documents (path, artifact_type, fingerprint, indexed_at) VALUES ('x', 'knowledge', 'sha256:x', 0)`); err != nil {
		t.Fatalf("seeding a document row: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO chunks (document_id, heading, anchor, content, start_line, end_line, token_estimate) VALUES (1, 'Retry Policy', 'retry-policy', 'body', 1, 2, 3)`); err != nil {
		t.Fatalf("chunks table does not accept an anchor value: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO chunks (document_id, heading, anchor, content, start_line, end_line, token_estimate) VALUES (1, 'No Anchor', NULL, 'body', 4, 5, 3)`); err != nil {
		t.Errorf("chunks table does not accept NULL anchor: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO links (source_artifact_id, target_artifact_id, relation, target_anchor) VALUES ('SPEC-001', 'KNOW-003', 'wikilink', 'retry-policy')`); err != nil {
		t.Fatalf("links table does not accept a target_anchor value: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO links (source_artifact_id, target_artifact_id, relation, target_anchor) VALUES ('SPEC-001', 'FEAT-001', 'parent', NULL)`); err != nil {
		t.Errorf("links table does not accept NULL target_anchor: %v", err)
	}
}

// TestEnsureSchema_RecreatesFromPreAnchorSchemaVersion2 proves spec 040
// research.md #5: a database built under schemaVersion 2 (038's own
// shape, no anchor columns) is detected as stale and transparently
// rebuilt under schemaVersion 3 — no manual migration.
func TestEnsureSchema_RecreatesFromPreAnchorSchemaVersion2(t *testing.T) {
	db := openRawDB(t)

	if _, err := db.Exec(`CREATE TABLE links (
		id                 INTEGER PRIMARY KEY,
		source_artifact_id TEXT NOT NULL,
		target_artifact_id TEXT NOT NULL,
		relation           TEXT NOT NULL,
		source_section     TEXT,
		source_line        INTEGER
	)`); err != nil {
		t.Fatalf("seeding pre-anchor links table: %v", err)
	}
	if _, err := db.Exec("PRAGMA user_version = 2"); err != nil {
		t.Fatalf("seeding stale user_version: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO links (source_artifact_id, target_artifact_id, relation, target_anchor) VALUES ('SPEC-001', 'KNOW-003', 'wikilink', 'retry-policy')`,
	); err != nil {
		t.Errorf("schema was not transparently rebuilt from version 2: %v", err)
	}
	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != schemaVersion {
		t.Errorf("user_version after rebuild = %d, want %d", version, schemaVersion)
	}
}

// TestEnsureSchema_PacksTableHasExpectedColumns is
// 043-incremental-context-reuse T004 (Foundational): the packs table
// accepts pack_id/config_hash/target/created_at/items_json, and
// pack_id enforces uniqueness as its own primary key
// (data-model.md "StoredPack").
func TestEnsureSchema_PacksTableHasExpectedColumns(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	_, err := db.Exec(
		`INSERT INTO packs (pack_id, config_hash, target, created_at, items_json) VALUES ('sha256:aaa', 'sha256:bbb', 'SPEC-014', 100, '[]')`,
	)
	if err != nil {
		t.Fatalf("packs table does not accept the expected columns: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO packs (pack_id, config_hash, target, created_at, items_json) VALUES ('sha256:aaa', 'sha256:ccc', 'SPEC-020', 200, '[]')`,
	)
	if err == nil {
		t.Error("inserting a duplicate pack_id succeeded, want a primary-key violation")
	}
}

// TestEnsureSchema_RecreatesFromPreV4SchemaVersion3 is
// 043-incremental-context-reuse T004: a database built under
// schemaVersion 3 (040's own shape, no packs table) is detected as
// stale and transparently rebuilt under schemaVersion 4 — no manual
// migration, mirroring every prior schema-bump precedent in this file.
func TestEnsureSchema_RecreatesFromPreV4SchemaVersion3(t *testing.T) {
	db := openRawDB(t)

	if _, err := db.Exec("PRAGMA user_version = 3"); err != nil {
		t.Fatalf("seeding stale user_version: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO packs (pack_id, config_hash, target, created_at, items_json) VALUES ('sha256:aaa', 'sha256:bbb', 'SPEC-014', 100, '[]')`,
	); err != nil {
		t.Errorf("schema was not rebuilt with a packs table from version 3: %v", err)
	}
	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != schemaVersion {
		t.Errorf("user_version after rebuild = %d, want %d", version, schemaVersion)
	}
}

// TestEnsureSchema_CodeTablesHaveExpectedColumns is
// 044-architecture-code-context-rules T006 (Foundational): code_files
// and code_declarations accept the columns data-model.md documents.
func TestEnsureSchema_CodeTablesHaveExpectedColumns(t *testing.T) {
	db := openRawDB(t)

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO code_files (path, fingerprint, indexed_at) VALUES ('internal/foo/foo.go', 'sha256:aaa', 100)`,
	); err != nil {
		t.Fatalf("code_files table does not accept the expected columns: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO code_files (path, fingerprint, indexed_at) VALUES ('internal/foo/foo.go', 'sha256:bbb', 200)`,
	); err == nil {
		t.Error("inserting a duplicate code_files.path succeeded, want a uniqueness violation")
	}

	if _, err := db.Exec(
		`INSERT INTO code_declarations (code_file_id, name, kind, signature, body, start_line, end_line, is_test) VALUES (1, 'Foo', 'func', 'func Foo()', 'func Foo() {}', 1, 3, 0)`,
	); err != nil {
		t.Fatalf("code_declarations table does not accept the expected columns: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO code_declarations (code_file_id, name, kind, signature, body, start_line, end_line, is_test) VALUES (1, 'Bar', 'func', 'func Bar()', NULL, 5, 5, 1)`,
	); err != nil {
		t.Errorf("code_declarations table does not accept a NULL body: %v", err)
	}
}

// TestEnsureSchema_RecreatesFromPreV5SchemaVersion4 is
// 044-architecture-code-context-rules T006: a database built under
// schemaVersion 4 (043's own shape, no code tables) is detected as
// stale and transparently rebuilt — no manual migration.
func TestEnsureSchema_RecreatesFromPreV5SchemaVersion4(t *testing.T) {
	db := openRawDB(t)

	if _, err := db.Exec("PRAGMA user_version = 4"); err != nil {
		t.Fatalf("seeding stale user_version: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		t.Fatalf("ensureSchema() unexpected error: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO code_files (path, fingerprint, indexed_at) VALUES ('internal/foo/foo.go', 'sha256:aaa', 100)`,
	); err != nil {
		t.Errorf("schema was not rebuilt with code_files from version 4: %v", err)
	}
	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != schemaVersion {
		t.Errorf("user_version after rebuild = %d, want %d", version, schemaVersion)
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
