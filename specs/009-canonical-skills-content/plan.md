# Implementation Plan: Canonical Skills Content

**Branch**: `009-canonical-skills-content` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-canonical-skills-content/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Author the nine canonical Skills `docs/architecture-specification.md`
§38 names (create-knowledge-base, create-constitution, create-program,
create-feature, create-specs, create-plan, create-tasks, implement,
analyze), each a `kit/skills/<name>/SKILL.md` conforming to §39's
required structure, §40's deterministic-operations transparency, and
§50's completion contract — replacing 008-cli-cobra's single placeholder
file. Grounding this against what 001-008 actually built surfaced one
real, necessary technical change: `internal/installer`'s `ListFS`/
`InstallFS` only ever listed a source directory's immediate children,
never descending into subdirectories, so a real `<skill-name>/SKILL.md`
layout would silently fail to install. This plan makes `ListFS`/
`InstallFS` recursive (backward-compatible, verified against every
existing flat fixture) as the Foundational enabler every Skill's
installation depends on, then each user story adds its own Skills'
content plus machine-checked structural conformance. No other package
changes; §41-49's two illustrative operations that don't exist yet
(`references`, Task ID allocation) are resolved using what already
exists rather than built new (spec.md Assumptions, research.md).

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged) for the one code change; Markdown + YAML frontmatter for the nine Skills themselves.
**Primary Dependencies**: Go standard library only (`io/fs`, `path/filepath`) for the installer change — no new external dependency. No new dependency for content authoring.
**Storage**: Filesystem. This feature's only Go-level write-path change is *what* gets discovered and installed (`internal/installer`'s recursive walk) — the atomic-write, no-silent-overwrite, and containment mechanisms themselves are unchanged (006-agent-adapter, unmodified). Content itself lives at `kit/skills/<name>/SKILL.md`, embedded via `kit.SkillsFS` (008-cli-cobra's shape, unchanged).
**Testing**: `go test` — a filesystem-integration regression suite (005/006/008's full existing suites, re-run unmodified, as the recursive-install change's primary proof) plus new tests: `internal/installer`'s own nested-resource behavior, and a structural-conformance suite parsing all nine installed `SKILL.md` files against data-model.md's contract table (heading completeness, frontmatter, operations allowlist, completion-contract concepts). Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged; `filepath.FromSlash` used explicitly where an `fs.FS`-style relative name becomes an OS path (research.md), so the recursive-install change is correct beyond Linux, not only coincidentally so.
**Project Type**: Single Go module. One extended package (`internal/installer`, additive behavior change only — no exported signature changes). One extended embed (`kit/kit.go`'s `SkillsFS` content). Nine new content directories under `kit/skills/`.
**Performance Goals**: Trivial — nine small Markdown files; no performance-sensitive path. The recursive walk itself is bounded by the embedded kit's own small size, not runtime project data.
**Constraints**: `ListFS`/`InstallFS`'s exported signatures must not change (every existing caller — `Install`, `List`, `claude.Install`, `bootstrap.Bootstrap` transitively — keeps compiling and passing unmodified); every Skill's "Deterministic Operations" section names only a real, already-registered `misterspec internal <op>` (FR-003, machine-checked); no new `internal/operations`/`internal/validation` function is added (spec.md Assumptions — this is content, not new deterministic behavior).
**Scale/Scope**: `internal/installer/{installer.go modified, installer_test.go extended}`; `kit/kit.go` doc-comment touch-up (shape unchanged); `kit/skills/README.md` removed; nine new `kit/skills/<name>/SKILL.md` files; one new structural-conformance test (in `internal/example` or a small dedicated package). No CLI change, no new adapter, no second coding agent.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass, realized for the first time on the content side.** Every prior feature built the deterministic half of this boundary; this feature authors the semantic half — Skills that *decide* (extract facts, choose boundaries, judge pass/fail) while their own "Deterministic Operations" sections delegate every mechanical step to what 001-008 already built (FR-008, FR-009). |
| II. Deterministic Ops as Only Mutation Primitive | **Pass, reinforced.** Every Skill's Procedure section instructs the agent to mutate the repository only via an existing `misterspec internal <op>` — never a hand-authored file write, hand-picked ID, or guessed path (FR-008). |
| III. Filesystem Is Single Source of Truth | **Pass (N/A).** This feature adds no new persisted project state — Skills are framework-shipped content, installed the same way kit templates already are. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** The two operations §41-49 illustrate but that don't exist (`references`, Task-ID allocation) are deliberately *not* built — existing capability (`inspect`'s `depends_on`/`supersedes`; direct Task authoring, matching 003's own precedent) is reused instead, exactly the discipline every prior feature has applied to its own scope boundary. |
| V. Test-First Discipline | **Gate carried into tasks.** The recursive-install change is test-first per usual; the structural-conformance suite is this feature's own equivalent test-first discipline applied to content rather than only code. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `ListFS`/`InstallFS` are extended in place, not duplicated into a Skill-specific installer (research.md) — the same "generalize, don't fork" discipline 006-agent-adapter's own `InstallFS`/`ListFS` generalization set. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced and now content-level too.** Every Skill's own "Allowed Reads"/"Allowed Creates"/"Allowed Modifications"/"Forbidden Mutations" sections state its boundary explicitly (FR-007) — the same discipline this project's Go code has always had, now expressed for the agent-facing layer directly. |
| VIII. Safety by Construction | **Pass, reinforced.** The recursive-install change reuses `WriteAtomicFile`'s and `RelativeWithinRoot`'s already-proven guarantees unmodified — no new write path, no new containment logic. |
| IX. Transparent, Machine-Readable Contracts | **Pass, extended to content.** §40's own rule — a Skill must explicitly state which deterministic operations it calls — is this principle applied to agent-facing documentation instead of only JSON; FR-003/SC-004 make it verifiable, not just readable. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/009-canonical-skills-content/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── skills.md          # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # unchanged — no new dependency
├── cmd/misterspec/                       # unchanged
├── kit/
│   ├── kit.go                            # doc-comment touch-up only; SkillsFS shape unchanged
│   ├── templates/                        # unchanged
│   └── skills/
│       ├── README.md                      # REMOVED (research.md)
│       ├── create-knowledge-base/
│       │   └── SKILL.md                    # NEW
│       ├── create-constitution/
│       │   └── SKILL.md                    # NEW
│       ├── create-program/
│       │   └── SKILL.md                    # NEW
│       ├── create-feature/
│       │   └── SKILL.md                    # NEW
│       ├── create-specs/
│       │   └── SKILL.md                    # NEW
│       ├── create-plan/
│       │   └── SKILL.md                    # NEW
│       ├── create-tasks/
│       │   └── SKILL.md                    # NEW
│       ├── implement/
│       │   └── SKILL.md                    # NEW
│       └── analyze/
│           └── SKILL.md                    # NEW
└── internal/
    ├── project/, artifacts/, ids/, testutil/                # unchanged
    ├── lock/, templates/, validation/, operations/           # unchanged
    ├── agents/, agents/claude/, agents/builtin/, bootstrap/   # unchanged
    ├── cli/, cli/internalcmd/                                  # unchanged
    ├── installer/
    │   ├── installer.go              # MODIFIED: ListFS/InstallFS recursive (research.md)
    │   ├── installer_test.go         # MODIFIED: + nested-resource test cases
    │   ├── filesystem.go, filesystem_test.go, containment_test.go, doc.go  # unchanged
    └── example/
        └── skills_content_quickstart_test.go   # NEW — structural-conformance + install end-to-end
```

**Structure Decision**: One package modified in place
(`internal/installer`, additive behavior only), nine new content
directories under `kit/skills/`, one removed placeholder file, one new
test file. No new package. This is the smallest-footprint Go change of
any feature since 001, deliberately — the feature's actual scope is
content, and research.md's recursive-install decision is the one
technical prerequisite that content needs to actually work.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
