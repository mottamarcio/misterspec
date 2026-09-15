---

description: "Task list template for feature implementation"
---

# Tasks: Agent Adapter Layer

**Input**: Design documents from `/specs/006-agent-adapter/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/agents.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 005-embedded-kit's full test suite is an explicit, named regression gate for `internal/installer`'s generalization — not just "the usual suite."

**Organization**: Tasks are grouped by user story. This feature's shape is a sixth distinct one across six features so far, and is called out explicitly rather than forced into a uniform pattern: **User Story 1 (Registry) and User Story 2 (the Claude adapter) are independent of each other** — US1's tests use a fake `Adapter` defined inline, never the real `claude` package, so Registry can be built and proven without Claude's `Install` being ready (the same "independent lanes" shape 003-entity-creation's US1/US2 had). **User Story 3 depends on User Story 2** — per spec.md's own Independent Test wording, detecting what's installed is tested by first actually installing something (US2), then reading it back — there is no meaningful "detect what's installed" test without a real install already having happened.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/installer/           # existing package — Foundational (extended)
internal/agents/               # NEW package — Foundational (types), US1 (registry), US3 (record read)
internal/agents/claude/         # NEW package — US2
internal/agents/builtin/         # NEW package — US2 (wiring)
internal/example/                 # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the directory skeletons this feature adds.

- [X] T001 Create `internal/agents/`, `internal/agents/claude/`, and `internal/agents/builtin/` directories, each with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: Module builds (`go build ./...`) with the new empty packages before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Generalize `internal/installer` so `internal/agents/claude` has a proven materialization primitive to build on, and define the shared `agents` types every user story needs to even compile against.

**⚠️ CRITICAL**: No user story implementation may begin until this phase — including the `internal/installer` regression check — is complete.

- [X] T002 [P] Unit tests for `InstallFS`/`ListFS` — the same behavior `List`/`Install` already prove (fresh directory, no-overwrite skip, overwrite replace, containment rejection), but against a fixture `fs.FS` and `sourceDir` instead of `kit.TemplatesFS`/`"templates"`, proving the generalization is genuinely source-independent — in `internal/installer/installer_test.go`.
- [X] T003 Implement `InstallFS`/`ListFS` in `internal/installer/installer.go`; make `Install`/`List` thin wrappers over them (`Install(t,o) { return InstallFS(kit.TemplatesFS, "templates", t, o) }`). Depends on T002.
- [X] T004 Run `go test ./internal/installer/...` and confirm 005-embedded-kit's full existing test suite passes unmodified after the generalization.
- [X] T005 [P] Define `Adapter`, `InstallRequest`, `InstallResult` in `internal/agents/adapter.go`.
- [X] T006 [P] Define `InstallRecord` (with `json:` tags matching `docs/architecture-specification.md` §22's `install.json` schema exactly) in `internal/agents/record.go` — types only; `recordInstall`/`CurrentInstall` come later (US2, US3).

**Checkpoint**: Foundation ready — `internal/installer` generalized and regression-checked; `internal/agents`'s shared types compile. `go test ./internal/installer/...` green.

---

## Phase 3: User Story 1 - Discover Available Adapters (Priority: P1) 🎯 MVP

**Goal**: List every registered adapter's ID, name, and target integration location; select one by ID with a distinct not-found result.

**Independent Test**: List adapters and confirm at least Claude Code is present with correct metadata; select by a known ID and by an unknown one, confirming the distinct not-found result — no filesystem writes, no dependency on `claude` package's `Install` being ready.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T008.

- [X] T007 [P] [US1] Unit tests for `Registry` — `List()` returns every registered adapter sorted by `ID()`; `Get()` returns `(adapter, true)` for a registered ID and `(nil, false)` for an unregistered one — using a fake `Adapter` implementation defined directly in the test file (its `Install` need not do anything real), never the `claude` package — in `internal/agents/registry_test.go`.

### Implementation for User Story 1

- [X] T008 [US1] Implement `Registry`, `NewRegistry`, `List`, `Get` in `internal/agents/registry.go`. Depends on T005, T007.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/agents/... -run TestRegistry` passes on its own, with zero dependency on `internal/agents/claude`.

---

## Phase 4: User Story 2 - Install the Kit for a Selected Adapter (Priority: P2)

**Goal**: The Claude Code adapter materializes a fixture set of Skill resources into `.claude/skills` and writes an accurate `install.json` record.

**Independent Test**: Install a fixture Skills `fs.FS` for the Claude adapter into a fresh target directory; confirm every resource lands at `.claude/skills`, byte-identical, and the resulting record names the right adapter and path. Independent of User Story 1's `Registry` — `Adapter.Install` doesn't need a `Registry` to run.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T011–T013.

- [X] T009 [P] [US2] Filesystem-integration tests for the Claude adapter — `ID()`, `Name()`, `TargetPath()` return the expected constants (`"claude-code"`, `.claude/skills`); `Install()` against a fixture Skills `fs.FS` into a fresh target directory materializes every resource, byte-identical, under `.claude/skills`; re-running without overwrite skips everything already present; `install.json` is written at `.misterspec/install.json` matching `InstallRecord`'s schema — in `internal/agents/claude/claude_test.go`.
- [X] T010 [P] [US2] Unit tests for `recordInstall` — writes `install.json` atomically with the correct `schema_version` (`1`), `misterspec_version` (`"0.1.0"`), and `agent.id`/`agent.integration_path` fields — in `internal/agents/record_test.go`.

### Implementation for User Story 2

- [X] T011 [US2] Implement `recordInstall` in `internal/agents/record.go` (writes `InstallRecord` atomically, reusing `internal/installer`'s atomic-write helper rather than a third implementation). Depends on T006, T010.
- [X] T012 [US2] Implement `claude.New()`, `ID()`, `Name()`, `TargetPath()`, and `Install()` (calls `installer.InstallFS(req.Skills, ".", filepath.Join(req.ProjectRoot, TargetPath()), req.Overwrite)`, then `recordInstall`) in `internal/agents/claude/claude.go`. Depends on T003, T011, T009.
- [X] T013 [US2] Implement `builtin.Default()` (`agents.NewRegistry(claude.New())`) in `internal/agents/builtin/builtin.go`. Depends on T012.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/agents/claude/...` passes on its own.

---

## Phase 5: User Story 3 - Detect the Currently Installed Adapter (Priority: P3)

**Goal**: Read back a project's installation record without re-installing or guessing from directory contents.

**Independent Test**: Install the Claude adapter into a fixture project (via User Story 2's real `Install`), then confirm `CurrentInstall` reports the exact adapter and path that was installed; separately, confirm a never-installed project reports "not installed," not an error. **Note**: this story's test genuinely depends on User Story 2's `Install` having actually run — spec.md's own Independent Test is phrased that way, and there is no meaningful "what's installed" answer to detect without a real installation to read back.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T015.

- [X] T014 [US3] Filesystem-integration tests for `CurrentInstall` — after a real `claude.New().Install(...)` call, `CurrentInstall` returns the exact `AgentID`/`IntegrationPath` that was installed, `installed == true`; for a project directory that has never had anything installed, `CurrentInstall` returns `installed == false` and `err == nil` — in `internal/agents/record_test.go` (same file as T010; sequential, not `[P]`, since T010/T011 are already complete by the time this runs). Depends on T012.

### Implementation for User Story 3

- [X] T015 [US3] Implement `CurrentInstall` in `internal/agents/record.go` (reads `.misterspec/install.json`; absent file → `(zero, false, nil)`; malformed file → `(zero, false, err)`). Depends on T011, T014.

**Checkpoint**: All three user stories pass their own tests — the full discover → install → detect cycle works end to end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T016 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 005-embedded-kit's packages included) and fix any findings.
- [X] T017 [P] Verify/extend package-level doc comments on `internal/agents`, `internal/agents/claude`, and `internal/agents/builtin`, cross-checked against `specs/006-agent-adapter/contracts/agents.md`.
- [X] T018 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: `builtin.Default().List()`, `Get("claude-code")`, `Install` with a fixture Skills `fs.FS`, then `CurrentInstall` confirming the round trip.
- [X] T019 Reconcile `specs/006-agent-adapter/contracts/agents.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T020 Full regression run: `go test ./...` across the entire module (001 through 006) green, `go vet ./...` clean, `gofmt -l .` empty, `go test ./internal/operations/... -race` clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories, and includes a hard regression gate (T004) for `internal/installer`'s generalization.
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on Foundational only — not on User Story 1.
- **User Story 3 (Phase 5)**: Depends on User Story 2's implementation (T012) — it reads back a real install, not a fixture written independently.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Can start after Foundational, in parallel with User Story 1 — `Adapter.Install` needs no `Registry`.
- **User Story 3 (P3)**: Depends on User Story 2 (T012) — genuinely, not just for test-writing convenience.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational's `internal/installer` generalization verified non-regressive, and `agents`'s shared types compiling, before any story's implementation.
- User Story 3 cannot start in earnest until User Story 2's `Install` is real.

### Parallel Opportunities

- T002 (installer generalization tests) and T005/T006 (agents' shared types) in parallel — different packages.
- Once Foundational is done: **User Story 1 and User Story 2 can proceed fully in parallel** by different contributors, the same independence 003-entity-creation's US1/US2 had.
- Within User Story 2: T009 and T010 (tests, different files) in parallel.
- Within Polish: T016 and T017 in parallel.

---

## Parallel Example: User Story 1 and User Story 2 together

```bash
# Once Foundational is done, these two stories need no coordination:
Task: "Unit tests for Registry in internal/agents/registry_test.go"
Task: "Filesystem-integration tests for the Claude adapter in internal/agents/claude/claude_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (`internal/installer` generalized and regression-checked; `agents` types defined).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/agents/... -run TestRegistry` green, independently.
5. This alone already proves adapter discovery works — the prerequisite for a future `misterspec init`'s agent-selection step, even before any adapter can actually install anything.

### Incremental Delivery

1. Setup + Foundational (installer generalized, regression-checked; shared types ready).
2. Add US1 → validate independently → discovery usable (MVP).
3. Add US2 (independent of US1, can be parallel) → validate independently → Claude installation usable.
4. Add US3 (depends on US2) → validate independently → detecting what's installed usable.
5. Polish (Phase 6), including the full-module `-race` regression run (T020).

### Team Strategy

Once Foundational is done, Developer A can take US1 and Developer B can
take US2 fully in parallel — the same independence 003-entity-creation's
US1/US2 had. US3 must wait specifically on US2 (not US1), so it's the
natural next pickup for whichever developer finishes US2 first.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- An adapter writes only to its own `TargetPath()` and `.misterspec/install.json` — never a project's `ai/` artifact tree (plan.md's Constitution Check, Principle VII).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
