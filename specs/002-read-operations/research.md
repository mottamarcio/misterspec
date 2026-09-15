# Phase 0 Research: Read-Only Deterministic Operations

As with 001-core-foundation, the frozen envelope
(`docs/architecture-specification.md` §9, §32, §70 and the ratified
constitution) leaves no open `NEEDS CLARIFICATION`. This document records
the specific decisions made within that envelope — several of which
extend already-shipped `001-core-foundation` code in small, additive,
backward-compatible ways once this feature's real requirements exposed a
gap in it. Each extension is called out explicitly, with why it was
necessary rather than speculative.

## Decision: New package `internal/operations`

- **Decision**: One new package, `internal/operations`, with one file per
  operation (`resolve.go`, `inspect.go`, `parent.go`, `children.go`,
  `inventory.go`, `fingerprint.go`), per
  `docs/architecture-specification.md` §32's own layout.
- **Rationale**: Matches the frozen package layout; keeps each operation's
  SRP boundary clean (Constitution Principle VI) without introducing any
  new architectural layer.
- **Alternatives considered**: Folding these into `internal/artifacts` or
  `internal/ids` directly — rejected. Those packages are pure
  computation/parsing primitives; `operations` is the first layer that
  *composes* them into the outcomes an agent actually asks for (Constitution
  Principle I's semantic/deterministic boundary is about the agent vs. the
  binary, not about which internal package does what — but SRP still
  argues for keeping "compose primitives into an answer" separate from
  "the primitives themselves").

## Decision: Extend `ids.ScanResult` with per-ID locations

- **Decision**: Add a `Paths map[int][]string` field to `ids.ScanResult`
  (number → every claiming path, relative to root) — populated for every
  number found, not just duplicates. `IDs` and `Duplicates` are unchanged.
- **Rationale**: `Resolve` needs the path of a *single* match, not only
  the paths of a *duplicate* match — 001-core-foundation's `ScanResult`
  only ever recorded paths for numbers claimed more than once
  (`Duplicates`), because nothing consumed a single match's path before
  now. This is the first real consumer that needs it.
- **Alternatives considered**: Re-scanning the filesystem a second time,
  specifically for the one ID being resolved — rejected as duplicated,
  slower logic doing what `Scan` already does internally. Making `Resolve`
  reimplement its own directory walk — rejected for the same reason
  (Constitution Principle VI, DRY). This field addition is purely additive
  — 001-core-foundation's existing tests (`ids.NextID(result.IDs, ...)`,
  `result.Duplicates`) are unaffected and must still pass unmodified.

## Decision: Promote raw-ID-string parsing into `ids.ParseAny`

- **Decision**: Add `ids.ParseAny(raw string) (EntityID, error)` — parses
  a raw ID string like `"SPEC-014"` by deriving both its `EntityType`
  (from the prefix, via the already-exported `ids.TypeForPrefix`) and its
  width (from the suffix length) from the string itself, without the
  caller needing to already know the type. Refactor
  `internal/artifacts/parser.go`'s private `parseFieldID` to call it
  instead of duplicating the same logic.
- **Rationale**: Every one of `resolve`/`inspect`/`parent`/`children`
  takes a bare ID string as its primary input (mirroring
  `misterspec internal resolve SPEC-014`'s eventual CLI shape) and needs
  exactly this parsing. `artifacts.ParseMetadata` already needed the same
  logic privately for frontmatter fields (`id`, `parent`, `depends_on`).
  Two independent implementations of the same parsing would violate
  Constitution Principle VI (DRY) the moment this feature landed.
- **Alternatives considered**: Leaving `parseFieldID` private and copying
  it into `operations` — rejected, exactly the duplication DRY forbids.

## Decision: Export the path-containment check from `internal/artifacts`

- **Decision**: Export `artifacts`'s existing private `relativeWithinRoot`
  as `artifacts.RelativeWithinRoot(root, path string) (string, error)`,
  reused by `operations.Inventory` and `operations.Fingerprint` for
  FR-013's traversal guarantee, instead of re-implementing the same check.
- **Rationale**: This exact containment logic was already written, tested
  (`TestResolvePath_RejectsTraversalOutsideRoot`,
  `TestClassifyPath_RejectsTraversalOutsideRoot`), and proven in
  001-core-foundation. Constitution Principle VI (DRY) and Principle VIII
  (safety by construction) both argue for one proven implementation, not
  a second one in `operations` that could drift from it.
- **Alternatives considered**: A parallel containment check in
  `operations` — rejected as both duplicative and a safety risk (two
  implementations can disagree; one has already been tested against
  traversal inputs).

## Decision: `Children` reuses `Scan` + a path-prefix filter, not a new scan variant

- **Decision**: `Children(root, cfg, parentRawID, filterType)` resolves the
  parent to its directory, calls `ids.Scan` for the candidate child type
  (which now carries `Paths` per the extension above), and keeps only
  entries whose path is nested under the parent's directory. No new
  `ids.Scan` variant is added.
- **Rationale**: The project's structural nesting (Program → Feature →
  Spec) is exactly filesystem nesting — a prefix check on the path
  `Scan` already computed is sufficient and requires no new primitive.
  Adding a parent-scoped scan function now, before a second caller needs
  it, would be exactly the kind of speculative addition Constitution
  Principle IV rules out.
- **Alternatives considered**: A dedicated `ScanUnder(parent, childType)`
  function in `ids` — rejected for now as unjustified surface-area growth;
  revisit only if a second, genuinely different consumer needs it.

## Decision: Task metadata scope for `Inspect`/`Parent`

- **Decision**: For a Task ID, `Inspect` returns only `ID`, `Status`
  (derived from its Markdown checkbox — `- [ ]` = pending, `- [x]`/`- [X]`
  = complete) and `Parent` (the Spec named by its owning `tasks.md`'s own
  `for:` frontmatter field). `DependsOn`/`Supersedes` are left empty — not
  applicable to Task, per spec.md FR-004's "as applicable to its type."
  Extracting a Task's `Requirements:`/`Depends on:`/`Scope:`/
  `Verification:` body lines is explicitly out of scope for this feature.
- **Rationale**: Those body fields are how one Task references *Spec
  requirement markers* (`SPEC-001:R1`) and *other Tasks* — deterministically
  extracting and validating cross-references is the `references` operation
  (`docs/architecture-specification.md` §15), a distinct, later feature.
  Building it here, before its own spec exists, would be scope creep past
  what 002-read-operations committed to.
- **Alternatives considered**: Fully parsing every Task body field now —
  rejected as speculative scope expansion (Constitution Principle IV).

## Decision: Extend `Metadata` with a `For *ids.EntityID` field

- **Decision**: Add a `For *ids.EntityID` field to
  `internal/artifacts.Metadata`, populated from a `for:` frontmatter key
  (Plan/Tasks/Validation's own schema, `docs/architecture-specification.md`
  §28-30) the same way `Parent` already is from a `parent:` key.
- **Rationale**: `Parent(Task)` needs this to identify the owning Spec.
  Plan and Validation artifacts declare the identical `for:` field, so
  this is a real, shared, already-schema'd need — not speculative.
- **Alternatives considered**: A special-cased raw-YAML read inside
  `operations`, bypassing `artifacts.ParseMetadata` — rejected as
  duplicating frontmatter-extraction logic outside the package that owns
  it (DRY).

## Decision: Streaming SHA-256 fingerprinting

- **Decision**: `Fingerprint` opens the file and streams it through
  `io.Copy` into a `crypto/sha256` hash, rather than reading the whole
  file into memory first.
- **Rationale**: `docs/architecture-specification.md` §13 explicitly notes
  fingerprinting must "support any file type" without understanding it —
  Raw sources include PDFs, which can be large; streaming avoids an
  unnecessary full-file memory allocation with no added complexity.
- **Alternatives considered**: `os.ReadFile` + `sha256.Sum256` — rejected
  only on the memory-efficiency point for large files; otherwise
  equivalent and simpler, but the streaming form costs nothing extra here.

## Decision: Ambiguous-ID errors carry every matching location

- **Decision**: `Resolve` (and anything built on it) reports an ambiguous
  ID as a typed error value (not a bare sentinel) carrying the raw ID and
  every claiming path, e.g. `*operations.AmbiguousIDError{ID, Locations
  []string}`, wrapping a `operations.ErrEntityAmbiguous` sentinel for
  `errors.Is` checks.
- **Rationale**: FR-003 requires naming every matching location, not just
  signaling "ambiguous" — a bare sentinel error would lose that data. This
  mirrors 001-core-foundation's `project.ConfigError` pattern (a typed
  error wrapping a sentinel) already proven in this codebase.
- **Alternatives considered**: A bare `errors.New("ambiguous")` — rejected,
  loses the location data FR-003 requires the caller to have.

## Output

All unknowns resolved. Every extension to 001-core-foundation code above
is additive (new field, new exported function, or refactor-to-shared-helper)
and must be verified not to break `001-core-foundation`'s existing test
suite as part of this feature's own test run.
