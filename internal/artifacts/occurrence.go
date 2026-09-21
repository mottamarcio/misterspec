package artifacts

// ReferenceOccurrence is one specific, identifiable place a WikiLink
// was written (038-wikilink-chunk-provenance data-model.md
// "ReferenceOccurrence") — which section of the source artifact it
// falls under, and its own file-absolute line.
type ReferenceOccurrence struct {
	// SourceSection is the enclosing Chunk's own Heading — empty when
	// the link falls before any heading (research.md #1; not an error
	// condition, spec Edge Cases).
	SourceSection string
	// SourceLine is link's own Line, unchanged — always populated. Its
	// coordinate system (body-relative vs. file-absolute) is whatever
	// link.Line and chunks' own StartLine/EndLine already agree on —
	// OccurrenceFor performs no offset adjustment of its own. Callers
	// reporting this value externally MUST use file-absolute line
	// numbers, matching every other location this codebase already
	// reports (033-context-pack-output-contract research.md Decision
	// 1) — see internal/operations/references.go's own use for the
	// offset-adjustment pattern.
	SourceLine int
}

// OccurrenceFor reports link's own ReferenceOccurrence within chunks —
// the already-computed section segmentation of the artifact link was
// found in (research.md #1: reuse Chunks, never a second parser). The
// enclosing chunk is the one whose [StartLine, EndLine] contains
// link.Line — both MUST already be in the same coordinate system
// (caller's responsibility; see SourceLine's own doc). Returns
// ReferenceOccurrence{SourceLine: link.Line} with an empty
// SourceSection, never an error, when no chunk contains it.
func OccurrenceFor(link WikiLink, chunks []Chunk) ReferenceOccurrence {
	for _, c := range chunks {
		if link.Line >= c.StartLine && link.Line <= c.EndLine {
			return ReferenceOccurrence{SourceSection: c.Heading, SourceLine: link.Line}
		}
	}
	return ReferenceOccurrence{SourceLine: link.Line}
}

// ChunksWithOffset returns Chunks(relPath, ParseDocument(body)) with
// every StartLine/EndLine shifted by offset — the file-absolute
// adjustment internal/context/collector.go's chunkArtifact and
// internal/context/index/sync.go's indexOneArtifact each already apply
// by hand (offset = bodyStartLine-1 from ReadBodyWithOffset).
// operations.References/Backlinks need the identical adjustment to
// correlate a WikiLink's own file-absolute line against Chunks'
// section boundaries (038-wikilink-chunk-provenance research.md #1) —
// factored here rather than hand-rolled a third time (Constitution
// Principle VI).
func ChunksWithOffset(relPath string, body []byte, offset int) []Chunk {
	doc := ParseDocument(body)
	chunks := Chunks(relPath, doc)
	for i := range chunks {
		chunks[i].StartLine += offset
		chunks[i].EndLine += offset
	}
	return chunks
}
