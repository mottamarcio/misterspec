package artifacts_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestReadBody_ReturnsPostFrontmatterContent(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", ""+
		"---\n"+
		"id: KNOW-001\n"+
		"type: knowledge\n"+
		"status: active\n"+
		"---\n"+
		"# Authentication\n\n## Summary\n\nSome facts.\n")

	body, err := artifacts.ReadBody(path)
	if err != nil {
		t.Fatalf("ReadBody() unexpected error: %v", err)
	}

	want := "# Authentication\n\n## Summary\n\nSome facts.\n"
	if string(body) != want {
		t.Errorf("ReadBody() = %q, want %q", body, want)
	}
}

func TestReadBody_EmptyBody(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md", ""+
		"---\n"+
		"id: KNOW-002\n"+
		"type: knowledge\n"+
		"status: active\n"+
		"---\n")

	body, err := artifacts.ReadBody(path)
	if err != nil {
		t.Fatalf("ReadBody() unexpected error: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("ReadBody() = %q, want empty, not an error", body)
	}
}

func TestReadBody_MissingFrontmatterDelimiters(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-z.md", "# Just a heading, no frontmatter\n")

	_, err := artifacts.ReadBody(path)
	if !errors.Is(err, artifacts.ErrFrontmatterMalformed) {
		t.Fatalf("ReadBody() error = %v, want errors.Is(err, ErrFrontmatterMalformed) — the same vocabulary ParseMetadata already uses", err)
	}
}

func TestReadBody_ArtifactNotFound(t *testing.T) {
	root := testutil.Project(t)

	_, err := artifacts.ReadBody(root + "/ai/knowledge/KNOW-999-missing.md")
	if !errors.Is(err, artifacts.ErrArtifactNotFound) {
		t.Fatalf("ReadBody() error = %v, want errors.Is(err, ErrArtifactNotFound)", err)
	}
}

// TestReadBodyWithOffset_ReportsFileAbsoluteBodyStart covers
// 033-context-pack-output-contract research.md Decision 1: the body's
// own 1-indexed starting line in the whole file, accounting for the
// opening/closing "---" delimiters plus every frontmatter line.
func TestReadBodyWithOffset_ReportsFileAbsoluteBodyStart(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-001-x.md", ""+
		"---\n"+ // line 1
		"id: KNOW-001\n"+ // line 2
		"type: knowledge\n"+ // line 3
		"status: active\n"+ // line 4
		"---\n"+ // line 5
		"# Authentication\n\n## Summary\n\nSome facts.\n") // body starts at line 6

	body, offset, err := artifacts.ReadBodyWithOffset(path)
	if err != nil {
		t.Fatalf("ReadBodyWithOffset() unexpected error: %v", err)
	}
	if offset != 6 {
		t.Errorf("offset = %d, want 6 (the body's own file-absolute starting line)", offset)
	}
	want := "# Authentication\n\n## Summary\n\nSome facts.\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

// TestReadBodyWithOffset_EmptyFrontmatter covers the degenerate case:
// an immediately-closed frontmatter block still yields the correct
// offset (no off-by-one from an empty frontmatter string).
func TestReadBodyWithOffset_EmptyFrontmatter(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-002-y.md", "---\n---\nBody text.\n")

	_, offset, err := artifacts.ReadBodyWithOffset(path)
	if err != nil {
		t.Fatalf("ReadBodyWithOffset() unexpected error: %v", err)
	}
	if offset != 3 {
		t.Errorf("offset = %d, want 3", offset)
	}
}

// TestReadBodyWithOffset_MissingFrontmatterDelimiters covers that
// ReadBodyWithOffset shares ReadBody's existing error vocabulary for a
// file with no frontmatter at all — every canonical misterspec artifact
// requires frontmatter, so this stays an error, not a zero-offset
// fallback.
func TestReadBodyWithOffset_MissingFrontmatterDelimiters(t *testing.T) {
	root := testutil.Project(t)
	path := testutil.WriteFile(t, root, "ai/knowledge/KNOW-003-z.md", "# Just a heading, no frontmatter\n")

	_, _, err := artifacts.ReadBodyWithOffset(path)
	if !errors.Is(err, artifacts.ErrFrontmatterMalformed) {
		t.Fatalf("ReadBodyWithOffset() error = %v, want errors.Is(err, ErrFrontmatterMalformed)", err)
	}
}
