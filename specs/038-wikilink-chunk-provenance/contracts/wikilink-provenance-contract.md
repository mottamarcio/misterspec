# Contract: Wikilink Occurrence Provenance

This documents the machine-readable contracts this feature adds/changes
(Constitution Principle IX). All three surfaces below are additive —
existing default output is byte-identical to before this feature,
except where a contract's own `schema_version` already required a
bump for other reasons.

## 1. `internal/artifacts` (extended)

```go
package artifacts

// ReferenceOccurrence is one specific, identifiable place a wikilink
// was written (data-model.md "ReferenceOccurrence").
type ReferenceOccurrence struct {
	SourceSection string // enclosing Chunk's Heading; "" if none
	SourceLine    int    // file-absolute (033 research.md Decision 1 precedent)
}

// OccurrenceFor reports the ReferenceOccurrence for link within doc's
// own already-computed chunks — the enclosing Chunk is the one whose
// [StartLine, EndLine] contains link.Line (research.md #1). Returns
// the zero value (SourceSection: "") when link.Line falls before any
// heading — never an error.
func OccurrenceFor(link WikiLink, chunks []Chunk) ReferenceOccurrence
```

## 2. `internal/operations` (extended)

```go
package operations

// ReferenceEntry gains occurrence fields (data-model.md
// "ReferenceEntry / BacklinkEntry (extended)"). SourcePath is always
// populated — the queried target's own path, known regardless of
// relation kind. SourceSection/SourceLine are populated only for a
// semantic entry (Relation == "wikilink"); zero-valued for a formal
// one — a formal relationship has no line-level wikilink origin (spec
// FR-009).
type ReferenceEntry struct {
	Relation      string
	Target        ids.EntityID
	SourcePath    string // new
	SourceSection string // new
	SourceLine    int    // new
}

// BacklinkEntry gains the same three fields, populated from the
// referencing (Source) artifact's own occurrence. SourcePath is always
// populated (formal or semantic) — the only way a caller learns
// Source's own path without a separate lookup; Backlinks already has
// it in scope while scanning, for every entry it emits.
type BacklinkEntry struct {
	Relation      string
	Source        ids.EntityID
	SourcePath    string // new
	SourceSection string // new
	SourceLine    int    // new
}
```

`References`/`Backlinks` themselves keep their existing signatures —
only the entry types' shape grows.

## 3. `misterspec internal references <id>` / `internal backlinks <id>` (additive, no version bump — research.md #4)

Existing fields (`relation`, `target`/`source`) are unchanged. Each
entry gains three new keys, always present (empty string / `0` for a
formal entry):

```json
{
  "relation": "wikilink",
  "target": "KNOW-003",
  "source_path": "specs/014-sqlite-index/spec.md",
  "source_section": "Requirements",
  "source_line": 42
}
```

```json
{
  "relation": "parent",
  "target": "FEAT-001",
  "source_path": "specs/012-references-backlinks/spec.md",
  "source_section": "",
  "source_line": 0
}
```

No new error conditions; existing `entity_not_found`/`entity_ambiguous`
/`invalid_target` behavior is unchanged.

## 4. `misterspec internal context <id> --provenance` (new opt-in flag, `schema_version: 4`)

**Args**: `--provenance` (optional boolean flag, any output mode).
Omitted (default) → response identical to `schema_version: 3`'s own
shape except the version number itself, mirroring 036's own
`--diagnostic-scores` precedent (research.md #3) — no new field
appears unless requested.

When passed, each item with at least one Reason carrying real
occurrence data gains a `provenance` array — one entry per such
Reason. This includes both an outgoing `"wikilink"` reason and an
incoming `"backlink"` reason built from a semantic (wikilink-sourced)
`BacklinkEntry` — `internal/context/collector.go` always labels
incoming reasons `"backlink"`, never `"wikilink"`, regardless of the
underlying relationship, so gating on the relation string alone would
silently drop every incoming reference's own provenance; the correct
test is whether `source_section`/`source_line` are populated, not the
relation label (data-model.md "Reason (extended)"):

```json
{
  "path": "specs/012-references-backlinks/spec.md",
  "heading": "Requirements",
  "tier": "semantic",
  "reasons": ["wikilink"],
  "score": 18,
  "tokens": 240,
  "provenance": [
    {
      "source_path": "specs/014-sqlite-index/spec.md",
      "source_section": "Requirements",
      "source_line": 42
    }
  ]
}
```

```json
{
  "path": "specs/014-sqlite-index/spec.md",
  "heading": "Requirements",
  "tier": "semantic",
  "reasons": ["backlink"],
  "score": 18,
  "tokens": 240,
  "provenance": [
    {
      "source_path": "specs/012-references-backlinks/spec.md",
      "source_section": "Related Specs",
      "source_line": 17
    }
  ]
}
```

An item with no Reason carrying occurrence data (e.g. `reasons:
["target"]`, `["constitution"]`, `["parent"]`, or `["depends_on"]`, and
a `"backlink"` reason built from a *formal* relation) has no
`provenance` field at all, even with `--provenance` passed — never an
empty array standing in for "not applicable."

`schema_version` in every mode's response becomes `4` regardless of
whether `--provenance` was passed, per 036's own precedent of bumping
the version for the whole feature's contract growth, not per-flag.

## 5. `misterspec internal context <id> --prefer-section` (new, explicit, off-by-default — spec FR-005/FR-006)

**Args**: `--prefer-section` (optional boolean flag, any output mode).
Omitted (default, and the only behavior evaluated/shipped as default
by this feature) → ordering is byte-identical to `--prefer-section`
never having existed.

When passed: among items whose relevant `Reason.Relation == "wikilink"`,
one whose `SourceSection` case-insensitively matches a fixed,
recognized set of requirements-bearing heading names (`"Requirements"`,
`"Functional Requirements"`) is scored as more relevant than an
otherwise-equivalent item whose wikilink occurrence came from any other
section, when tier and other score components are otherwise equal.
This flag is diagnostic/experimental: passing it MUST NOT be
interpreted by any other command, Skill, or default configuration as
implying a permanent behavior change — it exists solely so
037-eval-quality-efficiency's harness can record a comparison against
default ordering (research.md #5). Promoting its effect to default
ordering (dropping the requirement to pass this flag) requires a
future Spec's own `ranking_version` bump, evidenced by that comparison
— never this feature on its own.

## 6. `links` table (widened, index `schemaVersion: 2`)

```sql
ALTER TABLE links ADD COLUMN source_section TEXT;  -- illustrative; the
ALTER TABLE links ADD COLUMN source_line    INTEGER; -- index is rebuilt
                                                        -- from scratch,
                                                        -- never migrated
                                                        -- in place
                                                        -- (Principle III)
```

A caller holding an existing cache database built under
`schemaVersion: 1` observes no manual step — the next `Sync`/`Rebuild`
call detects the `PRAGMA user_version` mismatch and recreates the
schema automatically (existing mechanism, `internal/context/index/
schema.go`), repopulating every table including the two new columns.
