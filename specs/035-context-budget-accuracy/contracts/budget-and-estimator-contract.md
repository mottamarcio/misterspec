# Contract: `internal context`'s new `--hard-limit` flag and budget diagnostics (`schema_version: 2`)

This documents the machine-readable contract this feature adds/changes (Constitution Principle IX).

## 1. New flag: `internal context ... --hard-limit N`

**Args**: `--hard-limit N` (optional, integer, any output mode) — the mandatory-content ceiling for this request. Independent of the existing `--budget N` flag (the soft, optional-content target — unchanged meaning). Omitted → resolves to `DefaultHardLimit` (`12000`, twice `DefaultBudget`). A `--hard-limit` smaller than the effective `--budget` is accepted, not rejected (data-model.md's Request validation rule).

| Condition | Code | Exit |
|---|---|---|
| `--hard-limit` given a non-integer or negative value | `invalid_argument` (existing sentinel) | 2 |

No new error sentinel is introduced.

## 2. `schema_version: 2` — response shape change

Every `internal context` JSON response (all three output modes) now reports `"schema_version": 2` instead of `1`. A consumer that branches on `schema_version` (033's own documented pattern) must treat `2` as carrying the semantic change below; a consumer that ignores `schema_version` and reads `budget_exceeded` under the old assumption will observe a behavior change, which is exactly what the version bump exists to signal.

### 2.1 `budget_exceeded` / `overage` — redefined, same field names

**Before (`schema_version: 1`)**: `budget_exceeded = mandatory_tokens > budget` (the soft budget).
**Now (`schema_version: 2`)**: `budget_exceeded = mandatory_tokens > hard_limit` (the new, separate hard limit — resolved value, always present, defaulting to `12000`).

```json
{
  "ok": true,
  "schema_version": 2,
  "diagnostics": {
    "candidates_considered": 9,
    "items_selected": 6,
    "tokens_available": 6000,
    "tokens_selected": 5820,
    "tokens_excluded": 1400,
    "reduction_percent": 19.4,
    "estimator": "default",
    "hard_limit": 12000,
    "exclusions": [
      {"path": "ai/.../SPEC-014/plan.md", "heading": "Alternatives Considered", "reason": "did_not_fit_remaining_budget"},
      {"path": "ai/.../SPEC-014/spec.md", "heading": "Edge Cases", "reason": "no_coherent_unit_fit"}
    ]
  },
  "budget_exceeded": false,
  "overage": 0
}
```

`budget_exceeded: true` never shrinks `items` — every mandatory candidate is always present in full (spec FR-007); only the flag and `overage` (now `mandatory_tokens - hard_limit`) change.

### 2.2 New diagnostics fields

| Field | Type | Meaning |
|---|---|---|
| `diagnostics.estimator` | string | The `Name()` of the `Estimator` that produced every token count in this response (spec FR-001). Today, always `"default"`. |
| `diagnostics.hard_limit` | int | The resolved hard limit actually applied — always present, even when `--hard-limit` was not given (spec FR-006). |
| `diagnostics.exclusions` | array of `{path, heading, reason}` | Every optional candidate, or partial-unit fallback, left out of the response and why (spec FR-010). `reason` is one of `"did_not_fit_remaining_budget"` or `"no_coherent_unit_fit"` — a closed, stable set; a consumer may safely match on exact string. |

### 2.3 Backfill — no new field, a behavior change to which candidates appear in `items`

An optional candidate that doesn't fit in the remaining soft-budget space no longer aborts the optional-content loop for every subsequent candidate (spec FR-008). The loop keeps trying later candidates, in the same tier-ascending/score-descending order `ranked` already established — a smaller, lower-scored, same-tier candidate may now appear in `items` even though an earlier, higher-scored candidate of the *same* tier was excluded. A higher-priority tier is still always fully decided before a lower-priority tier is considered (spec FR-009) — this ordering guarantee is unchanged and requires no new field to observe; it is simply true of `items`' own tier composition.

### 2.4 Coherent-unit fallback — surfaces only as `exclusions` + possibly-partial `items`

When a single optional candidate is too large to include whole but at least its first coherent unit fits, `items` may contain that candidate's own content truncated to a coherent-unit boundary (never mid-code-block, never mid-paragraph — spec FR-011) instead of the whole candidate being excluded outright. No new per-item field marks this; a consumer that needs to know can compare the returned content's own length against the candidate's original size, but this is not part of the tested contract — only the exclusion-vs-inclusion outcome and the coherence guarantee are.

## 3. Example: soft/hard limit decoupling in practice

```text
$ misterspec internal context --target SPEC-014 --budget 2000 --hard-limit 3000
```

A request whose mandatory content alone is 2500 tokens: `budget_exceeded: true`? No — `2500 <= 3000` (the hard limit), so `budget_exceeded: false`, even though mandatory content already exceeds the 2000-token soft budget. This is the exact false-alarm spec User Story 2 exists to eliminate (research.md Decision 3).
