---

description: "Task list template for feature implementation"
---

# Tasks: Atomic Entity Creation

**Input**: Design documents from `/specs/003-entity-creation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/creation.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). This is the first feature that writes to the filesystem, so 001-core-foundation's and 002-read-operations's full suites are explicit regression gates (T016), not optional.

**Organization**: Tasks are grouped by user story (from spec.md). Unlike 002-read-operations, User Story 1 (`Create`) and User Story 2 (`CreateArtifact`) are genuinely independent of each other's implementation — neither calls the other; both call only the Foundational packages (`lock`, `templates`) and 002's `Resolve`. User Story 3 (concurrency/recovery) is different again: its *mechanism* (the lock's mutual exclusion and staleness recovery) is built in Foundational, because `Create`/`CreateArtifact` cannot compile or be trusted without it — US3's own phase is validation of that mechanism at the `Create` level, not new production code. All three of these dependency shapes are called out explicitly below rather than forced into a uniform pattern.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

Single Go module (unchanged):

```text
internal/lock/         # NEW package — Foundational (prerequisite for all stories)
internal/templates/    # NEW package — Foundational (prerequisite for US1, US2)
internal/operations/   # existing package — US1 (create.go), US2 (create_artifact.go), US3 (concurrency test)
internal/example/       # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the two new packages this feature adds.

- [X] T001 Create `internal/lock/` and `internal/templates/` (with a `files/` subdirectory) directories, each with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: Module builds (`go build ./...`) with both new empty packages before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: `internal/lock` and `internal/templates` are not partial stubs here — they are fully built, independently tested, complete primitives, because `Create` and `CreateArtifact` cannot even compile (let alone be trusted for FR-006/FR-007) without a working lock and working templates. This is why this phase is larger than 001/002's Foundational phases.

**⚠️ CRITICAL**: No user story implementation may begin until this phase — including both packages' own tests passing — is complete.

- [X] T002 [P] Unit tests for `lock.Acquire`/`Release` — successful acquire-then-release; a second `Acquire` call blocks until the first `Release`s; a lock file older than `staleAfter` is recovered automatically; `Acquire` returns `ErrLockTimeout` when the lock is genuinely held beyond `timeout` — in `internal/lock/lock_test.go`.
- [X] T003 [P] Unit tests for `templates.Render` — one case per `Kind` (Program, Feature, Spec, Knowledge, Learning, Plan, Tasks, Validation) asserting the rendered frontmatter and body match `docs/architecture-specification.md` §22-31 exactly, plus a mismatched-data-type rejection — in `internal/templates/templates_test.go`.
- [X] T004 Implement `lock.Acquire`/`Release`, `ErrLockTimeout`, and the `DefaultStaleAfter`/`DefaultTimeout` constants (`O_CREATE|O_EXCL` acquisition, staleness-based recovery, bounded retry) in `internal/lock/lock.go`. Depends on T002.
- [X] T005 [P] Write the 8 embedded template files (`program.md.tmpl`, `feature.md.tmpl`, `spec.md.tmpl`, `knowledge.md.tmpl`, `learning.md.tmpl`, `plan.md.tmpl`, `tasks.md.tmpl`, `validation.md.tmpl`) under `internal/templates/files/`, matching `docs/architecture-specification.md` §22-31's frontmatter and body schemas exactly.
- [X] T006 Implement `templates.Kind`, the per-kind `*Data` structs, `go:embed` wiring, and `Render` in `internal/templates/templates.go`. Depends on T005, T003.

**Checkpoint**: Foundation ready — `go test ./internal/lock/... ./internal/templates/...` passes on its own; `Create` and `CreateArtifact` implementation can now begin (both depend only on this phase and on 002-read-operations's existing `Resolve`, `ids.Scan`/`NextID`, `artifacts.ResolvePath`/`RelativeWithinRoot` — not on each other).

---

## Phase 3: User Story 1 - Atomically Create a New Entity (Priority: P1) 🎯 MVP

**Goal**: Given a type and (where required) a parent, allocate the next ID and write the initial artifact as one all-or-nothing operation.

**Independent Test**: Populate a fixture project and request creation of each supported type (including a Feature under a nonexistent Program) and confirm sequential ID allocation, immediate inspectability, and nothing-written-on-rejection — no other story's code required.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T008.

- [X] T007 [P] [US1] Filesystem-integration tests for `Create` — a Program with no parent; a Feature under a valid Program; three Specs created in sequence under the same Feature (asserting `SPEC-001`, `SPEC-002`, `SPEC-003`); a Feature requested under a nonexistent Program (`ErrInvalidParent`, nothing written); a Knowledge/Learning entity with a supplied slug; a request whose target directory already exists (`ErrAlreadyExists`) — in `internal/operations/create_test.go`.

### Implementation for User Story 1

- [X] T008 [US1] Implement `CreateRequest`, `CreateResult`, `ErrInvalidParent`, `ErrAlreadyExists`, `ErrUnsupportedType`, and `Create` — holds `lock.Acquire`/`defer Release` for the whole operation; validates a required parent via `operations.Resolve`; allocates via `ids.Scan`+`ids.NextID`; computes the canonical path via `artifacts.ResolvePath`; rejects an existing target; renders via `templates.Render`; writes atomically (temp file in the target directory → `fsync` → `os.Rename`) — in `internal/operations/create.go`. Depends on T004, T006, T007.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/operations/... -run TestCreate` passes on its own.

---

## Phase 4: User Story 2 - Create Subordinate Spec Artifacts (Priority: P2)

**Goal**: Given an existing Spec's ID, write its Plan, Tasks, or Validation artifact at its fixed location — no new independent ID.

**Independent Test**: Against a fixture project with one existing Spec, request a Plan, then Tasks, then Validation for it, and confirm exact paths/frontmatter; confirm a nonexistent Spec and a duplicate target are both rejected. **Note**: this story does not call `Create` — it depends only on Phase 2 (Foundational) and on 002-read-operations's existing `Resolve`, so it can proceed in parallel with User Story 1, not after it.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T010.

- [X] T009 [P] [US2] Filesystem-integration tests for `CreateArtifact` — a Plan, a Tasks artifact, and a Validation artifact each created for the same existing Spec, asserting exact paths and a correct `for:` field; a Tasks artifact requested for a nonexistent Spec ID (`ErrInvalidParent`); a second Validation artifact requested for a Spec that already has one (`ErrAlreadyExists`); an unsupported `Kind` (e.g. `TypeSpec`) rejected as `ErrUnsupportedType` — in `internal/operations/create_artifact_test.go`.

### Implementation for User Story 2

- [X] T010 [US2] Implement `CreateArtifactRequest`, `CreateArtifactResult`, and `CreateArtifact` — same lock-held, validate-then-write-atomically shape as `Create`, but computes a fixed path under the `For` Spec's own directory instead of allocating an ID — in `internal/operations/create_artifact.go`. Depends on T004, T006, T009.

**Checkpoint**: User Stories 1 AND 2 both independently pass their own tests.

---

## Phase 5: User Story 3 - Safe, Recoverable Concurrency Protection (Priority: P3)

**Goal**: Prove, at the `Create` level, the two guarantees the Foundational lock (T004) was built to provide: concurrent requests never collide, and a stale lock never permanently blocks recovery.

**Independent Test**: Issue two concurrent `Create` calls for the same type and confirm both succeed with different, sequential IDs; pre-place a stale lock file and confirm the next `Create` call still succeeds unaided. **Note**: this phase adds no new production code — the mechanism was already built and unit-tested in Phase 2 (T002, T004). This phase is end-to-end validation that `Create` actually holds the lock correctly, so it depends on User Story 1 (T008) existing.

### Tests for User Story 3

> These tests exercise already-built mechanism end-to-end; still write them expecting them to pass immediately once T008 exists, and treat a failure here as a `Create`-level integration bug, not a `lock` package bug (T002 already covers the package in isolation).

- [X] T011 [US3] Concurrency and recovery integration tests for `Create` — launch several goroutines calling `Create` for the same entity type simultaneously against one fixture project and assert every result holds a unique, sequential ID with fully-written, valid content; separately, pre-write a `.misterspec/.lock` file with a timestamp older than `lock.DefaultStaleAfter` and confirm a subsequent `Create` call still succeeds without any manual cleanup — in `internal/operations/create_concurrency_test.go`. Depends on T008.

**Checkpoint**: All three user stories independently pass their own tests — `Create`/`CreateArtifact` are safe under concurrency and interruption, not merely "usually fine."

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T012 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation and 002-read-operations's packages included) and fix any findings.
- [X] T013 [P] Verify/extend package-level doc comments on `internal/lock`, `internal/templates`, and the new `operations` functions, cross-checked against `specs/003-entity-creation/contracts/creation.md`.
- [X] T014 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: create a Program, a Feature under it, a Spec under that, then a Plan for the Spec via `CreateArtifact`.
- [X] T015 Reconcile `specs/003-entity-creation/contracts/creation.md`'s signatures and error-code mapping table against the actual implementation; fix any drift introduced during implementation (same discipline as 001's T025 and 002's T022).
- [X] T016 Full regression run: `go test ./...` across the entire module (001-core-foundation, 002-read-operations, and this feature's code) green, `go vet ./...` clean, `gofmt -l .` empty. Run the concurrency test (T011) with `-race` enabled at least once and confirm it is clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories. Both `lock` and `templates` must be complete and self-tested (not stubbed) before any story starts, because `Create`/`CreateArtifact` depend on their full behavior, not just their types.
- **User Stories (Phase 3–5)**: All depend on Foundational completion. **User Story 1 and User Story 2 do not depend on each other** — `CreateArtifact` does not call `Create`. **User Story 3 depends on User Story 1** (T008) — it validates `Create`'s use of the already-built lock mechanism, not a standalone capability.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2. US3 depends on it.
- **User Story 2 (P2)**: Can start after Foundational, in parallel with US1 — calls only `lock`, `templates`, and 002's `Resolve`.
- **User Story 3 (P3)**: Can start only after User Story 1's `Create` (T008) exists; it is validation of `Create`'s correctness under concurrency/interruption, not new capability.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational packages complete and self-tested before any story's implementation.
- Story complete and its own tests green before moving to the next priority, if working sequentially.

### Parallel Opportunities

- T002 and T003 (Foundational tests, different packages) in parallel; T005 (template files) in parallel with either.
- Once Foundational (T002–T006) is complete, **User Story 1 and User Story 2 can proceed fully in parallel** by different contributors — a first in this project's features so far, since neither depends on the other's implementation.
- User Story 3 (T011) must wait for User Story 1 (T008) specifically, regardless of User Story 2's progress.
- Within Polish, T012 and T013 in parallel.

---

## Parallel Example: Foundational Phase

```bash
# Launch Foundational test-writing tasks together (different packages):
Task: "Unit tests for lock.Acquire/Release in internal/lock/lock_test.go"
Task: "Unit tests for templates.Render in internal/templates/templates_test.go"
Task: "Write the 8 embedded .tmpl files under internal/templates/files/"
```

## Parallel Example: User Story 1 and User Story 2 together

```bash
# Once Foundational is done, these two stories need no coordination:
Task: "Filesystem-integration tests for Create in internal/operations/create_test.go"
Task: "Filesystem-integration tests for CreateArtifact in internal/operations/create_artifact_test.go"
# ...followed by their respective implementations, also independently.
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (`lock` + `templates`, both self-tested).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/operations/... -run TestCreate` green, independently.
5. This alone already lets any later Skill create a Program, Feature, Spec, Knowledge entry, or Learning — the core write capability every `/create-*` Skill needs.

### Incremental Delivery

1. Setup + Foundational (lock + templates, fully built and tested) → shared write primitives ready.
2. Add US1 → validate independently → entity creation usable (MVP).
3. Add US2 → validate independently, in parallel with US1 if staffed → subordinate artifact creation usable.
4. Add US3 → validate independently (depends on US1 existing) → concurrency/recovery guarantees proven, not assumed.
5. Polish (Phase 6), including a `-race`-enabled run of the concurrency test (T016).

### Parallel Team Strategy

With more than one contributor:

1. Complete Setup + Foundational together first (it blocks everyone, and both `lock` and `templates` should be reviewed once for the whole team, not twice).
2. Once Foundational is done: Developer A takes US1, Developer B takes US2 — fully independent, unlike 002-read-operations's US1→US2 dependency.
3. Once US1 lands, either developer can pick up US3 (concurrency validation) — it only needs `Create` to exist.
4. Stories integrate at Phase 6 (Polish), where the `-race`-enabled regression run (T016) is the final gate.

---

## Notes

- [P] tasks touch different files/packages with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- This is the first feature in the project that writes to the filesystem — every write goes through the atomic temp-file-then-rename recipe (plan.md's Constraints), and the lock (Foundational, not story-specific) is what makes "reject if already exists" race-free against other misterspec-initiated creators.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
