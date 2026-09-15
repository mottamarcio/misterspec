---

description: "Task list template for feature implementation"
---

# Tasks: Dogfooding and Evaluation

**Input**: Design documents from `/specs/019-dogfooding-evaluation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/evaluation-protocol.md, quickstart.md

**Tests**: Not applicable (plan.md's own Technical Context) — this feature ships no deterministic logic of its own; its own "verification" is the Dogfooding Report being internally consistent and traceable to real `internal context` output or a directly-observed live session, plus 001-018's own full suites staying green (untouched) throughout.

**Organization**: Tasks are grouped by user story. **User Story 1** (static evaluation against the fixture) and **User Story 2** (a live Claude Code session) can both be performed directly within this very session, since this assistant *is* Claude Code with the tool access the Skill's own instructions require. **User Story 3** (a live Antigravity session) requires the user's own hands-on action outside this session — its tasks are flagged accordingly and block only the parts of **User Story 4** that need its evidence specifically. Setup and Foundational are shared prerequisites for all four stories.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US4)
- Every task names its exact file path

## Path Conventions

```text
specs/019-dogfooding-evaluation/fixture/   # the real-history-encoding project (research.md #1)
specs/019-dogfooding-evaluation/report.md  # the Dogfooding Report (research.md #5)
```

---

## Phase 1: Setup (Fixture Construction)

**Purpose**: Build the one real, filesystem-backed misterspec project every later phase queries — encoding this repository's own actual, already-known feature dependency graph (research.md #1), never invented data.

- [X] T001 [P] Create `specs/019-dogfooding-evaluation/fixture/.misterspec/config.yaml` (standard schema, matching `testutil.DefaultConfigYAML`'s own shape).
- [X] T002 [P] Create `specs/019-dogfooding-evaluation/fixture/ai/memory/constitution.md`, condensed from the real `.specify/memory/constitution.md` Principles III, IV, and IX only (data-model.md).
- [X] T003 [P] Create `specs/019-dogfooding-evaluation/fixture/ai/programs/PRG-001/program.md` and `.../features/FEAT-001/feature.md` ("Context Engine" — data-model.md).
- [X] T004 [P] Create `specs/019-dogfooding-evaluation/fixture/ai/knowledge/KNOW-001-constitution-principles.md` and `KNOW-002-agent-skills-conventions.md`, condensed from real content (data-model.md).
- [X] T005 [P] Create `specs/019-dogfooding-evaluation/fixture/ai/memory/learnings/LRN-001-budget-flag-correction.md`, condensed from 017's own real `--budget` string-flag correction (data-model.md). **Correction during implementation**: Learning's `status` field has a restricted vocabulary (`candidate`/`promoted`/`dismissed`) discovered via `internal status`'s own `structural_errors` count — used `status: promoted` (this correction is itself real, already-adopted knowledge), not `active`.
- [X] T006 [P] Create `SPEC-006` (`.../specs/SPEC-006/spec.md`), `depends_on: []`, condensed from `specs/006-agent-adapter/spec.md`'s own real Summary.
- [X] T007 [P] Create `SPEC-011` (`.../specs/SPEC-011/spec.md`), `depends_on: []`, condensed from `specs/011-wikilink-foundation/spec.md`'s own real Summary.
- [X] T008 [P] Create `SPEC-012` (`.../specs/SPEC-012/spec.md`), `depends_on: [SPEC-011]`, condensed from `specs/012-references-backlinks/spec.md`'s own real Summary.
- [X] T009 [P] Create `SPEC-013` (`.../specs/SPEC-013/spec.md`), `depends_on: [SPEC-011]`, condensed from `specs/013-document-model-chunking/spec.md`'s own real Summary.
- [X] T010 [P] Create `SPEC-014` (`.../specs/SPEC-014/spec.md`), `depends_on: [SPEC-012, SPEC-013]`, condensed from `specs/014-sqlite-index/spec.md`'s own real Summary.
- [X] T011 [P] Create `SPEC-015` (`.../specs/SPEC-015/spec.md`), `depends_on: [SPEC-011, SPEC-012, SPEC-013, SPEC-014]`, condensed from `specs/015-context-collector/spec.md`'s own real Summary, wikilinking `[[KNOW-001]]`.
- [X] T012 [P] Create `SPEC-016` (`.../specs/SPEC-016/spec.md`), `depends_on: [SPEC-015]`, condensed from `specs/016-ranking-budgeting/spec.md`'s own real Summary, wikilinking `[[KNOW-001]]`.
- [X] T013 [P] Create `SPEC-017` (`.../specs/SPEC-017/spec.md`), `depends_on: [SPEC-014, SPEC-015, SPEC-016]`, condensed from `specs/017-internal-context-command/spec.md`'s own real Summary, wikilinking `[[LRN-001]]` and `[[KNOW-001]]`.
- [X] T014 [P] Create `SPEC-018` (`.../specs/SPEC-018/spec.md`), `depends_on: [SPEC-006, SPEC-017]`, condensed from `specs/018-multi-agent-skill-integration/spec.md`'s own real Summary, wikilinking `[[KNOW-002]]` and `[[KNOW-001]]`.

**Checkpoint**: `misterspec internal status --dir specs/019-dogfooding-evaluation/fixture` reports all 9 Specs, `internal validate` reports zero structural findings.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Install both live-testable agents' Skills into the fixture — needed by User Story 2/3 (live sessions), not by User Story 1 (which only needs `internal context`, no installed Skills).

**⚠️ CRITICAL**: No User Story 2/3 task may begin until this phase completes. User Story 1 has no dependency on this phase and may proceed in parallel.

- [X] T015 [P] Install the `claude-code` adapter's Skills into the fixture (`internal/agents/claude`, via `kit.SkillsFS`) — `specs/019-dogfooding-evaluation/fixture/.claude/skills/`. Performed via a throwaway `cmd/dogfood-install/main.go` calling `builtin.Default().Get("claude-code").Install(...)` (`internal/` packages can't be imported from outside the module, so the installer had to run from inside it), deleted immediately after use — not committed, since it has no product role (research.md #3's own "one-off evaluation artifact" reasoning applies equally to the installer script itself).
- [X] T016 [P] Install the `agy` adapter's Skills into the fixture (`internal/agents/agy`) — `specs/019-dogfooding-evaluation/fixture/.agents/skills/`. Same throwaway tool as T015 (both adapters installed in one run).
- [X] T017 Regression checkpoint: confirm both installs succeeded and neither's own files were corrupted by the other (018's own coexistence guarantee, since `.agents/skills` is `agy`'s own target — no collision here since `codex` was not installed). Depends on T015, T016.

**Checkpoint**: Foundation ready — the fixture has both Skill sets installed. User Story 2 and User Story 3 may begin.

---

## Phase 3: User Story 1 - Validate Retrieval Against This Project's Own Real History (Priority: P1) 🎯 MVP

**Goal**: For every fixture Spec, confirm `internal context`'s own returned pack contains every real, historically-known dependency, and record any omission or over-inclusion as a specific finding.

**Independent Test**: Run `internal context` against each of the 9 fixture Specs; compare `context.items` against data-model.md's own `depends_on` table.

- [X] T018 [US1] Run `misterspec internal context <id> --intent planning --dir specs/019-dogfooding-evaluation/fixture` for all 9 Specs (T006-T014); save each raw JSON response inline in `specs/019-dogfooding-evaluation/report.md`'s own "User Story 1 — Raw Responses" section. Depends on T001-T014 (Setup). **Extended during implementation**: the default budget (6000) never trimmed anything for this small fixture (0% reduction on every Spec), so two additional `SPEC-017` runs (`--budget 700`, `--budget 300`) were performed to actually exercise budget-constrained behavior — real, valuable evidence the static protocol as originally scoped would have missed entirely (see report.md Findings F3-F5).
- [X] T019 [P] [US1] Analyze `SPEC-006`, `SPEC-011`, `SPEC-012`, `SPEC-013` (no or single-dependency Specs): confirm each real `depends_on` entry (data-model.md) appears in its own `context.items`; record any omission or over-inclusion as a Dogfooding Finding row in `report.md`. Depends on T018.
- [X] T020 [P] [US1] Analyze `SPEC-014`, `SPEC-015` (multi-dependency Specs): same check against their own real `depends_on` sets; record findings. Depends on T018.
- [X] T021 [P] [US1] Analyze `SPEC-016`, `SPEC-017` (ranking/command Specs, including `[[LRN-001]]`/`[[KNOW-001]]` wikilinks): same check, also confirming the wikilinked Knowledge/Learning appear; record findings. Depends on T018.
- [X] T022 [P] [US1] Analyze `SPEC-018` (multi-dependency + `[[KNOW-002]]` wikilink): same check; record findings. Depends on T018.
- [X] T023 [US1] Write the "User Story 1 — Summary" section of `report.md`: total Specs evaluated, total findings by kind (`omission`/`over-inclusion`), and each `context.diagnostics` figure quoted per Spec. Depends on T019-T022.

**Checkpoint**: User Story 1 is independently complete — every fixture Spec has a recorded pass/fail judgment, entirely without any live agent session (SC-001, SC-002).

---

## Phase 4: User Story 2 - Dogfood a Real Skill Invocation With Claude Code (Priority: P2)

**Goal**: Directly observe, within this very session (this assistant *is* Claude Code), one live invocation of a Context-Pack-Aware Skill against the fixture, recording request behavior, sufficiency, and elapsed time.

**Independent Test**: Follow `kit/skills/create-plan/SKILL.md`'s own real, installed instructions (T015) against `SPEC-018` in the fixture; record whether `internal context` was called first, whether its pack was sufficient, and how long the whole invocation took.

- [X] T024 [US2] Following `create-plan`'s own installed instructions, run `internal resolve SPEC-018`, `internal inspect SPEC-018`, then `internal context SPEC-018 --intent planning` (in that order) against the fixture, timing the whole sequence; record in `report.md`'s own "User Story 2 — Claude Code" section whether the pack alone would have been sufficient to draft `SPEC-018`'s own Plan (it already contains `SPEC-006`, `SPEC-017`, `KNOW-002`, `KNOW-001` per data-model.md), or what further exploration was genuinely needed. Depends on T015, T017 (Foundational).

**Checkpoint**: User Story 2 is independently complete — one concrete, directly-observed Claude Code session is recorded (SC-003).

---

## Phase 5: User Story 3 - Confirm the Same Behavior With a Second, Independent Agent (Priority: P3)

**Goal**: Repeat User Story 2's own observation through Antigravity, to confirm the behavior isn't Claude-specific.

**Independent Test**: Same protocol as User Story 2, run through Antigravity instead.

> **⚠️ Requires the user's own hands-on action** — this assistant has no access to Antigravity (spec.md's own explicit constraint). This phase cannot be completed inside this session.

- [X] T025 [US3] **Deferred by explicit user decision**: "podemos ignorar os testes no antigravity por ora, se no claude funcionou" — User Story 2's own Claude Code observation already confirmed the integration works end-to-end, and both agents run the byte-identical Context Engine/Skill content this decision concerns; a second live observation was judged to add no marginal evidence worth the extra session. Not attempted, not blocked — explicitly out of scope for this pass. `fixture/.agents/skills/` remains installed (T016) if revisited later.
- [X] T026 [US3] `report.md`'s own "User Story 3 — Antigravity" section records the deferral above (with its own reasoning) rather than a completed observation.

**Checkpoint**: User Story 3 is closed by explicit user decision, not left open.

---

## Phase 6: User Story 4 - Decide Whether Ranking Needs Tuning, Based Only on Observed Evidence (Priority: P4)

**Goal**: Review every Finding from User Story 1-3 and record one evidence-gated Tuning Decision.

**Independent Test**: Every Finding is classified ranking-attributable or not; the Decision cites specific Findings or states plainly that none warrant a change.

- [X] T027 [US4] Draft a preliminary Tuning Decision in `report.md` using only User Story 1 and User Story 2's own available findings (T023, T024) — classify each as ranking-attributable or not (FR-007). Depends on T023, T024.
- [X] T028 [US4] Finalized the Tuning Decision in `report.md`: **no change warranted**, based on User Story 1 and User Story 2's own findings (F1-F6) alone — User Story 3 (T025/T026) contributes no new evidence by explicit user decision, not by omission, so the preliminary decision (T027) stands as final without modification. Depends on T026, T027.

**Checkpoint**: This feature's own governing rule (§30 Phase 9: "tune ranking only from observed failures") is satisfied by construction — no ranking file is touched anywhere in this task list. Final outcome: no follow-up ranking task is warranted.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency check — no new capability.

- [X] T029 Verify `go test ./...` across the entire module (001 through 018) is still green — this feature touches no source file, so this is a pure confirmation, not a fix cycle.
- [X] T030 Re-read `report.md` end to end: confirmed every Finding (F1-F6) cites concrete evidence, the Tuning Decision is unambiguous ("no change warranted," reasoning stated), and every fixture entity's real-content provenance (data-model.md's own source column) is traceable — no inconsistency found.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — the fixture's own content comes from already-published `specs/006/011-018` files.
- **Foundational (Phase 2)**: Depends on Setup (needs the fixture project to install into). Blocks User Story 2/3 only.
- **User Story 1 (Phase 3)**: Depends on Setup only — proceeds in parallel with Foundational.
- **User Story 2 (Phase 4)**: Depends on Foundational.
- **User Story 3 (Phase 5)**: Depends on Foundational; blocked on user action (T025).
- **User Story 4 (Phase 6)**: Depends on User Story 1 and 2 for its preliminary draft (T027); depends on User Story 3 for its final form (T028).
- **Polish (Phase 7)**: Depends on User Story 4 completing.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on any other story.
- **User Story 2 (P2)**: Depends on Foundational only.
- **User Story 3 (P3)**: Depends on Foundational only; independent of User Story 1/2's own content, but requires the user.
- **User Story 4 (P4)**: The only story genuinely dependent on all three others' own output.

### Parallel Opportunities

- T001-T014 (all of Setup) can be authored in parallel — 14 independent files.
- T015/T016 (Foundational installs) can proceed in parallel.
- T019-T022 (User Story 1's four analysis groups) can proceed in parallel once T018 exists.
- User Story 1 (Phase 3) as a whole can proceed in parallel with Foundational (Phase 2) and, once Foundational completes, with User Story 2 (Phase 4).

---

## Parallel Example: Setup

```bash
# All 9 Spec files and 5 scaffold files are independent:
Task: "Create SPEC-011 in fixture/ai/programs/PRG-001/features/FEAT-001/specs/SPEC-011/spec.md"
Task: "Create SPEC-012 in fixture/ai/programs/PRG-001/features/FEAT-001/specs/SPEC-012/spec.md"
Task: "Create KNOW-001 and KNOW-002 in fixture/ai/knowledge/"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (the fixture).
2. Complete Phase 3: User Story 1.
3. **STOP and VALIDATE**: Every fixture Spec has a recorded pass/fail judgment against its own real dependency graph — the one piece of evidence achievable without any live agent session at all.

### Incremental Delivery

1. Setup → the fixture exists, reviewable on its own.
2. User Story 1 → static retrieval evidence recorded (MVP).
3. Foundational + User Story 2 → one real Claude Code observation recorded.
4. User Story 3 → hand off T025 to the user; resume once they report back.
5. User Story 4 → preliminary decision after step 3, finalized after step 4.
6. Polish (Phase 7).

### Team Strategy

User Story 1 and User Story 2 can both be completed by this assistant
directly, in either order, once their own prerequisite phase is done.
User Story 3 is the one genuine handoff point in this feature — every
other task can proceed without waiting on it, except User Story 4's
own final form (T028).

---

## Notes

- No task in this feature modifies any file under `internal/`, `cmd/`, or `kit/` — everything lives under `specs/019-dogfooding-evaluation/` (research.md).
- Every fixture entity's content must trace to a real, already-published source file (data-model.md's own "source" column) — never invented data (research.md #1).
- T025/T026 are the only tasks this session cannot complete alone — they require the user's own Antigravity access, per spec.md's own explicit constraint (FR-006).
- A ranking/budgeting code change is explicitly out of this task list's own scope unless T028 identifies one — at that point it becomes its own new, evidence-cited task, not retrofitted into this list.
- Commit after each phase; stop at the User Story 3 handoff point to wait for the user.
