# Contract: `internal eval-retrieval` and `internal eval-compare`

This documents the machine-readable contract this feature adds
(Constitution Principle IX). Both commands are new, hidden `internal`
subcommands (Architecture Constraints — not advertised in normal
`--help`), following the same `ok`/`valid`/`result`, `schema_version`,
and error-code conventions already established by `internal context`
(033/035/036).

## 1. `misterspec internal eval-retrieval`

Runs a set of `EvaluationCase` files (data-model.md) against a target
fixture/repository, entirely offline and deterministic (spec User
Story 1, FR-001).

**Args**:

| Flag | Type | Notes |
|---|---|---|
| `--cases` | string, required | Path to a directory of case YAML files (data-model.md `EvaluationCase`). |
| `--out` | string, optional | Path to write the full `RunRecord` JSON. Omitted → `RunRecord` is only embedded in the command's own stdout response, not persisted. |
| `--json` | boolean, default true | Structured output only; no non-JSON mode exists for this command (Principle IX — internal commands default to structured JSON). |

**Response** (`evalRetrievalSchemaVersion: 1`):

```json
{
  "ok": true,
  "schema_version": 1,
  "result": {
    "run": {
      "run_id": "retrieval-2026-09-21T10:00:00Z",
      "kind": "retrieval",
      "config": { "repo_revision": "3ccbb06", "variant": "default" },
      "results": [
        {
          "case_id": "free-text-punctuation-01",
          "passed": true,
          "missing": [],
          "unexpected": [],
          "positions": { "SPEC-036": 1 }
        }
      ]
    },
    "summary": { "total": 12, "passed": 11, "failed": 1 }
  }
}
```

| Condition | Code | Exit |
|---|---|---|
| `--cases` path does not exist or contains no case files | `invalid_argument` (existing sentinel) | 2 |
| A case file fails validation (data-model.md Validation Rules — e.g. both `query` and `task` set) | `invalid_case` (new sentinel) | 2 |
| A case's own `dir` does not resolve | `invalid_case` | 2 |
| Every case loads and runs (regardless of individual pass/fail) | — | `ok: true`, exit 0. Case-level failures are reported in `result`, not surfaced as a command failure — a failing case is a valid, successful measurement (Principle IX's `ok` vs `result` distinction). |

## 2. `misterspec internal eval-compare`

Compares a candidate `RunRecord` against a named `Baseline`
(spec User Story 3, FR-012/FR-013).

**Args**:

| Flag | Type | Notes |
|---|---|---|
| `--baseline` | string, required | Name of a `Baseline` file under `eval/baselines/`. |
| `--candidate` | string, required | Path to a candidate `RunRecord` JSON file (e.g. `--out` from `eval-retrieval`, or a hand-authored task-execution `RunRecord`). |
| `--repetitions` | string, optional | Path to a directory of additional same-config `RunRecord` files, used to compute `variance_note` (data-model.md, Research #5/#6). |

**Response** (`evalCompareSchemaVersion: 1`):

```json
{
  "ok": true,
  "schema_version": 1,
  "result": {
    "comparison": {
      "baseline_name": "retrieval-2026-09",
      "candidate_run_id": "retrieval-2026-09-25T09:00:00Z",
      "dimension_diff": [],
      "changed_cases": [
        {
          "id": "free-text-punctuation-01",
          "baseline_outcome": "fail",
          "candidate_outcome": "pass",
          "direction": "improved"
        }
      ],
      "aggregate": {
        "success_rate": 0.92,
        "total_tokens_per_correct_task": 14500
      }
    }
  }
}
```

| Condition | Code | Exit |
|---|---|---|
| Named baseline does not exist | `invalid_argument` | 2 |
| Candidate `RunRecord.kind` differs from baseline's `kind` | `incompatible_run` (new sentinel) | 2 |
| Baseline was recorded against a different `contextSchemaVersion`/`rankingVersion` than the candidate declares in its `config` | `stale_baseline` (new sentinel) — a warning surfaced in `result`, not a hard failure, per spec Edge Cases ("must surface that the baseline needs refreshing rather than produce a misleading delta") | 0, `ok: true`, `result.stale: true` |
| `--repetitions` supplied and its `RunRecord`s' `config` (minus `repetition_index`) does not match the candidate's | `invalid_argument` | 2 |

### `dimension_diff` semantics (FR-006)

Computed by comparing every field of `candidate.config` against
`baseline`'s own recorded `config` except `repetition_index`. Any
field that differs is listed by name. An empty `dimension_diff` means
the comparison is a valid single-variable isolation; a non-empty one
MUST be surfaced to the caller before `changed_cases`/`aggregate` are
trusted as isolating one cause (spec User Story 3, Acceptance Scenario
2).

### `variance_note` semantics (FR-013)

Present only when `--repetitions` is supplied. A changed case is
reported as `direction: "neutral"` instead of `"improved"`/`"regressed"`
when its outcome is inconsistent across the supplied repetitions
(i.e. the repetitions themselves already disagree on that case) — the
comparison MUST NOT claim a directional change the baseline's own
repeated runs don't support.

## 3. New error sentinels

| Code | Meaning |
|---|---|
| `invalid_case` | A case or task file failed structural validation (data-model.md Validation Rules). |
| `incompatible_run` | Baseline and candidate `RunRecord.kind` differ; a comparison across kinds is meaningless and MUST be refused. |
| `stale_baseline` | Baseline predates a contract/config dimension the candidate now reports differently; comparison still runs but is flagged, never silently trusted (spec Edge Cases). |
