---

description: "Task list template for feature implementation"
---

# Tasks: CLI Command Layer (Cobra)

**Input**: Design documents from `/specs/008-cli-cobra/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 001-007's full test suites are explicit, named regression gates (Polish), since none of their packages change in this feature.

**Organization**: Tasks are grouped by user story. This feature's shape: **User Story 1 (the ten `internal` commands) depends only on Foundational** (the shared envelope/classify helpers and the bare command tree) — not on User Story 2 or 3. **User Story 2 (`init`) depends only on Foundational** too — its own underlying capability (`bootstrap.Bootstrap`) is independent of User Story 1's ten commands; it shares only the same envelope/classify infrastructure. **User Story 3 (minimal help surface) is a thin hardening pass over whatever commands User Story 1 and 2 already registered** — it sets `Hidden: true` on the `internal` parent command and proves it via tests, genuinely independent of *which* commands exist underneath, but most meaningfully verified once both are in place.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
cmd/misterspec/                # NEW — binary entrypoint — Foundational
internal/cli/                   # NEW package — Foundational, US2, US3
internal/cli/internalcmd/        # NEW package — Foundational, US1
kit/                               # existing package — extended, US2
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add this feature's one new dependency and create the directory skeletons it adds.

- [X] T001 Add `github.com/spf13/cobra` v1.10.2 as a direct dependency (`go get github.com/spf13/cobra@v1.10.2`; updates `go.mod`/`go.sum`).
- [X] T002 [P] Create `cmd/misterspec/`, `internal/cli/`, and `internal/cli/internalcmd/` directories, `internal/cli/` and `internal/cli/internalcmd/` each with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: `go build ./...` succeeds with the new empty packages and the new dependency resolved before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the bare command tree and the shared JSON-envelope/error-classification helpers every user story's commands call — nothing below this phase is registerable without it.

**⚠️ CRITICAL**: No user story implementation may begin until this phase is complete.

- [X] T003 Implement `cmd/misterspec/main.go` — calls `cli.Execute()` and `os.Exit` with its returned code. Depends on T002.
- [X] T004 Implement `newRootCmd` and `Execute` (a bare root `*cobra.Command` — `Use: "misterspec"`, no subcommands yet) in `internal/cli/root.go`. Depends on T003.
- [X] T005 Implement `newInternalCmd` (the `"internal"` parent `*cobra.Command` — not yet `Hidden`, no subcommands attached yet) in `internal/cli/internal.go`, attached to root via `newRootCmd`. Depends on T004.
- [X] T006 [P] Unit tests for `WriteSuccess`/`WriteError` in `internal/cli/internalcmd/envelope_test.go` — success envelope shape (`{"ok":true,"<key>":value}`), error envelope shape (`{"ok":false,"error":{"code","message"}}`), and that `WriteError` returns the correct exit code for a representative error.
- [X] T007 [P] Unit tests for `classify` in `internal/cli/internalcmd/errors_test.go` — one subtest per sentinel in data-model.md's Error Code table (`project.ErrNotInitialized`, `operations.ErrEntityNotFound`, `operations.ErrEntityAmbiguous`, `operations.ErrInvalidTarget`, `operations.ErrInvalidParent`, `operations.ErrAlreadyExists`, `operations.ErrUnsupportedType`, `operations.ErrInvalidSlug`, `artifacts.ErrPathOutsideProject`, `bootstrap.ErrAlreadyInitialized`, `bootstrap.ErrUnknownAgent`), plus an unrecognized error classifying as `("unexpected_failure", 1)`.
- [X] T008 Implement `WriteSuccess`/`WriteError` in `internal/cli/internalcmd/envelope.go`. Depends on T006.
- [X] T009 Implement `classify` in `internal/cli/internalcmd/errors.go`, covering every sentinel from data-model.md's table. Depends on T007.

**Checkpoint**: Foundation ready — `go build ./...` (including `./cmd/misterspec`) succeeds; envelope/classify unit tests green; the root command runs (showing Cobra's default help) before any user story's commands exist.

---

## Phase 3: User Story 1 - Invoke Deterministic Operations via Stable, Scriptable Commands (Priority: P1) 🎯 MVP

**Goal**: Every one of 001-007's ten deterministic operations is invocable from the terminal, producing the JSON/exit-code contract data-model.md defines.

**Independent Test**: Invoke each command against a fixture project directory (in-process via `RunE`, plus a subprocess check for a couple of them) and assert JSON shape and exit code — both success and failure cases — independent of User Story 2 or 3.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T020–T029.

- [X] T010 [P] [US1] Tests for `resolve` in `internal/cli/internalcmd/resolve_test.go` — known ID success (`{"entity":{"id","type","path"}}`, exit 0), unknown ID (`entity_not_found`, exit 3), ambiguous ID (`entity_ambiguous`, exit 3), directory not a project (`project_not_initialized`, exit 6).
- [X] T011 [P] [US1] Tests for `inspect` in `internal/cli/internalcmd/inspect_test.go` — success shape including `status`/`parent`/`depends_on`/`supersedes` (`parent: null` when absent, arrays never `null` when empty), not-found.
- [X] T012 [P] [US1] Tests for `parent` in `internal/cli/internalcmd/parent_test.go` — has-parent case (`{"parent":{"id","type"}}`), no-parent case (`{"ok":true,"parent":null}` — not an error).
- [X] T013 [P] [US1] Tests for `children` in `internal/cli/internalcmd/children_test.go` — no filter, `--type feature` filter, an unrecognized `--type` value (`invalid_argument`, exit 2, before `Children` is called — research.md).
- [X] T014 [P] [US1] Tests for `create` in `internal/cli/internalcmd/create_test.go` — success for each createable type (program/feature/spec/knowledge/learning), invalid parent (`invalid_parent`, exit 5), unsupported type name (`invalid_argument`, exit 2), invalid slug (`invalid_argument`, exit 2).
- [X] T015 [P] [US1] Tests for `create-artifact` in `internal/cli/internalcmd/create_artifact_test.go` — success for plan/tasks/validation, unrecognized kind name (`invalid_argument`, exit 2).
- [X] T016 [P] [US1] Tests for `fingerprint` in `internal/cli/internalcmd/fingerprint_test.go` — success (`{"source":{"path","algorithm","fingerprint"}}`), path outside project (`path_outside_project`, exit 5).
- [X] T017 [P] [US1] Tests for `inventory` in `internal/cli/internalcmd/inventory_test.go` — populated directory, empty/non-existent directory (empty `files: []`, not an error), path outside project.
- [X] T018 [P] [US1] Tests for `validate` in `internal/cli/internalcmd/validate_test.go` — no-arg calls `ValidateProject`, one-arg calls `ValidateEntity`; a valid project (`valid:true`, `findings:[]`, exit 0); a project with a structural anomaly (`ok:true`, `valid:false`, findings populated, **exit 4** — research.md's dedicated decision, distinct from `classify`'s table).
- [X] T019 [P] [US1] Tests for `status` in `internal/cli/internalcmd/status_test.go` — `counts` keyed by singular lowercase type name (research.md), `specs` mirrors `SpecsByState`, `structural_errors` against a fixture project with a few entities of known counts.

### Implementation for User Story 1

- [X] T020 [P] [US1] Implement `NewResolveCmd` in `internal/cli/internalcmd/resolve.go` (calls `operations.Resolve`). Depends on T008, T009, T010.
- [X] T021 [P] [US1] Implement `NewInspectCmd` in `internal/cli/internalcmd/inspect.go` (calls `operations.Inspect`). Depends on T008, T009, T011.
- [X] T022 [P] [US1] Implement `NewParentCmd` in `internal/cli/internalcmd/parent.go` (calls `operations.Parent`). Depends on T008, T009, T012.
- [X] T023 [P] [US1] Implement `NewChildrenCmd` (+ `--type`, its package-local name table — research.md) in `internal/cli/internalcmd/children.go` (calls `operations.Children`). Depends on T008, T009, T013.
- [X] T024 [P] [US1] Implement `NewCreateCmd` (+ `--parent`, `--slug`, its package-local entity-type name table) in `internal/cli/internalcmd/create.go` (calls `operations.Create`). Depends on T008, T009, T014.
- [X] T025 [P] [US1] Implement `NewCreateArtifactCmd` (+ `--for`, its package-local artifact-kind name table) in `internal/cli/internalcmd/create_artifact.go` (calls `operations.CreateArtifact`). Depends on T008, T009, T015.
- [X] T026 [P] [US1] Implement `NewFingerprintCmd` in `internal/cli/internalcmd/fingerprint.go` (calls `operations.Fingerprint`). Depends on T008, T009, T016.
- [X] T027 [P] [US1] Implement `NewInventoryCmd` in `internal/cli/internalcmd/inventory.go` (calls `operations.Inventory`). Depends on T008, T009, T017.
- [X] T028 [P] [US1] Implement `NewValidateCmd` in `internal/cli/internalcmd/validate.go` (calls `validation.ValidateProject`/`ValidateEntity`; computes its own exit code from `len(findings)` per research.md — not via `classify`). Depends on T008, T009, T018.
- [X] T029 [P] [US1] Implement `NewStatusCmd` in `internal/cli/internalcmd/status.go` (calls `operations.Status`; builds `counts` via `EntityType.String()` keys — research.md). Depends on T008, T009, T019.
- [X] T030 [US1] Register all ten `NewXxxCmd()` constructors onto the `"internal"` parent command in `internal/cli/internal.go`. Depends on T020, T021, T022, T023, T024, T025, T026, T027, T028, T029.
- [X] T031 [US1] Subprocess-level integration test in `internal/cli/internal_test.go` — `go build` the real `cmd/misterspec` binary and run `misterspec internal resolve ...` and `misterspec internal status ...` against a fixture project directory as actual subprocesses, confirming exact stdout JSON and exit code end to end (quickstart.md's validation strategy). Depends on T030.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/cli/...` passes on its own; every one of the ten operations is invocable directly by name, with zero dependency on `init` or the help-hiding behavior.

---

## Phase 4: User Story 2 - Bootstrap a New Project Non-Interactively (Priority: P2)

**Goal**: `misterspec init --agent <id> [--dir <path>]` bootstraps a new project via `bootstrap.Bootstrap`, reporting the same structured, per-part outcome, and rejecting an already-initialized target or unregistered agent the same way.

**Independent Test**: Run the public init command against a fresh, empty target with a registered agent (succeeds); against the same directory again (rejected, already initialized); against a fresh directory with an unregistered agent (rejected, unknown agent) — independent of User Story 1's commands.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T034.

- [X] T032 [P] [US2] Add `SkillsFS embed.FS` (`//go:embed skills`) to `kit/kit.go` and `kit/skills/.gitkeep` as its placeholder (research.md — `go:embed` requires at least one matched file).
- [X] T033 [US2] Tests for `init` in `internal/cli/init_test.go` — fresh directory + registered agent (`{"ok":true,"bootstrap":{"project_root","config_written","templates","agent"}}`, exit 0); already-initialized target (`already_initialized`, exit 5); unregistered agent (`unknown_agent`, exit 5); missing `--agent` flag (`invalid_argument`, exit 2). Uses the real `builtin.Default()` registry and the real (currently empty) `kit.SkillsFS`, the same way 007-project-bootstrap's own example test does. Depends on T008, T009, T032.

### Implementation for User Story 2

- [X] T034 [US2] Implement `newInitCmd` (`--agent` required, `--dir` default `.`) in `internal/cli/init.go` — calls `bootstrap.Bootstrap(dir, agent, builtin.Default(), kit.SkillsFS)` and reports via the envelope helper; registered onto root in `newRootCmd`. Depends on T033.
- [X] T035 [US2] Subprocess-level integration test — `go build` the real binary and run `misterspec init --agent claude-code --dir <fresh temp dir>` end to end, confirming the target is bootstrapped and afterward detectable via `project.Detect` (quickstart.md §2). Depends on T034.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/cli/... -run TestInit` passes on its own; `misterspec init` bootstraps a real project from the terminal, with zero dependency on User Story 1's ten commands.

---

## Phase 5: User Story 3 - Keep the Human-Facing Command Surface Minimal (Priority: P3)

**Goal**: `misterspec --help` (and bare `misterspec`) list only the public surface — never any `internal` operation name — while every `internal` command remains fully invocable directly.

**Independent Test**: Run top-level help and confirm it lists only the public command(s); invoke an `internal` command directly by name and confirm it still runs.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T036.

- [X] T037 [US3] Tests in `internal/cli/root_test.go` confirming `misterspec --help`/bare `misterspec` list only `init` (plus Cobra's own built-in `help`/`completion`) and never `internal`, while `misterspec internal resolve ...` (in-process) still runs correctly — plus one subprocess-level `--help` output check reusing T031/T035's real binary. Depends on T030, T034.

### Implementation for User Story 3

- [X] T036 [US3] Set `Hidden: true` on the `"internal"` parent command in `internal/cli/internal.go` (FR-005). Depends on T005.

**Checkpoint**: All three user stories independently verified — the full CLI surface (public `--help`, `init`, and every hidden `internal` command) matches spec.md end to end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T038 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 007-project-bootstrap's packages included) and fix any findings.
- [X] T039 [P] Verify/extend package-level doc comments on `internal/cli` and `internal/cli/internalcmd`, cross-checked against `specs/008-cli-cobra/contracts/cli.md`.
- [X] T040 Reconcile `specs/008-cli-cobra/contracts/cli.md`'s signatures and JSON shapes against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T041 Full regression run: `go test ./...` across the entire module (001 through 008) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/cli/... -race` clean (mutating commands — `create`, `create-artifact`, `init` — touch shared filesystem state, mirroring 003-entity-creation's and 007-project-bootstrap's `-race` discipline).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on Foundational only — not on User Story 1.
- **User Story 3 (Phase 5)**: Depends on Foundational (T005) for its `Hidden: true` target, and its own test (T037) is most meaningfully run once User Story 1 (T030) and User Story 2 (T034) have registered real commands to hide/show — though the `Hidden: true` change itself (T036) needs only T005.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Can start after Foundational, in parallel with User Story 1 — `bootstrap.Bootstrap` needs no `internal` command to exist first.
- **User Story 3 (P3)**: Its `Hidden: true` change (T036) needs only Foundational; its verifying test (T037) is written against whatever US1/US2 have already registered, so it is sequenced last for a meaningful assertion, even though the underlying code change is small and independent.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational's envelope/classify helpers and bare command tree verified working before any story's implementation.
- User Story 1's ten commands are mutually independent (different files) but each individually depends on Foundational's T008/T009.

### Parallel Opportunities

- T006 and T007 (envelope tests, classify tests) in parallel — different files.
- Once Foundational is done: **User Story 1 and User Story 2 can proceed fully in parallel** by different contributors — the same independence 003-entity-creation's and 006-agent-adapter's parallel stories had.
- Within User Story 1: all ten test tasks (T010-T019) in parallel; all ten implementation tasks (T020-T029) in parallel once their own test exists and T008/T009 are done.
- Within Polish: T038 and T039 in parallel.

---

## Parallel Example: User Story 1's ten commands

```bash
# Once Foundational (T008, T009) is done, every command's test+impl pair
# is independent of every other command's:
Task: "Tests for resolve in internal/cli/internalcmd/resolve_test.go"
Task: "Tests for inspect in internal/cli/internalcmd/inspect_test.go"
Task: "Tests for parent in internal/cli/internalcmd/parent_test.go"
Task: "Tests for children in internal/cli/internalcmd/children_test.go"
Task: "Tests for create in internal/cli/internalcmd/create_test.go"
Task: "Tests for create-artifact in internal/cli/internalcmd/create_artifact_test.go"
Task: "Tests for fingerprint in internal/cli/internalcmd/fingerprint_test.go"
Task: "Tests for inventory in internal/cli/internalcmd/inventory_test.go"
Task: "Tests for validate in internal/cli/internalcmd/validate_test.go"
Task: "Tests for status in internal/cli/internalcmd/status_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (bare command tree, envelope/classify helpers).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/cli/...` green, independently; every deterministic operation is now invocable from a terminal.
5. This alone already delivers the machine-facing contract canonical Skills will eventually call — the single largest value slice in this feature.

### Incremental Delivery

1. Setup + Foundational (bare tree, shared helpers ready).
2. Add US1 → validate independently → every operation invocable (MVP).
3. Add US2 (independent of US1, can be parallel) → validate independently → public bootstrap usable.
4. Add US3 (hardening over US1+US2) → validate independently → help surface minimal.
5. Polish (Phase 6), including the full-module `-race` regression run (T041).

### Team Strategy

Once Foundational is done, Developer A can take User Story 1 (ten
commands) and Developer B can take User Story 2 (`init`) fully in
parallel — the same independence 003-entity-creation's and
006-agent-adapter's parallel stories had. User Story 3 is a natural
final pickup for whichever developer finishes first, since its test is
most meaningful once both are registered.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- No command contains business logic beyond argument/flag parsing and JSON shaping — every command calls exactly one existing 001-007 function (plan.md's Constitution Check, Principle IX now realized rather than deferred).
- `validate`'s exit code is the one command-specific exception to `classify`'s error-path table — computed directly from `len(findings)` on its own successful return (research.md).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
