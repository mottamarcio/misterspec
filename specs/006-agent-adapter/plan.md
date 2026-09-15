# Implementation Plan: Agent Adapter Layer

**Branch**: `006-agent-adapter` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-agent-adapter/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build Phase 4's Agent Adapter layer as scoped in spec.md: the `Adapter`
interface and a `Registry` for discovery/selection (US1); one working
adapter, Claude Code, whose `Install` materializes Skill resources into
`.claude/skills` and writes an `install.json` record (US2); and a way to
read that record back to learn what's currently installed, without
re-installing or guessing (US3). `internal/installer` (005-embedded-kit)
is generalized additively (`InstallFS`/`ListFS`) so this feature reuses
its proven atomic-write and no-silent-overwrite guarantees rather than
reimplementing them a third time. Additional adapters and canonical
Skill content itself remain explicitly out of scope (spec.md
Assumptions) until they're actually needed. No new external dependency,
still no CLI/JSON surface.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged)
**Primary Dependencies**: Go standard library only (`io/fs`, `encoding/json`, `context`) — no new external dependency.
**Storage**: Filesystem. This feature's writes are scoped to exactly two things: an adapter's own `TargetPath()` (e.g. `.claude/skills`) and `.misterspec/install.json` — never a project's `ai/` artifact tree (research.md, matching 005-embedded-kit's identical boundary).
**Testing**: `go test` — unit tests (`Registry` lookup/listing) and filesystem-integration tests (`Adapter.Install` against a fixture Skills `fs.FS`, `CurrentInstall` round-tripping), per Constitution Principle V. 005-embedded-kit's full suite is an explicit, named regression gate for `internal/installer`'s generalization.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. `internal/installer` extended additively (two new functions, existing ones now thin wrappers). Three new packages: `internal/agents` (interface + types + registry), `internal/agents/claude` (the one concrete adapter), `internal/agents/builtin` (wiring, to avoid an import cycle — research.md).
**Performance Goals**: Trivial — a handful of small Skill files; no performance-sensitive path.
**Constraints**: `Adapter.Install` is one call producing a complete, recorded outcome (research.md); `internal/agents` never imports `internal/agents/claude` (the cycle `builtin` exists specifically to avoid); `install.json` is explicitly non-authoritative project state (FR-008, Constitution Principle III).
**Scale/Scope**: `internal/installer/{installer.go}` gains `InstallFS`/`ListFS`; `internal/agents/{adapter.go,registry.go,record.go}` plus tests; `internal/agents/claude/claude.go`; `internal/agents/builtin/builtin.go`. No CLI, no second adapter, no canonical Skill content, no interactive init flow.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Installing Skills for a chosen agent is mechanical materialization — no judgment about which agent to pick (the caller decides) or what a Skill means. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A, distinct boundary).** Like `Install` in 005, this allocates no entity ID and never touches `ai/` — it's a resource-materialization primitive, not entity creation. |
| III. Filesystem Is Single Source of Truth | **Pass, reinforced.** `install.json` is explicitly documented (FR-008, §22) as non-authoritative — a report, never re-derived as project state; `CurrentInstall` only ever reads what `Install` wrote, never infers from `TargetPath()`'s contents. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** Exactly one concrete adapter (Claude Code) — the architecture spec itself calls further adapters "a release decision"; `Registry` is a plain constructed value, not a dynamic plugin-loading system; no adapter self-registration magic for a single adapter. |
| V. Test-First Discipline | **Gate carried into tasks.** Unit + filesystem-integration tests mandatory; 005-embedded-kit's suite is a named regression gate for `internal/installer`'s generalization, not just "the usual suite." |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `InstallFS`/`ListFS` generalize 005's mechanism instead of a third reimplementation; `recordInstall` is a shared helper every future adapter reuses rather than each writing `install.json` itself; the `builtin` package cleanly separates "what an adapter is" from "which adapters exist," avoiding a cycle without compromising either concern. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** An adapter's writes are scoped to exactly `TargetPath()` and `install.json` — the same "framework resources only, never `ai/`" boundary 005-embedded-kit established, now proven to extend cleanly to a second writer. |
| VIII. Safety by Construction | **Pass, reinforced.** Every write here goes through `internal/installer`'s already-proven atomic-write and containment guarantees — no new write path invented. |
| IX. Transparent, Machine-Readable Contracts | **Pass (deferred correctly, same as 001-005).** No CLI/JSON layer yet, but `InstallRecord` already serializes directly to §22's frozen `install.json` shape via `json:` tags, and `Registry.Get`/`CurrentInstall`'s bool-signaled absence already matches the `{"ok": true, ...: null}` shape a future JSON layer will use. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/006-agent-adapter/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── agents.md         # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Still a single Go module, following `docs/architecture-specification.md`
§32's `internal/agents/{registry.go, adapter.go, claude/, codex/, ...}`
layout, with one addition (`builtin/`) research.md justifies:

```text
misterspec/
├── go.mod
├── go.sum
├── kit/                                # unchanged
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── lock/, templates/, validation/                     # unchanged
    ├── operations/                                         # unchanged
    ├── installer/
    │   ├── installer.go        # MODIFIED: + InstallFS, ListFS; List/Install become thin wrappers
    │   ├── installer_test.go   # MODIFIED: + tests for InstallFS/ListFS
    │   ├── filesystem.go, containment_test.go, filesystem_test.go, doc.go  # unchanged
    └── agents/                  # NEW package (US1)
        ├── doc.go
        ├── adapter.go            # Adapter interface, InstallRequest, InstallResult
        ├── registry.go            # Registry, NewRegistry, List, Get
        ├── record.go               # InstallRecord, recordInstall, CurrentInstall (US2, US3)
        ├── registry_test.go
        ├── record_test.go
        └── claude/                  # NEW package (US2)
            ├── claude.go
            └── claude_test.go
        └── builtin/                  # NEW package (US1 wiring)
            └── builtin.go
```

**Structure Decision**: Three new packages (`agents`, `agents/claude`,
`agents/builtin`) plus an additive extension to the existing `installer`
package. `agents/builtin` is the one deliberate addition beyond §32's own
sketch, existing specifically to avoid the `agents` ↔ `claude` import
cycle a naive "agents package also knows how to build its own default
registry" design would create (research.md). Still no `cmd/` entrypoint;
still no second adapter; still no canonical Skill content.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
