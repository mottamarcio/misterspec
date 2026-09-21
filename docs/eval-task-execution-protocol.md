# Agent Task-Execution Evaluation Protocol

This documents how to run one `EvaluationTask` (037-eval-quality-
efficiency data-model.md "EvaluationTask") through a live coding-agent
session and record the result as a `RunRecord` the harness's `internal
eval-compare` command can then compare against a baseline.

There is deliberately no `misterspec` command that performs this
protocol automatically — the binary never launches, drives, or scrapes
a live LLM/agent session (Constitution Principle IV; spec037
research.md #1). This document, plus the shared `RunRecord`/`Metrics`
JSON shape `internal/eval` already validates, is the whole mechanism.

## Steps

1. **Pick a task.** Load one `EvaluationTask` from a task directory
   (e.g. `specs/037-eval-quality-efficiency/fixture/tasks/`) — its
   `description` is what you give the agent; its `dir` is the
   repository state to start from; its `acceptance_test` is the only
   thing that decides correctness.

2. **Run it live.** In a real coding-agent session (Claude Code,
   Antigravity, or any other agent you can exercise directly), give
   the agent exactly the task's own `description`, working inside its
   `dir`. Note the session's own reported effort figures as it
   finishes: input/output/cached tokens (if the session surfaces
   them), number of tool/LLM calls, any additional file reads beyond
   what a Context Pack request supplied, any repeated/corrective
   ("rework") actions you observe, and wall-clock elapsed time.

3. **Run the acceptance test.** Execute the task's own
   `acceptance_test` command exactly as written, in the same `dir`.
   Its exit code — not the agent's own narrative — decides the
   `TaskResult.Outcome`:
   - exit `0` → `"pass"`
   - non-zero, and the failure is clearly attributable to the task
     itself → `"fail"`
   - the run was aborted or errored for a reason unrelated to
     MisterSpec (e.g. an unrelated environment failure) → `"inconclusive"`
     (spec Edge Cases — recorded, never miscounted as pass or fail)

4. **Hand-author the `RunRecord`.** Following `internal/eval`'s
   `RunRecord`/`TaskResult`/`Metrics` shape (data-model.md), write a
   JSON file:

   ```json
   {
     "run_id": "task-exec-2026-09-25T10:00:00Z",
     "kind": "task_execution",
     "created_at": "2026-09-25T10:00:00Z",
     "config": {
       "repo_revision": "<git rev-parse --short HEAD>",
       "model": "<the agent/model used>",
       "variant": "default"
     },
     "results": [
       {
         "task_id": "add-forbidden-case-01",
         "outcome": "pass",
         "metrics": {
           "input_tokens": 12000,
           "input_tokens_estimated": false,
           "output_tokens": 3000,
           "output_tokens_estimated": false,
           "calls": 14,
           "extra_reads": 1,
           "rework": 0,
           "latency_seconds": 95.4
         }
       }
     ]
   }
   ```

   Every numeric effort figure you include MUST carry its own sibling
   `*_estimated` boolean (spec FR-008, SC-004) — this is a structural
   requirement `internal/eval`'s `Metrics.Validate` enforces, not a
   style preference. If your agent session does not expose token
   telemetry at all, omit those fields entirely rather than guessing,
   or set them with `*_estimated: true` and note the estimation method
   you used.

5. **Save it.** Write the file under `eval/runs/<run_id>.json` (the
   project-level directory `internal eval-compare` reads from — plan.md
   Research #3). To make it a comparison baseline, also write an
   `eval/baselines/<name>.json`:

   ```json
   {
     "name": "task-exec-2026-09",
     "run_ids": ["task-exec-2026-09-25T10:00:00Z"],
     "recorded_at": "2026-09-25T10:05:00Z"
   }
   ```

6. **Compare.** `misterspec internal eval-compare --baseline <name>
   --candidate eval/runs/<later-run>.json` works identically for
   task-execution `RunRecord`s as it does for retrieval ones — see
   `specs/037-eval-quality-efficiency/quickstart.md` §3-5 and
   `contracts/eval-commands-contract.md` §2.

## Why this is manual, not automated

Automating step 2 would require the `misterspec` binary to reach into
another program's own process (a different coding agent's CLI/IDE) —
architecture this project has never taken on, and doing so would
violate both the Constitution's Principle IV (no new machine command
for an external-orchestration decision) and the "no network in core
commands" architecture constraint every command except `--update`
holds to. Spec 019's own dogfooding evaluation already proved this
manual pattern works in practice: the assistant directly observed and
recorded its own Skill invocation behavior without any new tooling.
