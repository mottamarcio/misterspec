---

description: "Task list template for feature implementation"
---

# Tasks: Disposable SQLite Index

**Input**: Design documents from `/specs/014-sqlite-index/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/index.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 011-wikilink-foundation's, 012-references-backlinks's, and 013-document-model-chunking's full test suites are explicit, named regression gates (Polish), since this feature adds no call site into any of them requiring a change at all.

**Organization**: Tasks are grouped by user story. Unlike 011-013, this feature has real **Setup** work (a new external dependency) and a substantial **Foundational** phase (schema + Store lifecycle) both User Story 1 and later stories build on. **User Story 1 (build + search) must complete before User Story 2 (incremental sync)** — US2 extends the same per-artifact indexing logic US1 establishes with change/delete detection. **User Story 3 (relationship lookups) depends on both US1 and US2** — it extends the same indexing/cleanup logic again to add `links` population, and needs US2's delete-on-change/delete-on-removal logic already in place to keep `links` consistent too. This feature's story chain is therefore strictly sequential, closer to 007's/011's own shape than 012's/013's parallel-story pairs — each story deepens the same underlying `Sync` implementation rather than adding an independent capability beside it.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
go.mod, go.sum                        # + modernc.org/sqlite v1.39.0
internal/context/index/                 # new package — Foundational, US1, US2, US3
internal/example/                         # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add this feature's one new external dependency, verified not to bump the module's `go` directive (research.md #1), and stand up the new package's own skeleton.

- [X] T001 [P] Add `modernc.org/sqlite` pinned to `v1.39.0` via `go get modernc.org/sqlite@v1.39.0` followed by `go mod tidy`; confirm `go.mod`'s `go` line is still exactly `1.23.4` (research.md #1 — the exact check 010's own experience established as mandatory before trusting a new dependency); confirm `go build ./...` still succeeds.
- [X] T002 [P] Create `internal/context/index/doc.go` with the package's own doc comment (data-model.md, contracts/index.md).

**Checkpoint**: The new dependency is present and verified safe; the package exists and compiles as an empty shell.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Stand up the schema (with version handling) and the `Store`'s own open/close lifecycle — the one piece of shared infrastructure every later story's own indexing/search/lookup logic runs on top of.

**⚠️ CRITICAL**: No user story implementation may begin until this phase's regression checkpoint (T007) passes.

- [X] T003 [P] Tests for schema creation and versioning in new file `internal/context/index/schema_test.go`: opening a fresh path creates every table (`documents`, `chunks`, `chunks_fts`, `links`) and sets `PRAGMA user_version` to this package's current schema version; opening an existing database whose `user_version` doesn't match recreates the schema cleanly (research.md #7).
- [X] T004 Implement the DDL (data-model.md's schema exactly) and version-check/recreate logic in new file `internal/context/index/schema.go`. Depends on T003, T001.
- [X] T005 [P] Tests for `Open`/`Close` in new file `internal/context/index/sqlite_test.go`: `Open` creates `.misterspec/cache/`'s equivalent parent directory if missing and returns a working `Store`; `Close` releases the handle without error; reopening the same path afterward succeeds and preserves whatever schema/data was already there.
- [X] T006 Implement `Store`, `SyncReport`, `SyncError`, `SearchResult`, `Link` in new file `internal/context/index/store.go` (contracts/index.md); implement the concrete SQLite-backed store plus `Open`/`Close` in new file `internal/context/index/sqlite.go`. Depends on T004, T005.
- [X] T007 Regression checkpoint: `go build ./...` succeeds with the new dependency and package wired in; `go vet ./internal/context/...` clean. Depends on T006.

**Checkpoint**: Foundation ready — a `Store` can be opened and closed, with a correct, versioned, empty schema. User story implementation can begin.

---

## Phase 3: User Story 1 - Build a Searchable Index From the Project's Own Content (Priority: P1) 🎯 MVP

**Goal**: `Sync` performs a correct initial build over every eligible artifact (`ArtifactType`'s full 9 values, research.md #3), `Search` returns matching indexed content, and `Rebuild` reproduces a search-equivalent index from scratch at any time.

**Independent Test**: Build an index over a fixture project with a representative artifact of each eligible type plus a source-code-only directory; confirm every eligible artifact's content is indexed and none of the source code is; search for a term known to appear in exactly one indexed chunk and confirm it's returned; delete the index and rebuild it, confirming the same search still returns the same result.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T011-T013.

- [X] T008 [P] [US1] Tests for `Sync`'s initial-build path in new file `internal/context/index/sync_test.go`: a fixture with one artifact of each `ArtifactType` value is fully indexed (every `documents`/`chunks` row present); a fixture containing only non-artifact files (e.g. `*.go`) indexes nothing; `SyncReport.Indexed` matches the actual count; an artifact with an unclassifiable path (`ClassifyPath` fails) is silently skipped, not reported as an error (research.md #3).
- [X] T009 [P] [US1] Tests for `Search` in new file `internal/context/index/search_test.go`: a term appearing in exactly one indexed chunk's content or heading returns that chunk with correct `Path`/`Heading`/`StartLine`/`EndLine`; a term appearing nowhere returns an empty, non-error result; the `limit` parameter is respected when more matches exist than the limit.
- [X] T010 [P] [US1] Tests for `Rebuild` in `internal/context/index/sync_test.go`: rebuilding an already-populated, unchanged-since index produces a search-equivalent result (same terms return the same underlying chunks, SC-004); rebuilding after the project's content has changed entirely since the index was last built converges to the current state, with no stale rows left from before.

### Implementation for User Story 1

- [X] T011 [US1] Implement per-artifact indexing (via `artifacts.ClassifyPath`/`ParseMetadata`/`ReadBody`/`ParseDocument`/`Chunks`, inserting `documents`/`chunks`/`chunks_fts` rows — no `links` yet, US3's own concern) and the initial-build walk over `cfg.ArtifactsDir` in new file `internal/context/index/sync.go`'s `Sync`. Depends on T008, T007.
- [X] T012 [US1] Implement `Search` (an FTS5 `MATCH` query against `chunks_fts`, ranked via `bm25()`, joined back to `chunks`/`documents` for full provenance) in new file `internal/context/index/search.go`. Depends on T009, T011.
- [X] T013 [US1] Implement `Rebuild` (clear every row from `documents`/`chunks`/`chunks_fts`/`links`, then run the same per-artifact indexing walk `Sync` uses) in `internal/context/index/sync.go`. Depends on T010, T011.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/context/index/... -run "TestSync_InitialBuild|TestSearch|TestRebuild"` passes on its own, with zero dependency on incremental sync or relationship lookups.

---

## Phase 4: User Story 2 - Keep the Index Cheaply Up to Date (Priority: P2)

**Goal**: `Sync` correctly distinguishes unchanged, changed, new, and deleted artifacts using each artifact's existing content fingerprint, reprocessing only what actually needs it, wrapped in one all-or-nothing transaction.

**Independent Test**: Build an index, then change one artifact, add a new one, and delete another, leaving the rest untouched; re-run `Sync` and confirm only the changed/new/deleted artifacts' indexed rows were affected — every other artifact's own rows are provably identical to before.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T016.

- [X] T014 [P] [US2] Tests for incremental `Sync` in `internal/context/index/sync_test.go`: an unchanged artifact (same fingerprint) is not reprocessed (its `documents.indexed_at` is untouched); a changed artifact's old `chunks`/`chunks_fts` rows are fully replaced with its current content; a brand-new artifact is added; a previously-indexed, now-deleted artifact's rows are removed entirely; `SyncReport`'s `Updated`/`Skipped`/`Indexed`/`Removed` counts are all correct for one mixed batch covering all four cases at once.
- [X] T015 [P] [US2] Tests for transactional safety in `internal/context/index/sync_test.go`: a `Sync` run that fails partway (simulated) leaves the database exactly as it was before the call — no partially-applied rows from the failed run are observable afterward; no project artifact file is ever modified by any `Sync`/`Rebuild` call, verified via each fixture's own file content/mtime.

### Implementation for User Story 2

- [X] T016 [US2] Extend `Sync` in `internal/context/index/sync.go` with fingerprint comparison (via `operations.Fingerprint`) driving the unchanged/changed/new/deleted branches — deleting an artifact's stale `chunks`/`chunks_fts` rows before reindexing it on change, and entirely on deletion — all wrapped in one transaction with rollback on any error (research.md #9). Depends on T014, T015, T011 (US1's own per-artifact indexing, reused).

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/context/index/... -run "TestSync_Incremental|TestSync_Transactional"` passes on its own, building on User Story 1's already-working initial build.

---

## Phase 5: User Story 3 - Look Up an Artifact's Relationships From the Index (Priority: P3)

**Goal**: `Outgoing`/`Incoming` report an already-indexed artifact's relationships, sourced from `operations.References` for the five entity types it already covers (research.md #4), matching that direct computation exactly.

**Independent Test**: Build an index over a fixture with known formal and semantic relationships between artifacts; look up one artifact's relationships from the index and confirm the result matches `operations.References`/`Backlinks`'s own direct computation for the same artifact and fixture.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T018-T019.

- [X] T017 [P] [US3] Tests for `Outgoing`/`Incoming` in new file `internal/context/index/links_test.go`: after syncing a fixture with known formal and semantic relationships among the five ID-bearing types, `Outgoing(id)`/`Incoming(id)` match `operations.References`/`Backlinks`'s own direct computation exactly for the same fixture; an artifact with no relationships returns a well-formed empty result, not an error; a synced Plan/Tasks/Validation/Constitution artifact contributes zero `links` rows (research.md #4) even though it was indexed for search.

### Implementation for User Story 3

- [X] T018 [US3] Extend the per-artifact indexing step in `internal/context/index/sync.go` to also populate `links` rows via `operations.References` for the five ID-bearing types (`ArtifactType.HasEntityID()`), and extend User Story 2's own change/delete cleanup to also remove an artifact's stale `links` rows. Depends on T017, T016.
- [X] T019 [US3] Implement `Outgoing`/`Incoming` (querying `links` by `target_artifact_id`/`source_artifact_id`) in `internal/context/index/sqlite.go`. Depends on T018.

**Checkpoint**: User Story 3 is independently complete and testable — `go test ./internal/context/index/... -run "TestOutgoing|TestIncoming"` passes on its own.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T020 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 013-document-model-chunking's packages included) and fix any findings.
- [X] T021 [P] Verify/extend `internal/context/index`'s own package-level doc comment, cross-checked against `specs/014-sqlite-index/contracts/index.md`.
- [X] T022 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: initial build, search, an incremental sync after a targeted edit, a deletion, an `Outgoing`/`Incoming` lookup matching `operations.References`/`Backlinks`, a full rebuild, and an artifact that fails to parse not halting the run — each matching `quickstart.md`'s own documented result exactly.
- [X] T023 Reconcile `specs/014-sqlite-index/contracts/index.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T024 Full regression run: `go test ./...` across the entire module (001 through 014) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/context/... -race` clean, and `go.mod`'s `go` directive is confirmed still exactly `1.23.4` (research.md #1's own concern, checked one final time against the fully resolved dependency graph).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — the new dependency and package skeleton come first.
- **Foundational (Phase 2)**: Depends on Setup. BLOCKS all three user stories — none can index or query anything without a working, versioned schema and a `Store` to open.
- **User Story 1 (Phase 3)**: Depends on Foundational.
- **User Story 2 (Phase 4)**: Depends on User Story 1 — it extends the same per-artifact indexing logic with change detection, rather than adding an independent capability.
- **User Story 3 (Phase 5)**: Depends on User Story 1 and User Story 2 both — it extends the same indexing logic again (to add `links`) and needs US2's own delete/cleanup logic already in place to keep `links` consistent through changes and deletions too.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational.
- **User Story 2 (P2)**: Depends on User Story 1's own per-artifact indexing existing to extend.
- **User Story 3 (P3)**: Depends on both User Story 1 and User Story 2.

This feature's story chain (US1 → US2 → US3) is strictly sequential —
closer to 007-project-bootstrap's and 011-wikilink-foundation's own
shape than 012's/013's parallel-story pairs, because each story here
deepens the same underlying `Sync` implementation rather than adding an
independent, separately-testable capability beside it.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational's regression checkpoint (T007) passes before any user story's implementation begins.

### Parallel Opportunities

- T001 and T002 (Setup) in parallel.
- T003 and T005 (Foundational's two test files) in parallel.
- T008, T009, and T010 (User Story 1's three test files) in parallel.
- T014 and T015 (User Story 2's two test concerns) in parallel — though both land in the same `sync_test.go` file, so coordinate to avoid a merge conflict if worked on literally simultaneously.
- Within Polish: T020 and T021 in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (`modernc.org/sqlite` added and verified safe).
2. Complete Phase 2: Foundational (schema + `Store` lifecycle).
3. Complete Phase 3: User Story 1 (build, search, rebuild).
4. **STOP and VALIDATE**: `go test ./internal/context/index/... -run "TestSync_InitialBuild|TestSearch|TestRebuild"` green, independently.
5. This alone already proves the core value: the project's own content can be searched locally.

### Incremental Delivery

1. Setup + Foundational (schema/lifecycle, regression-proven).
2. Add User Story 1 → validate independently → build-and-search usable (MVP).
3. Add User Story 2 (extends US1) → validate independently → the index stays cheap to keep current.
4. Add User Story 3 (extends US1+US2) → validate independently → relationship lookups usable from the index.
5. Polish (Phase 6), including the full-module `-race` regression run and the final `go.mod` directive check (T024).

### Team Strategy

Unlike 012's/013's genuinely parallel story pairs, this feature's chain
is sequential by nature (each story extends the same `Sync`
implementation) — a single developer working through it in order is the
natural shape, the same as 007's own Inspect → Bootstrap → Verify chain
or 011's own US1 → US2 → US3 progression.

---

## Notes

- [P] tasks touch different files (or, for T014/T015, different concerns within the same new file) with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `links` rows are populated only for the five entity types `operations.References` already covers (research.md #4) — Plan/Tasks/Validation/Constitution are indexed for full-text search only; this is a deliberate scope decision, not something to "finish later" within this feature.
- A `Sync`/`Rebuild` run is exactly one transaction (research.md #9) — implement the transaction boundary once, in User Story 2's own task (T016), and never re-open it per-artifact.
- No CLI-layer file is touched anywhere in this feature (research.md) — Phase 8's own "internal context" command is the later, actual CLI surface this capability feeds into.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
