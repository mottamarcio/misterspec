# Quickstart: Validating End-to-End Quality and Efficiency Evaluation

Prerequisites: a built `misterspec` binary from this branch; the
initial fixture case set under `specs/037-eval-quality-efficiency/
fixture/cases/` (User Story 1); this repository's own `eval/` directory
present at the project root (created on first `eval-compare` baseline
save, or manually).

## 1. Deterministic retrieval evaluation runs offline and reproducibly (User Story 1)

```sh
misterspec internal eval-retrieval \
  --cases specs/037-eval-quality-efficiency/fixture/cases \
  --out /tmp/retrieval-run.json
```

**Expected**: exits `0`, `"ok": true`; `result.summary` reports
`total`/`passed`/`failed` counts; every case in `result.run.results`
reports `passed`, `missing`, `unexpected`, and `positions` (contract
§1).

Run the same command again and diff the two `--out` files:

```sh
diff <(jq -S . /tmp/retrieval-run.json) <(misterspec internal eval-retrieval --cases specs/037-eval-quality-efficiency/fixture/cases --out /dev/stdout | jq -S .)
```

**Expected**: identical except `run_id`/`created_at` — confirms
SC-001's reproducibility requirement.

## 2. A missing required item is named, not just scored (User Story 1, Acceptance Scenario 2)

Edit one fixture case's `required` list to include an identifier that
is deliberately absent from its `dir`'s fixture project, then re-run:

```sh
misterspec internal eval-retrieval --cases specs/037-eval-quality-efficiency/fixture/cases
```

**Expected**: that case's `passed` is `false` and `missing` names the
specific absent identifier — never only a lowered aggregate score.

## 3. Record a baseline and compare a candidate run (User Story 3)

```sh
mkdir -p eval/baselines
cp /tmp/retrieval-run.json eval/baselines/retrieval-2026-09.json
# ... make a change to internal/context/rank.go, rebuild ...
misterspec internal eval-retrieval \
  --cases specs/037-eval-quality-efficiency/fixture/cases \
  --out /tmp/retrieval-candidate.json
misterspec internal eval-compare \
  --baseline retrieval-2026-09 \
  --candidate /tmp/retrieval-candidate.json
```

**Expected**: `result.comparison.changed_cases` lists every case whose
`passed`/`positions` differ from the baseline, each with an explicit
`direction` (contract §2). `result.comparison.dimension_diff` is empty
when only the ranking code changed (no config dimension moved).

## 4. A non-isolated comparison is flagged, not silently trusted (User Story 3, Acceptance Scenario 2)

Re-run the candidate command with a different `--variant` value baked
into its `config` (or against a different `repo_revision`), then
compare again:

**Expected**: `result.comparison.dimension_diff` names the differing
field(s) (e.g. `["variant"]`); the response still succeeds (`ok: true`)
but a caller reading `dimension_diff` knows the comparison does not
isolate a single cause.

## 5. Agent task-execution runs are recorded, not automated (User Story 2)

Following the protocol below (no new `misterspec` command performs
this step — Constitution Principle IV, research.md #1):

1. Pick an `EvaluationTask` from `specs/037-eval-quality-efficiency/
   fixture/tasks/`.
2. Run it to completion in a live coding-agent session (Claude Code or
   Antigravity), noting the session's own reported token/call/latency
   figures.
3. Run the task's `acceptance_test` command; record its exit code as
   `outcome`.
4. Hand-author a `RunRecord` JSON file (data-model.md shape) with the
   observed `Metrics`, setting every `*_estimated` flag explicitly.
5. Save it under `eval/runs/`, then compare it the same way as step 3
   above (`--candidate eval/runs/<file>.json`).

**Expected**: the resulting `ComparisonReport`'s `aggregate.success_rate`
and `aggregate.total_tokens_per_correct_task` are both present together
(never the ratio alone, per spec FR-009/SC constraints), and any
estimated figure is visibly labeled in the report.

See `data-model.md` for full entity shapes and `contracts/
eval-commands-contract.md` for the complete request/response contract
and error sentinels.
