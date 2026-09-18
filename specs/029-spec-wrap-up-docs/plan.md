# Implementation Plan: `/mister-wrap-up` — Spec Documentation for Future Official Docs

**Branch**: `029-spec-wrap-up-docs` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/029-spec-wrap-up-docs/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add a 10th canonical Skill, `mister-wrap-up`, that synthesizes a Spec's own artifacts, its commit history, the Knowledge base, the Constitution, and related Learnings into one Markdown document under `cortex/SPEC-###-<slug>.md` — raw material for a future official docs site or GitBook, not a canonical validated project artifact. The one piece of real Go logic this feature needs — "every commit on the current branch from the one that first added this Spec's `plan.md`, through now" — is a mechanical, deterministic computation (Constitution Principle I/II), so it becomes a new, small `internal/vcs` function exposed via a new `internal commits-since-file` CLI command, following the exact precedent `022-feature-branch-automation` and `027-constitution-frontmatter-task-deps` already established for Git-related mechanics: no commit-message convention required, and no raw `git log` parsing left to the agent. Everything else — reading the Spec's artifacts, Knowledge, Constitution, Learnings, and composing synthesized prose — is the Skill's own judgment, same as every other canonical Skill.

## Technical Context

**Language/Version**: Go 1.23.4 (repo-wide). The new deterministic operation is Go; `mister-wrap-up` itself is Markdown + YAML frontmatter prompt content embedded via `go:embed`, identical in kind to the other 9 canonical Skills.
**Primary Dependencies**: `internal/vcs` (extended with a new function computing a file's "added" commit and the commit range since), `internal/cli/internalcmd` (new CLI command exposing it under the existing hidden `misterspec internal …` namespace), `kit.SkillsFS` (embedded Skill content, now 10 directories), `internal/example/skills_content_test.go` (extended for the 10th Skill).
**Storage**: N/A — reads Git history and existing artifacts (`ai/programs/…`); writes one Markdown file per Spec under `cortex/` at the project root, a new output location parallel to (not inside) `ai/`. No database, no new persisted state beyond that one file, which is itself fully regenerated on each run (Constitution Principle III — nothing is amended or accumulated).
**Testing**: `go test ./internal/vcs/...` (new commit-range function: file never committed, file added at repo root commit with no parent, ordinary case with commits before and after, not a git repo at all), `go test ./internal/cli/internalcmd/...` (new CLI command's JSON contract), `go test ./internal/example/...` (`skills_content_test.go` extended: `mister-wrap-up` added to `skillOperationsAllowlist` and `canonicalSkillNames`; the existing `TestSkillsContent_AllNineInstalled` — whose own name and doc comment say "nine" — renamed to reflect 10 canonical Skills, since this feature makes that count stale otherwise).
**Target Platform**: Same as always — the hidden `misterspec internal …` command surface plus the agent Skill prompt surface.
**Project Type**: Single Go project with an embedded content kit — unchanged structurally; `cortex/` is a new top-level directory this feature's Skill creates in a *consumer* project (not in this repository's own layout, except as the target of a manual quickstart validation).
**Performance Goals**: N/A — one `git log` invocation per run, same order of cost as `022`'s existing branch-lookup calls.
**Constraints**: MUST NOT require any commit-message convention (spec.md FR-005) — commit range is derived solely from `plan.md`'s own Git history. MUST NOT fabricate content for a source with nothing relevant (FR-010) — an empty Knowledge/Constitution/Learnings contribution is stated as "none found," never invented. MUST degrade gracefully (not error) when the project isn't a Git repository or `plan.md`'s history can't be determined (FR-007) — the document still generates from every other source, with a clear note that commit history is missing. MUST regenerate the document wholesale on every run (FR-009) — unlike every other canonical Skill's amend-in-place discipline (`create-constitution`, `create-tasks`, etc.), `mister-wrap-up`'s own output is explicitly non-canonical downstream content, so overwriting it in full each time is correct, not a violation of Principle VII, as long as nothing *other* than its own `cortex/SPEC-###-*.md` file is ever written to. `cortex/` MUST NOT require a new project configuration field — it is a fixed convention, per spec.md's own Assumptions, with no stated need to vary between projects (Constitution Principle IV).
**Scale/Scope**: One new `internal/vcs` function + one new CLI command (small, mirrors `022`/`027`'s own precedent for Git-mechanics); one new canonical Skill (`kit/skills/mister-wrap-up/SKILL.md`, the 10th); `internal/example/skills_content_test.go` extended (new allowlist entry, new `canonicalSkillNames` entry, one existing test renamed); `docs/architecture-specification.md` §38 updated to list all 10 names and note `mister-wrap-up`'s own optional, downstream-documentation role distinct from the core MVP pipeline.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I/II (Semantic/Deterministic Separation, Deterministic Ops as sole mutation primitive)**: PASS. "Which commits touch this Spec" is computed mechanically from `plan.md`'s own Git history via the new deterministic operation — never left to the agent to run raw `git log` or infer from commit message conventions. Judging which Knowledge/Constitution/Learning content is *relevant* to the Spec remains the agent's own semantic call (Learnings in particular have no structured link to a Spec at all — confirmed via `docs/architecture-specification.md` §31, their schema carries no `for`/`parent` field), correctly left to the Skill, not forced into a deterministic query that doesn't exist.
- **Principle III (Filesystem is Source of Truth)**: PASS. The commit range is derived live from Git on each run; the wrap-up document itself is fully regenerated each time, not amended — no new persisted "last run" state anywhere.
- **Principle IV (Simplicity First / YAGNI)**: PASS. One small, narrowly-scoped new CLI command, justified by the same demonstrated need `022`/`027` already established for Git-mechanics; no new config field for `cortex/`'s location, since nothing in the request states it needs to vary between projects.
- **Principle V (Test-First, NON-NEGOTIABLE)**: Applies fully to the new `internal/vcs` function and CLI command (real deterministic Go logic, real unit tests written first). Applies in its content-conformance form to the new Skill file itself, consistent with every other Skill this session.
- **Principle VI (Clean Code, DRY)**: PASS. The new commit-range function reuses `internal/vcs`'s existing `IsRepo`/git-invocation conventions rather than a parallel implementation; filename slugification reuses (generalizes, not duplicates) the existing branch-slug logic already in `internal/vcs`.
- **Principle VII (Explicit Mutation Boundaries)**: PASS, with an explicit, documented exception: `mister-wrap-up` is the first canonical Skill whose own output is regenerated wholesale rather than amended — because its output (`cortex/*.md`) is downstream documentation material, not a canonical validated artifact like every other Skill's own output. It remains strictly read-only against every other artifact (Spec, Plan, Tasks, Validation, Knowledge, Constitution, Learnings) — it may create/overwrite only its own one file per Spec.
- **Principle VIII (Safety by Construction)**: PASS. Writes are scoped to exactly `cortex/SPEC-###-<slug>.md` for the named Spec — never another Spec's own file, never anything outside `cortex/`.
- **Principle IX (Transparent, Machine-Readable Contracts)**: PASS. The new CLI command returns structured JSON (commit list, or a clear "not available" shape) consistent with every other `internal` command's own contract; the Skill itself still carries the full, machine-checked Completion Contract every canonical Skill requires.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/029-spec-wrap-up-docs/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/vcs/
└── vcs.go                  # + CommitsSinceFileAdded(root, path string) ([]Commit,
                            # bool, error) — bool reports whether history could be
                            # determined at all (false when not a repo, or the file
                            # was never committed); generalizes the existing
                            # slug-derivation helper for reuse by filenames too

internal/cli/internalcmd/
└── commits_since_file.go     # New: `misterspec internal commits-since-file
                            # --path <path>` — thin Cobra wrapper returning
                            # structured JSON, mirroring 022's/027's own existing
                            # command shapes

kit/skills/
└── mister-wrap-up/SKILL.md  # New, 10th canonical Skill: full §39-required section
                            # set, Deterministic Operations section naming the new
                            # `internal commits-since-file` call plus existing
                            # `internal resolve`/`internal inspect`/`internal
                            # inventory`, Forbidden Mutations explicitly scoping
                            # writes to its own cortex/*.md file only

internal/example/
└── skills_content_test.go  # + "mister-wrap-up" in skillOperationsAllowlist and
                            # canonicalSkillNames; TestSkillsContent_AllNineInstalled
                            # renamed (its own name/doc comment say "nine" — no
                            # longer accurate with a 10th canonical Skill)

docs/
└── architecture-specification.md  # §38 Canonical Skill Set: add mister-wrap-up,
                                   # note its optional/downstream role distinct
                                   # from the required MVP pipeline
```

**Structure Decision**: Single Go project, embedded-content model already in use. Mirrors `022`/`027`'s own established pattern exactly: one small, additive `internal/vcs` function + CLI command for the mechanical Git piece, one new canonical Skill for everything requiring judgment. No new package, no new top-level directory in this repository itself (`cortex/` is created by the Skill in a *consumer* project, not here).

## Complexity Tracking

*No Constitution Check violations — table not needed.*
