---

description: "Task list template for feature implementation"
---

# Tasks: Document Model and Chunking

**Input**: Design documents from `/specs/013-document-model-chunking/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/document-chunking.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 011-wikilink-foundation's full test suite is an explicit, named regression gate (Foundational), since this feature must not change `ExtractWikiLinks`'s behavior at all.

**Organization**: Tasks are grouped by user story. **Foundational blocks User Story 1 and User Story 2** — both build on the flat, heading-boundary detection that reuses `fencedLines`. **User Story 3 (token estimation) depends on nothing in this feature at all** — it operates on any string, needs no `fencedLines`, no `Document`, no `Chunk` — so it can proceed fully in parallel with Foundational, User Story 1, and User Story 2. **User Story 2 depends on User Story 1** — a Chunk is derived from an already-parsed Document's Sections. Unlike 011/012, this feature adds **no dedicated non-regression user story**: it wires into no existing consumer at all (no CLI command, no call site in `internal/validation` or `internal/operations`) — the only pre-existing behavior it could plausibly affect is `ExtractWikiLinks` itself, and that is Foundational's own named regression gate (T003), not a separate story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/artifacts/            # existing package — Foundational, US1, US2, US3
internal/example/                # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package and no new dependency (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Factor 011's own `ExtractWikiLinks` fence-tracking loop into a small, shared `fencedLines` helper — the one piece of already-shipped code this feature touches, and the only infrastructure User Story 1's heading detection needs (research.md #3).

**⚠️ CRITICAL**: No user story implementation may begin until this phase's regression gate (T003) passes. (User Story 3 is exempt — it needs nothing from this phase at all.)

- [X] T001 [P] Regression tests confirming `ExtractWikiLinks`'s existing behavior is completely unchanged after the planned `fencedLines` extraction, in `internal/artifacts/wikilink_test.go` — every existing test in this file is itself part of this gate; add any additional case needed to pin down current behavior precisely before the refactor.
- [X] T002 Implement `fencedLines(lines []string) []bool` (data-model.md) in `internal/artifacts/wikilink.go`, refactoring `ExtractWikiLinks` to call it instead of its own inline fence-tracking loop — behavior-identical. Depends on T001.
- [X] T003 Regression checkpoint: run `go test ./internal/artifacts/...` and confirm every existing test (002-read-operations through 011-wikilink-foundation's own fixtures) passes unmodified — the explicit, named proof this refactor changed nothing observable about `ExtractWikiLinks`. Depends on T002.

**Checkpoint**: Foundation ready — `fencedLines` exists, shared and proven behavior-identical to 011's own fence-tracking. User Story 1 and 2 can begin; User Story 3 could already have started in parallel.

---

## Phase 3: User Story 1 - See an Artifact's Body as Its Real Structure, Not a Blob of Text (Priority: P1) 🎯 MVP

**Goal**: `ParseDocument` deterministically splits any artifact body into an ordered, flat list of heading-bounded `Section`s — correctly handling nested headings, a heading-free body, an empty body, and fenced example syntax.

**Independent Test**: Feed `ParseDocument` a body with top-level and nested headings, some with content and some without, plus a fenced code block showing example heading syntax; confirm every real heading appears in order at its correct level with exactly its own content, and the fenced example is never mistaken for a boundary.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T005.

- [X] T004 [P] [US1] Unit tests for `ParseDocument` in `internal/artifacts/document_test.go`, covering `docs/context-engine-implementation.md` §29.4's own list: a body with several headings at different levels, in document order, each with correct `Level`/`Body`/`StartLine`/`EndLine`; nested headings, where a parent heading with content only before its first child produces its own non-empty section, and one with no content of its own before its first child produces an empty-`Body` section (spec.md's own Edge Case — no duplication into children); an empty section (heading immediately followed by another heading); a body with no headings at all (exactly one section covering everything); a completely empty body (zero sections, FR-011); a fenced code block containing heading-like text is never treated as a boundary (reusing `fencedLines`); running `ParseDocument` twice on the same input produces identical output (FR-008).

### Implementation for User Story 1

- [X] T005 [US1] Implement `Section`, `Document`, and `ParseDocument` in new file `internal/artifacts/document.go` (data-model.md) — an ATX-heading-only (research.md #4), flat, non-duplicating line partition (research.md #2) reusing `fencedLines`. Depends on T004, T003 (Foundational).

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/artifacts/... -run TestParseDocument` passes on its own, with zero dependency on `Chunks` or token estimation.

---

## Phase 4: User Story 2 - Break an Artifact Into Traceable, Retrieval-Sized Pieces (Priority: P2)

**Goal**: `Chunks` derives one fully-traceable `Chunk` per non-empty `Section` of an already-parsed `Document`, preserving any wikilink syntax inside verbatim.

**Independent Test**: Take a `Document` with several sections, including one with an empty `Body`; call `Chunks` and confirm exactly one `Chunk` per non-empty section (none for the empty one), each correctly naming its path, heading, and line range, and confirm calling it again on the same input produces an identical result.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T007.

- [X] T006 [P] [US2] Unit tests for `Chunks` in `internal/artifacts/chunk_test.go`: one `Chunk` per non-empty-`Body` `Section`, in document order; no `Chunk` for an empty-`Body` `Section` (FR-007); every `Chunk`'s `Path`/`Heading`/`Level`/`StartLine`/`EndLine` matches its originating `Section` and the `path` argument exactly; a `Chunk`'s `Content` preserves an explicit wikilink (`[[TARGET]]`) inside it exactly as written, unmodified (FR-012); calling `Chunks` twice on the same `Document`/`path` produces identical output (FR-008).

### Implementation for User Story 2

- [X] T007 [US2] Implement `Chunk` and `Chunks` in new file `internal/artifacts/chunk.go` (data-model.md, contracts/document-chunking.md) — pure function of `path` and an already-parsed `Document`, no I/O of its own (FR-010). Depends on T006, T005 (US1).

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/artifacts/... -run TestChunks` passes on its own.

---

## Phase 5: User Story 3 - Know Roughly How Much a Piece of Content Will Cost (Priority: P3)

**Goal**: `EstimateTokens`/`Estimator` provide a deterministic, documented approximate token count for any piece of text — usable independently of `Document`/`Chunk`.

**Independent Test**: Estimate the same text twice and confirm identical results; estimate two texts of clearly different length and confirm the longer one's estimate is larger; estimate an empty string and confirm zero.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T009. Can be done at any point relative to Foundational/US1/US2 — no dependency on either.

- [X] T008 [P] [US3] Unit tests for `EstimateTokens` and `Estimator`/`DefaultEstimator` in new file `internal/artifacts/tokens_test.go`: the same text estimated twice yields identical results (determinism, FR-008-style); a longer text yields a larger-or-equal estimate than a clearly shorter one (monotonicity); an empty string yields exactly `0`; `DefaultEstimator{}.Estimate(text)` always equals `EstimateTokens(text)` for the same `text`.

### Implementation for User Story 3

- [X] T009 [US3] Implement `Estimator`, `DefaultEstimator`, and `EstimateTokens` in new file `internal/artifacts/tokens.go` (data-model.md, contracts/document-chunking.md) — `ceil(rune count / 4)` (research.md #7). Depends on T008 only.

**Checkpoint**: User Story 3 is independently complete and testable — `go test ./internal/artifacts/... -run TestEstimateTokens` passes on its own, with zero dependency on any other part of this feature.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T010 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 012-references-backlinks' packages included) and fix any findings.
- [X] T011 [P] Verify/extend `internal/artifacts`'s own package-level doc comment, cross-checked against `specs/013-document-model-chunking/contracts/document-chunking.md`.
- [X] T012 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: parse a fixture body with nested headings into a `Document`, derive its `Chunks`, confirm the empty-section/no-chunk case, confirm a fenced example heading is excluded, confirm an `EstimateTokens` call on a resulting chunk, confirming each matches `quickstart.md`'s own documented result exactly.
- [X] T013 Reconcile `specs/013-document-model-chunking/contracts/document-chunking.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T014 Full regression run: `go test ./...` across the entire module (001 through 013) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/artifacts/... -race` clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS User Story 1 and User Story 2 (both need `fencedLines`) — does not block User Story 3, which needs nothing from this phase.
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on User Story 1 (`Document`/`Section` must exist) and, transitively, Foundational.
- **User Story 3 (Phase 5)**: Depends on nothing in this feature — can start immediately, in parallel with Foundational/US1/US2.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational. No dependency on User Story 2 or 3.
- **User Story 2 (P2)**: Depends on User Story 1's `Document`/`Section` existing.
- **User Story 3 (P3)**: No dependency on User Story 1, 2, or Foundational — the most independent story of any feature so far.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational's regression gate (T003) passes before User Story 1's or 2's implementation begins.

### Parallel Opportunities

- T001 (Foundational's own test) and T008 (User Story 3's own test) in parallel — genuinely unrelated files and concerns.
- Once Foundational is done: **User Story 1 can proceed**; User Story 3 could already be fully done by this point, having started immediately.
- User Story 2 must wait for User Story 1's `Document`/`Section` to exist — not independently parallel with it, the way User Story 3 is with everything else.
- Within Polish: T010 and T011 in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (`fencedLines`, regression-proven).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/artifacts/... -run TestParseDocument` green, independently.
5. This alone already proves an artifact's real structure can be deterministically recovered — the foundation everything else (chunking, and eventually indexing) builds on.

### Incremental Delivery

1. Setup (nothing) + Foundational (`fencedLines`, regression-proven).
2. Add User Story 1 → validate independently → structure usable (MVP).
3. Add User Story 2 (depends on US1) → validate independently → traceable chunks usable.
4. Add User Story 3 (independent — may already be done in parallel) → validate independently → token estimation usable.
5. Polish (Phase 6), including the full-module `-race` regression run (T014).

### Team Strategy

Unlike 011's strictly sequential chain, this feature has real parallel
opportunity: one developer could take Foundational → User Story 1 →
User Story 2 in sequence while a second takes User Story 3 entirely
independently, the same shape 012's References/Backlinks pair
demonstrated — except here the independent story (US3) needs literally
nothing from the rest of the feature, not even Foundational.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `ParseDocument`/`Chunks`/`EstimateTokens` are pure functions of their inputs — no I/O, no mutation, deterministic output every time (FR-008, FR-010).
- A `Chunk` carries no `ArtifactID` and no `Tokens` field — both deliberate scope decisions (research.md #5, #6), not omissions to fix later in this feature.
- No CLI-layer file is touched anywhere in this feature (research.md) — Phase 8's own "internal context" command is the later, actual CLI surface this capability feeds into.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
