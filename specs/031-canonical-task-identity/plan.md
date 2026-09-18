# Implementation Plan: Modelo canônico de tarefas e identidade composta

**Branch**: `031-canonical-task-identity` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/031-canonical-task-identity/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Task identity is currently ambiguous across Specs: `internal/ids/scan.go`'s `scanTaskHeadings` globs every `*/features/*/specs/*/tasks.md` and keys all `## TASK-NNN` headings in one shared `map[int][]string`, so a `TASK-001` in `SPEC-001` and a `TASK-001` in `SPEC-002` collide as if they were the same identity — both in `operations.Resolve`/`Inspect` (ambiguous-resolution error) and in `validation.ValidateProject` (false-positive duplicate finding). The Skills (`mister-implement`, `mister-tasks`, `mister-analyze`) already document Task numbering as per-Spec, so the fix is to make the code's task-identity model match that documented contract: scope duplicate detection to one Spec's `tasks.md`, and let a bare `TASK-NNN` resolve globally *only* when its number is unambiguous project-wide, otherwise require the caller to supply the owning Spec (via a composite `SPEC-###:TASK-###` form or an accompanying Spec argument).

## Technical Context

**Language/Version**: Go 1.23.4 (existing `go.mod`)
**Primary Dependencies**: Cobra (CLI), existing internal packages only — `internal/ids`, `internal/operations`, `internal/validation`, `internal/project`, `internal/artifacts`. No new external dependency.
**Storage**: N/A — filesystem (Markdown + YAML frontmatter) is the sole source of truth; no database, no persistent counter (Constitution Principle III).
**Testing**: Go `testing` package — unit tests for parsing/scanning (`internal/ids`), filesystem-integration tests using temp repositories for resolve/validate (`internal/operations`, `internal/validation`), matching existing `*_test.go` conventions in those packages.
**Target Platform**: `misterspec` CLI binary (Linux/macOS/Windows), invoked directly and via agent Skills.
**Project Type**: Single Go module, CLI + embedded kit (existing structure — no new top-level project).
**Performance Goals**: No regression vs. today's single-pass directory scan; task resolution/validation must remain one filesystem walk per invocation (no per-Spec re-walk of the whole tree).
**Constraints**: MUST NOT change the human-authored `## TASK-NNN — Title` heading syntax; MUST NOT renumber existing tasks automatically (spec FR-007); MUST NOT introduce a persistent ID counter (Constitution Principle III); MUST NOT add a new package layer without demonstrated need (Constitution Principle IV) — this fits inside the existing `ids`/`operations`/`validation` boundaries.
**Scale/Scope**: Touches `internal/ids/scan.go`, `internal/ids/ids.go` (or a new small file in that package), `internal/operations/resolve.go`, `internal/operations/inspect.go`, `internal/validation/validator.go`, `internal/cli/internalcmd/inspect.go` (and any sibling internalcmd that accepts a task argument), plus the already-correct Skill docs (`mister-implement`, `mister-tasks`, `mister-analyze`) which need no behavioral change, only confirmation the binary now matches them.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Semantic/Deterministic Separation** — PASS. Task identity resolution and duplicate detection are purely mechanical (string parsing + filesystem scanning); no semantic judgment is added to the Go binary, and Skills continue to call `misterspec internal …` rather than reproducing resolution logic.
- **II. Deterministic Operations Are the Only Mutation Primitive** — PASS. This feature is read/validate-only (resolve, inspect, validate); it does not add or change any mutating `create`/`create-artifact` path, and does not touch ID allocation (`NextID` is untouched — Task numbering-on-creation is out of scope here, per spec Assumptions).
- **III. Filesystem Is the Single Source of Truth** — PASS. The composite identity (`SpecID + local TASK-NNN`) is derived at scan/resolve time from `tasks.md` headings and directory names; nothing new is persisted.
- **IV. Simplicity First — YAGNI & Minimal Configuration** — PASS. No new package layer; the composite identity type and per-Spec scan live inside `internal/ids`, alongside `EntityID`/`Scan`, which already own this responsibility. No new config field — Spec context comes from the existing command argument / invocation context, not a new `.misterspec/config.yaml` knob.
- **V. Test-First Discipline** — Applies. New/changed behavior (composite parsing, per-Spec duplicate scoping, ambiguous-without-context resolution) MUST get unit tests in `internal/ids` and filesystem-integration tests in `internal/operations`/`internal/validation` before/alongside implementation, per existing `*_test.go` conventions in those packages.
- **VI. Clean Code & SOLID** — PASS. `ids` keeps owning parsing/scanning; `operations` keeps owning resolution; `validation` keeps owning structural checks — no responsibility shifts between packages, avoiding the blur Principle VI warns against.
- **VII. Explicit Mutation Boundaries** — N/A. No artifact-owning Skill or operation gains write access to a file it doesn't already own.
- **VIII. Safety by Construction** — N/A. No new filesystem write path is introduced.
- **IX. Transparent, Machine-Readable Contracts** — Applies. The new "Spec context required" outcome (spec FR-003) and the migration diagnostic (FR-008) MUST be structured JSON with stable error/finding codes, consistent with existing `ErrEntityAmbiguous`/`AmbiguousIDError` and `validation.Finding` conventions — not new prose-only output.

No violations requiring justification; Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/031-canonical-task-identity/
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
├── ids/
│   ├── types.go          # EntityID, EntityType (existing) — add TaskID composite type here
│   ├── ids.go             # Parse/ParseAny/NextID (existing) — add ParseTaskRef
│   ├── scan.go            # Scan/scanTaskHeadings (existing) — rescope Task scanning per-Spec
│   └── scan_test.go / ids_test.go   # existing test files — extended, not replaced
├── operations/
│   ├── resolve.go         # Resolve (existing) — Task path gains Spec-context handling
│   ├── inspect.go         # Inspect/inspectTask (existing) — consumes the new resolution path
│   └── resolve_test.go / inspect_test.go
├── validation/
│   ├── validator.go       # ValidateProject (existing) — duplicate check rescoped per-Spec
│   └── validator_test.go
└── cli/internalcmd/
    ├── inspect.go          # `inspect <id>` — gains optional Spec-context argument/flag
    └── inspect_test.go (or resolve_test.go, mirroring existing patterns)

kit/skills/
├── mister-implement/SKILL.md   # already documents composite intent — verify wording still matches
├── mister-tasks/SKILL.md
└── mister-analyze/SKILL.md
```

**Structure Decision**: No new top-level directory or package. This is a scoped change inside the existing `internal/ids` → `internal/operations` → `internal/validation` → `internal/cli/internalcmd` pipeline (single Go module, CLI project type), matching Constitution Principle IV and Principle VI's existing package boundaries.

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
