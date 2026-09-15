# Implementation Plan: Feature-Level Git Branch Automation

**Branch**: `022-feature-branch-automation` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/022-feature-branch-automation/spec.md`

## Summary

`misterspec internal create feature` currently writes the new Feature's
artifact to whichever Git branch is already checked out, with no branch
management of its own — exactly the gap the user's own validation surfaced.
This feature adds a small `internal/vcs` package that shells out to the
system `git` binary to create-and-checkout (or resume) one dedicated branch
per Feature, deterministically named from the Feature's own ID, wires it
into `operations.Create` only for `Type == Feature`, and adds a matching
non-blocking mismatch warning when a Spec is created while checked out
somewhere other than its parent Feature's own branch. A new
`git_branch_automation` config field (default: on) lets a project disable
the whole behavior and fall back to exactly today's behavior.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: Standard library only (`os/exec` to shell out to
the system `git` binary) — no new Go module dependency
**Storage**: Filesystem + Git (unchanged — no new persistent state; branch
existence is itself the only "state," queried from Git each time per
Constitution Principle III)
**Testing**: `go test` — unit tests for `internal/vcs` (with a fixture Git
repo created via the `git` CLI in `t.TempDir()`) and filesystem-integration
tests for `operations.Create`'s new branch-aware behavior
**Target Platform**: Same as the rest of `misterspec` — Linux, macOS,
Windows (the `git` CLI is a widely available prerequisite already implied
by the project's own "History: Git" architecture constraint)
**Project Type**: Single Go CLI project (existing structure, no new project)
**Performance Goals**: Not applicable beyond ordinary local `git` command
latency (sub-second) — no throughput requirement
**Constraints**: Requires a `git` executable on `PATH` only when the target
project is itself a Git repository and automation is enabled; every other
case degrades gracefully with zero behavior change from today
**Scale/Scope**: Single Feature-creation call; no concurrency beyond the
existing allocation lock (`internal/lock`) already guarding `operations.Create`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Branch naming
  is a pure function of the Feature's own already-allocated ID — nothing
  here asks the agent to guess or asks the Go binary to make a judgment
  call; it is exactly the kind of "computable from repository structure"
  work Principle I assigns to the binary.
- **Principle II (Deterministic Operations Are the Only Mutation
  Primitive)**: PASS. Branch creation is folded into the existing atomic
  `operations.Create` call for `Type == Feature` — no new split
  allocate-then-branch sequence, no new public mutation entrypoint outside
  `operations`.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. No
  new persistent state is introduced; branch existence and the current
  branch are queried from Git live, every time, never cached or recorded in
  `.misterspec/`.
- **Principle IV (YAGNI & Minimal Configuration)**: PASS, with one new
  config field (`git_branch_automation`) — justified because User Story 3
  demonstrates a real, stated variance (teams with their own branching
  convention), not a hypothetical one, and it is a single boolean, not a
  new layer.
- **Principle V (Test-First Discipline)**: Applies — `internal/vcs` unit
  tests and `operations.Create` filesystem-integration tests are written
  before/alongside implementation (Phase 2 tasks).
- **Principle VI (SOLID/Clean Code)**: PASS. A new `internal/vcs` package
  owns exactly one responsibility (talking to `git`), mirroring how `ids`
  allocates and `validation` validates without blurring into `operations`,
  which stays the orchestrator.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. `internal/vcs`
  only ever creates/checks out branches it is explicitly asked to manage —
  it never touches branches, commits, or artifacts outside that scope.
- **Principle VIII (Safety by Construction)**: PASS. Every Git operation
  used (`rev-parse`, `branch --show-current`, `checkout [-b]`) is
  non-destructive — no `reset --hard`, no `push --force`, no branch
  deletion; a pre-existing branch is resumed, never overwritten; failure to
  invoke `git` degrades to "skip, don't fail" rather than corrupting
  anything.
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS. The new
  Git-related information (branch name, whether it was newly created, a
  skip reason, or a mismatch warning) is added as explicit, named JSON
  fields on `internal create`'s existing success payload — never prose.

No violations — Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/022-feature-branch-automation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── vcs/                       # NEW — thin wrapper around the system `git` binary
│   ├── vcs.go                 # IsRepo, CurrentBranch, EnsureBranch, BranchName
│   └── vcs_test.go
├── operations/
│   ├── create.go              # MODIFIED — call internal/vcs for Type == Feature/Spec
│   └── create_test.go         # MODIFIED — new branch-aware assertions
├── project/
│   ├── config.go              # MODIFIED — add GitBranchAutomation bool (default true)
│   └── config_test.go         # MODIFIED
└── cli/internalcmd/
    ├── create.go               # MODIFIED — surface new git.* fields in JSON payload
    └── create_test.go          # MODIFIED
```

**Structure Decision**: Single existing Go CLI project — no new project or
module boundary. One new leaf package (`internal/vcs`) added alongside the
existing `ids`/`validation`/`installer` siblings, consumed only by
`internal/operations`, per Principle VI's own SRP guidance.
