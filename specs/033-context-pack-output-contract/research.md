# Phase 0 Research: Context Pack completo e contrato de saída versionado

No `[NEEDS CLARIFICATION]` markers were left in the Technical Context. This document records the design decisions the current code's actual behavior forces, verified against `internal/cli/internalcmd/context.go`, `internal/context/*.go`, and `internal/artifacts` rather than assumed from the backlog document alone.

## Decision 1 — Absolute line numbers are broader than the backlog doc's own premise: every Candidate, not just wikilinks

**Decision**: Fix the line-offset bug at its source — `internal/artifacts` gains a new function (`ReadBodyWithOffset`, or equivalent) that reuses `splitFrontmatter` but also reports the body's own 1-indexed starting line in the whole file. `internal/context/collector.go`'s `chunkArtifact` (the one place every `Candidate` is built from a file) adds that offset to every `Chunk.StartLine`/`EndLine` before constructing a `Candidate`.

**Rationale**: The backlog document frames this as "o parser atual de wikilinks usa linhas relativas ao corpo" — true (`internal/artifacts/wikilink.go:18-21` documents it explicitly), but verified against the actual code, the *same* bug affects every single Context Pack item: `chunkArtifact` (collector.go:206-226) calls `artifacts.ReadBody` (post-frontmatter) then `artifacts.ParseDocument(body)`, so `Section.StartLine`/`EndLine` — and everything built on them (`Chunk`, `Candidate`, `ScoredCandidate`, `ResultItem`) — are body-relative for every item the Context Pack has ever returned, not a wikilink-specific edge case. Fixing it once at the `artifacts` layer (the actual owner of frontmatter/body splitting) means every downstream type (`Chunk`, `Candidate`, etc.) needs no shape change at all — only its *values* become correct, which is also why this fix is purely additive to the existing default/`--render` output (spec.md FR-008): the field names `path`/`heading` stay the same; only the (currently silently wrong) numbers they'd have carried, had this feature exposed them, become right, and this feature is what first exposes them.

**Alternatives considered**:
- Passing the whole file (frontmatter included) to `artifacts.ParseDocument` instead of just the body — rejected: YAML frontmatter has no ATX heading syntax, so it would silently fold into `ParseDocument`'s existing "untitled preamble Section" (Level 0), producing a bogus Chunk containing raw frontmatter text mixed with any real preamble prose — a correctness regression, not a fix.
- Fixing only `internal/artifacts/wikilink.go`'s own `Line` field, matching the backlog document's literal wording — rejected once the broader bug was confirmed: `WikiLink.Line` is not itself part of the Context Pack's JSON output today (wikilinks only feed `Reasons`/`Relation` strings, never their own line number), so fixing only it would leave every actual Context Pack item's location still wrong — the bug this spec's User Story 2 is about.

## Decision 2 — Output modes: additive `--mode` flag, default and `--render` untouched

**Decision**: Add a new `--mode` flag to `internal context <id>` with three values: `manifest` (today's default shape — metadata only, no content, no `rendered`), `package` (new — every item carries full content, file-absolute location, and a fingerprint; no `rendered` field in the same response), and `markdown` (today's `--render` behavior, now also requestable as its own named mode). Omitting `--mode` keeps today's exact default. `--render` keeps working exactly as today (`--mode` unset + `--render` = today's "metadata items + rendered Markdown" combination) — it is not removed or reinterpreted.

**Rationale**: Spec.md FR-008/SC-003 requires zero-migration compatibility for every currently-shipped Skill, none of which pass `--render` (per `mister-{plan,tasks,analyze,implement,wrap-up}/SKILL.md`, none reference it) — so the safest design is additive-only: nothing about the existing default JSON shape or the `--render` flag's own behavior changes at all. `--mode=package` is the only way to opt into the new full-content shape, so an old caller that never learns about `--mode` sees byte-for-byte identical output forever.

**Alternatives considered**:
- Making `--render` itself return structured per-item content (reinterpreting the existing flag) — rejected: `--render`'s documented behavior (a single Markdown string) is exactly what Skills and `site/commands.html` already document; silently changing its meaning would violate FR-008 even if no Skill currently reads the changed field, since it's still a behavior change to a name callers already depend on.
- A single new flag that always includes content in `items` regardless of mode (e.g. `--include-content`) layered onto the existing default — rejected: this reintroduces exactly the duplication FR-005 forbids the moment someone also passes `--render`, and blurs "which shape am I getting" without a named, versioned mode to point at.

## Decision 3 — `--mode=package` and `--render` are mutually exclusive

**Decision**: Passing `--render` together with `--mode=package` is rejected as `invalid_argument` (the CLI-boundary sentinel `internalcmd.ErrInvalidArgument`, already mapped by `classify`) — never silently prioritizing one over the other.

**Rationale**: `package` mode's items already carry full content (Decision 2); also setting `rendered` in that same response would duplicate that same content as prose, exactly what spec.md FR-005 forbids by default. Rejecting the combination outright (rather than silently dropping `--render` or silently omitting `items[].content`) keeps the contract's behavior explicit and machine-checkable (Constitution Principle IX) instead of a caller-surprising implicit rule.

**Alternatives considered**: Silently ignoring `--render` when `--mode=package` is set — rejected: a caller who explicitly asked for both would get one of them dropped with no signal why, the same class of silent-surprise behavior Constitution Principle IX exists to prevent.

## Decision 4 — Per-item fingerprint is a content hash, not a whole-file hash

**Decision**: Each `package`-mode item's fingerprint is a SHA-256 digest of that item's own selected `Content` text (the same bytes returned in `content`), rendered `"sha256:<hex>"` — matching `operations.FileFingerprint`'s existing string format (`internal/operations/fingerprint.go:14-29`) for consistency, but computed as a new, small in-package function over already-in-memory bytes, not a second file read.

**Rationale**: `operations.Fingerprint` (the existing `internal fingerprint <path>` operation) hashes an entire file — coarser than what a Context Pack item needs: two different items from the same file (e.g. two Requirement sections in one `spec.md`) would collapse to the same fingerprint if hashed at the file level, even though only one of them actually changed. A content-level hash lets a consumer detect exactly when *this* item's own text changed, which is what spec.md's Edge Cases ("um arquivo de origem muda... o fingerprint de cada item deve permitir detectar essa divergência") asks for. Reusing the existing `"sha256:<hex>"` string shape (rather than inventing a new format) keeps the project's one fingerprint vocabulary consistent (Constitution Principle VI, DRY) without importing `internal/operations` into `internal/context` for the formatting alone — the format is trivial enough to reproduce locally with `fmt.Sprintf("sha256:%x", sha256.Sum256(...))`.

**Alternatives considered**: Reusing `operations.Fingerprint` directly (a whole-file hash) for every item from that file — rejected for the coarseness reason above; it would also require a second file read per item (the operation reads from disk), when the content is already held in memory from `Collect`.

## Decision 5 — `schema_version` lives on the envelope, not per-item

**Decision**: The `context` JSON envelope gains a top-level `schema_version` integer field (starting at `1`), sibling to `target`/`intent`/`budget`. It is present in every mode's response, including the unchanged default/`--render` shape (an additive field costs nothing to existing consumers that don't look for it).

**Rationale**: Spec.md FR-006 asks the *response* to identify its contract version — the envelope, not each item, is what a consumer branches on ("do I know how to read this shape"), matching the existing precedent of `project.Configuration.SchemaVersion` versioning the project config file as a whole rather than per-field. Starting the counter at `1` (not `0`) makes an absent field trivially distinguishable from "explicitly version 0" for any future consumer that stores or logs this value.

**Alternatives considered**: A version string embedded in each item (e.g. tagging `content`'s own format) — rejected: nothing about an individual item's shape is expected to version independently of the envelope; one counter for the whole response is simpler (Principle IV) and matches how a consumer actually needs to reason about compatibility (the whole shape, not per-field).

## Decision 6 — Output-size diagnostics are additive fields, not a redefinition of existing ones

**Decision**: `Diagnostics.TokensSelected` (and its siblings) keep their existing, already-relied-upon meaning: content-only token estimation (`artifacts.EstimateTokens(c.Content)` summed over selected items — unchanged, `budget.go:79`). A new diagnostics field, `payload_tokens` (or equivalent), is added alongside them, estimating the actual serialized size of the response in the mode requested (item metadata + content, or the Markdown string, as applicable) — always `>= tokens_selected`, since it additionally accounts for envelope/per-item metadata overhead.

**Rationale**: Spec.md FR-007 requires the reported size to reflect what was actually delivered — verified that `budget.go` counts content only today (research confirms no envelope/serialization overhead anywhere in the budget math), so a `package`-mode response with rich per-item metadata would otherwise under-report its own true size. Keeping `tokens_selected` unchanged (rather than redefining it to include overhead) preserves FR-008 compatibility, since any existing consumer reading that field for budget-fit reasoning keeps getting the same number it always did; the new field is strictly additive.

**Alternatives considered**: Redefining `tokens_selected` itself to include serialization overhead — rejected: this is the one existing field a consumer might already reason about numerically (e.g. comparing it against a requested `--budget`); silently changing its meaning would violate FR-008 even though the field name stays the same, which is a subtler but real compatibility break Constitution Principle IX's "stable, machine-readable contracts" guidance warns against.

## Decision 7 — A Task-derived item's location needs no new mechanism beyond Decision 1

**Decision**: No dedicated "Task location" handling is added beyond fixing line numbers to be file-absolute (Decision 1). A Chunk built from a Task's own `## TASK-NNN — Title` heading (`kit/templates/tasks.md.tmpl`'s convention, already an ordinary ATX heading `artifacts.ParseDocument` already chunks per-section) already carries `Heading` and an `[StartLine, EndLine]` range bounded to exactly that Task's own section — nothing about a Task's location is special-cased elsewhere in the collection pipeline.

**Rationale**: Verified `internal/artifacts/chunk.go`'s `Chunks` derives one `Chunk` per non-empty `Section`, and a Task's own body already sits inside its own heading-bounded Section — so once Decision 1's offset fix lands, a Task-derived Context Pack item's `heading` + file-absolute `[start_line, end_line]` already identify that one Task specifically within a shared `tasks.md`, satisfying spec.md FR-009 with no additional code.

**Alternatives considered**: Reusing 031-canonical-task-identity's `"<path>#TASK-NNN"` location string format for Task-derived Context Pack items specifically — rejected as unnecessary duplication of information already fully carried by `heading` + absolute `start_line`/`end_line`; introducing a second, format-specific location string for one entity type would be inconsistent with every other item's plain path+line location and adds a special case Constitution Principle IV would flag without a demonstrated need.
