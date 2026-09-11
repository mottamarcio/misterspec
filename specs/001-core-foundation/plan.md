# Implementation Plan: Core Repository Foundation

**Branch**: `001-core-foundation` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-core-foundation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build the read-only foundation layer misterspec's entire deterministic/
semantic split depends on: detect whether a directory is an initialized
misterspec project and load its configuration (US1); compute canonical
artifact paths and classify artifact files by type, safely (US2); and parse
artifact frontmatter into structured metadata plus scan the project tree
for existing entity IDs, flagging invalid/duplicate ones and computing the
next available ID by scanning rather than a stored counter (US3). Three new
internal Go packages (`project`, `artifacts`, `ids`) with no CLI, no TUI, and
no file-writing code path — this is Phase 1 of
`docs/architecture-specification.md` §68 only.

## Technical Context

**Language/Version**: Go 1.23+ (module `github.com/mottamarcio/misterspec`)
**Primary Dependencies**: `gopkg.in/yaml.v3` (frontmatter parsing); Go standard library (`path/filepath`, `io/fs`, `errors`) for everything else. Cobra and Bubble Tea are project-wide frozen choices (§70) but are out of scope for this feature — no CLI or TUI code is added here.
**Storage**: Filesystem only (Markdown + YAML frontmatter files under the project's `ai/` tree) — no database, no vector store (Constitution Principle III).
**Testing**: `go test` — unit tests for pure logic (ID parsing, `NextID`, path computation) and filesystem-integration tests against fixture trees built with `t.TempDir()`, per Constitution Principle V.
**Target Platform**: Cross-platform Go module (Linux, macOS, Windows) — no OS-specific code paths anticipated; path handling uses `path/filepath` throughout for portability.
**Project Type**: Single Go module, library-first (no `cmd/` entrypoint added yet — see Structure Decision).
**Performance Goals**: No hard targets; must stay comfortably interactive (well under 1s) for project sizes up to several hundred artifacts, since every deterministic operation and Skill invocation will call into this layer synchronously.
**Constraints**: No network access required (all logic is local filesystem I/O); no persisted counters or caches (Constitution Principle III); every path operation must guarantee containment inside the project root (Constitution Principle VIII).
**Scale/Scope**: Three packages (`internal/project`, `internal/artifacts`, `internal/ids`) implementing exactly the FRs in `spec.md` — no `internal/operations`, `internal/cli`, `internal/tui`, `internal/validation`, `internal/installer`, or `internal/agents` code in this feature (those are later features per §68 Phases 2–6).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Every capability in this feature is pure computation over filesystem structure (detection, path math, parsing, scanning) — no semantic judgment is made or required. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A for mutation).** This feature performs no mutation at all — it is read-only by design (see research.md "Scope boundary"). The atomic `create` operation this layer will eventually support is explicitly out of scope here. |
| III. Filesystem Is Single Source of Truth | **Pass.** `ids.NextID` is a pure function over a freshly `Scan`ned slice; no counter file is read or written anywhere in this feature (FR-013). |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass.** Exactly three packages, matching §32's own layout; no speculative `domain/`, `services/`, or `repositories/` layers; no new config fields beyond what §21's schema already defines. |
| V. Test-First Discipline | **Gate carried into tasks.** Unit tests for `ids`/path-math and filesystem-integration tests (fixture trees) are mandatory deliverables of this feature, not a follow-up. |
| VI. Clean Code & SOLID | **Pass.** Package boundaries enforce SRP (`project` detects/configures, `artifacts` resolves/classifies/parses, `ids` parses/scans entity IDs); `contracts/packages.md` keeps each package's exported surface small and consumer-shaped. |
| VII. Explicit Mutation Boundaries | **Pass (N/A).** No artifact is written by this feature; nothing to bound. |
| VIII. Safety by Construction | **Gate carried into design.** `artifacts.ResolvePath`/`ClassifyPath` MUST reject any result outside `Project.Root` *before* touching the filesystem (FR-008, SC-004) — enforced by `data-model.md`'s Canonical Path validation rules and directly tested. |
| IX. Transparent, Machine-Readable Contracts | **Pass (deferred correctly).** The JSON CLI layer itself is out of scope, but every distinguishable failure is already a typed/sentinel Go error (research.md "Error modeling") mapped 1:1 to the future JSON error codes in `contracts/packages.md`, so nothing will need renaming when that layer is built. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/001-core-foundation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── packages.md       # Phase 1 output (/speckit-plan command) — internal Go package API contracts
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

This is a single Go module. It does not use the generic `src/`/`tests`
layout — it follows the concrete package layout already frozen in
`docs/architecture-specification.md` §32, scoped to exactly the packages
this feature needs:

```text
misterspec/
├── go.mod
├── go.sum
└── internal/
    ├── project/
    │   ├── root.go          # project root detection (US1 / FR-001..003)
    │   ├── config.go        # config loading + validation (US1 / FR-004..005)
    │   ├── root_test.go
    │   └── config_test.go
    ├── artifacts/
    │   ├── types.go          # ArtifactType / EntityType classification (US2 / FR-006)
    │   ├── paths.go          # canonical path resolution + traversal safety (US2 / FR-007..008)
    │   ├── metadata.go        # Metadata struct + frontmatter extraction (US3 / FR-009)
    │   ├── parser.go          # YAML frontmatter parsing (US3 / FR-009)
    │   ├── types_test.go
    │   ├── paths_test.go
    │   └── parser_test.go
    └── ids/
        ├── ids.go             # EntityID type, Parse, NextID (US3 / FR-010, FR-013)
        ├── scan.go             # project-tree scanning (US3 / FR-011..012, FR-014)
        ├── ids_test.go
        └── scan_test.go
```

**Structure Decision**: Single Go module, no `cmd/` package yet. This
feature is purely the foundation library — `cmd/misterspec/main.go` and the
`misterspec internal ...` CLI surface belong to later features (§68 Phases
2 and 5) that consume these three packages. Adding a `cmd/` entrypoint now,
before there is any command to wire it to, would be exactly the kind of
speculative structure Constitution Principle IV forbids. `go.mod`/`go.sum`
are created by this feature since it's the first code in the repository.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
