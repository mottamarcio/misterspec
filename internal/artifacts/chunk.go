package artifacts

// Chunk is one retrieval-sized piece of an artifact, derived from
// exactly one non-empty Section, carrying enough provenance to explain
// where it came from (FR-006). No ArtifactID field (013-document-
// model-chunking/research.md #5 — Path is this feature's only identity)
// and no Tokens field (research.md #6 — token estimation stays a fully
// independent, on-demand capability; see EstimateTokens).
type Chunk struct {
	// Path is the artifact's own path, exactly as the caller supplied
	// it to Chunks — never resolved, validated, or re-read.
	Path string
	// Heading, Level, StartLine, EndLine are copied from the
	// originating Section.
	Heading   string
	Level     int
	StartLine int
	EndLine   int
	// Content is copied from the originating Section's Body, verbatim —
	// including any wikilink syntax written inside it (FR-012); this
	// feature never strips, alters, or resolves a link.
	Content string
}

// Chunks derives one Chunk per non-empty-Body Section of doc, in
// document order, attributing each to path exactly as given (FR-005,
// FR-006, FR-007). A Section whose Body is empty produces no Chunk.
// Pure function of its inputs — no I/O of its own (FR-008, FR-010).
func Chunks(path string, doc Document) []Chunk {
	chunks := make([]Chunk, 0, len(doc.Sections))
	for _, s := range doc.Sections {
		if s.Body == "" {
			continue
		}
		chunks = append(chunks, Chunk{
			Path:      path,
			Heading:   s.Heading,
			Level:     s.Level,
			StartLine: s.StartLine,
			EndLine:   s.EndLine,
			Content:   s.Body,
		})
	}
	return chunks
}
