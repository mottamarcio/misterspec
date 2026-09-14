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
