# Implementation Plan: Validação de cobertura de requisitos e dependências do SDD

**Branch**: `032-requirement-coverage-dependency-validation` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/032-requirement-coverage-dependency-validation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

`kit/skills/mister-tasks/SKILL.md` and `mister-analyze/SKILL.md` already document a `R#` requirement-ID convention (`kit/templates/spec.md.tmpl`'s `### R1 — <Requirement>`) and a `Serves: SPEC-###:R#` Task-coverage convention, and `mister-tasks/SKILL.md:135-136` explicitly promises that `internal validate SPEC-###` "confirms no Task referencing a nonexistent requirement" — but `internal/validation` never parses Task or Spec body content at all today; it only checks frontmatter, Task-heading duplicates (031-canonical-task-identity), and that each `depends_on` entry resolves to an existing Spec (no graph, no cycle detection). This plan closes that gap: parse `R#` headings out of a Spec's own body and `Serves:` lines out of a Task's own body (reusing `artifacts.ParseDocument`, the existing heading-chunking utility from 013-document-model-chunking, rather than a new ad hoc scanner), build a project-wide Spec dependency graph and detect cycles, and gate coverage completeness on the Spec's own lifecycle `status`.

## Technical Context

**Language/Version**: Go 1.23.4 (existing `go.mod`)
**Primary Dependencies**: existing internal packages only — `internal/artifacts` (`ReadBody`, `ParseDocument`/`Section` for heading-bounded body chunks), `internal/ids`, `internal/validation`, `internal/project`. No new external dependency.
**Storage**: N/A — filesystem is the sole source of truth (Constitution Principle III); the dependency graph and coverage sets are recomputed from `spec.md`/`tasks.md` content on every validation run, never persisted.
**Testing**: Go `testing` — unit tests for pure parsing (requirement headings, `Serves:` line grammar, cycle detection over an in-memory graph) and filesystem-integration tests via `testutil.Project`, matching `internal/validation`'s existing test conventions.
**Target Platform**: `misterspec` CLI binary, via `internal validate` (project-wide and single-Spec).
**Project Type**: Single Go module — extends the existing `internal/validation` package; no new package layer.
**Performance Goals**: No regression vs. today's single-pass-per-entity-type validation; the dependency graph is built once per `ValidateProject`/`ValidateEntity` call, not once per Spec.
**Constraints**: MUST NOT introduce a new persisted state (Constitution Principle III); MUST NOT add a new package layer without demonstrated need (Principle IV) — this stays inside `internal/validation`, `internal/artifacts`'s already-general `ParseDocument`, and `internal/ids`'s existing `EntityID`; MUST NOT let a Plan's own prose "Requirement Coverage" section become a second, redundant mechanical check — the enforceable signal is the Task's own `Serves:` line (research.md Decision 4 records this scoping choice explicitly).
**Scale/Scope**: Touches `internal/validation/validator.go`, plus new files `internal/validation/requirements.go` (requirement/coverage parsing) and `internal/validation/dependency_graph.go` (cycle detection); `internal/validation/findings.go` (new Finding codes); `internal/validation/tables.go` is read, not modified (phase-gate logic keys off its existing `allowedStatesByType[ids.Spec]`, `"draft"` vs. everything else).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Semantic/Deterministic Separation** — PASS. Every new check (requirement existence, coverage counting, cycle detection, phase gating) is purely mechanical string/graph analysis; judging whether a requirement is *semantically* satisfied stays with the agent (spec.md FR-015, an explicit non-goal).
- **II. Deterministic Operations Are the Only Mutation Primitive** — PASS. This feature is read-only validation; it adds no `create`/mutation path and allocates no new IDs (Requirement numbers are authored directly in Markdown by the agent, exactly like Task numbers already are — `mister-tasks/SKILL.md:113-116`).
- **III. Filesystem Is the Single Source of Truth** — PASS. The dependency graph and coverage sets are rebuilt from `spec.md`/`tasks.md` content on every call; nothing new is persisted.
- **IV. Simplicity First — YAGNI & Minimal Configuration** — PASS. No new package layer: requirement/coverage parsing and cycle detection are added to `internal/validation`, which already owns "structural correctness of the project." Reuses `artifacts.ParseDocument` (already general-purpose) instead of writing a second heading-and-body scanner. No new config field — the `R#`/`Serves:` grammar is a fixed convention, not a per-project setting.
- **V. Test-First Discipline** — Applies. New parsing (requirement headings, `Serves:` grammar, cycle detection) and new integration checks (`ValidateProject`/`ValidateEntity`) MUST get unit tests and filesystem-integration tests before/alongside implementation, per `internal/validation`'s existing `*_test.go` conventions.
- **VI. Clean Code & SOLID** — PASS. `internal/validation` keeps owning "is this structurally correct," `internal/artifacts` keeps owning "how is Markdown structured" (`ParseDocument` reused, not duplicated) — no responsibility blur between packages.
- **VII. Explicit Mutation Boundaries** — N/A. No artifact-owning Skill or operation gains new write access; this is read-only validation exactly like every other check in `internal/validation`.
- **VIII. Safety by Construction** — N/A. No new filesystem write path.
- **IX. Transparent, Machine-Readable Contracts** — Applies. Every new problem (duplicate requirement, uncovered requirement, unknown/cross-Spec reference, task without requirement, dependency cycle, phase-gate block) MUST be a `validation.Finding` with a stable `Code`, consistent with the existing `Finding` shape and code vocabulary in `internal/validation/findings.go` — never a new, parallel output shape.

No violations requiring justification; Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/032-requirement-coverage-dependency-validation/
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
├── artifacts/
│   ├── document.go        # ParseDocument/Section (existing, reused unchanged)
│   └── markdown.go        # ReadBody (existing, reused unchanged)
├── validation/
│   ├── requirements.go     # NEW: R# heading parsing (Spec body), Serves: line parsing
│   │                        # (Task body), per-Spec coverage-set computation
│   ├── dependency_graph.go # NEW: builds the project-wide Spec depends_on graph,
│   │                        # detects cycles (including self-reference), returns
│   │                        # full cycle paths
│   ├── validator.go        # ValidateProject/ValidateEntity — wires in the new
│   │                        # per-Spec coverage checks, the cycle check, and the
│   │                        # phase-gate rule (extends the existing Task-duplicate
│   │                        # loop pattern from 031-canonical-task-identity)
│   ├── findings.go         # New Finding Code* constants
│   ├── requirements_test.go        # NEW
│   ├── dependency_graph_test.go    # NEW
│   └── validator_test.go           # extended
└── ids/                    # unchanged — EntityID/Scan already provide everything
                             # this feature needs (Spec existence, DependsOn parsing
                             # already lives in artifacts.Metadata)
```

**Structure Decision**: No new top-level directory or package. This is a scoped addition inside the existing `internal/validation` package (two new files for two distinct responsibilities — requirement/coverage parsing vs. graph/cycle detection — each independently unit-testable, per Constitution Principle VI's SRP guidance), reusing `internal/artifacts.ParseDocument` rather than a new Markdown scanner.

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
