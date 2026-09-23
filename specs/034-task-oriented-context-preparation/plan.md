# Implementation Plan: Recuperação orientada à tarefa e preparação de execução

**Branch**: `034-task-oriented-context-preparation` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/034-task-oriented-context-preparation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today, preparing to execute one Task means `mister-implement`'s Skill manually chaining `internal resolve` → `internal inspect` → `internal context` and then reading the Task's own free prose to decide "if its own dependencies are not all complete, stop" (`kit/skills/mister-implement/SKILL.md:143-147`) — no deterministic operation backs any of that reasoning. Worse, `internal/context/collector.go:27-29` explicitly rejects a Task as a context target today (`ArtifactType.HasEntityID()` is false for Task), so a Task-specific Context Pack cannot even be requested through the existing pipeline. This plan adds: a formal, parseable syntax for a Task's own `Depends on:`/`Verify:`/`Scope:` lines (mirroring 032's `Serves:` precedent); Task-to-Task readiness/blocking and cycle detection, reusing 032's already domain-agnostic `DependencyGraph`/`detectCycles` at Task granularity; and a new `internal prepare SPEC-### [--task TASK-###]` operation in a new `internal/prepare` package that assembles Task + deterministic context (served requirements' own text, Plan sections sharing those same requirements, scope, verification method) + readiness in one read-only response — without needing `Collect` to ever accept a Task target, since this context is a dedicated, purpose-built extraction, not a wrapper around the general wikilink/backlink pipeline.

## Technical Context

**Language/Version**: Go 1.23.4 (existing `go.mod`)
**Primary Dependencies**: existing internal packages only — `internal/validation` (new Task-body field parsing + Task-level dependency graph, reusing its existing `parseTaskCoverage`/`ParseDocument` patterns and its already-generic `DependencyGraph`/`detectCycles`), `internal/operations` (`ResolveTask`, `Inspect`), `internal/context` (`Fingerprint`, reused for consistency with 033's per-item shape), `internal/ids` (`ScanTasks`, `EntityID`), `internal/cli/internalcmd`. No new external dependency.
**Storage**: N/A — every result (readiness, context, blockers) is recomputed from the filesystem on each call (Constitution Principle III); nothing new is persisted.
**Testing**: Go `testing` — unit tests for the new Task-body field parsers and Task-level cycle detection (pure functions, mirroring `internal/validation`'s existing test conventions), and CLI-level tests for `internal prepare` (mirroring `internal/cli/internalcmd`'s existing conventions).
**Target Platform**: `misterspec` CLI binary, via a new `internal prepare` subcommand.
**Project Type**: Single Go module — one new package (`internal/prepare`), additive changes to `internal/validation` (new file) and `internal/cli/internalcmd` (new file + registration).
**Performance Goals**: No regression to existing commands; `prepare` performs one bounded read pass per Spec's `tasks.md`/`spec.md`/`plan.md`, no different in order of magnitude from `internal validate`'s existing per-Spec work.
**Constraints**: MUST NOT change `internal/context/collector.go`'s existing five-entity-type restriction (Constitution IV — no speculative widening of an established contract without a demonstrated need this feature doesn't have, per research.md Decision 3); MUST NOT introduce a new persisted state (Principle III); MUST NOT execute any command/script/test as part of `prepare` itself (spec FR-002); new Task-body fields (`Depends on:`, `Verify:`, `Scope:`) MUST default to "absent = no info" for Tasks authored before this feature exists (spec FR-003's compatibility requirement, mirroring 032/033's own zero-migration precedent).
**Scale/Scope**: Touches `internal/validation/task_dependencies.go` (new), `internal/validation/findings.go` (two new Finding codes), `internal/prepare/*.go` (new package: field parsing for `Verify`/`Scope`, readiness computation, plan-section association, response assembly), `internal/cli/internalcmd/prepare.go` (new command), `internal/cli/internal.go` (registration).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Semantic/Deterministic Separation** — PASS. Every new behavior (field parsing, readiness computation, cycle detection, requirement/plan-section text lookup) is purely mechanical; judging whether a Task is *actually* done, or whether its scope/verification text is *correct*, stays with the agent (spec FR-002's "no execution" boundary keeps this feature strictly assembly, not judgment).
- **II. Deterministic Operations Are the Only Mutation Primitive** — PASS. `prepare` is read-only; it allocates no ID and mutates nothing.
- **III. Filesystem Is the Single Source of Truth** — PASS. Readiness and context are recomputed from `tasks.md`/`spec.md`/`plan.md` on every call (spec FR-013); nothing new is persisted.
- **IV. Simplicity First — YAGNI & Minimal Configuration** — Applies directly to the one real architectural decision this plan makes: whether `prepare`'s orchestration deserves its own package. research.md Decision 5 justifies `internal/prepare` the same way `internal/context` earned its own package in an earlier feature — a materially new, sizable capability, not a tweak to an existing one. No new config field is added — the new Task-body labels are a fixed convention, not a per-project setting.
- **V. Test-First Discipline** — Applies. New parsing (`Depends on:`/`Verify:`/`Scope:`), Task-level cycle detection, readiness computation, and the `prepare` CLI surface MUST get unit/integration tests before/alongside implementation, per existing `internal/validation` and `internal/cli/internalcmd` conventions.
- **VI. Clean Code & SOLID** — PASS. `internal/validation` keeps owning "is this structurally correct" (Task-dependency cycles and invalid references become new Findings, exactly like 032's Spec-level cycles); `internal/prepare` owns the new "assemble everything needed to start this Task" responsibility; `internal/context` is untouched. No package's existing responsibility blurs into another's.
- **VII. Explicit Mutation Boundaries** — N/A. `prepare` is read-only, like every other `internal` command except `create`/`create-artifact`.
- **VIII. Safety by Construction** — N/A. No new filesystem write path.
- **IX. Transparent, Machine-Readable Contracts** — Applies. A blocked Task is reported as a structured, stable outcome (`ready: false` + named `blockers`), never an ad hoc error string; a Task-dependency cycle is a new stable Finding code in `internal validate`, consistent with 032's own precedent for Spec-level cycles.

No violations requiring justification; Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/034-task-oriented-context-preparation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── validation/
│   ├── task_dependencies.go     # NEW: Depends on: parsing, Task-level
│   │                              # DependencyGraph (reuses the existing
│   │                              # generic detectCycles), new Finding
│   │                              # codes (task_dependency_cycle,
│   │                              # invalid_task_dependency)
│   ├── task_dependencies_test.go # NEW
│   ├── findings.go               # + 2 new Code* constants
│   └── validator.go              # wires the new checks into
│                                   # ValidateProject/ValidateEntity
├── prepare/                      # NEW package
│   ├── fields.go                  # Verify:/Scope: line parsing
│   ├── readiness.go               # TaskReadiness computation from
│   │                                # task_dependencies.go's data
│   ├── context.go                 # requirement-text lookup, Plan-section
│   │                                # association, response assembly
│   ├── selection.go                # auto-selection (lowest-numbered
│   │                                # Ready Task, stable tie-break)
│   ├── doc.go
│   └── *_test.go                   # NEW, per file above
└── cli/internalcmd/
    ├── prepare.go                  # `internal prepare SPEC-### [--task
    │                                # TASK-###]`
    └── prepare_test.go             # NEW
```

**Structure Decision**: One new package (`internal/prepare`) for the genuinely new "assemble everything needed to start a Task" capability (research.md Decision 5), plus one new file in the existing `internal/validation` package for Task-level dependency parsing/validation (the same package that already owns Spec-level `depends_on` validation and Task-body field parsing via `Serves:`). No other package's boundary changes.

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
