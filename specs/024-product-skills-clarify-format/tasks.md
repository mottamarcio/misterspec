---

description: "Task list template for feature implementation"
---

# Tasks: Interactive Clarification and Richer Output for Product Skills

**Input**: Design documents from `/specs/024-product-skills-clarify-format/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/skill-behavior.md, quickstart.md

**Tests**: No unit-test framework applies to instruction/prompt text (plan.md's own Technical Context) — each user story's own manual verification task, drawn from `quickstart.md`, is this feature's adapted form of Constitution Principle V, matching spec 023's own precedent.

**Organization**: Tasks are grouped by user story, mapping 1:1 to spec.md's 5 user stories. User Story 5 (formatting) touches all 9 product Skills, four of which (`create-specs`, `create-plan`, `create-tasks`, `analyze`) are also touched by User Stories 1-4 — those four files' formatting edits are sequenced after their own behavioral edit lands first (same-file rule), never in parallel with it.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Every task names its exact file path

## Path Conventions

```text
kit/skills/*/SKILL.md   # 9 product Skill files, all under this one directory
```

---

## Phase 1: Setup

**Purpose**: No new dependency or directory to create before any user story begins.

**Checkpoint**: Nothing to verify — proceed directly to User Story 1.

---

## Phase 2: User Story 1 - Ambiguous requirements get asked about, not silently resolved (Priority: P1) 🎯 MVP

**Goal**: `/create-specs` presents a genuinely ambiguous requirement boundary as a question with concrete options (one recommended) and waits for the user's answer, instead of silently narrowing the requirement.

**Independent Test**: Run `/create-specs` against a Feature with a genuinely ambiguous requirement; confirm a question with options and a recommended pick is presented and the answer is reflected directly in the written requirement.

- [X] T001 [US1] In `kit/skills/create-specs/SKILL.md`: replace the `## Interaction Rules` section's current "prefer a narrower, clearly testable requirement plus a named open question" guidance with the new interactive-question behavior (research.md #1): for a requirement boundary with multiple reasonable interpretations and materially different implications, present a full-sentence question with 2-4 concrete options (one marked "(Recommended)" with a one-sentence reason) and wait for the answer; low-stakes ambiguity with an obvious default keeps today's silent default-and-annotate behavior unchanged (FR-002). Add a note in `## Procedure` step 4 (where requirement text is written) pointing to this rule. Add to `## Decision Rules` or `## Interaction Rules`: when the interactive quota is reached with real ambiguity still unresolved, record it under the new Spec's own `## Unresolved Questions` section (FR-003, research.md #3).
- [X] T002 [US1] Manual verification per `quickstart.md` §1 (a genuinely ambiguous requirement triggers the question/options/recommended structure) and §3 (a low-stakes ambiguity is still handled silently, unchanged). Depends on T001.

**Checkpoint**: User Story 1 is independently complete and testable — `create-specs` never silently resolves a genuine requirement-boundary ambiguity.

---

## Phase 3: User Story 2 - Planning-level forks get asked about, not silently resolved (Priority: P1)

**Goal**: `/create-plan` presents a genuine strategic fork as a question with concrete options (one recommended), instead of silently choosing and recording the tradeoff as a note.

**Independent Test**: Run `/create-plan` against a Spec with a genuine strategic fork; confirm the same question/options/recommended structure as User Story 1, applied to the planning decision.

- [X] T003 [US2] In `kit/skills/create-plan/SKILL.md`: replace the `## Interaction Rules` section's current "name the tradeoff in Risks or Assumptions rather than silently picking one" guidance with the same interactive-question behavior as T001 (research.md #1), applied to a Plan decision with more than one reasonable strategy and materially different tradeoffs (FR-004); a decision with only one reasonable approach keeps today's behavior unchanged.
- [X] T004 [US2] Manual verification per `quickstart.md` §2. Depends on T003.

**Checkpoint**: User Story 2 is independently complete and testable — `create-plan` never silently resolves a genuine strategic fork.

---

## Phase 4: User Story 3 - Tasks quote a requirement's own constraint, not just its ID (Priority: P2)

**Goal**: `/create-tasks` quotes a Requirement's own explicit constraint verbatim in the Task description, in addition to the existing `SPEC-###:R#` reference.

**Independent Test**: Run `/create-tasks` against a Spec with a constrained requirement; confirm the resulting Task's description quotes the constraint's exact text.

- [X] T005 [US3] In `kit/skills/create-tasks/SKILL.md`'s `## Decision Rules` section: add "When a Requirement states an explicit constraint (a specific limit, required format, or measurable threshold), quote that constraint's exact text in the Task description, in addition to the `SPEC-###:R#` reference — not instead of it" (research.md #4, FR-005). A Requirement with no explicit constraint keeps today's behavior unchanged (ID reference only, FR-005's own second clause).
- [X] T006 [US3] Manual verification per `quickstart.md` §4. Depends on T005.

**Checkpoint**: User Story 3 is independently complete and testable — a Task touching a constrained Requirement always carries that constraint's own verbatim text.

---

## Phase 5: User Story 4 - Analyze can offer to track a remaining gap as a Task, with the user's confirmation (Priority: P3)

**Goal**: `/analyze` offers, with explicit user confirmation, to append exactly one narrowly-scoped remediation Task when it finds an "implementation incomplete" gap — never automatically, never for Spec/Plan-layer gaps.

**Independent Test**: Run `/analyze` against a Spec with a deliberate implementation gap; confirm it asks before appending, appends exactly one Task only on confirmation, and makes no other write; confirm it never offers this for a Spec/Plan-layer gap.

- [X] T007 [US4] In `kit/skills/analyze/SKILL.md`: widen `## Allowed Modifications` to add the one narrow, gated exception (data-model.md's own before/after table) — appending exactly one new `## TASK-NNN` entry to the existing Tasks artifact, only on the user's explicit per-finding confirmation, and only for a finding whose responsible layer is "implementation incomplete" (FR-006, FR-007, FR-008). Add a matching clarifying line to `## Forbidden Mutations` (this is an append, never a modification, and never automatic). Add a new step to `## Procedure`, immediately after a `Result: fail` with responsible layer "implementation incomplete" is recorded: present the gap and ask whether to append a tracking Task, using the same recommended-option interaction pattern as T001/T003 (research.md #5); the appended Task's shape matches data-model.md's own template (`Serves`, `Depends on`, `Verification`, `Origin` lines), with the next `TASK-NNN` number scanned from existing headings (the same convention `create-tasks` already documents, no new ID allocator).
- [X] T008 [US4] Manual verification per `quickstart.md` §5 (confirm-then-append and decline-then-nothing, both against an implementation-incomplete gap) and §6 (a Spec/Plan-layer gap never offers the append). Depends on T007.

**Checkpoint**: User Story 4 is independently complete and testable — `analyze` never appends a Task without explicit, per-finding confirmation, and never for a Spec/Plan-layer gap.

---

## Phase 6: User Story 5 - Every slash command's completion summary is scannable, not a wall of prose (Priority: P1)

**Goal**: Every one of the 9 product Skills' completion (and failure/stop) reports render `Artifacts` as a table/single bullet and `Important findings`/`Attention` as bullet lists, instead of prose paragraphs.

**Independent Test**: Run any of the 9 Skills to completion; confirm its own report uses the structured formatting instead of a prose paragraph.

- [X] T009 [US5] In `kit/skills/create-specs/SKILL.md`'s `## Completion Contract` section, immediately after "Every invocation ends with a concise operational summary naming:", insert the formatting instruction (research.md #6, verbatim): render `Artifacts` as a Markdown table when 2+ artifacts are involved (or a single bullet for exactly 1), and `Important findings`/`Attention` as bullet lists — applies equally to a failure/stop report. Depends on T001 (same file, US1's own edit lands first).
- [X] T010 [US5] Same edit as T009 in `kit/skills/create-plan/SKILL.md`. Depends on T003 (same file, US2's own edit lands first).
- [X] T011 [US5] Same edit as T009 in `kit/skills/create-tasks/SKILL.md`. Depends on T005 (same file, US3's own edit lands first).
- [X] T012 [US5] Same edit as T009 in `kit/skills/analyze/SKILL.md`. Depends on T007 (same file, US4's own edit lands first).
- [X] T013 [P] [US5] Same edit as T009 in `kit/skills/create-program/SKILL.md`.
- [X] T014 [P] [US5] Same edit as T009 in `kit/skills/create-feature/SKILL.md`.
- [X] T015 [P] [US5] Same edit as T009 in `kit/skills/create-knowledge-base/SKILL.md`.
- [X] T016 [P] [US5] Same edit as T009 in `kit/skills/create-constitution/SKILL.md`.
- [X] T017 [P] [US5] Same edit as T009 in `kit/skills/implement/SKILL.md`.
- [X] T018 [US5] Manual verification per `quickstart.md` §7 (run all 9 Skills at least once each; confirm each completion report is structured, not prose). Depends on T009-T017.

**Checkpoint**: User Story 5 is independently complete and testable — every product Skill's own completion report is scannable.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency check across all 5 stories.

- [X] T019 [P] Grep all 9 `kit/skills/*/SKILL.md` files for the exact formatting-instruction sentence from T009 and confirm it appears exactly once, immediately after "Every invocation ends with a concise operational summary naming:", in every file.
- [ ] T020 Full manual regression: re-run every `quickstart.md` scenario (§1-§7) once, end to end, against a single disposable throwaway test project, to confirm no story's own change regressed another's.
- [X] T021 Reconcile `specs/024-product-skills-clarify-format/contracts/skill-behavior.md` against the actual final wording in each Skill; fix any drift.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do.
- **User Stories 1-4 (Phases 2-5)**: Each independent of the others at the requirement level — different files, different concerns.
- **User Story 5 (Phase 6)**: Independent at the requirement level, but four of its nine file edits (T009, T010, T011, T012) must land after User Stories 1-4's own edits to those same four files.
- **Polish (Phase 7)**: Depends on all 5 stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Independent — the true starting point.
- **User Story 2 (P1)**: Independent of User Story 1 (different file), same priority.
- **User Story 3 (P2)**: Independent of 1 and 2.
- **User Story 4 (P3)**: Independent of 1, 2, and 3 — the most constitution-sensitive, so it ships after the others are proven out, though nothing technically blocks it running first.
- **User Story 5 (P1)**: Depends on 1-4 only for same-file sequencing on 4 of its 9 files; otherwise independent.

### Parallel Opportunities

- T001, T003, T005, T007 (the four behavioral edits, User Stories 1-4) can all proceed in parallel — four different files, no shared dependency.
- T013, T014, T015, T016, T017 (User Story 5's five Skill-only-touched files) can all proceed in parallel with each other and with T001/T003/T005/T007, since they share no file with any of them.
- Within Polish: T019 alone is parallel-safe; T020/T021 are sequential checks.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: User Story 1.
2. **STOP and VALIDATE**: `quickstart.md` §1 and §3 pass — `create-specs` asks about genuine ambiguity and still defaults silently for low-stakes cases.

### Incremental Delivery

1. User Story 1 → `create-specs` asks instead of silently resolving (MVP).
2. User Story 2 → same fix in `create-plan`.
3. User Story 3 → `create-tasks` quotes constraints verbatim.
4. User Story 4 → `analyze` offers confirmed, narrowly-scoped remediation Tasks.
5. User Story 5 → every Skill's completion report becomes scannable.
6. Polish (Phase 7).

### Team Strategy

User Stories 1-4 can be worked on by four different people in full
parallel (four different files). User Story 5 can start immediately on
its five Skill-only-touched files, and its remaining four files as soon
as each corresponding behavioral story lands.

---

## Notes

- This feature touches only `kit/skills/*/SKILL.md` — no Go code, no
  `kit/templates/*.tmpl`, no `.specify/`/`.claude/` dev-tooling (spec.md's
  own Assumptions; verified during specification that no template change
  is needed since the Spec artifact already has an `## Unresolved
  Questions` section).
- User Story 4's own widening of `analyze`'s mutation boundary is the
  single most constitution-sensitive change in this feature — keep its
  three gating conditions (gap-type-scoped, user-confirmed, append-only)
  intact exactly as documented; do not generalize them during
  implementation.
- Commit after each phase.
