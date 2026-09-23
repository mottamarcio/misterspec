---

description: "Task list for Análise de Impacto e Revisão Incremental"
---

# Tasks: Análise de Impacto e Revisão Incremental

**Input**: Design documents from `/specs/042-impact-analysis-review/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/analyze-impact-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic diffing/traversal/classification logic, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US3)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. New package `internal/impact` (change-set diffing, reverse-relation propagation, classification, orchestration); additive extensions to `internal/vcs`, `internal/validation`, `internal/cli/internalcmd`. No new top-level directory beyond `internal/impact`.

---

## Phase 1: Setup

**Purpose**: Scaffold the one new package this feature adds.

- [X] T001 Create `internal/impact/doc.go` with a package doc explaining `internal/impact` computes a change-triggered impact report by diffing two Git revisions and walking reverse formal/coverage/evidence/wikilink relations already computed by `internal/operations`, `internal/validation`, and `internal/evidence` — a package layered like `internal/prepare` (plan.md "Scale/Scope").

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The Git-diff, per-Requirement-fingerprint, error-code, and Change Set/classification primitives every user story builds on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundational

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T002 [P] Add cases to `internal/vcs/vcs_test.go` (write first): `DiffNameStatus(root, from, to)` reports an added/modified/removed path correctly between two commits; with `to == ""` it diffs `from` against the working tree (an uncommitted edit shows up); a `from` that does not resolve returns a non-nil error (contracts §1). Expected to FAIL until T006.
- [X] T003 [P] Add cases to `internal/vcs/vcs_test.go` (write first): `FileAtRevision(root, path, rev)` returns a file's content and `found == true` at a commit where it exists; `found == false`, no error, at a commit before the file was added (mirrors `CommitsSinceFileAdded`'s own convention); `rev == ""` reads the working-tree file directly, reflecting an uncommitted edit (contracts §1). Expected to FAIL until T007.
- [X] T004 [P] Add cases to `internal/validation/requirements_test.go` (write first): `RequirementSections(spec, body)` returns one `artifacts.Section` per `R<N>` heading found, keyed by number, matching exactly the same numbers `parseSpecRequirements` already finds for the same `body` (regression case: the existing function's own behavior is unchanged) (contracts §2). Expected to FAIL until T008.
- [X] T005 [P] Add cases to `internal/cli/internalcmd/errors_test.go` (write first): `classify()` maps an error wrapping the new `impact.ErrRevisionNotFound` to code `"revision_not_found"`, and one wrapping `impact.ErrNotARepository` to code `"not_a_repository"`, each with its own stable exit code, distinct from any existing code (contracts §4 "Error envelope"). Expected to FAIL until T009.

### Implementation for Foundational

- [X] T006 Add `vcs.DiffEntry` and `vcs.DiffNameStatus(root, from, to string) ([]DiffEntry, error)` to `internal/vcs/vcs.go`, via `git diff --name-status <from> [<to>]` (`to == ""` omits the second argument, diffing against the working tree) (contracts §1, research.md #2) (depends on T002).
- [X] T007 Add `vcs.FileAtRevision(root, path, rev string) (content []byte, found bool, err error)` to `internal/vcs/vcs.go`, via `git show <rev>:<path>` (`rev == ""` reads the working-tree file with `os.ReadFile` instead) (contracts §1, research.md #2) (depends on T003).
- [X] T008 Add `validation.RequirementSections(spec ids.EntityID, body []byte) map[int]artifacts.Section` to `internal/validation/requirements.go`, reusing the existing `requirementHeadingPattern`/`artifacts.ParseDocument` walk `parseSpecRequirements` already performs, now also capturing each match's own `Section` — `parseSpecRequirements` itself is left unchanged (contracts §2, research.md #3) (depends on T004).
- [X] T009 Add `revision_not_found` and `not_a_repository` error codes to `internal/cli/internalcmd/errors.go`'s `classify()`, matching new sentinel errors `impact.ErrRevisionNotFound`/`impact.ErrNotARepository` (defined alongside `AnalyzeImpact` in T019) (contracts §4) (depends on T005).
- [X] T010 [P] Add `internal/impact/change_set_test.go` (write first): `BuildChangeSet` produces one whole-artifact `ChangedElement` per added/modified/removed Program/Feature/Spec/Knowledge/Learning path in the diff; a modified `spec.md` additionally produces one `ChangedElement` per Requirement number whose `RequirementSections` snapshot differs between `from`/`to` (added/changed/removed), with `RequirementNumber` set; a changed `plan.md`/`tasks.md` is attributed to its owning Spec's `ids.EntityID` (data-model.md "ChangedElement", research.md #8); a changed path outside the five types and `plan.md`/`tasks.md` increments `UnmappedCodePaths` instead of appearing in `Elements` (research.md #9). Expected to FAIL until T012.
- [X] T011 [P] Add `internal/impact/classify_test.go` (write first): every `RelationKind` (`depends_on`, `parent`, `supersedes`, `coverage`, `evidence`, `wikilink`) maps to exactly one `Classification` per the fixed table (`wikilink` → `suggested_review`, every other kind → `deterministic_invalidation`, unconditionally — no input flips this); every `(Classification, RelationKind)` pair maps to exactly one `Severity`, and calling the classification/severity functions twice with the same input always returns the same result (spec FR-010, research.md #7). Expected to FAIL until T013.

- [X] T012 [P] Implement `internal/impact/change_set.go`: the `ChangeStatus`/`ChangedElement`/`ChangeSet` types and `BuildChangeSet(root string, cfg project.Configuration, from, to, path string) (ChangeSet, error)`, using `vcs.DiffNameStatus`, `vcs.FileAtRevision`, and `validation.RequirementSections` (data-model.md "ChangedElement"/"ChangeSet", research.md #1/#3/#8/#9) (depends on T006, T007, T008, T010).
- [X] T013 [P] Implement `internal/impact/classify.go`: the `Classification`/`Severity` types and their two fixed lookup tables, keyed exactly as research.md #7 describes (data-model.md "Classification"/"Severity") (depends on T011).

**Checkpoint**: Foundation ready — Change Set construction (whole-artifact and per-Requirement) and the classification/severity tables are available; nothing downstream walks relations or reports yet.

---

## Phase 3: User Story 1 - Saber o que revisar depois de mudar um artefato (Priority: P1) 🎯 MVP

**Goal**: Given a changed Requirement or a Spec another Spec formally `depends_on`, report every Task/Spec with a known formal-dependency or Requirement-coverage relation to it, each with a specific reason — nothing reported for an artifact with no such relation.

**Independent Test**: Change a Requirement's text served by a Task with existing 041 evidence, run `analyze-impact`, and confirm that Task is reported with a reason naming the changed Requirement; confirm an unrelated artifact is not reported.

**Scope note**: This phase implements one-hop propagation over `depends_on`/`parent`/`supersedes` (formal) and Requirement→Task (`coverage`) relations only — `wikilink` (Semantic) relations are Phase 4 (User Story 2)'s own addition; a wikilink-only relation is simply not yet walked at all until then, so no premature classification decision is made about it.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T014 [P] [US1] Add `internal/impact/propagate_test.go` (write first): a `ChangedElement` for Spec `Y` with a one-hop `depends_on` backlink from Spec `X` (via a fixture using `operations.Backlinks`' own `Formal` list) produces one `PropagationHop{Relation: "depends_on"}` and one `AffectedItem` for `X`; a `ChangedElement` for Requirement `SPEC-014:R2` with a Task in the same Spec whose `Serves:` names it (via `validation.ParseTaskCoverage`) produces one `PropagationHop{Relation: "coverage"}` and one `AffectedItem` for that Task; a `ChangedElement` with zero backlinks and zero coverage produces zero `AffectedItem`s and appears in a separate `NoKnownRelationElements`-shaped result instead (data-model.md "PropagationHop"/"PropagationPath", spec FR-002/FR-008). Expected to FAIL until T017.
- [X] T015 [P] [US1] Add `internal/impact/analyze_test.go` (write first), building a real two-commit fixture Git repo: `AnalyzeImpact` reports a Task as `deterministic_invalidation` with a `Reason` naming the changed Requirement number after that Requirement's own text changes between the two commits (spec US1 Acceptance Scenario 1); reports a dependent Spec as `deterministic_invalidation` after a Spec it `depends_on` changes (Scenario 2); reports nothing for an artifact with no relation to the change (Scenario 3). Expected to FAIL until T018.
- [X] T016 [P] [US1] Add `internal/cli/internalcmd/analyze_impact_test.go` (write first): the success envelope's shape matches contracts §4's example (`analyze_impact_schema_version`, `change_set`, `affected_items`, `no_known_relation_elements` all present, arrays never `null`); an unresolvable `--from` produces an error envelope with `"code": "revision_not_found"`; running outside a Git repository produces `"code": "not_a_repository"`. Expected to FAIL until T019.

### Implementation for User Story 1

- [X] T017 [US1] Implement `internal/impact/propagate.go`: the `RelationKind`/`PropagationHop`/`PropagationPath` types and a one-hop walk that, for each whole-artifact `ChangedElement`, calls `operations.Backlinks` and emits a `PropagationHop` per `Formal` entry, and, for each per-Requirement `ChangedElement`, reads the owning Spec's `tasks.md` via `validation.ParseTaskCoverage` and emits a `coverage` `PropagationHop` per matching Task; tracks a `map[ids.EntityID]bool` visited set so no element is expanded twice (data-model.md "PropagationHop", research.md #4/#5/#6) (depends on T012, T013, T014).
- [X] T018 [US1] Implement `internal/impact/analyze.go`: the `AffectedItem`/`ImpactReport` types, `ErrRevisionNotFound`/`ErrNotARepository` sentinels, and `AnalyzeImpact(root string, cfg project.Configuration, req AnalyzeImpactRequest) (ImpactReport, error)`, orchestrating `BuildChangeSet` + `propagate` + `classify`, populating `NoKnownRelationElements` for every `ChangedElement` the walk found zero hops for, and ordering `AffectedItems` by severity then path (data-model.md "ImpactReport", contracts §3) (depends on T017, T015).
- [X] T019 [US1] Implement `internal/cli/internalcmd/analyze_impact.go`: `"internal analyze-impact"` wiring `--from`/`--to`/`--path`/`--dir`, calling `impact.AnalyzeImpact`, and rendering the JSON envelope at `analyze_impact_schema_version = 1` per contracts §4 (depends on T016, T018, T009).

**Checkpoint**: User Story 1 is independently functional — a changed Requirement or a formally-depended-upon Spec is reported with a specific reason via `misterspec internal analyze-impact`; unrelated artifacts are not.

---

## Phase 4: User Story 2 - Distinguir o que foi invalidado do que só merece revisão (Priority: P1)

**Goal**: A wikilink-only mention of a changed artifact is reported as `suggested_review`, never `deterministic_invalidation` — added to the same report User Story 1 already produces.

**Independent Test**: Change an artifact that is only referenced via wikilink by another (no formal/coverage/evidence relation), run `analyze-impact`, and confirm the referencing artifact is reported with `classification: "suggested_review"`.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T020 [P] [US2] Add cases to `internal/impact/propagate_test.go` (write first): a `ChangedElement` with a wikilink-only backlink (via `operations.Backlinks`' own `Semantic` list, including its `SourceSection`/`SourceLine` provenance from 038) produces a `PropagationHop{Relation: "wikilink"}` and an `AffectedItem` classified `suggested_review`; the same element additionally reached by a `depends_on` backlink still classifies that separate path `deterministic_invalidation` — the two classifications never blend into one for the same item (spec FR-003/FR-004, data-model.md "Classification"). Expected to FAIL until T021.
- [X] T021 [P] [US2] Add a fixture-repo case to `internal/impact/analyze_test.go` (write first): artifact `A`, which only `[[wikilinks]]` a changed artifact `B` with no other relation, is reported by `AnalyzeImpact` with `classification: "suggested_review"` (spec US2 Acceptance Scenario 1); in the same run, a Task with a formal/coverage relation to a different changed element is reported `deterministic_invalidation` (Scenario 2). Expected to FAIL until T022.

### Implementation for User Story 2

- [X] T022 [US2] Extend `internal/impact/propagate.go`'s walk to also consume `operations.Backlinks`' `Semantic` list, emitting one `wikilink` `PropagationHop` (with `SourceSection`/`SourceLine` carried through) per occurrence, alongside the existing formal/coverage hops — `classify.go`'s fixed table (already built in Foundational) is consulted unchanged, so no new classification logic is added here, only a new relation kind feeding it (research.md #7) (depends on T017, T020, T021).

**Checkpoint**: User Story 2 is independently functional — a wikilink-only relation is reported, and always as `suggested_review`; a formal/coverage relation in the same run is unaffected and still `deterministic_invalidation`.

---

## Phase 5: User Story 3 - Caminho de propagação, motivo e severidade legíveis (Priority: P2)

**Goal**: A multi-hop chain (e.g. a Task's coverage relation reached through an intermediate Spec's own formal dependency) reports its *full* path, not just the nearest hop; the same affected item reached by more than one distinct path keeps every path; a reference/dependency cycle terminates cleanly; multiple simultaneous changes in one run are consolidated into one report.

**Independent Test**: Construct a two-hop chain (A depends_on B; a Task's coverage/evidence ties it to A) and confirm the reported `AffectedItem` for the Task carries both hops, in order, from the original change; construct a wikilink cycle (A ↔ B) and confirm `analyze-impact` terminates with a bounded result.

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T023 [P] [US3] Add cases to `internal/impact/propagate_test.go` (write first): a fixture with a two-hop chain (`B` changed → `A` depends_on `B` → a Task covers a Requirement in `A`) produces one `PropagationPath` for the Task whose `Hops[0].FromID == B` and whose `Hops[len-1].ToID` is the Task's own ID, with `len(Hops) == 2`; a `ChangedElement` reached by both a `depends_on` hop and a separate `wikilink` hop to the same target keeps two distinct `PropagationPath`s in that target's `AffectedItem`, never collapsed into one (spec FR-005, data-model.md "AffectedItem.Paths"). Expected to FAIL until T025.
- [X] T024 [P] [US3] Add a fixture-repo case to `internal/impact/analyze_test.go` (write first): two artifacts wikilinking each other (`A ↔ B`); changing `A` and running `AnalyzeImpact` returns promptly with `B` appearing at most once in `AffectedItems` despite more than one path existing back to it (spec FR-006, Edge Case "ciclo de referências"); a `ChangeSet` with two simultaneously changed elements (e.g. a Requirement and its owning Spec's `plan.md`) produces one consolidated `AffectedItems` list, not two separate reports (Edge Case "duas mudanças na mesma execução"). Expected to FAIL until T026.

### Implementation for User Story 3

- [X] T025 [US3] Extend `internal/impact/propagate.go` from a one-hop walk to full transitive closure: after emitting a hop to an intermediate artifact, continue expanding that artifact's own `Backlinks`/coverage relations in turn, appending to the same `PropagationPath` rather than stopping — still bounded by the existing visited-element set (research.md #6), and still accumulating every distinct path to a given item rather than keeping only the first found (data-model.md "PropagationPath", "AffectedItem.Paths") (depends on T022, T023).
- [X] T026 [US3] Update `internal/impact/analyze.go` to merge `AffectedItem`s reached from more than one `ChangedElement` in the same `ChangeSet` into a single row per item (by ID), keeping every distinct `PropagationPath` from every originating change (spec Edge Case "duas mudanças na mesma execução") (depends on T018, T025, T024).

**Checkpoint**: User Story 3 is independently functional — full multi-hop propagation paths are reported, multiple distinct paths to the same item are preserved, cycles terminate cleanly, and simultaneous changes are consolidated into one report.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification across all three stories together.

- [X] T027 [P] Add a case to `internal/impact/analyze_test.go`: a diff that includes a `.go` source file alongside a Spec change reports `ChangeSet.UnmappedCodePaths >= 1` and still reports the Spec-side impact normally — the code path is declared, never silently dropped nor treated as "no impact" (spec FR-007, research.md #9).
- [X] T028 [P] Run `gofmt -l` and `go vet ./...` across every new/changed file in this feature and fix any findings.
- [X] T029 Manually run all 6 scenarios in `quickstart.md` against a built `misterspec` binary in a real fixture repository, confirming each "Expected" outcome.
- [X] T030 [P] Final pass on `internal/impact/doc.go` and each new exported symbol's doc comment, confirming they accurately describe the shipped behavior (formal + coverage + evidence + wikilink, multi-hop, consolidated) rather than the phase-by-phase scope notes used during build-out.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion. MVP — delivers formal + coverage propagation.
- **User Story 2 (Phase 4)**: Depends on Foundational completion; in practice extends `propagate.go`/`propagate_test.go` from Phase 3, so implement after US1 lands (T017) even though it introduces no new dependency on US1's own report shape.
- **User Story 3 (Phase 5)**: Depends on Foundational completion; extends the same `propagate.go`/`analyze.go` files US1 and US2 touch, so implement after both (sequential by file, not by feature dependency).
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other stories — a wikilink-only relation simply isn't walked yet in this phase.
- **User Story 2 (P1)**: Independently testable once implemented, but its implementation task (T022) touches the same file US1's T017 creates — sequence by file, not by required feature order.
- **User Story 3 (P2)**: Independently testable once implemented; its implementation tasks (T025, T026) touch the same files US1/US2 already modified — sequence last.

### Within Each User Story

- Tests written and failing before implementation (Principle V).
- Types/data structures before the logic that populates them.
- `internal/impact` internals (`propagate.go`, `analyze.go`) before the CLI command that exposes them.

### Parallel Opportunities

- T002-T005 (all Foundational tests) can run in parallel — different files.
- T010-T011 (Change Set / classify tests) can run in parallel once T006-T009 land.
- T012-T013 (Change Set / classify implementations) can run in parallel — different files, no shared dependency between them.
- T014-T016 (US1 tests) can run in parallel.
- T020-T021 (US2 tests) can run in parallel.
- T023-T024 (US3 tests) can run in parallel.
- T027-T028, T030 (Polish) can run in parallel.

---

## Parallel Example: Foundational

```bash
# Launch all Foundational tests together:
Task: "Add DiffNameStatus/FileAtRevision cases to internal/vcs/vcs_test.go"
Task: "Add RequirementSections cases to internal/validation/requirements_test.go"
Task: "Add new error-code cases to internal/cli/internalcmd/errors_test.go"

# Once T006-T009 land, launch the two Change Set / classify test files together:
Task: "Add internal/impact/change_set_test.go"
Task: "Add internal/impact/classify_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run `quickstart.md` scenarios 1, 3, and 4 against a real fixture repo.
5. `misterspec internal analyze-impact` is now genuinely useful for the most common case (Requirement/dependency changes) even before wikilink suggestions or multi-hop chaining exist.

### Incremental Delivery

1. Setup + Foundational → Change Set and classification primitives ready.
2. Add User Story 1 → one-hop formal/coverage propagation → validate independently (MVP!).
3. Add User Story 2 → wikilink suggestions added to the same report → validate independently (quickstart scenario 2).
4. Add User Story 3 → multi-hop chains, multi-path items, cycle termination, consolidated multi-change reports → validate independently (quickstart scenarios 3 and 5).
5. Polish → unmapped-code-path declaration (quickstart scenario 6), full quickstart re-run, doc-comment accuracy pass.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- US2 and US3 both extend files US1 creates (`propagate.go`, `analyze.go`) — they are independently *testable* (each adds its own acceptance-scenario coverage) even though they are not independently *deployable* as separate binaries; this mirrors `041-task-evidence-fingerprint`'s own precedent of a later story extending an earlier story's own file with a documented scope note.
- Verify each story's tests fail before implementing that story's tasks.
- Commit after each task or logical group.
- Stop at any checkpoint to validate a story independently.
