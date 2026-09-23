package index

import (
	"database/sql"
	"fmt"
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

func testStoredPack(packID string) StoredPack {
	return StoredPack{
		PackID:     packID,
		ConfigHash: "sha256:cfg",
		Target:     "SPEC-014",
		CreatedAt:  1000,
		Items: []StoredPackItem{
			{Path: "ai/.../SPEC-014/spec.md", StartLine: 10, EndLine: 20, Fingerprint: "sha256:aaa", Content: "### R1"},
		},
	}
}

// TestSavePack_LookupPack_RoundTrip is 043-incremental-context-reuse
// T005 (Foundational): SavePack followed by LookupPack with the same
// pack_id returns found == true and every field matching what was
// saved (contracts §2).
func TestSavePack_LookupPack_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	want := testStoredPack("sha256:pack1")
	if err := store.SavePack(want); err != nil {
		t.Fatalf("SavePack() unexpected error: %v", err)
	}

	got, found, err := store.LookupPack("sha256:pack1")
	if err != nil {
		t.Fatalf("LookupPack() unexpected error: %v", err)
	}
	if !found {
		t.Fatal("LookupPack() found = false, want true after SavePack")
	}
	if got.PackID != want.PackID || got.ConfigHash != want.ConfigHash || got.Target != want.Target {
		t.Errorf("LookupPack() = %+v, want matching %+v", got, want)
	}
	if len(got.Items) != 1 || got.Items[0] != want.Items[0] {
		t.Errorf("LookupPack() Items = %+v, want %+v", got.Items, want.Items)
	}
}

// TestLookupPack_UnknownPackID is 043-incremental-context-reuse T005:
// an unrecognized pack_id returns found == false and err == nil — the
// same "bool separate from error" convention used elsewhere in this
// codebase (contracts §2).
func TestLookupPack_UnknownPackID(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	_, found, err := store.LookupPack("sha256:never-saved")
	if err != nil {
		t.Fatalf("LookupPack() unexpected error: %v", err)
	}
	if found {
		t.Error("LookupPack() found = true for a pack_id that was never saved, want false")
	}
}

// TestSavePack_EvictsOldestBeyondCap is 043-incremental-context-reuse
// T005 (research.md #3): saving more packs than the fixed cap evicts
// the oldest (by created_at) first, keeping the table bounded.
func TestSavePack_EvictsOldestBeyondCap(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	for i := 0; i < packCacheCap+5; i++ {
		p := testStoredPack(fmt.Sprintf("sha256:pack%03d", i))
		p.CreatedAt = int64(i)
		if err := store.SavePack(p); err != nil {
			t.Fatalf("SavePack() #%d unexpected error: %v", i, err)
		}
	}

	if _, found, _ := store.LookupPack("sha256:pack000"); found {
		t.Error("oldest pack still present after exceeding the cap, want it evicted")
	}
	last := packCacheCap + 4
	if _, found, _ := store.LookupPack(fmt.Sprintf("sha256:pack%03d", last)); !found {
		t.Error("most recently saved pack was evicted, want it retained")
	}
}

// TestSavePack_TiedCreatedAtStillRetainsMostRecentlySaved is a code
// review finding fix: when every row shares the same created_at (a
// same-second burst of SavePack calls, e.g. a scripted eval loop),
// eviction must still keep the pack just written rather than an
// arbitrary tied row — ordering by created_at alone leaves SQLite's
// own tie resolution unspecified.
func TestSavePack_TiedCreatedAtStillRetainsMostRecentlySaved(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	for i := 0; i < packCacheCap+5; i++ {
		p := testStoredPack(fmt.Sprintf("sha256:tied%03d", i))
		p.CreatedAt = 1000 // identical for every row — forces a tie
		if err := store.SavePack(p); err != nil {
			t.Fatalf("SavePack() #%d unexpected error: %v", i, err)
		}
	}

	last := packCacheCap + 4
	if _, found, _ := store.LookupPack(fmt.Sprintf("sha256:tied%03d", last)); !found {
		t.Error("the pack saved last (all created_at tied) was evicted, want it retained")
	}
}

// TestSavePack_LostSafelyOnUnrelatedSchemaBump is 043-incremental-
// context-reuse T020 (Polish, spec FR-011, research.md #3): a pack
// saved under one schemaVersion is gone after an unrelated schema
// change triggers ensureSchema's own drop-and-recreate path — a
// missed reuse opportunity, never a correctness problem, since
// LookupPack still safely reports found == false rather than an error
// or stale data.
func TestSavePack_LostSafelyOnUnrelatedSchemaBump(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	if err := store.SavePack(testStoredPack("sha256:pack1")); err != nil {
		t.Fatalf("SavePack() unexpected error: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() unexpected error: %v", err)
	}

	// Simulate an unrelated future schema bump by forcing a stale
	// user_version directly, then reopening (mirrors
	// TestEnsureSchema_RecreatesOnVersionMismatch's own technique).
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() unexpected error: %v", err)
	}
	if _, err := raw.Exec("PRAGMA user_version = 999"); err != nil {
		t.Fatalf("forcing stale user_version: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("closing raw handle: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("Open() (reopen) unexpected error: %v", err)
	}
	defer reopened.Close()

	_, found, err := reopened.LookupPack("sha256:pack1")
	if err != nil {
		t.Fatalf("LookupPack() unexpected error: %v", err)
	}
	if found {
		t.Error("LookupPack() found the pack after an unrelated schema bump, want it lost (found == false)")
	}
}

func writeGoFixture(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", relPath, err)
	}
}

// TestSyncCode_IndexesGoFiles is 044-architecture-code-context-rules
// T007 (Foundational): SyncCode indexes every .go file under root
// except excluded paths, one code_files row plus one code_declarations
// row per top-level declaration.
func TestSyncCode_IndexesGoFiles(t *testing.T) {
	dir := t.TempDir()
	writeGoFixture(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")
	writeGoFixture(t, dir, "vendor/other/other.go", "package other\n\nfunc Baz() {}\n")

	store, err := Open(filepath.Join(dir, ".cache", "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	report, err := store.SyncCode(dir, []string{"vendor/**"})
	if err != nil {
		t.Fatalf("SyncCode() unexpected error: %v", err)
	}
	if report.Indexed != 1 {
		t.Errorf("Indexed = %d, want 1 (vendor/** excluded)", report.Indexed)
	}

	decls, err := store.DeclarationsForFiles([]string{"internal/foo/foo.go"})
	if err != nil {
		t.Fatalf("DeclarationsForFiles() unexpected error: %v", err)
	}
	if len(decls) != 1 || decls[0].Name != "Bar" {
		t.Errorf("DeclarationsForFiles() = %+v, want one declaration named Bar", decls)
	}

	excluded, err := store.DeclarationsForFiles([]string{"vendor/other/other.go"})
	if err != nil {
		t.Fatalf("DeclarationsForFiles() unexpected error: %v", err)
	}
	if len(excluded) != 0 {
		t.Errorf("DeclarationsForFiles(vendor path) = %+v, want none (excluded)", excluded)
	}
}

// TestSyncCode_UnchangedFileSkipped is 044-architecture-code-context-
// rules T007: a second SyncCode call with no filesystem change leaves
// the file's own fingerprint unchanged (reported as Skipped, not
// re-Indexed).
func TestSyncCode_UnchangedFileSkipped(t *testing.T) {
	dir := t.TempDir()
	writeGoFixture(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")

	store, err := Open(filepath.Join(dir, ".cache", "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	if _, err := store.SyncCode(dir, nil); err != nil {
		t.Fatalf("SyncCode() (first) unexpected error: %v", err)
	}
	report, err := store.SyncCode(dir, nil)
	if err != nil {
		t.Fatalf("SyncCode() (second) unexpected error: %v", err)
	}
	if report.Skipped != 1 || report.Indexed != 0 {
		t.Errorf("second SyncCode() report = %+v, want Skipped=1, Indexed=0", report)
	}
}

// TestSyncCode_ChangedFileReplacesDeclarations is
// 044-architecture-code-context-rules T007: a changed file's own
// declarations are replaced, not duplicated.
func TestSyncCode_ChangedFileReplacesDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeGoFixture(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")

	store, err := Open(filepath.Join(dir, ".cache", "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	if _, err := store.SyncCode(dir, nil); err != nil {
		t.Fatalf("SyncCode() (first) unexpected error: %v", err)
	}

	writeGoFixture(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n\nfunc Qux() {}\n")
	if _, err := store.SyncCode(dir, nil); err != nil {
		t.Fatalf("SyncCode() (second) unexpected error: %v", err)
	}

	decls, err := store.DeclarationsForFiles([]string{"internal/foo/foo.go"})
	if err != nil {
		t.Fatalf("DeclarationsForFiles() unexpected error: %v", err)
	}
	if len(decls) != 2 {
		t.Errorf("DeclarationsForFiles() = %+v, want exactly 2 (no duplicates from the first sync)", decls)
	}
}

// TestSyncCode_DeletedFileRemoved is 044-architecture-code-context-
// rules T007: a deleted file's own rows are removed.
func TestSyncCode_DeletedFileRemoved(t *testing.T) {
	dir := t.TempDir()
	writeGoFixture(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")

	store, err := Open(filepath.Join(dir, ".cache", "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	if _, err := store.SyncCode(dir, nil); err != nil {
		t.Fatalf("SyncCode() (first) unexpected error: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "internal/foo/foo.go")); err != nil {
		t.Fatalf("removing fixture: %v", err)
	}
	report, err := store.SyncCode(dir, nil)
	if err != nil {
		t.Fatalf("SyncCode() (second) unexpected error: %v", err)
	}
	if report.Removed != 1 {
		t.Errorf("Removed = %d, want 1", report.Removed)
	}

	decls, err := store.DeclarationsForFiles([]string{"internal/foo/foo.go"})
	if err != nil {
		t.Fatalf("DeclarationsForFiles() unexpected error: %v", err)
	}
	if len(decls) != 0 {
		t.Errorf("DeclarationsForFiles() after deletion = %+v, want none", decls)
	}
}

// TestDeclarationsForFiles_IncludesAssociatedTests is
// 044-architecture-code-context-rules T008 (Foundational):
// DeclarationsForFiles includes declarations from a file's own
// associated _test.go file in the same directory (spec FR-007).
func TestDeclarationsForFiles_IncludesAssociatedTests(t *testing.T) {
	dir := t.TempDir()
	writeGoFixture(t, dir, "internal/foo/foo.go", "package foo\n\nfunc Bar() {}\n")
	writeGoFixture(t, dir, "internal/foo/foo_test.go", "package foo\n\nimport \"testing\"\n\nfunc TestBar(t *testing.T) {}\n")

	store, err := Open(filepath.Join(dir, ".cache", "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	if _, err := store.SyncCode(dir, nil); err != nil {
		t.Fatalf("SyncCode() unexpected error: %v", err)
	}

	decls, err := store.DeclarationsForFiles([]string{"internal/foo/foo.go"})
	if err != nil {
		t.Fatalf("DeclarationsForFiles() unexpected error: %v", err)
	}

	var sawBar, sawTestBar bool
	for _, d := range decls {
		switch d.Name {
		case "Bar":
			sawBar = true
			if d.IsTest {
				t.Errorf("Bar.IsTest = true, want false")
			}
		case "TestBar":
			sawTestBar = true
			if !d.IsTest {
				t.Errorf("TestBar.IsTest = false, want true")
			}
		}
	}
	if !sawBar || !sawTestBar {
		t.Errorf("decls = %+v, want both Bar and its associated TestBar", decls)
	}
}

// TestDeclarationsForFiles_UnknownPathReturnsNoRows is
// 044-architecture-code-context-rules T008: a path with no indexed
// file returns no rows for it, no error.
func TestDeclarationsForFiles_UnknownPathReturnsNoRows(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, ".cache", "context.db"))
	if err != nil {
		t.Fatalf("Open() unexpected error: %v", err)
	}
	defer store.Close()

	decls, err := store.DeclarationsForFiles([]string{"internal/never/indexed.go"})
	if err != nil {
		t.Fatalf("DeclarationsForFiles() unexpected error: %v", err)
	}
	if len(decls) != 0 {
		t.Errorf("DeclarationsForFiles() = %+v, want none", decls)
	}
}
