# Quickstart: Document Model and Chunking

Builds directly on 011-wikilink-foundation's example: SPEC-014's body.

## 1. Structure an artifact's body

```markdown
<!-- SPEC-014/spec.md's body -->

## Intent

Rotate refresh tokens on every use.

## Requirements

### R1

Refresh tokens MUST be single-use.

### R2

A reused refresh token MUST revoke the session.

## Acceptance Scenarios

...
```

```go
body, _ := artifacts.ReadBody("ai/.../SPEC-014/spec.md")
doc := artifacts.ParseDocument(body)
// doc.Sections == []Section{
//   {Heading: "Intent",              Level: 2, Body: "\nRotate refresh tokens on every use.\n", ...},
//   {Heading: "Requirements",        Level: 2, Body: "",  ...}, // no content of its own before R1
//   {Heading: "R1",                  Level: 3, Body: "\nRefresh tokens MUST be single-use.\n", ...},
//   {Heading: "R2",                  Level: 3, Body: "\nA reused refresh token MUST revoke the session.\n", ...},
//   {Heading: "Acceptance Scenarios", Level: 2, Body: "...", ...},
// }
```

Note "Requirements" has an empty `Body` of its own — its content lives
entirely in its two children, R1 and R2, each its own Section (research
.md #2's flat, non-duplicating partition).

## 2. Break it into traceable pieces

```go
chunks := artifacts.Chunks("ai/.../SPEC-014/spec.md", doc)
// len(chunks) == 4 — "Requirements" produced no Chunk (FR-007):
// its own Body was empty.
// chunks[0] == Chunk{
//   Path: "ai/.../SPEC-014/spec.md", Heading: "Intent", Level: 2,
//   Content: "\nRotate refresh tokens on every use.\n",
//   StartLine: 3, EndLine: 5,
// }
```

Running this twice on the same, unchanged file produces byte-for-byte
identical output (FR-008, SC-002).

## 3. Estimate a chunk's token cost

```go
n := artifacts.EstimateTokens(chunks[0].Content)
// A deterministic approximation (~ rune count / 4) — not a real
// tokenizer, but consistent every time (FR-009).

var e artifacts.Estimator = artifacts.DefaultEstimator{}
e.Estimate(chunks[0].Content) == n // always true
```

## 4. A heading-free or empty body still works

```go
artifacts.ParseDocument([]byte("Just prose, no headings at all.\n"))
// One Section: {Heading: "", Level: 0, Body: "Just prose, no headings at all.\n", ...}

artifacts.ParseDocument([]byte(""))
// Document{Sections: nil} — zero Sections, not an error (FR-011)
```

## 5. Fenced example syntax is never mistaken for a real heading

```markdown
Some prose.

```markdown
## Not a real heading — just documentation
```

## A Real Heading
```

```go
doc := artifacts.ParseDocument(body)
// Exactly one real Section ("A Real Heading") plus the preamble —
// the fenced "## Not a real heading" line is never a boundary.
```

## Validation

Validated by: `internal/artifacts`'s new `ParseDocument`/`Chunks`/
`EstimateTokens` tests (headings, nested headings, empty sections,
fenced-block exclusion, no-heading body, empty body, duplicate-content
exclusion, determinism, wikilink-content preservation); `fencedLines`'s
extraction proven behavior-identical via 011's full existing
`ExtractWikiLinks` suite re-run unmodified as the explicit non-
regression gate.
