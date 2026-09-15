# Phase 1 Data Model: Document Model and Chunking

Entities extracted from `spec.md`'s Key Entities plus
`docs/context-engine-implementation.md` §8-9. This feature adds three
new files to `internal/artifacts` plus one small, DRY-motivated refactor
of an existing one — no new package, no new dependency.

## `internal/artifacts.Section` (new)

| Field | Type | Notes |
|---|---|---|
| `Heading` | string | Heading text, without the leading `#`s. Empty for the untitled preamble section (research.md #2). |
| `Level` | int | `1`-`6` for a real ATX heading; `0` for the untitled preamble section. |
| `Body` | string | Raw content strictly between this heading's own line and the next heading (at any level) — never trimmed (research.md #8). Empty when a heading is immediately followed by another heading. |
| `StartLine` | int | 1-indexed: the heading's own line, or `1` for the preamble section. |
| `EndLine` | int | 1-indexed: the last line included in `Body`; equals `StartLine` when `Body` is empty. |

## `internal/artifacts.Document` (new)

| Field | Type | Notes |
|---|---|---|
| `Sections` | `[]Section` | In document order (FR-001). Empty (not nil) for a body with no headings and no content — see `ParseDocument`'s own rules below. |

## `internal/artifacts.ParseDocument` (new)

| Symbol | Notes |
|---|---|
| `ParseDocument(body []byte) Document` | Pure function of `body` — no filesystem access, no path, matching `ExtractWikiLinks`'s own determinism discipline (FR-008). Splits on ATX headings (`#` through `######`, research.md #4), skipping heading-like text inside a fenced code block (FR-003, reusing `fencedLines`, research.md #3). A body with no headings at all still produces exactly one Section (Heading `""`, Level `0`) covering the entire body (FR-004). A completely empty body produces zero Sections (FR-011). |

**Validation rules**: Every line of `body` belongs to exactly one
Section's `Body` (flat partition, research.md #2) — never zero, never
more than one.

## `internal/artifacts.Chunk` (new)

| Field | Type | Notes |
|---|---|---|
| `Path` | string | Exactly as the caller supplied it to `Chunks` — never resolved, validated, or re-read (research.md #5). |
| `Heading` | string | Copied from the originating Section. |
| `Level` | int | Copied from the originating Section. |
| `Content` | string | Copied from the originating Section's `Body`, verbatim — including any wikilink syntax written inside it (FR-012). |
| `StartLine` | int | Copied from the originating Section. |
| `EndLine` | int | Copied from the originating Section. |

No `ArtifactID` field (research.md #5) and no `Tokens` field
(research.md #6) — both deliberate deviations from
`docs/context-engine-implementation.md` §8.2's own illustrative struct,
justified there.

## `internal/artifacts.Chunks` (new)

| Symbol | Notes |
|---|---|
| `Chunks(path string, doc Document) []Chunk` | One Chunk per non-empty-`Body` Section of `doc`, in document order (FR-005, FR-007) — a Section whose `Body` is empty produces no Chunk. Pure function of its inputs (FR-008); performs no I/O of its own (FR-010 — read-only by construction, since it never touches the filesystem at all). |

## `internal/artifacts.Estimator` / `EstimateTokens` (new)

| Symbol | Notes |
|---|---|
| `type Estimator interface { Estimate(text string) int }` | Kept replaceable per `docs/context-engine-implementation.md` §9's own explicit instruction (research.md #7). |
| `type DefaultEstimator struct{}` | Implements `Estimator` by calling `EstimateTokens`. |
| `EstimateTokens(text string) int` | `ceil(len([]rune(text)) / 4)`; `EstimateTokens("") == 0` (FR-009, research.md #7). Deterministic: same input always yields the same output. |

## `internal/artifacts.ExtractWikiLinks` (extended, behavior-unchanged)

| Symbol | Notes |
|---|---|
| `fencedLines(lines []string) []bool` (new, unexported) | Factored out of `ExtractWikiLinks`'s own existing fence-tracking loop (research.md #3) — shared with `ParseDocument`. `ExtractWikiLinks`'s own observable behavior is unchanged; its full existing test suite (011) is this refactor's regression gate. |

## State / Flow Summary

```text
artifact body (via 011's artifacts.ReadBody, or any raw []byte)
      ↓ artifacts.ParseDocument                                    [US1]
Document{Sections: [...]}  — flat, ordered, one Section per heading
      (+ one untitled preamble Section if there's content before the
       first heading, or if there are no headings at all)
      ↓ artifacts.Chunks(path, doc)                                 [US2]
[]Chunk — one per non-empty Section, each carrying Path/Heading/
  Level/Content/StartLine/EndLine

Independently, for any text (a Chunk's Content or anything else):
      ↓ artifacts.EstimateTokens(text) / DefaultEstimator.Estimate   [US3]
int — deterministic, approximate token cost
```
