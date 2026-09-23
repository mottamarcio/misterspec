# Contract: Stable Section Anchors

This documents the machine-readable contracts this feature adds/changes
(Constitution Principle IX). Every surface below is additive and
reached only when an anchor is actually present (research.md #6) —
existing default output for a non-anchor wikilink/Section is
byte-identical to before this feature, except where a contract's own
`schema_version` already requires a bump.

## 1. `internal/artifacts` (extended)

```go
package artifacts

// Section gains Anchor (data-model.md "Section (extended)"). Stripped
// from the trailing "{#slug}" suffix on the heading's own line, if
// present; "" otherwise. Heading no longer includes the suffix.
type Section struct {
	Heading   string
	Anchor    string // new
	Level     int
	Body      string
	StartLine int
	EndLine   int
}

// Chunk gains Anchor, copied verbatim from the originating Section.
type Chunk struct {
	Path      string
	Heading   string
	Anchor    string // new
	Level     int
	StartLine int
	EndLine   int
	Content   string
}

// WikiLink gains Anchor — the "#anchor" token between Target and any
// "|Alias", if present; "" otherwise. Target itself never includes it.
type WikiLink struct {
	Target string
	Anchor string // new
	Alias  string
	Line   int
}
```

## 2. `internal/validation` (extended)

```go
package validation

const (
	// ... existing codes unchanged ...
	CodeUnknownAnchor   = "unknown_anchor"
	CodeDuplicateAnchor = "duplicate_anchor"
)
```

`classifyWikilink` gains one more branch, reached only when the target
resolves to exactly one artifact and `link.Anchor != ""`: read that
artifact's own Sections and confirm one has a matching `Anchor`;
`CodeUnknownAnchor` otherwise. A new `checkAnchors(root, cfg, filePath)`
(same shape as `checkWikilinks`) scans one artifact's own Sections for
a repeated non-empty `Anchor`, raising `CodeDuplicateAnchor` — run
independent of any wikilink referencing that anchor.

## 3. `internal/operations` (extended)

```go
package operations

// ReferenceEntry/BacklinkEntry each gain TargetAnchor (data-model.md
// "ReferenceEntry / BacklinkEntry (extended)"), populated only for a
// "wikilink" Relation whose originating WikiLink.Anchor != "".
type ReferenceEntry struct {
	Relation      string
	Target        ids.EntityID
	SourcePath    string
	SourceSection string
	SourceLine    int
	TargetAnchor  string // new
}

type BacklinkEntry struct {
	Relation      string
	Source        ids.EntityID
	SourcePath    string
	SourceSection string
	SourceLine    int
	TargetAnchor  string // new
}
```

## 4. `misterspec internal references <id>` / `internal backlinks <id>` (additive, no version bump)

Each entry gains one new key, always present (empty string for a
non-anchor or formal entry):

```json
{
  "relation": "wikilink",
  "target": "KNOW-003",
  "source_path": "specs/040-stable-section-anchors/spec.md",
  "source_section": "Requirements",
  "source_line": 12,
  "target_anchor": "retry-policy"
}
```

No new error conditions on these two commands themselves — an unknown
anchor is a `validate` Finding (§2), not a `references`/`backlinks`
resolution failure; these commands still report whatever `Target`
resolved to, anchor or not.

## 5. `internal/context` (extended)

```go
package context

// Candidate gains HeadingPath (data-model.md "Candidate / Reason
// (extended)"), non-nil if and only if this Candidate came from
// chunkArtifactAnchor — []string{} (non-nil, empty), never nil, when
// the anchored Section has no ancestors; gate on nil-ness, never on
// length (correction found during implementation, data-model.md).
type Candidate struct {
	Path       string
	Heading    string
	Content    string
	StartLine  int
	EndLine    int
	Reasons    []Reason
	HeadingPath []string // new
}

// Reason gains TargetAnchor, copied from the driving
// ReferenceEntry/BacklinkEntry.
type Reason struct {
	Tier          Tier
	Relation      string
	SourcePath    string
	SourceSection string
	SourceLine    int
	TargetAnchor  string // new
}

// chunkArtifactAnchor resolves exactly one Candidate for the Section
// whose Anchor == anchor within the artifact at relPath — including a
// Section with an empty Body (unlike chunkArtifact/artifacts.Chunks,
// which skip those). HeadingPath is that Section's ancestor headings'
// titles, outermost first. Returns errAnchorNotFound (wrapped) when no
// Section declares anchor — this IS reachable in normal use (see
// correction below), not just an I/O-failure case.
func chunkArtifactAnchor(root, relPath, anchor string, reason Reason) (Candidate, error)
```

`connectedCandidates` branches to `chunkArtifactAnchor` instead of
`chunkArtifact` whenever the driving entry's `Relation == "wikilink"`
and its `TargetAnchor != ""`. **Correction found by code review, after
initial implementation**: this contract originally asserted
anchor-not-found is "never reached here in practice because
collector.go only calls this for a TargetAnchor a prior resolution step
already confirmed exists" — that precondition was never actually
enforced. `internal/operations.References`/`Backlinks` populate
`TargetAnchor` straight from the wikilink's own raw `#anchor` text with
no existence check (existence checking is `internal/validation`'s
separate job, `CodeUnknownAnchor`, §2) — so a wikilink naming a
nonexistent anchor on an otherwise-real target reaches
`chunkArtifactAnchor` in ordinary `Collect` use, not just as a
theoretical edge case. The first implementation's `chunkArtifactAnchor`
masked this by silently returning a near-empty `Candidate{Path:
relPath, Reasons: []Reason{reason}}` (no Heading/Content, StartLine/
EndLine 0) — a phantom, contentless item with a nonzero score, silently
included in the Context Pack. Fixed: `chunkArtifactAnchor` now returns
a sentinel `errAnchorNotFound`, and `connectedCandidates` checks for it
with `errors.Is` and silently omits that one reference (`continue`),
matching the codebase's existing tolerance for an unresolved wikilink
target (`references.go`'s own `if err != nil || len(paths) != 1 {
continue }`) — never fabricating a placeholder Candidate, never failing
the whole `Collect` call over one stale anchor. The actual diagnostic
still surfaces separately via `internal validate`'s `CodeUnknownAnchor`
Finding (§2), exactly as this contract's separation of concerns always
intended — only the retrieval-side enforcement of that separation was
missing.

## 6. `misterspec internal context <id>` (`schema_version: 5`, was 4)

Existing `--provenance` shape (038, §4) is unchanged aside from the
version bump. Any item produced via an anchor-qualified reference gains
two additional keys, `heading_path` and `anchor` (the causing Reason's
`TargetAnchor`), alongside its existing `provenance` entries when
`--provenance` is also passed:

```json
{
  "path": "specs/knowledge/retry-policies.md",
  "heading": "Client Retry Policy",
  "heading_path": ["Networking Policies"],
  "anchor": "retry-policy",
  "tier": "semantic",
  "reasons": ["wikilink"],
  "score": 21
}
```

An item reached without an anchor omits both `heading_path` and
`anchor` entirely (not empty-valued — absent), matching 038's own
"no provenance field at all" convention for items with no occurrence
data.

## 7. `internal/context/index` (`schemaVersion: 3`, was 2)

```sql
ALTER TABLE chunks ADD COLUMN anchor TEXT;         -- conceptual; schema.go recreates from scratch
ALTER TABLE links  ADD COLUMN target_anchor TEXT;  -- conceptual; schema.go recreates from scratch
```

A version mismatch (including this bump) triggers the index's own
already-existing full rebuild — no manual migration statement is ever
actually run (Constitution Principle III, 038 precedent).
