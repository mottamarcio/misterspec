# Implementation Plan: Constitution Frontmatter Guarantee, and Task Dependency/Parallelism Reporting

**Branch**: `027-constitution-frontmatter-task-deps` | **Date**: 2026-09-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/027-constitution-frontmatter-task-deps/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Two independent, additive changes bundled into one spec per the user's request. **User Story 1** closes the Constitution frontmatter gap two ways: (a) `create-constitution/SKILL.md`'s own Outputs/Procedure/Validation Rules gain explicit instructions to write and preserve the `type: constitution`/`schema_version: 1` frontmatter §24 already requires, and (b) — the stronger, model-independent guarantee — `internal validate` (which `create-constitution` already calls at the end of its own Procedure) is extended with a new, small, read-only structural check for the Constitution file, so a missing/malformed frontmatter is caught deterministically regardless of what the underlying model did, not left purely to prompt-following. This reuses `internal/artifacts.ParseMetadata` and existing Finding codes (`CodeFrontmatterMalformed`, `CodeRequiredFieldMissing`) — no new CLI command, no new Finding vocabulary. **User Story 2** is pure Skill-prompt content: `create-tasks/SKILL.md`'s own Completion Contract gains explicit dependency/parallel-group reporting, computed from data the Skill already has (each Task's own recorded dependency, already required by its existing Outputs) — no new Go code.

## Technical Context

**Language/Version**: Go 1.23.4 (repo-wide). `internal/artifacts` and `internal/validation` changes are Go; `create-constitution`/`create-tasks` changes are Markdown + YAML frontmatter prompt content embedded via `go:embed`.
**Primary Dependencies**: `internal/artifacts` (frontmatter parsing — `ParseMetadata`, `Metadata`, extended with a `SchemaVersion` field), `internal/validation` (`ValidateProject`'s existing Finding-aggregation pipeline, extended with one new Constitution-specific check function), `kit.SkillsFS` (embedded Skill content).
**Storage**: N/A — reads `ai/memory/constitution.md` and `tasks.md`; no database, no new persisted state.
**Testing**: `go test ./internal/artifacts/...` (new `SchemaVersion` field parsing), `go test ./internal/validation/...` (new Constitution check: missing file → no finding since it may not exist yet per Preconditions; present-but-malformed → `CodeFrontmatterMalformed`; present-but-missing-`schema_version` → `CodeRequiredFieldMissing`; present-and-correct → no finding), `go test ./internal/example/...` (`skills_content_test.go`, extended with assertions that `create-constitution`'s Outputs/Validation Rules mention the required frontmatter fields, and that `create-tasks`'s Completion Contract mentions dependency/parallel reporting).
**Target Platform**: Same as always — the hidden `misterspec internal …` command surface (existing `internal validate`, unchanged interface) plus agent Skill prompt content.
**Project Type**: Single Go project with an embedded content kit — unchanged.
**Performance Goals**: N/A — one additional lightweight file read/parse per `internal validate` run, same order of cost as the existing per-entity checks.
**Constraints**: MUST NOT alter `ValidateProject`'s existing behavior for Program/Feature/Spec/Knowledge/Learning (`004-structural-validation`'s own `projectEntityTypes`/`checkEntity` stay untouched) — the new Constitution check is a separate, additive function, not a modification of `checkEntity`'s existing logic (Constitution has no `EntityType`/ID and would break `checkEntity`'s ID-location-mismatch assumption if forced through it). MUST treat a *missing* Constitution file as "nothing to check" (not a Finding) — `create-constitution`'s own Preconditions already allow first-run creation, so an absent file is not yet a structural problem. MUST reuse existing Finding codes (`CodeFrontmatterMalformed`, `CodeRequiredFieldMissing`) rather than inventing new ones, per `findings.go`'s own established vocabulary. `create-tasks`'s dependency/parallel reporting MUST NOT require any new deterministic operation — it is prose the Skill composes from data (`depends_on`, `[P]` markers) its own existing Outputs already require it to record.
**Scale/Scope**: One new field on `internal/artifacts.Metadata`/its internal YAML shape; one new, small, read-only check function in `internal/validation` wired into `ValidateProject`'s existing aggregation; `create-constitution/SKILL.md`'s Outputs, Procedure, and Validation Rules sections; `create-tasks/SKILL.md`'s Completion Contract section; `internal/example/skills_content_test.go` extended with two new assertions.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I/II (Semantic/Deterministic Separation, Deterministic Ops as sole mutation primitive)**: PASS. "Does this file's frontmatter contain the two required fields" is purely mechanical — exactly what `internal validate` already exists to check for every other artifact type; this closes the one gap where that mechanical check was missing, rather than leaving it to model behavior.
- **Principle III (Filesystem is Source of Truth)**: PASS. The check reads the Constitution file live on each `internal validate` run — no caching, no new persisted state.
- **Principle IV (Simplicity First / YAGNI)**: PASS, and notably simpler than prior features this session — no new CLI command, no new config field, no new Finding code. `create-constitution` already calls `internal validate` in its own existing Procedure (step 5); this feature only makes that existing call more complete.
- **Principle V (Test-First, NON-NEGOTIABLE)**: Applies fully here, unlike the prompt-only features earlier this session — this is real deterministic Go logic (a frontmatter field check) and gets real unit tests in `internal/validation`, written to fail against a malformed/incomplete fixture before the check exists, per the project's existing `004-structural-validation` test conventions.
- **Principle VI (Clean Code & SOLID, DRY)**: PASS. Reuses `internal/artifacts.ParseMetadata` rather than writing a second, parallel frontmatter parser for Constitution specifically.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. `internal validate` only reads and reports — it does not write to or repair the Constitution file itself; `create-constitution`'s own Procedure remains the only thing that writes it, per its own existing Allowed Modifications.
- **Principle VIII (Safety by Construction)**: PASS. Read-only structural check; no mutation, so no destructive-operation risk at all.
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS. Reuses the exact same `Finding{Code, Severity, Path, Message}` shape every other check already produces — an agent reading `internal validate`'s JSON output needs no new parsing logic to understand a Constitution finding.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/027-constitution-frontmatter-task-deps/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/artifacts/
└── parser.go               # + SchemaVersion field on frontmatterYAML/Metadata
                            # (constitution-only field, parsed the same way
                            # existing optional fields like Parent already are)

internal/validation/
└── validator.go             # + checkConstitution(root) []Finding — reads
                            # ai/memory/constitution.md if it exists (no
                            # finding if absent), reports
                            # CodeFrontmatterMalformed /
                            # CodeRequiredFieldMissing as appropriate;
                            # wired into ValidateProject's existing
                            # aggregation, alongside (not inside)
                            # projectEntityTypes' per-entity loop

kit/skills/
├── create-constitution/SKILL.md  # + Outputs: explicit required-frontmatter
│                                # block shown verbatim (mirrors how the
│                                # Quality Requirements baseline is already
│                                # shown verbatim); + Procedure step:
│                                # preserve/add frontmatter before writing
│                                # body; + Validation Rules: frontmatter
│                                # requirement stated explicitly
└── create-tasks/SKILL.md         # Completion Contract: + explicit
                                 # dependency-relationship and
                                 # parallel-safe-group reporting, computed
                                 # from data already in Outputs

internal/example/
└── skills_content_test.go  # + assertion: create-constitution's Outputs/
                            # Validation Rules sections mention the
                            # required frontmatter fields; + assertion:
                            # create-tasks's Completion Contract mentions
                            # dependency/parallel reporting
```

**Structure Decision**: Single Go project, embedded-content model already in use. This feature is the most code-light of the session so far for its Go portion (one struct field, one new check function wired into an existing aggregation) and purely additive to two existing Skill files otherwise. No new package, no new CLI command.

## Complexity Tracking

*No Constitution Check violations — table not needed.*
