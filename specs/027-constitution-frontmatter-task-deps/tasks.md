---

description: "Task list for feature implementation"
---

# Tasks: Constitution Frontmatter Guarantee, and Task Dependency/Parallelism Reporting

**Input**: Design documents from `/specs/027-constitution-frontmatter-task-deps/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/validate-constitution.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First, NON-NEGOTIABLE), User Story 1 has real Go tests written first (new deterministic logic: a `SchemaVersion` field and a `checkConstitution` check). User Story 2 is prompt-content only (no new Go logic — plan.md's own Constraints are explicit that no new deterministic operation is introduced), so its test-first discipline takes the same form used for prompt-content changes elsewhere this session: a machine-checked content assertion (`skills_content_test.go`) written first and confirmed failing, then the `SKILL.md` edit that makes it pass.

**Organization**: Tasks are grouped by user story. Unlike prior specs this session, **User Story 1 and User Story 2 touch entirely different Go packages and different `SKILL.md` files** — they are genuinely independent and can be implemented in parallel by two different people, with the sole shared touch point being `internal/example/skills_content_test.go` (each adds one distinct new assertion to it).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Single Go project with an embedded content kit. US1's Go changes live under `internal/artifacts/` and `internal/validation/`; both stories' Skill-content changes live under `kit/skills/`; both stories' test additions live in `internal/example/skills_content_test.go`.

---

## Phase 1: Setup

**Purpose**: Establish an accurate baseline of the files each story will touch before editing them.

- [X] T001 [P] Read `internal/artifacts/parser.go` and `internal/artifacts/metadata.go` in full to confirm the current `frontmatterYAML`/`Metadata` shape (baseline for adding `SchemaVersion`, User Story 1).
- [X] T002 [P] Read `internal/validation/validator.go` and `internal/validation/findings.go` in full to confirm `ValidateProject`'s current aggregation shape and the existing `CodeFrontmatterMalformed`/`CodeRequiredFieldMissing` codes (baseline for `checkConstitution`, User Story 1).
- [X] T003 [P] Read `kit/skills/create-tasks/SKILL.md`'s current `Outputs` and `Completion Contract` sections in full to confirm exactly what dependency/parallel data is already recorded per Task (baseline for User Story 2).

**Checkpoint**: Both stories' target files are understood; no blocking prerequisite exists between the two stories themselves — both can now proceed in parallel.

---

## Phase 2: User Story 1 - The Constitution always has its required frontmatter (Priority: P1) 🎯 MVP

**Goal**: `internal validate` deterministically catches a missing/malformed Constitution frontmatter (`type`/`schema_version`, `docs/architecture-specification.md` §24), and `create-constitution/SKILL.md`'s own instructions explicitly require and preserve it — regardless of which model is running the Skill.

**Independent Test**: Follow `quickstart.md` steps 1–4 — create a Constitution, confirm correct frontmatter is written; break it manually; confirm `internal validate` now reports a Finding; re-run `/create-constitution` and confirm it repairs the frontmatter.

### Tests for User Story 1 (write first, confirm failing)

- [X] T004 [US1] In `internal/artifacts/parser_test.go`, add a test asserting `ParseMetadata` populates a new `SchemaVersion` field from a `schema_version:` frontmatter line (e.g. on a `type: constitution` fixture). Run `go test ./internal/artifacts/...` and confirm it **fails** (field doesn't exist yet).
- [X] T005 [US1] In `internal/validation/validator_test.go`, add tests for a `checkConstitution` check covering all four cases from `data-model.md`'s Constitution Finding table: (a) `ai/memory/constitution.md` absent → no Finding; (b) file present with unparseable/missing frontmatter block → one Finding with `Code: CodeFrontmatterMalformed`, `Path: "ai/memory/constitution.md"`; (c) file present, frontmatter parses, `schema_version` absent → one Finding with `Code: CodeRequiredFieldMissing`; (d) file present with valid `type: constitution` + `schema_version` → no Finding. Run `go test ./internal/validation/...` and confirm these **fail** (`checkConstitution` doesn't exist yet, or isn't wired into `ValidateProject`).
- [X] T006 [US1] In `internal/example/skills_content_test.go`, add an assertion (near the existing `TestSkillsContent_KnowledgeAndConstitution`) that reads `kit/skills/create-constitution/SKILL.md`'s `Outputs` section (via `extractSection`) and confirms it explicitly shows the required frontmatter fields `type: constitution` and `schema_version`. Run `go test ./internal/example/... -run TestSkillsContent` and confirm it **fails** against the current file.

### Implementation for User Story 1

- [X] T007 [US1] In `internal/artifacts/parser.go`, add a `SchemaVersion int` field to `frontmatterYAML` (yaml tag `schema_version`) and to `Metadata` (mirroring how existing optional fields like `Parent` are already carried through `ParseMetadata`). Run `go test ./internal/artifacts/...` and confirm T004 now **passes**.
- [X] T008 [US1] In `internal/validation/validator.go`, implement `checkConstitution(root string) []Finding`: return no findings if `ai/memory/constitution.md` does not exist (`os.IsNotExist`); otherwise call `artifacts.ParseMetadata` on it and return `CodeFrontmatterMalformed` if parsing fails, `CodeRequiredFieldMissing` if `SchemaVersion == 0` (absent), or no findings if both `type`/`schema_version` are present and valid. Call it once from `ValidateProject`, appending its result to the aggregated `findings` slice, alongside (not inside) the existing `projectEntityTypes` loop and the Task-duplicate-ID check. Run `go test ./internal/validation/...` and confirm T005 now **passes**, and confirm the existing Program/Feature/Spec/Knowledge/Learning/Task tests in the same package are still green (no regression to `004-structural-validation`'s own behavior).
- [X] T009 [US1] In `kit/skills/create-constitution/SKILL.md`'s `## Outputs` section, show the required frontmatter block verbatim (mirroring how the `Quality Requirements` baseline is already shown verbatim in that same section): `---\ntype: constitution\nschema_version: 1\n---`, stated as required before the body sections. Run `go test ./internal/example/... -run TestSkillsContent` and confirm T006 now **passes**.
- [X] T010 [US1] In `kit/skills/create-constitution/SKILL.md`'s `## Procedure` step 4 (the "write or amend `ai/memory/constitution.md` directly" step), add explicit instructions: when creating the file, write the required frontmatter first; when amending an existing file, preserve already-correct frontmatter unchanged and add any missing required field rather than leaving it incomplete (FR-002, FR-003, spec.md).
- [X] T011 [US1] In `kit/skills/create-constitution/SKILL.md`'s `## Validation Rules` section, add: the Constitution's frontmatter must contain `type: constitution` and `schema_version`, alongside the section already stated there ("must contain all seven required sections... must not contain an entity ID").
- [X] T012 [US1] In `kit/skills/create-constitution/SKILL.md`'s `## Deterministic Operations` section, note that `internal validate` (already listed as a required operation, already called in step 5 of the Procedure) now also structurally checks the Constitution's own frontmatter — no new operation is being added, this is a one-line clarification of existing behavior.

**Checkpoint**: At this point, User Story 1 is fully functional and independently testable via `quickstart.md` steps 1–4, with zero dependency on User Story 2.

---

## Phase 3: User Story 2 - create-tasks reports Task dependencies and parallel-safe groups (Priority: P1)

**Goal**: `/create-tasks`'s completion summary explicitly names which Tasks depend on which, and which Tasks have no dependency between them (safe to implement in parallel) — computed from data already recorded per Task, with no new deterministic operation.

**Independent Test**: Follow `quickstart.md` steps 5–6 — generate Tasks for a Spec with both a dependency chain and an independent pair; confirm the completion summary names both distinctly; re-run on an already-extended `tasks.md` and confirm the reporting reflects the full current file.

### Tests for User Story 2 (write first, confirm failing)

- [X] T013 [P] [US2] In `internal/example/skills_content_test.go`, add an assertion (near `TestSkillsContent_PlanTasksImplementAnalyze`) that reads `kit/skills/create-tasks/SKILL.md`'s `Completion Contract` section (via `extractSection`) and confirms it mentions both a Task dependency concept (e.g. "depend") and a parallel-safe-group concept (e.g. "parallel"). Run `go test ./internal/example/... -run TestSkillsContent` and confirm it **fails** against the current file.

### Implementation for User Story 2

- [X] T014 [US2] In `kit/skills/create-tasks/SKILL.md`'s `## Completion Contract` section, extend the **Important findings** (or add a dedicated bullet) to explicitly state: (a) every dependency relationship among the Tasks file's current Tasks (e.g. "Task B depends on Task A"), and (b) every group of Tasks with no dependency between them, named as safe to implement in parallel — both derived from data already recorded per Task (FR-004, FR-005, spec.md), covering the full current `tasks.md` state, not only newly added Tasks in this run (FR-006). When there is nothing to report (e.g. a single Task, or a Spec whose Tasks form one strict sequential chain with no parallel opportunity), state that plainly rather than omitting the topic (FR-007). Run `go test ./internal/example/... -run TestSkillsContent` and confirm T013 now **passes**.

**Checkpoint**: At this point, User Stories 1 AND 2 both work independently — neither shares any implementation file with the other besides `skills_content_test.go`, where each added a distinct, non-conflicting assertion.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Confirm the whole feature regresses nothing and behaves as documented end-to-end.

- [X] T015 [P] Run `go test ./...` in full and confirm no regressions anywhere in the repository, not only in the packages this feature touched.
- [X] T016 Execute `quickstart.md`'s manual validation steps 1–6 end-to-end (or, if a live multi-turn Skill invocation isn't available in this session, a structural trace of each step against the edited `SKILL.md`/Go logic, noting explicitly which form of validation was performed) and confirm every acceptance scenario in `spec.md` (User Stories 1–2) passes as described. NOTE: User Story 1 was validated live — built the `misterspec` binary, ran `misterspec init` in a scratch project (confirming the embedded, edited `create-constitution/SKILL.md` installs with the new frontmatter wording intact), then ran `misterspec internal validate` against: no Constitution (clean), a well-formed Constitution (clean), a Constitution missing `schema_version` (`required_field_missing`), and a Constitution with no frontmatter at all — the exact reported bug scenario (`frontmatter_malformed`). All four matched `data-model.md`'s Constitution Finding table exactly. User Story 2 (prompt content only, no live CLI surface) was validated by structural trace: the edited Completion Contract wording was checked directly against spec.md's FR-004–FR-007 and all four User Story 2 acceptance scenarios — a live multi-turn `/create-tasks` invocation would need a separate agent session.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately; all three tasks are `[P]` (different file sets).
- **User Story 1 (Phase 2)**: Depends only on Setup. Fully independent of User Story 2 — no shared implementation file.
- **User Story 2 (Phase 3)**: Depends only on Setup. Fully independent of User Story 1 — no shared implementation file.
- **Polish (Phase 4)**: Depends on both User Stories being complete (a full regression run needs both sets of changes present).

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on User Story 2. Internal task order within it *is* sequential (test → field → test → check → wire → Skill content), since each step's own test must fail before the next step makes it pass (TDD).
- **User Story 2 (P1)**: No dependency on User Story 1. Its own two tasks (T013 → T014) are sequential for the same test-first reason, but the story as a whole has zero file overlap with User Story 1's own implementation tasks.

### Parallel Opportunities

- **T001, T002, T003** (Setup) — three different file sets, fully parallel.
- **User Story 1 (T004–T012) and User Story 2 (T013–T014) as a whole** — genuinely independent: different Go packages (`internal/artifacts`, `internal/validation`) and a different `SKILL.md` file (`create-constitution` vs. `create-tasks`). Two people could implement both stories fully in parallel, merging at the end — the only shared file is `internal/example/skills_content_test.go`, where each story adds one distinct, non-conflicting test function (T006/T013 land in different functions, so a merge conflict there is trivial to resolve, not a real coordination blocker).
- **T015** (Polish, full regression) has no file dependency on any single prior task, but is ordered last because it needs both stories' changes present to be a meaningful full-repo check.

---

## Parallel Example: Setup + Both User Stories

```bash
# Phase 1 — three people, three files, no coordination needed:
Task: "Read internal/artifacts/parser.go and metadata.go"
Task: "Read internal/validation/validator.go and findings.go"
Task: "Read kit/skills/create-tasks/SKILL.md's Outputs/Completion Contract"

# After Setup — two people, two entirely independent stories:
Person A: T004 -> T005 -> T006 -> T007 -> T008 -> T009 -> T010 -> T011 -> T012  (User Story 1)
Person B: T013 -> T014                                                           (User Story 2)
# Both merge into the same internal/example/skills_content_test.go — distinct
# functions (T006 vs. T013), trivial to reconcile.
```

---

## Implementation Strategy

### MVP First (Either Story Alone)

Both User Stories are independently P1 and independently shippable — there is no "MVP is just US1" ordering forced by dependency here, unlike prior specs this session. A team could ship either one alone as a complete increment:

1. Complete Phase 1: Setup.
2. Complete **either** Phase 2 (User Story 1) **or** Phase 3 (User Story 2) — each is a complete, independently valuable, independently testable increment on its own.
3. **STOP and VALIDATE**: run that story's own `quickstart.md` steps.
4. Add the other story whenever convenient — no ordering constraint between them.

### Incremental Delivery

1. Setup → both stories' target files understood.
2. Add User Story 1 → validate independently (`quickstart.md` steps 1–4) → deterministic Constitution frontmatter guarantee live.
3. Add User Story 2 → validate independently (`quickstart.md` steps 5–6) → dependency/parallel reporting live.
4. Polish → full-repo regression + end-to-end trace of both stories together.

### Parallel Team Strategy

This is the one spec this session where a genuine two-person parallel split is both possible and low-risk: Developer A takes User Story 1 (Go-heavy: `internal/artifacts`, `internal/validation`, `create-constitution/SKILL.md`), Developer B takes User Story 2 (prompt-content only: `create-tasks/SKILL.md`). Both converge only in `internal/example/skills_content_test.go`, in two separate, easily-mergeable test functions.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- T004/T005/T006 and T013 must each fail before their corresponding implementation task makes them pass (test-first, Constitution Principle V).
- Commit after each user-story phase (after T012 for US1, after T014 for US2) rather than after every single task, so each story's diff stays reviewable as one coherent unit.
- Avoid: modifying `checkEntity`'s existing per-entity logic or `projectEntityTypes` — `checkConstitution` (T008) must be additive, called separately from `ValidateProject`, never folded into the existing entity-scanning loop (per `research.md`'s own explicit decision).
