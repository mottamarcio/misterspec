# Implementation Plan: Embedded Kit and Resource Installer

**Branch**: `005-embedded-kit` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-embedded-kit/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build the Embedded Kit as scoped in spec.md: relocate
003-entity-creation's 8 artifact templates from a private,
feature-scoped embed into a shared top-level `kit` package
(`kit/templates/*.tmpl`) matching `docs/architecture-specification.md`
§32's intended layout, so templates have exactly one source of truth
(US3); expose kit-content discovery (US1, `internal/installer.List`);
and build a generic, agent-agnostic materialization primitive (US2,
`internal/installer.Install`) that writes embedded resources to a target
directory atomically, never silently overwriting, never leaving a
partial file. Canonical Skill content and agent-specific integration
resources are explicitly deferred (spec.md Assumptions) until Phase 6 and
Phase 4 respectively produce real content to embed. No new external
dependency, still no CLI/JSON surface.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged)
**Primary Dependencies**: Go standard library only (`embed`, `io/fs`) — no new external dependency.
**Storage**: Filesystem. This feature's `Install` is this project's second write path (after 003-entity-creation's `Create`/`CreateArtifact`) — but writes to a caller-chosen target directory for framework resources, never into a project's `ai/` artifact tree (research.md's "raw materialization, not entity creation" decision).
**Testing**: `go test` — unit tests (`List`'s derivation from the embedded FS) and filesystem-integration tests (fresh target directory, already-populated target directory with and without overwrite, traversal rejection), per Constitution Principle V. 001/002/003/004's full suites — especially 003-entity-creation's — are mandatory regression gates, since this feature relocates content 003 depends on.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. One new top-level package (`kit`), one new package (`internal/installer`), one modified package (`internal/templates`, internals only — its exported API is unchanged).
**Performance Goals**: Trivial — a handful of small embedded files; no performance-sensitive path here.
**Constraints**: Every `Install` write is atomic (FR-003, FR-007); no silent overwrite without explicit request (FR-004); every destination is containment-checked via the already-proven `artifacts.RelativeWithinRoot` before any write (FR-006); `internal/installer` does not import `internal/operations` (research.md — a sibling primitive package, not a consumer of operations' compositions; its own small atomic-write helper is a deliberate, documented duplication rather than that cross-dependency).
**Scale/Scope**: `kit/kit.go` + 8 relocated `.tmpl` files; `internal/installer/{installer.go, filesystem.go}` plus tests; `internal/templates/templates.go` modified to read from `kit` instead of its own embed. No CLI, no Skill content, no agent adapter logic.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Listing and installing embedded resources is purely mechanical — no judgment about what should be installed or how, only "copy what's embedded, safely." |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (distinct from Create).** `Install` is not entity creation — it allocates no ID and never touches a project's `ai/` tree; it's a lower-level resource-materialization primitive `misterspec init` (a later feature) will orchestrate. This boundary is explicit in research.md and spec.md's Assumptions, not assumed. |
| III. Filesystem Is Single Source of Truth | **Pass, extended.** The compiled binary's embedded `kit.TemplatesFS` is now the single source of truth for template *content* (FR-008) — `internal/templates` and `internal/installer` both read the same bytes; no independent copy exists anywhere to drift. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** `kit/skills/` and `kit/integrations/` are deliberately not created with placeholder content (spec.md Assumptions) — building them now, before Phase 4/6 have real content, would be speculative. `installer`'s own atomic-write helper is a small, justified duplication rather than a premature shared package for two call sites (research.md). |
| V. Test-First Discipline | **Gate carried into tasks.** Unit + filesystem-integration tests mandatory; 003-entity-creation's suite is an explicit, named regression gate (not just "the usual suite") since this feature relocates content it depends on. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `Install` reuses `artifacts.RelativeWithinRoot` rather than a fourth containment implementation; `kit` becomes templates' one source of truth (closing a duplication `internal/templates`'s private embed would otherwise have created against this feature's kit root). |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** `Install` writes only to its caller-chosen target directory for framework resources — it has no path into a project's `ai/` artifact tree at all, a boundary enforced by what `Install` even accepts as input, not just documented. |
| VIII. Safety by Construction | **Pass, reinforced.** Atomic writes (temp+fsync+rename, `filesystem.go`), no silent overwrite (`Skipped` is a first-class, reported outcome, not silence), containment checked before any write — the same discipline 003-entity-creation established, applied to a new write path. |
| IX. Transparent, Machine-Readable Contracts | **Pass (deferred correctly, same as 001-004).** No CLI/JSON layer yet, but `Outcome.Status` is already a structured per-resource result (`Installed`/`Skipped`/`Failed`) mapped to a future JSON `status` field (contracts/kit-installer.md), and containment failures reuse the already-mapped `artifacts.ErrPathOutsideProject` rather than minting a new, unmapped error. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/005-embedded-kit/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── kit-installer.md  # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Still a single Go module. `kit/` appears at the module root for the first
time, matching `docs/architecture-specification.md` §32's own layout
exactly (research.md):

```text
misterspec/
├── go.mod
├── go.sum
├── kit/                          # NEW top-level package
│   ├── kit.go                     # //go:embed templates/*.tmpl
│   └── templates/                  # relocated from internal/templates/files/
│       ├── program.md.tmpl
│       ├── feature.md.tmpl
│       ├── spec.md.tmpl
│       ├── knowledge.md.tmpl
│       ├── learning.md.tmpl
│       ├── plan.md.tmpl
│       ├── tasks.md.tmpl
│       └── validation.md.tmpl
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── lock/                                              # unchanged (003)
    ├── operations/                                        # unchanged — still the only writer into a project's ai/ tree
    ├── validation/                                         # unchanged (004)
    ├── templates/
    │   ├── doc.go, templates_test.go                       # unchanged
    │   ├── templates.go                                     # MODIFIED: reads via kit.TemplatesFS, not its own embed
    │   └── files/                                            # REMOVED — content now lives in kit/templates/
    └── installer/                    # NEW package (US1, US2)
        ├── doc.go
        ├── installer.go               # List, Install, Resource, Outcome
        ├── filesystem.go               # atomic-write helper (its own — research.md)
        ├── installer_test.go
        └── filesystem_test.go
```

**Structure Decision**: One new top-level package (`kit`), one new
`internal/installer` package matching §32's file list exactly, and a
content-only relocation inside `internal/templates` (its exported API is
untouched). Still no `cmd/` entrypoint. `kit/skills/` and
`kit/integrations/` are not created in this feature — they arrive with
the features that have real content for them.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
