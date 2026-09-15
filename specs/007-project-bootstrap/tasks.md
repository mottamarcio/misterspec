---

description: "Task list template for feature implementation"
---

# Tasks: Project Bootstrap

**Input**: Design documents from `/specs/007-project-bootstrap/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/bootstrap.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 001-core-foundation's, 005-embedded-kit's, and 006-agent-adapter's full test suites are explicit, named regression gates (Polish), since this feature composes all three unmodified.

**Organization**: Tasks are grouped by user story. This feature's dependency shape: **User Story 1 (`Inspect`) has no dependency on the others** — it composes only `project.Detect` and `agents.CurrentInstall`, both already proven, and is tested purely against fixture directory states. **User Story 2 (`Bootstrap`) depends on User Story 1's `Inspect`** — it calls `Inspect` internally as its pre-write rejection check (research.md, FR-004). **User Story 3 (`Verify`) depends on both**: implementation-wise it is `Inspect` plus one comparison (research.md's DRY decision), but its own test, per spec.md's Independent Test wording, needs a real `Bootstrap` output to verify against — so it cannot be meaningfully tested until User Story 2's implementation exists, the same "genuinely depends on the prior story's real output" shape 006-agent-adapter's US3/US2 had.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/bootstrap/            # NEW package — Setup, US1, US2, US3
internal/example/               # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the directory skeleton this feature adds.

- [X] T001 Create `internal/bootstrap/` directory with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: Module builds (`go build ./...`) with the new empty package before any user story work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: This feature modifies no existing package (`internal/project`, `internal/installer`, and `internal/agents` all remain unmodified per research.md) — there is no shared infrastructure to build before User Story 1 beyond the Setup skeleton. This phase is intentionally empty; User Story 1 starts directly after Setup.

**Checkpoint**: Foundation ready (trivially — Setup alone suffices). User Story 1 implementation can begin.

---

## Phase 3: User Story 1 - Inspect a Target Directory Before Bootstrapping (Priority: P1) 🎯 MVP

**Goal**: Determine, for a target directory and without writing anything, whether it is not yet a misterspec project, or already is one (and if so, which agent, if any, is installed there).

**Independent Test**: Inspect a fresh, empty directory (reported as uninitialized) and a directory with an existing project and a known installed agent (reported as already initialized, naming that agent) — no filesystem writes involved in either case.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T003.

- [X] T002 [P] [US1] Unit tests for `Inspect` in `internal/bootstrap/inspect_test.go`: an empty fixture directory reports `Initialized == false`; a fixture project with a known agent recorded (via a fixture `install.json`) reports `Initialized == true, AgentInstalled == true, InstalledAgent == "claude-code"`; a fixture project with no `install.json` reports `Initialized == true, AgentInstalled == false, InstalledAgent == ""` (Acceptance Scenario 3); a fixture directory with a malformed `.misterspec/config.yaml` returns an error satisfying `errors.Is(err, project.ErrInvalidConfiguration)`, distinct from the uninitialized case (Edge Case).

### Implementation for User Story 1

- [X] T003 [US1] Implement `InspectResult` and `Inspect` in `internal/bootstrap/inspect.go` — calls `project.Detect(targetDir)` (propagating `ErrNotInitialized` as `InspectResult{Initialized: false}, nil` and any other error, i.e. `ErrInvalidConfiguration`, as `InspectResult{}, err`), then `agents.CurrentInstall(project.Root)` when detected. Depends on T002.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/bootstrap/... -run TestInspect` passes on its own, with zero dependency on `Bootstrap`/`Verify`.

---

## Phase 4: User Story 2 - Bootstrap a New Project for a Chosen Agent (Priority: P2)

**Goal**: Given an uninitialized target directory and one already-chosen, valid agent ID, atomically bootstrap a new misterspec project — configuration, kit resources, and agent Skills — with the specific outcome of every part reported.

**Independent Test**: Bootstrap a fresh temporary directory for a known, registered agent (using a fixture `Registry` and fixture Skills `fs.FS`) and confirm the configuration file, kit resources, and agent Skills all land correctly with specific per-resource outcomes; confirm both an already-initialized target and an unregistered agent ID are rejected before anything is written. Depends on User Story 1's `Inspect` (used internally as the pre-write rejection check).

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T006–T008.

- [X] T004 [P] [US2] Unit tests for `writeDefaultConfig` in `internal/bootstrap/config_test.go`: writes a `.misterspec/config.yaml` that `project.Load` reads back successfully with every `Default*` value from `internal/project` (`DefaultArtifactsDir`, `DefaultRawDir`, `DefaultKnowledgeDir`, `DefaultConstitutionPath`, `DefaultLearningsDir`, `DefaultProgramsRoot`, `DefaultIDWidth`) and the given `agentID`; written atomically (reusing `installer.WriteAtomicFile`).
- [X] T005 [P] [US2] Filesystem-integration tests for `Bootstrap` in `internal/bootstrap/bootstrap_test.go`, using a fixture `*agents.Registry` (one fake `Adapter`, the same inline-fake pattern 006-agent-adapter's `registry_test.go` used — no dependency on the real `claude` package) and a fixture Skills `fs.FS`: (a) bootstrapping a fresh, non-existent target directory creates it and succeeds, with `ConfigWritten == true`, every `TemplateOutcomes` entry `Installed`, and `AgentInstall` matching the fake adapter's result; (b) bootstrapping an already-initialized target (a fixture project) returns `errors.Is(err, bootstrap.ErrAlreadyInitialized)` and leaves the target's existing files untouched; (c) bootstrapping a fresh target with an unregistered agent ID returns `errors.Is(err, bootstrap.ErrUnknownAgent)` before any file is written (assert the target directory remains empty).

### Implementation for User Story 2

- [X] T006 [US2] Implement `writeDefaultConfig` in `internal/bootstrap/config.go` — builds a `project.Configuration` from `internal/project`'s exported `Default*` constants plus the given `agentID`, marshals via `yaml.Marshal`, writes atomically via `installer.WriteAtomicFile` to `<targetDir>/.misterspec/config.yaml`. Depends on T004.
- [X] T007 [US2] Implement `BootstrapOutcome`, `ErrAlreadyInitialized`, `ErrUnknownAgent`, and `Bootstrap` in `internal/bootstrap/bootstrap.go`: call `Inspect(targetDir)` and reject if already initialized; call `registry.Get(agentID)` and reject if not found; `os.MkdirAll(targetDir)`; call `writeDefaultConfig`; call `installer.Install(targetDir, false)`; call `adapter.Install(ctx, agents.InstallRequest{targetDir, skills, false})`; assemble and return `BootstrapOutcome`. Depends on T003, T006, T005.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/bootstrap/... -run TestBootstrap` passes on its own.

---

## Phase 5: User Story 3 - Verify a Completed Bootstrap (Priority: P3)

**Goal**: Confirm a completed bootstrap is genuinely usable — the directory detects as a valid project, and the agent actually installed matches the one requested.

**Independent Test**: Bootstrap a fixture directory (User Story 2's real `Bootstrap`), then verify it — confirming the project detects successfully and the recorded agent matches exactly what was requested. Genuinely depends on User Story 2's `Bootstrap` having actually run — there is no meaningful "verify a completed bootstrap" test without a real bootstrap to verify (the same 006-agent-adapter US3/US2 shape).

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T009.

- [X] T008 [US3] Filesystem-integration tests for `Verify` in `internal/bootstrap/verify_test.go`, building on a real `Bootstrap` call from T005/T007 (fixture registry, fixture Skills `fs.FS`): after a successful bootstrap, `Verify(targetDir, "claude-code")` returns `Detected == true, InstalledAgent == "claude-code", AgentMatches == true`; verifying against a different expected agent ID returns `Detected == true, AgentMatches == false` (mismatch, not an error); verifying a directory that was never bootstrapped returns `Detected == false`. Depends on T007.

### Implementation for User Story 3

- [X] T009 [US3] Implement `VerifyResult` and `Verify` in `internal/bootstrap/verify.go` — calls `Inspect(targetDir)` and compares `InstalledAgent` against `wantAgentID`; no independent detection or record-reading logic (research.md). Depends on T003, T008.

**Checkpoint**: All three user stories pass their own tests — the full inspect → bootstrap → verify cycle works end to end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T010 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 006-agent-adapter's packages included) and fix any findings.
- [X] T011 [P] Verify/extend the package-level doc comment on `internal/bootstrap`, cross-checked against `specs/007-project-bootstrap/contracts/bootstrap.md`.
- [X] T012 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: `Inspect` on a fresh directory, `Bootstrap` it for a fixture agent, `Inspect` again confirming it's now initialized, then `Verify` confirming the match.
- [X] T013 Reconcile `specs/007-project-bootstrap/contracts/bootstrap.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T014 Full regression run: `go test ./...` across the entire module (001 through 007) green, `go vet ./...` clean, `gofmt -l .` empty, `go test ./internal/bootstrap/... -race` clean (concurrency-sensitive: `Bootstrap`'s multi-part write sequence, mirroring 003-entity-creation's `-race` discipline for `Create`).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — intentionally empty for this feature (no existing package changes required, research.md).
- **User Story 1 (Phase 3)**: Depends on Setup only.
- **User Story 2 (Phase 4)**: Depends on User Story 1's `Inspect` (T003) — a genuine implementation dependency, not just convenience.
- **User Story 3 (Phase 5)**: Depends on User Story 1's `Inspect` (T003) for implementation, and on User Story 2's `Bootstrap` (T007) for a meaningful test.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Setup. No dependency on US2/US3.
- **User Story 2 (P2)**: Depends on User Story 1 (`Inspect` is called internally by `Bootstrap`, FR-004).
- **User Story 3 (P3)**: Depends on User Story 1 (`Verify` is `Inspect` plus a comparison) and, for its test, on User Story 2 (needs a real bootstrap to verify against).

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- User Story 1 complete and regression-checked before User Story 2's implementation begins.
- User Story 2 complete before User Story 3's test can meaningfully run.

### Parallel Opportunities

- T004 (`writeDefaultConfig` tests) and T005 (`Bootstrap` filesystem-integration tests) in parallel — different files, though T005's own test authoring can proceed even before T006 exists (tests are written first).
- Within Polish: T010 and T011 in parallel.
- Unlike 006-agent-adapter's US1/US2, this feature's three stories form a strict chain (US1 → US2 → US3) — no cross-story parallelism is available once Setup is done; each story's implementation genuinely gates the next.

---

## Parallel Example: User Story 2's tests

```bash
# Once User Story 1 (Inspect) is done, these two test files can be written in parallel:
Task: "Unit tests for writeDefaultConfig in internal/bootstrap/config_test.go"
Task: "Filesystem-integration tests for Bootstrap in internal/bootstrap/bootstrap_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (trivially — nothing to do).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/bootstrap/... -run TestInspect` green, independently.
5. This alone already proves the read-only pre-flight check `misterspec init` will eventually run — usable by any future caller before anything destructive happens.

### Incremental Delivery

1. Setup + (trivial) Foundational.
2. Add US1 → validate independently → inspection usable (MVP).
3. Add US2 (depends on US1) → validate independently → full bootstrap usable.
4. Add US3 (depends on US1 and US2) → validate independently → verification usable.
5. Polish (Phase 6), including the full-module `-race` regression run (T014).

### Team Strategy

This feature's strict US1 → US2 → US3 chain means, unlike 006-agent-adapter's
parallel US1/US2, a single developer working sequentially is the natural
shape — there is no meaningful way to split US2 and US3 across
developers before US1 (and, for US3, US2) lands.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `Bootstrap` rejects before writing anything — an already-initialized target or an unregistered agent ID must never leave a partially-written resource (plan.md's Constitution Check, Principles VII and VIII).
- No new locking is introduced for `Bootstrap` — `Inspect`-then-reject is the entire concurrency story for this feature (research.md, Constitution Principle IV).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
