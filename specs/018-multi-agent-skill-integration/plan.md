# Implementation Plan: Multi-Agent Skill Integration

**Branch**: `018-multi-agent-skill-integration` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/018-multi-agent-skill-integration/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Two independent deliverables, both scoped by spec.md: (1) five new
`agents.Adapter` implementations — `agy` (Antigravity), `codex` (Codex
CLI), `copilot` (GitHub Copilot), `cursor-agent` (Cursor), `devin`
(Devin for Terminal) — each a near-verbatim copy of the existing
`claude` adapter (006-agent-adapter) differing only in `id`/`name`/
`targetPath`, since research confirmed all five agents already scan
the same directory-per-skill `SKILL.md` "Agent Skills" convention
misterspec's own canonical Skills already use; and (2) four of
misterspec's own canonical Skill files (`implement`, `create-plan`,
`create-tasks`, `analyze`) gain one new early step in their existing
`Procedure`/`Deterministic Operations` sections directing the agent to
request an intent-appropriate Context Pack (via 017's `internal
context` command) before broader exploration, while explicitly
remaining free to exceed it. No change to the Context Engine (011-017)
itself; no change to any other Skill; no new CLI command.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged) for the five new adapters; Markdown/YAML-frontmatter (misterspec's own canonical Skill format, unchanged) for the four updated Skill files.
**Primary Dependencies**: None new — the five adapters reuse `internal/installer.InstallFS` and `internal/agents.RecordInstall`, both already exported and already used by `claude`; the four Skill files use only the already-existing `internal context` command (017).
**Storage**: None new — adapters write into each agent's own project-relative integration directory using the same atomic-write mechanism `claude`'s adapter already uses; no new persisted state.
**Testing**: `go test` — new `internal/agents/{agy,codex,copilot,cursoragent,devin}` package tests mirroring `internal/agents/claude/claude_test.go`'s own shape (ID/Name/TargetPath, successful install, no-clobber-on-overwrite-false, `agents.RecordInstall` called); `internal/agents/builtin/builtin_test.go` extended to assert all six IDs are present and multi-agent installs don't collide; a new `kit` golden-fixture check (or extension of an existing one) confirming each of the four updated `SKILL.md` files still parses as valid frontmatter and still contains its own required sections. Per Constitution Principle V. 001-017's full suites re-run unmodified as the named regression gate.
**Target Platform**: Cross-platform Go module, unchanged; the five new integration directories (`.agents/skills`, `.github/skills`, `.cursor/skills`, `.devin/skills`) are plain project-relative paths, no agent-specific runtime dependency.
**Project Type**: Single Go module (CLI) plus embedded content (`kit/skills`). Five new small Go packages; one line changed in `internal/agents/builtin/builtin.go`; four existing Markdown files edited in place.
**Performance Goals**: Each new adapter's `Install` is the same `O(n)` byte-copy over `n` Skill files `claude`'s own adapter already performs — no new cost. The four updated Skills add one `internal context` invocation to each Skill's own already-existing runtime, bounded by 017's own established performance characteristics.
**Constraints**: Every new adapter MUST NOT alter Skill content or meaning (FR-003, Constitution Principle VII/§35) — content is copied byte-for-byte, exactly as `claude`'s adapter already does. Installing for one agent MUST NOT disturb another agent's own installed integration (FR-004) — each writes only under its own `targetPath`, and `installer.InstallFS`'s own existing containment/no-clobber guarantees are reused unchanged. Every updated Skill MUST keep the agent free to exceed the Context Pack (FR-007) and MUST fall back gracefully on a failed request (FR-008) — both are instruction-text requirements, not code paths, verified by inspecting the Skill file's own content.
**Scale/Scope**: `internal/agents/{agy,codex,copilot,cursoragent,devin}/{<name>.go,<name>_test.go}` (10 new files, ~20 lines of production code each); `internal/agents/builtin/builtin.go` (+5 lines) and `builtin_test.go` (extended); `kit/skills/{implement,create-plan,create-tasks,analyze}/SKILL.md` (4 files edited in place, each gaining ~4-6 lines across two existing sections).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** The five adapters perform zero judgment — mechanical byte-copy into a fixed, agent-specific directory, identical to `claude`'s own already-approved adapter. The four Skill edits add a mechanical `internal context` call as a new first step; the agent's own semantic judgment about what to do with the returned pack, and whether to look further, is explicitly preserved (FR-007) — never replaced by anything mechanical. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass.** No adapter invents an ID, path, or hash — `installer.InstallFS` (already-existing, already-deterministic) performs every write; Skill files gain a call to `internal context`, an already-existing deterministic operation, never a new one. |
| III. Filesystem Is Single Source of Truth | **Pass (N/A new surface).** Adapters write only into each agent's own non-authoritative integration directory (mirroring `.claude/skills`'s own established non-authoritative status); no new persisted state, no new database. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** Research (research.md #1-#2) confirmed zero content transformation is needed for any of the five agents once their real, current integration conventions were verified — an earlier, more complex "flattening adapter" design was discarded specifically because it was unneeded speculative surface once the facts were confirmed. No new configuration field is added; each `targetPath` is a fixed internal constant, exactly `claude`'s own precedent. |
| V. Test-First Discipline | **Gate carried into tasks.** Each new adapter gets its own test file mirroring `claude_test.go`'s coverage; `builtin_test.go` gains multi-agent assertions; the four updated Skill files get a golden-fixture-style structural check. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** Every new adapter reuses `installer.InstallFS`/`agents.RecordInstall` verbatim rather than reimplementing install logic five more times (DRY); one small file per adapter keeps each package's own single responsibility identical to `claude`'s (SRP). |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** No adapter may alter Skill semantics (§35, unchanged); the four updated Skills still only write what they already wrote before (their own artifact type) — the new step only reads (`internal context` is read-only, 017 FR-008). |
| VIII. Safety by Construction | **Pass (N/A new surface).** Each adapter's `targetPath` is a fixed, hardcoded constant joined against `project.Detect`'s own already-validated root — no new user-supplied path is ever accepted. |
| IX. Transparent, Machine-Readable Contracts | **Pass (N/A this feature).** No new CLI command is added; the five adapters are discoverable through the existing `agents.Registry` contract 006 already established. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/018-multi-agent-skill-integration/
├── plan.md                          # This file (/speckit-plan command output)
├── research.md                      # Phase 0 output (/speckit-plan command)
├── data-model.md                    # Phase 1 output (/speckit-plan command)
├── quickstart.md                    # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── adapters-and-skills.md       # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md              # /speckit-specify quality checklist
└── tasks.md                         # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                          # unchanged — no new dependency
├── cmd/misterspec/, internal/cli/            # unchanged — no CLI surface added
├── internal/
│   ├── installer/, project/, ids/              # unchanged — reused, not modified
│   ├── context/                                  # unchanged — this feature is a new caller, not a modifier
│   └── agents/
│       ├── adapter.go, registry.go, record.go      # unchanged (006) — reused, not modified
│       ├── claude/                                   # unchanged — the pattern every new adapter copies
│       ├── agy/agy.go, agy_test.go                     # NEW
│       ├── codex/codex.go, codex_test.go                 # NEW
│       ├── copilot/copilot.go, copilot_test.go              # NEW
│       ├── cursoragent/cursoragent.go, cursoragent_test.go     # NEW
│       ├── devin/devin.go, devin_test.go                          # NEW
│       └── builtin/
│           ├── builtin.go                                          # MODIFIED — +5 registrations
│           └── builtin_test.go                                       # MODIFIED — +multi-agent coverage
└── kit/
    └── skills/
        ├── implement/SKILL.md                                          # MODIFIED — +Context Pack step
        ├── create-plan/SKILL.md                                          # MODIFIED — +Context Pack step
        ├── create-tasks/SKILL.md                                           # MODIFIED — +Context Pack step
        └── analyze/SKILL.md                                                  # MODIFIED — +Context Pack step
```

**Structure Decision**: Five new one-file-per-adapter packages under
`internal/agents/`, mirroring `claude`'s own existing shape exactly
(research.md #2, #4); one small registration change in `builtin.go`;
four existing canonical Skill files edited in place with the smallest
change consistent with FR-009/FR-010 (research.md #6). No new package,
no new CLI surface, no change to the Context Engine.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
