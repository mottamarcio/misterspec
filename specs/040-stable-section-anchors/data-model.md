# Data Model: Referências a Seções com Âncoras Estáveis

## Section (extended — `internal/artifacts`)

Existing type (`document.go`); this feature adds one field.

| Field (new) | Type | Notes |
|---|---|---|
| `Anchor` | `string` | The explicit `{#slug}` suffix declared on this heading's own line, stripped out of `Heading`; `""` when no suffix is present (the common case). |

**Validation rules**: `Anchor`, when non-empty, is opaque author text —
`ParseDocument` itself never validates its shape or uniqueness (that is
`internal/validation`'s job, spec FR-005). Two different Sections *may*
carry the same non-empty `Anchor` at the parse layer; that condition is
a validation Finding, not a parse error (keeps `ParseDocument` a pure,
always-succeeding function, matching its existing contract).

## Chunk (extended — `internal/artifacts`)

Existing type (`chunk.go`); this feature adds one field, copied
straight from the originating `Section.Anchor`.

| Field (new) | Type | Notes |
|---|---|---|
| `Anchor` | `string` | Same value/emptiness rule as `Section.Anchor`. |

**Note**: `Chunks()` still skips a Section whose `Body` is empty (its
existing, unchanged contract) — an anchor on such a Section produces no
`Chunk` from `Chunks()`. `chunkArtifactAnchor` (below) therefore
resolves against `Document`/`Section`s directly, not against
`Chunks()`'s already-filtered output, so an anchor on an empty-Body
heading (spec Edge Cases) still resolves successfully to an empty-
content Candidate rather than "not found."

## WikiLink (extended — `internal/artifacts`)

Existing type (`wikilink.go`); this feature adds one field.

| Field (new) | Type | Notes |
|---|---|---|
| `Anchor` | `string` | The `#anchor` token found between `Target` and any `|Alias`; `""` when absent. `Target` itself is unchanged — never includes the `#anchor` suffix. |

**Validation rules**: `Anchor` is a raw token exactly as written — not
yet checked for existence; that stays `internal/validation`'s separate
job (§6.1 precedent: `Target` is likewise unresolved at parse time).

## ReferenceEntry / BacklinkEntry (extended — `internal/operations`)

Existing types (`references.go`, `backlinks.go`); this feature adds one
field to each, mirroring how 038 added `SourceSection`/`SourceLine`.

| Field (new) | Type | Notes |
|---|---|---|
| `TargetAnchor` | `string` | The anchor named by the underlying wikilink, if any. Populated only when `Relation == "wikilink"` and the originating `WikiLink.Anchor != ""`; `""` for a formal entry or a non-anchor semantic entry. |

## Candidate / Reason (extended — `internal/context`)

Existing types (`result.go`); this feature adds fields to each.

| Type | Field (new) | Type | Notes |
|---|---|---|---|
| `Candidate` | `HeadingPath` | `[]string` | Ordered ancestor heading titles (outermost first) of this Candidate's own Section, populated only when the Candidate came from an anchor-qualified reference (`chunkArtifactAnchor`); `nil` for every other Candidate — including a Candidate for the *same* Section reached without an anchor, since only the anchor-qualified path needs the breadcrumb to stand in for the artifact context it's deliberately not including (spec FR-006). |
| `Reason` | `TargetAnchor` | `string` | Copied from the driving `ReferenceEntry`/`BacklinkEntry.TargetAnchor` — lets a consumer see *which* anchor caused this Candidate's inclusion. |

**Validation rule**: `HeadingPath` is `nil` if and only if the Candidate
was *not* produced by `chunkArtifactAnchor`. **Correction found during
implementation**: `chunkArtifactAnchor` always sets `HeadingPath` to a
non-nil slice, `[]string{}` when the anchored Section has no ancestors
(a top-level anchored heading) — nil-ness, not length, is the only
reliable signal that a Candidate is anchor-derived, since a real,
anchor-derived `HeadingPath` can legitimately be empty. A first
implementation gated the JSON `heading_path`/`anchor` fields on
`len(HeadingPath) > 0` and silently dropped both for exactly this
top-level case; caught by a manual end-to-end walkthrough of
quickstart.md §1 (whose own example anchors a top-level heading) and
fixed by gating on non-nil-ness instead, in both
`internal/context/collector.go` (`chunkArtifactAnchor`) and
`internal/context/result.go` (`mergeAndSort`'s own HeadingPath-forwarding
check on a merged duplicate).

## Validation Codes (new — `internal/validation/findings.go`)

| Code | Constant | Raised when | Distinct from |
|---|---|---|---|
| `unknown_anchor` | `CodeUnknownAnchor` | A wikilink's `Anchor` is non-empty, its `Target` resolves to exactly one existing artifact, but that artifact declares no Section with a matching `Anchor`. | `CodeBrokenWikilink` (target artifact itself doesn't exist) — spec FR-009 requires these stay distinguishable. |
| `duplicate_anchor` | `CodeDuplicateAnchor` | Two or more Sections within the same artifact declare the same non-empty `Anchor`. Raised once per artifact scan, independent of whether anything currently references that anchor (spec Assumptions — an unreferenced anchor is not itself an error; a *duplicated* one is). | Any existing ID-duplication code (`CodeDuplicateID`, `CodeDuplicateRequirementID`) — those cover entity/requirement identity, not in-document heading anchors. |

## Index Schema (extended — `internal/context/index`)

`schemaVersion` bumps 2 → 3 (research.md #5).

| Table | Column (new) | Type | Notes |
|---|---|---|---|
| `chunks` | `anchor` | `TEXT` | Nullable; empty/NULL when the Chunk's own `Anchor` is `""`. |
| `links` | `target_anchor` | `TEXT` | Nullable; empty/NULL for a non-anchor-qualified link row. |

No new table — anchor data widens the two tables 038 already
established the pattern for.
