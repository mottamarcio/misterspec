# Phase 2 Contracts: References and Backlinks

**Reconciled against the actual implementation (T020)** — zero drift.
Every signature below (`ids.ResolveTarget`; `ReferenceEntry`/
`ReferencesResult`/`References`; `BacklinkEntry`/`BacklinksResult`/
`Backlinks`; both CLI commands' JSON shapes, including field names
`target`/`source`/`relation`) matches what shipped, verbatim. Three
unexported helpers introduced during implementation
(`formalReferences`, `matchingFormalBacklinks`,
`sortedScanNumbers` — plus `renderReferenceEntries`/
`renderBacklinkEntries` in `internal/cli/internalcmd`) are internal
decomposition detail, not part of this feature's exported/documented
surface. `classifyWikilink`'s refactor to call `ids.ResolveTarget`
(research.md #2) is behavior-identical, proven by 011's own full
existing suite re-run unmodified — no drift in `internal/validation`'s
own already-published contract either.

## `internal/ids` (extended)

```go
package ids

// ResolveTarget parses raw via ParseAny (width-tolerant, type-inferring)
// then Scans for it, returning every matching artifact path — zero
// (unresolved), one (resolved), or more (ambiguous — a duplicate ID).
// Only a genuinely malformed raw string is an error.
func ResolveTarget(root string, cfg project.Configuration, raw string) (EntityID, []string, error)
```

## `internal/operations` (extended)

```go
package operations

// ReferenceEntry is one outgoing relationship from a queried artifact.
type ReferenceEntry struct {
    Relation string // "parent" | "depends_on" | "supersedes" | "wikilink"
    Target   ids.EntityID
}

// ReferencesResult groups a queried artifact's outgoing relationships,
// each list in its own deterministic order (declaration/document order).
type ReferencesResult struct {
    Formal   []ReferenceEntry
    Semantic []ReferenceEntry
}

// References reports every outgoing relationship of rawID — every
// formal one (parent, depends_on, supersedes) exactly as declared in
// frontmatter, and every semantic one (a wikilink) that resolves to
// exactly one real artifact. rawID must name one of Program, Feature,
// Spec, Knowledge, or Learning — any other type returns
// ErrInvalidTarget, the same sentinel Resolve/Inspect already use.
func References(root string, cfg project.Configuration, rawID string) (ReferencesResult, error)

// BacklinkEntry is one incoming relationship into a queried artifact.
type BacklinkEntry struct {
    Relation string // "parent" | "depends_on" | "supersedes" | "wikilink"
    Source   ids.EntityID
}

// BacklinksResult groups a queried artifact's incoming relationships,
// discovered by scanning every artifact of the five scoped types, in a
// fixed deterministic type-then-number order.
type BacklinksResult struct {
    Formal   []BacklinkEntry
    Semantic []BacklinkEntry
}

// Backlinks reports every other artifact that formally or semantically
// references rawID. Same target-type scope and error behavior as
// References. Computed directly from filesystem state on every call —
// never depends on any index or cache existing (docs/context-engine-
// implementation.md §7.2).
func Backlinks(root string, cfg project.Configuration, rawID string) (BacklinksResult, error)
```

**Guarantees**:
- `References`/`Backlinks` are strictly read-only — neither writes, nor
  triggers any write, to any file (FR-010).
- A broken, malformed, or ambiguous wikilink target never contributes an
  edge in either direction (FR-002, FR-009, data-model.md's resolution
  table) — a graph edge always means "resolves to exactly one real
  artifact."
- A formal relationship is always reported exactly as declared, with no
  existence check of its own — that integrity question remains
  004-structural-validation's job (`missing_parent`,
  `unresolved_dependency`), never duplicated here.
- An artifact with zero outgoing or zero incoming relationships returns
  a well-formed, empty (non-nil-slice) result — never an error (FR-006).
- Both functions return identical output for identical project state
  across repeated calls (FR-007) — no hidden mutable state between
  calls.

## `internal/cli/internalcmd` (extended)

Two new commands, wired into `internal/cli/internal.go` alongside the
existing ten. No change to `errors.go`'s `classify` — both commands
reuse `operations.ErrInvalidTarget`/`ErrEntityNotFound`/
`ErrEntityAmbiguous`, already mapped to `invalid_target` (2),
`entity_not_found` (3), `entity_ambiguous` (3).

```text
misterspec internal references <id> [--dir <path>]
misterspec internal backlinks <id> [--dir <path>]
```

Success (`references`):

```json
{"ok":true,"target":"SPEC-014","references":{"formal":[{"relation":"parent","target":"FEAT-004"},{"relation":"depends_on","target":"SPEC-011"}],"semantic":[{"relation":"wikilink","target":"KNOW-003"}]}}
```

Success (`backlinks`):

```json
{"ok":true,"target":"SPEC-014","backlinks":{"formal":[{"relation":"depends_on","source":"SPEC-020"}],"semantic":[{"relation":"wikilink","source":"SPEC-025"}]}}
```

Failure (either command, same shape every existing internal command
already uses):

```json
{"ok":false,"error":{"code":"entity_not_found","message":"operations: entity not found: SPEC-999"}}
```

`target`/`source` are rendered via `ids.EntityID.String()` (e.g.
`"SPEC-014"`), matching every existing command's own ID-rendering
convention (`idStrings` in `inspect.go`). `formal`/`semantic` are always
present, non-null arrays (`[]`, never `null`, when empty).

## Cross-cutting: no change to any existing contract

`internal/validation`'s exported surface (`ValidateProject`,
`ValidateEntity`, the Finding codes) is unchanged — `classifyWikilink`'s
internal refactor to call `ids.ResolveTarget` (research.md #2) is
behavior-identical, proven by 011's own full existing suite re-run
unmodified. `internal/cli/internalcmd/errors.go`'s `classify` function
is unchanged verbatim (research.md #4) — the strongest evidence yet,
after 011's own zero-CLI-change precedent, that this project's generic
adapter/deterministic-core split keeps paying for itself.
