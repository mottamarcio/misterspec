# Implementation Plan: Interactive Clarification and Richer Output for Product Skills

**Branch**: `024-product-skills-clarify-format` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/024-product-skills-clarify-format/spec.md`

## Summary

Comparing upstream `github/spec-kit` against misterspec's own product Skills
(`kit/skills/`) surfaced that ambiguity in `create-specs`/`create-plan` is
resolved silently (narrowed and annotated) rather than asked about, that
`create-tasks` references a Requirement's own constraint only by ID (risking
drift), that `analyze` can only recommend remediation in prose with no
tracked follow-up, and that every Skill's own completion report — despite
already having a structured `Completion Contract` naming five distinct
buckets of information — renders as prose at runtime rather than a
scannable table/bullet format. This feature updates 5 of the 9
`kit/skills/*/SKILL.md` files (`create-specs`, `create-plan`, `create-tasks`,
`analyze` get behavioral changes; all 9 get the formatting change) — text
only, no Go code, no template schema change (verified during specification
that the Spec artifact already has an `## Unresolved Questions` section to
reuse).

## Technical Context

**Language/Version**: N/A — Markdown Skill definitions (YAML frontmatter +
prose instructions); no Go code touched
**Primary Dependencies**: None new — reuses `internal validate`/`internal
inspect`/`internal create-artifact`, all already used by these Skills
**Storage**: N/A — no new artifact type or persisted state; User Story 4's
Task-append reuses the existing Tasks-artifact-is-a-directly-authored-file
convention `create-tasks` itself already relies on (no independent Task-ID
allocator exists — confirmed in `create-tasks/SKILL.md`'s own `Allowed
Modifications`)
**Testing**: No unit-test framework applies to instruction text; each user
story's own manual verification (drawn from `quickstart.md`) is this
feature's adapted form of Constitution Principle V, matching the approach
already used for spec 023 (the analogous dev-tooling-only feature)
**Target Platform**: Any coding agent capable of installing and following
misterspec's own `kit/skills/` Skills in an end user's project
**Project Type**: Prompt-engineering / instruction-text change to the
product misterspec ships — no application project structure applies
**Performance Goals**: N/A
**Constraints**: Changes are scoped to `kit/skills/*/SKILL.md` only; no
`internal/`/`cmd/` Go package, and no `kit/templates/*.tmpl`, is touched
(verified during specification that the one template field this feature
might have needed, an open-questions field on Spec, already exists)
**Scale/Scope**: 4 Skills gain new behavioral rules (`create-specs`,
`create-plan`, `create-tasks`, `analyze`); all 9 Skills gain the same
formatting instruction at one identical anchor point each

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS, and directly
  advanced. Resolving a genuine requirement/strategy ambiguity is exactly
  the kind of judgment call Principle I assigns to the agent, never the Go
  binary — this feature makes that judgment visible to the user (asking)
  instead of silent, but does not move the decision itself into
  deterministic code.
- **Principle II (Deterministic Operations Are the Only Mutation
  Primitive)**: PASS. User Story 4's Task-append reuses the exact same
  direct-file-authoring convention `create-tasks` already uses for the same
  artifact — no new mutation primitive, no new `internal` command.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. No
  new persisted state; the "Unresolved Questions" section already exists in
  the Spec artifact and is read/written the same way any other section is.
- **Principle IV (YAGNI & Minimal Configuration)**: PASS. Zero new
  configuration fields, zero new Skills, zero new artifact/template
  sections — every change either adds an instruction to an existing
  `SKILL.md` section or reuses an already-existing template field.
- **Principle V (Test-First Discipline)**: Adapted — manual verification
  per story, matching spec 023's own precedent for instruction-only
  features with no Go code to unit-test.
- **Principle VI (SOLID/Clean Code)**: N/A — no Go code.
- **Principle VII (Explicit Mutation Boundaries)**: PASS, and the single
  most constitution-sensitive point in this feature. User Story 4
  deliberately, narrowly widens `analyze/SKILL.md`'s own `Allowed
  Modifications` (today: "The Validation artifact's own content, on a
  re-run") to add exactly one new permitted write — appending one `##
  TASK-NNN` entry to the Tasks artifact — gated by three independent
  constraints so the boundary stays narrow rather than open-ended: (a) only
  for a finding whose responsible layer is "implementation incomplete"
  (never Spec/Plan-layer findings); (b) only on the user's explicit,
  per-finding confirmation (never automatic); (c) append-only — the Spec,
  Plan, and every pre-existing Task remain untouched. This is the exact
  same append-only discipline already validated for the dev-tooling
  `speckit-converge` skill (spec 023), applied here to the product's own
  `analyze` Skill instead.
- **Principle VIII (Safety by Construction)**: PASS. No destructive
  operation is introduced anywhere; User Story 4's only write is additive
  and user-confirmed.
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS, and
  directly advanced by User Story 5 — every Skill's own `Completion
  Contract` already names five distinct informational buckets (`Outcome`,
  `Artifacts`, `Important findings`, `Attention`, `Recommended next step`);
  this feature makes their *rendering* match that same structure (tables/
  bullet lists) instead of letting runtime output flatten them back into
  undifferentiated prose.

No violations — Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/024-product-skills-clarify-format/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source (repository root — product Skills only, no Go code)

```text
kit/skills/
├── create-specs/SKILL.md         # MODIFIED — US1 (interactive clarification), US5 (formatting)
├── create-plan/SKILL.md          # MODIFIED — US2 (interactive clarification), US5 (formatting)
├── create-tasks/SKILL.md         # MODIFIED — US3 (constraint verbatim), US5 (formatting)
├── analyze/SKILL.md              # MODIFIED — US4 (confirmed remediation Task), US5 (formatting)
├── create-program/SKILL.md       # MODIFIED — US5 (formatting) only
├── create-feature/SKILL.md       # MODIFIED — US5 (formatting) only
├── create-knowledge-base/SKILL.md # MODIFIED — US5 (formatting) only
├── create-constitution/SKILL.md  # MODIFIED — US5 (formatting) only
└── implement/SKILL.md            # MODIFIED — US5 (formatting) only
```

**Structure Decision**: No application project structure applies — every
change is a Markdown instruction edit inside `kit/skills/`, the same
directory whose content is embedded into the `misterspec` binary via
`go:embed` and installed verbatim into an end user's project. Nothing
outside this directory changes.
