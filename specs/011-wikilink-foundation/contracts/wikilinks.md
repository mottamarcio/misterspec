# Phase 1 Contracts: Wikilink Graph Foundation

**Reconciled against the actual implementation (T016)** — zero drift.
Every signature below (`WikiLink`, `ExtractWikiLinks`, `ReadBody`, the
three Finding code constants and their exact string values) matches
what shipped, verbatim. `checkWikilinks`'s signature
(`checkWikilinks(root string, cfg project.Configuration, filePath
string) []Finding`, data-model.md) also matches exactly, wired into
`checkEntity` at one new call site as planned. No unexported helper
introduced during implementation (`classifyWikilink`, the actual
per-link classification helper `checkWikilinks` calls out to) changes
any of the contract below — it is an internal decomposition detail,
not part of this feature's exported/documented surface.

No HTTP/CLI surface is added by this feature (research.md) — its
contract is the new Go API in `internal/artifacts` and the new,
generically-flowing Finding codes in `internal/validation`.

## `internal/artifacts` (extended)

```go
package artifacts

// WikiLink is one explicit, author-written semantic link from an
// artifact's body to another artifact's ID (docs/context-engine-
// implementation.md §6.2).
type WikiLink struct {
    Target string
    Alias  string
    Line   int // 1-indexed, relative to the body passed in (research.md)
}

// ExtractWikiLinks parses body for "[[TARGET]]" and "[[TARGET|Alias]]"
// occurrences, in document order (FR-001, FR-002). It never resolves a
// target against real project state (FR-004) — that is
// internal/validation's job. It never treats a standard Markdown link,
// text inside an inline code span or fenced code block, or an
// unterminated "[[" as a link (FR-003).
func ExtractWikiLinks(body []byte) ([]WikiLink, error)

// ReadBody reads path and returns its Markdown body — everything after
// the artifact's frontmatter block's closing delimiter. The
// counterpart to ParseMetadata, which returns only the frontmatter
// half of the same file.
func ReadBody(path string) ([]byte, error)
```

**Guarantees**:
- `ExtractWikiLinks` is a pure function of its input — no filesystem
  access, no target resolution, deterministic output for the same
  input (matches every prior feature's own determinism discipline).
- `ParseMetadata`'s own behavior (every existing caller, every existing
  test) is unchanged — `splitFrontmatter`'s widened return shape is
  purely additive and has exactly one call site before this feature and
  after it (research.md).

## `internal/validation` (extended)

```go
package validation

// New Finding codes (docs/context-engine-implementation.md §6.4).
const (
    CodeInvalidWikilink   = "invalid_wikilink"
    CodeBrokenWikilink    = "broken_wikilink"
    CodeAmbiguousWikilink = "ambiguous_wikilink"
)
```

`ValidateProject` and `ValidateEntity`'s own exported signatures are
unchanged — both already return `[]Finding`, and both already call
`checkEntity` internally, which now additionally calls the new,
unexported `checkWikilinks` for each of the five already-validated
entity types.

**Guarantees**:
- Every link classifies into exactly one of the three codes above, or
  produces no finding at all — never zero-or-more-than-one for the same
  link (FR-006, FR-007, FR-008).
- An artifact with no links in its body contributes zero findings from
  this check — `ValidateProject`/`ValidateEntity`'s output for such an
  artifact is byte-for-byte identical to before this feature existed
  (FR-009, SC-003).

## Cross-cutting: no change to the machine-readable contract surface

`internal/cli/internalcmd/validate.go`'s JSON shape
(`{"ok":true,"valid":bool,"findings":[{"code","severity","path","message"}]}`,
008-cli-cobra) is unchanged verbatim — the three new codes appear in
`findings[].code` exactly like every existing code already does, with
no new field, no new command, no new flag (research.md).
