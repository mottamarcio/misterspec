---

description: "Task list template for feature implementation"
---

# Tasks: Interactive Init TUI

**Input**: Design documents from `/specs/010-interactive-init-tui/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tui.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). `Model.Update`/`View` are tested directly with synthetic `tea.Msg` values, no real terminal needed (research.md). 007-project-bootstrap's and 008-cli-cobra's full test suites are explicit, named regression gates — the non-interactive `init` path (SC-005) must not change at all.

**Organization**: Tasks are grouped by user story. **User Story 1 (happy-path bootstrap) depends only on Foundational.** **User Story 2 (non-empty/already-initialized warning) depends only on Foundational** for its own mechanics (the Warning screen displaying and transitioning correctly) — but its own Acceptance Scenario 4 ("continue anyway" reaching a clear explanation) is only fully demonstrable once User Story 3's Error screen exists, so that specific end-to-end chain is a User Story 3 task, honestly reflecting the real dependency rather than a forced independence. **User Story 3 (clear failure recovery) depends only on Foundational** for its own Error-screen mechanics, and additionally on User Story 2's Warning screen for the one integration test that ties both together.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/tui/                  # NEW package — Foundational, US1, US2, US3
internal/bootstrap/             # existing package — Foundational (extended)
internal/cli/                    # existing package — US1 (extended)
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add this feature's new dependencies and create the directory skeleton it adds.

- [X] T001 Add `github.com/charmbracelet/bubbletea` v1.3.10, `github.com/charmbracelet/lipgloss` v1.1.0, and `golang.org/x/term` (latest) as direct dependencies (`go get`; updates `go.mod`/`go.sum`).
- [X] T002 [P] Create `internal/tui/` directory with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: `go build ./...` succeeds with the new empty package and the new dependencies resolved before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Give every screen a shared `Model`/`Screen` type to extend, and the one additive `bootstrap` capability (`Empty`) the Warning screen needs — nothing user-story-specific yet.

**⚠️ CRITICAL**: No user story implementation may begin until this phase is complete.

- [X] T003 [P] Unit tests for `InspectResult.Empty` in `internal/bootstrap/inspect_test.go`: an empty (or non-existent) uninitialized directory reports `Empty == true`; a non-empty uninitialized directory reports `Empty == false`; an already-initialized directory reports `Empty == false` regardless of its own contents (research.md).
- [X] T004 Implement `InspectResult.Empty` and its population in `Inspect`, `internal/bootstrap/inspect.go`. Existing fields and `Inspect`'s exported signature unchanged. Depends on T003.
- [X] T005 Define `Screen` (the seven-state enum, §37 verbatim), `Model` (unexported fields per data-model.md), and `NewModel` in `internal/tui/model.go`; implement `Model.Init()` returning a `tea.Cmd` that runs `bootstrap.Inspect` and produces an `inspectResultMsg`.
- [X] T006 [P] Unit test for `Model.Init()`/`inspectResultMsg` handling in `internal/tui/model_test.go` — confirms the command `Init()` returns, when executed, produces an `inspectResultMsg` wrapping `bootstrap.Inspect`'s real result against a fixture directory. Depends on T005.

**Checkpoint**: Foundation ready — `go build ./...` succeeds; `Empty` and `Model.Init()` tests green. User story implementation can begin.

---

## Phase 3: User Story 1 - Bootstrap a Project Interactively, Start to Finish (Priority: P1) 🎯 MVP

**Goal**: `misterspec init` with no flags, in an interactive terminal, against a fresh empty directory, completes the full Inspect → Agent Selection → Preview → Confirm → Install → Success flow, bootstrapping a real project — and `--agent` still runs 008-cli-cobra's unchanged non-interactive path.

**Independent Test**: Run `misterspec init` with no flags in a fresh, empty, writable directory; select the one available agent; review the preview; confirm; observe a success screen — using only `bootstrap.Bootstrap` underneath.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T012–T017.

- [X] T007 [P] [US1] `Update` tests for `ScreenAgentSelection` in `internal/tui/update_test.go`: cursor movement between registered agents; Enter on a selection builds the preview and transitions to `ScreenPreview`; Esc/Ctrl+C sets `quitting`.
- [X] T008 [P] [US1] `Update` tests for `ScreenPreview` in `internal/tui/update_test.go`: confirm (`y`/Enter) transitions to `ScreenInstalling` and returns a `tea.Cmd` that will call `bootstrap.Bootstrap`; decline (`n`/Esc) sets `quitting` with nothing written.
- [X] T009 [P] [US1] `Update` tests for the `ScreenInstalling` → `ScreenSuccess` transition in `internal/tui/update_test.go`: a `bootstrapResultMsg` carrying a successful `BootstrapOutcome` populates `Model.outcome` and transitions to `ScreenSuccess`.
- [X] T010 [P] [US1] `View` tests for `ScreenAgentSelection`, `ScreenPreview`, `ScreenInstalling`, and `ScreenSuccess` in `internal/tui/view_test.go` — each renders non-empty output containing the expected content (agent name(s), preview resource counts and paths, a success summary naming what was installed).
- [X] T011 [P] [US1] Tests for `internal/cli/init.go`'s new branch in `internal/cli/init_test.go`, using package-level overridable `isInteractiveTerminal`/`tuiRunInit` function variables (so no real TTY or Bubble Tea program is needed): (a) `isInteractiveTerminal` stubbed `true`, no `--agent` → `tuiRunInit` is invoked with the right `dir`, exit `0`; (b) `isInteractiveTerminal` stubbed `false`, no `--agent` → `invalid_argument`, exit `2` (research.md's regression-safety case — same outward shape as before this feature); (c) `--agent` provided → `tuiRunInit` is never invoked, existing 008-cli-cobra behavior unchanged (re-run alongside `TestInitCmd_Success`/`TestInitCmd_AlreadyInitialized`/`TestInitCmd_UnknownAgent` unmodified as a regression check).

### Implementation for User Story 1

- [X] T012 [US1] Implement `ScreenAgentSelection`'s `Update` logic in `internal/tui/update.go`. Depends on T007.
- [X] T013 [US1] Implement `ScreenPreview`'s `Update` logic and `previewContent` assembly (`installer.List`, `installer.ListFS(skills, ".", "skill")`, `selectedAgent.TargetPath()` — research.md's read-only composition) in `internal/tui/update.go`. Depends on T008.
- [X] T014 [US1] Implement the `ScreenInstalling` → `ScreenSuccess` transition (the `bootstrap.Bootstrap` `tea.Cmd` and its result handling) in `internal/tui/update.go`. Depends on T009.
- [X] T015 [US1] Implement `View` rendering and `internal/tui/styles.go`'s `lipgloss` styles for `ScreenAgentSelection`, `ScreenPreview`, `ScreenInstalling`, `ScreenSuccess` in `internal/tui/view.go`. Depends on T010.
- [X] T016 [US1] Implement `RunInit` in `internal/tui/model.go` — builds and runs the `tea.Program`, accepting variadic `tea.ProgramOption`s so a caller (tests, T018) can redirect I/O away from a real terminal. Depends on T012, T013, T014, T015.
- [X] T017 [US1] Implement `isInteractiveTerminal` (wrapping `term.IsTerminal`) and `tuiRunInit` (`= tui.RunInit`) as overridable package-level variables, and the new `RunE` branch (`--agent` set → unchanged; unset + interactive → `tuiRunInit`; unset + non-interactive → `internalcmd.WriteError(...ErrInvalidArgument..."interactive terminal"...)`) in `internal/cli/init.go`. Depends on T011, T016.
- [X] T018 [US1] Full happy-path integration test in `internal/tui/runinit_test.go`: drive `RunInit` via `tea.WithInput`/`tea.WithOutput` against in-memory buffers with a scripted key sequence (select the one fixture agent, confirm the preview) and a fixture `*agents.Registry`/`fs.FS`, confirming the target directory is actually bootstrapped afterward (`project.Detect` succeeds) — quickstart.md's validation strategy. Depends on T017.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/tui/... ./internal/cli/...` passes on its own; the full happy path works end to end via the real binary.

---

## Phase 4: User Story 2 - Warn Before Touching an Already-Initialized or Non-Empty Directory (Priority: P2)

**Goal**: The Warning screen displays and transitions correctly for both the "already initialized" and "non-empty" cases, before any further prompt.

**Independent Test**: Run against a directory that is already an initialized misterspec project, and separately against a non-empty non-project directory; confirm each shows the correct warning before any agent-selection prompt, and declining leaves the directory untouched.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T021–T022.

- [X] T019 [P] [US2] `Update` tests for the `ScreenInspect` → `ScreenNonEmptyWarning` transition (both the already-initialized case and the non-empty-but-uninitialized case) and the Warning screen's own Cancel (→ `quitting`) / Continue (→ `ScreenAgentSelection`) choices, in `internal/tui/update_test.go`.
- [X] T020 [P] [US2] `View` tests for `ScreenNonEmptyWarning` in `internal/tui/view_test.go` — renders the already-initialized case naming the installed agent, and the non-empty case, as two distinct messages, never a generic one.

### Implementation for User Story 2

- [X] T021 [US2] Implement the `ScreenInspect` → `ScreenNonEmptyWarning` transition and the Warning screen's own `Update` logic in `internal/tui/update.go`. Depends on T019.
- [X] T022 [US2] Implement `ScreenNonEmptyWarning`'s `View` rendering (both distinct messages) in `internal/tui/view.go`. Depends on T020.

**Checkpoint**: User Story 2 is independently complete and testable — the Warning screen itself is fully verified; its "continue anyway into a doomed bootstrap" full chain is completed in User Story 3 (T027), which needs the Error screen to exist.

---

## Phase 5: User Story 3 - Clear Recovery When Installation Fails Partway (Priority: P3)

**Goal**: A failure — either a genuine `Inspect` error, or a `Bootstrap` failure after confirmation — shows a specific Error screen naming the failure and confirming that retrying is safe.

**Independent Test**: Simulate a failure partway through installation (e.g. a target that becomes unwritable after confirmation) and confirm the resulting screen names the specific failure and confirms re-running is safe.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T025–T026.

- [X] T023 [P] [US3] `Update` tests for the `ScreenInstalling` → `ScreenError` transition (a `bootstrapResultMsg` carrying an error) and the `ScreenInspect` → `ScreenError` transition (a genuine `Inspect` error, e.g. `project.ErrInvalidConfiguration`) in `internal/tui/update_test.go`.
- [X] T024 [P] [US3] `View` tests for `ScreenError` in `internal/tui/view_test.go` — renders the specific failure (not a generic message) and an explicit statement that re-running `misterspec init` is safe (FR-008).

### Implementation for User Story 3

- [X] T025 [US3] Implement both `ScreenError`-entry transitions (`Installing`→`Error` on `Bootstrap` failure; `Inspect`→`Error` on a genuine `Inspect` error) in `internal/tui/update.go`. Depends on T023.
- [X] T026 [US3] Implement `ScreenError`'s `View` rendering, including the retry-safety statement, in `internal/tui/view.go`. Depends on T024.
- [X] T027 [US3] Integration test completing spec.md's User Story 2 Acceptance Scenario 4: "continue anyway" past an already-initialized Warning screen proceeds through agent selection and preview, `bootstrap.Bootstrap` rejects with `ErrAlreadyInitialized`, and the resulting `ScreenError` shows that specific, clear explanation — never a silent no-op or a forced overwrite. In `internal/tui/runinit_test.go`. Depends on T021 (US2's Warning transition) and T025/T026.

**Checkpoint**: All three user stories pass their own tests — every acceptance scenario in spec.md is demonstrable end to end, including the one that spans User Story 2 and 3.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T028 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 009-canonical-skills-content's packages included) and fix any findings.
- [X] T029 [P] Verify/extend package-level doc comments on `internal/tui`, cross-checked against `specs/010-interactive-init-tui/contracts/tui.md`.
- [X] T030 Reconcile `specs/010-interactive-init-tui/contracts/tui.md`'s signatures and screen-transition table against the actual implementation; document any drift (same discipline as every prior feature's final reconciliation task).
- [X] T031 Full regression run: `go test ./...` across the entire module (001 through 010) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go mod tidy` clean, `go test ./internal/bootstrap/... ./internal/cli/...` (007's and 008's own suites) specifically re-confirmed green as the named non-interactive-path regression gate (SC-005), `go test ./internal/tui/... -race` clean (the `Bootstrap` call runs as an async `tea.Cmd`, mirroring this project's established `-race` discipline for concurrency-sensitive code).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on Foundational only.
- **User Story 3 (Phase 5)**: Depends on Foundational only for its own Error-screen mechanics (T023-T026); its final integration test (T027) additionally depends on User Story 2's T021.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Can start after Foundational, in parallel with User Story 1 — the Warning screen's own display/transition logic needs nothing from US1's Agent Selection/Preview screens to be correct on its own.
- **User Story 3 (P3)**: Can start after Foundational, in parallel with User Story 1 and 2, for its own Error-screen mechanics — only its final cross-story integration test (T027) needs User Story 2's transition already in place.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Within User Story 1: T007-T010 (Update/View tests, different concerns) in parallel; T011 (cli branch tests) in parallel with those, different package.

### Parallel Opportunities

- Once Foundational is done: **User Story 1 and User Story 2 can proceed fully in parallel** by different contributors (different screens, same package but non-overlapping `Update`/`View` cases) — the same independence 003's and 006's parallel stories had. User Story 3's own mechanics (not its final T027) can proceed in parallel too.
- Within User Story 1: T007, T008, T009, T010, T011 in parallel (five different test concerns).
- Within User Story 2: T019, T020 in parallel.
- Within User Story 3: T023, T024 in parallel.
- Within Polish: T028 and T029 in parallel.

---

## Parallel Example: All three stories' test-writing together

```bash
# Once Foundational is done, these can all be written in parallel:
Task: "Update tests for ScreenAgentSelection/ScreenPreview/ScreenInstalling->Success in internal/tui/update_test.go (US1)"
Task: "Update tests for ScreenInspect->Warning transition in internal/tui/update_test.go (US2)"
Task: "Update tests for ScreenInstalling->Error and Inspect->Error transitions in internal/tui/update_test.go (US3)"
```

(Note: US1/US2/US3 add distinct test functions to the same `update_test.go`/`view_test.go` files — genuinely parallelizable by different contributors as long as each adds new functions rather than editing shared ones, the same file-level coordination note every multi-contributor story in this project's tasks.md files has carried.)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (`Empty` field, `Model`/`Screen` skeleton, `Init`).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/tui/... ./internal/cli/...` green, independently; a real `misterspec init` run in an empty directory works start to finish.
5. This alone already delivers §4.1's promised public experience for the common case — a human bootstrapping a fresh project with no prior knowledge of agent IDs.

### Incremental Delivery

1. Setup + Foundational (`Empty`, `Model`/`Screen` skeleton ready).
2. Add US1 → validate independently → full happy path usable (MVP).
3. Add US2 (independent of US1, can be parallel) → validate independently → safety warning usable.
4. Add US3 (independent of US1/US2 for its own mechanics, can be parallel) → validate independently, then complete T027 once US2 is also done → full cross-story acceptance scenario demonstrable.
5. Polish (Phase 6), including the full-module `-race` regression run (T031) and the explicit 007/008 non-interactive-path regression re-confirmation (SC-005).

### Team Strategy

Once Foundational is done, three developers can take User Story 1, 2,
and 3 largely in parallel — the same independence 003-entity-creation's
and 006-agent-adapter's parallel stories had, though whoever finishes
User Story 3's own mechanics first should coordinate with User Story
2's owner for the one task (T027) that spans both.

---

## Notes

- [P] tasks touch different files (or clearly separate functions within a shared test file) with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `internal/tui.Model` holds UI state only, never filesystem business rules (§37's own explicit rule, plan.md's Constitution Check) — every actual decision still comes from `bootstrap`/`agents`/`installer`.
- The non-interactive `--agent` path (008-cli-cobra) must remain byte-for-byte unchanged — T011(c) and T031's explicit re-confirmation are this feature's own regression discipline for that guarantee (SC-005).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
