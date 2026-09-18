---

description: "Task list template for feature implementation"
---

# Tasks: Validação de cobertura de requisitos e dependências do SDD

**Input**: Design documents from `/specs/032-requirement-coverage-dependency-validation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/requirement-coverage-and-dependency-validation.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story and MUST be written and confirmed failing before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths are included in every description

## Path Conventions

Single Go module/CLI project (plan.md "Structure Decision") — all paths are relative to the repository root, inside the existing `internal/validation` package (two new files: `requirements.go`, `dependency_graph.go`). No new top-level directory.

---

## Phase 1: Setup

**Purpose**: Establish a clean, verified starting point. No new dependencies, no new project structure.

- [X] T001 Run `go build ./...` and `go test ./...` from the repository root and confirm a clean baseline (no pre-existing failures) before making any change, so any later red test is attributable to this feature's work.

**Checkpoint**: Baseline confirmed green — safe to start Foundational work.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The new stable Finding codes (contracts §1) both User Story 1 and User Story 2 emit into the shared `internal/validation/findings.go` file.

**⚠️ CRITICAL**: No user story task may start until this phase is complete.

- [X] T002 Add the six new Finding code constants to `internal/validation/findings.go`, matching contracts/requirement-coverage-and-dependency-validation.md §1 exactly: `CodeDuplicateRequirementID = "duplicate_requirement_id"`, `CodeUncoveredRequirement = "uncovered_requirement"`, `CodeUnknownRequirementReference = "unknown_requirement_reference"`, `CodeCrossSpecRequirementReference = "cross_spec_requirement_reference"`, `CodeTaskWithoutRequirement = "task_without_requirement"`, `CodeDependencyCycle = "dependency_cycle"`, `CodePhaseGateBlocked = "phase_gate_blocked"` — following the existing `Code* = "..."` declaration style already in that file (no test needed for these constants alone, matching this file's existing convention of having no dedicated `findings_test.go`).

**Checkpoint**: New Finding codes exist — User Stories 1, 2, and 3 can now proceed in priority order.

---

## Phase 3: User Story 1 - Detectar cobertura de requisitos por tarefas (Priority: P1) 🎯 MVP

**Goal**: Every `### R<N>` requirement in a Spec is checked against every `Serves:` reference in that Spec's own `tasks.md` — an uncovered requirement, an unknown/malformed reference, a cross-Spec reference, and a Task with no `Serves:` line at all are all detected and reported with a stable Finding code.

**Independent Test**: A Spec with `R1`/`R2`, a Task serving only `R1` → `R2` reported uncovered; a Task serving a nonexistent or cross-Spec requirement → reported invalid (quickstart.md Scenarios 1–4).

### Tests for User Story 1 ⚠️

- [X] T003 [P] [US1] Write failing unit tests in `internal/validation/requirements_test.go` for Requirement-heading parsing (`data-model.md` "SpecRequirements"): a body with `### R1` and `### R2` headings yields `Numbers: [1, 2]`; a repeated `### R1` heading yields `Duplicates: [1]`; headings at other levels (`## R3`, `#### R4`) still match; a non-`R<digits>` heading (e.g. `### Retry policy`) is not matched.
- [X] T004 [P] [US1] Write failing unit tests in `internal/validation/requirements_test.go` for `parseRequirementRef` (`data-model.md` "RequirementRef", contracts §1): valid `"SPEC-014:R3"` parses to `{Spec: SPEC-014, Number: 3}`; malformed input (missing `R`, non-numeric suffix, missing Spec half, invalid Spec-ID syntax) is a parse error.
- [X] T005 [P] [US1] Write failing unit tests in `internal/validation/requirements_test.go` for Task `Serves:` parsing (`data-model.md` "TaskCoverage", research.md Decision 3): a single `Serves: SPEC-014:R1` line yields one reference; `Serves: SPEC-014:R1, SPEC-014:R2` yields two; two separate `Serves:` lines in the same Task body are unioned; a malformed entry lands in `MalformedReferences`, not silently dropped; a Task body with no `Serves:` line yields empty `References`.
- [X] T006 [P] [US1] Write failing integration test in `internal/validation/validator_test.go`: a Spec with `R1`/`R2` in `spec.md` and a `tasks.md` Task serving only `SPEC-001:R1` → `ValidateProject` reports `CodeUncoveredRequirement` for `R2` and nothing for `R1` (spec Acceptance Scenario 1).
- [X] T007 [P] [US1] Write failing integration test in `internal/validation/validator_test.go`: a Task serving `SPEC-001:R9` (nonexistent) → `ValidateProject` reports `CodeUnknownRequirementReference` naming the Task and `R9` (spec Acceptance Scenario 2).
- [X] T008 [P] [US1] Write failing integration test in `internal/validation/validator_test.go`: a Task owned by `SPEC-001` serving `SPEC-002:R1` → `ValidateProject` reports `CodeCrossSpecRequirementReference`, and `SPEC-002:R1` still shows as uncovered from `SPEC-002`'s own perspective (spec Acceptance Scenario 3).
- [X] T009 [P] [US1] Write failing integration test in `internal/validation/validator_test.go`: a Task with no `Serves:` line at all → `ValidateProject` reports `CodeTaskWithoutRequirement` naming that Task (spec Acceptance Scenario 4).
- [X] T010 [P] [US1] Write failing integration test in `internal/validation/validator_test.go`: a Spec where every requirement is covered and every reference is valid and same-Spec → `ValidateProject` reports zero coverage-related findings (spec Acceptance Scenario 5).
- [X] T011 [P] [US1] Write failing integration test in `internal/validation/validator_test.go`: two `### R1` headings in the same Spec's `spec.md` → `ValidateProject` reports `CodeDuplicateRequirementID` (spec FR-002, Edge Cases).

### Implementation for User Story 1

- [X] T012 [US1] Implement Requirement-heading parsing in `internal/validation/requirements.go`, using `artifacts.ReadBody` + `artifacts.ParseDocument` and a regex matching `^R(\d+)\b` against each `Section.Heading` (research.md Decision 1–2), to make T003 pass.
- [X] T013 [US1] Implement `RequirementRef` and `parseRequirementRef(raw string, cfg project.Configuration) (RequirementRef, error)` in `internal/validation/requirements.go` per data-model.md "RequirementRef" — the Spec half via `ids.Parse(ids.Spec, ..., cfg.IDWidth)`, the Requirement half via a dedicated unpadded `R<digits>` parser (research.md Decision 1) — to make T004 pass.
- [X] T014 [US1] Implement Task `Serves:` line parsing in `internal/validation/requirements.go`, scanning a Task's own `Section.Body` (from `tasks.md`'s `ParseDocument`) for `^Serves:\s*(.+)$` lines, splitting each on commas, and unioning across repeated lines (research.md Decision 3), to make T005 pass (depends on T013).
- [X] T015 [US1] Implement `RequirementCoverageReport` computation in `internal/validation/requirements.go`, joining one Spec's `SpecRequirements` with every Task's `TaskCoverage` from its own `tasks.md`, classifying each reference as valid / `unknown_requirement_reference` / `cross_spec_requirement_reference` per data-model.md "RequirementCoverageReport" (depends on T012, T014).
- [X] T016 [US1] Wire `RequirementCoverageReport` into `validation.ValidateProject` (`internal/validation/validator.go`): for every Spec discovered via the existing `projectEntityTypes` loop, parse its `spec.md` and (if present) `tasks.md`, and emit `CodeDuplicateRequirementID` / `CodeUncoveredRequirement` / `CodeUnknownRequirementReference` / `CodeCrossSpecRequirementReference` / `CodeTaskWithoutRequirement` Findings per contracts §2 — a Spec with no `tasks.md` yet MUST NOT be treated as an error, only as "no coverage yet" — to make T006–T011 pass (depends on T015).
- [X] T017 [US1] Wire the same per-Spec coverage checks into `validation.ValidateEntity` for a `Spec`-typed target (`internal/validation/validator.go`), mirroring 031-canonical-task-identity's existing per-Spec Task-duplicate wiring pattern, satisfying spec FR-013 (depends on T016).

**Checkpoint**: User Story 1 is independently functional and testable — requirement coverage gaps and broken/cross-Spec references are now caught deterministically.

---

## Phase 4: User Story 2 - Detectar ciclos no grafo de dependências entre Specs (Priority: P1)

**Goal**: The project-wide `depends_on` graph between Specs is built once per validation run and every cycle — including a Spec depending on itself — is detected and reported with its full path, without halting validation of the rest of the project.

**Independent Test**: Three Specs in a cycle (`A → B → C → A`) → one finding naming the full path; a self-referencing Spec → a degenerate cycle finding; an acyclic graph, however deep → zero findings (quickstart.md Scenarios 5–7).

### Tests for User Story 2 ⚠️

- [X] T018 [P] [US2] Write failing unit tests in `internal/validation/dependency_graph_test.go` for cycle detection over an in-memory `DependencyGraph` (data-model.md "DependencyGraph"/"DependencyCycle", research.md Decision 5): a 3-node cycle returns one `DependencyCycle` with the full ordered `Path`; a self-referencing node returns a degenerate 2-element `Path` (`[SPEC-001, SPEC-001]`); an acyclic graph (including a deep, branching one) returns zero cycles; two independent cycles in the same graph are both detected.
- [X] T019 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: three Specs with `depends_on` forming `SPEC-001 → SPEC-002 → SPEC-003 → SPEC-001` → `ValidateProject` reports exactly one `CodeDependencyCycle` finding naming the full path, and an unrelated fourth Spec still validates normally (spec Acceptance Scenario 1).
- [X] T020 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: a Spec whose own `depends_on` names itself → `ValidateProject` reports `CodeDependencyCycle`, distinct from `CodeUnresolvedDependency` (spec Acceptance Scenario 2).
- [X] T021 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: a large acyclic, branching `depends_on` graph → `ValidateProject` reports zero `CodeDependencyCycle` findings (spec Acceptance Scenario 3).
- [X] T022 [P] [US2] Write failing integration test in `internal/validation/validator_test.go`: with the Scenario-1 cycle present, `ValidateEntity(root, cfg, "SPEC-002")` (a Spec participating in the cycle) reports the same cycle `ValidateProject` would; a Spec not on any cycle produces no `CodeDependencyCycle` finding when validated alone (spec Acceptance Scenario 4, contracts §3).

### Implementation for User Story 2

- [X] T023 [US2] Implement `DependencyGraph` construction in `internal/validation/dependency_graph.go`: every Spec found via `ids.Scan(root, cfg, ids.Spec)` as a node, edges from each Spec's own `Metadata.DependsOn` filtered to `Type == ids.Spec` (data-model.md "DependencyGraph"), to support T018.
- [X] T024 [US2] Implement cycle detection in `internal/validation/dependency_graph.go` — a DFS/three-color (white/gray/black) walk over `DependencyGraph.Edges`, returning `[]DependencyCycle` with the full ordered `Path` per cycle, treating a self-edge as the same algorithm's trivial case (research.md Decision 5), to make T018 pass (depends on T023).
- [X] T025 [US2] Wire cycle detection into `validation.ValidateProject` (`internal/validation/validator.go`): build the `DependencyGraph` once per call (not once per Spec) and emit one `CodeDependencyCycle` Finding per distinct cycle per contracts §2 step 5, to make T019–T021 pass (depends on T024).
- [X] T026 [US2] Wire the same project-wide graph/cycle check into `validation.ValidateEntity` for a `Spec`-typed target (`internal/validation/validator.go`) per contracts §3 — the named Spec's participation is checked against the full project graph, mirroring `checkDependencyList`'s existing full-project-scan behavior — to make T022 pass (depends on T025).

**Checkpoint**: User Stories 1 AND 2 both work independently — coverage gaps and dependency cycles are both caught deterministically, without either interfering with the other.

---

## Phase 5: User Story 3 - Aplicar regras de validação sensíveis à fase da Spec (Priority: P2)

**Goal**: A Spec in `draft` is exempt from the coverage phase gate; any Spec past `draft` with an uncovered requirement or a Task without a requirement is reported as blocked from being ready for implementation.

**Independent Test**: A `draft` Spec with no `tasks.md` at all is not blocked; the same Spec moved to `status: ready` without coverage is blocked (quickstart.md Scenarios 8–9).

### Tests for User Story 3 ⚠️

- [X] T027 [P] [US3] Write failing integration test in `internal/validation/validator_test.go`: a Spec with `status: draft` and no `tasks.md` at all → `ValidateProject` reports no `CodePhaseGateBlocked` finding and does not error on the missing `tasks.md` (spec Acceptance Scenario 1, Edge Cases).
- [X] T028 [P] [US3] Write failing integration test in `internal/validation/validator_test.go`: the same Spec with `status: ready` and still no `tasks.md` → `ValidateProject` reports `CodePhaseGateBlocked`, alongside the underlying `CodeUncoveredRequirement`/`CodeTaskWithoutRequirement` findings it escalates (spec Acceptance Scenario 2).
- [X] T029 [P] [US3] Write failing integration test in `internal/validation/validator_test.go`: a Spec with `status: ready` and full requirement coverage → `ValidateProject` reports no `CodePhaseGateBlocked` finding (spec Acceptance Scenario 3).

### Implementation for User Story 3

- [X] T030 [US3] Implement `SpecPhaseGate` computation in `internal/validation/requirements.go`: `Exempt = (Status == "draft")`; `Blocked = !Exempt && (len(UncoveredRequirements) > 0 || len(TasksWithoutCoverage) > 0)` per research.md Decision 6 and data-model.md "SpecPhaseGate" (depends on T015).
- [X] T031 [US3] Wire `CodePhaseGateBlocked` into `validation.ValidateProject` and `validation.ValidateEntity` (`internal/validation/validator.go`), emitted in addition to (never instead of) the underlying coverage Findings, per contracts §1's table, to make T027–T029 pass (depends on T030, T016, T017).

**Checkpoint**: All three user stories are independently functional — requirement coverage, dependency cycles, and phase-aware gating are all enforced deterministically.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the whole feature end-to-end and confirm the Skills' already-documented promises now hold.

- [X] T032 [P] Re-read `kit/skills/mister-tasks/SKILL.md` (specifically its `internal validate SPEC-###` promise at lines 135–136) and `kit/skills/mister-analyze/SKILL.md`'s `Serves:` wording, and confirm both now match the implemented behavior exactly; update wording only if a genuine mismatch is found.
- [X] T033 [P] Review `docs/architecture-specification.md` for wording about requirement coverage, `Serves` references, or dependency-cycle detection that still describes them as unimplemented or absent, and correct it to describe the now-implemented checks (mirrors 031-canonical-task-identity's equivalent polish task).
- [X] T034 Manually execute every scenario in `quickstart.md` (1–10) against a disposable temp project built from the finished binary, and confirm each documented "Pass condition" holds.
- [X] T035 Run `go build ./...`, `go test ./...`, and `go vet ./...` from the repository root and confirm a fully green result across the whole module, not just the packages touched by this feature.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup (T001). BLOCKS every user story (T003–T031).
- **User Story 1 (Phase 3)**: Depends on Foundational (T002) only.
- **User Story 2 (Phase 4)**: Depends on Foundational (T002) only. Independent of User Story 1 — touches a different new file (`dependency_graph.go` vs. `requirements.go`) and a disjoint set of Finding codes.
- **User Story 3 (Phase 5)**: Depends on User Story 1's `RequirementCoverageReport` (T015) and its `ValidateProject`/`ValidateEntity` wiring (T016, T017) — the phase gate escalates User Story 1's own coverage facts, so it cannot be implemented before them (spec.md "Why this priority" for User Story 3 states this explicitly). Independent of User Story 2.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on US2 or US3.
- **User Story 2 (P1)**: No dependency on US1 or US3 — shares only the Phase 2 Finding codes.
- **User Story 3 (P2)**: Depends on User Story 1 (reuses `RequirementCoverageReport`); independent of User Story 2.

User Stories 1 and 2 touch disjoint new files (`requirements.go` vs. `dependency_graph.go`) and can proceed in parallel if staffed; User Story 3 must follow User Story 1 (though it can proceed in parallel with User Story 2).

### Within Each User Story

- Tests are written and confirmed failing before their corresponding implementation task (Principle V).
- Pure parsing (T012–T014, T023–T024) before the aggregate/report computation that consumes it (T015, T024 itself is the aggregate for US2).
- Aggregate computation before wiring into `ValidateProject` (T016, T025) before wiring into `ValidateEntity` (T017, T026).
- User Story 3's `SpecPhaseGate` computation (T030) before its wiring (T031).

### Parallel Opportunities

- T003, T004, T005 (US1 foundational parsing tests, independent functions in the same file) — mark `[P]` only if each is its own test function with no shared fixture state.
- T006–T011 (US1 integration tests) — independent fixtures within `validator_test.go`; run in parallel.
- T018 (US2 unit tests) can run in parallel with any US1 task — disjoint file.
- T019–T022 (US2 integration tests) — independent fixtures within `validator_test.go`; run in parallel with each other, and with US1's own integration tests once both are past their own Foundational dependency.
- T027–T029 (US3 integration tests) — independent fixtures; run in parallel with each other, but only after T015–T017 (US1) land.
- Once Phase 2 (Foundational) is complete, User Story 1 and User Story 2 implementation work can proceed in parallel by different contributors, since they touch disjoint new files; User Story 3 starts once User Story 1's T015–T017 land.

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together once Phase 2 is complete:
Task: "Unit tests for Requirement-heading parsing in internal/validation/requirements_test.go"
Task: "Unit tests for parseRequirementRef in internal/validation/requirements_test.go"
Task: "Unit tests for Task Serves: parsing in internal/validation/requirements_test.go"
Task: "Integration test: uncovered requirement in internal/validation/validator_test.go"
Task: "Integration test: unknown requirement reference in internal/validation/validator_test.go"
Task: "Integration test: cross-Spec requirement reference in internal/validation/validator_test.go"
Task: "Integration test: task without requirement in internal/validation/validator_test.go"
Task: "Integration test: fully covered Spec produces no findings in internal/validation/validator_test.go"
Task: "Integration test: duplicate requirement ID in internal/validation/validator_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001).
2. Complete Phase 2: Foundational (T002) — CRITICAL, blocks all stories.
3. Complete Phase 3: User Story 1 (T003–T017).
4. **STOP and VALIDATE**: run quickstart.md Scenarios 1–4 against a temp project.
5. This alone already closes `mister-tasks/SKILL.md`'s broken promise (requirement coverage checking) and is independently shippable.

### Incremental Delivery

1. Setup + Foundational → new Finding codes ready.
2. Add User Story 1 → requirement coverage/reference checks → validate independently (MVP).
3. Add User Story 2 → dependency-cycle detection → validate independently.
4. Add User Story 3 → phase-aware gating (builds on US1) → validate independently.
5. Polish → confirm Skills/docs match code, run full quickstart, whole-repo test gate.

### Parallel Team Strategy

1. One contributor completes Setup + Foundational (T001–T002) — the shared dependency every story needs.
2. Once Phase 2 lands:
   - Contributor A: User Story 1 (`internal/validation/requirements.go` + its `validator.go` wiring).
   - Contributor B: User Story 2 (`internal/validation/dependency_graph.go` + its `validator.go` wiring).
   - Contributor C starts User Story 3 once Contributor A's T015–T017 land.
3. Stories integrate independently since US1 and US2 touch disjoint new files; Polish (Phase 6) runs once all three land.

---

## Notes

- `[P]` tasks touch different files or independent fixtures within a shared test file — no shared mutable state.
- `[Story]` labels map every Phase 3+ task to its spec.md user story for traceability.
- Every user story is independently completable and testable, except User Story 3's explicit, spec-documented dependency on User Story 1.
- Confirm each test fails before writing its implementation (Principle V, NON-NEGOTIABLE).
- Avoid: vague tasks, same-file conflicts inside a `[P]` group, and cross-story dependencies beyond the one US3→US1 dependency spec.md itself documents.
