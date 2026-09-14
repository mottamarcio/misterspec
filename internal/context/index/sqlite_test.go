package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_CreatesParentDirectoryAndSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "context.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(path); err != nil {
		t.Errorf("Open() did not create the database file: %v", err)
	}
}

func TestOpen_CloseReleasesHandleWithoutError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}

func TestOpen_ReopenPreservesExistingData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	ss, ok := store.(*sqliteStore)
	if !ok {
		t.Fatalf("Open() returned %T, want *sqliteStore", store)
	}
	if _, err := ss.db.Exec("INSERT INTO documents (path, artifact_type, fingerprint, indexed_at) VALUES ('x', 'spec', 'sha256:x', 0)"); err != nil {
		t.Fatalf("seeding a row: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() unexpected error: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("Open() (reopen) unexpected error: %v", err)
	}
	defer reopened.Close()

	ss2, ok := reopened.(*sqliteStore)
	if !ok {
		t.Fatalf("Open() (reopen) returned %T, want *sqliteStore", reopened)
	}
	var count int
	if err := ss2.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&count); err != nil {
		t.Fatalf("counting documents after reopen: %v", err)
	}
	if count != 1 {
		t.Errorf("reopening lost existing data: rows = %d, want 1", count)
	}
}
