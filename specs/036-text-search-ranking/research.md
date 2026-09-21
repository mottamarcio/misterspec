# Phase 0 Research: Robust Text Search and Measurable Ranking

No `NEEDS CLARIFICATION` markers remain in the Technical Context — all defaults were resolvable from the existing codebase (`internal/context/index`, `internal/context/rank.go`, `internal/context/collector.go`, `internal/cli/internalcmd/context.go`) and prior Specs (016, 019, 033, 035). This document records the technical decisions the plan depends on.

## 1. How to make free-text queries safe against FTS5 syntax

**Decision**: Build the FTS5 `MATCH` expression by tokenizing the free-text query into words and wrapping each token in double quotes as an individual FTS5 string literal (e.g. `foo-bar:baz` → `"foo-bar:baz"`, an internal double-quote in a token is escaped by doubling it, matching SQLite's own string-literal escaping), joined with implicit AND (FTS5's default). This neutralizes every FTS5 reserved character (`"`, `*`, `:`, `-`, parentheses, `NEAR`, `AND`/`OR`/`NOT` as bare words) by never handing FTS5 anything but quoted string literals in free-text mode.

**Rationale**: FTS5's own query grammar treats a double-quoted string as a literal phrase/token with no operator interpretation inside it (per SQLite FTS5 documentation) — quoting is the documented, engine-native way to pass arbitrary text through safely, rather than hand-rolling a character blacklist that could miss an edge case (e.g. `NEAR`, a future FTS5 keyword) or over-escape and break legitimate matches.

**Alternatives considered**:
- **Blacklist/escape specific characters** (`-`, `:`, `*`, `(`, `)`): rejected — brittle, must be kept in sync with FTS5's full reserved-character set, and does not handle bareword operators (`AND`, `OR`, `NOT`, `NEAR`) which are syntax, not characters.
- **Switch to a non-FTS5 `LIKE`/substring fallback for free text**: rejected — throws away BM25 ranking entirely (the exact signal User Story 2 needs), and duplicates query logic outside the virtual table the schema already builds for this purpose.
- **Pre-validate and reject queries containing reserved characters**: rejected — fails the reliability goal directly; User Story 1 requires such queries to *succeed*, not to be refused.

## 2. How to offer an explicit advanced-syntax mode without ambiguity

**Decision**: Add a query mode value at the `Store.Search` call boundary (an explicit parameter/flag, not inferred from the query's own content) — free-text mode quotes/tokenizes as in #1; advanced mode passes the caller's string to FTS5 `MATCH` unmodified. At the CLI layer (`internal/cli/internalcmd/context.go`), this surfaces as an opt-in flag (e.g. `--query-mode=advanced`) alongside the existing `--query` flag, defaulting to free-text mode.

**Rationale**: Spec Assumption explicitly rules out guessing intent from characters present in the input — that would reintroduce the exact ambiguity User Story 1 exists to remove (a hyphen could be a literal dash or a NOT operator; only an explicit mode disambiguates deterministically, per Constitution Principle I's semantic/deterministic split: which mode applies is not something the binary should infer).

**Alternatives considered**:
- **A sigil/prefix in the query string itself** (e.g. leading `!` means "advanced"): rejected — sigils can themselves collide with legitimate free-text content, and silently changes meaning based on the first character rather than an explicit, separately-named parameter.
- **Two separate commands/flags with no shared code path**: rejected as unnecessarily duplicative — both modes share the same `Store.Search(query string, limit int) ([]SearchResult, error)` signature; only the string construction before the call differs.

## 3. How to surface FTS5 syntax errors in advanced mode as a clear diagnostic (FR-004)

**Decision**: Advanced-mode syntax errors returned by the underlying `chunks_fts MATCH` query (SQLite returns a query-time error, not a schema error, for malformed FTS5 syntax) are wrapped into a distinct, stable error identifiable as a query-syntax error (not a generic index/IO failure), following the existing pattern in `internal/context/index/search.go` of wrapping driver errors with `fmt.Errorf("index: ...: %w", err)` — the wrap gets a syntax-specific message prefix so callers (and the CLI's JSON error envelope, per Constitution Principle IX) can distinguish "your advanced query was malformed" from "the index itself failed."

**Rationale**: Constitution Principle IX requires Skills to reason about stable error codes/messages, not prose; a leaked raw SQLite error message is neither stable across driver versions nor CLI-diagnosable at the code level FR-004 requires.

**Alternatives considered**:
- **Let the raw driver error propagate unchanged**: rejected — violates FR-004's "not a generic or leaked low-level engine error" requirement directly.
- **Silently fall back to free-text interpretation on advanced-mode syntax error**: rejected — spec's edge cases explicitly forbid a silent fallback without saying so; it would also hide a real user mistake in the advanced query.

## 4. How to use the real BM25 signal in same-tier ordering without breaking the tier-first invariant

**Decision**: Extend `Candidate` (`internal/context/collector.go`'s Tier 4 branch) to carry the `SearchResult.Rank` (BM25) value already available from `Search`, and change `rank.go`'s `scoreCandidate`/`textRelevance` path to consume that stored value for `text_match` reasons instead of recomputing a term-occurrence count. Because FTS5's `bm25()` is lower-is-better and unbounded in magnitude (unlike the existing `0..40`-range heuristic), it is normalized/inverted into the same additive integer scoring space `rank.go` already uses (`relationWeight` + `intentWeight` + text-relevance contribution), preserving the documented "meaningful only within one Tier" contract (research.md #3 of 016) and the absolute tier-before-score sort order in `Rank()`.

**Rationale**: `Rank()`'s tier-first, then-Score, then-Path, then-StartLine sort (016-ranking-budgeting) is a deliberately protected invariant (FR-002/FR-003 of that Spec, and this Spec's own FR-006/FR-012) — this feature only replaces *how* the text-relevance component of Score is computed for `text_match` reasons, it does not touch tier assignment or the sort function itself.

**Alternatives considered**:
- **Use raw bm25() value directly as Score**: rejected — bm25()'s scale and sign convention (lower/more-negative is better) are incompatible with `rank.go`'s existing higher-is-better additive convention used for `relationWeight`/`intentWeight`, and would require every other scoring contributor to be renormalized too, a much larger and riskier change than this Spec's scope.
- **Replace the whole scoring model with `bm25()` alone**: rejected — would discard `relationWeight`/`intentWeight`, which encode information (relation type, intent match) that free-text relevance alone does not capture; out of scope per spec Assumptions ("existing priority-tier model... is out of scope to change").

## 5. How to expose score components in diagnostic output (FR-008) without polluting default output

**Decision**: Extend the existing diagnostics machinery already present in `internal/cli/internalcmd/context.go` (`result.Diagnostics...`, surfaced today for budget/estimator info per 035) with an opt-in diagnostic block per scored item — tier, BM25-derived contribution, relation weight, intent bonus — gated behind the same kind of explicit request the command already uses for verbosity (consistent with 033's "avoid duplicating text in JSON/Markdown by default" decision), not emitted in the default compact manifest.

**Rationale**: FR-008 explicitly requires this only "in a diagnostic/debug output mode," and the Spec's own Assumptions state this detail is not part of the compact, default output agents consume during normal task execution — matching the precedent 033 and 035 already set for keeping the default envelope lean while still making richer detail available on request.

**Alternatives considered**:
- **Always include score components in every response**: rejected — spec explicitly scopes this to diagnostic mode; also increases default payload size for information most task executions don't need (in tension with 035's token-budget accuracy work).

## 6. How to version ranking changes and gate them on evaluation evidence (FR-009, FR-010, SC-005)

**Decision**: Add a `ranking_version` constant (parallel to `contextSchemaVersion` in `internal/cli/internalcmd/context.go`, currently 2) surfaced in the `context` command's diagnostic output, bumped whenever the scoring formula or its weights change. Any proposed weight change is evaluated first using the existing evaluation protocol (`specs/019-dogfooding-evaluation/contracts/evaluation-protocol.md`'s User Story 1/4 comparative protocol) against a documented baseline before the new `ranking_version` becomes the shipped default — following the same "evidence before promotion" discipline already called for in the source proposal document and in 019's own Tuning Decision step.

**Rationale**: 019 already defined exactly this kind of "Finding → Tuning Decision, evidence-gated" workflow for ranking/budgeting behavior; reusing it (rather than inventing a second evaluation mechanism) satisfies Constitution Principle IV (no speculative new layer) and directly answers FR-010/SC-005.

**Alternatives considered**:
- **A new, separate ranking-changelog document**: rejected as redundant — `ranking_version` plus the existing 019 protocol already gives traceability and an evidence gate without a new artifact type.
- **No explicit version, rely on git history alone**: rejected — FR-009 requires that a given result set be traceable to the ranking behavior that produced it at the point of use (i.e., from the JSON response itself), which git history alone cannot answer without out-of-band lookup.

## 7. Tie-break stability (FR-007, SC-003)

**Decision**: No change to the existing tie-break chain in `Rank()` (Tier → Score → Path → StartLine, from 016-ranking-budgeting) — replacing `textRelevance` with a BM25-derived contribution (research item #4) still feeds into the same `Score` field, so the existing deterministic tie-break continues to apply unchanged when BM25-derived scores are equal.

**Rationale**: The Spec's own edge cases and FR-007 require the *existing* deterministic convention be preserved, not reinvented; 016 already proved this tie-break stable and it requires no modification for this feature's scope.

**Alternatives considered**: None — introducing a new tie-break key was never indicated by the Spec and would be unjustified scope expansion under Constitution Principle IV.
