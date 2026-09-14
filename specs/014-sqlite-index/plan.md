# Implementation Plan: Disposable SQLite Index

**Branch**: `014-sqlite-index` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/014-sqlite-index/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 4
("Disposable SQLite Index") exactly: a new `internal/context/index`
package wrapping a pure-Go, CGO-free SQLite database
(`modernc.org/sqlite`, pinned to `v1.39.0` after verifying — the same
way 010 learned to — that neither it nor any transitive dependency
requires a `go` directive newer than this module's own `1.23.4`). The
index stores `documents`/`chunks`/`links` rows plus an FTS5 table,
built and incrementally synchronized from the project's own files using
`ClassifyPath` (001) to decide what's eligible, `Fingerprint` (002) to
detect change, `ReadBody`/`ParseDocument`/`Chunks` (011/013) to derive
content, and `References` (012) to populate relationship edges for the
five entity types 012 already covers — consistency with 012 by
construction, not by parallel reimplementation. Fully disposable: never
authoritative, always safe to delete and rebuild. No new CLI surface —
Phase 8's own later "internal context" command is the real entry point.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged — verified `modernc.org/sqlite@v1.39.0` and its full transitive dependency tree all declare `go <= 1.23.0`, research.md #1).
**Primary Dependencies**: `modernc.org/sqlite v1.39.0` (new — the first new external dependency since 010's Bubble Tea stack). Go standard library's `database/sql` for the driver interface.
**Storage**: A new disposable SQLite database at `.misterspec/cache/context.db` (research.md #6) — never authoritative (Constitution Principle III), never committed to version control, always safe to delete. The project's own files remain untouched (FR-011) — every read into them reuses already-existing, already read-only functions.
**Testing**: `go test` — new `internal/context/index` tests covering `docs/context-engine-implementation.md` §29.5's own list (initial build, unchanged sync, changed/new/deleted file, transaction rollback, database rebuild, FTS synchronization, incoming/outgoing links); 011's, 012's, and 013's own full suites re-run unmodified as the named regression gate, since this feature adds no call site into any of them requiring a change. Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged — `modernc.org/sqlite`'s pure-Go implementation preserves this project's own single-static-binary, no-C-toolchain property exactly (research.md #1).
**Project Type**: Single Go module. One new package (`internal/context/index`); no existing package modified at all — this feature only reads from four existing packages, adding no new call site *into* any of them.
**Performance Goals**: A full initial build/rebuild is O(artifacts) with an accepted, documented O(k) cost per ID-bearing artifact from reusing `operations.References` directly (research.md #5) — the same order of cost 012's own `Backlinks` already accepts per call today, at this project's own stated scale (hundreds to low thousands of artifacts). Incremental `Sync` (Phase 5's own focus) only reprocesses artifacts whose fingerprint actually changed.
**Constraints**: Strictly read-only with respect to every project artifact (FR-011) — every mutation this feature performs is confined to its own disposable database. A `Sync`/`Rebuild` run is exactly one SQL transaction (research.md #9) — it either fully commits or leaves the database completely unchanged (FR-007). `links` rows are populated only for the five entity types `operations.References` already covers (research.md #4) — Plan/Tasks/Validation/Constitution are indexed for full-text search only, no new formal-relationship scope invented for them.
**Scale/Scope**: `internal/context/index/{doc.go, schema.go, store.go, sqlite.go, sync.go, search.go}` (new) + matching `_test.go` files. `go.mod`/`go.sum` (+ `modernc.org/sqlite` and its transitive dependencies). No existing package modified, no CLI change.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Indexing, fingerprint comparison, and FTS querying are purely mechanical — no judgment about what content *means* happens anywhere in this feature; ranking/interpretation remains later phases' own job. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A — different boundary).** This feature mutates only its own disposable database, never a project artifact; the existing deterministic-operation mutation boundary (Create/CreateArtifact) is completely untouched. |
| III. Filesystem Is Single Source of Truth | **Pass, the central discipline of this entire feature.** The index is explicitly, permanently non-authoritative (FR-002, FR-012) — always fully derivable from the project's own files, always safe to delete, exactly mirroring `internal/lock`'s own established precedent for non-authoritative state. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised repeatedly.** Only `internal/context/index` is created now — not `internal/context`'s own collector/ranker files, which have no real consumer yet (research.md #2); `ArtifactType` is reused verbatim for "what gets indexed" rather than a new list (research.md #3); `links` scope for Plan/Tasks/Validation/Constitution is deliberately not extended beyond 012's own already-drawn boundary (research.md #4); schema versioning reuses SQLite's own built-in `PRAGMA user_version` rather than a bespoke metadata table (research.md #7); FTS sync is explicit Go-driven writes rather than trigger machinery (research.md #8); sync cost is accepted as-is rather than prematurely optimized (research.md #5). |
| V. Test-First Discipline | **Gate carried into tasks.** `docs/context-engine-implementation.md` §29.5's own Index test list is the explicit floor; 011/012/013's full suites are named regression gates. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `links` is sourced directly from `operations.References` rather than a second, parallel relationship computation (research.md #4) — consistency (FR-009) holds by construction; `ClassifyPath` is reused verbatim for indexing scope rather than duplicated (research.md #3); the `Store` interface keeps SQL entirely out of any future caller's view (§10.4). |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature never writes to any project artifact — its only writes are to its own disposable database, a boundary even stricter than 012's own read-only operations. |
| VIII. Safety by Construction | **Pass, newly exercised for this feature's own write path.** The database itself is the only thing this feature ever writes, and every write happens inside one transaction that either fully commits or fully rolls back (FR-007, research.md #9) — the same atomic-write discipline Principle VIII already requires of project-artifact writes, applied here to the index's own storage. |
| IX. Transparent, Machine-Readable Contracts | **Pass (N/A this feature).** No CLI command is added — Phase 8's own "internal context" command is the later, actual machine-readable surface this capability feeds into. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/014-sqlite-index/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md           # Phase 1 output (/speckit-plan command)
├── quickstart.md            # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── index.md               # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md    # /speckit-specify quality checklist
└── tasks.md                 # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # MODIFIED: + modernc.org/sqlite v1.39.0 (research.md #1)
├── cmd/misterspec/, internal/cli/         # unchanged — no CLI surface added
├── internal/
│   ├── project/, testutil/, example/        # example/ extended (+1 quickstart test)
│   ├── lock/, templates/                      # unchanged
│   ├── installer/, agents/, bootstrap/, tui/   # unchanged
│   ├── validation/, operations/                 # unchanged — reused, not modified
│   ├── artifacts/                                 # unchanged — reused, not modified
│   ├── ids/                                         # unchanged — reused, not modified
│   └── context/
│       └── index/
│           ├── doc.go                                 # NEW
│           ├── schema.go                                # NEW — DDL, PRAGMA user_version handling
│           ├── store.go                                   # NEW — Store interface, SearchResult/Link/SyncReport/SyncError types
│           ├── sqlite.go                                    # NEW — concrete SQLite-backed Store, Open/Close
│           ├── sync.go                                        # NEW — Sync/Rebuild orchestration
│           ├── search.go                                        # NEW — FTS5 query implementation
│           └── *_test.go                                          # NEW
└── kit/                                    # unchanged
```

**Structure Decision**: One new package, `internal/context/index` — no
existing package modified at all, the cleanest "zero blast radius" of
any feature so far (011 modified `wikilink.go`; 012 refactored
`classifyWikilink`; 013 modified `wikilink.go` again for `fencedLines`;
this feature touches none of them, only reading their already-published
functions). `internal/context` itself carries no other files yet
(research.md #2) — Phase 6's own collector is what will eventually add
them.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
