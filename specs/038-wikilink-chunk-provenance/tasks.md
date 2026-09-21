---

description: "Task list for Wikilinks with Chunk-Level Provenance (038-wikilink-chunk-provenance)"
---

# Tasks: Wikilinks with Chunk-Level Provenance

**Input**: Design documents from `/specs/038-wikilink-chunk-provenance/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/wikilink-provenance-contract.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story and are written before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (US1 = explain why a reference appeared, US2 = section-based preference scoring, US3 = bounded/non-redundant retrieval) to enable independent implementation and testing of each.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single Go project. No new package. Changes land in `internal/artifacts`,
`internal/operations`, `internal/context` (+ `internal/context/index`),
and `internal/cli/internalcmd`.

---

## Phase 1: Setup

**Purpose**: None required — this feature extends existing packages
with no new project scaffolding, dependencies, or directory structure
(plan.md Structure Decision: no new package layer).

- [X] T001 Confirm the branch's baseline builds and tests pass before any change: `go build ./...` and `go test ./...` from repo root, both clean.

**Checkpoint**: Baseline confirmed green; safe to start Foundational work.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The `ReferenceOccurrence` correlation helper and its
threading into `ReferenceEntry`/`BacklinkEntry` are used by every user
story — US1 surfaces it, US2 scores against it, and US3's regression
tests exercise it under cycles/hubs. This phase must land first.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 [P] Add `ReferenceOccurrence` type and `OccurrenceFor` helper in `internal/artifacts/occurrence.go` (contract §1): `type ReferenceOccurrence struct { SourceSection string; SourceLine int }`; `func OccurrenceFor(link WikiLink, chunks []Chunk) ReferenceOccurrence` returns the enclosing `Chunk`'s own `Heading` for the chunk whose `StartLine <= link.Line <= EndLine` (data-model.md Validation rules: "`SourceSection` MAY be empty — a wikilink before any heading is still a valid occurrence"); returns the zero value, never an error, when no chunk contains `link.Line`.
- [X] T003 [P] [Foundational] Unit tests for T002 in `internal/artifacts/occurrence_test.go`: a link whose line falls inside a chunk's `[StartLine, EndLine]` range returns that chunk's `Heading`; a link before the first heading returns `ReferenceOccurrence{}` (empty `SourceSection`, zero `SourceLine` is never asserted here — `SourceLine` always equals `link.Line`); two links on different lines within the same chunk both resolve to that chunk's same `Heading`, each keeping its own distinct `SourceLine`.
- [X] T004 Extend `operations.ReferenceEntry` in `internal/operations/references.go` with `SourcePath`, `SourceSection string`, `SourceLine int` (data-model.md "ReferenceEntry / BacklinkEntry (extended)"): `SourcePath` is always populated — the queried target's own already-known `result.Location.Path` — for every entry, formal or semantic; `SourceSection`/`SourceLine` are populated (via `artifacts.OccurrenceFor`, using the same `chunks := artifacts.Chunks(...)` this function already needs to compute once) only for a semantic (`Relation == "wikilink"`) entry, left as `""`/`0` for a formal one.
- [X] T005 [P] Extend `operations.BacklinkEntry` in `internal/operations/backlinks.go` with the same three fields: `SourcePath` is always populated from `filePath` (already in scope in the existing scan loop, per data-model.md's own note that this is "the only way the caller learns that Source's own path without a separate lookup"); `SourceSection`/`SourceLine` populated via `artifacts.OccurrenceFor` only for the semantic (wikilink) backlink entries this function already builds — zero-valued for formal ones.
- [X] T006 [P] [Foundational] Unit tests for T004 in `internal/operations/references_test.go`: a semantic (wikilink) `ReferenceEntry` has non-empty `SourceSection` matching the fixture's own known heading and `SourceLine` matching the fixture's own known line; a formal `ReferenceEntry` (`parent`/`depends_on`/`supersedes`) has `SourcePath` populated but `SourceSection: ""`, `SourceLine: 0`; two wikilinks to the same target from two different sections of the source produce two distinct `ReferenceEntry` values in `Semantic`, each with its own `SourceSection` (spec FR-004).
- [X] T007 [P] [Foundational] Unit tests for T005 in `internal/operations/backlinks_test.go`: mirrors T006 from the incoming side — a semantic `BacklinkEntry` names the correct `SourcePath`/`SourceSection`/`SourceLine` of the artifact that referenced the queried target; a formal `BacklinkEntry` has `SourcePath` populated, `SourceSection`/`SourceLine` zero-valued.
- [X] T008 Extend `contextengine.Reason` in `internal/context/result.go` with `SourcePath`, `SourceSection string`, `SourceLine int` (data-model.md "Reason (extended)"): "A `Reason` with `Relation != "wikilink"` carries these three fields empty/zero — never populated, never implying a wikilink occurrence exists where none does."
- [X] T009 Change `chunkArtifact`'s signature in `internal/context/collector.go` from `(root, relPath string, tier Tier, relation string)` to accept a full `Reason` template (e.g. `(root, relPath string, reason Reason)`), attaching that same `Reason` (already carrying `Tier`/`Relation`/`SourcePath`/`SourceSection`/`SourceLine`) to every resulting `Candidate`; update the two direct call sites (Tier 0 Constitution, Tier 1 target) to pass a `Reason` with only `Tier`/`Relation` set (occurrence fields left zero — these are not reference-derived).
- [X] T010 Change `connectedCandidates`'s `reasonOf` callback signature in `internal/context/collector.go` from `func(E) (Tier, string)` to `func(E) Reason`, updating every call site (outgoing formal/semantic in `Collect`, incoming formal/semantic backlinks in `Collect`, both formal/semantic in `secondHopCandidates`) to build a full `Reason` from each entry's own new `SourcePath`/`SourceSection`/`SourceLine` fields (T004/T005) — Tier and `Relation` string unchanged from today's behavior at each call site (e.g. incoming backlinks still label `Relation: "backlink"` regardless of the underlying entry's own `"wikilink"`/formal relation, exactly as today).
- [X] T011 [Foundational] Unit tests for T009/T010 in `internal/context/collector_test.go`: a `Candidate` produced from an outgoing wikilink reference carries a `Reason` with non-empty `SourcePath`/`SourceSection`/`SourceLine`; a `Candidate` produced from `TierMandatory`/`TierStructural` (constitution/target/formal `depends_on`) carries a `Reason` with those three fields zero-valued; a second-hop `Candidate` also carries correct occurrence data from its own originating entry.
- [X] T012 Widen the `links` table in `internal/context/index/schema.go`: add nullable `source_section TEXT` and `source_line INTEGER` columns to the `CREATE TABLE links` statement (data-model.md "`links` table (widened)"); bump `schemaVersion` from `1` to `2` (contract §6 — "the existing version-mismatch rebuild path... recreates the schema automatically, no manual migration").
- [X] T013 Update `insertLink`/`indexLinks` in `internal/context/index/sync.go` to populate the two new columns from `operations.ReferenceEntry`/`BacklinkEntry`'s own new `SourceSection`/`SourceLine` fields (T004/T005) — `insertLink`'s signature grows to accept them; `NULL` for a formal relationship (matching data-model.md: "the two new columns are the only nullable ones, reflecting that a formal relationship legitimately has no line-level origin").
- [X] T014 [Foundational] Unit/integration tests for T012/T013 in `internal/context/index/schema_test.go` and `sync_test.go`: a fresh index built under the new schema has `source_section`/`source_line` populated for a wikilink-derived `links` row and `NULL` for a formal one; an existing database file carrying the old `schemaVersion` (`1`) is transparently rebuilt on the next `Open`/`Sync` call, after which the new columns are present and correctly populated (contract §6, spec FR-010).

**Checkpoint**: `ReferenceOccurrence` correlation, extended `ReferenceEntry`/`BacklinkEntry`/`Reason`, and the widened index schema are all in place and tested. User story work can now begin.

---

## Phase 3: User Story 1 - Explain Exactly Why a Referenced Artifact Appeared (Priority: P1) 🎯 MVP

**Goal**: A retrieved Context Pack item, and the lower-level `references`/`backlinks` commands, can name the exact artifact/section/line that justified including something — not just "referenced somewhere."

**Independent Test**: Request a Context Pack for a target another artifact references from a specific, identifiable section, and confirm the returned explanation names that exact referencing artifact, section, and location (quickstart.md §1-§3).

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T015 [P] [US1] Golden/contract test in `internal/cli/internalcmd/references_test.go`: `internal references <id>` response includes `source_path`/`source_section`/`source_line` on every entry (contract §3); a formal entry shows `source_section: ""`, `source_line: 0` but a non-empty `source_path`.
- [X] T016 [P] [US1] Golden/contract test in `internal/cli/internalcmd/backlinks_test.go`: mirrors T015 for `internal backlinks <id>`, asserting the fields are populated from the referencing (source) artifact's own perspective.
- [X] T017 [P] [US1] Golden/contract test in `internal/cli/internalcmd/context_test.go`: `internal context <id>` without `--provenance` returns `schema_version: 4` with no `provenance` field on any item (contract §4 — default output additive-but-unchanged besides the version number); `internal context <id> --provenance` adds a `provenance` array only to items whose `reasons` includes `"wikilink"`, each entry shaped `{source_path, source_section, source_line}`; an item with no wikilink-derived reason has no `provenance` key at all (not an empty array).
- [X] T018 [P] [US1] Integration test in `internal/cli/internalcmd/context_test.go` (or a fixture-based test alongside it): given a target referenced from two different sections of the same source artifact, `--provenance` reports two distinct `provenance` entries (or two entries split across the relevant items), each with its own `source_section` (spec FR-004, Acceptance Scenario 2).

### Implementation for User Story 1

- [X] T019 [P] [US1] Update `renderReferenceEntries` in `internal/cli/internalcmd/references.go` to include `source_path`/`source_section`/`source_line` per entry (contract §3).
- [X] T020 [P] [US1] Update the equivalent backlink-rendering helper in `internal/cli/internalcmd/backlinks.go` to include the same three fields, from `BacklinkEntry`'s own new data.
- [X] T021 [US1] Add the `--provenance` boolean flag to `NewContextCmd` in `internal/cli/internalcmd/context.go`; when set, attach a `provenance` array (built from each item's own `Reasons` whose `Relation == "wikilink"`, using their `SourcePath`/`SourceSection`/`SourceLine`) to `renderContextItems`/`renderPackageItems`'s per-item output — only for items that actually have a wikilink-derived reason, never an empty array standing in for "not applicable" (contract §4).
- [X] T022 [US1] Bump `contextSchemaVersion` from `3` to `4` in `internal/cli/internalcmd/context.go` (contract §4 — version bumps for the whole feature's contract growth, independent of whether `--provenance` was actually passed on a given call).
- [X] T023 [US1] Run `go test ./internal/cli/internalcmd/... ./internal/context/... ./internal/operations/... ./internal/artifacts/...` and confirm T015-T018 now pass; run quickstart.md §1-§3 manually against a built binary and confirm the documented `Expected` outcomes.

**Checkpoint**: `references`/`backlinks`/`context --provenance` all report exact reference-occurrence provenance — the MVP for this feature, usable and testable independently of US2/US3.

---

## Phase 4: User Story 2 - Prefer References Made Where They Matter (Priority: P2)

**Goal**: An explicit, off-by-default `--prefer-section` capability exists that scores a wikilink-derived candidate higher when its occurrence's `SourceSection` is a requirements-bearing heading — never applied unless the caller passes the flag, never promoted to default without a future, evidence-recorded Spec.

**Independent Test**: Request a Context Pack for a target referenced both from a Requirements section and from an unrelated section; confirm ordering is unchanged without `--prefer-section` and confirm the Requirements-section reference is preferred with it (quickstart.md §4).

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T024 [P] [US2] Unit test for `Request.PreferSection`/`ScoreComponents.SectionPreference` in `internal/context/rank_test.go`: with `PreferSection: false` (the zero value), a candidate whose wikilink `Reason.SourceSection` is `"Requirements"` scores identically to one with an unrelated `SourceSection`, all else equal (data-model.md "Preference Score": "Absent/false → byte-identical ordering to today's"); with `PreferSection: true`, the `"Requirements"`-sourced candidate's `Total` is strictly higher, driven by a non-zero `SectionPreference` component.
- [X] T025 [P] [US2] Unit test confirming the match is case-insensitive and covers exactly the fixed set (`"Requirements"`, `"Functional Requirements"`) named in contract §5 — a `SourceSection` of `"requirements"` (lowercase) also matches; an unrelated heading like `"Notes"` does not.
- [X] T026 [P] [US2] Golden/contract test in `internal/cli/internalcmd/context_test.go`: `internal context <id>` (no flag) and `internal context <id> --prefer-section` are called against the same fixture; without the flag, item ordering is byte-identical to the pre-038 baseline; with the flag, the Requirements-sourced item moves ahead of its otherwise-equivalent, non-preferred sibling within the same tier (contract §5).

### Implementation for User Story 2

- [X] T027 [US2] Add `PreferSection bool` to `contextengine.Request` in `internal/context/request.go` (data-model.md: "zero value `false`... matching `Request`'s existing `QueryMode`-style... convention").
- [X] T028 [US2] Add `SectionPreference int` to `ScoreComponents` in `internal/context/rank.go`; extend `computeScoreComponents` to take `req.PreferSection` (or the relevant field) and compute a fixed positive bonus when a candidate has a wikilink-origin `Reason` whose `SourceSection` case-insensitively matches `"Requirements"` or `"Functional Requirements"`, added into `Total` alongside the existing `RelationWeight`/`IntentBonus`/`TextRelevance` contributors — zero whenever `PreferSection` is false.
- [X] T029 [US2] Add the `--prefer-section` boolean flag to `NewContextCmd` in `internal/cli/internalcmd/context.go`, threading it into `contextengine.Request.PreferSection`; document in the flag's help text and in a code comment that this is diagnostic/experimental (contract §5: "MUST NOT be interpreted... as implying a permanent behavior change").
- [X] T030 [US2] If `--diagnostic-scores` is also passed, ensure `renderScoreComponents` (or equivalent) in `internal/cli/internalcmd/context.go` includes the new `section_preference` field in `score_components` so the effect is visible, not just inferred from reordering.
- [X] T031 [US2] Run `go test ./internal/context/... ./internal/cli/internalcmd/...` and confirm T024-T026 now pass; run quickstart.md §4 manually and confirm ordering is unchanged without the flag and changes as documented with it.

**Checkpoint**: The preference-scoring capability exists, is off by default, and is independently testable — User Stories 1 and 2 both work without depending on each other's own command-line surface beyond the shared `context` command.

---

## Phase 5: User Story 3 - Retrieval Stays Bounded and Non-Redundant (Priority: P3)

**Goal**: Confirm — with tests written against this feature's own changes, not new mechanism — that a reference cycle or a heavily-referenced hub never causes unbounded expansion, and that no referenced content is loaded twice, exactly preserving the guarantees already built into `internal/context/collector.go`'s existing 2-hop cap and dedup.

**Independent Test**: Request a Context Pack for a target inside a two-artifact reference cycle, and for a target referenced from many other artifacts; confirm both requests complete, stay within the existing documented bound, and never duplicate content (quickstart.md §5).

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T032 [P] [US3] Regression test in `internal/context/collector_test.go`: two fixture artifacts that wikilink each other (A → B, B → A) — `Collect` for either one completes and `CandidateSet.Candidates` contains no unbounded/duplicated entries; each entry that does exist carries correct, distinct occurrence data (T009/T010 did not regress the existing 2-hop cap).
- [X] T033 [P] [US3] Regression test in `internal/context/collector_test.go`: one fixture artifact wikilinked by an unusually large number (e.g. 20+) of other fixture artifacts — `Collect` for it completes promptly and `Diagnostics.CandidatesConsidered`-equivalent count stays within the same bound as an equivalent pre-038 fixture would have produced (spec SC-002).
- [X] T034 [P] [US3] Regression test in `internal/context/collector_test.go`: two separate reference paths (e.g. a direct wikilink and a second-hop path) that both lead to the same underlying chunk — that chunk appears exactly once in the deduplicated `CandidateSet`, not once per path (spec FR-007, reusing 015's existing `mergeAndSort` dedup — confirms T009/T010's `Reason` plumbing did not bypass it).

### Implementation for User Story 3

- [X] T035 [US3] Run T032-T034 against the current (post-Phase-2/3/4) implementation. If any fails, fix the specific regression in `internal/context/collector.go` introduced by T009/T010's signature changes (e.g. a `reasonOf`/`chunkArtifact` call site that lost its original tier/relation behavior) — this phase is expected to require no new mechanism, only confirming none was accidentally broken.
- [X] T036 [US3] Run `go test ./internal/context/...` and confirm T032-T034 pass; run quickstart.md §5 manually against a built binary with a cycle/hub fixture and confirm the documented `Expected` outcomes.

**Checkpoint**: All three user stories are independently functional and tested; the feature's own safety guarantees are confirmed unregressed.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories.

- [X] T037 [P] Run `go vet ./...` and the project's full test suite (`go test ./...`) to confirm no regressions anywhere else in the codebase from the `ReferenceEntry`/`BacklinkEntry`/`Reason`/`chunkArtifact`/`connectedCandidates` signature changes.
- [X] T038 [P] Run the complete `quickstart.md` end-to-end (§1-§6) against a freshly built binary, including §6's stale-cache-rebuild scenario using a `.misterspec/cache/context.db` built before this feature (or a database with `PRAGMA user_version` manually set to `1`), confirming the automatic rebuild described in contract §6.
- [X] T039 Cross-reference this feature from `docs/context-engine-implementation.md`'s existing wikilink/reference sections (§5-§6, mirroring how 037's own note was added to §Phase 9), so a future reader of the source design document finds the provenance capability without re-reading Spec 038 in full.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup (T001) — BLOCKS all user stories. This phase touches shared types (`ReferenceOccurrence`, extended `ReferenceEntry`/`BacklinkEntry`/`Reason`, widened `links` schema) every story depends on.
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) only.
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) only — reads the same `Reason.SourceSection` data US1 surfaces, but does not depend on US1's own CLI/rendering changes (T019-T022).
- **User Story 3 (Phase 5)**: Depends on Foundational (Phase 2), and in practice is most meaningfully run after Phase 3/4's own `collector.go` changes (T009/T010) have landed, since it is a regression check against exactly those changes.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2/US3 — independently testable per quickstart.md §1-§3.
- **US2 (P2)**: No dependency on US1's own CLI surface, though both read the same Foundational-phase data; independently testable per quickstart.md §4.
- **US3 (P3)**: A regression check against Foundational-phase changes (T009/T010) rather than new user-facing surface; independently testable per quickstart.md §5, but most informative once US1/US2 exist to exercise realistic candidate sets.

### Within Each User Story

- Tests (T015-T018, T024-T026, T032-T034) MUST be written and FAIL before their corresponding implementation task.
- Foundational data/schema changes (Phase 2) before any story's own CLI-surface changes.
- Story complete (checkpoint) before moving to the next priority, though parallel staffing across US1/US2 is possible once Phase 2 completes (US3 benefits from following both).

### Parallel Opportunities

- T002 and T003 can run in parallel (implementation vs. its own test, though T003 must be written first per TDD and then made to pass by T002).
- T005, T006, T007 can run in parallel with each other once T004 lands (T004/T005 touch different files).
- T015, T016, T017, T018 (US1 tests) can run in parallel with each other.
- T019 and T020 (US1 implementation) can run in parallel — different files.
- T024, T025, T026 (US2 tests) can run in parallel with each other.
- T032, T033, T034 (US3 tests) can run in parallel with each other.
- T037 and T038 (Polish) can run in parallel.

---

## Parallel Example: Foundational Phase

```bash
# Launch independent Foundational tasks together once T002 exists:
Task: "Extend operations.BacklinkEntry in internal/operations/backlinks.go"
Task: "Unit tests for extended ReferenceEntry in internal/operations/references_test.go"
Task: "Unit tests for extended BacklinkEntry in internal/operations/backlinks_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run quickstart.md §1-§3 independently.
5. `references`/`backlinks`/`context --provenance` are now usable — the MVP for this feature.

### Incremental Delivery

1. Setup + Foundational → shared occurrence data and widened index ready.
2. Add User Story 1 → validate independently → provenance explainability shipped (MVP).
3. Add User Story 2 → validate independently → off-by-default preference scoring available for future evaluation.
4. Add User Story 3 → validate independently → safety guarantees reconfirmed under this feature's own changes.
5. Polish → cross-cutting regression pass and documentation.

### Parallel Team Strategy

With multiple contributors, after Phase 2 completes: one contributor takes US1 (Phase 3), another takes US2 (Phase 4) — both build on the same Foundational data without touching each other's files (US1: `references.go`/`backlinks.go`/`context.go`'s `--provenance`; US2: `request.go`/`rank.go`/`context.go`'s `--prefer-section`) — and a third writes US3's regression tests (T032-T034) as soon as Phase 2 lands, since they exercise Foundational-phase behavior directly.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- Constitution Principle IV boundary (plan.md Constitution Check) is load-bearing for scope: `--prefer-section` (T027-T030) MUST remain off by default through every task in this list — no task here promotes it to default ordering. That promotion is explicitly out of scope, deferred to a future Spec with recorded 037-eval-quality-efficiency evidence.
- Every `Reason`/`ReferenceEntry`/`BacklinkEntry` occurrence field task (T004, T005, T008) must preserve the "never conflated with a formal dependency" rule from spec FR-009 — a formal entry's `SourceSection`/`SourceLine` stay empty/zero, never populated from an unrelated wikilink occurrence.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
