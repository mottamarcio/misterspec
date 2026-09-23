# Phase 1 Data Model: Robust Text Search and Measurable Ranking

This feature adds no new persistent entity — the FTS5 index schema (`internal/context/index/schema.go`, `schemaVersion = 1`) is unchanged. What changes is the shape of values flowing between `index.Search`, `contextengine.Rank`, and the `context` CLI envelope.

## QueryMode

An in-memory enum, not stored, describing how a caller's query text should be interpreted before reaching `Store.Search`.

| Value | Meaning |
|---|---|
| `free` (default) | Every token is treated as literal text; FTS5 reserved characters/keywords have no special meaning. |
| `advanced` | The caller's raw string is passed to FTS5 `MATCH` unmodified, honoring FTS5's own query grammar (phrase, prefix, boolean operators, `NEAR`). |

**Validation rules**:
- Only these two values are valid; an unrecognized mode is a caller error (CLI argument validation), not a search-time error.
- Mode is always explicit at the call boundary — never inferred from the query string's content (research.md #2).

## SearchResult (extended, not restructured)

`internal/context/index/store.go`'s existing `SearchResult` struct is unchanged in shape — it already carries `Rank float64` (FTS5's `bm25()` value). This feature changes who *reads* `Rank` (previously unused past the `Search` call; now consumed by ranking) — no field is added or removed here.

| Field | Type | Notes |
|---|---|---|
| `Path` | string | unchanged |
| `Heading` | string | unchanged |
| `Content` | string | unchanged |
| `StartLine` | int | unchanged |
| `EndLine` | int | unchanged |
| `Rank` | float64 | unchanged shape; now propagated into `Candidate`/`Score` instead of being discarded after `Search` returns |

## Candidate (extended)

`internal/context/collector.go`'s `Candidate` struct (Tier 4 branch) gains a field carrying the originating `SearchResult.Rank`, so it survives past the `Search` call into `rank.go`.

| Field | Type | Notes |
|---|---|---|
| *(existing fields: Path, Heading, Content, StartLine, EndLine, Reasons)* | | unchanged |
| `TextRank` (new) | float64, pointer-or-zero-value convention TBD in implementation | Populated only for candidates originating from Tier 4 (`text_match` reason); zero/absent for candidates from other tiers, which never use this value in scoring. |

**Validation rules**: `TextRank` is meaningful only when a candidate carries a `text_match` Reason (Tier 4) — `scoreCandidate` MUST NOT read it for candidates whose Reasons don't include `text_match`, consistent with research.md #4's tier-scoped contract.

## ScoreComponents (new — diagnostic-only)

A new, diagnostic-only view of what produced one candidate's final `Score` — never part of the default compact manifest (research.md #5).

| Field | Type | Notes |
|---|---|---|
| `Tier` | int | the candidate's `minTier(Reasons)` (016's own concept) |
| `RelationWeight` | int | `relationWeight()`'s existing contribution |
| `IntentBonus` | int | `intentWeight()`'s existing contribution |
| `TextRelevance` | int | the BM25-derived contribution replacing today's `textRelevance()` term-count value |
| `Total` | int | sum, equal to `ScoredCandidate.Score` |

**Relationships**: One `ScoreComponents` per `ScoredCandidate`, computed alongside `Score` in `scoreCandidate`, discarded unless diagnostic output is requested.

## RankingVersion (new — constant, surfaced in output)

Not a stored entity — a build-time constant (parallel to `contextSchemaVersion`) identifying the scoring formula/weight set currently in effect.

| Field | Type | Notes |
|---|---|---|
| value | int, starting at `1` for this feature's shipped formula | Bumped whenever weights or the formula change (FR-009); a bump requires prior evaluation evidence per research.md #6 before becoming default (FR-010). |

## State / Lifecycle

No new state machine. Query mode is a per-call parameter (no persistence). `ranking_version` changes only through a deliberate, evidence-gated code change (not a runtime toggle) — consistent with Constitution Principle III (no authoritative secondary state) and Principle IV (no speculative configuration surface).
