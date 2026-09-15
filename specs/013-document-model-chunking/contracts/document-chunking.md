# Phase 3 Contracts: Document Model and Chunking

**Reconciled against the actual implementation (T013)** — zero drift.
Every signature below (`Section`, `Document`, `ParseDocument`, `Chunk`,
`Chunks`, `Estimator`, `DefaultEstimator`, `EstimateTokens`) matches
what shipped, verbatim, including the `ceil(rune count / 4)` estimation
method. The only cosmetic difference is `Chunk`'s actual field
declaration order (`Path, Heading, Level, StartLine, EndLine, Content`)
versus this document's own illustrative order — Go struct field order
carries no API meaning here, so this is not treated as drift. The
unexported `fencedLines` helper (shared with `ExtractWikiLinks`,
research.md #3) and `document.go`'s internal `flush` closure are
implementation detail, not part of this feature's exported surface.

No HTTP/CLI surface is added by this feature (research.md) — its
contract is the new Go API in `internal/artifacts`.

## `internal/artifacts` (extended)

```go
package artifacts

// Section is one heading-bounded region of a Document, in document
// order (docs/context-engine-implementation.md §8).
type Section struct {
    Heading   string // without the leading "#"s; "" for the untitled preamble section
    Level     int    // 1-6 for a real heading; 0 for the untitled preamble section
    Body      string // raw content between this heading and the next (any level) — never trimmed
    StartLine int    // 1-indexed
    EndLine   int    // 1-indexed; == StartLine when Body is empty
}

// Document is body's structured representation — an ordered list of
// Sections. Distinct from Metadata (frontmatter); see ParseMetadata.
type Document struct {
    Sections []Section
}

// ParseDocument splits body into Sections using ATX headings as
// boundaries (FR-001, FR-002). Never treats heading-like text inside a
// fenced code block as a boundary (FR-003). A body with no headings at
// all still produces exactly one Section (FR-004); a completely empty
// body produces zero Sections (FR-011). Pure function of body — no I/O,
// no path (FR-008, FR-010).
func ParseDocument(body []byte) Document

// Chunk is one retrieval-sized piece derived from exactly one non-empty
// Section, carrying enough provenance to explain where it came from
// (FR-006).
type Chunk struct {
    Path      string
    Heading   string
    Level     int
    Content   string
    StartLine int
    EndLine   int
}

// Chunks derives one Chunk per non-empty-Body Section of doc, in
// document order, attributing each to path exactly as given (FR-005,
// FR-006, FR-007). Pure function of its inputs — no I/O (FR-010).
func Chunks(path string, doc Document) []Chunk

// Estimator estimates the token cost of a piece of text, kept as an
// interface so a real provider tokenizer can be substituted later
// without changing any caller (docs/context-engine-implementation.md §9).
type Estimator interface {
    Estimate(text string) int
}

// DefaultEstimator is this project's own deterministic approximation —
// see EstimateTokens.
type DefaultEstimator struct{}

func (DefaultEstimator) Estimate(text string) int

// EstimateTokens returns a deterministic, approximate token count for
// text (FR-009) — ceil(rune count / 4); EstimateTokens("") == 0. Not a
// real provider tokenizer (docs/context-engine-implementation.md §9).
func EstimateTokens(text string) int
```

**Guarantees**:
- `ParseDocument`/`Chunks`/`EstimateTokens` are pure functions of their
  inputs — no filesystem access, no mutation, deterministic output for
  the same input every time (FR-008, FR-010, matching every prior
  feature's own determinism discipline).
- Every line of a non-empty `body` belongs to exactly one `Section`'s
  `Body` — never zero, never more than one (data-model.md).
- A `Section`/`Chunk` never spans a heading boundary, and a parent
  heading's own content never duplicates a child heading's own content
  (research.md #2).
- A `Chunk`'s `Content` preserves any wikilink syntax exactly as
  authored — this feature never strips, alters, or resolves a link
  (FR-012; that remains 011's and 012's own separate concern).
- `ParseMetadata`'s own behavior (every existing caller, every existing
  test) is unchanged — this feature adds no new call site to it at all.
- `ExtractWikiLinks`'s own observable behavior is unchanged after the
  `fencedLines` extraction (research.md #3) — 011's full existing test
  suite is the explicit regression gate.

## Cross-cutting: no change to any existing contract

`internal/validation` and `internal/operations`'s exported surfaces
(Finding codes, `References`/`Backlinks`) are completely untouched —
this feature adds no new call site into either package, and no CLI
command (research.md's own "Summary of Go footprint").
