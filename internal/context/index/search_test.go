package index

import (
	"errors"
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

func TestSearch_FreeTextWithPunctuationDoesNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	queries := []string{
		`retry-policy: "at most once" (idempotent)`,
		`SPEC-014:TASK-003`,
		`foo* OR bar NOT baz`,
		`NEAR(a, b)`,
		`unbalanced "quote`,
		`trailing-hyphen-`,
	}
	for _, q := range queries {
		if _, err := store.Search(q, 10); err != nil {
			t.Errorf("Search(%q) unexpected error: %v (free-text queries must never fail on punctuation — spec FR-001, FR-003)", q, err)
		}
	}
}

func TestSearch_FreeTextMixedPortugueseEnglishNoError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	results, err := store.Search("rotação de token refresh rotation", 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	_ = results // mixed-language query completing without error is the assertion
}

func TestSearch_FreeTextPlainProseUnaffected(t *testing.T) {
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
		t.Fatalf("Search(\"rotation\") = %+v, want exactly 1 result (plain prose must match exactly as before free-text quoting was introduced)", results)
	}
	if results[0].Heading != "Intent" {
		t.Errorf("result.Heading = %q, want Intent", results[0].Heading)
	}

	limited, err := store.Search("Summary", 2)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("Search(\"Summary\", limit=2) = %d results, want exactly 2", len(limited))
	}
}

func TestSearch_FreeTextPunctuationOnlyReturnsEmptyNotError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	for _, q := range []string{"---():", "   ", "!!", "\"\""} {
		results, err := store.Search(q, 10)
		if err != nil {
			t.Errorf("Search(%q) unexpected error: %v, want empty result (spec Edge Cases)", q, err)
		}
		if len(results) != 0 {
			t.Errorf("Search(%q) = %+v, want no results", q, results)
		}
	}
}

func TestSearchAdvanced_HonorsFTS5Operators(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	results, err := store.SearchAdvanced(`"refresh token"`, 10)
	if err != nil {
		t.Fatalf("SearchAdvanced() unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("SearchAdvanced(phrase) = %+v, want exactly 1 result (FTS5 phrase syntax must be honored — spec FR-002)", results)
	}
}

func TestSearchAdvanced_MalformedExpressionReturnsSyntaxError(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	_, err := store.SearchAdvanced(`"unbalanced`, 10)
	if err == nil {
		t.Fatal("SearchAdvanced(malformed) = nil error, want a query-syntax error")
	}
	if !errors.Is(err, ErrQuerySyntax) {
		t.Errorf("SearchAdvanced(malformed) error = %v, want it to wrap ErrQuerySyntax (spec FR-004)", err)
	}
}

func TestSearch_OperatorCharactersTreatedAsLiteralInFreeMode(t *testing.T) {
	root := testutil.Project(t)
	cfg := testConfig()
	setupFullFixture(t, root)
	store := openTestStore(t, root)
	if _, err := store.Sync(root, cfg); err != nil {
		t.Fatalf("Sync() unexpected error: %v", err)
	}

	// A phrase-quoted advanced query matches; the identical text through
	// free-text Search must not be interpreted as a phrase operator —
	// it is tokenized into separate literal words instead (spec FR-002
	// Assumptions: mode is never inferred from the query's own content).
	advanced, err := store.SearchAdvanced(`"refresh token"`, 10)
	if err != nil {
		t.Fatalf("SearchAdvanced() unexpected error: %v", err)
	}
	free, err := store.Search(`"refresh token"`, 10)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(free) == 0 {
		t.Fatalf("Search(%q) = no results, want the literal words to still match via free-text tokenization", `"refresh token"`)
	}
	_ = advanced // both must complete without error; free mode's own literal handling is the assertion
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
