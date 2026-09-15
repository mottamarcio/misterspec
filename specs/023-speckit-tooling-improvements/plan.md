# Implementation Plan: Spec-Kit Tooling Improvements

**Branch**: `023-speckit-tooling-improvements` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/023-speckit-tooling-improvements/spec.md`

## Summary

Comparing upstream `github/spec-kit`'s `templates/` against this repository's own
`.specify/` dev-tooling surfaced 8 concrete gaps, one of them a demonstrated bug
(a mandatory hook that can be described without actually being invoked — the
exact ambiguity that let this session's own git-branch-creation hook go unrun
for several prior features). This feature ports the relevant upstream text and
mechanics into our own `.claude/skills/speckit-*/SKILL.md` files and
`.specify/templates/checklist-template.md`, adds one new skill
(`speckit-converge`), and adds two new hook keys to `.specify/extensions.yml`
for it — all text/config changes, no Go code, no changes to `kit/skills/` (the
product misterspec ships to its own end users).

## Technical Context

**Language/Version**: N/A — Markdown skill definitions (YAML frontmatter +
Markdown instructions) and YAML configuration; no Go code touched
**Primary Dependencies**: None new — reuses the GitHub MCP server tools
(`list_issues`, `issue_write`) `speckit-taskstoissues` already depends on
**Storage**: N/A — no new files beyond one new skill and its own directory
**Testing**: No unit-test framework applies to prompt/instruction text;
verification is manual — invoke each modified skill against a real,
disposable Spec-Kit-managed feature and confirm the observable behavior
change (a hook's effect actually exists; a checklist file is byte-identical
before/after; a second `taskstoissues` run creates zero new issues; etc.)
**Target Platform**: Claude Code (or any other `/speckit-*`-compatible
coding agent) operating on this repository's own `.specify/` tooling
**Project Type**: Dev-tooling / prompt-engineering change — no application
project structure applies
**Performance Goals**: N/A
**Constraints**: Changes are scoped to `.claude/skills/speckit-*/SKILL.md`,
`.specify/templates/checklist-template.md`, and `.specify/extensions.yml`
only — `kit/skills/` and every `internal/`/`cmd/` Go package are explicitly
out of scope (spec.md's own Assumptions)
**Scale/Scope**: 9 existing skill files edited, 1 new skill file created
(`speckit-converge`), 1 template file edited, 1 config file edited (2 new
hook keys)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: N/A — this feature
  edits agent-facing instructions, not the Go binary's own deterministic
  operations; no boundary between semantic and mechanical work is affected.
- **Principle II (Deterministic Operations Are the Only Mutation
  Primitive)**: N/A for the same reason — no new mutation primitive is
  introduced in the Go binary. `speckit-converge`'s own append-only
  discipline (FR-009) is this same principle's *spirit* applied to this
  repo's dev-tooling: one narrow, explicit write path, never an implicit
  rewrite.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS —
  every check this feature adds (hook effect verification, checklist
  read-only-ness, issue deduplication, convergence gap detection) is
  computed live from the current state of files/Git/GitHub, never a new
  persisted counter or cache.
- **Principle IV (YAGNI & Minimal Configuration)**: PASS, with one new
  skill (`speckit-converge`) and two new `extensions.yml` hook keys
  (`before_converge`/`after_converge`) — justified because User Story 6
  is a real, demonstrated capability gap (no drift-detection tool exists
  today), not a hypothetical one, and the two new hook keys mirror the
  exact pattern every other command already uses rather than inventing a
  new mechanism.
- **Principle V (Test-First Discipline)**: Adapted, not literally
  applicable — there is no Go code to unit-test. The adapted discipline is
  "manually verify the *previous*, unfixed behavior reproduces the gap
  before editing, then verify the fix" (documented per-story in
  quickstart.md), preserving the same before/after discipline in spirit.
- **Principle VI (SOLID/Clean Code)**: N/A — no Go code.
- **Principle VII (Explicit Mutation Boundaries)**: PASS, and directly
  advanced — User Story 2 (checklists), User Story 4 (constitution scope
  guard), and User Story 6 (`speckit-converge`'s append-only contract) are
  all instances of this exact principle, ported from upstream into skills
  that didn't yet state it as explicitly as they should.
- **Principle VIII (Safety by Construction)**: PASS — `speckit-converge`
  never writes application code or deletes anything; its only write is a
  strictly-additive `tasks.md` append, and it leaves the file byte-for-byte
  unchanged when there is nothing to add.
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS, and
  directly advanced — User Story 1's FR-002 (visible hook-parse failures)
  and User Story 6's severity-graded findings table are both instances of
  "never leave an ambiguous, unparseable outcome," this principle's own
  core concern, applied to this repo's dev-tooling rather than the Go
  binary's JSON contracts.

No violations — Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/023-speckit-tooling-improvements/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command) — hook/skill contract, not an API
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source (repository root — dev-tooling only, no Go code)

```text
.claude/skills/
├── speckit-specify/SKILL.md        # MODIFIED — US1 (hook boilerplate), US5 (Done When)
├── speckit-plan/SKILL.md           # MODIFIED — US1, US5, US8 (quickstart scope)
├── speckit-tasks/SKILL.md          # MODIFIED — US1, US5, US8 (constraint-verbatim)
├── speckit-implement/SKILL.md      # MODIFIED — US1, US2 (checklist read-only), US5
├── speckit-clarify/SKILL.md        # MODIFIED — US1, US7 (question quality, checklist re-validation)
├── speckit-checklist/SKILL.md      # MODIFIED — US1, US2 (unchecked-by-default)
├── speckit-constitution/SKILL.md   # MODIFIED — US1, US4 (Scope Guard)
├── speckit-taskstoissues/SKILL.md  # MODIFIED — US1, US3 (deduplication)
└── speckit-converge/SKILL.md       # NEW — US6

.specify/
├── templates/checklist-template.md # MODIFIED — US2 (Marker Semantics / Review Ownership)
└── extensions.yml                  # MODIFIED — US6 (before_converge/after_converge hook keys)
```

**Structure Decision**: No application project structure applies — every
change is a Markdown skill definition or YAML config edit in this
repository's own `.specify`/`.claude` dev-tooling tree, mirroring the file
layout upstream's own `templates/commands/` already uses one-to-one.
