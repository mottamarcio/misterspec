---

description: "Task list template for feature implementation"
---

# Tasks: Structural Validation and Project Status

**Input**: Design documents from `/specs/004-structural-validation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/validation.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 001-core-foundation's, 002-read-operations's, and 003-entity-creation's full suites are explicit regression gates (T015).

**Organization**: Tasks are grouped by user story (from spec.md). This feature has a **linear dependency chain** — a third shape distinct from both prior features: 002-read-operations had US2 depend on US1 while US3 was independent; 003-entity-creation had US1 and US2 fully independent of each other. Here, **US2 depends on US1** (`ValidateProject` runs `ValidateEntity`'s own per-entity checks internally, per data-model.md's flow) and **US3 depends on US2** (`Status.StructuralErrors` is `len(validation.ValidateProject(...))`, per research.md's explicit decision not to re-derive that count separately). This is called out here rather than forced into a parallel-stories narrative that wouldn't be honest about this feature's actual shape.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

Single Go module (unchanged):

```text
internal/validation/   # NEW package — Foundational (tables, Finding), US1, US2
internal/operations/   # existing package — US3 (status.go)
internal/example/       # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the one new package this feature adds.

- [X] T001 Create the `internal/validation/` directory with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: Module builds (`go build ./...`) with the new empty package before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: `Finding`/`Severity` and the two fixed reference tables (allowed lifecycle states, required parent type) are what every check `ValidateEntity` runs depends on — they must exist and be correct before US1 can be implemented meaningfully.

**⚠️ CRITICAL**: No user story implementation may begin until this phase is complete.

- [X] T002 [P] Unit tests for the fixed reference tables — every entity type's allowed lifecycle states match `docs/architecture-specification.md`'s frozen list exactly (Program/Feature: draft/active/done/cancelled; Spec: draft/ready/in_progress/validated/blocked/superseded/cancelled; Learning: candidate/promoted/dismissed; Knowledge: active); required-parent-type lookups (Feature→Program, Spec→Feature, and a "no entry" result for Program/Knowledge/Learning) — in `internal/validation/tables_test.go`.
- [X] T003 [P] Define `Severity`, `SeverityError`, and `Finding` (`Code`, `Severity`, `Path`, `Message`) in `internal/validation/findings.go`.
- [X] T004 Implement the allowed-lifecycle-states and required-parent-type tables (hardcoded constants, not configuration — research.md) in `internal/validation/tables.go`. Depends on T002.

**Checkpoint**: Foundation ready — `go test ./internal/validation/...` passes on its own (table tests only so far); User Story 1 implementation can now begin.

---

## Phase 3: User Story 1 - Validate a Single Entity (Priority: P1) 🎯 MVP

**Goal**: Given one entity's ID, run every applicable structural check and return every finding — or an empty list if there are none.

**Independent Test**: Populate a fixture project with a mix of well-formed and deliberately broken entities (missing parent, wrong-type parent, invalid status, unresolved Spec dependency, an entity with more than one problem at once) and confirm each validates to exactly the findings that apply — no other story's code required.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T006.

- [X] T005 [P] [US1] Filesystem-integration tests for `ValidateEntity` — a well-formed Spec (empty findings); a missing parent; a parent that resolves but is the wrong type; an invalid status value; a Spec whose `depends_on` names a nonexistent Spec; an entity with more than one problem simultaneously (all applicable findings present together); a not-found ID; two artifacts sharing one ID (duplicate/ambiguous finding); a syntactically malformed `rawID` (a Go error, not a Finding); a Task ID requested directly (a Go error — unsupported target) — in `internal/validation/validator_test.go`.

### Implementation for User Story 1

- [X] T006 [US1] Implement `ValidateEntity` in `internal/validation/validator.go` — parses `rawID` (mirroring `operations.Resolve`'s two-pass type-then-width approach, but via `ids.ParseAny`/`ids.Parse` directly, never importing `operations` — research.md), looks up `ids.Scan(...).Paths[number]` directly (0/`>`1/1, turning 0 and `>`1 into Findings, never errors), and for the single-match case runs the frontmatter/required-field/ID-location/status/parent/dependency checks. Depends on T004, T003, T005.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/validation/... -run TestValidateEntity` passes on its own.

---

## Phase 4: User Story 2 - Validate the Whole Project (Priority: P2)

**Goal**: Run every entity's checks across the whole project in one pass, aggregating every finding, including whole-project-only problems like duplicate IDs.

**Independent Test**: Against a fixture project mixing valid entities, several individually-broken ones, and a duplicated ID, confirm the aggregated result contains every individual problem plus the duplicate-ID finding. **Note**: this story's implementation genuinely depends on User Story 1 (`ValidateProject` runs the same per-entity logic `ValidateEntity` does, per entity found by `Scan` — see data-model.md's flow) — this is not merely a testing convenience, `ValidateProject` cannot be correctly implemented before `ValidateEntity`'s checks exist.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T008.

- [X] T007 [P] [US2] Filesystem-integration tests for `ValidateProject` — an all-clean project (empty result); several entities with different individual problems, all present in the aggregated result; two artifacts of the same type sharing one ID under different parents (a `duplicate_id` finding); two Task headings sharing one number across `tasks.md` files (a `duplicate_id` finding, proving Task's free duplicate-detection scope from research.md); a project with zero entities at all (empty result, not an error) — in `internal/validation/validator_test.go` (same file as T005; sequential, not `[P]`, since T005/T006 are already complete by the time this runs).

### Implementation for User Story 2

- [X] T008 [US2] Implement `ValidateProject` in `internal/validation/validator.go` — for each of Program/Feature/Spec/Knowledge/Learning, `ids.Scan` and run T006's per-entity checks against every found number, appending `Duplicates` as `duplicate_id` findings; for Task, `ids.Scan` and append only its `Duplicates` (no per-entity checks — research.md's Task scope decision). Depends on T006, T007.

**Checkpoint**: User Stories 1 AND 2 both pass their own tests.

---

## Phase 5: User Story 3 - Get an Aggregate Project Status Summary (Priority: P3)

**Goal**: A single call returning entity counts by type, Spec counts by lifecycle state, and the total structural-finding count.

**Independent Test**: Against a fixture project with a known mix of entities and lifecycle states (some structurally broken), confirm `Status`'s counts match a direct count and its `StructuralErrors` matches a separate `ValidateProject` call's finding count. **Note**: this story's implementation depends on User Story 2 (`Status.StructuralErrors` is literally `len(validation.ValidateProject(...))` — research.md was explicit that re-deriving this count separately would violate DRY) and reuses `operations.Inspect` (002-read-operations) for per-Spec states.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T010.

- [X] T009 [P] [US3] Filesystem-integration tests for `Status` — a project with a known number of Programs/Features/Specs (`Counts` matches exactly); Specs in several different lifecycle states (`SpecsByState` matches exactly); a project with known structural problems (`StructuralErrors` matches a direct `validation.ValidateProject` call's finding count on the same fixture); a project with zero entities (every count zero, not an error) — in `internal/operations/status_test.go`.

### Implementation for User Story 3

- [X] T010 [US3] Implement `StatusSummary` and `Status` in `internal/operations/status.go` — `ids.Scan` per type for `Counts`, `operations.Inspect` on every found Spec for `SpecsByState`, `len(validation.ValidateProject(root, cfg))` for `StructuralErrors`. Depends on T008, T009.

**Checkpoint**: All three user stories pass their own tests — the full chain (`ValidateEntity` → `ValidateProject` → `Status`) is correct end to end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T011 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation, 002-read-operations, and 003-entity-creation's packages included) and fix any findings.
- [X] T012 [P] Verify/extend package-level doc comments on `internal/validation` and the new `operations.Status`, cross-checked against `specs/004-structural-validation/contracts/validation.md`.
- [X] T013 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: validate a clean entity, validate a deliberately broken one, validate the whole project, then call `Status` and confirm its `StructuralErrors` matches.
- [X] T014 Reconcile `specs/004-structural-validation/contracts/validation.md`'s signatures and Finding-code mapping table against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T015 Full regression run: `go test ./...` across the entire module (001, 002, 003, and this feature's code) green, `go vet ./...` clean, `gofmt -l .` empty.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Stories (Phase 3–5)**: Form a **linear chain**, not independent lanes: US1 → US2 → US3. Each one's implementation is a genuine prerequisite for the next, not just a sequencing convenience (see each phase's Independent Test note above for why).
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Depends on User Story 1's implementation (T006) — `ValidateProject` runs `ValidateEntity`'s own checks per discovered entity.
- **User Story 3 (P3)**: Depends on User Story 2's implementation (T008) — `Status.StructuralErrors` calls `ValidateProject` directly.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational tables/types complete before any story's implementation.
- Given the linear chain, working sequentially (rather than attempting parallel stories) is the natural execution order for this feature.

### Parallel Opportunities

- T002 and T003 (Foundational, different files) in parallel.
- Within Polish, T011 and T012 in parallel.
- Unlike 003-entity-creation, there is no cross-story parallel-team opportunity here — the linear dependency chain means a second contributor has nothing independent to start until the first story ahead of them lands.

---

## Parallel Example: Foundational Phase

```bash
# Launch Foundational tasks together (different files):
Task: "Unit tests for lifecycle-state/parent-type tables in internal/validation/tables_test.go"
Task: "Define Finding/Severity in internal/validation/findings.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/validation/... -run TestValidateEntity` green, independently.
5. This alone already lets any Skill validate the one entity it just created or modified — the check `/create-feature`, `/create-specs`, and `/create-plan` all need per their own contracts.

### Incremental Delivery

1. Setup + Foundational → shared types/tables ready.
2. Add US1 → validate independently → single-entity validation usable (MVP).
3. Add US2 (depends on US1) → validate independently → whole-project validation usable, including duplicate-ID detection.
4. Add US3 (depends on US2) → validate independently → fast status summaries usable.
5. Polish (Phase 6), full-module regression (T015).

### Team Strategy

Given the linear US1 → US2 → US3 chain, this feature does not offer the
parallel-team opportunity 003-entity-creation's independent US1/US2 did.
One contributor working the chain in order is the natural approach;
splitting it across people would mean most of the team waiting on the
person ahead of them in the chain.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- This feature is entirely read-only (FR-014) — no task here writes anything outside test fixtures under `t.TempDir()`.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
