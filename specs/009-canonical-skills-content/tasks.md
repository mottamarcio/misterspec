---

description: "Task list template for feature implementation"
---

# Tasks: Canonical Skills Content

**Input**: Design documents from `/specs/009-canonical-skills-content/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/skills.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE) — applied here to *content* conformance, not only Go behavior (research.md's "structural conformance is machine-checked" decision). 005-embedded-kit's, 006-agent-adapter's, and 008-cli-cobra's full test suites are explicit, named regression gates for the recursive-install change (Foundational), since none of those features' own packages are otherwise touched.

**Organization**: Tasks are grouped by user story, matching §52's pipeline order. **User Story 1 (Knowledge/Constitution) depends only on Foundational** (the recursive-install change every Skill needs). **User Story 2 (Program/Feature/Specs) depends only on Foundational** too — its three Skills' installability and structural conformance is independently provable without User Story 1's two Skills having been authored first, even though the *narrative* pipeline runs through them. **User Story 3 (Plan/Tasks/Implement/Analyze) likewise depends only on Foundational** for the same reason. All three stories share one reusable structural-conformance test helper, introduced in User Story 1 (first story to need it) and reused, not reimplemented, by User Story 2 and 3's own tests.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/installer/            # existing package — Foundational (extended)
kit/skills/                      # existing dir — US1, US2, US3 (content), Foundational (placeholder removal)
internal/example/                 # existing package — extended in US1, Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package, no new dependency, and no new directory skeleton beyond what already exists (`kit/skills/` was created by 008-cli-cobra). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Make `internal/installer`'s `ListFS`/`InstallFS` recursive — the one technical change every canonical Skill's installability depends on (research.md). No Skill content can be authored meaningfully until this is proven safe.

**⚠️ CRITICAL**: No user story's content is considered "installable" until this phase's regression gate (T003) passes.

- [X] T001 [P] Unit tests for `ListFS`/`InstallFS`'s recursive behavior in `internal/installer/installer_test.go`: a fixture `fstest.MapFS` with a nested file (e.g. `"skills/create-plan/SKILL.md"`) is discovered with `Resource.Name == "create-plan/SKILL.md"` (not skipped as a directory entry); `InstallFS` materializes it at `targetDir/create-plan/SKILL.md`, parent directory created automatically; every existing flat fixture (`fixtureSkillsFS()`, `kit.TemplatesFS`) produces an identical `Resource` set before and after (asserted directly, not just "still green").
- [X] T002 Implement the recursive walk (`fs.WalkDir`) in `ListFS`, `internal/installer/installer.go` — `Resource.Name` becomes the discovered file's path relative to `sourceDir`; `installOneFS`'s OS-path construction uses `filepath.FromSlash(r.Name)` (research.md). `ListFS`/`InstallFS`'s exported signatures unchanged. Depends on T001.
- [X] T003 Regression checkpoint: run `go test ./internal/installer/... ./internal/agents/... ./internal/agents/claude/... ./internal/bootstrap/... ./internal/cli/...` and confirm every existing test (005/006/008) passes unmodified — the explicit, named proof the recursive change is backward compatible (research.md). Depends on T002.

**Checkpoint**: Foundation ready — nested installation proven correct and non-regressive. User story content authoring can begin.

---

## Phase 3: User Story 1 - Start a New Project's Knowledge and Memory (Priority: P1) 🎯 MVP

**Goal**: `create-knowledge-base` and `create-constitution` exist as real, structurally conformant, installable Skills — the two the whole pipeline depends on.

**Independent Test**: Bootstrap a fresh project, place a sample document under its raw-sources directory, invoke `/create-knowledge-base` then `/create-constitution`, confirming each produces its documented artifact and completion summary — without any later-stage Skill installed.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T006–T007.

- [X] T004 [US1] Write the shared structural-conformance helper and its first two call sites in `internal/example/skills_content_test.go`: `assertSkillConformant(t, skillName)` — reads `kit/skills/<skillName>/SKILL.md` via `kit.SkillsFS`, asserts non-empty frontmatter `name`/`description`, all 29 of §39's H2 headings present in order, every `misterspec internal <word>` token in its "Deterministic Operations" section is one of the ten real commands *and* matches that skill's own row in data-model.md's per-Skill table, and its "Completion Contract" section names all five of §50's concepts. `TestSkillsContent_KnowledgeAndConstitution` calls it for `"create-knowledge-base"` and `"create-constitution"`.
- [X] T005 [US1] Confirm T004 fails — `kit/skills/create-knowledge-base/` and `kit/skills/create-constitution/` do not exist yet (only `kit/skills/README.md`, 008's placeholder).

### Implementation for User Story 1

- [X] T006 [US1] Author `kit/skills/create-knowledge-base/SKILL.md` — frontmatter (`name: create-knowledge-base`, `description`), all 29 §39 sections, Deterministic Operations per data-model.md's row (`internal inventory raw`, `internal fingerprint <source>`, `internal create knowledge --slug <slug>`, `internal inspect <knowledge-id>`, `internal validate`), Completion Contract per §50/§51's own example shape. Depends on T004, T005.
- [X] T007 [US1] Author `kit/skills/create-constitution/SKILL.md` — same structure; Deterministic Operations `internal inventory knowledge`, `internal resolve KNOW-###`, `internal validate` (§42). Depends on T004, T005.
- [X] T008 [US1] Remove `kit/skills/README.md` (008-cli-cobra's compile-only placeholder, no longer needed now that real content exists — research.md). Depends on T006, T007.
- [X] T009 [US1] Confirm `TestSkillsContent_KnowledgeAndConstitution` passes and `go build ./...` succeeds (confirms `kit.SkillsFS` still compiles with the placeholder removed). Depends on T008.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/example/... -run TestSkillsContent_KnowledgeAndConstitution` passes on its own, with zero dependency on User Story 2 or 3's Skills existing.

---

## Phase 4: User Story 2 - Decompose a Project into Programs, Features, and Specs (Priority: P2)

**Goal**: `create-program`, `create-feature`, and `create-specs` exist as real, structurally conformant, installable Skills.

**Independent Test**: Against a bootstrapped project, invoke `/create-program`, then `/create-feature` for the resulting Program, then `/create-specs` for the resulting Feature — independent of whether User Story 1's Skills were actually invoked first.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T012–T014.

- [X] T010 [US2] `TestSkillsContent_ProgramFeatureSpecs` in `internal/example/skills_content_test.go`, reusing T004's `assertSkillConformant` for `"create-program"`, `"create-feature"`, `"create-specs"`.
- [X] T011 [US2] Confirm T010 fails — none of the three directories exist yet.

### Implementation for User Story 2

- [X] T012 [US2] [P] Author `kit/skills/create-program/SKILL.md` — Deterministic Operations `internal status`, `internal create program`, `internal validate PRG-###` (§43). Depends on T010, T011.
- [X] T013 [US2] [P] Author `kit/skills/create-feature/SKILL.md` — Deterministic Operations `internal resolve PRG-###`, `internal children PRG-### --type feature`, `internal create feature --parent PRG-###`, `internal validate FEAT-###` (§44). Depends on T010, T011.
- [X] T014 [US2] [P] Author `kit/skills/create-specs/SKILL.md` — Deterministic Operations `internal resolve FEAT-###`, `internal children FEAT-### --type spec`, `internal create spec --parent FEAT-###`, `internal validate SPEC-###` (§45). Depends on T010, T011.
- [X] T015 [US2] Confirm `TestSkillsContent_ProgramFeatureSpecs` passes. Depends on T012, T013, T014.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/example/... -run TestSkillsContent_ProgramFeatureSpecs` passes on its own.

---

## Phase 5: User Story 3 - Plan, Break Down, Implement, and Verify a Spec (Priority: P3)

**Goal**: `create-plan`, `create-tasks`, `implement`, and `analyze` exist as real, structurally conformant, installable Skills.

**Independent Test**: Against an existing Spec (seeded directly, independent of whether User Story 1/2's Skills produced it), invoke `/create-plan`, then `/create-tasks`, then `/implement` for one task, then `/analyze` — including `/analyze` correctly reporting a gap for a deliberately incomplete implementation.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T019–T021.

- [X] T016 [US3] `TestSkillsContent_PlanTasksImplementAnalyze` in `internal/example/skills_content_test.go`, reusing T004's `assertSkillConformant` for `"create-plan"`, `"create-tasks"`, `"implement"`, `"analyze"`.
- [X] T017 [US3] Confirm T016 fails — none of the four directories exist yet.

### Implementation for User Story 3

- [X] T018 [US3] [P] Author `kit/skills/create-plan/SKILL.md` — Deterministic Operations `internal resolve SPEC-###`, `internal inspect SPEC-###` (its `depends_on`/`supersedes` fields stand in for §46's `references` — research.md's gap resolution), `internal create-artifact plan --for SPEC-###`, `internal validate SPEC-###`. Depends on T016, T017.
- [X] T019 [US3] [P] Author `kit/skills/create-tasks/SKILL.md` — Deterministic Operations `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal create-artifact tasks --for SPEC-###`, `internal validate SPEC-###`; Procedure instructs the agent to author `## TASK-NNN` headings directly (no allocator exists — 003-entity-creation's own precedent, research.md). Depends on T016, T017.
- [X] T020 [US3] [P] Author `kit/skills/implement/SKILL.md` — Deterministic Operations `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal validate SPEC-###` (§48, minus `internal references`). Depends on T016, T017.
- [X] T021 [US3] [P] Author `kit/skills/analyze/SKILL.md` — Deterministic Operations `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal create-artifact validation --for SPEC-###`, `internal validate SPEC-###`; Decision Rules/Recommended Next Step section encodes §53's branching rule (recommend the responsible artifact layer, not a fixed next command, on failure). Depends on T016, T017.
- [X] T022 [US3] Confirm `TestSkillsContent_PlanTasksImplementAnalyze` passes. Depends on T018, T019, T020, T021.

**Checkpoint**: All three user stories pass their own tests — all nine canonical Skills exist and are individually conformant.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Whole-feature consistency and validation across all nine Skills, no new capability.

- [X] T023 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 008-cli-cobra's packages included) and fix any findings.
- [X] T024 `TestSkillsContent_AllNineInstalled` in `internal/example/skills_content_test.go`: exactly the nine names §38 lists exist under `kit.SkillsFS` (no more, no fewer — catches a stray leftover file the same way T008 removing `README.md` was meant to); a real end-to-end install via `claude.New().Install(ctx, ...)` into a fresh temp directory confirms all nine land at `.claude/skills/<name>/SKILL.md`, byte-identical to `kit.SkillsFS`'s own content (quickstart.md's validation strategy). Depends on T009, T015, T022.
- [X] T025 Reconcile `specs/009-canonical-skills-content/contracts/skills.md` against the actual implementation; document any drift (same discipline as every prior feature's final reconciliation task).
- [X] T026 Full regression run: `go test ./...` across the entire module (001 through 009) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/installer/... -race` clean (mutating recursive installs, mirroring 003's/007's/008's `-race` discipline).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS all user stories — no Skill's installability can be proven until the recursive-install change is verified non-regressive (T003).
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on Foundational only — not on User Story 1's Skills existing.
- **User Story 3 (Phase 5)**: Depends on Foundational only — not on User Story 1 or 2's Skills existing.
- **Polish (Phase 6)**: Depends on all three user stories being complete (T024 in particular needs all nine to exist for its "exactly nine" assertion).

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Can start after Foundational, in parallel with User Story 1 and User Story 3 — each story's Skills are structurally and installably independent of the others existing, even though the *narrative* pipeline (§52) runs through them in order.
- **User Story 3 (P3)**: Can start after Foundational, in parallel with User Story 1 and User Story 2, for the same reason.

### Within Each User Story

- Tests (the story's own `TestSkillsContent_*` function, reusing T004's shared helper) written and failing before any of that story's `SKILL.md` files are authored (Constitution Principle V).
- Within a story, its own Skills are mutually independent files — `[P]` in User Story 2 and 3's implementation tasks.

### Parallel Opportunities

- Once Foundational is done: **User Story 1, User Story 2, and User Story 3 can all proceed fully in parallel** by different contributors — a three-way independence beyond 003's/006's own two-way parallel-story precedent, since every Skill's own structural conformance and installability is self-contained.
- Within User Story 2: T012, T013, T014 (three Skills, three files) in parallel.
- Within User Story 3: T018, T019, T020, T021 (four Skills, four files) in parallel.
- Within Polish: T023 alone; T024-T026 are sequential (each depends on the full Skill set existing).

---

## Parallel Example: All three user stories together

```bash
# Once Foundational (T003) is done, these three stories need no coordination:
Task: "Author create-knowledge-base/SKILL.md and create-constitution/SKILL.md (User Story 1)"
Task: "Author create-program/SKILL.md, create-feature/SKILL.md, create-specs/SKILL.md (User Story 2)"
Task: "Author create-plan/SKILL.md, create-tasks/SKILL.md, implement/SKILL.md, analyze/SKILL.md (User Story 3)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (recursive install proven safe).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/example/... -run TestSkillsContent_KnowledgeAndConstitution` green, independently.
5. This alone already lets a coding agent turn raw documents into structured Knowledge and a Constitution — the pipeline's own starting point (§52), usable even before any later-stage Skill exists.

### Incremental Delivery

1. Setup (nothing) + Foundational (recursive install).
2. Add US1 → validate independently → Knowledge/Constitution usable (MVP).
3. Add US2 (independent of US1, can be parallel) → validate independently → Program/Feature/Spec decomposition usable.
4. Add US3 (independent of US1/US2, can be parallel) → validate independently → Plan/Tasks/Implement/Analyze usable.
5. Polish (Phase 6), including the all-nine comprehensive check and the full-module `-race` regression run (T026).

### Team Strategy

Once Foundational is done, three developers can take User Story 1, 2,
and 3 fully independently and in parallel — the widest story-level
parallelism this project has had since 001-core-foundation, because
authoring one Skill's content has no code-level dependency on any
other Skill existing.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — applied to content structure, not only Go behavior; write the assertion, watch it fail because the file doesn't exist, then author the file.
- No Skill's "Deterministic Operations" section may name an operation the CLI doesn't actually have (FR-003) — `internal references` and Task-ID allocation are deliberately absent from every Skill (research.md).
- The recursive-install change (Foundational) must not alter any existing test's outcome — T003 is a named, explicit regression gate, not an implicit assumption.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
