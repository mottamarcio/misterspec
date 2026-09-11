# Implementation Plan: Read-Only Deterministic Operations

**Branch**: `002-read-operations` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-read-operations/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build the first layer that *composes* 001-core-foundation's primitives
into the answers an agent actually asks for: find any entity by ID alone
and report its exact location or a distinct not-found/ambiguous result
(US1: `Resolve`, `Inspect`); walk structural parent/children relationships
using only fixed nesting rules (US2: `Parent`, `Children`); and discover
files in an area plus compute a content fingerprint, without interpreting
file contents (US3: `Inventory`, `Fingerprint`). One new package,
`internal/operations`, plus three small, additive, justified extensions to
001-core-foundation's `internal/ids` and `internal/artifacts` (a
`Paths` field on `ScanResult`, a `For` field on `Metadata`, and exporting
two already-tested private helpers) — all recorded in research.md. Still
no CLI/JSON surface, no mutation, no locking.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module `github.com/mottamarcio/misterspec`, unchanged)
**Primary Dependencies**: Go standard library only (`crypto/sha256`, `io`, `os`, `path/filepath`) — no new external dependency; `gopkg.in/yaml.v3` remains as-is from 001-core-foundation.
**Storage**: Filesystem only, read-only for this feature (Constitution Principle III; FR-014).
**Testing**: `go test` — unit tests for pure composition logic (e.g. path-prefix filtering in `Children`) and filesystem-integration tests against fixture trees (`internal/testutil`, extended if a new fixture shape is needed), per Constitution Principle V. 001-core-foundation's existing test suite must continue to pass unmodified, since this feature extends its types additively.
**Target Platform**: Cross-platform Go module (Linux, macOS, Windows) — unchanged from 001-core-foundation.
**Project Type**: Single Go module, library-first (still no `cmd/` — see Structure Decision).
**Performance Goals**: No hard targets; `Fingerprint` streams via `io.Copy` rather than buffering whole files, so large Raw sources (PDFs) don't force a large single allocation.
**Constraints**: Every operation is read-only (FR-014) — no operation in this feature may open a file for writing, create, rename, or delete anything, verified by test (SC-006). All path handling reuses 001-core-foundation's proven traversal-containment check rather than a second implementation.
**Scale/Scope**: One new package (`internal/operations`, 6 files) plus additive extensions to two existing packages — no `internal/cli`, `internal/tui`, `internal/validation`, `internal/installer`, or `internal/agents` code (later phases per architecture spec §68).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Every operation here is pure search/composition over filesystem structure and already-parsed metadata — no semantic judgment. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A for mutation).** This feature adds zero write paths; `create`/`create-artifact` remain a distinct, later feature. |
| III. Filesystem Is Single Source of Truth | **Pass.** `Resolve`/`Children` derive entirely from a fresh `ids.Scan` each call — no caching, no stored index. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass.** `Children` deliberately reuses `Scan` + a path-prefix filter instead of adding a new scan primitive before a second consumer needs one (research.md). No new config fields. |
| V. Test-First Discipline | **Gate carried into tasks.** Unit + filesystem-integration tests are mandatory deliverables per story, and 001-core-foundation's existing suite must stay green throughout — both are explicit task-level checks, not a follow-up. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** This plan explicitly removes duplication that would otherwise appear: `ids.ParseAny` (shared by `artifacts.ParseMetadata` and `operations`) and `artifacts.RelativeWithinRoot` (shared by `Inventory`/`Fingerprint` and `ResolvePath`/`ClassifyPath`) are each one implementation reused, not two. |
| VII. Explicit Mutation Boundaries | **Pass (N/A).** No artifact is written; nothing to bound. |
| VIII. Safety by Construction | **Pass, reinforced.** `Inventory`/`Fingerprint` reuse the already-tested containment check rather than a parallel one — one proven implementation instead of a second that could silently diverge. |
| IX. Transparent, Machine-Readable Contracts | **Pass (deferred correctly, same as 001).** No JSON/CLI layer yet, but every new error (`ErrEntityNotFound`, `ErrEntityAmbiguous`, `ErrInvalidTarget`) is already a typed sentinel mapped 1:1 to the stable JSON codes in `docs/architecture-specification.md` §7 (`entity_not_found`, `entity_ambiguous`, `invalid_target` — codes the architecture spec already reserved), so nothing will need renaming later. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/002-read-operations/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── operations.md     # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Still a single Go module, following `docs/architecture-specification.md`
§32's concrete package layout — this feature adds one new package and
touches two existing ones additively:

```text
misterspec/
├── go.mod
├── go.sum
└── internal/
    ├── project/            # unchanged (001-core-foundation)
    ├── artifacts/
    │   ├── metadata.go     # MODIFIED: + For field on Metadata
    │   ├── parser.go       # MODIFIED: decode `for:`; parseFieldID delegates to ids.ParseAny
    │   ├── paths.go         # MODIFIED: relativeWithinRoot exported as RelativeWithinRoot
    │   └── ...              # types.go, doc.go unchanged
    ├── ids/
    │   ├── ids.go            # MODIFIED: + ParseAny
    │   ├── scan.go            # MODIFIED: + Paths field on ScanResult, populated in buildScanResult
    │   └── ...                # types.go, doc.go unchanged
    ├── testutil/             # unchanged, reused as-is
    └── operations/            # NEW package (this feature)
        ├── doc.go
        ├── resolve.go         # Resolve, ResolvedLocation, ErrEntityNotFound, ErrEntityAmbiguous, AmbiguousIDError (US1)
        ├── inspect.go          # Inspect, InspectResult, ErrInvalidTarget (US1)
        ├── parent.go           # Parent, ParentResult (US2)
        ├── children.go         # Children (US2)
        ├── inventory.go        # Inventory, FileEntry (US3)
        ├── fingerprint.go      # Fingerprint (US3)
        ├── resolve_test.go
        ├── inspect_test.go
        ├── parent_test.go
        ├── children_test.go
        ├── inventory_test.go
        └── fingerprint_test.go
```

**Structure Decision**: One new package (`internal/operations`), plus
additive-only changes to `internal/ids` and `internal/artifacts` (new
fields, new exported functions, one internal delegation) — no existing
exported signature from 001-core-foundation is removed or changed
incompatibly, and 001-core-foundation's test suite is a required, explicit
regression gate for this feature (Constitution Principle V). Still no
`cmd/` entrypoint — the CLI surface remains a later feature.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
