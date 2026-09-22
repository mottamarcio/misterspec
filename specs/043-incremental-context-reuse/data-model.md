# Data Model: Reutilização Incremental de Context Packs

All new types live in `internal/context` (pure computation) and
`internal/context/index` (the one new persisted table) — no new
package (research.md #1/#3).

## ConfigIdentity

The deterministic *configuration* inputs — every `PackIdentity`
component except the selected items themselves (spec Key Entity
"Identificador de Pacote-Base", research.md #2/#3). Hashed on its own
(`config_hash`) precisely so a live request's own freshly computed
value can be compared against a stored row's value without needing to
"invert" a hash (research.md #3's correction).

| Field | Type | Notes |
|---|---|---|
| `Target` | `string` | The resolved request's own `Target`. |
| `Intent` / `Task` / `Query` / `QueryMode` | `string` | Copied from `contextengine.Request`, exactly as resolved for this call. |
| `Budget` / `HardLimit` | `int` | The *resolved* values (after `DefaultBudget`/`DefaultHardLimit` substitution), never the raw nilable request fields — two calls that resolve to the same effective budget must hash identically regardless of whether one passed `--budget` explicitly. |
| `PreferSection` | `bool` | Copied from `contextengine.Request`. |
| `RankingVersion` / `ContextSchemaVersion` / `Estimator` | `int` / `int` / `string` | The command's own current constants/estimator name at call time. |

## PackItemIdentity

One selected item's own identity+content-fingerprint pair (reuses
`contextengine.Candidate`'s existing `(Path, StartLine, EndLine)`
identity convention, research.md #5).

| Field | Type | Notes |
|---|---|---|
| `Path` / `StartLine` / `EndLine` | `string` / `int` / `int` | Same identity `Candidate` already uses for deduplication. |
| `Fingerprint` | `string` | `contextengine.Fingerprint` of the item's own selected `Content` — unchanged means byte-identical content. |

## PackIdentity

`ConfigIdentity` plus the ordered selected-item list — the full set of
inputs `pack_id` is derived from.

| Field | Type | Notes |
|---|---|---|
| `Config` | `ConfigIdentity` | As above. |
| `Items` | `[]PackItemIdentity` | Every selected item's own identity+fingerprint, in the response's own final order. |

## ConfigHash / PackID

Both are plain `string`s of the form `"sha256:<hex>"`, computed with
`contextengine.Fingerprint`'s existing hashing helper (reused, not
duplicated) over a canonical serialization of their respective inputs:

- `ConfigHash` = hash of `ConfigIdentity` alone.
- `PackID` = hash of `ConfigHash` concatenated with the ordered
  `Items` list (research.md #2/#3) — so any single differing
  component, config or content, still changes `PackID`, while
  `ConfigHash` alone answers "is this call shaped the same way as the
  one that produced a given stored pack" independently of content.

## StoredPack

One row in the new `packs` table (`internal/context/index`) — the
persisted, disposable record a future `--base` lookup reads
(research.md #3).

| Column | Type | Notes |
|---|---|---|
| `pack_id` | `TEXT PRIMARY KEY` | The `PackID` string. |
| `config_hash` | `TEXT NOT NULL` | The `ConfigHash` string this pack was computed under — compared directly against a live call's own freshly computed `ConfigHash` (research.md #3/#4); this is what actually detects "the caller's request shape changed" (different `--intent`/`--budget`/ranking version/contract version/…), which the opaque `pack_id` alone cannot answer. |
| `target` | `TEXT NOT NULL` | For a fast existence/target-mismatch pre-check before touching `config_hash`. |
| `created_at` | `INTEGER NOT NULL` | Unix seconds — used only for the bounded eviction policy (research.md #3), never for correctness. |
| `items_json` | `TEXT NOT NULL` | JSON array of `{path, start_line, end_line, fingerprint, content}` — the base item list `DiffAgainstBase` compares the current selection against. |

## DiffEntry

One line of a diff response — reused across `"reuse"`, `"added"`, and
`"modified"` kinds (data-model presented positionally, in final order;
research.md #5).

| Field | Type | Notes |
|---|---|---|
| `Kind` | `"reuse" \| "added" \| "modified"` | |
| `BaseIndex` | `*int` | Non-nil only for `Kind == "reuse"` — the 0-based index into the named base pack's own stored item list to copy content from. |
| `Item` | `*PackageItem` | Non-nil only for `Kind == "added"`/`"modified"` — the item's own full current data (same shape `--mode package` already returns per item, 033). |

## RemovedEntry

One base identity absent from the current selection.

| Field | Type | Notes |
|---|---|---|
| `Path` / `StartLine` / `EndLine` | `string` / `int` / `int` | The removed item's own identity, from the base pack — no content (it no longer exists in the current selection). |

## PackDiff

The full diff computation result (spec Key Entity "Diferença de
Context Pack").

| Field | Type | Notes |
|---|---|---|
| `BasePackID` | `PackID` | The base actually used — echoed back so the caller can confirm which base a diff applies to. |
| `Entries` | `[]DiffEntry` | Every position of the *current* selection's own final order. |
| `Removed` | `[]RemovedEntry` | Every base identity no longer present. |

## RecoveryResult

What `--base` produces when the named base cannot be trusted
(research.md #4, spec Key Entity "Gatilho de Invalidação" made
concrete as the reason field below).

| Field | Type | Notes |
|---|---|---|
| `Recovered` | `bool` | Always `true` on this path — the explicit marker spec FR-005 requires. |
| `Reason` | `"unknown_base" \| "invalidated"` | `unknown_base`: no stored row for the given `pack_id`. `invalidated`: a stored row existed but the current call's own freshly computed `ConfigHash` no longer matches the row's stored `config_hash` (a different `--intent`/`--budget`/ranking version/contract version/estimator — spec FR-006, research.md #3/#4). A pure content change is never `invalidated` — it stays config-compatible and is instead reported as `"modified"`/`"added"`/`"removed"` diff entries. |
| `Items` | `[]PackageItem` | The full current package — identical to what a non-`--base` `--mode package` call would return. |

## ReuseDiagnostics

Per-call measurement (spec Key Entity "Medição de Reutilização",
research.md #6) — present whenever `--base` was supplied, on both the
diff and the recovery path (0 values on recovery, since nothing was
reused that call).

| Field | Type | Notes |
|---|---|---|
| `ItemsReused` | `int` | Count of `Kind == "reuse"` entries (0 on a `RecoveryResult`). |
| `ItemsSent` | `int` | Count of `"added"`+`"modified"` entries, or `len(Items)` on a `RecoveryResult`. |
| `TokensSavedEstimate` | `int` | The configured `Estimator`'s own estimate of what the reused items' content would have cost, had it been resent. |

## Reused, unmodified types (no new definitions)

- `contextengine.PackageItem` / `contextengine.Fingerprint` (033) —
  the source of every item's own content/fingerprint this feature
  compares.
- `contextengine.Request` / `Result` / `ResultItem` (015/016/035) —
  unchanged; this feature reads their already-computed output, never
  changes how selection/ranking/budgeting work.
- `index.Store` (014) — the same `Open`/`Sync`/`Close` lifecycle
  `internal context` already uses; `packs` is a sibling table, not a
  new store type.

## New, additive changes to existing contracts

- `internal/context/index`: `schemaVersion` bumped; `packs` table
  added (research.md #3).
- `misterspec internal context`: new `--base <pack_id>` flag
  (`--mode package` only); `contextSchemaVersion` bumped — every
  `--mode package` response (with or without `--base`) gains
  `pack_id`; a `--base` response additionally gains either `diff` +
  `reuse`, or `recovered` + `reason` + `reuse` (contracts/
  incremental-context-reuse-contract.md).

No existing exported signature changes; a call that never passes
`--base` sees byte-identical behavior except for the additive
`pack_id` field (Constitution Principle VII/IX — additive, not
breaking).
