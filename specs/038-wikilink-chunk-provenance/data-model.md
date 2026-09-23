# Phase 1 Data Model: Wikilinks with Chunk-Level Provenance

Entities correspond to spec.md's Key Entities section. Field names
below are illustrative (the authoritative shape is `contracts/`); this
document defines relationships, ownership, and validation rules.

## ReferenceOccurrence

One specific, identifiable place a wikilink was written — spec's
"Reference Occurrence" (Key Entities). Not a new artifact or entity ID
(Constitution Principle II N/A) — a computed, read-side value.

| Field | Type | Notes |
|---|---|---|
| `SourcePath` | string | The referencing artifact's own path (relative to project root) — the artifact the wikilink was written *in*. |
| `SourceID` | string | The referencing artifact's own entity ID (e.g. `SPEC-014`), when it has one. |
| `SourceSection` | string | The heading of the `artifacts.Chunk` whose `[StartLine, EndLine]` contains the wikilink's own line (research.md #1). Empty when the wikilink falls outside any heading (e.g. before the first `##`). |
| `SourceLine` | int | The wikilink's own line, file-absolute (matching every other file-absolute line number this codebase already reports — 033 research.md Decision 1 precedent). |
| `Relation` | string | `"wikilink"` — kept for symmetry with `ReferenceEntry.Relation`/`BacklinkEntry.Relation`; a formal relationship (`parent`/`depends_on`/`supersedes`) has no `ReferenceOccurrence` of its own, since it is not written as a wikilink (spec FR-009 — a reference occurrence is never conflated with a formal dependency). |
| `Target` | ids.EntityID | The artifact the wikilink resolves to. |

**Validation rules**: `SourcePath`/`SourceLine` always present for a
wikilink-derived occurrence (spec FR-001); `SourceSection` MAY be
empty (a wikilink before any heading is still a valid occurrence, just
with no enclosing section — spec Edge Cases: "not an error
condition").

## ReferenceEntry / BacklinkEntry (extended)

`operations.ReferenceEntry` (012-references-backlinks) and `operations.
BacklinkEntry` each gain the `ReferenceOccurrence` fields directly.
`SourcePath` is populated for every entry, formal or semantic — the
referencing artifact's own path is cheaply known regardless of relation
kind (both directions already have it in scope where these entries are
built). `SourceSection`/`SourceLine` are populated only for a semantic
(`Relation == "wikilink"`) entry — a formal entry (`parent`/
`depends_on`/`supersedes`) carries those two empty/zero, since it has
no line-level wikilink origin to report; its formal nature is
unaffected either way (spec FR-009).

| Field (new) | Type | Notes |
|---|---|---|
| `SourcePath` | string | The referencing artifact's own path. For `ReferenceEntry` (outgoing from a known target) this is trivially the queried target's own path; included for symmetry and so callers never need direction-specific logic. For `BacklinkEntry` (incoming, from some other Source artifact) this is the only way the caller learns that Source's own path without a separate lookup — `operations.Backlinks` already has it in scope while scanning (`filePath`, `internal/operations/backlinks.go`), so it costs nothing extra to include here. |
| `SourceSection` | string | See `ReferenceOccurrence.SourceSection`. Empty for a formal entry. |
| `SourceLine` | int | See `ReferenceOccurrence.SourceLine`. Zero for a formal entry. |

`ReferenceEntry.Relation`/`Target` and `BacklinkEntry.Relation`/`Source`
are unchanged (research.md #4 — additive only, no existing field's
meaning changes).

## Reason (extended, `internal/context`)

`contextengine.Reason` (015-context-collector) — currently
`{Tier, Relation}` — gains the same occurrence fields for a
`Relation == "wikilink"` reason, populated from the
`ReferenceOccurrence` that produced the `Candidate` it is attached to.

| Field (new) | Type | Notes |
|---|---|---|
| `SourcePath` | string | The referencing artifact's own path — populated for any reference-derived `Relation` (`"parent"`, `"depends_on"`, `"supersedes"`, `"wikilink"`, `"backlink"`), mirroring `ReferenceEntry`/`BacklinkEntry`'s own "always populated" rule. Empty for `"constitution"`/`"target"`/`"text_match"`, which have no referencing-artifact concept at all. |
| `SourceSection` | string | |
| `SourceLine` | int | |

**Validation rules**: `SourceSection`/`SourceLine` are populated only
when the underlying occurrence was itself a wikilink — i.e. `Relation
== "wikilink"`, or `Relation == "backlink"` whose underlying
`BacklinkEntry` was itself semantic. A formal relation (`"parent"`,
`"depends_on"`, `"supersedes"`, or a `"backlink"` reason built from a
formal `BacklinkEntry`) carries these two fields empty/zero — never
populated, never implying a wikilink occurrence exists where none
does — while still carrying a populated `SourcePath` (spec FR-009 is
about not fabricating a *dependency*, not about withholding which
artifact established a real, already-formal relationship). In
practice, `internal/context/collector.go`'s own `reasonOf` callbacks
copy all three fields straight from the already-correctly-shaped
`ReferenceEntry`/`BacklinkEntry` they are built from — no
relation-based branching needed at the `contextengine` layer.

## Preference Score (new, off by default)

An additional, explicit scoring input (spec User Story 2) computed
from a `ReferenceOccurrence.SourceSection` — whether it matches a
Spec's own "Requirements" content, or the section covering the
currently active task. Not a new persisted entity; a request-scoped
computation.

| Concept | Notes |
|---|---|
| Activation | `contextengine.Request.PreferSection bool` — new field, zero value `false` (off by default, matching `Request`'s existing `QueryMode`-style "zero value means today's behavior" convention). Absent/false → byte-identical ordering to today's (spec FR-006). |
| Computation | `ScoreComponents` (036-text-search-ranking) gains a new `SectionPreference int` field. Zero unless `Request.PreferSection` is true. When true, a candidate with a wikilink-origin `Reason` whose `SourceSection` case-insensitively matches a fixed set (`"Requirements"`, `"Functional Requirements"`) contributes a fixed positive bonus, included in `Total` — additive alongside `RelationWeight`/`IntentBonus`/`TextRelevance`, never replacing them. |
| Promotion path | Adopting this as default ordering (making `PreferSection`'s effect apply without the caller passing it) requires a future Spec recording a comparison through 037-eval-quality-efficiency's harness and bumping `ranking_version` (research.md #5) — out of scope for this feature's own default behavior. |

## `links` table (widened, `internal/context/index`)

| Column (new) | Type | Notes |
|---|---|---|
| `source_section` | TEXT, nullable | `ReferenceOccurrence.SourceSection`; NULL for a formal relationship or an occurrence with no enclosing heading. |
| `source_line` | INTEGER, nullable | `ReferenceOccurrence.SourceLine`; NULL for a formal relationship. |

`schemaVersion` bumps 1 → 2 (research.md #6); the index's own existing
version-mismatch rebuild path (`internal/context/index/schema.go`)
recreates the table with the new columns automatically — no manual
migration, no data carried forward (Constitution Principle III, spec
FR-010).

**Validation rules**: unchanged from the existing `links` table's own
constraints (`source_artifact_id`/`target_artifact_id`/`relation` all
`NOT NULL`); the two new columns are the only nullable ones, reflecting
that a formal relationship legitimately has no line-level origin.

## Relationships

```text
artifacts.WikiLink (existing, has Line)
        │  correlated against
        ▼
artifacts.Chunks (existing segmentation)
        │  produces
        ▼
ReferenceOccurrence (new)
        │  attached to
        ▼
operations.ReferenceEntry / BacklinkEntry (extended)
        │  consumed by
        ▼
contextengine.Candidate.Reason (extended)  ──▶  context command's
        │                                        opt-in "provenance"
        │  also persisted as                     field (contextSchemaVersion 4)
        ▼
index "links" table (widened, schemaVersion 2)
```

## Validation Rules Summary

- A `ReferenceOccurrence`/extended `ReferenceEntry`/`BacklinkEntry`/
  `Reason` MUST NOT be presented, stored, or interpreted as a formal
  execution dependency (spec FR-009) — occurrence fields are additive
  metadata on a semantic wikilink relation only.
- Every occurrence field populated for a wikilink-derived entry MUST
  reflect the source artifact's own state as of the last successful
  index sync (research.md #2) — never manually reconciled or
  request-time-recomputed beyond that.
- The preference-score capability (spec FR-005) MUST default to
  inactive; its presence or absence MUST NOT change response ordering
  unless explicitly requested (spec FR-006).
