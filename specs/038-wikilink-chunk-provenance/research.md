# Phase 0 Research: Wikilinks with Chunk-Level Provenance

No `NEEDS CLARIFICATION` markers remained in the Technical Context —
every decision below was resolvable from this project's own existing
code (verified by reading it directly, not assumed) and established
precedent (033/035/036's contract-versioning discipline, 037's
evidence-gated promotion protocol).

## 1. How is a wikilink's enclosing section determined?

**Decision**: Reuse `artifacts.Chunks`'s existing section segmentation.
`artifacts.WikiLink.Line` (already computed by `ExtractWikiLinks`) is
correlated against the same `[]Chunk` a source artifact is already
split into everywhere else in this codebase (`StartLine <= Line <=
EndLine`); the matching chunk's own `Heading` is the wikilink's
section. No new parser, no new line-tracking mechanism.

**Rationale**: `Chunks` is already the single source of truth this
project uses to decide "what section is this line in" — `internal/
context/collector.go`'s own `chunkArtifact` and `internal/context/
index/sync.go`'s own `indexOneArtifact` both already call it. Building
a second, independent notion of "section" for wikilinks specifically
would violate Constitution Principle VI (no duplicated segmentation
logic) and risk the two disagreeing on a document's own real section
boundaries.

**Alternatives considered**: A dedicated regex/heading-scan specific to
wikilink provenance — rejected: duplicates `Chunks`'s own logic for no
benefit, and would need to be kept in sync with it by hand forever.

## 2. Where does the source content's own fingerprint (for staleness) come from, and is new staleness-detection logic needed?

**Decision**: No new staleness-detection logic. The existing
incremental-sync mechanism (015-incremental-synchronization) already
deletes and recomputes a changed artifact's own `links` rows whenever
its fingerprint changes (`internal/context/index/sync.go`'s
`deleteDocument`/`indexOneArtifact` path, already covering the `links`
table per the code's own comment: "clean up a stale artifact's own
links rows"). Spec FR-011 ("refreshed whenever content changes") is
already satisfied by this existing machinery; this feature only adds
new columns that get populated the same way the existing ones already
are, on every reindex.

**Rationale**: Reinventing staleness tracking on top of an index that
already fully reindexes a changed artifact's own links on every sync
would be redundant Principle IV complexity with no observable benefit
— the new provenance columns are recomputed at exactly the same time
and frequency the existing `relation`/`target` columns already are.

**Alternatives considered**: Storing a separate fingerprint per
occurrence and comparing it against the source's current fingerprint
at query time to flag "possibly stale" — rejected: the existing
sync-on-change model already guarantees the stored data reflects the
source's current state as of the last sync; a request-time freshness
check would only matter between an artifact changing and the next
sync running, which 014/015's own existing "Index Readiness" step
(`store.Sync` called transparently before every `context` request,
`internal/cli/internalcmd/context.go`) already closes.

## 3. What shape does the `context` command's new provenance field take, and does it require a contract version bump?

**Decision**: A new opt-in `--provenance` boolean flag (mirroring
036's `--diagnostic-scores`), omitted by default. When passed, each
item in the response gains a `provenance` array — one entry per
`ReferenceOccurrence` that justified its inclusion — each with
`source_path`, `source_section`, and `source_line`. `contextSchemaVersion`
bumps 3 → 4 to reflect the new field's availability, following the
exact precedent 036 set when it added `--diagnostic-scores` and bumped
2 → 3.

**Rationale**: Matches this project's own established pattern exactly
— an additive, opt-in field that never changes the default response
shape, still versioned because the contract's own vocabulary grew
(033/035/036's own reconciliation of "additive fields are still a
contract change worth a version marker, even when default output is
byte-identical").

**Alternatives considered**: Folding provenance into the existing
`reasons` field (currently `[]string`) by changing its shape to a list
of objects — rejected: this would be a breaking change to every
existing consumer of `reasons`, for no benefit over an additive
opt-in field; the project's own established pattern (036) is
"add a new opt-in field," not "reshape an existing default one."

## 4. Do `references`/`backlinks` commands need their own contract version marker for their new fields?

**Decision**: No. Their new occurrence fields (`source_section`,
`source_line`) are added directly to each existing entry, unconditionally
— not behind a flag — since these are already low-level introspection
commands whose entire purpose is exposing exactly this kind of detail,
and the change is purely additive (new keys; `relation`/`target` keys
and their meaning are unchanged). Neither command currently has a
`schema_version` field at all.

**Rationale**: Introducing new versioning machinery for two commands
that have never needed it before, for a purely additive change, would
be exactly the kind of premature infrastructure Principle IV warns
against. If either command's shape ever needs a genuinely breaking
change, that is the point to introduce `schema_version` for them —
not preemptively here.

**Alternatives considered**: Adding `schema_version` to `references`/
`backlinks` now, for consistency with `context` — rejected as
premature; `context`'s versioning exists because its shape has already
changed multiple times (033, 035, 036) and needed a way to say so
explicitly. A first-ever additive change to a previously-unversioned
command does not yet demonstrate that need.

## 5. How does the Requirements/active-task-section preference (spec User Story 2) get implemented without becoming a silent default change?

**Decision**: A new score-component computation, exposed only under
an explicit, non-default request parameter (planning-level name TBD in
data-model.md/contracts, e.g. a `--prefer-section` flag or intent-scoped
parameter) — when absent, ordering is byte-identical to today's. Actually
adopting it as default ordering requires a separate, future Spec that
records a comparison through 037-eval-quality-efficiency's harness and
bumps `ranking_version`, exactly the path 036 already established for
its own BM25 change.

**Rationale**: Directly satisfies spec FR-006 ("MUST NOT be applied as
default retrieval behavior unless... evaluated... through the
project's own evidence-based evaluation process") and reuses the
promotion mechanism 036 already built (`ranking_version`, the
evaluation-protocol gate) rather than inventing a second one.

**Alternatives considered**: Shipping the preference as default
immediately, since it is intuitively reasonable — explicitly rejected
by the spec itself (FR-006) and by this project's own Constitution
Principle IV precedent (dogfooding/evaluation-first tuning rule
already established in 019 and continued by 036/037).

## 6. Does widening the `links` table require touching `chunks`/`documents` or `chunks_fts`?

**Decision**: No. Only the `links` table's own schema changes (new
columns); `documents`, `chunks`, and `chunks_fts` are untouched. The
schema's overall `schemaVersion` still bumps (1 → 2) because it is one
single version tag covering the whole schema, per the existing design
— not because those other tables changed.

**Rationale**: Confirmed by reading `internal/context/index/schema.go`
and `sync.go` directly: `links` rows are populated independently of
`chunks`/`chunks_fts` rows, in `indexLinks`, a separate call from
`insertChunk`. No cross-table dependency exists that would require
touching the other two.
