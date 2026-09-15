---

description: "Task list template for feature implementation"
---

# Tasks: Context Collector and Retrieval

**Input**: Design documents from `/specs/015-context-collector/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/collector.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 011-wikilink-foundation's, 012-references-backlinks's, 013-document-model-chunking's, and 014-sqlite-index's full test suites are explicit, named regression gates (Polish), since this feature adds no call site into any of them requiring a change at all.

**Organization**: Tasks are grouped by user story. Like 014, this feature's story chain is **strictly sequential** — each story extends the same `Collect` function body rather than adding an independent capability beside it: **User Story 1** (mandatory baseline) must exist before **User Story 2** (structural/semantic) can extend it, which must exist before **User Story 3** (text) and **User Story 5** (second-hop, which needs Story 2's own first-hop set) can extend it further. **User Story 4** (deduplication) is, like 011's own User Story 3, a *regression/correctness guarantee* about behavior the Foundational `mergeAndSort` helper and every story's own use of it already provide — it has a dedicated proving test but no separate implementation task of its own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Every task names its exact file path

## Path Conventions

```text
internal/context/            # existing package (014's own index/ subpackage) — gains its first top-level files here
internal/example/              # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package and no new dependency (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Stand up the shared data model every user story's own `Collect` logic builds on: the `Request`/`Intent` shape (with validation) and the `Tier`/`Reason`/`Candidate`/`CandidateSet` shape (with the merge-and-sort logic that gives `Collect` its final deterministic, deduplicated output at every stage of this feature's own incremental build-out).

**⚠️ CRITICAL**: No user story implementation may begin until this phase's regression checkpoint (T005) passes.

- [X] T001 [P] Unit tests for intent validation in new file `internal/context/request_test.go`: an empty `Intent` is valid; each of the five recognized values (`planning`, `tasks`, `implementation`, `validation`, `analysis`) is valid; any other string is rejected (FR-002, research.md #4).
- [X] T002 [P] Unit tests for the merge-and-sort helper in new file `internal/context/result_test.go` (data-model.md): two entries sharing the same `(Path, StartLine, EndLine)` merge into one `Candidate` carrying every distinct `Reason` (FR-008); a `TierSecondHop` reason for a candidate that also has any direct (Tier 0-3) reason for the exact same chunk is dropped, never kept alongside it (FR-011, research.md #8); the final result is ordered by `Tier` ascending, then `Path`, then `StartLine` (research.md #10), with no separate scoring step.
- [X] T003 [P] Implement `Request`, `Intent` and its five constants, and `validateIntent` in new file `internal/context/request.go` (data-model.md, contracts/collector.md). Depends on T001.
- [X] T004 [P] Implement `Tier` and its five constants, `Reason`, `Candidate`, `CandidateSet`, and the merge-and-sort helper in new file `internal/context/result.go` (data-model.md, contracts/collector.md). Depends on T002.
- [X] T005 Regression checkpoint: `go build ./...` succeeds with the new files in place; `go vet ./internal/context/...` clean. Depends on T003, T004.

**Checkpoint**: Foundation ready — the `Request`/`Intent`/`Tier`/`Reason`/`Candidate`/`CandidateSet` shapes exist, with validated intent and proven merge-and-sort behavior. User story implementation can begin.

---

## Phase 3: User Story 1 - Always Get the Non-Negotiable Baseline (Priority: P1) 🎯 MVP

**Goal**: `Collect` resolves the target (rejecting an unknown or out-of-scope one consistently with 012/014's own sentinel), and always includes every chunk of the project's Constitution (if one exists) and every chunk of the target itself.

**Independent Test**: Request context for an existing target with nothing else connected to it; confirm the result contains the target's own chunks and the Constitution's own chunks (if one exists), each correctly labeled, and confirm an unknown or out-of-scope target is rejected the same way `operations.References` already rejects it.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T007.

- [X] T006 [P] [US1] Unit tests for `Collect`'s mandatory-tier behavior in new file `internal/context/collector_test.go`: a target with no connections at all still returns its own chunks and the project's Constitution's own chunks, each labeled `{TierMandatory, "target"}`/`{TierMandatory, "constitution"}`; a project with no Constitution yet still returns the target's own chunks, not an error (FR-004); a request naming a target that does not exist, or one outside the five entity types `operations.References` covers, is rejected via `operations.ErrInvalidTarget`; an unrecognized `Intent` value is rejected (FR-002).

### Implementation for User Story 1

- [X] T007 [US1] Implement `Collect` in new file `internal/context/collector.go` (data-model.md's Algorithm, steps 1-4 and 8-9 only at this stage): validate intent, resolve the target via `operations.Inspect` with the five-type scope check (research.md #3), chunk the Constitution and the target (`artifacts.ReadBody`/`ParseDocument`/`Chunks`, 011/013), and return the merged-and-sorted result. Depends on T006, T005 (Foundational).

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/context/... -run TestCollect_Mandatory` passes on its own, with zero dependency on structural, semantic, text, or second-hop collection.

---

## Phase 4: User Story 2 - Discover Everything Structurally and Semantically Connected (Priority: P2)

**Goal**: `Collect` additionally gathers the target's own formal relationships and semantic connections (both outgoing and incoming), reusing `operations.References`/`Backlinks` directly, each chunk labeled with its specific relationship kind.

**Independent Test**: Request context for a target with a known parent, a known dependency, and a known incoming reference from another artifact; confirm all three appear, each labeled with its specific relationship.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T009.

- [X] T008 [P] [US2] Unit tests for `Collect`'s structural/semantic tier in `internal/context/collector_test.go`: a target with a declared parent and a declared dependency surfaces both, labeled `TierStructural` with their own relation; a target with an outgoing wikilink surfaces the linked artifact labeled `TierSemantic`/`"wikilink"`; an artifact that formally or semantically references the target surfaces it labeled `TierSemantic`/`"backlink"`; a target with no such connections contributes nothing beyond User Story 1's own baseline.

### Implementation for User Story 2

- [X] T009 [US2] Extend `Collect` (data-model.md's Algorithm step 5) to call `operations.References`/`Backlinks` for the target, resolving and chunking each connected artifact directly from the filesystem (research.md #5 — never through `internal/context/index`), labeling each per data-model.md's `Reason` table. Depends on T008, T007.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/context/... -run TestCollect_Structural` passes on its own, on top of User Story 1's already-working baseline.

---

## Phase 5: User Story 3 - Surface Relevant Text Matches (Priority: P3)

**Goal**: `Collect` additionally queries the local search index (014) using the request's own query or task text (never a fabricated one), labeling matches as text hits.

**Independent Test**: Request context with a free-text question known to match content in exactly one other artifact; confirm that content appears, labeled as a text match. Request context with no query and no task; confirm this tier contributes nothing.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T011.

- [X] T010 [P] [US3] Unit tests for `Collect`'s text tier in `internal/context/collector_test.go`: a request with a `Query` matching content elsewhere in the project surfaces it labeled `TierText`/`"text_match"`; a request with no `Query` but a non-empty `Task` uses `Task`'s own text verbatim as the search query (research.md #6); a request with neither contributes nothing from this tier, not an error; a `Query` matching nothing contributes nothing, not a failure.

### Implementation for User Story 3

- [X] T011 [US3] Extend `Collect` (data-model.md's Algorithm step 6) to call `store.Search` with `Request.Query`, falling back to `Request.Task` verbatim when `Query` is empty (research.md #6), labeling every result `TierText`/`"text_match"`. Depends on T010, T009.

**Checkpoint**: User Story 3 is independently complete and testable — `go test ./internal/context/... -run TestCollect_Text` passes on its own.

---

## Phase 6: User Story 4 - Never Show the Same Content Twice (Priority: P4)

**Goal**: The same underlying chunk discovered through more than one signal (structural, semantic, text, or second-hop) appears exactly once, carrying every distinct reason.

**Independent Test**: Construct a target whose dependency is also a free-text match for the request's own query, and a target reachable via two distinct structural relationships from the same source; confirm each collapses to one `Candidate` with multiple `Reason`s.

### Tests for User Story 4

> Write this test FIRST; confirm it passes only once User Story 1-3's own implementation is complete and correct (this story is a regression/correctness guarantee about the Foundational merge-and-sort helper already proven in T002 and already in continuous use since T007 — not new `Collect` behavior of its own, mirroring 011-wikilink-foundation's own User Story 3).

- [X] T012 [US4] Dedicated tests in `internal/context/collector_test.go` directly proving spec.md's own User Story 4 acceptance scenarios end-to-end through `Collect` itself (not just the Foundational unit-level `mergeAndSort` tests from T002): a target's own direct dependency that is also a free-text match for the request's query appears exactly once, listing both `TierStructural` and `TierText` reasons; a target reachable through two distinct structural relationships from the same source artifact (for example, both a formal dependency and an incoming backlink) appears exactly once, listing both. Depends on T011.

**Checkpoint**: User Story 4 is independently complete and testable — `go test ./internal/context/... -run TestCollect_Deduplicates` passes on its own, proving a guarantee that already held once User Stories 1-3 existed.

---

## Phase 7: User Story 5 - Optionally Look One Step Further (Priority: P5)

**Goal**: `Collect` additionally expands one bounded, outgoing-only hop from each first-hop (Tier 2/3) artifact, never a third hop, and never duplicating what a direct connection already found (User Story 4's own guarantee holding for this tier too).

**Independent Test**: Construct a target whose direct dependency itself has its own further dependency; confirm the further dependency appears, labeled `TierSecondHop`, distinguishable from a direct connection — and confirm nothing three or more hops out is ever included.

### Tests for User Story 5

> Write these tests FIRST; confirm they fail before implementing T014.

- [X] T013 [P] [US5] Unit tests for `Collect`'s second-hop tier in `internal/context/collector_test.go`: a target's direct dependency that itself has its own further dependency surfaces that further artifact labeled `TierSecondHop` with its own relation (research.md #7); a second-hop discovery that duplicates something already found directly does not add a duplicate entry, and the merged entry's reasons never include a redundant `TierSecondHop` alongside the direct one (FR-011, already proven at the unit level in T002 — this test proves it end-to-end through `Collect`); a chain three or more hops from the target never appears.

### Implementation for User Story 5

- [X] T014 [US5] Extend `Collect` (data-model.md's Algorithm step 7) to compute each first-hop artifact's own outgoing `operations.References` — never that artifact's own backlinks, never a third hop (research.md #7) — labeling every result `TierSecondHop` with its own relation. Depends on T013, T009 (User Story 2's own first-hop set).

**Checkpoint**: User Story 5 is independently complete and testable — `go test ./internal/context/... -run TestCollect_SecondHop` passes on its own.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all five stories, no new capability.

- [X] T015 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 014-sqlite-index's packages included) and fix any findings.
- [X] T016 [P] Verify/extend `internal/context`'s own package-level doc comment, cross-checked against `specs/015-context-collector/contracts/collector.md`.
- [X] T017 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: mandatory baseline, structural/semantic connections, a text match, cross-tier deduplication, bounded second-hop expansion, and an out-of-scope target rejection — each matching `quickstart.md`'s own documented result exactly.
- [X] T018 Reconcile `specs/015-context-collector/contracts/collector.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T019 Full regression run: `go test ./...` across the entire module (001 through 015) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/context/... -race` clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS every user story — none can produce a result without the `Request`/`Candidate`/`CandidateSet` shapes and a proven merge-and-sort helper.
- **User Story 1 (Phase 3)**: Depends on Foundational.
- **User Story 2 (Phase 4)**: Depends on User Story 1's own `Collect` to extend.
- **User Story 3 (Phase 5)**: Depends on User Story 2's own `Collect` to extend.
- **User Story 4 (Phase 6)**: Depends on User Story 1-3 all existing — nothing to prove deduplicated until they can actually overlap.
- **User Story 5 (Phase 7)**: Depends on User Story 2's own first-hop set specifically (not User Story 3's text tier).
- **Polish (Phase 8)**: Depends on all five user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational.
- **User Story 2 (P2)**: Depends on User Story 1.
- **User Story 3 (P3)**: Depends on User Story 2 (extends the same `Collect` body, though functionally independent of Tier 2/3's own content).
- **User Story 4 (P4)**: Depends on User Story 1, 2, and 3 all being complete.
- **User Story 5 (P5)**: Depends on User Story 2 specifically.

This feature's chain (US1 → US2 → US3 → [US4 proving] / US5 extending
US2) is strictly sequential, the same shape 014's own Setup →
Foundational → US1 → US2 → US3 chain took — every story deepens the
same `Collect` function rather than adding an independent capability
beside it.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V) — except User Story 4, a regression/correctness guarantee about already-implemented behavior, matching 011's own User Story 3 precedent.
- Foundational's regression checkpoint (T005) passes before any user story's implementation begins.

### Parallel Opportunities

- T001 and T002 (Foundational's two test files) in parallel.
- T003 and T004 (Foundational's two implementation files) in parallel, once their own tests exist.
- T006, T008, T010, and T013 (each story's own test-writing task) are only truly parallelizable with each other in the sense of authoring — since each story's *implementation* task depends on the previous story's implementation existing, write tests for a later story only once ready to implement it, to avoid tests describing behavior `Collect` doesn't support yet.
- Within Polish: T015 and T016 in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (`Request`/`Intent`/`Tier`/`Reason`/`Candidate`/`CandidateSet`, merge-and-sort proven).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/context/... -run TestCollect_Mandatory` green, independently.
5. This alone already proves the one guarantee nothing else in this feature may ever violate: the mandatory baseline is never missing.

### Incremental Delivery

1. Setup (nothing) + Foundational (data model + merge-and-sort, proven).
2. Add User Story 1 → validate independently → mandatory baseline usable (MVP).
3. Add User Story 2 (extends US1) → validate independently → structural/semantic context usable.
4. Add User Story 3 (extends US2) → validate independently → text-match context usable.
5. Add User Story 4 (proves US1-3's own combined behavior) → validate independently → deduplication formally proven.
6. Add User Story 5 (extends US2) → validate independently → bounded second-hop context usable.
7. Polish (Phase 8), including the full-module `-race` regression run (T019).

### Team Strategy

Unlike 012's/013's parallel story pairs, this feature's chain is
sequential by nature (each story extends the same `Collect` function
body) — a single developer working through it in order is the natural
shape, the same as 007's, 011's, and 014's own strictly sequential
chains.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `Collect` is strictly read-only (FR-012) — no mutation of any project artifact, the reference graph, or the search index, anywhere in this feature.
- Tiers 0-3 never read from `internal/context/index` — only Tier 3's text-match half (User Story 3) does (research.md #5).
- No CLI-layer file is touched anywhere in this feature (research.md) — Phase 8's own "internal context" command is the later, actual CLI surface this capability feeds into.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
