package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestSync_InitialBuild_IndexesEveryArtifactType(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)

	report, err := store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	if report.Indexed != 9 {
		t.Errorf("report.Indexed = %d, want 9 (one per ArtifactType value): %+v", report.Indexed, report)
	}
	if len(report.Errors) != 0 {
		t.Errorf("report.Errors = %+v, want none", report.Errors)
	}

	var docCount, chunkCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("counting documents: %v", err)
	}
	if docCount != 9 {
		t.Errorf("documents row count = %d, want 9", docCount)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM chunks").Scan(&chunkCount); err != nil {
		t.Fatalf("counting chunks: %v", err)
	}
	if chunkCount == 0 {
		t.Error("chunks row count = 0, want at least one chunk per indexed artifact")
	}
}

func TestSync_UnclassifiableContentIsSkippedNotErrored(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupUnrelatedFixture(t, root)
	store := openTestStore(t, root)

	report, err := store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	if report.Indexed != 0 {
		t.Errorf("report.Indexed = %d, want 0 — unclassifiable content must never be indexed", report.Indexed)
	}
	if len(report.Errors) != 0 {
		t.Errorf("report.Errors = %+v, want none — an unrecognized file is skipped silently, not reported as an error", report.Errors)
	}
}

func TestRebuild_AlreadyPopulatedIndexStaysSearchEquivalent(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	before, err := store.Search("rotation", 10)
	if err != nil {
		t.Fatalf("Search() before rebuild unexpected error: %v", err)
	}

	report, err := store.Rebuild(root, cfg)
	if err != nil {
		t.Fatalf("Rebuild() unexpected error: %v", err)
	}
	if report.Indexed != 9 {
		t.Errorf("Rebuild() report.Indexed = %d, want 9", report.Indexed)
	}

	after, err := store.Search("rotation", 10)
	if err != nil {
		t.Fatalf("Search() after rebuild unexpected error: %v", err)
	}
	if len(before) != len(after) || len(after) != 1 || before[0].Path != after[0].Path {
		t.Errorf("Rebuild() is not search-equivalent: before=%+v after=%+v", before, after)
	}
}

func TestRebuild_ConvergesToCurrentStateWithNoStaleRows(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	// The project's content changes completely between the last index
	// and the rebuild.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nCompletely different content now.\n")

	if _, err := store.Rebuild(root, cfg); err != nil {
		t.Fatalf("Rebuild() unexpected error: %v", err)
	}

	stale, err := store.Search("rotation", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("Search(\"rotation\") after rebuild = %+v, want none — the old content must not survive a rebuild", stale)
	}
	fresh, err := store.Search("Completely different", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(fresh) != 1 {
		t.Errorf("Search(\"Completely different\") after rebuild = %+v, want exactly 1", fresh)
	}
}

func TestSync_Incremental_MixedBatch(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)

	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() (first) unexpected error: %v", err)
	}

	var untouchedIndexedAt int64
	if err := store.db.QueryRow("SELECT indexed_at FROM documents WHERE path = ?", "ai/knowledge/KNOW-001-auth.md").Scan(&untouchedIndexedAt); err != nil {
		t.Fatalf("reading baseline indexed_at: %v", err)
	}

	// Change one artifact.
	testutil.WriteFile(t, root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md",
		"---\nid: SPEC-001\ntype: spec\nstatus: ready\nparent: FEAT-001\ndepends_on: []\nsupersedes: []\n---\n## Intent\n\nSomething entirely new.\n")
	// Add one new artifact.
	testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-payments.md",
		"---\nid: KNOW-002\ntype: knowledge\nstatus: active\n---\n## Summary\n\nPayments facts.\n")
	// Delete one artifact.
	if err := os.Remove(filepath.Join(root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/validation.md")); err != nil {
		t.Fatalf("removing validation.md: %v", err)
	}

	report, err := store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() (second) unexpected error: %v", err)
	}

	if report.Indexed != 1 {
		t.Errorf("report.Indexed = %d, want 1 (KNOW-002)", report.Indexed)
	}
	if report.Updated != 1 {
		t.Errorf("report.Updated = %d, want 1 (SPEC-001)", report.Updated)
	}
	if report.Removed != 1 {
		t.Errorf("report.Removed = %d, want 1 (validation.md)", report.Removed)
	}
	if report.Skipped != 7 {
		t.Errorf("report.Skipped = %d, want 7 (everything else, untouched)", report.Skipped)
	}

	var untouchedIndexedAtAfter int64
	if err := store.db.QueryRow("SELECT indexed_at FROM documents WHERE path = ?", "ai/knowledge/KNOW-001-auth.md").Scan(&untouchedIndexedAtAfter); err != nil {
		t.Fatalf("reading indexed_at after second sync: %v", err)
	}
	if untouchedIndexedAtAfter != untouchedIndexedAt {
		t.Errorf("KNOW-001's indexed_at changed (%d -> %d) even though it was never touched", untouchedIndexedAt, untouchedIndexedAtAfter)
	}

	results, err := store.Search("entirely new", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search(\"entirely new\") = %+v, want exactly 1 — the changed artifact's new content must be indexed", results)
	}

	var validationCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/validation.md").Scan(&validationCount); err != nil {
		t.Fatalf("counting validation.md rows: %v", err)
	}
	if validationCount != 0 {
		t.Errorf("validation.md still has %d documents row(s) after deletion, want 0", validationCount)
	}
}

func TestSync_TransactionalRollbackOnFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root — permission-based failure injection does not apply")
	}

	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)

	blocked := filepath.Join(root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001")
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(blocked, 0o755)

	store := openTestStore(t, root)

	if _, err := store.Sync(root, cfg); err == nil {
		t.Fatal("Sync() expected an error from the unreadable directory, got nil")
	}

	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&count); err != nil {
		t.Fatalf("counting documents: %v", err)
	}
	if count != 0 {
		t.Errorf("documents count = %d after a failed Sync, want 0 — the whole transaction must have rolled back", count)
	}
}

func TestSync_NeverModifiesProjectFiles(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	specPath := filepath.Join(root, "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md")

	before, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading fixture before Sync: %v", err)
	}

	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}
	if _, err := store.Rebuild(root, cfg); err != nil {
		t.Fatalf("Rebuild() unexpected error: %v", err)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading fixture after Sync/Rebuild: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("SPEC-001's own file content changed after Sync/Rebuild — this feature must be strictly read-only with respect to project artifacts (FR-011)")
	}
}

func TestSync_NoArtifactsDirectoryYetIsNotAnError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	store := openTestStore(t, root)

	report, err := store.Sync(root, cfg)
	if err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}
	if report.Indexed != 0 {
		t.Errorf("report.Indexed = %d, want 0", report.Indexed)
	}
}
