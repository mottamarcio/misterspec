# Phase 0 Research: Wikilink Graph Foundation

`docs/context-engine-implementation.md` §5-6, §29.1-29.2 give this
feature a concrete envelope (syntax, parsing rules, validation finding
names, a full test list) but leaves package placement and a few
implementation choices for this project's own actual code to resolve.
This document records those decisions.

## Decision: No new Markdown-parsing dependency — a hand-written, line-based scanner

- **Decision**: `ExtractWikiLinks` is implemented as a small, hand-written
  line-by-line scanner in `internal/artifacts` — no `goldmark` or other
  CommonMark library dependency.
- **Rationale**: The only two exclusion rules `ExtractWikiLinks` must get
  right are "not inside a fenced code block" and "not inside an inline
  code span" — both are straightforward with a line-based state machine
  (track fenced-block state across lines; mask backtick-delimited spans
  within a line before matching `[[...]]`). Every artifact this project
  authors (including `docs/architecture-specification.md` itself)
  already uses only simple, single-backtick inline code and simple
  triple-backtick fences — a full CommonMark parser would handle CommonMark
  edge cases (nested emphasis, multi-backtick spans, lazy continuation
  lines) this project has no actual content exercising. Every dependency
  this project has added so far (`cobra`, `bubbletea`, `lipgloss`) was for
  something a hand-written implementation could not reasonably replace
  (a CLI framework, a full TUI runtime); a ~150-line scanner does not
  clear that bar (Constitution Principle IV, YAGNI).
- **Alternatives considered**: `github.com/yuin/goldmark` (the standard,
  actively maintained, pure-Go CommonMark parser) walking its AST for
  text-node positions — rejected as the more "correct" but disproportionate
  choice for what this project's own real Markdown content actually
  needs; revisit only if real artifacts are found using Markdown
  constructs the scanner gets wrong.
- **Explicit implementation choices carried into tests**: fenced code
  blocks are detected by a line whose trimmed content starts with
  ` ``` ` or `~~~` (matching fence markers close the block; this project's
  own docs never mix fence styles or vary fence length); inline code
  spans are single-backtick delimited (`` `...` ``) — the only style used
  anywhere in this repository's existing Markdown; a wikilink is always
  fully contained on one line (no attempt to match `[[` against a `]]`
  on a later line — an unterminated `[[` is simply not a link, per
  spec.md's own Edge Case).

## Decision: `WikiLink.Line` is relative to the body passed to `ExtractWikiLinks`, not the whole file

- **Decision**: `ExtractWikiLinks(body []byte)` returns line numbers
  counted from the start of `body` itself (1-indexed) — it has no
  knowledge of, or offset for, a frontmatter block that may have
  preceded it in the original file.
- **Rationale**: `ExtractWikiLinks` is a pure, self-contained lexical
  function per its own contract (`docs/context-engine-implementation.md`
  §6.3: "avoid resolving targets during lexical parsing") — it does not
  know whether its input came from a real artifact file, a fixture, or
  a hand-constructed byte slice in a test, and should not need to. A
  caller that wants file-absolute line numbers (not required by this
  feature's own spec, which asks only for "exact source line," not
  specifically file-absolute) can add its own known frontmatter-line
  offset; no such caller exists yet in this feature's own scope.
- **Alternatives considered**: Passing the whole file (including
  frontmatter) and reporting file-absolute lines — rejected; it would
  make `ExtractWikiLinks` implicitly depend on frontmatter always being
  present and well-formed, coupling a pure lexical function to a
  concern (frontmatter parsing) `internal/artifacts/parser.go` already
  owns separately.

## Decision: `parser.go`'s `extractFrontmatter` is refactored (not duplicated) to also return the body

- **Decision**: `extractFrontmatter` (private, one call site —
  `ParseMetadata`) becomes `splitFrontmatter(data []byte) (frontmatter,
  body []byte, err error)`, returning both halves. `ParseMetadata`'s own
  behavior is completely unchanged — it simply ignores the new second
  return value. A new exported function, `ReadBody(path string)
  ([]byte, error)`, reads a file and returns `splitFrontmatter`'s body
  half, for `ExtractWikiLinks`'s own callers (`internal/validation`).
- **Rationale**: The delimiter-scanning logic (find the first `---`,
  find the closing `---`) is identical whether the caller wants the
  frontmatter half or the body half — writing a second, independent
  scanner in a new file would duplicate it for no reason. `extractFrontmatter`
  is unexported with exactly one call site, so widening its return
  shape is a safe, fully internal change — the same low-risk internal
  refactor precedent 002-read-operations's and 006-agent-adapter's own
  research.md documents used for comparably small, single-call-site
  helpers.
- **Alternatives considered**: A second, independent body-extraction
  scanner in `markdown.go` — rejected as needless duplication of
  `parser.go`'s already-correct, already-tested delimiter logic
  (Constitution Principle VI, DRY).

## Decision: The three wikilink Finding codes reuse `ids.ParseAny`/`ids.Scan` exactly — no new ID semantics

- **Decision**: For each extracted link, `internal/validation`'s new
  check classifies it using only two already-existing primitives:
  - `ids.ParseAny(link.Target)` fails → `invalid_wikilink` (the target
    is not even syntactically a recognizable entity ID — wrong prefix,
    non-numeric suffix, empty, or similar).
  - `ids.ParseAny` succeeds, then `ids.Scan(root, cfg,
    parsed.Type).Paths[parsed.Number]` has zero entries →
    `broken_wikilink` (syntactically fine, but nothing with that ID
    actually exists).
  - The same `Paths[parsed.Number]` has more than one entry →
    `ambiguous_wikilink` (the target genuinely resolves to more than
    one artifact) — the *link's own* finding, attached at the linking
    artifact's location, distinct from the `duplicate_id` finding
    `internal/validation` already attaches at the *target's own*
    definition site(s) for the identical underlying condition.
  - Exactly one entry → no finding; the link is valid.
- **Rationale**: This is precisely the same "syntax via `ids.ParseAny`,
  existence/count via `ids.Scan`" pattern `internal/operations.Resolve`
  and `internal/validation`'s own existing `checkParent`/
  `checkDependencyList` already use — FR-005's own requirement ("the
  exact same entity-ID rules the project already uses everywhere else")
  made concrete, and the reason `ambiguous_wikilink` needs no new
  concept of "ambiguity" invented for this feature: this project
  already has exactly one meaning for it (`ids.Scan` finding more than
  one path for the same type+number), established since
  002-read-operations.
- **Alternatives considered**: Treating "ambiguous" as referring to
  something else (e.g. a target string that could plausibly parse as
  more than one entity type) — rejected; this project's ID scheme
  (`internal/ids`) gives every entity type a fixed, non-overlapping
  prefix, so that condition cannot actually occur — inventing a check
  for an impossible case would be speculative, not real (Constitution
  Principle IV).

## Decision: Wikilink validation is scoped to the same five entity types `ValidateEntity`/`checkEntity` already validate

- **Decision**: The new check runs inside `internal/validation`'s
  existing `checkEntity` (shared by `ValidateProject` and
  `ValidateEntity`), for Program, Feature, Spec, Knowledge, and Learning
  only — the exact same `validatableTypes` set 004-structural-validation
  already established. Task, Plan, Tasks, Validation, and Constitution
  artifacts are not wikilink-checked by this feature.
- **Rationale**: `checkEntity` already runs exactly for these five types,
  reading each one's file via the same `artifacts.ParseMetadata`-adjacent
  path resolution this feature's `ReadBody` reuses — adding the wikilink
  check here costs nothing new in artifact discovery. Task only ever
  gets a lighter, heading-based duplicate-ID check today (no per-entity
  structural validation exists for it yet); Plan, Tasks, Validation, and
  Constitution have no structural validation at all today — extending
  wikilink-checking to any of them would first require building that
  validation from scratch, a materially larger scope this feature's own
  spec.md does not ask for.
- **Alternatives considered**: Extending validation to Plan/Tasks/
  Validation/Constitution as part of this feature specifically so their
  own "Relevant Knowledge"-style sections could be wikilink-checked too
  — rejected as scope creep; those artifact types gaining structural
  validation at all is a distinct, future concern, not something this
  feature's own narrow Phase 1 boundary should absorb.

## Decision: No CLI change — the existing generic Finding→JSON mapping already covers new codes

- **Decision**: `internal/cli/internalcmd/validate.go` (008-cli-cobra)
  is not touched by this feature at all.
- **Rationale**: `NewValidateCmd`'s `RunE` already maps every
  `validation.Finding`'s `Code`/`Severity`/`Path`/`Message` into JSON
  generically — it holds no hardcoded allowlist of known codes. Three
  new codes flow through the existing, already-tested envelope
  automatically, the same way 004-structural-validation's own nine
  original codes already do.
- **Alternatives considered**: None meaningfully different — this is
  simply confirming the existing layering already does what's needed,
  the payoff of every prior feature's own discipline about keeping the
  CLI layer a thin, generic adapter (008-cli-cobra's own research.md).

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
