# Contract: Incremental Context Pack Reuse (`internal context --base`)

This documents the additive changes this feature makes to the existing
`misterspec internal context` command (033-context-pack-output-contract,
035-context-budget-accuracy). No existing field changes meaning; a
call that never passes `--base` is byte-identical except for the new
`pack_id` field in `--mode package` responses.

## 1. `internal/context` (extended)

```go
package contextengine

// ConfigHash / PackID are both "sha256:<hex>", reusing Fingerprint's
// own hashing convention (data-model.md "ConfigHash / PackID",
// research.md #2/#3).
type ConfigHash string
type PackID string

// ConfigIdentity is every PackIdentity input except the selected
// items — hashed on its own so a live call's own freshly computed
// ConfigHash can be compared directly against a stored row's
// config_hash, without needing to invert a hash (data-model.md,
// research.md #3's correction).
type ConfigIdentity struct {
	Target               string
	Intent               string
	Task                 string
	Query                string
	QueryMode            string
	Budget               int // resolved, not raw nilable request field
	HardLimit            int // resolved
	PreferSection        bool
	RankingVersion       int
	ContextSchemaVersion int
	Estimator            string
}

// PackItemIdentity is one selected item's own identity+fingerprint
// (data-model.md; reuses Candidate's (Path, StartLine, EndLine)
// identity convention, research.md #5).
type PackItemIdentity struct {
	Path        string
	StartLine   int
	EndLine     int
	Fingerprint string
}

// PackIdentity is ConfigIdentity plus the ordered selected-item list
// (data-model.md).
type PackIdentity struct {
	Config ConfigIdentity
	Items  []PackItemIdentity
}

// ComputeConfigHash derives cfg's own deterministic ConfigHash.
func ComputeConfigHash(cfg ConfigIdentity) ConfigHash

// ComputePackID derives id's own deterministic PackID — hash of
// ComputeConfigHash(id.Config) concatenated with id.Items.
func ComputePackID(id PackIdentity) PackID

// DiffEntry / RemovedEntry / PackDiff / RecoveryResult / ReuseDiagnostics
// — see data-model.md for full field tables.

// DiffAgainstBase compares current (the freshly computed
// []PackageItem for this call) against base (a StoredPack's own
// deserialized items), returning the ordered PackDiff data-model.md
// describes — pure, no I/O (research.md #5).
func DiffAgainstBase(current []PackageItem, base []index.StoredPackItem) PackDiff
```

## 2. `internal/context/index` (extended)

```go
package index

// StoredPackItem is one persisted item inside a StoredPack's own
// items_json (data-model.md "StoredPack").
type StoredPackItem struct {
	Path        string
	StartLine   int
	EndLine     int
	Fingerprint string
	Content     string
}

// StoredPack is one packs table row (data-model.md).
type StoredPack struct {
	PackID     string
	ConfigHash string
	Target     string
	CreatedAt  int64
	Items      []StoredPackItem
}

// Store (extended) — two new methods on the existing interface
// (store.go), implemented on the existing unexported *sqliteStore
// (sqlite.go), the same way every other Store method already is; no
// new concrete type.
type Store interface {
	// ... existing methods (Sync, Rebuild, Search, SearchAdvanced,
	// Outgoing, Incoming, Close) unchanged ...

	// SavePack upserts pack into the packs table, then evicts the
	// oldest rows beyond a fixed cap (research.md #3) — never fails
	// the caller's own response if eviction itself has nothing to do
	// (an empty/under-cap table).
	SavePack(pack StoredPack) error

	// LookupPack returns the stored pack for packID, and found ==
	// false (never an error) when no such row exists — the same "bool
	// separate from error" convention CommitsSinceFileAdded/HeadCommit
	// already use elsewhere in this codebase, applied here for "no
	// row" being a perfectly normal outcome (research.md #4), not a
	// failure. The caller (context.go) is responsible for comparing
	// the returned pack's own ConfigHash against the current call's
	// freshly computed ConfigHash before treating it as diffable
	// (research.md #4) — LookupPack itself does no validation, only
	// retrieval.
	LookupPack(packID string) (pack StoredPack, found bool, err error)
}
```

`schemaVersion` (currently 3) is bumped to 4; `packs` joins the
existing `dropStatements`/`schemaStatements` pair — no separate
migration path (data-model.md, research.md #3).

## 3. CLI: `misterspec internal context <id> --mode package [--base <pack_id>]`

`--base` is only valid alongside `--mode package`; combined with
`--mode manifest`/`markdown`, or with `--render`, it fails with
`invalid_argument` — mirroring the existing `--mode package` + `--render`
rejection precedent.

### Success envelope — first `--mode package` call (no `--base`)

Identical to today's `--mode package` shape, plus one new field:

```json
{
  "ok": true,
  "context": {
    "schema_version": 6,
    "target": "SPEC-014",
    "...": "... unchanged existing fields ...",
    "pack_id": "sha256:3f9a...",
    "items": [ { "path": "...", "content": "...", "fingerprint": "sha256:...", "...": "..." } ]
  }
}
```

### Success envelope — repeat call with `--base` (content unchanged)

```json
{
  "ok": true,
  "context": {
    "schema_version": 6,
    "target": "SPEC-014",
    "pack_id": "sha256:3f9a...",
    "diff": {
      "base_pack_id": "sha256:2a11...",
      "entries": [],
      "removed": []
    },
    "reuse": {"items_reused": 9, "items_sent": 0, "tokens_saved_estimate": 589}
  }
}
```

### Success envelope — repeat call with `--base` (one item changed)

```json
{
  "ok": true,
  "context": {
    "pack_id": "sha256:7bd1...",
    "diff": {
      "base_pack_id": "sha256:2a11...",
      "entries": [
        {"kind": "reuse", "base_index": 0},
        {"kind": "modified", "item": {"path": "ai/.../SPEC-014/spec.md", "content": "### R1 — ...", "fingerprint": "sha256:9e21...", "...": "..."}},
        {"kind": "reuse", "base_index": 2}
      ],
      "removed": []
    },
    "reuse": {"items_reused": 2, "items_sent": 1, "tokens_saved_estimate": 214}
  }
}
```

### Success envelope — `--base` with an unknown/invalidated pack

```json
{
  "ok": true,
  "context": {
    "pack_id": "sha256:af02...",
    "recovered": true,
    "reason": "unknown_base",
    "items": [ "... full package, same shape as a --base-less call ..." ],
    "reuse": {"items_reused": 0, "items_sent": 9, "tokens_saved_estimate": 0}
  }
}
```

`reason` is `"unknown_base"` (no stored row for the given `pack_id`) or
`"invalidated"` (a stored row existed, but the request's own freshly
recomputed identity — content, config, or contract version — no longer
matches it, research.md #4).

### Error envelope

Reuses `{"ok": false, "error": {"code", "message"}}`. No new error
code: `--base` combined with an incompatible `--mode`/`--render`
classifies as the existing `invalid_argument`.

## 4. Behavioral guarantees this contract makes (traceable to spec FRs)

- FR-002/FR-003: applying a returned `diff` to its own `base_pack_id`
  (substituting each `"reuse"` entry with that index's own stored item,
  in the response's own listed order) reproduces byte-identical output
  to a direct, `--base`-less `--mode package` call for the same target
  and configuration.
- FR-004/FR-005: `recovered: true` is present on every response where
  `--base` could not be honored — never a silently-incomplete `diff`.
- FR-006: a `pack_id` mismatch on any input dimension (content, config,
  or `contextSchemaVersion`) always falls onto the recovery path, never
  a diff computed against stale inputs.
- FR-009/FR-010: `reuse` is present on every `--base` call, whether diff
  or recovery, and never asserts a dollar/provider-cache saving — only
  a locally measured item/token count (research.md #6/#7).
