# Phase 1 Data Model: End-to-End Quality and Efficiency Evaluation

Entities correspond to spec.md's Key Entities section. Field names
below are illustrative (the authoritative shape is `contracts/`); this
document defines relationships, ownership, and validation rules.

## EvaluationCase (retrieval)

One deterministic retrieval check (spec's "Evaluation Case").

| Field | Type | Notes |
|---|---|---|
| `id` | string | Unique within its case-file directory; stable across runs so comparisons can match cases by identity. |
| `target` | string | Required entity ID `contextengine.Collect` resolves context around (`Request.Target`, e.g. `SPEC-014`) — every case exercises a Context Pack request for one specific target, matching the existing `internal context <id>` contract this evaluation reuses. |
| `query` | string, optional | Free-text query mode (exercises 036's BM25 path). Mutually exclusive with `task`. |
| `task` \| `intent` | string, optional | Task-oriented mode (exercises 034's `internal prepare` path). Mutually exclusive with `query`. |
| `dir` | string | Path to the fixture project this case runs against — relative to the case-file's own directory, or absolute (used as-is). |
| `budget` | int, optional | Overrides the default budget for this case only (exercises budget-pressure edge cases, spec FR-010). |
| `required` | []string | Artifact/chunk identifiers that MUST appear in the returned pack for this case to pass. |
| `forbidden` | []string, optional | Identifiers that MUST NOT appear (over-inclusion check). |

**Validation rules**: `target` non-empty (required — `contextengine.
Request.Target` has no zero-value/"any" meaning); exactly one of
`query`/`task` set, or neither (a target-only case is valid — it
exercises Tier 0-3 structural/semantic collection with no free-text or
task-oriented Tier 4 behavior); `id` unique within the loaded set;
`dir` MUST resolve to an existing path; `required` non-empty unless the
case is explicitly documented (via a `note` field) as a negative/
empty-result case.

## EvaluationTask (agent task-execution)

One agent task-execution scenario (spec's "Evaluation Task"). Unlike
`EvaluationCase`, this is not run by the binary — it is a specification
a maintainer or live agent session follows, and the binary only
consumes the `RunRecord` produced afterward.

| Field | Type | Notes |
|---|---|---|
| `id` | string | Unique within its task-file directory. |
| `description` | string | The task given to the agent, in plain language. |
| `dir` | string | Fixture/repository state the task starts from. |
| `acceptance_test` | string | A command whose exit code determines correctness (Research #4). Never interpreted, only recorded and later re-run/checked by the maintainer's own session. |

## RunRecord

One execution of either evaluation set under a specific, recorded
configuration (spec's "Evaluation Run" + "Metric Record", merged into
one file per run for traceability).

| Field | Type | Notes |
|---|---|---|
| `run_id` | string | Stable identifier, typically `<kind>-<timestamp>`. |
| `kind` | `"retrieval"` \| `"task_execution"` | Which evaluation set produced this record. |
| `created_at` | string (RFC 3339) | When the run was recorded. |
| `config` | object | `{repo_revision, model, variant, repetition_index?}` — every dimension a comparison (FR-006) must check for equality. |
| `results` | []CaseResult or []TaskResult | Per-case/per-task outcomes; see below. |

### CaseResult (retrieval, nested in RunRecord.results)

| Field | Type | Notes |
|---|---|---|
| `case_id` | string | Matches an `EvaluationCase.id`. |
| `passed` | bool | True iff all `required` present and all `forbidden` absent. |
| `missing` | []string | Required identifiers not found (empty on pass). |
| `unexpected` | []string | Forbidden identifiers found (empty on pass). |
| `positions` | map[string]int | Rank position of each required identifier that was found, for later ranking-position comparisons (User Story 1, Acceptance Scenario 3). |
| `score` | float, optional | The item's own composite score (036's diagnostic score components), when available. |

### TaskResult (agent task-execution, nested in RunRecord.results)

| Field | Type | Notes |
|---|---|---|
| `task_id` | string | Matches an `EvaluationTask.id`. |
| `outcome` | `"pass"` \| `"fail"` \| `"inconclusive"` | `"inconclusive"` per spec Edge Cases — a run aborted for reasons unrelated to MisterSpec. |
| `metrics` | Metrics | See below. |

### Metrics

| Field | Type | Notes |
|---|---|---|
| `input_tokens`, `output_tokens`, `cached_tokens` | int, optional | Optional because not every provider exposes all three. |
| `input_tokens_estimated`, `output_tokens_estimated`, `cached_tokens_estimated` | bool | Research #6 — structurally paired with each figure. |
| `calls` | int | Number of agent-tool/LLM calls in the run. |
| `extra_reads` | int | Reads performed beyond what the Context Pack supplied (spec FR-007). |
| `rework` | int | Count of repeated/corrective actions observed. |
| `latency_seconds` | float | Wall-clock elapsed time. |
| `cost` | float, optional | Computed from published provider pricing (spec Assumptions); `cost_estimated` mirrors the token-level flags. |

## Baseline

A named, recorded `RunRecord` (or an aggregate over several repetition
`RunRecord`s sharing the same `config` except `repetition_index`)
designated as the comparison reference.

| Field | Type | Notes |
|---|---|---|
| `name` | string | e.g. `retrieval-2026-09`, `task-exec-036-baseline`. |
| `run_ids` | []string | One or more `RunRecord.run_id` values this baseline aggregates (repetitions of the same config). |
| `recorded_at` | string (RFC 3339) | |

## ComparisonReport

Output of `misterspec internal eval-compare` (spec's "Comparison
Report").

| Field | Type | Notes |
|---|---|---|
| `baseline_name` | string | |
| `candidate_run_id` | string | |
| `dimension_diff` | []string | Non-empty when candidate's `config` differs from baseline's in more than the one intended dimension (FR-006) — e.g. `["model"]` if the model also changed unexpectedly. |
| `changed_cases` | []ChangedCase | Every case/task whose outcome differs, per FR-012. |
| `aggregate` | object | `{success_rate, total_tokens_per_correct_task}` per FR-009 — always presented together, per spec SC (never the ratio alone). |
| `variance_note` | string, optional | Set when repetitions (Research #5) show a changed outcome is within normal run-to-run variance rather than a real effect (FR-013). |

### ChangedCase

| Field | Type | Notes |
|---|---|---|
| `id` | string | Case or task id. |
| `baseline_outcome` | string | |
| `candidate_outcome` | string | |
| `direction` | `"improved"` \| `"regressed"` \| `"neutral"` | |

## Relationships

```text
EvaluationCase ──run──▶ RunRecord(kind=retrieval) ──contains──▶ CaseResult
EvaluationTask ──run──▶ RunRecord(kind=task_execution) ──contains──▶ TaskResult ──has──▶ Metrics
RunRecord ──named as──▶ Baseline
(Baseline, RunRecord) ──compared by── ▶ ComparisonReport ──contains──▶ ChangedCase
```

## Validation Rules Summary

- A `RunRecord` MUST declare exactly one `kind`; its `results` entries
  MUST all match that kind's shape.
- A `ComparisonReport` MUST be producible only when baseline and
  candidate share the same `kind`.
- Every `Metrics` numeric field that is present MUST have its sibling
  `_estimated` boolean present and explicit (never implied by absence).
- `dimension_diff` MUST be computed before any other comparison output
  is produced, since a non-empty result changes how the rest of the
  report should be read (FR-006).
