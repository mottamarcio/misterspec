# Implementation Plan: Project Bootstrap

**Branch**: `007-project-bootstrap` | **Date**: 2026-09-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-project-bootstrap/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build Phase 5's deterministic bootstrap core as scoped in spec.md: a
read-only `Inspect` that determines whether a target directory is
already a misterspec project and, if so, which agent is installed (US1);
`Bootstrap`, which atomically sets up a new project's configuration,
kit resources, and a chosen agent's Skills for an uninitialized target,
rejecting an already-initialized target or an unregistered agent ID
before writing anything and reporting every part's own outcome (US2);
and `Verify`, confirming a completed bootstrap actually detects and the
right agent actually matches (US3). All three compose
001-core-foundation's project detection/configuration,
005-embedded-kit's template installation, and 006-agent-adapter's
registry/adapter/install-record primitives directly, in a new top-level
package, `internal/bootstrap` — no existing package gains a new
cross-dependency on another (research.md). `cmd/misterspec`, Cobra, and
the interactive Bubble Tea flow (agent selection, preview, confirm)
remain explicitly out of scope (spec.md Assumptions), the same
"operations before CLI" boundary every prior feature has kept. No new
external dependency, still no CLI/JSON surface.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged)
**Primary Dependencies**: Go standard library only (`os`, `io/fs`, `context`, `errors`) plus the existing `gopkg.in/yaml.v3` (already a dependency since 001-core-foundation, reused here to marshal a `project.Configuration` value) — no new external dependency.
**Storage**: Filesystem. This feature's writes are scoped to exactly what `Bootstrap` composes: `<targetDir>/.misterspec/config.yaml` (new, this feature), the kit's templates under `<targetDir>` (`internal/installer`, unmodified), the chosen agent's `TargetPath()` and `.misterspec/install.json` (`internal/agents`, unmodified) — never any other path. `Inspect`/`Verify` perform no writes at all (FR-001, FR-002).
**Testing**: `go test` — unit tests (`Inspect` against fixture directory states) and filesystem-integration tests (`Bootstrap` against a fresh temporary directory with a fixture `Registry` and fixture Skills `fs.FS`, `Verify` against `Bootstrap`'s own output), per Constitution Principle V. 001-core-foundation's, 005-embedded-kit's, and 006-agent-adapter's full suites are explicit, named regression gates since this feature composes all three unmodified.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. One new package: `internal/bootstrap` (`Inspect`, `Bootstrap`, `Verify`). No existing package is modified.
**Performance Goals**: Trivial — one config file, a handful of kit templates, a handful of Skill files; no performance-sensitive path.
**Constraints**: `internal/project` must not gain a dependency on `internal/installer` or `internal/agents` (would invert its foundational-layer role — research.md); `Bootstrap` rejects before writing anything, never partially (FR-004, FR-007, FR-009); the written configuration matches `project.Configuration`'s actual flat schema, not architecture §21's illustrative nested example (research.md); no new locking — `Inspect`-then-reject is sufficient for this one-time, interactively-confirmed operation (research.md, Constitution Principle IV).
**Scale/Scope**: `internal/bootstrap/{inspect.go, bootstrap.go, verify.go, config.go}` plus tests. No CLI, no Bubble Tea, no new adapter, no canonical Skill content (still Phase 6, unchanged scoping from 006-agent-adapter).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Bootstrapping is mechanical composition of already-proven deterministic primitives (detect, write config, install templates, install Skills) — no judgment about *which* agent to pick or *what* a project should contain; the caller supplies `agentID` already decided (spec.md Assumptions). |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A, distinct boundary).** Like `installer.Install` and `Adapter.Install` before it, `Bootstrap` allocates no entity ID and never touches a project's `ai/` artifact tree — it is project-scaffolding, not entity creation (`operations.Create`'s domain, untouched here). |
| III. Filesystem Is Single Source of Truth | **Pass, reinforced.** `Inspect`/`Verify` never cache or re-derive state of their own — every call re-reads `project.Detect` and `agents.CurrentInstall` fresh; `install.json`'s already-established non-authoritative status (006-agent-adapter) is unchanged. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** No new locking mechanism for `Bootstrap` despite it being a multi-step write sequence — `Inspect`-then-reject is sufficient for a one-time, interactively-confirmed operation, and adding `internal/lock` protection with no identified concurrent-usage scenario would be exactly the speculative infrastructure this principle rules out (research.md). |
| V. Test-First Discipline | **Gate carried into tasks.** Unit + filesystem-integration tests mandatory; 001/005/006's suites are named regression gates, not just "the usual suite," since this feature composes all three unmodified. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `internal/bootstrap` composes rather than duplicates — config-writing reuses `project.Configuration`'s own `Default*` constants and struct tags (not a hand-written YAML template), `Verify` reuses `Inspect` rather than a parallel check (research.md); `internal/project` stays dependency-light, exactly preserving the single-responsibility layering `operations` and `validation` already established. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** `Bootstrap`'s writes are exhaustively enumerated above (config.yaml, kit templates, agent Skills, install.json) — no path into any other part of the filesystem exists in this feature at all. |
| VIII. Safety by Construction | **Pass, reinforced.** Every write `Bootstrap` performs goes through an already-proven atomic-write path (`installer.WriteAtomicFile` for config.yaml, `installer.Install` for templates, `Adapter.Install`/`agents.RecordInstall` for Skills/install.json) — no new write path invented; rejection checks (`Inspect`, `registry.Get`) run strictly before any of them. |
| IX. Transparent, Machine-Readable Contracts | **Pass (deferred correctly, same as 001-006).** No CLI/JSON layer yet, but `BootstrapOutcome`'s per-part fields and the two new sentinel errors already match the "no aggregate pass/fail, distinct named conditions" shape a future JSON layer will use directly (contracts/bootstrap.md's mapping table). |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/007-project-bootstrap/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── bootstrap.md      # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Still a single Go module. `internal/bootstrap` is a new top-level
composition package, the same role `internal/operations` and
`internal/validation` already play (research.md), sitting above
`project`/`installer`/`agents` rather than inside any of them:

```text
misterspec/
├── go.mod
├── go.sum
├── kit/                                # unchanged
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── lock/, templates/, validation/                     # unchanged
    ├── operations/                                         # unchanged
    ├── installer/                                          # unchanged
    ├── agents/, agents/claude/, agents/builtin/            # unchanged
    └── bootstrap/                # NEW package
        ├── doc.go
        ├── inspect.go             # InspectResult, Inspect (US1)
        ├── config.go               # writeDefaultConfig — bootstrap-local, reuses project.Configuration + installer.WriteAtomicFile (research.md)
        ├── bootstrap.go             # BootstrapOutcome, Bootstrap, ErrAlreadyInitialized, ErrUnknownAgent (US2)
        ├── verify.go                 # VerifyResult, Verify (US3)
        ├── inspect_test.go
        ├── bootstrap_test.go
        └── verify_test.go
```

**Structure Decision**: One new package, `internal/bootstrap`, composing
three existing, unmodified packages. `internal/project` deliberately
does **not** gain `Bootstrap`-writing logic itself — that would invert
its foundational, dependency-light role (research.md) — so the
config-writing helper (`writeDefaultConfig`) lives in `bootstrap` and
reuses `project.Configuration`'s already-exported `Default*` constants
plus `internal/installer`'s already-exported `WriteAtomicFile`, rather
than either duplicating atomic-write logic a third time or creating a
backwards dependency. Still no `cmd/` entrypoint; still no second
adapter; still no canonical Skill content.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
