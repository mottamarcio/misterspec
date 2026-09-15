package index

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestSearch_ReturnsMatchingChunkWithProvenance(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	results, err := store.Search("rotation", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search(\"rotation\") = %+v, want exactly 1 result", results)
	}
	r := results[0]
	if r.Path != "ai/programs/PRG-001/features/FEAT-001/specs/SPEC-001/spec.md" {
		t.Errorf("result.Path = %q, unexpected", r.Path)
	}
	if r.Heading != "Intent" {
		t.Errorf("result.Heading = %q, want Intent", r.Heading)
	}
	if r.StartLine == 0 || r.EndLine == 0 {
		t.Errorf("result line range = [%d,%d], want non-zero", r.StartLine, r.EndLine)
	}
}

func TestSearch_NoMatchReturnsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	results, err := store.Search("xyzzyunmatchable", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() = %+v, want no results", results)
	}
}

func TestSearch_RespectsLimit(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	// "Summary" appears as a heading in three of the fixture's own
	// artifacts (Plan, Knowledge, Learning).
	results, err := store.Search("Summary", 2)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Search(limit=2) = %d results, want exactly 2", len(results))
	}
}
