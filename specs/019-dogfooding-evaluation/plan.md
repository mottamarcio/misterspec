# Implementation Plan: Dogfooding and Evaluation

**Branch**: `019-dogfooding-evaluation` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/019-dogfooding-evaluation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

An evaluation exercise, not a code feature: build one small, real
misterspec fixture project encoding this repository's own actual
feature dependency graph (011-018, plus 006) as `ai/programs/PRG-001/
features/FEAT-001/specs/SPEC-{006,011..018}` (research.md #1); run
`misterspec internal context` against every one of those Specs and
check the returned pack against each feature's own real, already-known
dependencies (User Story 1); perform one live Skill invocation each
through Claude Code and Antigravity against that same fixture (User
Story 2/3, research.md #2); and record every finding plus a final,
evidence-gated Tuning Decision in one Dogfooding Report (User Story 4).
No change to the Context Engine (011-017) or Skill Integration (018)
is planned up front — a ranking/budgeting change is added to scope
only if the evidence gathered here actually calls for one
(research.md #6).

## Technical Context

**Language/Version**: Go 1.23.4 (existing, unmodified) — the already-built `misterspec` binary is the only tool this evaluation runs; no new Go code is planned.
**Primary Dependencies**: None new. Uses `misterspec internal context` (017) and the `agy`/`claude-code` adapters (018), all already shipped.
**Storage**: One new, committed fixture project (`specs/019-dogfooding-evaluation/fixture/`, its own `.misterspec/config.yaml` + `ai/...` tree) — real filesystem content, not a database; its own disposable `.misterspec/cache/context.db` is git-ignored like any other project's, per 014's own established convention.
**Testing**: This feature has no automated test suite of its own — it is a human-reviewed evaluation (spec.md's own Assumptions); its "verification" is the Dogfooding Report itself being internally consistent (every finding traceable to an actual `internal context` JSON response or a directly-observed live session) and 001-018's own full suites re-run unmodified as the standing regression gate (nothing here should ever break them, since nothing here modifies product code).
**Target Platform**: Cross-platform Go module, unchanged. Live sessions (User Story 2/3) run wherever Claude Code and Antigravity are available to the user.
**Project Type**: Evaluation/documentation effort layered on an existing single Go module CLI. New content lives entirely under `specs/019-dogfooding-evaluation/` (fixture + report); zero files change under `internal/`, `cmd/`, or `kit/` unless User Story 4 identifies a justified ranking change (research.md #6), in which case that change is scoped and executed as its own explicit follow-up, not pre-designed here.
**Performance Goals**: N/A — no new code path. Elapsed time for live Skill invocations is recorded as an observed data point (research.md #4), not measured against a target.
**Constraints**: Strictly read-only with respect to 011-018's own shipped behavior unless FR-007/FR-008's own evidence bar is met (spec.md FR-010). The fixture must encode each Spec's own *real* `depends_on` and *real* summary content — never invented data (research.md #1) — so every User Story 1 finding is checking against a genuine, verifiable answer.
**Scale/Scope**: One fixture project (~9 Specs, 2 Knowledge entries, 1 Learning, 1 Constitution excerpt — research.md #1) plus one report file (`specs/019-dogfooding-evaluation/report.md`, research.md #5). No source code files.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** This feature performs no semantic judgment in code — every judgment (was this pack sufficient? is this omission acceptable?) is a human review recorded in the Dogfooding Report, exactly the boundary the constitution draws. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** The fixture's own artifacts are hand-authored content mirroring real, already-known history, not IDs/paths/hashes an operation would need to allocate; `internal create` is used where creating the fixture's own entities mechanically, exactly as any project owner would. |
| III. Filesystem Is Single Source of Truth | **Pass, reinforced.** The fixture is itself a real filesystem project; its own disposable index is rebuildable from it at any time, the same guarantee 014 already established. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** No latency instrumentation, no index-size tracking, no pre-drafted ranking changes, no second fixture for live vs. static evaluation — every one of these was considered and rejected for lack of a demonstrated present need (research.md #2, #4, #6). |
| V. Test-First Discipline | **N/A, documented.** This feature produces no deterministic logic of its own to test-first; 001-018's own full suites remain the standing regression gate (Technical Context, Testing). |
| VI. Clean Code & SOLID | **Pass (N/A new code).** No new package or file under `internal/`. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes only inside its own fixture project and its own report file — it never touches 011-018's own shipped artifacts (Skill files, Go source) unless User Story 4 explicitly justifies a change, tracked as its own separate, evidence-cited edit. |
| VIII. Safety by Construction | **Pass (N/A new surface).** The fixture is a normal, `misterspec init`-shaped project; no new path-handling code is introduced. |
| IX. Transparent, Machine-Readable Contracts | **Pass (N/A this feature).** No new CLI command; every `internal context` call used for evidence already returns 017's own established JSON envelope, quoted directly into the report. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/019-dogfooding-evaluation/
├── plan.md                          # This file (/speckit-plan command output)
├── research.md                      # Phase 0 output (/speckit-plan command)
├── data-model.md                    # Phase 1 output (/speckit-plan command)
├── quickstart.md                    # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── evaluation-protocol.md       # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md              # /speckit-specify quality checklist
├── fixture/                         # Phase 2/implementation output — the real-history-encoding project (research.md #1)
│   ├── .misterspec/config.yaml
│   └── ai/
│       ├── memory/constitution.md
│       ├── knowledge/KNOW-001-*.md, KNOW-002-*.md
│       └── programs/PRG-001/features/FEAT-001/specs/SPEC-{006,011..018}/spec.md
├── report.md                        # Phase 2/implementation output — the Dogfooding Report (research.md #5)
└── tasks.md                         # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                # unchanged — no new dependency
├── internal/, cmd/, kit/           # unchanged unless User Story 4 justifies a specific, cited ranking change (research.md #6) — tracked as its own follow-up, not designed here
└── specs/019-dogfooding-evaluation/  # this feature's entire footprint (see above)
```

**Structure Decision**: Everything this feature produces lives under
its own `specs/019-dogfooding-evaluation/` directory — a fixture
project and a report — with zero planned changes to `internal/`,
`cmd/`, or `kit/`. This matches the feature's own nature (an
evaluation, not a capability) and keeps its evidence and conclusions
reviewable in one self-contained place.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
