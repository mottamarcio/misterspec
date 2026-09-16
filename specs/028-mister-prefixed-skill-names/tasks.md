---

description: "Task list for feature implementation"
---

# Tasks: `mister-`-Prefixed Skill Names to Avoid Slash-Command Collisions

**Input**: Design documents from `/specs/028-mister-prefixed-skill-names/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/skill-names.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First, NON-NEGOTIABLE), a new permanent regression test (`TestSkillsContent_NoStaleSkillNameReferences`) is written in Foundational — before the renames happen, so it starts by failing correctly (the 9 old names are still the current, legitimate names at that point) — and is confirmed passing only once both User Stories are complete. Existing structural-conformance tests (`skillOperationsAllowlist`-driven, `canonicalSkillNames`-driven) already exist and serve as User Story 1's own test-first signal: they start failing the moment directories are renamed (Foundational) and are fixed to pass by User Story 1's own tasks.

**Organization**: Tasks are grouped by user story. User Story 1 (each Skill's own identity: frontmatter `name:`, `## Invocation`) and User Story 2 (every Skill's cross-references to the other 8) both touch all 9 `SKILL.md` files, but as two distinct passes — User Story 2's edit to a given file always follows that same file's own User Story 1 edit (same-file dependency), while different Skills' files remain fully parallel across both stories.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single Go project with an embedded content kit. All Skill renames and content edits live under `kit/skills/`; the one Go test file touched is `internal/example/skills_content_test.go`; one documentation file (`docs/architecture-specification.md`) is touched once.

---

## Phase 1: Setup

**Purpose**: Establish an accurate baseline before any rename or edit.

- [X] T001 [P] Run a full-repo grep across `kit/skills/` for each of the 9 old bare command-name strings (`/analyze`, `/create-constitution`, `/create-feature`, `/create-knowledge-base`, `/create-plan`, `/create-program`, `/create-specs`, `/create-tasks`, `/implement`) and record the per-file occurrence count as a baseline to compare against after the rename (research.md's own ~72-occurrence figure, confirmed fresh).
- [X] T002 [P] Read `internal/example/skills_content_test.go` in full, noting the exact current line ranges of `skillOperationsAllowlist`'s 9 map keys, the three `assertSkillConformant(t, "...")` call groups, `canonicalSkillNames`'s 9 elements, and the three feature-specific tests' literal `fs.ReadFile(kit.SkillsFS, "<name>/SKILL.md")` calls.
- [X] T003 [P] Read `docs/architecture-specification.md` §38 (Canonical Skill Set) in full, noting its exact code-block content and confirming the separate "Canonical source: `.misterspec/skills/`" line beneath it (pre-existing, unrelated drift — out of scope, must not be touched by this feature).

**Checkpoint**: Baseline established; no files changed yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Write the whole-feature regression test first (per Constitution Principle V), then perform the 9 directory renames both User Stories build on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Add a new test `TestSkillsContent_NoStaleSkillNameReferences` in `internal/example/skills_content_test.go` that walks every file under `kit.SkillsFS` and fails if any of the 9 old bare command-name strings (as whole-word slash-command tokens, e.g. `` `/implement` ``, `` `/create-tasks` ``) appears anywhere in `kit/skills/` content. Run `go test ./internal/example/... -run TestSkillsContent_NoStaleSkillNameReferences` and confirm it **fails** — at this point the 9 old names are still the current, legitimate Skill names, so their presence is expected and correct; this failure is the test-first baseline the rest of this feature drives to green.
- [X] T005 [P] `git mv kit/skills/analyze kit/skills/mister-analyze`
- [X] T006 [P] `git mv kit/skills/create-constitution kit/skills/mister-constitution`
- [X] T007 [P] `git mv kit/skills/create-feature kit/skills/mister-features`
- [X] T008 [P] `git mv kit/skills/create-knowledge-base kit/skills/mister-knowledge-base`
- [X] T009 [P] `git mv kit/skills/create-plan kit/skills/mister-plan`
- [X] T010 [P] `git mv kit/skills/create-program kit/skills/mister-program`
- [X] T011 [P] `git mv kit/skills/create-specs kit/skills/mister-specify`
- [X] T012 [P] `git mv kit/skills/create-tasks kit/skills/mister-tasks`
- [X] T013 [P] `git mv kit/skills/implement kit/skills/mister-implement`

**Checkpoint**: All 9 directories exist only under their new names. `go test ./internal/example/...` now fails broadly (every test referencing an old path or old map key) — expected; User Stories 1 and 2 drive this back to green.

---

## Phase 3: User Story 1 - A misterspec user's slash commands no longer collide with another framework's (Priority: P1) 🎯 MVP

**Goal**: Each renamed Skill's own frontmatter `name:` and `## Invocation` reflect its new name; the structural test lists and the three literal-path tests reference the new names; the canonical documentation listing is updated.

**Independent Test**: Follow `quickstart.md` steps 1–2 — run `misterspec init` and confirm the 9 installed commands are exactly the new `mister-`-prefixed set, with none of the old names present.

### Implementation for User Story 1

- [X] T014 [P] [US1] In `kit/skills/mister-analyze/SKILL.md`, update the frontmatter `name:` field to `mister-analyze` and the `## Invocation` section's own slash command to `` `/mister-analyze` ``.
- [X] T015 [P] [US1] In `kit/skills/mister-constitution/SKILL.md`, update `name:` to `mister-constitution` and `## Invocation` to `` `/mister-constitution` ``.
- [X] T016 [P] [US1] In `kit/skills/mister-features/SKILL.md`, update `name:` to `mister-features` and `## Invocation` to `` `/mister-features` ``.
- [X] T017 [P] [US1] In `kit/skills/mister-knowledge-base/SKILL.md`, update `name:` to `mister-knowledge-base` and `## Invocation` to `` `/mister-knowledge-base` ``.
- [X] T018 [P] [US1] In `kit/skills/mister-plan/SKILL.md`, update `name:` to `mister-plan` and `## Invocation` to `` `/mister-plan` ``.
- [X] T019 [P] [US1] In `kit/skills/mister-program/SKILL.md`, update `name:` to `mister-program` and `## Invocation` to `` `/mister-program` ``.
- [X] T020 [P] [US1] In `kit/skills/mister-specify/SKILL.md`, update `name:` to `mister-specify` and `## Invocation` to `` `/mister-specify` `` (note: this Skill's own old Invocation was `` `/create-specs` `` — the new name follows the confirmed `mister-specify`, not `mister-specs`, exception).
- [X] T021 [P] [US1] In `kit/skills/mister-tasks/SKILL.md`, update `name:` to `mister-tasks` and `## Invocation` to `` `/mister-tasks` ``.
- [X] T022 [P] [US1] In `kit/skills/mister-implement/SKILL.md`, update `name:` to `mister-implement` and `## Invocation` to `` `/mister-implement` `` (both of this Skill's own invocation forms — `SPEC-###` alone and `SPEC-### TASK-NNN`, per `specs/026-implement-single-task`).
- [X] T023 [US1] In `internal/example/skills_content_test.go`, update `skillOperationsAllowlist`'s 9 map keys and `canonicalSkillNames`'s 9 elements to the new names, per `data-model.md`'s mapping table.
- [X] T024 [US1] In `internal/example/skills_content_test.go`, update the three feature-specific tests' literal paths and matching error-message string literals: `TestSkillsContent_ImplementDualInvocation` (`"implement/SKILL.md"` → `"mister-implement/SKILL.md"`), `TestSkillsContent_ConstitutionFrontmatterDocumented` (`"create-constitution/SKILL.md"` → `"mister-constitution/SKILL.md"`), `TestSkillsContent_CreateTasksReportsDependenciesAndParallelism` (`"create-tasks/SKILL.md"` → `"mister-tasks/SKILL.md"`). Run `go test ./internal/example/...` and confirm every `TestSkillsContent_*` test **except** `TestSkillsContent_NoStaleSkillNameReferences` now passes.
- [X] T025 [US1] In `docs/architecture-specification.md` §38, update the Canonical Skill Set code block to the 9 new names, leaving the separate "Canonical source: `.misterspec/skills/`" line immediately below it untouched (pre-existing, unrelated drift — out of scope).

**Checkpoint**: A fresh `misterspec init` now installs the 9 new names with correct self-identity; every existing structural-conformance test passes. Only cross-reference content (User Story 2) and the new whole-feature regression test remain.

---

## Phase 4: User Story 2 - Renamed Skills still correctly guide the user to each other (Priority: P1)

**Goal**: Every mention any of the 9 Skills makes of any of the other 8 — next-step guidance, stated dependencies, failure-condition recommendations, ordinary prose — uses that other Skill's new name.

**Independent Test**: Follow `quickstart.md` steps 3–4 — confirm `mister-tasks/SKILL.md`'s own next-step guidance names `/mister-implement`, and confirm a Skill known to mention another in ordinary prose (not just its own next-step block) is also updated; confirm `TestSkillsContent_NoStaleSkillNameReferences` passes.

### Implementation for User Story 2

- [X] T026 [P] [US2] In `kit/skills/mister-analyze/SKILL.md` (depends on T014 completing first — same file), find and update every cross-reference to another Skill (in `## Related Skills`, `## Recommended Next Step`, `## Failure Conditions`, and any other prose mention) to that Skill's new name.
- [X] T027 [P] [US2] Same as T026, for `kit/skills/mister-constitution/SKILL.md` (depends on T015).
- [X] T028 [P] [US2] Same as T026, for `kit/skills/mister-features/SKILL.md` (depends on T016).
- [X] T029 [P] [US2] Same as T026, for `kit/skills/mister-knowledge-base/SKILL.md` (depends on T017).
- [X] T030 [P] [US2] Same as T026, for `kit/skills/mister-plan/SKILL.md` (depends on T018) — known from the pre-plan investigation to mention `/implement` in ordinary prose, not only its own next-step block; confirm that mention is caught too.
- [X] T031 [P] [US2] Same as T026, for `kit/skills/mister-program/SKILL.md` (depends on T019).
- [X] T032 [P] [US2] Same as T026, for `kit/skills/mister-specify/SKILL.md` (depends on T020).
- [X] T033 [P] [US2] Same as T026, for `kit/skills/mister-tasks/SKILL.md` (depends on T021) — confirm its own `## Recommended Next Step` names `/mister-implement`, not `/implement` (spec.md Acceptance Scenario 1).
- [X] T034 [P] [US2] Same as T026, for `kit/skills/mister-implement/SKILL.md` (depends on T022) — confirm both its named next steps (naming another `/mister-implement TASK-NNN` invocation, and `/mister-analyze` once all Tasks are done, per `specs/026-implement-single-task`) are correct.
- [X] T035 [US2] Run `go test ./internal/example/... -run TestSkillsContent_NoStaleSkillNameReferences` and confirm it now **passes**.

**Checkpoint**: Both User Stories complete — every Skill's own identity and every cross-reference use the new names; the whole-feature regression test is green.

---

## Phase 5: User Story 3 - Old command names are fully retired, not kept as a fallback (Priority: P2)

**Goal**: Confirm no alias, redirect, or dual-registration exists for any old name — the rename is clean, matching spec.md FR-004.

**Independent Test**: Attempt to invoke an old command name after the rename and confirm it is not recognized.

### Validation for User Story 3

- [X] T036 [US3] Run `quickstart.md`'s own full-repo search command (`grep -rn` for all 9 old bare command-name strings scoped to `kit/skills/`) and confirm **zero matches** — this is the explicit, human-readable re-confirmation of what T035's automated test already guarantees.
- [X] T037 [US3] Build the `misterspec` binary from this branch and run `misterspec init --dir <scratch> --agent claude-code` against a scratch project; run `ls <scratch>/.claude/skills/` and confirm exactly the 9 new names are present with none of the 9 old names — live confirmation that no alias/dual-registration was accidentally introduced anywhere in the install path.

**Checkpoint**: All three user stories complete and independently verified.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Confirm the whole feature regresses nothing.

- [X] T038 [P] Run `go build ./...` and `go test ./...` in full and confirm no regressions anywhere in the repository, not only in the packages this feature touched. NOTE: this surfaced a real gap the original task list missed — `internal/example/multi_agent_skill_integration_quickstart_test.go` (from `specs/018-multi-agent-skill-integration`, an active Go test, not a historical document) also hardcoded `"implement"`/`"create-plan"`/`"create-tasks"`/`"analyze"` as map keys read via `kit.SkillsFS`. Fixed to the new names (FR-006 covers this: "every automated check that verifies a Skill's own content against its name" — this file qualifies even though it wasn't in the original per-file task list). Confirmed unrelated: `internal/installer/installer_test.go`'s own `create-plan`/`create-tasks` strings are synthetic fixture names in `fixtureNestedSkillsFS()`, testing generic nested-install mechanics, never touching `kit.SkillsFS` or the real 9 Skills — correctly left untouched. Full suite green after the fix.
- [X] T039 Execute `quickstart.md`'s remaining steps end-to-end (or, if a step needs a live multi-turn agent invocation unavailable in this session, a structural trace noting explicitly which form of validation was performed) and confirm every acceptance scenario in `spec.md` (User Stories 1–3) passes as described. NOTE: steps 1–2 (US1) and step 5 (US3) validated live in T037 (real `misterspec init` against a scratch project, `.claude/skills/` listing showed exactly the 9 new names). Steps 3–4 (US2 cross-reference wording) validated live via direct grep/read in T030/T033/T034's own spot-checks, plus the automated `TestSkillsContent_NoStaleSkillNameReferences` regression test (T035) passing as the mechanical proof no old name survives anywhere in `kit/skills/`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — all three tasks are `[P]` (read-only, different targets).
- **Foundational (Phase 2)**: Depends on Setup. T004 (the new regression test) should be written and confirmed failing *before* T005–T013 (the renames) run, so its failure is attributable to "old names still legitimately exist," the correct test-first baseline. T005–T013 are mutually `[P]` (9 independent directories) but all must complete before any User Story task starts — nothing in `kit/skills/` can be edited at its new path until the `git mv` has happened.
- **User Story 1 (Phase 3)**: Depends on Foundational. Its own 9 per-Skill tasks (T014–T022) are mutually `[P]` (different files); T023–T025 (test file structural lists, docs) have no file dependency on T014–T022 and could run in parallel with them, but are listed after for narrative clarity.
- **User Story 2 (Phase 4)**: Depends on Foundational directly, and on User Story 1's *own same-file task* specifically (T026 depends on T014, T027 on T015, etc. — editing a file's cross-references after its own identity edit, not before, keeps each file's own diff coherent) — but is otherwise mutually `[P]` across the 9 different Skills' files, and has no dependency on any *other* Skill's User Story 1 task (T026 does not wait for T015–T022, only T014).
- **User Story 3 (Phase 5)**: Depends on both User Stories 1 and 2 being complete — validating "nothing old remains reachable" is only meaningful once the rename and cross-reference fix are both done.
- **Polish (Phase 6)**: Depends on all three user stories.

### Parallel Opportunities

- **T001–T003** (Setup): fully parallel.
- **T005–T013** (Foundational renames): fully parallel — 9 independent directories.
- **T014–T022** (User Story 1, per-Skill identity): fully parallel — 9 different files.
- **T026–T034** (User Story 2, per-Skill cross-references): fully parallel *across Skills*, each individually gated only by its own same-file User Story 1 predecessor (T026 waits only for T014, not for T015–T022). A team of up to 9 people could each own one Skill's file end-to-end (its own T01x identity task, immediately followed by its own T02x cross-reference task) with zero coordination needed between different people's files, converging only at T035's shared regression-test check.
- **T038** (Polish, full regression): no file dependency on any single prior task, ordered last because it needs the whole feature present to be meaningful.

---

## Parallel Example: One Person Per Skill

```bash
# Each of 9 people owns one Skill's file end-to-end, after Foundational completes:
Person 1: T014 (mister-analyze identity) -> T026 (mister-analyze cross-refs)
Person 2: T015 (mister-constitution identity) -> T027 (mister-constitution cross-refs)
Person 3: T016 (mister-features identity) -> T028 (mister-features cross-refs)
# ...and so on through Person 9 (mister-implement).
# No coordination needed between people — different files throughout.
# Converge at T035 (shared regression test) once all 9 pairs are done.
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (the 9 renames + the new regression test, written failing).
3. Complete Phase 3: User Story 1 (each Skill's own identity + the structural test lists + docs).
4. **STOP and VALIDATE**: `misterspec init` now installs the correct 9 new names — the core of the reported collision problem is already resolved, even before cross-references are fixed.

### Incremental Delivery

1. Setup + Foundational → 9 directories renamed, regression test written and (correctly) failing.
2. Add User Story 1 → validate independently (`quickstart.md` steps 1–2) → new names install correctly.
3. Add User Story 2 → validate independently (`quickstart.md` steps 3–4, T035's test passing) → no user is ever stranded with a reference to a retired command.
4. Add User Story 3 → validate independently (T036–T037) → confirm no accidental alias exists.
5. Polish → full-repo regression.

### Parallel Team Strategy

The most parallelizable spec this session at the *individual task* level (9-way, not just 2-way): once Foundational is done, each of the 9 Skills' own identity-then-cross-reference pair (e.g. T014→T026) can be owned by a different person with zero file overlap with any other pair, converging only at the shared T035 regression-test checkpoint and the final T038 full-repo build/test.

---

## Notes

- [P] tasks = different files, no dependencies (except where a same-file same-Skill dependency is called out explicitly, e.g. T026 on T014).
- [Story] label maps task to specific user story for traceability.
- T004 must fail (for the expected reason — old names still current) before Foundational's renames run; T035 must pass only after both User Stories are complete.
- Commit after each phase (Foundational, User Story 1, User Story 2, User Story 3) rather than after every single task, so each phase's diff stays reviewable as one coherent unit — especially valuable here given the sheer file count.
- Avoid: touching `.claude/skills/speckit-*` (unrelated lineage), any already-completed historical Spec under `specs/` that mentions an old name, or the separate "Canonical source: `.misterspec/skills/`" drift in §38 — all explicitly out of scope per `research.md`'s own decisions.
