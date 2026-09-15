# Phase 1 Data Model: Context Collector and Retrieval

Entities extracted from `spec.md`'s Key Entities plus
`docs/context-engine-implementation.md` §13-18. This feature adds three
new files to `internal/context` (its first real content beyond the
`index/` subpackage 014 already built) — no new package, no new
dependency.

## `internal/context.Intent` (new)

| Value | Notes |
|---|---|
| `IntentPlanning` (`"planning"`) | |
| `IntentTasks` (`"tasks"`) | |
| `IntentImplementation` (`"implementation"`) | |
| `IntentValidation` (`"validation"`) | |
| `IntentAnalysis` (`"analysis"`) | |
| `""` (zero value) | Valid — "no particular intent." |

Any other string is rejected (FR-002, research.md #4).

## `internal/context.Request` (new)

| Field | Type | Notes |
|---|---|---|
| `Target` | string | Required (FR-001) — a raw entity ID, resolved via `operations.Inspect`. Must be one of the five types `operations.References` covers (research.md #3). |
| `Task` | string | Optional. Also used as Tier 4's own query when `Query` is empty (research.md #6). |
| `Intent` | `Intent` | Optional (empty is valid). |
| `Query` | string | Optional free-text question. |

No `Budget` field — deliberately deferred to Phase 7 (research.md #2).

## `internal/context.Tier` (new)

| Value | Meaning |
|---|---|
| `TierMandatory` | The Constitution and the target itself (FR-003). |
| `TierStructural` | A formal relationship (`parent`/`depends_on`/`supersedes`) directly on the target. |
| `TierSemantic` | A direct semantic connection (`wikilink`) or incoming reference (`backlink`) directly on the target. |
| `TierText` | A free-text match from the local search index. |
| `TierSecondHop` | An outgoing relationship one further hop from a Tier 2/3 artifact (research.md #7). |

Ordered ascending in this exact sequence — the `CandidateSet`'s own
primary sort key (research.md #10).

## `internal/context.Reason` (new)

| Field | Type | Notes |
|---|---|---|
| `Tier` | `Tier` | |
| `Relation` | string | `"constitution"` \| `"target"` (Tier 0/1); `"parent"` \| `"depends_on"` \| `"supersedes"` \| `"wikilink"` (Tier 2/3 outgoing, and reused at Tier 5); `"backlink"` (Tier 3 incoming); `"text_match"` (Tier 4). Never empty (FR-009). |

## `internal/context.Candidate` (new)

| Field | Type | Notes |
|---|---|---|
| `Path` | string | |
| `Heading` | string | |
| `Content` | string | |
| `StartLine` / `EndLine` | int | Together with `Path`, this is the Candidate's own identity for deduplication (research.md #8). |
| `Reasons` | `[]Reason` | Every distinct reason this exact chunk was included — never empty (FR-009), never re-derivable from a single field alone once merged (FR-008). |

## `internal/context.CandidateSet` (new)

| Field | Type | Notes |
|---|---|---|
| `Candidates` | `[]Candidate` | Deduplicated (FR-008), deterministically ordered (research.md #10). |

## `internal/context.Collect` (new)

| Symbol | Notes |
|---|---|
| `Collect(root string, cfg project.Configuration, store index.Store, req Request) (CandidateSet, error)` | See Algorithm below. Strictly read-only (FR-012) — never mutates a project artifact, the reference graph, or the search index. |

**Validation rules**: `req.Target` must resolve via `operations.Inspect`
to one of the five referenceable types, or `Collect` returns
`operations.ErrInvalidTarget` (FR-002, research.md #3). `req.Intent`
must be `""` or one of the five recognized values, or `Collect` returns
an error (FR-002).

## Algorithm

```text
Collect(root, cfg, store, req):
  1. Validate req.Intent (FR-002).
  2. target := operations.Inspect(root, cfg, req.Target)   [rejects an
     unknown or out-of-scope target — FR-002, research.md #3]

  3. Tier 0 (Mandatory — Constitution): if cfg.ConstitutionPath exists,
     chunk it (ReadBody + ParseDocument + Chunks, 011/013); every
     resulting Chunk -> Candidate{Reasons: [{TierMandatory, "constitution"}]}.
     Absence is not an error (FR-004).
  4. Tier 1 (Mandatory — target): chunk target.Location.Path the same
     way; every Chunk -> Candidate{Reasons: [{TierMandatory, "target"}]}.

  5. Tier 2/3 (structural/semantic, direct): refs :=
     operations.References(root, cfg, req.Target); backlinks :=
     operations.Backlinks(root, cfg, req.Target). For each entry in
     refs.Formal/refs.Semantic and backlinks.Formal/backlinks.Semantic,
     resolve its own artifact's path (operations.Inspect on
     entry.Target/entry.Source), chunk it, and add every Chunk as a
     Candidate labeled Tier 2 (formal) or Tier 3 (semantic/backlink) per
     data-model.md's own Reason table (research.md #5 — filesystem-
     direct, no index read).

  6. Tier 4 (text): query := req.Query, or req.Task if req.Query == ""
     (research.md #6); if query != "", store.Search(query, ...) ->
     each SearchResult -> Candidate{Reasons: [{TierText, "text_match"}]}.

  7. Tier 5 (second-hop): for each artifact discovered in step 5 (the
     first-hop set) that is itself one of the five referenceable types,
     operations.References(root, cfg, thatArtifact) -> its own
     outgoing entries -> Candidates labeled TierSecondHop, Relation =
     that entry's own relation (research.md #7) — never that artifact's
     own backlinks, never a third hop.

  8. Deduplicate by (Path, StartLine, EndLine): merge Reasons; a direct
     (Tier 2/3) classification always wins over a TierSecondHop one for
     the identical chunk (FR-011, research.md #8).
  9. Sort by Tier, then Path, then StartLine (research.md #10).
  10. Return CandidateSet{Candidates: ...}.
```

## State / Flow Summary

```text
Request{Target, Task, Intent, Query}
      ↓ operations.Inspect (scope-check, research.md #3)
target's own path
      ↓ artifacts.ReadBody/ParseDocument/Chunks (011/013)     [Tier 0/1]
      ↓ operations.References/Backlinks (012)                  [Tier 2/3]
      ↓ index.Store.Search (014)                                 [Tier 4]
      ↓ operations.References on first-hop artifacts (012)        [Tier 5]
      ↓ dedup by (Path, StartLine, EndLine); merge Reasons
      ↓ sort by Tier, Path, StartLine
CandidateSet — this feature's own output; not yet ranked by score, not
  yet trimmed to any budget (Phase 7's own job)
```
