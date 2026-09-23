# Contract: `internal context`'s query-mode flag, BM25-based ranking, and `schema_version: 3`

This documents the machine-readable contract this feature adds/changes (Constitution Principle IX).

## 1. New flag: `internal context ... --query-mode <free|advanced>`

**Args**: `--query-mode` (optional, string, any output mode) — how `--query` (or `--task`, when `--query` is empty) is interpreted before reaching the index. Omitted → `free` (default, backward-compatible: existing callers see safer behavior for the same input, never a behavior change for queries that already worked).

| Value | Behavior |
|---|---|
| `free` (default) | Every token in the query text is treated as literal content; FTS5 reserved characters/keywords never cause a syntax error (spec FR-001, FR-003). |
| `advanced` | The query text is passed to the underlying FTS5 query grammar unmodified — phrase, prefix, and boolean operators are honored (spec FR-002). |

| Condition | Code | Exit |
|---|---|---|
| `--query-mode` given a value other than `free`/`advanced` | `invalid_argument` (existing sentinel) | 2 |
| `--query-mode=advanced` with a syntactically malformed expression | `query_syntax_error` (new sentinel, see §2) | 2 |

## 2. New error code: `query_syntax_error`

Returned only when `--query-mode=advanced` is used and the underlying engine rejects the expression as malformed (spec FR-004). Never returned in `free` mode — free-text mode's own escaping (research.md #1) guarantees no query-syntax failure reaches this point.

```json
{
  "ok": false,
  "error": {
    "code": "query_syntax_error",
    "message": "advanced query syntax error: unbalanced quote near position 12"
  }
}
```

A consumer MUST treat this as distinct from `internal_error`/index-failure codes — it signals a caller-supplied malformed expression, not an index/storage problem.

## 3. `schema_version: 3` — response shape change

Every `internal context` JSON response (all output modes) now reports `"schema_version": 3` instead of `2`. The change: same-tier free-text ordering is now driven by the real BM25 relevance value instead of a recomputed term-occurrence heuristic (spec FR-005), and two new fields appear.

### 3.1 `ranking_version` — new top-level diagnostics field

| Field | Type | Meaning |
|---|---|---|
| `diagnostics.ranking_version` | int | Identifies the scoring formula/weight set that produced this response's ordering (spec FR-009). Starts at `1` for this feature's shipped BM25-based formula. A future weight change bumps this value only after evaluation evidence is recorded per the existing protocol (`specs/019-dogfooding-evaluation/contracts/evaluation-protocol.md`; spec FR-010, SC-005) — never as a silent default-behavior change under the same version number. |

### 3.2 `--diagnostic-scores` — new opt-in flag, per-item score components

**Args**: `--diagnostic-scores` (optional boolean flag, any output mode). Omitted (default) → response shape identical to today aside from `schema_version` and `diagnostics.ranking_version`; no per-item score breakdown is added to the compact/default output (spec FR-008, research.md #5).

When passed, each item in `items` (or `context.items`, per the existing mode's own shape) gains a `score_components` object:

```json
{
  "path": "specs/036-text-search-ranking/spec.md",
  "heading": "Free-text queries never fail on punctuation",
  "score_components": {
    "tier": 4,
    "relation_weight": 0,
    "intent_bonus": 0,
    "text_relevance": 27,
    "total": 27
  }
}
```

| Field | Type | Meaning |
|---|---|---|
| `score_components.tier` | int | The item's assigned priority tier (unchanged tier semantics from 016-ranking-budgeting). |
| `score_components.relation_weight` | int | Existing `relationWeight()` contribution — unchanged computation. |
| `score_components.intent_bonus` | int | Existing `intentWeight()` contribution — unchanged computation. |
| `score_components.text_relevance` | int | The BM25-derived contribution (spec FR-005) — replaces the previous term-occurrence-based value; present (possibly `0`) even for items with no `text_match` reason. |
| `score_components.total` | int | Sum of the three components above; equals the item's `Score` used for same-tier ordering. |

`score_components` never appears when `--diagnostic-scores` is omitted — this is additive-only, opt-in output, consistent with 033's "avoid duplicating detail in the default output" decision.

## 4. Tie-break — unchanged

`items` ordering remains: ascending tier, then descending `Score` (now BM25-derived for `text_match` items), then ascending `Path`, then ascending `StartLine` (spec FR-006, FR-007; 016-ranking-budgeting's own convention, unmodified by this feature).

## 5. Example: free-text query safe by default

```text
$ misterspec internal context --target SPEC-014 --query 'retry-policy: "at most once" (idempotent)'
```

**Before this feature**: this query could fail with an FTS5 syntax error (unbalanced/ambiguous punctuation reaching `MATCH` unescaped).
**After this feature**: the query succeeds in default (`free`) mode — every token, including the hyphen, colon, quotes, and parentheses, is treated as literal text.

## 6. Example: advanced mode for a deliberate phrase search

```text
$ misterspec internal context --target SPEC-014 --query '"exact phrase" NEAR/5 retry' --query-mode advanced
```

The query is passed to FTS5 unmodified, honoring `NEAR` and the quoted phrase as native FTS5 syntax — only available when `--query-mode advanced` is explicit (spec Assumptions: never inferred from query content).
