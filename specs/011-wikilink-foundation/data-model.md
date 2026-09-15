# Phase 1 Data Model: Wikilink Graph Foundation

Entities extracted from `spec.md`'s Key Entities plus
`docs/context-engine-implementation.md` §6, §6.4. This feature adds two
new files to `internal/artifacts` (plus one small refactor of an
existing unexported function), one new file to `internal/validation`,
and three new Finding codes — no new package, no new CLI surface.

## `internal/artifacts.WikiLink` (new)

The Link entity from spec.md's Key Entities, in its parsed form.

| Field | Type | Notes |
|---|---|---|
| `Target` | string | The raw target token as written (e.g. `"SPEC-014"`) — not yet validated as a real entity ID; that is a separate step (research.md, FR-004). |
| `Alias` | string | The display alias, if written (`[[TARGET\|Alias]]`); empty otherwise. Presentation-only — never affects resolution (§6.1). |
| `Line` | int | 1-indexed line number, relative to the `body` byte slice `ExtractWikiLinks` was given (research.md — not necessarily file-absolute). |

## `internal/artifacts` (extended)

| Symbol | Type | Notes |
|---|---|---|
| `ExtractWikiLinks` | `func(body []byte) ([]WikiLink, error)` | Pure lexical extraction, no target resolution (FR-002, FR-004). Returns links in document order. Never treats a standard Markdown link, inline-code-span content, fenced-code-block content, or an unterminated `[[` as a link (FR-003). |
| `ReadBody` | `func(path string) ([]byte, error)` | Reads path and returns everything after its frontmatter block's closing delimiter — the same body `ExtractWikiLinks` expects. Reuses `splitFrontmatter` (below), not a second parser. |
| `splitFrontmatter` (unexported) | `func(data []byte) (frontmatter, body []byte, err error)` | `extractFrontmatter`'s existing delimiter-scanning logic, widened to also return the body half (research.md). `ParseMetadata`'s own behavior is unchanged — it uses only the `frontmatter` half, exactly as before. |

**Validation rules**: `ExtractWikiLinks` never fails because a target
doesn't exist — only a genuine parse failure (none expected in this
feature's own scope; the function signature returns `error` for
forward-compatibility, matching `docs/context-engine-implementation.md`
§6.3's own suggested signature, but no failure path exists in this
feature's actual implementation beyond a well-formed empty result for
an empty or link-free body).

## `internal/validation` (extended)

### New Finding codes

| Code | Meaning | Detected via |
|---|---|---|
| `CodeInvalidWikilink` (`"invalid_wikilink"`) | The link's target token is not even syntactically a recognizable entity ID. | `ids.ParseAny(link.Target)` returns an error. |
| `CodeBrokenWikilink` (`"broken_wikilink"`) | The target is syntactically valid but nothing with that ID exists. | `ids.ParseAny` succeeds; `ids.Scan(...).Paths[number]` is empty. |
| `CodeAmbiguousWikilink` (`"ambiguous_wikilink"`) | The target resolves to more than one existing artifact. | `ids.ParseAny` succeeds; `ids.Scan(...).Paths[number]` has more than one entry. |

**Validation rules** (FR-006, FR-007, FR-008, research.md): exactly one
of the three codes above, or no finding at all, per link — never more
than one finding for the same link, never a generic/unlabeled finding
for a link-related problem.

### `checkWikilinks` (new, unexported)

| Symbol | Notes |
|---|---|
| `checkWikilinks(root string, cfg project.Configuration, filePath string) []Finding` | Called from `checkEntity` (existing, unmodified call site added) for each of the five already-validated entity types (Program, Feature, Spec, Knowledge, Learning — research.md's scope decision). Reads the artifact's body via `artifacts.ReadBody`, extracts links via `artifacts.ExtractWikiLinks`, classifies each per the table above. |

**Validation rules** (FR-009, SC-003): an artifact whose body contains
zero links produces zero findings from `checkWikilinks` — identical
`ValidateProject`/`ValidateEntity` output to before this feature
existed, for every artifact that doesn't use the new capability.

## State / Flow Summary

```text
artifact body (e.g. SPEC-014/spec.md's Markdown, frontmatter stripped)
      ↓ artifacts.ReadBody                                        [US1]
raw body bytes
      ↓ artifacts.ExtractWikiLinks
[]WikiLink (target, alias, line) — no resolution yet
      ↓ validation.checkWikilinks, called from checkEntity          [US2]
  for each WikiLink:
    ids.ParseAny(Target) fails        → Finding{invalid_wikilink}
    ids.Scan(...).Paths[N] empty      → Finding{broken_wikilink}
    ids.Scan(...).Paths[N] > 1 entry  → Finding{ambiguous_wikilink}
    exactly 1 entry                   → no Finding
      ↓
[]validation.Finding — flows through ValidateProject/ValidateEntity,
  then internal/cli/internalcmd/validate.go's already-generic JSON
  mapping, completely unchanged (research.md)                       [no new CLI code]

An artifact with zero links: checkWikilinks contributes nothing —
ValidateProject/ValidateEntity output identical to before this
feature existed                                                     [US3]
```
