---

description: "Task list template for feature implementation"
---

# Tasks: Read-Only Deterministic Operations

**Input**: Design documents from `/specs/002-read-operations/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/operations.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE) and this plan's Constitution Check gate. Additionally, this feature extends 001-core-foundation's shipped types additively — 001-core-foundation's own test suite is a mandatory regression gate (T005), not optional.

**Organization**: Tasks are grouped by user story (from spec.md) to enable independent implementation and testing of each story. Unlike 001-core-foundation, User Story 2 genuinely depends on User Story 1's implementation (not just its tests) — see Dependencies below; this is called out explicitly rather than glossed over.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

Single Go module (unchanged from 001-core-foundation):

```text
internal/ids/          # existing package — Foundational extensions here
internal/artifacts/    # existing package — Foundational extensions here
internal/operations/   # NEW package for this feature
internal/example/       # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the new package this feature adds.

- [X] T001 Create the `internal/operations/` directory with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: Module builds (`go build ./...`) with the new empty package before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The additive extensions to 001-core-foundation's `internal/ids` and `internal/artifacts` that more than one user story in this feature needs (per data-model.md and research.md) — done together so 001-core-foundation's regression suite is checked once, against the complete extended surface, not piecemeal.

**⚠️ CRITICAL**: No user story implementation may begin until this phase is complete, including T005's regression check.

- [X] T002 [P] Add `ParseAny(raw string) (EntityID, error)` to `internal/ids/ids.go` — derives `EntityType` (via the existing `TypeForPrefix`) and width from `raw` itself, wrapping `ErrInvalidIDSyntax` on failure the same way `Parse` does (research.md "Promote raw-ID-string parsing").
- [X] T003 [P] Add a `Paths map[int][]string` field to `ScanResult` in `internal/ids/scan.go`, populated in `buildScanResult` for every number found (not only duplicates) — `IDs` and `Duplicates` must remain unchanged in shape and behavior.
- [X] T004 [P] Add a `For *ids.EntityID` field to `Metadata` in `internal/artifacts/metadata.go`, decode a `for` YAML key in `internal/artifacts/parser.go` the same way `parent` already is, and refactor `parseFieldID` in `internal/artifacts/parser.go` to delegate to `ids.ParseAny` (T002) instead of duplicating its logic.
- [X] T005 Run `go test ./internal/project/... ./internal/artifacts/... ./internal/ids/... ./internal/example/...` and confirm every 001-core-foundation test still passes unmodified after T002–T004. Fix any regression before proceeding — this is a hard gate, not advisory.

**Checkpoint**: Foundation extended and verified non-regressive; US1, US2, US3 implementation can now begin (US2 additionally depends on US1's implementation — see Dependencies).

---

## Phase 3: User Story 1 - Locate and Inspect Any Entity by ID (Priority: P1) 🎯 MVP

**Goal**: Given an entity's ID alone, find exactly where it lives and retrieve its complete structured metadata.

**Independent Test**: Populate a fixture project with several entities (including a duplicated ID under different parents, and a Task ID inside a shared `tasks.md`) and confirm resolve/inspect behave per spec.md's acceptance scenarios — no other story's code required.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T008–T009.

- [X] T006 [P] [US1] Filesystem-integration tests for `Resolve` — a unique match, a not-found ID, an ambiguous ID (two artifacts with the same number under different parents, asserting `*AmbiguousIDError` names both locations), and a Task ID resolving to its owning `tasks.md#TASK-NNN` — in `internal/operations/resolve_test.go`.
- [X] T007 [P] [US1] Filesystem-integration tests for `Inspect` — a well-formed Spec's full metadata, a not-found ID, ambiguous-ID propagation from `Resolve`, and a Task's narrowed metadata (`ID`, `Status` derived from its checkbox, `Parent` derived from its `tasks.md`'s `for:` field) — in `internal/operations/inspect_test.go`.

### Implementation for User Story 1

- [X] T008 [US1] Implement `ResolvedLocation`, `ErrEntityNotFound`, `ErrEntityAmbiguous`, `AmbiguousIDError`, and `Resolve(root, cfg, rawID)` (uses `ids.ParseAny` and `ids.Scan`'s `Paths`: 0 paths → not found, 1 → `ResolvedLocation`, >1 → ambiguous) in `internal/operations/resolve.go`. Depends on T002, T003, T006.
- [X] T009 [US1] Implement `InspectResult`, `ErrInvalidTarget`, and `Inspect(root, cfg, rawID)` — delegates to `Resolve`, then `artifacts.ParseMetadata` for file-based types; for a Task, reads the checkbox state and the owning `tasks.md`'s `For` field instead (research.md's Task metadata scope decision) — in `internal/operations/inspect.go`. Depends on T004, T008, T007.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/operations/... -run 'Resolve|Inspect'` passes on its own.

---

## Phase 4: User Story 2 - Discover Structural Relationships (Priority: P2)

**Goal**: Given an entity, discover its structural parent and its direct children (optionally filtered by type), using only fixed nesting rules.

**Independent Test**: Against a fixture Program with two Features (one containing two Specs), confirm `Parent`/`Children` behave per spec.md's acceptance scenarios. **Note**: unlike 001-core-foundation's stories, this one is not independent of User Story 1's *implementation* — `Parent` calls `Inspect` and `Children` calls `Resolve` directly (data-model.md's flow), so T008–T009 must exist first; only its *tests* can be written before that.

### Tests for User Story 2

> Write these tests FIRST; they may fail to compile until T008–T009 (US1) exist — that is expected and consistent with TDD's red step here.

- [X] T010 [P] [US2] Filesystem-integration tests for `Parent` — a Spec with a Feature parent, a Program (`HasParent: false`, no error), and a Task (parent resolved via its `tasks.md`'s `for:` field) — in `internal/operations/parent_test.go`.
- [X] T011 [P] [US2] Filesystem-integration tests for `Children` — a Program with two Features (unfiltered and type-filtered), and a type filter that matches none (empty result, not an error) — in `internal/operations/children_test.go`.

### Implementation for User Story 2

- [X] T012 [US2] Implement `ParentResult` and `Parent(root, cfg, rawID)` — delegates to `Inspect`, reads `Metadata.Parent` (already normalized for Task by T009), returns `HasParent: false` for a type with no parent concept — in `internal/operations/parent.go`. Depends on T009, T010.
- [X] T013 [US2] Implement `Children(root, cfg, parentRawID, filterType)` — resolves the parent via `Resolve`, calls `ids.Scan` for the candidate child type(s), keeps only entries whose `Paths` value is nested under the parent's directory (research.md's "reuses Scan + prefix filter" decision, no new `ids` primitive) — in `internal/operations/children.go`. Depends on T008, T003, T011.

**Checkpoint**: User Stories 1 AND 2 both pass their own tests.

---

## Phase 5: User Story 3 - Discover Files and Verify Content Integrity (Priority: P3)

**Goal**: List files in a named area and compute a deterministic content fingerprint for any file, without interpreting its contents.

**Independent Test**: Against a fixture directory with mixed file types and an empty subdirectory, confirm `Inventory`/`Fingerprint` behave per spec.md's acceptance scenarios — fully independent of User Story 1 or 2.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T017–T018.

- [X] T014 [P] [US3] Filesystem-integration tests for `Inventory` — a populated directory, an empty directory, a not-yet-created-but-validly-located directory (empty result, not an error), and a traversal attempt (`ErrPathOutsideProject`) — in `internal/operations/inventory_test.go`.
- [X] T015 [P] [US3] Filesystem-integration tests for `Fingerprint` — the identical digest across repeated calls for the same content, a not-found path (`ErrArtifactNotFound`), and a traversal attempt (`ErrPathOutsideProject`) — in `internal/operations/fingerprint_test.go`.

### Implementation for User Story 3

- [X] T016 [US3] Export `internal/artifacts`'s private `relativeWithinRoot` as `RelativeWithinRoot` in `internal/artifacts/paths.go` — purely additive, no existing call site changes behavior. Depends on T014, T015.
- [X] T017 [P] [US3] Implement `FileEntry` and `Inventory(root, dir)` (lists files directly under `dir`, empty slice for an empty/nonexistent-but-valid directory) in `internal/operations/inventory.go`. Depends on T016.
- [X] T018 [P] [US3] Implement `Fingerprint` (`Algorithm`, `Digest`, `String()`) and `Fingerprint(root, path)` (streamed via `io.Copy` into `crypto/sha256`, per research.md) in `internal/operations/fingerprint.go`. Depends on T016.

**Checkpoint**: All three user stories independently pass their own tests.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T019 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation's packages included) and fix any findings.
- [X] T020 [P] Verify/extend the `internal/operations` package-level doc comment (`doc.go`) summarizing its exported contract, cross-checked against `specs/002-read-operations/contracts/operations.md`.
- [X] T021 Add a compiled, run-in-CI example in `internal/example` (extending 001-core-foundation's existing package) exercising `quickstart.md`'s full flow end-to-end: `Resolve` → `Inspect` → `Parent` → `Children` → `Inventory` → `Fingerprint`.
- [X] T022 Reconcile `specs/002-read-operations/contracts/operations.md`'s signatures and error-code mapping table against the actual implementation; fix any drift introduced during implementation (same discipline as 001-core-foundation's T025).
- [X] T023 Full regression run: `go test ./...` across the entire module (both 001-core-foundation and this feature's code) green, `go vet ./...` clean, `gofmt -l .` empty.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories, and includes a hard regression gate (T005) against 001-core-foundation's existing suite.
- **User Stories (Phase 3–5)**: All depend on Foundational completion. **User Story 2 additionally depends on User Story 1's implementation** (T008, T009) — this is a real dependency, not just a sequencing convenience, because `Parent` and `Children` call `Inspect`/`Resolve` directly. User Story 3 has no dependency on User Story 1 or 2.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Can start after Foundational, but its *implementation* tasks (T012, T013) depend on US1's implementation tasks (T008, T009) being done first. Its tests (T010, T011) can be written earlier, in parallel with US1.
- **User Story 3 (P3)**: Can start after Foundational. No dependency on US1 or US2 — fully independent, like every story in 001-core-foundation was.

### Within Each User Story

- Tests written and failing (or, for US2, at least failing to reference not-yet-built US1 symbols) before implementation (Constitution Principle V).
- Foundational types before story-specific behavior.
- Story complete and its own tests green before moving to the next priority, if working sequentially.

### Parallel Opportunities

- All Foundational tasks marked [P] (T002, T003, T004) touch different files and can run in parallel; T005 (regression gate) must run after all three.
- Within US1: T006 and T007 (tests, different files) in parallel; T008 before T009 (Inspect depends on Resolve).
- Within US2: T010 and T011 (tests) in parallel, though both wait on US1's implementation existing to be meaningfully green; T012 and T013 can proceed in parallel once US1 is done (different files, no dependency on each other).
- Within US3: T014 and T015 (tests) in parallel; T017 and T018 (implementation) in parallel once T016 (the shared export) is done.
- US3 as a whole can proceed in parallel with US1/US2 by a different contributor, since it has no dependency on either.

---

## Parallel Example: Foundational Phase

```bash
# Launch all Foundational extension tasks together (different files):
Task: "Add ids.ParseAny in internal/ids/ids.go"
Task: "Add ScanResult.Paths in internal/ids/scan.go"
Task: "Add Metadata.For in internal/artifacts/metadata.go + parser.go"
# Then, only after all three: T005's regression run.
```

## Parallel Example: User Story 3 (independent of US1/US2)

```bash
# Launch all US3 test tasks together (different files):
Task: "Filesystem-integration tests for Inventory in internal/operations/inventory_test.go"
Task: "Filesystem-integration tests for Fingerprint in internal/operations/fingerprint_test.go"

# After T016 (shared export), launch both implementations together:
Task: "Implement Inventory in internal/operations/inventory.go"
Task: "Implement Fingerprint in internal/operations/fingerprint.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (including the T005 regression gate).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/operations/...` green for Resolve/Inspect, independently.
5. This alone already lets any later Skill/operation locate and inspect any entity by ID — the single most-cited deterministic step across the architecture spec's own Skill contracts (§41-49).

### Incremental Delivery

1. Setup + Foundational (extended, regression-checked) → shared primitives ready.
2. Add US1 → validate independently → resolve/inspect usable (MVP).
3. Add US2 → validate independently (now genuinely depends on US1's code, not just its own tests) → structural navigation usable.
4. Add US3 → validate independently, any time after Foundational (no ordering constraint relative to US1/US2) → file discovery and fingerprinting usable.
5. Polish (Phase 6) once all three are in, including the full-module regression run (T023).

### Parallel Team Strategy

With more than one contributor:

1. Complete Setup + Foundational together first (it blocks everyone, and its regression gate should be checked once for the whole team).
2. Developer A takes US1 (and, once US1's implementation lands, US2 — since US2 genuinely depends on it). Developer B can take US3 immediately after Foundational, fully in parallel, since it has no dependency on US1/US2.
3. Stories integrate at Phase 6 (Polish), where the full-module regression run (T023) is the final gate.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- No task in this feature writes to the filesystem outside of test fixtures under `t.TempDir()` — every operation here is read-only by design (FR-014; see plan.md's Constitution Check, Principle II).
- This feature modifies 001-core-foundation's shipped files (`ids.go`, `scan.go`, `metadata.go`, `parser.go`, `paths.go`) — every such change is additive only (new field, new exported function, or an internal delegation), and T005/T023 are the explicit gates proving nothing broke.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
