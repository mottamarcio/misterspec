package artifacts_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestOccurrenceFor_LinkInsideChunkReturnsItsHeading(t *testing.T) {
	chunks := []artifacts.Chunk{
		{Heading: "Intent", StartLine: 6, EndLine: 9},
		{Heading: "Requirements", StartLine: 10, EndLine: 20},
	}
	link := artifacts.WikiLink{Target: "KNOW-001", Line: 15}

	got := artifacts.OccurrenceFor(link, chunks)

	if got.SourceSection != "Requirements" {
		t.Errorf("SourceSection = %q, want %q", got.SourceSection, "Requirements")
	}
	if got.SourceLine != 15 {
		t.Errorf("SourceLine = %d, want 15", got.SourceLine)
	}
}

func TestOccurrenceFor_LinkBeforeAnyHeadingReturnsEmptySection(t *testing.T) {
	chunks := []artifacts.Chunk{
		{Heading: "Intent", StartLine: 6, EndLine: 9},
	}
	link := artifacts.WikiLink{Target: "KNOW-001", Line: 2}

	got := artifacts.OccurrenceFor(link, chunks)

	if got.SourceSection != "" {
		t.Errorf("SourceSection = %q, want empty (link before any heading)", got.SourceSection)
	}
	if got.SourceLine != 2 {
		t.Errorf("SourceLine = %d, want 2 (never an error)", got.SourceLine)
	}
}

func TestOccurrenceFor_TwoLinksInSameChunkKeepDistinctLines(t *testing.T) {
	chunks := []artifacts.Chunk{
		{Heading: "Related Specs", StartLine: 10, EndLine: 20},
	}
	link1 := artifacts.WikiLink{Target: "SPEC-001", Line: 12}
	link2 := artifacts.WikiLink{Target: "SPEC-002", Line: 18}

	occ1 := artifacts.OccurrenceFor(link1, chunks)
	occ2 := artifacts.OccurrenceFor(link2, chunks)

	if occ1.SourceSection != "Related Specs" || occ2.SourceSection != "Related Specs" {
		t.Fatalf("both occurrences should resolve to the same chunk's heading, got %q and %q", occ1.SourceSection, occ2.SourceSection)
	}
	if occ1.SourceLine != 12 || occ2.SourceLine != 18 {
		t.Errorf("SourceLine = %d/%d, want 12/18 (each occurrence keeps its own distinct line)", occ1.SourceLine, occ2.SourceLine)
	}
}

func TestChunksWithOffset_ShiftsStartAndEndLine(t *testing.T) {
	body := []byte("## Heading\n\nSome body text.\n")

	chunks := artifacts.ChunksWithOffset("spec.md", body, 5)

	if len(chunks) != 1 {
		t.Fatalf("ChunksWithOffset() returned %d chunks, want 1", len(chunks))
	}
	unshifted := artifacts.Chunks("spec.md", artifacts.ParseDocument(body))
	if chunks[0].StartLine != unshifted[0].StartLine+5 || chunks[0].EndLine != unshifted[0].EndLine+5 {
		t.Errorf("ChunksWithOffset() = {StartLine: %d, EndLine: %d}, want shifted by 5 from {%d, %d}",
			chunks[0].StartLine, chunks[0].EndLine, unshifted[0].StartLine, unshifted[0].EndLine)
	}
}
