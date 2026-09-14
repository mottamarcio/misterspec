# Phase 0 Research: Disposable SQLite Index

All unknowns below were resolved by grounding
`docs/context-engine-implementation.md` §10-12, §29.5 against the actual
codebase (`internal/artifacts`, `internal/operations`, `internal/ids`,
`internal/lock`) rather than against the doc's own illustrative
snippets in isolation, per that document's own §32.1 instruction.

## 1. SQLite driver: pure Go, no CGO — `modernc.org/sqlite`, pinned

**Decision**: Use `modernc.org/sqlite` (a pure-Go, CGO-free SQLite
implementation with FTS5 support compiled in), pinned to `v1.39.0`.

**Rationale**: This project's own frozen constraints (Constitution
"Architecture Constraints": a single embedded-kit Go binary,
`go:embed`-based, no network access required to run) are fundamentally
incompatible with a CGO-based driver (`mattn/go-sqlite3`), which
requires a C toolchain at build time and complicates cross-compilation
— exactly the kind of portability regression this project has never
accepted anywhere else. `modernc.org/sqlite` needs nothing beyond `go
build` on any platform Go itself cross-compiles for, preserving the
single-static-binary property completely.

**Version pinning, following 010's own established method** (research,
not guesswork): queried `proxy.golang.org`'s per-version `.mod` files
directly for `modernc.org/sqlite` and its full transitive dependency
tree (`modernc.org/libc`, `modernc.org/fileutil`, `modernc.org/mathutil`,
`modernc.org/memory`, `golang.org/x/exp`, `github.com/google/pprof`,
`github.com/ncruces/go-strftime`, etc.). `v1.40.0` and later require `go
1.24.0`; `v1.37.0`–`v1.39.0` require exactly `go 1.23.0` — compatible
with this module's own `go 1.23.4` directive without any toolchain
bump. **Experimentally confirmed** in a scratch module (`go get
modernc.org/sqlite@v1.39.0`, `go build`): `go.mod`'s `go` line stayed at
`1.23.4` throughout, and the build succeeded with no C compiler
involved. `v1.39.0` is therefore the version to add.

**Alternatives considered**: `mattn/go-sqlite3` (rejected — CGO,
directly conflicting with this project's own portability commitment).
`modernc.org/sqlite@latest` without pinning (rejected — the exact
mistake 010 already made and fixed once; pin first, verify via the
proxy and a real build, exactly as 010's own contract now documents).

## 2. Package placement: `internal/context/index` only, nothing else yet

**Decision**: Create exactly one new package, `internal/context/index`
(`doc.go`, `schema.go`, `store.go`, `sqlite.go`, `sync.go`, `search.go`
+ tests). Do **not** create `internal/context`'s own top-level files
(`engine.go`, `request.go`, `collector.go`, `ranker.go`, `budget.go`,
`render.go`) yet — those belong to Phase 6-7's own context collector,
which does not exist yet and has no code to house.

**Rationale**: `docs/context-engine-implementation.md` §31's own
illustrative layout nests `index/` under `internal/context/`, but its
own §10.1 explicitly warns: "avoid introducing a broad generic indexing
subsystem unless a second concrete consumer appears." Creating the
*index* package now is justified — Phase 4 is exactly and only about
the index, with a real, immediate need — but creating the collector's
own empty scaffold files ahead of Phase 6 actually needing them would
be the precise anti-pattern the doc itself just warned against
(Constitution Principle IV, YAGNI). `internal/context` exists as a
directory purely because `internal/context/index` needs a parent; it
carries no code of its own yet.

**Alternatives considered**: Nesting the index directly under
`internal/artifacts` or `internal/operations` (rejected — this is
derived, disposable cache state fundamentally unlike either package's
own read-only-over-authoritative-files contract; it needs its own
clearly separate home, exactly as 010's `.misterspec/.lock` precedent
already established for a different kind of non-authoritative state).

## 3. "What gets indexed" — reusing `ArtifactType` exactly, not a new list

**Decision**: Sync walks the project's `cfg.ArtifactsDir` tree
(`ai/`), and for every `.md` file found, classifies it via
`artifacts.ClassifyPath` (already exported, from 001). A file that
`ClassifyPath` cannot classify is simply not an artifact this feature
indexes — skipped, not an error. Every one of `artifacts.ArtifactType`'s
9 defined values (Program, Feature, Spec, Plan, Tasks, Validation,
Knowledge, Learning, Constitution) is indexed — there is no separate,
new "eligible types" list to invent or maintain.

**Rationale**: `docs/context-engine-implementation.md` §12's own
indexing scope (Constitution, Knowledge, Learnings, Programs, Features,
Specs, Plans, Tasks, Validation) is, verbatim, exactly the full set
`ArtifactType` already enumerates — `ClassifyPath` already answers
"is this a real, canonical artifact, and which kind" with zero new
logic needed (Constitution Principle VI, DRY). Source code
(`*.go`/`*.ts`/etc., §12's own explicit exclusion) is automatically
excluded: `ClassifyPath` only ever matches the fixed canonical
filenames/prefixes real artifacts use, never an arbitrary source file
extension.

**Alternatives considered**: A new, index-specific "indexable types"
enum or config list (rejected — `ArtifactType` already *is* that list,
verified against the doc's own §12 line-for-line; a second one would be
exactly the kind of parallel vocabulary Constitution Principle VI
exists to prevent).

## 4. Populating `links`: reuse 012's `References` for the 5 ID-bearing types; semantic-only for the other 4

**Decision**: For an artifact of one of the five types 012 already
covers (Program, Feature, Spec, Knowledge, Learning), `links` rows are
populated by calling `operations.References` directly — both its
formal and semantic halves — and storing each entry keyed by that
artifact's own entity ID. For Plan, Tasks, Validation, and the
Constitution — types 012 deliberately excluded (012's own research.md
#1) — `links` receives no formal-relationship rows at all; only their
*content* is chunked and indexed for full-text search. Their own
wikilinks (if any) remain readable within their indexed chunk content,
just not materialized as separate graph edges in this MVP.

**Rationale**: FR-008/FR-009 promise the index-backed relationship
lookup is *consistent with* 012's own direct computation — sourcing
`links` directly from `operations.References` makes that true **by
construction**, not merely by parallel-and-hopefully-matching logic
(Constitution Principle VI). Extending formal-relationship semantics to
Plan/Tasks/Validation's own `for:` field is exactly the scope 012's own
research.md #1 explicitly deferred ("until a real caller needs it") —
this feature is about search and disposable caching, not about
reopening that boundary; doing so now would silently redraw a decision
a different, already-shipped feature made deliberately. FR-008/FR-009's
own guarantee is written to apply only where 012 itself has an opinion
to be consistent with — which is exactly these five types — so this
scope split introduces no contradiction with spec.md.

**Alternatives considered**: Inventing a path-based pseudo-ID so every
indexed type (including Plan/Tasks/Validation/Constitution) could
populate `links` symmetrically (rejected — the source document's own
schema (§10.3) declares `source_artifact_id`/`target_artifact_id` as
plain `TEXT NOT NULL`, implying real entity IDs, not a second
addressing scheme this feature would have to invent and maintain).

## 5. Sync cost: reuse over premature optimization

**Decision**: Accept `operations.References`'s own internal cost (a
fresh `ids.Scan` per call) for each of the five ID-bearing artifacts
processed during a sync — no batching or caching optimization in this
MVP.

**Rationale**: Constitution Principle IV cuts both ways: avoiding
speculative complexity applies as much to premature performance
optimization as to premature abstraction. At this project's own stated
scale (hundreds to low thousands of artifacts, `docs/context-engine-
implementation.md` §27), the resulting cost is the same order 012's own
`Backlinks` already accepts on every single call in production. A full
walk of the whole corpus is also the *rare* case in practice —
Phase 5's own incremental synchronization means steady-state syncs only
ever reprocess the handful of artifacts that actually changed.
Optimizing this ahead of any real measurement would be exactly the
premature complexity Principle IV warns against.

**Alternatives considered**: Bypassing `operations.References` and
reading `Metadata`/wikilinks directly during the walk to avoid the
redundant `Resolve` step (rejected — loses the "consistent by
construction" guarantee from Decision 4 for a performance gain with no
demonstrated need).

## 6. Database location and lifecycle

**Decision**: `.misterspec/cache/context.db`, created and populated
lazily by `Open` (which also creates `.misterspec/cache/` if missing).
Never authoritative, always safe to delete, never intended to be
committed to version control.

**Rationale**: `docs/context-engine-implementation.md` §10.2's own
explicit suggestion, directly under the existing `.misterspec/`
directory — the same convention `internal/lock`'s own `.misterspec/
.lock` already established for non-authoritative, disposable state
(Constitution Principle III). No new top-level convention introduced.

## 7. Schema versioning: `PRAGMA user_version`, not a metadata table

**Decision**: Use SQLite's built-in `PRAGMA user_version` integer to
tag the schema version. `Open` checks it against this package's own
current expected version; a mismatch (including a brand-new, empty
database) means the schema is (re)created from scratch — never
migrated in place.

**Rationale**: The index is explicitly disposable (Principle III) — it
never needs a real migration path, only "is this schema what I expect,
or do I start over." `PRAGMA user_version` is a single built-in integer
SQLite already provides for exactly this purpose, avoiding a bespoke
metadata table for a concern SQLite already solves (Constitution
Principle IV, KISS).

**Alternatives considered**: A dedicated `schema_meta` table (rejected
— strictly more code than a built-in pragma already covering the same
need).

## 8. FTS5 synchronization: explicit Go-driven writes, not triggers

**Decision**: `chunks_fts` is a standalone FTS5 virtual table (not
external-content-linked to `chunks`), with an `UNINDEXED` column
carrying the owning `chunks.id` back-reference. Every insert or delete
into `chunks` is paired with an explicit, corresponding insert or
delete into `chunks_fts` from Go code, inside the same transaction —
no SQL triggers.

**Rationale**: `docs/context-engine-implementation.md` §10.3 explicitly
allows either an external-content FTS5 table or "a simpler synchronized
table depending on complexity" — since this package (Decision 5)
already centrally orchestrates every write from Go during `Sync`,
adding the matching FTS statement at each of those same call sites is
simpler to read, test, and reason about than SQLite trigger semantics,
and keeps every write inside one explicit Go-level transaction
(Constitution Principle IV, KISS).

## 9. Transactional guarantee (FR-007)

**Decision**: A full `Sync` or `Rebuild` run is wrapped in exactly one
SQL transaction — `BEGIN` before the walk begins, `COMMIT` only after
every artifact has been processed (or skipped) and every stale entry
removed; any error aborts with `ROLLBACK`, leaving the database exactly
as it was before the call.

**Rationale**: The simplest possible way to satisfy "leave the index in
either its pre-run state or a fully correct post-run state — never a
partially-applied one" (FR-007, and §29.5's own explicit "transaction
rollback" test) — one transaction, one atomic outcome, no partial-commit
window to reason about. At this project's own stated scale, one
transaction covering a full sync is well within SQLite's own normal
operating envelope.

**Alternatives considered**: Per-artifact transactions with a separate
rollback/retry strategy for partial failure (rejected — more moving
parts for no benefit: FR-010's own "one bad artifact doesn't halt the
rest" is handled by catching and recording that artifact's own error
and continuing the walk, still inside the one outer transaction, rather
than by transaction boundaries).

## Summary of Go footprint

- `go.mod`/`go.sum`: + `modernc.org/sqlite v1.39.0` (Decision 1) and its
  transitive dependencies — the first new external dependency since
  010's TUI stack.
- `internal/context/index/{doc.go, schema.go, store.go, sqlite.go,
  sync.go, search.go}` (new), each with a matching `_test.go`.
- No change to `internal/artifacts`, `internal/validation`,
  `internal/operations`, or `internal/cli` — this feature only reads
  from them (`ReadBody`, `ParseDocument`, `Chunks`, `References`,
  `Fingerprint`, `ClassifyPath`, `ParseMetadata`), it adds no new call
  site *into* any of them requiring a change.
- No CLI surface (matching 013's own precedent) — Phase 8's own
  "internal context" command is the later, real CLI entry point for
  this capability.
