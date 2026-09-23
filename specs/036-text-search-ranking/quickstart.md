# Quickstart: Validating Robust Text Search and Measurable Ranking

Prerequisites: a built `misterspec` binary from this branch, and an indexed project (any repo with `misterspec init` already run and its FTS5 index present under the project's `.misterspec` state).

## 1. Free-text queries no longer fail on punctuation (User Story 1)

```sh
misterspec internal context --target SPEC-014 --query 'retry-policy: "at most once" (idempotent)'
```

**Expected**: exits `0`, returns a normal JSON response (`"ok": true`), never a query-syntax error — even though the query contains a hyphen, colon, quotes, and parentheses.

Repeat with a plain-prose query and confirm results are consistent with prior behavior:

```sh
misterspec internal context --target SPEC-014 --query 'context pack budget'
```

## 2. Same-tier ordering reflects real relevance (User Story 2)

```sh
misterspec internal context --target SPEC-014 --query 'budget estimator hard limit' --diagnostic-scores
```

**Expected**: `items` includes `score_components` per item (§3.2 of the contract). Among items sharing the same `score_components.tier`, higher `text_relevance` values appear earlier in `items`, consistent with the tier-then-score-then-path-then-line ordering (contract §4).

Run the same command twice and diff the raw JSON output — confirm byte-identical ordering (SC-003).

## 3. Advanced-syntax mode is explicit and separate (User Story 3)

```sh
misterspec internal context --target SPEC-014 --query '"exact phrase" NEAR/5 retry' --query-mode advanced
```

**Expected**: `NEAR` and the quoted phrase are honored as FTS5 operators — this only happens with `--query-mode advanced` explicit; omitting the flag treats the same string as literal free text (no crash, no operator interpretation).

Then try a deliberately malformed advanced query:

```sh
misterspec internal context --target SPEC-014 --query '"unbalanced' --query-mode advanced
```

**Expected**: exits non-zero, `"error": {"code": "query_syntax_error", ...}` (contract §2) — not a generic/leaked engine error, and not a silent fallback to free-text interpretation.

## 4. Ranking version is identifiable (FR-009)

```sh
misterspec internal context --target SPEC-014 --query 'budget estimator' | jq '.diagnostics.ranking_version, .schema_version'
```

**Expected**: `ranking_version` is present and equals the value documented in `contracts/search-and-ranking-contract.md` §3.1 (starts at `1`); `schema_version` equals `3`.

## 5. Regression check — existing budget/estimator behavior (035) unaffected

```sh
misterspec internal context --target SPEC-014 --budget 2000 --hard-limit 3000
```

**Expected**: `budget_exceeded`/`overage`/`diagnostics.estimator`/`diagnostics.hard_limit` semantics from 035-context-budget-accuracy are unchanged; only `schema_version` (now `3`) and the new `diagnostics.ranking_version` field are additive.

See `data-model.md` for the full field-level shape of `QueryMode`, `SearchResult`, `Candidate`, `ScoreComponents`, and `RankingVersion`, and `contracts/search-and-ranking-contract.md` for the complete request/response contract.
