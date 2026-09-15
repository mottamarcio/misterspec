# Phase 1 Data Model: References and Backlinks

Entities extracted from `spec.md`'s Key Entities plus
`docs/context-engine-implementation.md` §5, §7. This feature adds one
small exported helper to `internal/ids`, two new files to
`internal/operations`, and two new files to `internal/cli/internalcmd` —
no new package, no new Finding code, no CLI error-code change.

## `internal/ids` (extended)

| Symbol | Type | Notes |
|---|---|---|
| `ResolveTarget` | `func(root string, cfg project.Configuration, raw string) (EntityID, []string, error)` | Parses `raw` via `ParseAny` (width-tolerant, type-inferring — same tolerance `ParseMetadata`'s frontmatter fields and 011's wikilink classification already rely on), then `Scan`s for it. Returns every matching artifact path — zero (unresolved), one (resolved), or more (ambiguous — a duplicate ID). Only a genuinely malformed `raw` string is an `error`; zero or many matches are not (research.md #2). |

## `internal/operations.ReferenceEntry` (new)

The Reference entity from spec.md's Key Entities, in its outgoing form.

| Field | Type | Notes |
|---|---|---|
| `Relation` | string | One of `"parent"`, `"depends_on"`, `"supersedes"` (formal) or `"wikilink"` (semantic) — never a new fourth kind. |
| `Target` | `ids.EntityID` | The related artifact's ID. For a formal relation, exactly as declared in frontmatter (no existence check — FR-001 reports what's declared). For a semantic relation, only present when the wikilink resolved to exactly one real artifact (FR-002, research.md #3). |

## `internal/operations.ReferencesResult` (new)

| Field | Type | Notes |
|---|---|---|
| `Formal` | `[]ReferenceEntry` | In frontmatter declaration order: `parent` (if any), then each `depends_on` entry, then each `supersedes` entry, in file order (FR-007, research.md #5). Never `nil` when empty is meant — an empty, non-nil slice, so JSON serializes `[]`, not `null` (matching `internal/cli/internalcmd`'s existing `idStrings` convention). |
| `Semantic` | `[]ReferenceEntry` | In document order, exactly as `ExtractWikiLinks` already guarantees, filtered to only exactly-one-match targets. |

## `internal/operations.References` (new)

| Symbol | Notes |
|---|---|
| `References(root string, cfg project.Configuration, rawID string) (ReferencesResult, error)` | Calls `Inspect` (reusing its existing `Resolve`+`ParseMetadata` composition, research.md #4), rejects a non-referenceable type (Task and beyond) via `ErrInvalidTarget`, builds `Formal` directly from `Metadata`, builds `Semantic` by `ReadBody` + `ExtractWikiLinks` + `ids.ResolveTarget` per link. |

## `internal/operations.BacklinkEntry` (new)

The Backlink entity from spec.md's Key Entities, in its incoming form.

| Field | Type | Notes |
|---|---|---|
| `Relation` | string | Same four values as `ReferenceEntry.Relation`. |
| `Source` | `ids.EntityID` | The referencing artifact's own declared ID (its `Metadata.ID` — always non-nil for these five id-bearing types once `ParseMetadata` succeeds). |

## `internal/operations.BacklinksResult` (new)

| Field | Type | Notes |
|---|---|---|
| `Formal` | `[]BacklinkEntry` | Every scoped-type artifact whose `Parent`/`DependsOn`/`Supersedes` names the queried target, discovered by a full scan (research.md #6), in a fixed, deterministic type-then-number order (research.md #5). |
| `Semantic` | `[]BacklinkEntry` | Every scoped-type artifact whose body contains a wikilink resolving to exactly the queried target, same discovery pass, same ordering. |

## `internal/operations.Backlinks` (new)

| Symbol | Notes |
|---|---|
| `Backlinks(root string, cfg project.Configuration, rawID string) (BacklinksResult, error)` | Calls `Inspect` to resolve and validate the target's own scope (same rejection as `References`), then scans every artifact of the five scoped types (`ids.Scan` per type, `ParseMetadata` + `ReadBody` + `ExtractWikiLinks` per discovered path), comparing each candidate's own formal fields and resolved wikilink targets (`ids.ResolveTarget`) against the queried target's `EntityID`. |

## Scope (research.md #1)

Both operations are scoped to exactly the five entity types
`internal/validation` already validates as standalone entities — Program,
Feature, Spec, Knowledge, Learning. A target of any other type (Task,
or a syntactically valid but unsupported request) is rejected via
`operations.ErrInvalidTarget` — the same sentinel `Resolve`/`Inspect`
already use for this class of problem, already mapped by
`internal/cli/internalcmd/errors.go`'s `classify` to `invalid_target`
(no `classify` change needed).

## Resolution rules (research.md #3)

| Wikilink target | `ids.ResolveTarget` outcome | Counted as an edge? |
|---|---|---|
| Well-formed, resolves to exactly one real artifact | `(id, [onePath], nil)` | Yes — the only case that counts. |
| Malformed syntax | `(_, _, err)` | No (same as 011's `invalid_wikilink`). |
| Well-formed, resolves to nothing | `(id, nil, nil)` | No (same as 011's `broken_wikilink`). |
| Well-formed, resolves to more than one artifact (duplicate ID) | `(id, [path, path, ...], nil)` | No, in both directions (same as 011's `ambiguous_wikilink` — research.md #3). |

A formal relationship (`parent`/`depends_on`/`supersedes`) is always
counted as declared — this feature performs no existence check on a
formal target; that integrity question already belongs to
004-structural-validation's `missing_parent`/`unresolved_dependency`
findings (spec.md FR-001 vs. FR-002's own distinction).

## State / Flow Summary

```text
References(root, cfg, "SPEC-014")
  → operations.Inspect("SPEC-014")            [Resolve + ParseMetadata, reused]
  → reject if not in {Program,Feature,Spec,Knowledge,Learning}
  → Formal:   Metadata.Parent/DependsOn/Supersedes, as declared
  → Semantic: ReadBody → ExtractWikiLinks → ids.ResolveTarget each,
              keep only exactly-one-match targets
  → ReferencesResult{Formal, Semantic}

Backlinks(root, cfg, "SPEC-014")
  → operations.Inspect("SPEC-014")             [resolve + scope-check the target itself]
  → for each of the 5 scoped types, for each artifact found:
      meta := ParseMetadata(...)
      if meta.Parent/DependsOn/Supersedes names SPEC-014 → Formal entry
      links := ExtractWikiLinks(ReadBody(...))
      for each link: if ids.ResolveTarget(link.Target) == exactly SPEC-014 → Semantic entry
  → BacklinksResult{Formal, Semantic}, deterministically ordered
```
