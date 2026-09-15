# Implementation Plan: Atomic Entity Creation

**Branch**: `003-entity-creation` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-entity-creation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

This is the first feature that writes to the filesystem. Build the atomic
entity-creation primitive Constitution Principle II already mandates:
`Create` allocates a new entity's ID (Program/Feature/Spec/Knowledge/
Learning) and writes its initial artifact from a fixed template as one
all-or-nothing operation (US1); `CreateArtifact` does the same for a
Spec's subordinate Plan/Tasks/Validation, no new ID needed (US2); both are
protected by a short-lived, non-authoritative, stale-recoverable lock so
concurrent requests never collide (US3). Two new packages
(`internal/lock`, `internal/templates`) plus two new files in
`internal/operations`, built entirely on top of 001-core-foundation and
002-read-operations — no new external dependency, still no CLI/JSON
surface.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module `github.com/mottamarcio/misterspec`, unchanged)
**Primary Dependencies**: Go standard library only (`embed`, `text/template`, `os`, `path/filepath`, `time`) — no new external dependency.
**Storage**: Filesystem — this feature's first write path. Every write follows the atomic recipe (temp file → complete write → fsync → `os.Rename`) from `docs/architecture-specification.md` §57.
**Testing**: `go test` — unit tests (template rendering, lock staleness logic), filesystem-integration tests (fixture projects), and a concurrency test (overlapping goroutines calling `Create`), per Constitution Principle V. 001-core-foundation's and 002-read-operations's full suites remain mandatory regression gates.
**Target Platform**: Cross-platform Go module (Linux, macOS, Windows) — the lock uses only `os.OpenFile(O_CREATE|O_EXCL)`, portable across all three without OS-specific syscalls.
**Project Type**: Single Go module, library-first (still no `cmd/`).
**Performance Goals**: Lock defaults (`staleAfter` 10s, acquire `timeout` 5s, ~50ms poll) are generous relative to how long scaffolding a few files actually takes — a legitimate concurrent creation is never mistaken for staleness, while a crashed process's lock recovers well within one interactive workflow's patience.
**Constraints**: Every write is atomic (FR-007); every creation is rejected, not partially applied, on any invalid input (FR-004, FR-008, FR-011, FR-012, FR-014); the lock is explicitly non-authoritative project state (Constitution Principle III) and always safely recoverable (FR-006).
**Scale/Scope**: `internal/lock` (3 files), `internal/templates` (embeds exactly 8 `.tmpl` files — not the broader Embedded Kit), and `internal/operations/{create.go,create_artifact.go}` — no `cmd/`, no `internal/cli`, no `internal/installer`, no full `kit/` structure.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This is the first feature where Principles II, III, VII, and VIII move
from "N/A (no mutation yet)" to actively exercised — noted explicitly
below rather than glossed over.

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Creation is pure mechanical scaffolding (allocate, template, write); no judgment about *what* to create — the agent supplies type/parent/slug, this layer only executes them. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass, now actively fulfilled.** This feature *is* the atomic `create`/`create-artifact` primitive the constitution requires — ID allocation and initial artifact creation are a single operation, never a split `next-id` → `scaffold` sequence. |
| III. Filesystem Is Single Source of Truth | **Pass.** IDs are still allocated by scanning (`ids.Scan`/`ids.NextID`, reused from 001/002) — no persistent counter. The lock itself is explicitly non-authoritative and safely recoverable, exactly the carve-out Principle III names. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass.** Lock uses only stdlib `O_EXCL` (no flock dependency); templates cover exactly the 8 files this feature needs, not the full Embedded Kit; no new config fields (lock timeouts are constants, not `.misterspec/config.yaml` entries, until a real need for tuning them appears). |
| V. Test-First Discipline | **Gate carried into tasks.** Unit + filesystem-integration + a concurrency test are mandatory deliverables; 001/002's full suites are explicit regression gates, not optional. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `Create`'s parent validation reuses `operations.Resolve` rather than reimplementing it; `CreateArtifact.Kind` reuses `artifacts.ArtifactType` rather than a second enum; locking lives in its own package (SRP) rather than duplicated across `create.go`/`create_artifact.go`. |
| VII. Explicit Mutation Boundaries | **Pass, now actively exercised.** `Create`/`CreateArtifact` only ever write to a not-yet-existing canonical path — FR-008/FR-012 make overwriting an existing artifact an explicit, tested rejection, never a silent rewrite. |
| VIII. Safety by Construction | **Pass, now actively exercised.** Atomic writes (temp+rename+fsync), path containment (reusing `artifacts.RelativeWithinRoot`/`ResolvePath`), and the recoverable lock are this feature's central concern, not an afterthought. |
| IX. Transparent, Machine-Readable Contracts | **Pass (deferred correctly, same as 001/002).** No CLI/JSON layer yet, but `ErrInvalidParent`/`ErrAlreadyExists`/`ErrUnsupportedType`/`lock.ErrLockTimeout` are already typed sentinels mapped to the stable JSON codes §7 already reserves (`invalid_parent`, `already_exists`, `unsupported_type`). |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/003-entity-creation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── creation.md       # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Still a single Go module, following `docs/architecture-specification.md`
§32's package layout — two new packages, two new files in the existing
`operations` package:

```text
misterspec/
├── go.mod
├── go.sum
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── operations/
    │   ├── resolve.go, inspect.go, parent.go,
    │   │   children.go, inventory.go, fingerprint.go       # unchanged (002)
    │   ├── create.go              # NEW — Create (US1)
    │   ├── create_artifact.go     # NEW — CreateArtifact (US2)
    │   ├── create_test.go
    │   ├── create_concurrency_test.go   # US3
    │   └── create_artifact_test.go
    ├── lock/                       # NEW package (US3, prerequisite for US1/US2)
    │   ├── doc.go
    │   ├── lock.go
    │   └── lock_test.go
    └── templates/                  # NEW package (prerequisite for US1/US2)
        ├── doc.go
        ├── templates.go
        ├── templates_test.go
        └── files/
            ├── program.md.tmpl
            ├── feature.md.tmpl
            ├── spec.md.tmpl
            ├── knowledge.md.tmpl
            ├── learning.md.tmpl
            ├── plan.md.tmpl
            ├── tasks.md.tmpl
            └── validation.md.tmpl
```

**Structure Decision**: Two new packages plus two new files in the
existing `operations` package — no new architectural layer beyond what
§32 already lays out (`lock` and `templates` are both named there,
conceptually, as supporting the `operations` layer). Still no `cmd/`
entrypoint. `internal/templates` embeds exactly the 8 files this feature
needs, not the broader `kit/` structure `misterspec init` will require —
that remains a distinct, later feature.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
