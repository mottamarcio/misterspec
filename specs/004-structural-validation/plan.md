# Implementation Plan: Structural Validation and Project Status

**Branch**: `004-structural-validation` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-structural-validation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Complete the read-only Phase 2 operations begun in 002-read-operations:
`validate` (structural checks that collect every anomaly as data, never
stopping at the first or raising an error for a structural problem) and
`status` (a fast aggregate snapshot). One new package, `internal/validation`
(rule-checking, deliberately kept independent of `internal/operations` to
avoid an import cycle with the new `operations/status.go` — see
research.md), plus that one new file in the existing `operations` package.
No new external dependency, still no CLI/JSON surface, entirely read-only.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged)
**Primary Dependencies**: Go standard library only — no new external dependency.
**Storage**: Filesystem, read-only (Constitution Principle II/VII — this feature never writes).
**Testing**: `go test` — unit tests (lifecycle-state/parent-type tables, Finding construction) and filesystem-integration tests (fixture projects mixing valid and deliberately broken entities), per Constitution Principle V. 001/002/003's full suites remain mandatory regression gates.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module, library-first (still no `cmd/`).
**Performance Goals**: Whole-project validation and status are a single pass over `ids.Scan`'s results per type — no repeated full-tree walks per entity.
**Constraints**: `internal/validation` MUST NOT import `internal/operations` (research.md's import-cycle finding); every operation in this feature is read-only (FR-014); `Status.StructuralErrors` MUST be computed by calling `validation.ValidateProject`, never a separate, potentially-drifting count.
**Scale/Scope**: One new package (`internal/validation`, ~4 files) plus `internal/operations/status.go` — no CLI, no body-content parsing (duplicate requirement markers, Task→requirement references — explicitly deferred per spec.md Assumptions).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Every check is mechanical (ID syntax, fixed state tables, existence/type lookups) — no semantic judgment about whether an entity is "good," only whether it's structurally well-formed. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** This feature performs zero mutation — the entire point of "validate" and "status" is to observe, never change. |
| III. Filesystem Is Single Source of Truth | **Pass.** `ValidateProject`/`Status` recompute from `ids.Scan` on every call — no caching, no stored validation state. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass.** Lifecycle-state and parent-type tables are hardcoded constants, not new config fields (research.md); Task gets only the free duplicate-ID check it already qualifies for, not full body-content validation. |
| V. Test-First Discipline | **Gate carried into tasks.** Unit + filesystem-integration tests mandatory; 001/002/003's suites are explicit regression gates. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** The import-cycle discovery (research.md) is itself a Principle VI decision — the two packages' differing failure philosophies (fail-fast vs. collect-as-data) argue for the separation `docs/architecture-specification.md` §32 already specifies, not just permit it. `Status` composes `Inspect` and `validation.ValidateProject` rather than re-deriving either. |
| VII. Explicit Mutation Boundaries | **Pass (N/A).** Nothing is written; there is no boundary to bound. |
| VIII. Safety by Construction | **Pass (N/A for new risk).** No new filesystem-write surface is introduced; validation only reads via already-proven primitives (`ids.Scan`, `artifacts.ParseMetadata`). |
| IX. Transparent, Machine-Readable Contracts | **Pass, this feature's central concern.** The `ok` (command completed) vs. `valid` (structure is sound) distinction the constitution names explicitly is what `ValidateEntity`/`ValidateProject` returning Findings-not-errors *is*. Every `Finding.Code` is already mapped to a future JSON error code (contracts/validation.md). |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/004-structural-validation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── validation.md     # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Still a single Go module, following `docs/architecture-specification.md`
§32's layout — one new package, one new file in the existing `operations`
package:

```text
misterspec/
├── go.mod
├── go.sum
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── lock/, templates/                                  # unchanged (003)
    ├── operations/
    │   ├── resolve.go, inspect.go, parent.go, children.go,
    │   │   inventory.go, fingerprint.go, create.go,
    │   │   create_artifact.go, atomic_write.go            # unchanged
    │   ├── status.go              # NEW — Status (US3)
    │   └── status_test.go
    └── validation/                 # NEW package (US1, US2)
        ├── doc.go
        ├── findings.go             # Finding, Severity
        ├── tables.go                # allowed lifecycle states, required parent types
        ├── validator.go             # ValidateEntity, ValidateProject
        ├── validator_test.go
        └── tables_test.go
```

**Structure Decision**: One new package (`internal/validation`), depending
only on `internal/ids`/`internal/artifacts`/`internal/project` — never on
`internal/operations` (research.md's import-cycle finding), plus one new
file in the existing `operations` package for `Status`. Still no `cmd/`
entrypoint; still no body-content parsing.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
