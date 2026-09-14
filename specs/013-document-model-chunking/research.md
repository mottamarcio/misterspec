# Phase 0 Research: Document Model and Chunking

All unknowns below were resolved by grounding
`docs/context-engine-implementation.md` §8-9, §29.4 against the actual
codebase (`internal/artifacts`, especially 011's `wikilink.go`) rather
than against the doc's own illustrative snippets in isolation, per that
document's own §32.1 instruction.

## 1. Package placement: no new package

**Decision**: `Document`, `Section`, `Chunk`, and the token-estimation
helpers all live in `internal/artifacts` — three new files
(`document.go`, `chunk.go`, `tokens.go`) alongside 011's own
`wikilink.go`/`markdown.go`. No new package.

**Rationale**: `docs/context-engine-implementation.md` §31's own
illustrative layout already places `document.go` under
`internal/artifacts/`, not under a new `internal/context/`. The
`internal/context` package that layout also shows exists to house
Phase 4-8's own retrieval engine, index, and ranking — none of which
this feature touches. Constitution Principle IV (YAGNI) forbids
introducing a new package layer speculatively; this feature's own
content (structuring and chunking an artifact's already-read body) is a
direct, natural extension of what `internal/artifacts` already does
(`ReadBody`, `ExtractWikiLinks`), not a distinct architectural layer.

**Alternatives considered**: Creating `internal/context` now, pre-
populated with just this feature's own types (rejected — an empty-
except-for-this-feature package is exactly the kind of premature
structure Principle IV warns against; `internal/context` will have a
real, demonstrated reason to exist starting with Phase 4).

## 2. Flat, non-duplicating section model

**Decision**: `Section` is a flat, ordered list (not a nested tree).
Content between one heading (at any level, 1-6) and the next heading
(at any level) belongs exclusively to the first heading's own `Body` —
never repeated in a parent's `Body` once a child heading exists.

**Rationale**: `docs/context-engine-implementation.md` §8's own
illustrative `Document` struct already declares `Sections []Section`
(flat), with `Level int` present purely as descriptive metadata, not as
a nesting mechanism. A flat, boundary-based partition is also the
simplest possible correct answer to spec.md's own Edge Case ("a parent
heading's own piece must not duplicate content already captured by its
children's own pieces") — by construction, every line of body content
is assigned to exactly one Section, with no possibility of double-
counting (Constitution Principle IV, KISS).

**Alternatives considered**: A recursive tree (`Section.Children
[]Section`, with a parent's own `Body` covering everything under it
including descendants) — rejected: doesn't match the doc's own
illustrative shape, and would require an explicit, separate
"leaf-only" extraction step before chunking anyway (since a chunk must
never span both a parent and child's content per the no-duplication
edge case) — strictly more code for no behavioral gain over the flat
model.

## 3. Sharing fence-tracking with 011's wikilink extraction

**Decision**: Factor 011's own `ExtractWikiLinks` fence-tracking loop
(the `inFence`/`fenceMarker` state machine, `fenceMarkerOf`) into a new
small shared helper, `fencedLines(lines []string) []bool`, reporting for
each line of a split body whether it is itself a fence delimiter or
falls inside a fenced block. `ExtractWikiLinks` is refactored to call
it; the new section-parsing logic (Decision 2) calls the same function.

**Rationale**: Constitution Principle VI (DRY). Both wikilink extraction
and heading detection need the exact same guarantee — never treat
fenced content as real syntax (spec.md's own Edge Case, mirroring 011's
own). This is the direct continuation of 012's own precedent (extracting
`ids.ResolveTarget` out of `classifyWikilink` for the same reason): a
second, independent need for logic that already exists inline in one
place is exactly when DRY calls for a shared, tested extraction rather
than a second copy. `ExtractWikiLinks`'s own full existing test suite
(011) is this refactor's regression gate, proving the change is
behavior-identical.

**Alternatives considered**: A third, independent fence-tracking loop
inside the new section-parsing code (rejected — the exact duplication
Principle VI exists to prevent, immediately after having just fixed the
same class of duplication in 012).

## 4. Heading syntax: ATX only, no Setext

**Decision**: Only ATX headings (`#` through `######`, followed by at
least one space) are recognized as section boundaries. Setext headings
(`Title\n=====`) are not supported.

**Rationale**: This project's own Markdown content (every artifact
template, every prior spec/plan/tasks file, `docs/architecture-
specification.md` itself) exclusively uses ATX headings — the same
"cover this project's own actual content, not full CommonMark" scoping
011 already applied to code spans/fences (research.md). A line starting
with `#` but no following space (e.g. a prose `#123` reference) is
correctly not treated as a heading, matching CommonMark's own ATX rule
and avoiding false positives.

**Alternatives considered**: Full CommonMark heading detection via a
Markdown-parsing library (rejected for the same reason 011 rejected
`goldmark` — disproportionate to this project's own actual content,
Constitution Principle IV).

## 5. Chunk provenance: Path, not an entity ID

**Decision**: `Chunk`'s provenance is `Path` (exactly as the caller
supplies it) plus `Heading`/`Level`/`StartLine`/`EndLine` — no
`ArtifactID` field. `ParseDocument`/`Chunks` never look up or require an
entity ID at all.

**Rationale**: spec.md's own Assumptions already settled this: "any
managed artifact" means any file this project already reads a body
from, not a re-derivation of 012's own "which entity types are
addressable" scope question — a different concern (identity/resolution)
than structuring/chunking needs. Requiring an `ids.EntityID` here would
force a new dependency on `internal/ids` into a package that currently
has none, and would re-introduce exactly the "which five types" scoping
question 012 already answered for a different purpose. Whichever future
caller actually walks the project (Phase 4's indexer) already knows
each file's path and, where one exists, its ID — attaching an ID to a
chunk's persisted row is that caller's own job, not this feature's.

**Alternatives considered**: Accepting an optional `ids.EntityID`
parameter alongside `path` (rejected — no current caller has one to
give, and an always-empty optional field is speculative surface,
Principle IV).

## 6. Token estimation kept independent of Chunk construction

**Decision**: `Chunk` has no `Tokens` field. Token estimation
(`EstimateTokens`/`Estimator`) is a fully separate, general-purpose
capability operating on any string — including, but not limited to, a
chunk's own `Content` — computed on demand by whichever caller needs the
number, not baked into `Chunk` at construction time.

**Rationale**: Keeps User Story 2 (chunking) and User Story 3 (token
estimation) genuinely independent, exactly as spec.md's own "Independent
Test" for each describes — chunking never needs the estimator to
produce a correct, complete `Chunk`. `docs/context-engine-
implementation.md` §10's own SQLite schema (`CREATE TABLE chunks (...
tokens ...)`, Phase 4) is where a persisted token count actually gets
used; populating that column by calling `EstimateTokens(chunk.Content)`
at insert time is Phase 4's own job. Deferring the coupling to the
phase that actually consumes it avoids this feature computing and
discarding a number nothing here yet uses (Principle IV).

**Alternatives considered**: Including `Tokens int` on `Chunk` per the
source document's own illustrative struct exactly (rejected after
grounding it against this feature's own user-story boundaries above —
the doc itself says "exact field names may differ").

## 7. Token estimation algorithm

**Decision**: `EstimateTokens(text string) int` returns
`ceil(len([]rune(text)) / 4)` — a widely used, simple character-count
approximation for English prose (roughly 4 characters per token for
common tokenizers) — with an empty string returning exactly `0`. Also
exposed via a small `Estimator` interface (`Estimate(text string) int`)
implemented by `DefaultEstimator`, per `docs/context-engine-
implementation.md` §9's own suggested shape, so a real provider
tokenizer can be substituted later without changing any caller.

**Rationale**: `docs/context-engine-implementation.md` §9 explicitly
allows a documented, consistent approximation for the MVP ("does not
require an exact provider tokenizer... deterministic approximation is
acceptable"). Rune count (not byte count) is used so multi-byte UTF-8
characters are not over-counted. The interface wrapper directly
satisfies the doc's own explicit "keep the abstraction replaceable"
instruction at negligible cost — one interface, one implementing type.

**Alternatives considered**: A word-count-based heuristic (rejected —
character count is simpler, needs no tokenization/splitting logic of its
own, and is at least as standard an approximation). Integrating a real
tokenizer library (rejected — explicitly out of scope for the MVP per
the source document itself, and a new external dependency this feature
does not need, Principle IV).

## 8. No content trimming; raw line ranges

**Decision**: A Section's `Body` (and a Chunk's `Content`) is exactly
the raw lines between its boundaries — no leading/trailing blank-line
trimming. `StartLine` is the section's own heading line (or `1` for the
untitled preamble section); `EndLine` is the last line included in
`Body` (equal to `StartLine` when `Body` is empty).

**Rationale**: Simplicity and determinism (Constitution Principle IV,
KISS) — trimming would require deciding how it interacts with line-
number provenance for no behavioral requirement this feature actually
has (a later rendering step, if one ever wants trimmed presentation, can
trim `Content` itself without needing this layer to have already done
so). Avoids an entire category of "off by one blank line" bugs a
trimming step would risk introducing.

**Alternatives considered**: Trimming blank lines from `Body`/`Content`
while keeping `StartLine`/`EndLine` describing the untrimmed range
(rejected — a `Content` string whose length doesn't match its own
declared line range is a confusing, easy-to-misuse contract for zero
real benefit at this phase).

## Summary of Go footprint

- `internal/artifacts/wikilink.go` (modified): factor `fencedLines` out
  of `ExtractWikiLinks`'s existing inline fence-tracking loop —
  behavior-identical, 011's full existing suite is the regression gate.
- `internal/artifacts/document.go` (new): `Section`, `Document`,
  `ParseDocument`.
- `internal/artifacts/chunk.go` (new): `Chunk`, `Chunks`.
- `internal/artifacts/tokens.go` (new): `Estimator`, `DefaultEstimator`,
  `EstimateTokens`.
- No new package, no new external dependency, no CLI change (matching
  011's own zero-CLI-change precedent — this feature has no `internal
  …` command of its own; Phase 8's own "internal context" command is
  the eventual, later CLI surface).
