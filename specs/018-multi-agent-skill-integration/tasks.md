---

description: "Task list template for feature implementation"
---

# Tasks: Multi-Agent Skill Integration

**Input**: Design documents from `/specs/018-multi-agent-skill-integration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/adapters-and-skills.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). For User Story 1 (new adapters), each new package gets its own test file mirroring `internal/agents/claude/claude_test.go`'s own coverage, written first. For User Stories 2/3 (Skill content edits), the "test" is `internal/example/skills_content_test.go`'s own already-existing structural-conformance machinery (009-canonical-skills-content) — there is no new test file; each Skill's own edit is paired with the exact allowlist entry that makes the new `internal context` reference conformant, and the full existing suite is the regression gate proving nothing else broke. 001-017's own full suites are named regression gates throughout.

**Organization**: Tasks are grouped by user story. **User Story 1** (five new adapters) is entirely independent of **User Story 2**/**User Story 3** (Skill content edits) — different files, different packages, no shared code path — so all three stories can proceed in parallel once Foundational completes. Within User Story 1, the five adapters are mutually independent of each other and of the shared `builtin.go` registration (which depends on all five existing). Within User Story 3, the three Skills (`create-plan`, `create-tasks`, `analyze`) are mutually independent edits to three different files.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US3)
- Every task names its exact file path

## Path Conventions

```text
internal/agents/{agy,codex,copilot,cursoragent,devin}/   # new packages (US1)
internal/agents/builtin/                                    # existing package, extended (US1)
internal/example/skills_content_test.go                        # existing test, extended (US2, US3)
kit/skills/{implement,create-plan,create-tasks,analyze}/SKILL.md  # existing files, edited (US2, US3)
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package dependency and no new subsystem (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Register `"context"` as a recognized operation in `internal/example/skills_content_test.go`'s own `knownInternalCommands` allowlist — the one shared prerequisite User Story 2 and User Story 3 both need before any of their four Skill edits can reference `internal context` without the existing structural-conformance test rejecting it as an unrecognized command (FR-003 of 009-canonical-skills-content, still enforced). User Story 1 has no dependency on this phase at all — it touches an entirely different file.

**⚠️ CRITICAL**: No User Story 2/3 Skill edit may reference `internal context` in its own `Deterministic Operations` section until this phase completes. User Story 1 may proceed immediately, in parallel with this phase.

- [ ] T001 Add `"context": true` to `knownInternalCommands` in `internal/example/skills_content_test.go` (alongside the existing `resolve`, `inspect`, ..., `status` entries) — the real, registered `misterspec internal context` command (017-internal-context-command) this feature's four Skill edits will reference.
- [ ] T002 Regression checkpoint: `go test ./internal/example/...` still green (the allowlist addition alone changes no test outcome yet, since no Skill file references `internal context` until User Story 2/3's own tasks run) and `go build ./...` succeeds. Depends on T001.

**Checkpoint**: Foundation ready — `internal context` is a recognized operation for the structural-conformance check. User Story 2 and User Story 3 implementation may begin. User Story 1 was never blocked and may already be underway.

---

## Phase 3: User Story 1 - Install misterspec's Skills for More Coding Agents (Priority: P1) 🎯 MVP

**Goal**: Five new `agents.Adapter` implementations (`agy`, `codex`, `copilot`, `cursor-agent`, `devin`), each installing misterspec's canonical Skills byte-for-byte into that agent's own verified Skills directory (research.md #1), registered in `builtin.Default()` alongside the existing `claude-code` adapter.

**Independent Test**: For each of the five agents, select it by ID from `builtin.Default()`, install into a fixture project directory, and confirm every canonical Skill lands at `<root>/<TargetPath()>/<skill-name>/SKILL.md`, byte-identical to the source — entirely independent of User Story 2/3's own Skill-content changes.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail (undefined package) before implementing T004-T008/T010.

- [X] T003 [P] [US1] Tests for the `agy` adapter in new file `internal/agents/agy/agy_test.go`, mirroring `internal/agents/claude/claude_test.go`'s own three tests exactly (metadata: `ID()=="agy"`, non-empty `Name()`, `TargetPath()==".agents/skills"`; fresh-directory install producing byte-identical files plus a valid `install.json` record; a second install with `Overwrite: false` skips already-present files).
- [X] T004 [P] [US1] Tests for the `codex` adapter in new file `internal/agents/codex/codex_test.go`, same three-test shape (`ID()=="codex"`, `TargetPath()==".agents/skills"`, install/no-clobber).
- [X] T005 [P] [US1] Tests for the `copilot` adapter in new file `internal/agents/copilot/copilot_test.go`, same three-test shape (`ID()=="copilot"`, `TargetPath()==".github/skills"`, install/no-clobber).
- [X] T006 [P] [US1] Tests for the `cursoragent` package in new file `internal/agents/cursoragent/cursoragent_test.go`, same three-test shape (`ID()=="cursor-agent"`, `TargetPath()==".cursor/skills"`, install/no-clobber).
- [X] T007 [P] [US1] Tests for the `devin` adapter in new file `internal/agents/devin/devin_test.go`, same three-test shape (`ID()=="devin"`, `TargetPath()==".devin/skills"`, install/no-clobber).

### Implementation for User Story 1

- [X] T008 [P] [US1] Implement the `agy` adapter in new file `internal/agents/agy/agy.go`: `id="agy"`, `name="Antigravity"`, `targetPath=".agents/skills"`, `Install` identical in body to `claude.Adapter.Install` (data-model.md) — `installer.InstallFS` + `agents.RecordInstall`, no new logic. Depends on T003.
- [X] T009 [P] [US1] Implement the `codex` adapter in new file `internal/agents/codex/codex.go`: `id="codex"`, `name="Codex CLI"`, `targetPath=".agents/skills"`, same `Install` body. Depends on T004.
- [X] T010 [P] [US1] Implement the `copilot` adapter in new file `internal/agents/copilot/copilot.go`: `id="copilot"`, `name="GitHub Copilot"`, `targetPath=".github/skills"`, same `Install` body. Depends on T005.
- [X] T011 [P] [US1] Implement the `cursoragent` adapter in new file `internal/agents/cursoragent/cursoragent.go`: `id="cursor-agent"`, `name="Cursor"`, `targetPath=".cursor/skills"`, same `Install` body. Depends on T006.
- [X] T012 [P] [US1] Implement the `devin` adapter in new file `internal/agents/devin/devin.go`: `id="devin"`, `name="Devin for Terminal"`, `targetPath=".devin/skills"`, same `Install` body. Depends on T007.
- [X] T013 [US1] Extend `internal/agents/builtin/builtin_test.go`: assert `Default()` now returns all six adapters by ID (`claude-code`, `agy`, `codex`, `copilot`, `cursor-agent`, `devin`), each with its own documented `TargetPath()`; add a coexistence test installing both `agy` and `codex` into the same fixture project (they share `.agents/skills`). **Corrected during implementation**: `install.json` (§22's own schema, unmodified) records exactly one agent, overwritten per call — it does not track multiple installs, so the coexistence assertion is on the actual installed *Skill files* surviving byte-identical regardless of which adapter wrote them first (`Skipped` on the second call is correct, harmless behavior, not corruption), not on `install.json` somehow listing both (FR-004 is about file integrity, not `install.json`'s own single-record design). Depends on T008-T012.
- [X] T014 [US1] Register the five new adapters in `Default()` in `internal/agents/builtin/builtin.go` (`claude.New()`, `agy.New()`, `codex.New()`, `copilot.New()`, `cursoragent.New()`, `devin.New()`). Depends on T013.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/agents/...` passes on its own, entirely independent of User Story 2/3.

---

## Phase 4: User Story 2 - Implementation Begins From the Context Pack (Priority: P2)

**Goal**: `/implement`'s own instructions request a Context Pack (`internal context <Spec-ID> --intent implementation`) immediately after its existing `resolve`/`inspect` steps, begin from its returned items, fall back to prior behavior on failure, and explicitly state the agent remains free to explore further.

**Independent Test**: Inspect `kit/skills/implement/SKILL.md`'s own content directly — confirm the new required operation, the new early Procedure step, the fallback statement, and the exploration-remains-allowed statement are all present, and that the file still passes `internal/example/skills_content_test.go`'s own structural-conformance check.

### Tests for User Story 2

> This "test" is an existing golden-fixture check (009-canonical-skills-content), not a new test file — the allowlist entry below is written first, immediately followed by the SKILL.md edit it exists to permit.

- [X] T015 [US2] Add `"context"` to `implement`'s own entry in `skillOperationsAllowlist` in `internal/example/skills_content_test.go` (alongside its existing `resolve`, `inspect`, `validate`). Depends on T002 (Foundational).

### Implementation for User Story 2

- [X] T016 [US2] Edit `kit/skills/implement/SKILL.md` (data-model.md): add one bullet to `Deterministic Operations`'s "Required operations" — `` `internal context SPEC-### --intent implementation` — request a budgeted Context Pack before broader exploration``; insert one new numbered step in `Procedure` immediately after the existing `resolve`/`inspect` step, before Task selection — `` Run `internal context SPEC-### --intent implementation` and begin from its returned items. If the request fails, proceed using this Skill's own Required Context below instead.``; add one sentence to `Required Context` or `Optional Context` stating the agent remains free to read further repository files or run further deterministic operations beyond the Context Pack (FR-007). No other section changes (FR-009, FR-010). Depends on T015.
- [X] T017 [US2] Verify: `go test ./internal/example/... -run TestSkillsContent` passes with `implement` now referencing `internal context`, confirming FR-005/FR-007/FR-008 are all textually present and the file remains structurally conformant. Depends on T016.

**Checkpoint**: User Story 2 is independently complete and testable — the implementation Skill's own instructions request a Context Pack first, remain free to explore further, and fall back gracefully — verified without touching User Story 1 or User Story 3's own files.

---

## Phase 5: User Story 3 - Planning, Task Creation, and Analysis Begin From the Context Pack (Priority: P3)

**Goal**: `/create-plan`, `/create-tasks`, and `/analyze` each gain the same treatment as User Story 2, with their own matching intent (`planning`, `tasks`, `validation` respectively — research.md #5).

**Independent Test**: For each of the three Skills, inspect its own `SKILL.md` content directly — confirm the same four additions User Story 2 made (required operation, early Procedure step, fallback statement, exploration-remains-allowed statement), each using that Skill's own correct intent, and confirm the file still passes the structural-conformance check.

### Tests for User Story 3

> Same pattern as User Story 2 (T015): the allowlist entry is written first, immediately followed by each Skill's own SKILL.md edit.

- [X] T018 [P] [US3] Add `"context"` to `create-plan`'s own entry in `skillOperationsAllowlist` in `internal/example/skills_content_test.go`. Depends on T002 (Foundational).
- [X] T019 [P] [US3] Add `"context"` to `create-tasks`'s own entry in `skillOperationsAllowlist` in `internal/example/skills_content_test.go`. Depends on T002 (Foundational).
- [X] T020 [P] [US3] Add `"context"` to `analyze`'s own entry in `skillOperationsAllowlist` in `internal/example/skills_content_test.go`. Depends on T002 (Foundational).

### Implementation for User Story 3

- [X] T021 [P] [US3] Edit `kit/skills/create-plan/SKILL.md` (data-model.md): same four additions as T016, using `` `internal context SPEC-### --intent planning` `` and inserting the new Procedure step immediately after the existing `resolve`/`inspect` step, before strategy selection. Depends on T018.
- [X] T022 [P] [US3] Edit `kit/skills/create-tasks/SKILL.md` (data-model.md): same four additions, using `` `internal context SPEC-### --intent tasks` ``, inserted after the existing `resolve`/`inspect` step, before reading the Plan's own decomposition detail. Depends on T019.
- [X] T023 [P] [US3] Edit `kit/skills/analyze/SKILL.md` (data-model.md): same four additions, using `` `internal context SPEC-### --intent validation` ``, inserted after the existing `resolve`/`inspect` step, before per-requirement judgment. Depends on T020.
- [X] T024 [US3] Verify: `go test ./internal/example/... -run TestSkillsContent` passes with all three Skills now referencing `internal context`, each with its own correct intent, confirming FR-006/FR-007/FR-008 are textually present in all three and every file remains structurally conformant. Depends on T021, T022, T023.

**Checkpoint**: User Story 3 is independently complete and testable — planning, task creation, and analysis all request an intent-appropriate Context Pack first, remain free to explore further, and fall back gracefully.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T025 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 018-multi-agent-skill-integration's packages included) and fix any findings. Clean on first run.
- [X] T026 [P] Verify/extend `internal/agents`'s own package-level doc comment to mention the five new adapters, cross-checked against `specs/018-multi-agent-skill-integration/contracts/adapters-and-skills.md`.
- [X] T027 Add a compiled, run-in-CI example in `internal/example` (new file `multi_agent_skill_integration_quickstart_test.go`, extending the existing package) exercising `quickstart.md`'s flow end-to-end: listing all six registered adapters and their `TargetPath()`s; installing for `cursor-agent`; installing both `agy` and `codex` into one project; asserting the four updated `SKILL.md` files' own content contains `internal context` with their respective correct intents, the fallback statement, and the exploration-remains-allowed statement — each matching `quickstart.md`'s own documented result. Also discovered and fixed one now-outdated assertion in the pre-existing `internal/example/agents_quickstart_test.go` (006-agent-adapter), which asserted `registry.Get("codex")` returned `false` — legitimately no longer true now that `codex` is a real, registered adapter; replaced with a truly nonexistent ID (`not-a-real-agent`) to keep proving the same absence guarantee.
- [X] T028 Reconciled `specs/018-multi-agent-skill-integration/contracts/adapters-and-skills.md`'s adapter table and Skill/intent table against the actual implementation — verified every `id`/`name`/`targetPath` constant across all five new adapter files matches the contract exactly; zero drift found.
- [X] T029 Full regression run: `go test ./...` across the entire module (001 through 018) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/agents/... -race` clean. All confirmed.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS User Story 2 and User Story 3 only (the `knownInternalCommands` entry their Skill edits need). Does **not** block User Story 1.
- **User Story 1 (Phase 3)**: No dependency on Foundational or on User Story 2/3 — may start immediately, in parallel with Foundational.
- **User Story 2 (Phase 4)**: Depends on Foundational (T002). Independent of User Story 1 and User Story 3.
- **User Story 3 (Phase 5)**: Depends on Foundational (T002). Independent of User Story 1 and User Story 2.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on any other story — the true, unblocked entry point, and this feature's own MVP.
- **User Story 2 (P2)**: Depends only on Foundational; independent of User Story 1 and User Story 3.
- **User Story 3 (P3)**: Depends only on Foundational; independent of User Story 1 and User Story 2.

Unlike 016's or 017's own mostly-sequential chains, this feature's three
stories are **fully parallel** once Foundational completes — they share
no code, no file, and no runtime dependency, matching spec.md's own
framing of "two things the user explicitly bundled together" rather
than one linear pipeline.

### Within Each User Story

- User Story 1: tests written and failing (undefined package) before each adapter's own implementation (Constitution Principle V); the five adapters and their tests are mutually independent; `builtin.go`'s own registration (T014) depends on all five existing.
- User Story 2/3: the allowlist entry (this feature's own "test-first" step, since there is no separate test file) is written before the corresponding SKILL.md edit it exists to permit.

### Parallel Opportunities

- T003-T007 (User Story 1's five test files) can all be written in parallel — five different files, no dependency between them.
- T008-T012 (User Story 1's five adapter implementations) can all proceed in parallel once their own respective test exists.
- User Story 1 (Phase 3) as a whole can proceed in parallel with Foundational (Phase 2) — they touch entirely different files.
- User Story 2 (Phase 4) and User Story 3 (Phase 5) can proceed in parallel with each other once Foundational completes — different Skill files, no shared state beyond the one allowlist map, whose four entries (T015, T018, T019, T020) are themselves independent additions.
- T018-T020 (User Story 3's three allowlist entries) and T021-T023 (its three SKILL.md edits) can each run in parallel internally.
- Within Polish: T025 and T026 in parallel.

---

## Parallel Example: User Story 1 (all five adapters at once)

```bash
# Five independent test files, then five independent implementations:
Task: "Tests for the agy adapter in internal/agents/agy/agy_test.go"
Task: "Tests for the codex adapter in internal/agents/codex/codex_test.go"
Task: "Tests for the copilot adapter in internal/agents/copilot/copilot_test.go"
Task: "Tests for the cursoragent package in internal/agents/cursoragent/cursoragent_test.go"
Task: "Tests for the devin adapter in internal/agents/devin/devin_test.go"
```

## Parallel Example: User Story 2 + User Story 3 together

```bash
# Different Skill files — no shared state beyond four independent map entries:
Task: "Edit kit/skills/implement/SKILL.md"
Task: "Edit kit/skills/create-plan/SKILL.md"
Task: "Edit kit/skills/create-tasks/SKILL.md"
Task: "Edit kit/skills/analyze/SKILL.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 3: User Story 1 (does not require Foundational at all).
3. **STOP and VALIDATE**: `go test ./internal/agents/...` green, independently. All six agents can now install misterspec's canonical Skills — the literal prerequisite this feature exists to deliver first.

### Incremental Delivery

1. Setup (nothing) — Foundational (one allowlist entry) can run in parallel with User Story 1.
2. User Story 1 lands → five new agents can install Skills (MVP!).
3. User Story 2 lands (once Foundational is done) → `/implement` begins from a Context Pack.
4. User Story 3 lands (once Foundational is done, independent of User Story 2) → planning, task creation, and analysis do the same.
5. Polish (Phase 6), including the full-module `-race` regression run (T029).

### Team Strategy

All three stories are mutually independent once Foundational's one-line
addition exists — three developers could take User Story 1, User Story
2, and User Story 3 fully in parallel with zero coordination beyond
that single shared line.

---

## Notes

- [P] tasks touch different files, or the same file with independent, non-overlapping map entries.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — User Story 1's five adapters each get a real test file written first; User Story 2/3's Skill edits use the existing structural-conformance test as their own regression gate, with the allowlist entry as the enabling "test-first" step.
- No adapter may alter Skill content or meaning (FR-003) — every one of the five new `Install` methods is a byte-for-byte copy of `claude.Adapter.Install`'s own already-approved body.
- `agy` and `codex` share the identical `TargetPath()` (`.agents/skills`) — a verified fact (research.md #1), not an error; T013 explicitly proves their coexistence is safe.
- Every updated Skill (`implement`, `create-plan`, `create-tasks`, `analyze`) must state the agent remains free to explore beyond the Context Pack (FR-007) and must fall back to prior behavior on a failed request (FR-008) — T017/T024 are where this is verified.
- No other canonical Skill, and no part of the Context Engine (011-017), is touched anywhere in this feature (research.md).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
