# Implementation Plan: CLI Command Layer (Cobra)

**Branch**: `008-cli-cobra` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/008-cli-cobra/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build the CLI command layer scoped in spec.md: a hidden `misterspec
internal <op>` tree exposing all ten of 001-007's already-implemented
deterministic operations (resolve, inspect, parent, children, create,
create-artifact, fingerprint, inventory, validate, status) as
machine-readable JSON commands with stable error codes and
category-specific exit codes (US1, §5-8); a public, non-interactive
`misterspec init --agent <id> [--dir <path>]` wrapping
007-project-bootstrap's `Bootstrap` directly (US2, §4.1); and a minimal
top-level `--help` surface that never advertises the internal tree
(US3, §4). This is the first feature to add an external dependency
beyond `gopkg.in/yaml.v3` (Cobra) and the first to produce a real binary
(`cmd/misterspec`). No new business logic anywhere — every command is a
thin adapter over 001-007's already-tested Go packages (research.md).

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged)
**Primary Dependencies**: `gopkg.in/yaml.v3` (existing) plus this feature's one new dependency, `github.com/spf13/cobra` v1.10.2 — the project's first CLI framework, matching `docs/architecture-specification.md` §32's own illustrative layout (research.md).
**Storage**: Filesystem — unchanged writers. This feature introduces no new persisted state of its own; every command either reads (via `project.Detect` and 001-007's read operations) or writes exactly what `operations.Create`/`CreateArtifact` or `bootstrap.Bootstrap` already write. The one new surface is stdout/stderr itself — structured JSON out, non-zero exit codes on failure, never decorative styling (FR-010).
**Testing**: `go test` — fast, in-process tests calling `internalcmd`'s `RunE` functions and `cli.Execute` directly (capturing stdout/exit code without a subprocess), plus a smaller set of true subprocess integration tests that build the real `cmd/misterspec` binary and run it against a fixture project directory, confirming the exact JSON/exit-code contract end to end (quickstart.md). 001-007's full suites are named regression gates since this feature adds no changes to any of those packages.
**Target Platform**: Cross-platform Go module; this feature additionally produces `cmd/misterspec`, a real cross-platform binary artifact for the first time.
**Project Type**: Single Go module, now with one binary entrypoint. New packages: `internal/cli` (root/init/internal command wiring), `internal/cli/internalcmd` (the ten operation commands plus the shared envelope/error-classification helpers), `cmd/misterspec` (the `main` package). `kit/kit.go` gains one additive `SkillsFS embed.FS` (research.md) alongside the existing `TemplatesFS`.
**Performance Goals**: Trivial — every command is a single, already-fast deterministic operation; no performance-sensitive path.
**Constraints**: The `internal` command tree must never appear in `misterspec --help` (FR-005, Cobra's `Hidden: true`); every command's JSON marshaling and error classification goes through exactly one shared helper each, never duplicated per command (research.md, Constitution Principle VI); no command may contain business logic beyond argument/flag parsing and JSON shaping (FR-009); `validate`'s exit code (`4` on `valid: false`) is the one command-specific exception to the shared `classify` error-path table, computed directly from its own successful return (research.md).
**Scale/Scope**: `internal/cli/{root.go, init.go, internal.go}`; `internal/cli/internalcmd/{envelope.go, errors.go, resolve.go, inspect.go, parent.go, children.go, create.go, create_artifact.go, fingerprint.go, inventory.go, validate.go, status.go}` plus tests; `cmd/misterspec/main.go`; `kit/kit.go` modified, `kit/skills/.gitkeep` added. No Bubble Tea, no canonical Skill content, no second agent adapter, no `project`/`references` commands (research.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Every command invokes an already-deterministic operation exactly as-is — the CLI layer adds argument parsing and JSON shaping, never judgment about what an ID means or which entity to act on. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass, reinforced.** Every mutating command (`create`, `create-artifact`, `init`) calls only `operations.Create`/`CreateArtifact`/`bootstrap.Bootstrap` — the CLI layer performs no mutation of its own; it is purely an invocation surface. |
| III. Filesystem Is Single Source of Truth | **Pass.** No command caches anything between invocations — every run re-detects the project fresh via `project.Detect`, exactly like every underlying operation already does. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** Exactly the ten already-implemented operations are exposed — `project` and `references` commands are deliberately *not* added despite being named in §9/§15, because no `references` implementation exists and adding one here would be new business logic outside this feature's scope (research.md); Cobra itself is adopted specifically to avoid hand-rolling flag-parsing infrastructure this project has no reason to build itself. |
| V. Test-First Discipline | **Gate carried into tasks.** Both in-process command tests and subprocess-level binary tests are mandatory; 001-007's full suites are named regression gates since none of their packages change. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** A single shared envelope helper (`WriteSuccess`/`WriteError`) and a single shared error-classification function (`classify`) mean the JSON shape and error-code mapping are each defined exactly once, never duplicated across ten-plus commands (research.md) — the clearest DRY exercise since 006-agent-adapter's `RecordInstall`. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes nothing beyond what `operations`/`bootstrap` already write — its only new "output" is stdout/stderr, not a filesystem mutation. |
| VIII. Safety by Construction | **Pass, reinforced.** Every mutating command reuses an already-atomic underlying operation; no new write path is introduced anywhere in this feature. |
| IX. Transparent, Machine-Readable Contracts | **Pass — no longer deferred.** Every prior feature's contracts document included a "future JSON mapping" table explicitly marked as deferred; this feature is where those tables become real, tested, invocable behavior for the first time — stable error codes, category-specific exit codes, and machine-readable JSON are now the actual product surface, not a documented intention. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/008-cli-cobra/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── cli.md             # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

Matches `docs/architecture-specification.md` §32's own recommended
layout closely — this is the first feature to add both `cmd/` and a
binary entrypoint:

```text
misterspec/
├── go.mod                          # MODIFIED: + github.com/spf13/cobra
├── go.sum                          # MODIFIED
├── cmd/
│   └── misterspec/
│       └── main.go                 # NEW — calls cli.Execute(), os.Exit
├── kit/
│   ├── kit.go                      # MODIFIED: + SkillsFS embed.FS (research.md)
│   ├── templates/                  # unchanged
│   └── skills/
│       └── .gitkeep                # NEW — placeholder; content is Phase 6
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── lock/, templates/, validation/                     # unchanged
    ├── operations/, installer/, agents/, bootstrap/        # unchanged
    └── cli/                          # NEW package (US1, US2, US3)
        ├── doc.go
        ├── root.go                    # newRootCmd, Execute — bare root + hidden "internal" parent (Foundational); Hidden:true hardened + verified (US3)
        ├── internal.go                 # registers all ten internalcmd.NewXxxCmd() under the "internal" parent (US1)
        ├── init.go                      # newInitCmd (US2)
        ├── root_test.go
        ├── internal_test.go
        ├── init_test.go
        └── internalcmd/                 # NEW package (US1)
            ├── doc.go
            ├── envelope.go                # WriteSuccess, WriteError
            ├── errors.go                   # classify
            ├── resolve.go, inspect.go, parent.go, children.go
            ├── create.go, create_artifact.go
            ├── fingerprint.go, inventory.go
            ├── validate.go, status.go
            └── *_test.go                    # one per command, + envelope_test.go, errors_test.go
```

**Structure Decision**: Two new packages (`internal/cli`,
`internal/cli/internalcmd`) plus the module's first `cmd/` binary
entrypoint — file-for-file matching §32's own recommended structure.
`internalcmd` is a separate package from `cli` specifically so its ten
small command-constructor files and two shared helpers stay
independently testable and readable, rather than one large `internal.go`
(research.md). `kit.SkillsFS` is introduced now (empty embed) so the
public `init` command passes real, if currently empty, Skills data
rather than a fixture (research.md) — its *content* remains a future
feature's work.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
