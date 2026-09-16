# Implementation Plan: `mister-`-Prefixed Skill Names to Avoid Slash-Command Collisions

**Branch**: `028-mister-prefixed-skill-names` | **Date**: 2026-09-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/028-mister-prefixed-skill-names/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Rename the 9 directories under `kit/skills/` per the confirmed mapping (`analyze`→`mister-analyze`, `create-constitution`→`mister-constitution`, `create-feature`→`mister-features`, `create-knowledge-base`→`mister-knowledge-base`, `create-plan`→`mister-plan`, `create-program`→`mister-program`, `create-specs`→`mister-specify`, `create-tasks`→`mister-tasks`, `implement`→`mister-implement`), then update every place that names still refer to the old ones: each Skill's own frontmatter `name:` and `## Invocation`, every cross-reference any of the 9 makes to any other (a full mesh — every Skill mentions at least two others), `internal/example/skills_content_test.go`'s three hardcoded structures (`skillOperationsAllowlist` map keys, `canonicalSkillNames` slice, and three feature-specific tests from specs 026/027 that read a literal `"<name>/SKILL.md"` path), and `docs/architecture-specification.md` §38's Canonical Skill Set listing. No Go source logic changes at all — `kit/kit.go`'s `//go:embed skills` embeds the whole directory tree unconditionally, and no agent adapter (`internal/agents/*`) transforms a directory name before installing it, confirmed by investigation before this plan was written: installed command name = `kit/skills/` directory name, verbatim, with no translation layer anywhere.

## Technical Context

**Language/Version**: Go 1.23.4 (repo-wide, unaffected — no new Go logic). The actual changes are Markdown + YAML frontmatter content (9 renamed directories, their own cross-references) plus one Go test file's literal string constants and one documentation section.
**Primary Dependencies**: `kit.SkillsFS` (`kit/kit.go`'s `//go:embed skills` — needs no code change, embeds whatever directories exist under `kit/skills/` at build time), `internal/example/skills_content_test.go` (the sole Go file with hardcoded Skill-name string literals).
**Storage**: N/A.
**Testing**: `go test ./internal/example/...` — specifically `TestSkillsContent_AllNineInstalled` (already parametrized entirely by `canonicalSkillNames`; updating that one slice makes its real end-to-end install-and-byte-compare assertions verify the new names automatically), plus every `TestSkillsContent_*` conformance test (parametrized by `assertSkillConformant(t, "<name>")` calls and the `skillOperationsAllowlist` map), plus the three feature-specific tests (`TestSkillsContent_ImplementDualInvocation`, `TestSkillsContent_ConstitutionFrontmatterDocumented`, `TestSkillsContent_CreateTasksReportsDependenciesAndParallelism`) that each read one literal `"<old-name>/SKILL.md"` path directly.
**Target Platform**: Same as always — the hidden `misterspec internal …` / agent slash-command surface. No new platform surface.
**Project Type**: Single Go project with an embedded content kit — unchanged, no new package.
**Performance Goals**: N/A.
**Constraints**: MUST rename all 9 together, consistently — a partial rename would leave the cross-reference mesh pointing at a mix of old and new names, exactly the stranded-mid-pipeline failure mode User Story 2 exists to prevent. MUST NOT provide any alias/dual-registration for an old name (spec.md FR-004, confirmed with the user). MUST NOT touch `.claude/skills/speckit-*` (this repo's own separate, unrelated internal dev-tooling Skill set — a different lineage, not shipped to end users). MUST NOT rewrite already-completed historical Specs (009, 018, and every other existing spec under `specs/`) that reference an old name as part of their own historical record — only currently-active `kit/skills/` content and its own automated verification are in scope. MUST NOT "fix" `docs/architecture-specification.md` §38's separate, pre-existing, unrelated drift ("Canonical source: `.misterspec/skills/`" vs. the actual `kit/skills/`) while touching that section — out of scope, a different problem, noted only for awareness during investigation.
**Scale/Scope**: 9 directory renames (`git mv`); a full pass through all 9 renamed `SKILL.md` files correcting every cross-reference (per the pre-plan investigation: ~72 total slash-command-string occurrences across the 9 files, self-references and cross-references combined); one Go test file (`internal/example/skills_content_test.go`) with three distinct kinds of hardcoded name literals to update; one documentation section (`docs/architecture-specification.md` §38).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I/II (Semantic/Deterministic Separation, Deterministic Ops)**: N/A / PASS — this feature touches no deterministic operation or its logic at all; it is a content rename of prompt files and one test file's literals.
- **Principle III (Filesystem is Source of Truth)**: PASS — no persisted state is introduced; `kit.SkillsFS`'s directory listing *is* the source of truth for what Skills exist, unchanged in kind, just renamed.
- **Principle IV (Simplicity First / YAGNI)**: PASS — no new package, config, or command; confirmed by investigation that no name-transformation layer needs to be built, since none exists to begin with (directory name already *is* the command name everywhere).
- **Principle V (Test-First, NON-NEGOTIABLE)**: Applies in its content-conformance form (as with every prompt-content feature this session) — `internal/example/skills_content_test.go` already exists and already fails the moment directories are renamed without updating it (a real, natural test-first signal: renaming the directories first, before touching the test file, makes every `assertSkillConformant`/`fs.ReadFile` call in that suite fail immediately, which Phase 2's tasks will use as the starting "confirm it fails" checkpoint).
- **Principle VI (Clean Code, DRY)**: PASS — `canonicalSkillNames` already drives `TestSkillsContent_AllNineInstalled`'s real end-to-end install assertions; updating that one slice cascades correctly rather than requiring parallel edits to redundant lists.
- **Principle VII (Explicit Mutation Boundaries)**: N/A — no artifact-lifecycle mutation boundary is relevant to renaming the tool's own shipped prompt content.
- **Principle VIII (Safety by Construction)**: PASS — `git mv` preserves history; no destructive deletion is needed (a rename, not a delete-and-recreate).
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS — no change to any JSON/Finding contract shape; purely Skill-content and one test file.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/028-mister-prefixed-skill-names/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
kit/skills/
├── analyze/                 -> mister-analyze/                 (git mv, then edit cross-refs)
├── create-constitution/     -> mister-constitution/             (git mv, then edit cross-refs)
├── create-feature/          -> mister-features/                 (git mv, then edit cross-refs)
├── create-knowledge-base/   -> mister-knowledge-base/            (git mv, then edit cross-refs)
├── create-plan/             -> mister-plan/                     (git mv, then edit cross-refs)
├── create-program/          -> mister-program/                  (git mv, then edit cross-refs)
├── create-specs/            -> mister-specify/                  (git mv, then edit cross-refs)
├── create-tasks/            -> mister-tasks/                    (git mv, then edit cross-refs)
└── implement/                -> mister-implement/                (git mv, then edit cross-refs)

internal/example/
└── skills_content_test.go  # skillOperationsAllowlist map keys (9); canonicalSkillNames
                            # slice (9); 3 feature-specific tests' literal
                            # "<name>/SKILL.md" fs.ReadFile paths and their
                            # error-message name literals

docs/
└── architecture-specification.md  # §38 Canonical Skill Set code block (9 names)
```

**Structure Decision**: Single Go project, embedded-content model already in use. This is the largest *file-count* change this session (9 directories, each with internal cross-reference edits) but the smallest in terms of new logic — zero new Go code, one existing test file updated to match, one doc section updated. `kit/kit.go` itself needs no change (its `//go:embed skills` directive is directory-content-agnostic).

## Complexity Tracking

*No Constitution Check violations — table not needed.*
