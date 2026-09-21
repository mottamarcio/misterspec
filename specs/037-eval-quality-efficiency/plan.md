# Implementation Plan: End-to-End Quality and Efficiency Evaluation

**Branch**: `037-eval-quality-efficiency` | **Date**: 2026-09-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/037-eval-quality-efficiency/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Spec 019's dogfooding evaluation was a one-off, manually-recorded exercise against a 9-Spec fixture with no repetitions, no cost accounting, and no reusable comparison mechanism — every future retrieval/ranking proposal (036's BM25 work, and PROP-06/08/13 in the backlog) would otherwise have to invent its own ad hoc measurement each time. The technical approach: (1) a new `internal/eval` package providing a deterministic retrieval-evaluation runner that loads versioned case files (required/forbidden content per query), calls the existing `internal/context` collector/ranker in-process (no shelling out), and reports per-case pass/fail plus rank position — exposed as a new `misterspec internal eval-retrieval` command following the same JSON-envelope contract as `context`/`prepare`; (2) a `RunRecord` file format that a maintainer (or an agent session under this project's own dogfooding practice) fills in after a live agent task-execution run, capturing success, token/cost figures per FR-007, and an explicit `estimated: true/false` flag per figure per FR-008 — the binary never launches or drives a live LLM session itself, consistent with Principle IV (no new machine command for a semantic/external-orchestration decision) and the "no network in core commands" architecture constraint; (3) a `misterspec internal eval-compare` command that diffs a candidate `RunRecord` (or a set of them) against a named `Baseline` file, both retrieval- and task-execution-shaped, flagging multi-dimension differences and surfacing per-case regressions separately from aggregates (FR-006, FR-011-FR-013).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: existing `internal/context` (`contextengine` package) and `internal/context/index` for retrieval evaluation; `github.com/spf13/cobra` v1.10.2 for the new `internal eval` command family. No new third-party dependency is introduced — case files and run records are read as YAML/JSON with the standard library plus the project's existing `gopkg.in/yaml.v3` usage (already a dependency via artifact frontmatter parsing).
**Storage**: filesystem only — evaluation case files, run records, and named baselines are Markdown/YAML/JSON files under the feature's fixture and a project-level `eval/` directory (no database, no vector store; Constitution Principle III).
**Testing**: `go test` — unit tests for case loading/validation and comparison-diff logic, filesystem integration tests for the retrieval runner against a temporary fixture repo (pattern from `internal/context/fixture_test.go`), golden-file tests for the JSON envelope and comparison report shape.
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows); retrieval evaluation runs with no network access. Agent task-execution runs happen inside whatever live coding-agent session the maintainer is already using (Claude Code, Antigravity) — external to this binary, orchestrated by a documented protocol, not automated by `misterspec` itself.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: retrieval evaluation over the initial case set completes in under 2 minutes (SC-001); no new performance target for agent task-execution runs, which are inherently bounded by live LLM session latency.
**Constraints**: deterministic, reproducible retrieval-evaluation output (Constitution Principle I, V, IX); no new persistent/authoritative state (Principle III — run records and baselines are plain files, not a database); the binary MUST NOT attempt to drive a live LLM session (Principle IV — that decision stays with the agent/maintainer, not a new internal command); JSON contract changes to `internal eval` commands follow the same `schema_version` discipline as `context` (033/035/036 precedent).
**Scale/Scope**: new `internal/eval` package (case/record/baseline types, retrieval runner, comparison logic) + two new `internal/cli/internalcmd` commands (`eval-retrieval`, `eval-compare`) wired into `internal/cli/internal.go`; a fixture corpus and initial case set under `specs/037-eval-quality-efficiency/fixture/`; no change to `internal/context`, `internal/operations`, or existing Skills beyond documentation.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Retrieval-case checking (is required content present, at what rank) is fully mechanical and belongs in the binary. Judging whether an *agent's* completed task is correct is delegated to each task's own independent acceptance test (also mechanical, supplied per task) — never to LLM self-report — so no semantic judgment is pulled into the binary and no mechanical judgment is left to the agent.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. This feature creates no entities, IDs, or canonical artifact paths; case files, run records, and baselines are plain evaluation data, not `misterspec` artifacts.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. Case sets, run records, and named baselines are ordinary files, fully reconstructable/diffable via Git; no new authoritative counter or database is introduced.
- **Principle IV (Simplicity First — YAGNI)**: PASS with an explicit boundary: the binary gets exactly two new commands (`eval-retrieval`, `eval-compare`), both mechanical. It deliberately does NOT gain a command that launches or drives a live agent session — that would be a "decision that requires semantic reasoning" (routing an LLM conversation) misplaced into the Go binary, which Principle IV explicitly forbids. Live agent runs stay a documented protocol (extending 019's own precedent), not new binary surface.
- **Principle V (Test-First Discipline)**: Applies — case loading, required/forbidden checking, and comparison-diff logic are exactly the deterministic logic this principle requires unit tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. `internal/eval` owns case/record/comparison logic; it depends on `contextengine`'s existing public surface (already used by `internal/cli/internalcmd/context.go`) rather than duplicating collector/ranker logic.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. `eval-retrieval` only reads the target repository/fixture and writes its own report to stdout (or an explicit `--out` file the caller names); it never rewrites `spec.md`/`plan.md`/other artifacts it evaluates.
- **Principle VIII (Safety by Construction)**: PASS. No mutating filesystem writes beyond an explicitly-named `--out` report file; no path traversal risk beyond what `--dir`/`--out` flags already require validating (same posture as existing `internal context --dir`).
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — both new commands default to structured JSON with a `schema_version`, and comparison output distinguishes case-level results from aggregates per FR-011, continuing the `ok`/`valid` result-shape convention.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/037-eval-quality-efficiency/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
├── fixture/              # Initial retrieval case set + fixture corpus
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── eval/
│   ├── case.go              # EvaluationCase type + case-file loading/validation
│   ├── case_test.go
│   ├── retrieval.go         # Runs cases against contextengine.Collect/Score; RetrievalResult
│   ├── retrieval_test.go
│   ├── record.go            # RunRecord, Baseline types + file (de)serialization
│   ├── record_test.go
│   ├── compare.go           # Comparison logic: per-case delta, dimension-diff flag, regression surfacing
│   ├── compare_test.go
│   └── doc.go
├── context/                  # Unchanged; internal/eval depends on its existing public API
└── cli/
    └── internalcmd/
        ├── eval_retrieval.go # `misterspec internal eval-retrieval` — cobra command, JSON envelope
        ├── eval_retrieval_test.go
        ├── eval_compare.go   # `misterspec internal eval-compare` — cobra command, JSON envelope
        └── eval_compare_test.go

specs/037-eval-quality-efficiency/fixture/   # Initial retrieval case set + minimal fixture corpus
eval/                                         # Project-level home for recorded RunRecords/Baselines
├── baselines/
└── runs/
```

**Structure Decision**: Single Go project, no new top-level directory beyond `internal/eval` (mirrors the existing `internal/context`, `internal/validation` package-per-capability pattern) and a project-level `eval/` directory for recorded run data — analogous to how `.misterspec/` holds project state, but for evaluation artifacts, which are development-time data rather than product artifacts and therefore live outside `ai/`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations — table intentionally omitted.
