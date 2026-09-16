# Implementation Plan: Dual-Mode Implement — All Tasks or One Named Task

**Branch**: `026-implement-single-task` | **Date**: 2026-09-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/026-implement-single-task/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Give the native `/implement` Skill two invocation forms instead of one: `SPEC-###` alone keeps working the way it conceptually always has at the Spec level — sequentially implementing and verifying every currently executable task in that Spec within one invocation, without pausing to ask the user to re-issue the command between tasks (the "Antigravity" behavior the user wants to keep). `SPEC-### TASK-NNN` is new: it implements only that one named task and stops, leaving every other task in the Spec untouched (the "Claude" behavior the user wants to keep as an option for tighter control). Since Task IDs are numbered per-Spec (no global allocator), the Spec identifier is mandatory in both forms. No new deterministic operation is needed — task lookup, dependency checks, and sequencing all read data the Skill already loads via `internal inspect` / `internal context`. The change is confined to `kit/skills/implement/SKILL.md` (primary) plus two small consistency edits in Skills that recommend `/implement` as a next step (`create-tasks`, `analyze`).

## Technical Context

**Language/Version**: Go 1.23.4 (repo-wide); the artifacts under change here are Markdown + YAML frontmatter Skill prompts embedded via `go:embed`, not Go source.
**Primary Dependencies**: `kit.SkillsFS` (embedded skill content, `kit/kit.go`), the Claude/agent adapters under `internal/agents/` that render these Skills for each integration.
**Storage**: N/A — Skill content is plain files under `kit/skills/`; no database or persistent state introduced. No "active Spec" or "active task" state is tracked between invocations (would violate Constitution Principle III).
**Testing**: `go test ./internal/example/...` — specifically `skills_content_test.go` (structural/contract conformance: frontmatter, required §39 headings in order, Deterministic Operations allowlist, Completion Contract concepts) and `multi_agent_skill_integration_quickstart_test.go` (cross-agent rendering).
**Target Platform**: Agent slash-command surface (Claude Code and other integrations consuming the embedded kit), not a standalone runtime.
**Project Type**: Single Go project with an embedded content kit (`kit/skills/`) — no frontend/backend split.
**Performance Goals**: N/A (prompt content change, not a runtime hot path).
**Constraints**: Must preserve `kit/skills/implement/SKILL.md`'s existing required section set and order (`skills_content_test.go`'s `requiredSkillHeadings`); must keep every `` `internal <op>` `` mention inside this Skill's existing allowlist (`resolve, inspect, context, validate` — no new operation is being added); the Skill's existing "one task verified at a time, never batched unverified" discipline (its current Interaction Rules) must be preserved even in all-tasks mode — sequential means one-verified-task-after-another, not a single unverified multi-task change.
**Scale/Scope**: Three Skill files touched — `kit/skills/implement/SKILL.md` (primary: Invocation, Procedure, Decision/Interaction Rules, Failure Conditions, Completion Contract/Recommended Next Step) and `kit/skills/create-tasks/SKILL.md` + `kit/skills/analyze/SKILL.md` (each has one `/implement SPEC-###` example string, still valid as-is under the new design since bare `SPEC-###` remains supported — verify no wording implies "single task only" that would now be misleading) — plus one additive Go test assertion covering both invocation forms.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I/II (Semantic/Deterministic Separation, Deterministic Ops as sole mutation primitive)**: PASS. No new mutation path is introduced; the Skill continues to use `internal inspect` / `internal context` (already-loaded task data) to determine executable tasks, dependency state, and — in named-task mode — task existence. No new `internal` command, no LLM-invented ID/path logic.
- **Principle III (Filesystem is Source of Truth)**: PASS. All-tasks mode determines "next executable task" by re-reading `tasks.md`'s current checkbox/dependency state on each step within the same invocation — it does not cache or persist a work queue outside the file. No "active Spec" tracking is introduced between invocations.
- **Principle IV (Simplicity First / YAGNI)**: PASS. No new package, config field, or `internal` command. Two invocation forms are handled by one Skill via a single optional trailing argument, not a second Skill or a new config toggle.
- **Principle V (Test-First, NON-NEGOTIABLE)**: Applies at reduced scope — no new deterministic Go logic is added, but the Skill's own machine-checked contract (`skills_content_test.go`) still gates content changes. Phase 1/2 adds an assertion, written before the corresponding `SKILL.md` edit lands, that the `Invocation` section documents both forms (Spec-only and Spec+Task).
- **Principle VII (Explicit Mutation Boundaries)**: PASS. In both modes the Skill still only modifies the Tasks artifact's own completion checkboxes/evidence for the task(s) actually implemented in that run; it still must not rewrite the Spec or Plan.
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS. Completion Contract keeps naming Outcome/Artifacts/Findings/Attention/Recommended Next Step in both modes; all-tasks mode's Outcome additionally enumerates every task implemented (and any skipped as already-complete) in that run, and named-task mode's Recommended Next Step now also mentions the Spec-only form as an option.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/026-implement-single-task/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
kit/skills/
├── implement/
│   └── SKILL.md          # Primary change: Invocation (two forms), Procedure (branch on
│                          # whether a task ID was given), Decision/Interaction Rules
│                          # (sequential-but-verified in all-tasks mode), Failure Conditions
│                          # (blocked-task reporting in all-tasks mode), Completion Contract
├── create-tasks/
│   └── SKILL.md           # Verify its "/implement SPEC-###" example still reads correctly
│                          # under the dual-mode design (no wording change expected)
└── analyze/
    └── SKILL.md            # Same verification as create-tasks

internal/example/
└── skills_content_test.go  # Add assertion: implement's Invocation section documents
                             # both the Spec-only and the Spec+Task forms
```

**Structure Decision**: Single Go project, embedded-content model already in use (`kit/skills/*/SKILL.md` sourced by `kit.SkillsFS`, validated by `internal/example/skills_content_test.go`). No new directories or packages — this feature edits existing Skill content files in place and extends an existing conformance test.

## Complexity Tracking

*No Constitution Check violations — table not needed.*
