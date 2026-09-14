# Phase 1 Data Model: Disposable SQLite Index

Entities extracted from `spec.md`'s Key Entities plus
`docs/context-engine-implementation.md` §10.3-10.4. This feature adds
one new package, `internal/context/index`, and one new external
dependency (`modernc.org/sqlite`) — no change to any existing package.

## SQLite schema (`schema.go`)

```sql
CREATE TABLE documents (
    id            INTEGER PRIMARY KEY,
    path          TEXT NOT NULL UNIQUE,
    artifact_id   TEXT,           -- entity ID string, or NULL for a
                                   -- non-ID-bearing type (Plan/Tasks/
                                   -- Validation/Constitution)
    artifact_type TEXT NOT NULL,  -- artifacts.ArtifactType.String()
    title         TEXT,
    fingerprint   TEXT NOT NULL,  -- operations.FileFingerprint.String()
    indexed_at    INTEGER NOT NULL
);

CREATE TABLE chunks (
    id             INTEGER PRIMARY KEY,
    document_id    INTEGER NOT NULL REFERENCES documents(id),
    heading        TEXT,
    content        TEXT NOT NULL,
    start_line     INTEGER NOT NULL,
    end_line       INTEGER NOT NULL,
    token_estimate INTEGER NOT NULL
);

CREATE VIRTUAL TABLE chunks_fts USING fts5(
    heading,
    content,
    chunk_id UNINDEXED
);

CREATE TABLE links (
    id                 INTEGER PRIMARY KEY,
    source_artifact_id TEXT NOT NULL,
    target_artifact_id TEXT NOT NULL,
    relation           TEXT NOT NULL
);
```

`PRAGMA user_version` holds the schema version (research.md #7) — no
separate metadata table.

**Validation rules**:
- `documents.path` is unique — one row per indexed file.
- `chunks.document_id` always references a live `documents` row within
  the same transaction; deleting a document's row also deletes its own
  `chunks` (and matching `chunks_fts`) rows.
- `links` rows exist only for artifacts of the five types
  `operations.References` already covers (research.md #4) — a
  Plan/Tasks/Validation/Constitution document contributes chunks but no
  `links` rows.

## `internal/context/index.Store` (new, interface)

| Symbol | Notes |
|---|---|
| `Store` | `Sync`, `Rebuild`, `Search`, `Outgoing`, `Incoming`, `Close` — no SQL type ever crosses this interface (docs/context-engine-implementation.md §10.4). |
| `Open(path string) (Store, error)` | Creates `path`'s parent directory and the database file if missing; creates or recreates the schema per `PRAGMA user_version` (research.md #7). |

```go
type Store interface {
    Sync(root string, cfg project.Configuration) (SyncReport, error)
    Rebuild(root string, cfg project.Configuration) (SyncReport, error)
    Search(query string, limit int) ([]SearchResult, error)
    Outgoing(id string) ([]Link, error)
    Incoming(id string) ([]Link, error)
    Close() error
}
```

## `internal/context/index.SyncReport` / `SyncError` (new)

| Field | Type | Notes |
|---|---|---|
| `Indexed` | int | Newly-indexed artifacts this run. |
| `Updated` | int | Previously-indexed artifacts whose content changed. |
| `Skipped` | int | Previously-indexed artifacts whose fingerprint is unchanged (FR-004). |
| `Removed` | int | Previously-indexed artifacts no longer present on disk (FR-004). |
| `Errors` | `[]SyncError` | Artifacts that failed to read/structure — recorded, not fatal (FR-010). |

```go
type SyncError struct {
    Path string
    Err  error
}
```

## `internal/context/index.SearchResult` / `Link` (new)

| Field | Type | Notes |
|---|---|---|
| `SearchResult.Path` | string | The originating document's path. |
| `SearchResult.Heading` | string | The matching chunk's heading. |
| `SearchResult.Content` | string | The matching chunk's content. |
| `SearchResult.StartLine`/`EndLine` | int | Copied from the originating `Chunk` (013). |
| `SearchResult.Rank` | float64 | FTS5's own `bm25()` score — lower is more relevant, per FTS5's own convention. |
| `Link.Relation` | string | `"parent"`/`"depends_on"`/`"supersedes"`/`"wikilink"` — exactly `operations.ReferenceEntry`/`BacklinkEntry`'s own vocabulary (012), reused verbatim. |
| `Link.Source` / `Link.Target` | string | Entity ID strings. |

## Sync algorithm (`sync.go`)

```text
Sync/Rebuild(root, cfg):
  BEGIN transaction
  (Rebuild only: delete every row from documents, chunks, chunks_fts, links)
  walk root/cfg.ArtifactsDir for every ".md" file           [research.md #3]
    classify via artifacts.ClassifyPath — skip if unclassifiable
    compute current fingerprint via operations.Fingerprint
    compare against documents.fingerprint (by path)
      unchanged  → Skipped++
      new        → index it, Indexed++
      changed    → delete its old chunks/chunks_fts/links rows, re-index, Updated++
      (on any per-artifact read/parse error: record SyncError, continue — FR-010)
  for every previously-indexed path NOT found in this walk:
    delete its documents/chunks/chunks_fts/links rows, Removed++
  COMMIT (or ROLLBACK on any transaction-level failure — FR-007)
  return SyncReport
```

Indexing one artifact:

```text
index(path, artifactType):
  meta   := artifacts.ParseMetadata(path)
  body   := artifacts.ReadBody(path)
  doc    := artifacts.ParseDocument(body)
  chunks := artifacts.Chunks(path, doc)
  insert documents row (path, meta.ID, artifactType, fingerprint)
  for each chunk: insert chunks row + matching chunks_fts row
  if artifactType.HasEntityID():                              [research.md #4]
    refs := operations.References(root, cfg, meta.ID.String())
    for each formal/semantic entry in refs: insert links row
  # else (Plan/Tasks/Validation/Constitution): no links rows
```

## State / Flow Summary

```text
project files (ai/...)
      ↓ artifacts.ClassifyPath + operations.Fingerprint         [Sync, US1/US2]
new / changed / unchanged / deleted, per path
      ↓ artifacts.ParseMetadata / ReadBody / ParseDocument / Chunks (011/013)
documents + chunks + chunks_fts rows
      ↓ operations.References (012, five ID-bearing types only)
links rows
      ↓
Search(query) → []SearchResult                                  [US1]
Outgoing(id) / Incoming(id) → []Link, consistent with 012        [US3]
```
