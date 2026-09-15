---

description: "Task list template for feature implementation"
---

# Tasks: Spec-Kit Tooling Improvements

**Input**: Design documents from `/specs/023-speckit-tooling-improvements/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/skill-behavior.md, quickstart.md

**Tests**: No unit-test framework applies to instruction/prompt text (plan.md's own Technical Context) — each user story's own manual verification task, adapted from `quickstart.md`, replaces a written automated test per Constitution Principle V's own adaptation for this feature.

**Organization**: Tasks are grouped by user story, mapping 1:1 to spec.md's 8 user stories. All 8 are independent per spec.md's own Assumptions, but several touch the same file from different angles (`speckit-plan`, `speckit-tasks`, and `speckit-clarify` are each touched by more than one story) — those tasks are marked non-parallel against each other even across stories, per the "same file, sequential" rule.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US8)
- Every task names its exact file path

## Path Conventions

```text
.claude/skills/speckit-*/SKILL.md   # existing skills (modified) + speckit-converge (new)
.specify/templates/checklist-template.md
.specify/extensions.yml
```

---

## Phase 1: Setup

**Purpose**: No new dependency or directory to create before any user story begins.

**Checkpoint**: Nothing to verify — proceed directly to User Story 1.

---

## Phase 2: User Story 1 - A mandatory hook that's described is a hook that runs (Priority: P1) 🎯 MVP

**Goal**: Every `/speckit-*` command with a hook registered actually invokes a mandatory hook's own command (never just describes it), and visibly reports when `.specify/extensions.yml` can't be parsed instead of silently skipping.

**Independent Test**: Trigger a command with a mandatory hook registered (e.g. `/speckit-specify`'s `before_specify`); confirm the hook's own observable effect (a new git branch) exists afterward, not just its description.

- [X] T001 [P] [US1] In `.claude/skills/speckit-specify/SKILL.md`: after each mandatory-hook block (pre- and post-execution), add the "must actually invoke" sentence (research.md #1); replace both "skip hook checking silently and continue normally" lines with an explicit parse-failure report naming the error and which hooks couldn't be checked.
- [X] T002 [P] [US1] Same edit as T001 in `.claude/skills/speckit-analyze/SKILL.md`.
- [X] T003 [US1] Same edit as T001 in `.claude/skills/speckit-plan/SKILL.md`. Not [P] — this file is also touched by US5 (T019) and US8 (T027); do those after this one.
- [X] T004 [US1] Same edit as T001 in `.claude/skills/speckit-tasks/SKILL.md`. Not [P] — also touched by US5 (T020) and US8 (T028); do those after this one.
- [X] T005 [US1] Same edit as T001 in `.claude/skills/speckit-implement/SKILL.md`. Not [P] — also touched by US2 (T009) and US5 (T021); do those after this one.
- [X] T006 [US1] Same edit as T001 in `.claude/skills/speckit-clarify/SKILL.md`. Not [P] — also touched by US5 (T022) and US7 (T030); do those after this one.
- [X] T007 [P] [US1] Same edit as T001 in `.claude/skills/speckit-checklist/SKILL.md`.
- [X] T008 [P] [US1] Same edit as T001 in `.claude/skills/speckit-constitution/SKILL.md`.
- [X] T009 [P] [US1] Same edit as T001 in `.claude/skills/speckit-taskstoissues/SKILL.md`.
- [X] T010 [US1] Manual verification per `quickstart.md` §1: run `/speckit-specify` on a disposable throwaway feature from `dev`, confirm a new branch was actually created (not just described); corrupt `.specify/extensions.yml`, run any `/speckit-*` command, confirm the parse error is visibly reported. Depends on T001-T009 (this task references speckit-implement/checklist which are also edited in T005/T007, but T005's edit column here is US1's own portion only — the file may still be mid-edit for other stories; run this check against a fully-merged file if sequencing all stories together).

**Checkpoint**: All 9 skills actually invoke their own mandatory hooks and visibly report a broken hook config — the demonstrated bug this whole feature started from is fixed.

---

## Phase 3: User Story 2 - Checklists stay a read-only quality gate during implementation (Priority: P1)

**Goal**: `/speckit-implement` never modifies a checklist's own checkbox markers; `/speckit-checklist` always writes new items unchecked; the checklist template documents this contract explicitly.

**Independent Test**: Run `/speckit-implement` against a feature with a partially-checked checklist; confirm the file is byte-for-byte unchanged afterward.

- [X] T011 [P] [US2] In `.specify/templates/checklist-template.md`: add the `Review Ownership` and `Marker Semantics` lines and the note that `/speckit-implement` reads checklist state as a gate and must not modify markers (research.md #2).
- [X] T012 [US2] In `.claude/skills/speckit-implement/SKILL.md`'s existing checklist-status step: add "Treat checklist markers as a read-only gate: scan checkbox state, report status, and ask before proceeding when needed; do NOT modify checklist files or markers" and the `checklists/requirements.md`-vs-custom-checklist ownership distinction (research.md #2). Depends on T005 (same file, US1's edit lands first).
- [X] T013 [P] [US2] In `.claude/skills/speckit-checklist/SKILL.md`: add "This command generates or appends checklist items; it MUST NOT mark generated items `[x]`" plus the reviewer-ownership framing (research.md #2). Depends on T007 (same file, US1's edit lands first) — may run in parallel with T012 since they touch different files.
- [X] T014 [US2] Manual verification per `quickstart.md` §2: `md5sum` a checklist file with mixed checked/unchecked items, run `/speckit-implement`, `md5sum` again — confirm identical. Depends on T011, T012, T013.

**Checkpoint**: Checklists are provably untouched by implementation work; new checklist items are always born unchecked.

---

## Phase 4: User Story 3 - Re-running task-to-issue sync never duplicates issues (Priority: P2)

**Goal**: `/speckit-taskstoissues` skips any task that already has a matching GitHub issue.

**Independent Test**: Run `/speckit-taskstoissues` twice in a row against an unchanged `tasks.md`; confirm the second run creates zero new issues.

- [X] T015 [US3] In `.claude/skills/speckit-taskstoissues/SKILL.md`: add the deduplication algorithm (research.md #3) — `list_issues` with no `state` filter, `perPage: 100`, `after`-cursor pagination stopping once every task ID is matched or pages are exhausted; match issue titles against `\bT\d{3,}\b`; skip and report `"<ID> already has an issue, skipping"` for any already-matched task ID. Depends on T009 (same file, US1's edit lands first).
- [ ] T016 [US3] Manual verification per `quickstart.md` §3: run `/speckit-taskstoissues` on a disposable feature with unsynced tasks, note the issue numbers created, run it again with no changes, confirm zero new issues and the correct skip message per task. Depends on T015.

**Checkpoint**: Re-running task-to-issue sync is idempotent.

---

## Phase 5: User Story 4 - Constitution updates never silently absorb unrelated work (Priority: P2)

**Goal**: `/speckit-constitution` never executes a feature/code/deploy request mixed into its own prompt — it defers and surfaces such requests instead.

**Independent Test**: Send `/speckit-constitution` a prompt mixing a real governance change with an unrelated "also implement X" request; confirm the governance change lands, the unrelated request is not performed, and it's named under a `Next Actions` section.

- [X] T017 [US4] In `.claude/skills/speckit-constitution/SKILL.md`: add the `## Scope Guard` section (research.md #4) — classify input as constitution content vs. non-governance intent; never execute feature/code/refactor/deploy requests found mixed in; extract them as deferred intents; surface them in a `Next Actions` section (omitted when empty) naming the appropriate follow-up command without invoking it; ask for clarification when classification is genuinely ambiguous. Depends on T008 (same file, US1's edit lands first).
- [ ] T018 [US4] Manual verification per `quickstart.md` §4: send a mixed governance+implementation prompt; confirm the governance change applies, no code changes occur, and the implementation request appears under `Next Actions`. Depends on T017.

**Checkpoint**: `/speckit-constitution` never wanders outside its own governance lane.

---

## Phase 6: User Story 5 - Each core lifecycle command self-checks before reporting done (Priority: P3)

**Goal**: `speckit-specify`, `speckit-plan`, `speckit-tasks`, `speckit-implement`, and `speckit-clarify` each report an explicit self-check of their own required outputs before declaring completion.

**Independent Test**: Run any of the five; confirm its final report includes a `Done When` checklist naming its own required outputs, each confirmed.

- [X] T019 [US5] In `.claude/skills/speckit-plan/SKILL.md`: add a `## Done When` section (research.md #5) — "Plan workflow executed and design artifacts generated"; "Extension hooks dispatched or skipped..."; "Completion reported to user with branch, plan path, and generated artifacts". Depends on T003 (US1 lands first on this file).
- [X] T020 [US5] In `.claude/skills/speckit-tasks/SKILL.md`: add a `## Done When` section — "tasks.md generated with all phases, task IDs, and file paths"; hooks line; "Completion reported to user with task count, story breakdown, and MVP scope". Depends on T004.
- [X] T021 [US5] In `.claude/skills/speckit-implement/SKILL.md`: add a `## Done When` section — "All tasks in tasks.md completed and marked `[X]`"; "Implementation validated against specification, plan, and test coverage"; hooks line; "Completion reported to user with summary of completed work". Depends on T005, T012 (US1 and US2 land first on this file).
- [X] T022 [US5] In `.claude/skills/speckit-clarify/SKILL.md`: add a `## Done When` section — "Spec ambiguities identified and clarifications integrated into spec file"; "Spec quality checklist re-validated against updated spec (if it exists)"; hooks line; "Completion reported to user with questions answered, sections touched, checklist status, and coverage summary". Depends on T006 (US1 lands first on this file).
- [X] T023 [P] [US5] In `.claude/skills/speckit-specify/SKILL.md`: add a `## Done When` section — "Specification written to SPEC_FILE and validated against quality checklist"; hooks line; "Completion reported to user with feature directory, spec file path, and checklist results". Depends on T001.
- [X] T024 [US5] Manual verification per `quickstart.md` §5: run `/speckit-plan` (or any of the five); confirm its completion report includes the `Done When` checklist, each item confirmed. Depends on T019-T023.

**Checkpoint**: Every core lifecycle command's completion report is now self-checking, not just self-declaring.

---

## Phase 7: User Story 6 - A drift check between intent and code, run on demand (Priority: P3)

**Goal**: A new `/speckit-converge` command compares a feature's spec/plan/tasks (plus the Constitution) against the actual codebase, reports classified gaps, and appends only new tasks — never rewriting existing artifacts.

**Independent Test**: Run it against a feature whose code already fully satisfies its own spec/plan/tasks; confirm it reports no gaps and leaves every file unchanged. Run it against a feature with a deliberate gap; confirm the gap is reported and only a new `## Phase N: Convergence` section is appended to `tasks.md`.

- [X] T025 [US6] Create `.claude/skills/speckit-converge/SKILL.md` (research.md #6, contracts/skill-behavior.md): port upstream's `converge.md` — frontmatter matching every other `speckit-*` skill (`name`, `description`, `argument-hint`, `compatibility`, `metadata`); prerequisite check via `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` (which already covers `plan.md`/`tasks.md`) plus an explicit extra check for `spec.md` (the script has no `--require-spec` flag), stopping with a message naming the specific missing-prerequisite command when any of the three is absent; the append-only operating constraint (never modifies `spec.md`/`plan.md`/existing tasks; byte-for-byte-unchanged `tasks.md` when converged); the 4-step execution flow (build intent inventory, assess codebase, classify by gap-type and severity, present in-session findings table); the append contract (`## Phase N: Convergence`, continuing task IDs, Constitution violations always CRITICAL and listed first); this repo's own hook-boilerplate text (research.md #1) for `before_converge`/`after_converge`.
- [X] T026 [P] [US6] In `.specify/extensions.yml`: add `before_converge` and `after_converge` hook keys, each with the same optional, `enabled: true`, `speckit.git.commit` entry every other command's hook keys already have (research.md #6, data-model.md).
- [X] T027 [US6] Manual verification per `quickstart.md` §6: against a feature with a deliberately introduced gap, confirm `spec.md`/`plan.md` are unchanged and exactly one new Convergence phase is appended naming the gap; re-run with the gap fixed (or against an already-satisfied feature) and confirm a byte-for-byte-unchanged `tasks.md` and a "✅ Converged" report. Depends on T025, T026.

**Checkpoint**: A drift check between intent and code exists, is safe (append-only), and correctly reports both the "gap found" and "nothing to do" outcomes.

---

## Phase 8: User Story 7 - Clarification questions are genuinely questions, and checklists stay current (Priority: P3)

**Goal**: `/speckit-clarify` never presents a bare label/ID as a question, always states why a question matters, and re-validates the spec quality checklist once a clarification round completes.

**Independent Test**: Run `/speckit-clarify` against a spec with a genuine ambiguity; confirm each question is a complete sentence with a "why it matters" line, and that the checklist re-validation report appears once answers are integrated.

- [X] T028 [US7] In `.claude/skills/speckit-clarify/SKILL.md`: add the question-writing quality rules (research.md #7) — full interrogative ending in `?`; never a bare requirement ID or topic label as the question; only a parenthesized ID permitted after the `?`; a one-sentence "why it matters" statement immediately following. Depends on T006, T022 (US1 and US5 land first on this file).
- [X] T029 [US7] In the same file: add the post-clarify checklist re-validation step (research.md #7) — if `checklists/requirements.md` exists, re-evaluate each checkbox against the updated spec, toggle only markers whose state actually changed, and report newly-passing items and regressions in the completion report. Depends on T028 (same file, sequential).
- [X] T030 [US7] Manual verification per `quickstart.md` §7: run `/speckit-clarify` against a spec with a real ambiguity; confirm every question reads as a full sentence with a "why it matters" line; confirm the completion report shows before/after checklist pass counts. Depends on T028, T029.

**Checkpoint**: Clarification questions are legible on their own, and the spec quality checklist never silently goes stale after clarification.

---

## Phase 9: User Story 8 - Plans and task lists stay precise and appropriately scoped (Priority: P3)

**Goal**: `quickstart.md` never grows into a second task list; tasks touching a constrained field quote that constraint verbatim.

**Independent Test**: Generate a plan for a feature with a constrained data-model field; confirm `quickstart.md` has no full implementation code, and the resulting task quotes the constraint verbatim.

- [X] T031 [US8] In `.claude/skills/speckit-plan/SKILL.md`'s quickstart-generation step: add "Do not include full implementation code, model/service/controller bodies, migrations, or complete test suites. Keep this artifact as a validation/run guide" (research.md #8). Depends on T003, T019 (US1 and US5 land first on this file).
- [X] T032 [US8] In `.claude/skills/speckit-tasks/SKILL.md`'s "From Data Model" task-generation rule: add "For each field with constraints in data-model.md (max length, nullable/required, enum values, validation rules), quote the constraint verbatim in the task description so it is not left to implementation-time discretion" (research.md #8). Depends on T004, T020 (US1 and US5 land first on this file) — may run in parallel with T031 since they touch different files.
- [X] T033 [US8] Manual verification per `quickstart.md` §8: generate a plan/task list for a feature with a constrained field; confirm `quickstart.md` contains no full implementation code and the relevant task quotes the constraint verbatim. Depends on T031, T032.

**Checkpoint**: Planning and task-generation output stays precise and appropriately scoped.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency check across all 8 stories.

- [X] T034 [P] Grep all 9 modified `SKILL.md` files for the literal "must actually invoke" sentence and confirm it appears after every mandatory-hook block (2 occurrences per file: pre- and post-execution); confirm zero remaining occurrences of "skip hook checking silently and continue normally".
- [ ] T035 Full manual regression: re-run every `quickstart.md` scenario (§1-§8) once, end to end, against a single disposable throwaway feature, to confirm no story's own change regressed another's.
- [X] T036 Reconcile `specs/023-speckit-tooling-improvements/contracts/skill-behavior.md` against the actual final wording in each skill; fix any drift.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do.
- **User Story 1 (Phase 2)**: No dependency — this feature's own MVP, and the file every other story's own tasks depend on landing first (same-file sequencing).
- **User Stories 2-8 (Phases 3-9)**: Each independent of the others at the *requirement* level (spec.md's own Assumptions), but any task touching `speckit-plan`, `speckit-tasks`, `speckit-implement`, or `speckit-clarify` must land after User Story 1's own edit to that same file (see each task's own "Depends on" note).
- **Polish (Phase 10)**: Depends on all 8 stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Independent — the true starting point, and a same-file prerequisite for parts of US2, US5, US7, US8.
- **User Stories 2, 3, 4 (P1/P2)**: Each independent of every other story except their own same-file ordering against US1.
- **User Stories 5, 6, 7, 8 (P3)**: Independent of each other; US5/US7 share `speckit-clarify` (sequential); US5/US8 share `speckit-plan` and `speckit-tasks` (sequential).

### Parallel Opportunities

- T001, T002, T007, T008, T009 (US1's edits to files no other story touches) can all proceed in parallel once the pattern is agreed.
- T011 and T013 (US2's template and speckit-checklist edits) can proceed in parallel.
- T023 (US5's speckit-specify edit) can proceed in parallel with anything, once T001 lands.
- T026 (US6's extensions.yml edit) can proceed in parallel with T025 (the new skill file itself).
- T031 and T032 (US8's plan/tasks edits) can proceed in parallel with each other (different files), once their own US1/US5 prerequisites land.
- Within Polish: T034 alone is parallel-safe; T035/T036 are sequential checks.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: User Story 1.
2. **STOP and VALIDATE**: `quickstart.md` §1 passes — a mandatory hook's own effect is now observable, and a broken hook config is now visibly reported.

### Incremental Delivery

1. User Story 1 → the demonstrated bug is fixed (MVP).
2. User Story 2 → checklists become a provable read-only gate.
3. User Story 3 → task-to-issue sync becomes idempotent.
4. User Story 4 → constitution updates never silently absorb unrelated work.
5. User Story 5 → every core lifecycle command self-checks before reporting done.
6. User Story 6 → `/speckit-converge` exists as a new, safe drift check.
7. User Story 7 → clarification questions are legible, checklists stay current.
8. User Story 8 → plan/task output stays precise and appropriately scoped.
9. Polish (Phase 10).

### Team Strategy

Every story after User Story 1 can be worked on independently and in
parallel by different people, as long as whoever touches
`speckit-plan`/`speckit-tasks`/`speckit-implement`/`speckit-clarify` waits
for User Story 1's own edit to that same file to land first.

---

## Notes

- This feature touches only `.claude/skills/speckit-*/SKILL.md`,
  `.specify/templates/checklist-template.md`, and `.specify/extensions.yml`
  — `kit/skills/` (the product misterspec ships to its own end users) and
  every Go package under `internal/`/`cmd/` are explicitly untouched
  (spec.md's own Assumptions).
- No unit tests are written for this feature — there is no test framework
  for instruction/prompt text; each story's own manual verification task
  (drawn from `quickstart.md`) is this feature's adapted form of
  Constitution Principle V's test-first discipline.
- Commit after each phase.
