package artifacts_test

import (
	"reflect"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestChunks_OnePerNonEmptySectionInOrder(t *testing.T) {
	body := "## A\ntext a\n## B\ntext b\n## C\ntext c\n"
	doc := artifacts.ParseDocument([]byte(body))

	chunks := artifacts.Chunks("ai/.../SPEC-014/spec.md", doc)

	if len(chunks) != 3 {
		t.Fatalf("len(chunks) = %d, want 3: %+v", len(chunks), chunks)
	}
	want := []string{"A", "B", "C"}
	for i, w := range want {
		if chunks[i].Heading != w {
			t.Errorf("chunks[%d].Heading = %q, want %q", i, chunks[i].Heading, w)
		}
		if chunks[i].Path != "ai/.../SPEC-014/spec.md" {
			t.Errorf("chunks[%d].Path = %q, unexpected", i, chunks[i].Path)
		}
	}
}

func TestChunks_NoChunkForEmptySection(t *testing.T) {
	body := "## Requirements\n\n### R1\nReq 1 text.\n"
	doc := artifacts.ParseDocument([]byte(body))

	chunks := artifacts.Chunks("ai/.../SPEC-014/spec.md", doc)

	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1 (Requirements' own empty section produces none): %+v", len(chunks), chunks)
	}
	if chunks[0].Heading != "R1" {
		t.Errorf("chunks[0].Heading = %q, want R1", chunks[0].Heading)
	}
}

func TestChunks_ProvenanceMatchesOriginatingSection(t *testing.T) {
	body := "## Intro\nSome intro text.\n"
	doc := artifacts.ParseDocument([]byte(body))
	section := doc.Sections[0]

	chunks := artifacts.Chunks("ai/x/spec.md", doc)

	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	c := chunks[0]
	if c.Path != "ai/x/spec.md" {
		t.Errorf("Path = %q, want ai/x/spec.md", c.Path)
	}
	if c.Heading != section.Heading || c.Level != section.Level || c.Content != section.Body ||
		c.StartLine != section.StartLine || c.EndLine != section.EndLine {
		t.Errorf("Chunk = %+v, want it to match Section %+v exactly (Content <- Body)", c, section)
	}
}

func TestChunks_PreservesWikilinkSyntaxVerbatim(t *testing.T) {
	body := "## Relevant Knowledge\nSee [[KNOW-003|Authentication Model]] for details.\n"
	doc := artifacts.ParseDocument([]byte(body))

	chunks := artifacts.Chunks("ai/x/spec.md", doc)

	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	want := "See [[KNOW-003|Authentication Model]] for details."
	if chunks[0].Content != want {
		t.Errorf("Content = %q, want %q — wikilink syntax must be preserved exactly, unmodified and unresolved", chunks[0].Content, want)
	}
}

func TestChunks_Deterministic(t *testing.T) {
	body := "## A\ntext a\n## B\ntext b\n"
	doc := artifacts.ParseDocument([]byte(body))

	first := artifacts.Chunks("ai/x/spec.md", doc)
	second := artifacts.Chunks("ai/x/spec.md", doc)

	if !reflect.DeepEqual(first, second) {
		t.Errorf("Chunks() is not deterministic: %+v != %+v", first, second)
	}
}

func TestChunks_EmptyDocumentProducesNoChunks(t *testing.T) {
	doc := artifacts.ParseDocument([]byte(""))

	chunks := artifacts.Chunks("ai/x/spec.md", doc)

	if len(chunks) != 0 {
		t.Errorf("len(chunks) = %d, want 0: %+v", len(chunks), chunks)
	}
}
