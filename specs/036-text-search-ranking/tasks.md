---

description: "Task list for Busca textual robusta e ranking mensurável (PROP-06)"
---

# Tasks: Robust Text Search and Measurable Ranking

**Input**: Design documents from `/specs/036-text-search-ranking/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/search-and-ranking-contract.md, quickstart.md

**Tests**: Required for every user story per Constitution Principle V (Test-First, NON-NEGOTIABLE) — write each test, confirm it fails (`go test`), then implement until green.

**Organization**: Tasks are grouped by user story (US1/US2/US3, matching spec.md's priorities P1/P1/P2) so each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1, US2, or US3 — omitted for Setup/Foundational/Polish tasks

## Path Conventions

Single Go module. All paths are relative to the repository root (`/home/marciovcm/workspace/golang/misterspec`).

---

## Phase 1: Setup

**Purpose**: Confirm the branch starts from a green baseline before any change.

- [X] T001 Run `go build ./... && go vet ./... && go test ./...` on branch `036-text-search-ranking` and confirm it is green before making any change

**Checkpoint**: Baseline confirmed green.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Reserve the shared response-envelope constants once, in one place, so User Story 2 (BM25 diagnostics) and User Story 3 (advanced-mode flag/errors) don't independently race to edit the same `contextSchemaVersion` constant in `internal/cli/internalcmd/context.go` (contracts §3; Constitution Principle IX). User Story 1 needs none of this — it changes no response field, only `internal/context/index/search.go`'s internal behavior.

**⚠️ CRITICAL**: No User Story 2 or User Story 3 work should land its own `schema_version` bump — it is reserved here.

- [X] T002 [P] Write a failing test in `internal/cli/internalcmd/context_test.go` asserting every `internal context` JSON response reports `"schema_version": 3` (replacing 035's `2`) and `diagnostics.ranking_version == 1`, in the default (manifest) output mode (contracts §3, §3.1)
- [X] T003 Bump `contextSchemaVersion` from `2` to `3` in `internal/cli/internalcmd/context.go`; add `const rankingVersion = 1` alongside it, and include `"ranking_version": rankingVersion` in the `diagnostics` map for all three output modes (makes T002 pass; contracts §3.1)
- [X] T004 Run `go build ./... && go vet ./... && go test ./...` and confirm green, including every pre-existing test asserting the old `schema_version: 2` updated to `3`

**Checkpoint**: Shared envelope constants reserved. User stories can now proceed independently.

---

## Phase 3: User Story 1 - Free-text queries never fail on punctuation (Priority: P1) 🎯 MVP

**Goal**: `internal/context/index/search.go`'s `Search` never lets FTS5 reserved characters/keywords in a free-text query reach `MATCH` unescaped, so ordinary punctuation-bearing text never throws a syntax error (spec FR-001, FR-003, FR-011).

**Independent Test**: Run `store.Search` (or `misterspec internal context --query ...`) with queries containing quotes, hyphens, colons, and parentheses and confirm every call returns a result set, never an error.

### Tests for User Story 1

- [X] T005 [P] [US1] Write a failing test `TestSearch_FreeTextWithPunctuationDoesNotError` in `internal/context/index/search_test.go` covering queries such as `retry-policy: "at most once" (idempotent)` and `SPEC-014:TASK-003`, asserting `Search` returns no error (spec FR-001, FR-003, Acceptance Scenario 1; SC-001)
- [X] T006 [P] [US1] Write a failing test `TestSearch_FreeTextMixedPortugueseEnglishNoError` in `internal/context/index/search_test.go` asserting a query mixing Portuguese and English terms completes without error and returns relevance-ordered results (spec Acceptance Scenario 3)
- [X] T007 [P] [US1] Write a failing test `TestSearch_FreeTextPlainProseUnaffected` in `internal/context/index/search_test.go` asserting existing plain-word queries (e.g. `"rotation"`, `"Summary"`) return the same results as `search_test.go`'s current `TestSearch_ReturnsMatchingChunkWithProvenance`/`TestSearch_RespectsLimit` assertions — a regression guard proving free-text mode doesn't change matching for prose with no special characters (spec Acceptance Scenario 2)
- [X] T008 [P] [US1] Write a failing test `TestSearch_FreeTextPunctuationOnlyReturnsEmptyNotError` in `internal/context/index/search_test.go` asserting a query consisting only of punctuation (e.g. `"---():"`) or only sub-3-character fragments returns an empty result set, not an error (spec Edge Cases)

### Implementation for User Story 1

- [X] T009 [US1] Add an unexported `freeTextMatchExpr(query string) string` helper in `internal/context/index/search.go`: split `query` into whitespace-separated tokens, wrap each token as a double-quoted FTS5 string literal (doubling any internal `"` per SQLite string-literal escaping), and join with single spaces (FTS5's implicit `AND`); return `""` when no tokens remain (research.md #1)
- [X] T010 [US1] Change `Search`'s body in `internal/context/index/search.go` to build its `MATCH` argument via `freeTextMatchExpr(query)` instead of passing `query` directly to the SQL query; when `freeTextMatchExpr` returns `""`, return `[]SearchResult{}, nil` immediately without querying the database (makes T005/T006/T008 pass; handles the punctuation-only Edge Case)
- [X] T011 [US1] Run `go build ./... && go vet ./... && go test ./internal/context/... ./internal/example/...` and confirm green, including `search_test.go`'s, `sync_test.go`'s, and `sqlite_index_quickstart_test.go`'s existing calls to `store.Search(...)` (unchanged signature — pure internal behavior fix; makes T007 pass)

**Checkpoint**: User Story 1 is independently functional — free-text queries with punctuation never error.

---

## Phase 4: User Story 2 - Relevance signal reflects actual computed relevance (Priority: P1)

**Goal**: `SearchResult.Rank` (FTS5's own `bm25()` value, already computed by `Search`) survives into `Candidate` and drives same-tier ordering in `Rank()`, replacing `rank.go`'s `textRelevance` term-occurrence heuristic — without ever letting a free-text score cross the existing tier boundary (spec FR-005, FR-006, FR-007, FR-012).

**Independent Test**: Issue a query with several free-text matches of clearly differing relevance and confirm same-tier ordering reflects the underlying BM25 signal, with tier order and deterministic tie-break both preserved.

### Tests for User Story 2

- [X] T012 [P] [US2] Write a failing test in `internal/context/result_test.go` asserting `Candidate` has a `TextRank float64` field, and that `mergeAndSort` sets the merged candidate's `TextRank` from whichever occurrence (not necessarily the first-inserted one for that `candidateKey`) carries a `text_match` Reason (data-model.md Candidate validation rule)
- [X] T013 [P] [US2] Write a failing test in `internal/context/collector_test.go` asserting Tier 4's free-text branch (`Collect`'s handling of `store.Search` results) populates each resulting `Candidate.TextRank` from the corresponding `SearchResult.Rank` (spec FR-005)
- [X] T014 [P] [US2] Write a failing test in `internal/context/rank_test.go` asserting the text-relevance contribution to `scoreCandidate`'s `Score` is derived from a `text_match` candidate's `TextRank` (BM25, lower value = more relevant) rather than from counting term occurrences in `Heading`/`Content` — given two same-tier candidates with differing `TextRank`, the more relevant one (lower BM25) must score higher (spec FR-005, Acceptance Scenario 1)
- [X] T015 [P] [US2] Write a failing test in `internal/context/rank_test.go` asserting `Rank()`'s tier-first ordering is never crossed by the new BM25-derived score — a lower-tier candidate with an extremely strong `TextRank` still sorts after every higher-tier candidate (spec FR-006, FR-012, Acceptance Scenario 3)
- [X] T016 [P] [US2] Write a failing test in `internal/context/rank_test.go` asserting the existing deterministic tie-break (Path, then StartLine) applies unchanged when two same-tier candidates' BM25-derived scores are equal (spec FR-007, Acceptance Scenario 2; SC-003)
- [X] T017 [P] [US2] Write a failing test in `internal/cli/internalcmd/context_test.go` asserting `--diagnostic-scores` is omitted by default (no `score_components` on any item) and, when passed, each item in the response gains `score_components.{tier,relation_weight,intent_bonus,text_relevance,total}` matching contracts §3.2, with `total` equal to the item's own ordering score (spec FR-008; SC-004)

### Implementation for User Story 2

- [X] T018 [US2] Add `TextRank float64` field to `Candidate` in `internal/context/result.go`; in `mergeAndSort`, when an incoming duplicate candidate carries a `text_match` Reason, overwrite the merged candidate's `TextRank` with that duplicate's value (not only append Reasons) (makes T012 pass)
- [X] T019 [US2] In `internal/context/collector.go`'s Tier 4 branch (the loop building `Candidate` from `results`, around line 129), set `TextRank: r.Rank` on each constructed `Candidate` (makes T013 pass)
- [X] T020 [US2] In `internal/context/rank.go`, replace `textRelevance`'s role inside `scoreCandidate`: for a candidate carrying a `text_match` Reason, compute its text-relevance contribution from `c.TextRank` (BM25, lower-is-better) normalized/inverted into the existing higher-is-better additive integer range `textRelevance` used to occupy (research.md #4); a candidate with no `text_match` Reason contributes `0` as today (makes T014 pass)
- [X] T021 [US2] Confirm `Rank()` in `internal/context/rank.go` requires no structural change — its existing tier-then-Score-then-Path-then-StartLine sort already applies to the new `Score` values unmodified; add any missing regression coverage only if T015/T016 reveal a gap (documents that the tier-first invariant and tie-break survive the T020 change)
- [X] T022 [US2] Add a `ScoreComponents{Tier, RelationWeight, IntentBonus, TextRelevance, Total int}` type and a function computing it alongside `scoreCandidate` in `internal/context/rank.go` (data-model.md ScoreComponents), returned only when the caller opts in
- [X] T023 [US2] Add a `--diagnostic-scores` bool flag to the Cobra command in `internal/cli/internalcmd/context.go`; when set, attach each item's `score_components` (from T022, per contracts §3.2 field names) to its JSON representation in all three output modes (`--mode manifest`, `--mode package`, and the plain-Markdown diagnostics companion) (makes T017 pass)
- [X] T024 [US2] Run `go build ./... && go vet ./... && go test ./internal/context/... ./internal/cli/internalcmd/...` and confirm green

**Checkpoint**: User Stories 1 AND 2 both work independently — safe free-text search, and BM25-driven same-tier ordering with opt-in diagnostics.

---

## Phase 5: User Story 3 - Advanced query syntax remains available for power users (Priority: P2)

**Goal**: An explicit, separate advanced-syntax path lets a caller use FTS5's native query grammar deliberately, never inferred from the free-text query's own content, with malformed advanced expressions surfaced as a clear, distinguishable diagnostic (spec FR-002, FR-004).

**Independent Test**: Issue a query explicitly marked `--query-mode advanced` using phrase/boolean/`NEAR` syntax and confirm those operators are honored; issue a malformed advanced query and confirm a distinct `query_syntax_error`, not a leaked engine error or silent fallback.

### Tests for User Story 3

- [X] T025 [P] [US3] Write a failing test `TestSearchAdvanced_HonorsFTS5Operators` in `internal/context/index/search_test.go` asserting a new `Store.SearchAdvanced(query string, limit int) ([]SearchResult, error)` method interprets phrase, prefix, boolean, and `NEAR` syntax natively (spec FR-002, Acceptance Scenario 1)
- [X] T026 [P] [US3] Write a failing test `TestSearchAdvanced_MalformedExpressionReturnsSyntaxError` in `internal/context/index/search_test.go` asserting a malformed advanced query (e.g. an unbalanced quote) returns a distinguishable, wrapped error rather than a raw driver error (spec FR-004, Edge Cases)
- [X] T027 [P] [US3] Write a failing test `TestSearch_OperatorCharactersTreatedAsLiteralInFreeMode` in `internal/context/index/search_test.go` asserting `Search` (free-text mode) treats `NEAR`, `*`, and boolean words as literal text, never as operators — confirming mode separation is real, not just additive (spec FR-002 Acceptance Scenario 2, Assumptions)
- [X] T028 [P] [US3] Write a failing test in `internal/cli/internalcmd/context_test.go` asserting: `--query-mode advanced` routes the query through `SearchAdvanced`; `--query-mode` omitted or `=free` routes through `Search`; an unrecognized `--query-mode` value returns `invalid_argument` (exit 2); and a malformed advanced query returns the new `query_syntax_error` code (exit 2) (contracts §1, §2)

### Implementation for User Story 3

- [X] T029 [US3] Add `SearchAdvanced(query string, limit int) ([]SearchResult, error)` to the `Store` interface in `internal/context/index/store.go`, documented as passing `query` to FTS5 `MATCH` unmodified (research.md #2; Constitution Principle VI OCP — extends the interface rather than overloading `Search`'s existing contract)
- [X] T030 [US3] Implement `SearchAdvanced` in `internal/context/index/search.go`: same query/scan shape as `Search`, but skip `freeTextMatchExpr` and pass `query` straight to `MATCH` (makes T025/T027 pass)
- [X] T031 [US3] Add a distinguishable, stable-wrapped syntax-error path in `SearchAdvanced` (`internal/context/index/search.go` or a new `errors.go` in that package, matching the package's existing `fmt.Errorf("index: ...: %w", err)` convention) that a caller can detect via `errors.Is`/`errors.As`, separate from other index errors (makes T026 pass; contracts §2)
- [X] T032 [US3] Add a `QueryMode` field (or equivalent) to `Request` in `internal/context/request.go`, alongside the existing `Query`/`Task` fields, defaulting to free-text mode when unset; update `Collect`'s Tier 4 branch in `internal/context/collector.go` (around line 124) to call `store.SearchAdvanced` instead of `store.Search` when `Request.QueryMode` selects advanced mode (makes T028's routing assertions pass)
- [X] T033 [US3] Add a `--query-mode` string flag to the Cobra command in `internal/cli/internalcmd/context.go`, validated against `{free, advanced}` (reject any other value as `invalid_argument`, matching the existing `--mode` flag's validation pattern), wired into `Request.QueryMode` (makes T028's validation assertion pass; contracts §1)
- [X] T034 [US3] Map the T031 syntax-error sentinel to the `query_syntax_error` JSON error code with exit code 2 in `internal/cli/internalcmd/context.go`'s error handling, matching the existing `WriteError`/sentinel pattern used for `invalid_argument` (makes T028's error-code assertion pass; contracts §2)
- [X] T035 [US3] Run `go build ./... && go vet ./... && go test ./...` and confirm green

**Checkpoint**: All three user stories independently functional — safe free text (US1), BM25-driven measurable ranking (US2), and explicit advanced-syntax mode (US3).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Whole-feature validation and documentation, spanning all three user stories.

- [X] T036 Manually run every scenario in `specs/036-text-search-ranking/quickstart.md` against a built binary (`go build -o /tmp/ms-036-bin ./cmd/...`) in a disposable temp project fixture, then clean up
- [X] T037 [P] Update `site/commands.html` with the new `--query-mode` and `--diagnostic-scores` flags, the new `query_syntax_error` code, the new `diagnostics.ranking_version` field, and the `schema_version: 3` bump (matching the pattern established for specs 031–035)
- [X] T038 [P] Record this feature's `ranking_version = 1` formula as the evaluation baseline referenced by `specs/019-dogfooding-evaluation/contracts/evaluation-protocol.md`'s comparative protocol, so any future weight change (a `ranking_version` bump) has a documented "before" to compare against (spec FR-010, SC-005)
- [X] T039 Run the whole-repo gate one final time: `go build ./... && go vet ./... && go test ./...`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS User Story 2 and User Story 3 (both append fields to the same JSON envelope in `internal/cli/internalcmd/context.go`; the version bump is reserved once here per Constitution Principle IX). User Story 1 does not depend on Phase 2 and may start in parallel with it.
- **User Story 1 (Phase 3)**: Depends on Setup only — may run in parallel with Phase 2.
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2).
- **User Story 3 (Phase 5)**: Depends on Foundational (Phase 2). Shares `internal/context/index/search.go` and `internal/context/request.go`/`collector.go` with US1/US2 respectively — implement after US1 (so `freeTextMatchExpr` already exists as the "what advanced mode is *not*" baseline) to avoid rebasing, though its own tests/behavior are independent.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### Within Each User Story

- Tests are written first and confirmed to fail before implementation (Constitution Principle V).
- Implementation tasks within a story generally touch the same one or two files sequentially; tasks marked `[P]` are on genuinely different files or independent test functions.

### Parallel Opportunities

- T002 (Foundational test) has no sibling to parallelize with in that phase.
- T005–T008 (US1 tests, same file but independent test functions with no shared state) can be written in parallel by different contributors, though they land in one file.
- T012–T017 (US2 tests, different files: `result_test.go`, `collector_test.go`, `rank_test.go` x4, `context_test.go`) can run in parallel.
- T025–T028 (US3 tests, same/adjacent files, independent functions) can run in parallel.
- T037/T038 (Polish) can run in parallel.
- Once Phase 2 completes, User Story 2 and User Story 3 can proceed in parallel with each other (and User Story 1, if not already done) if staffed.

---

## Parallel Example: User Story 2 tests

```bash
# Launch all US2 tests together (different files, no shared state):
Task: "Write failing test in internal/context/result_test.go for Candidate.TextRank + mergeAndSort propagation"
Task: "Write failing test in internal/context/collector_test.go for Tier 4 TextRank population"
Task: "Write failing test in internal/context/rank_test.go for BM25-derived scoring ordering"
Task: "Write failing test in internal/context/rank_test.go for tier-first invariant under strong BM25"
Task: "Write failing test in internal/context/rank_test.go for stable tie-break on equal scores"
Task: "Write failing test in internal/cli/internalcmd/context_test.go for --diagnostic-scores opt-in score_components"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 3: User Story 1 (Phase 2 Foundational is not required for US1 — it exists only for US2/US3's shared envelope constants)
3. **STOP and VALIDATE**: run `go test ./internal/context/index/...` and confirm punctuation-bearing free-text queries never error
4. Demo if ready — this alone already resolves the reliability problem PROP-06 exists to fix

### Incremental Delivery

1. Setup → baseline confirmed
2. Add User Story 1 → free-text safety lands (MVP), independent of Foundational
3. Foundational → shared `schema_version`/`ranking_version` constants reserved
4. Add User Story 2 → BM25-driven measurable ranking + opt-in diagnostics land
5. Add User Story 3 → explicit advanced-syntax mode lands
6. Each story adds value without breaking the previous one — `Search`'s signature never changes (US1 fixes it in place); `SearchAdvanced` (US3) is additive; `score_components`/`ranking_version` (US2) are additive and opt-in

### Parallel Team Strategy

With multiple developers:

1. One developer starts User Story 1 immediately (no Foundational dependency)
2. Another completes Phase 2 Foundational
3. Once Foundational is done: two more developers take User Story 2 and User Story 3 in parallel
4. Stories complete and integrate independently

---

## Notes

- `[P]` tasks = different files (or independent test functions with no shared state), no dependencies
- `[Story]` label maps a task to US1/US2/US3 for traceability
- Verify each test fails before implementing (Constitution Principle V, NON-NEGOTIABLE)
- Run `go build ./... && go vet ./... && go test ./...` after every phase, not just at the end
- Avoid: vague tasks, same-file conflicts within a story, cross-story dependencies that break independence
- `Search`'s exported signature (`Search(query string, limit int) ([]SearchResult, error)`) is deliberately left unchanged by US1 — every existing caller (`internal/context/collector.go`, `internal/example/sqlite_index_quickstart_test.go`, `internal/context/index/sync_test.go`) keeps compiling unmodified; only its internal query-construction behavior changes
