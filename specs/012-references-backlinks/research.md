# Phase 0 Research: References and Backlinks

All unknowns below were resolved by grounding
`docs/context-engine-implementation.md` §5, §7, §29.3 against the actual
codebase (`internal/ids`, `internal/artifacts`, `internal/operations`,
`internal/validation`) rather than against the doc's own illustrative
snippets in isolation, per that document's own §32.1 instruction
("inspect the current `dev` branch... treat current code as
authoritative over illustrative snippets").

## 1. Entity-type scope: which artifacts participate in the graph

**Decision**: References/Backlinks operate on exactly the five entity
types `internal/validation` already validates as standalone entities —
Program, Feature, Spec, Knowledge, Learning. Task, Plan, Tasks,
Validation, and the Constitution are out of scope for this feature.

**Rationale**: Plan/Tasks/Validation have no `EntityType`/independent ID
at all (`internal/ids/types.go`'s own doc comment: "Plan, Tasks,
Validation, and the Constitution have no independent ID of their own...
and so have no EntityType value") — they cannot be a `references
<ID>`/`backlinks <ID>` query's *target* argument, since there is no ID
to name them by. Task has an `EntityType` and is independently
addressable (`operations.Resolve` already supports it), but has no
frontmatter of its own (`validatableTypes` in `internal/validation`
already excludes it for exactly this reason) — there is no `parent`/
`depends_on`/`supersedes` to report, and 011 never extended wikilink
*validation* to Task bodies either. Extending either side of the graph
to these types would need new discovery machinery (enumerating a
Spec's associated Plan/Tasks/Validation files, or deciding whether a
Task's own body participates) with no concrete, demonstrated need yet —
Constitution Principle IV (YAGNI). Keeping the same boundary 011 already
drew, rather than redrawing it per feature, is also more predictable for
any caller: "the reference graph" means the same five types everywhere
in the project.

**Alternatives considered**: Extending to Task (rejected — no
frontmatter to report a formal relationship from, and no validated body
to report a semantic one from, today). Extending formal relationships to
include Plan/Tasks/Validation's own `for:` field, surfaced only in
Backlinks (rejected — the spec's own first draft assumed this; grounding
it against the actual `internal/operations/create_artifact.go`
implementation showed it would require a second, different kind of
per-Spec file enumeration this feature doesn't otherwise need, for a
relationship kind no illustrative JSON in §7.1 actually shows; deferred
until a real caller needs it).

## 2. A shared, width-tolerant target-resolution helper

**Decision**: Add one new small exported function,
`ids.ResolveTarget(root string, cfg project.Configuration, raw string)
(EntityID, []string, error)`, composing `ParseAny` (width-tolerant,
type-inferring) and `Scan` exactly as `internal/validation/wikilinks.go`'s
existing unexported `classifyWikilink` already does inline. Refactor
`classifyWikilink` to call it instead of repeating the same two calls;
`internal/operations`'s new semantic-reference/backlink logic calls the
same function.

**Rationale**: Both this feature (a wikilink target that resolves to
zero, one, or many real artifacts) and 011's existing classification (the
exact same three-way outcome, just interpreted as a Finding instead of a
Reference/Backlink) need the identical composition — Constitution
Principle VI, DRY. `internal/ids` already owns both `ParseAny` and `Scan`
individually; a small composing function is the natural, minimal-surface
home for logic that is otherwise about to exist independently in two
different packages. This is the direct converse of 011's own
`canonicalFilename` precedent (duplicated between `internal/operations`
and `internal/validation` specifically *because* importing it would have
required a cycle) — here there is no cycle risk (`internal/operations`
already imports `internal/validation` via `status.go`), so the correct
move is the opposite one: share, don't duplicate.

**Alternatives considered**: Duplicating the two-call composition a
third time directly inside `internal/operations` (rejected — the exact
DRY violation Principle VI exists to prevent, with zero cycle-risk
justification this time). Exporting `classifyWikilink` itself from
`internal/validation` for `internal/operations` to call (rejected — its
return shape is a validation `Finding`, a concept `internal/operations`
has no business depending on; the shared part is genuinely just target
resolution, not Finding construction).

## 3. Ambiguous and broken/malformed wikilink targets

**Decision**: A wikilink counts as a graph edge (in either References or
Backlinks) only when it resolves, via `ids.ResolveTarget`, to *exactly
one* real artifact. Zero matches (broken) and more than one match
(ambiguous — a duplicate-ID situation) are both excluded, identically.

**Rationale**: An "ambiguous" wikilink target in 011's own model means
more than one physical file claims the *same* entity ID — not that the
ID could plausibly mean two different things. There is exactly one
logical `EntityID` either way; the ambiguity is about *which file*, a
question this feature has no principled way to answer that 011's own
validation (`ambiguous_wikilink`) doesn't already flag as the real
underlying problem (a duplicate ID). Symmetrically, querying
`references`/`backlinks` for an ambiguous ID as the target itself is
already impossible — `operations.Resolve` (reused per FR-008, see
Decision 4) already rejects an ambiguous target with
`ErrEntityAmbiguous` before this feature's own logic ever runs. Treating
"broken" and "ambiguous" identically (both simply omitted, never a
partial or best-guess entry) keeps the graph's contract simple: an edge
in the answer always means "this points at exactly one real thing."

**Alternatives considered**: Reporting one edge per ambiguous match
(rejected during drafting — spec.md's first draft assumed this; it
conflated "ambiguous ID" with "genuinely multiple valid meanings," which
is not what 011's classification actually represents, per the discussion
above).

## 4. Reusing existing operations rather than a parallel resolution path

**Decision**: `References` calls `operations.Inspect` (which already
composes `Resolve` + `ParseMetadata`, and already special-cases Task —
rejected afterward by this feature's own five-type scope check).
`Backlinks` calls `operations.Inspect` once for the queried target, then
iterates the five scoped types via `ids.Scan` and
`internal/operations/resolve.go`'s own existing unexported
`canonicalFilename` (same package, no duplication needed this time — a
direct call, not a copy). Both surface a genuinely unsupported or
unresolvable target via `operations.ErrInvalidTarget`,
`operations.ErrEntityNotFound`, and `operations.ErrEntityAmbiguous` —
the exact sentinels `internal/cli/internalcmd/errors.go`'s `classify`
already maps to `invalid_target` / `entity_not_found` / `entity_ambiguous`
(FR-008). No new sentinel and no `classify` change are needed at all.

**Rationale**: Constitution Principle VI (DRY) and Principle IX
(existing error-code contract). `Inspect` already does exactly the
"resolve, then read metadata" step this feature's own outgoing-reference
half needs; reusing it is strictly less code than re-deriving `Resolve`
+ `ParseMetadata` a second time. Reusing the exact same three sentinels
`inspect`/`children`/`parent` already use for an invalid or missing
target means a caller (an agent, or a test) needs to learn nothing new
about how this feature reports "that ID doesn't work" — direct evidence,
like 011's before it, that this project's generic CLI-adapter contract
keeps paying for itself.

**Alternatives considered**: A new, feature-specific "not a
referenceable type" error distinct from `operations.ErrInvalidTarget`
(rejected — `validation.ErrInvalidTarget` already established the
precedent of a second package reusing the same *name and meaning* for
its own scope boundary, both mapping to the same `invalid_target` CLI
code; this feature's own scope check follows that exact precedent using
`operations`'s own existing sentinel, since `References`/`Backlinks`
live in that same package).

## 5. Deterministic ordering

**Decision**: `References`' formal list preserves frontmatter
declaration order (parent, then `depends_on` entries in file order, then
`supersedes` entries in file order — already stable, since YAML list
order round-trips through `ParseMetadata` unchanged); its semantic list
preserves `ExtractWikiLinks`' own already-guaranteed document order.
Neither needs new sorting. `Backlinks` iterates the five scoped entity
types in a fixed order (Program, Feature, Spec, Knowledge, Learning —
the same order `internal/validation/validator.go`'s own
`projectEntityTypes` already uses), and within each type, ascending by
number (the same `sortedNumbers` pattern `validator.go` already
establishes) — giving a fully deterministic answer with no separate
sort step of its own, by construction rather than by a final sort call.

**Rationale**: FR-007. Reusing an already-proven ordering pattern
(`validator.go`'s own) rather than inventing a new one is itself a small
DRY/consistency win, even though the exact helper is duplicated locally
per package (matching the `canonicalFilename` precedent for a
genuinely tiny, package-local helper) rather than exported, since
`sortedNumbers` is unexported in `internal/validation` and this is not
a large enough piece of logic to justify promoting it to a shared
package on its own (Principle IV, YAGNI).

## 6. Backlinks' full-project scan cost

**Decision**: `Backlinks` scans every artifact of the five scoped types
directly (via `ids.Scan` + `ParseMetadata` + `ReadBody` +
`ExtractWikiLinks` per candidate source) — no caching, no index, no
persisted reverse-reference table.

**Rationale**: `docs/context-engine-implementation.md` §7.2's own
explicit requirement: "the implementation may initially scan through the
existing artifact model... correctness MUST NOT depend on the cache
existing." Constitution Principle III (filesystem as sole source of
truth) forbids a persisted reverse index as authoritative state; a
disposable SQLite index is explicitly a *later* phase (Phase 4) this
feature must not anticipate (Principle IV, YAGNI — no speculative
optimization ahead of a demonstrated need). At this project's own stated
scale (hundreds to low thousands of artifacts, `docs/context-engine-
implementation.md` §27), a full scan is the same order of cost
`ValidateProject` already pays on every run today.

**Alternatives considered**: Building an in-memory reverse-reference map
once per process and reusing it across multiple queries (rejected —
`References`/`Backlinks` are each single, independent, stateless calls
today, matching every other `internal/operations` function's own
contract; introducing shared mutable state across calls would be new
architecture this feature does not need).

## 7. CLI surface and JSON shape

**Decision**: Two new commands, `misterspec internal references <id>`
and `misterspec internal backlinks <id>`, each a thin
`internal/cli/internalcmd` wrapper (argument parsing + JSON shaping
only, per every existing command's own `spec.md FR-009`-style contract)
around the new `operations.References`/`operations.Backlinks`. JSON
shape follows `docs/context-engine-implementation.md` §7.1's own
illustrative example exactly, minus the `for` relation (Decision 1):

```json
{"ok":true,"target":"SPEC-014","references":{"formal":[...],"semantic":[...]}}
{"ok":true,"target":"SPEC-014","backlinks":{"formal":[...],"semantic":[...]}}
```

**Rationale**: Matches this project's already-established "generic
adapter over a deterministic core" contract (Constitution Principle IX)
exactly — see contracts/references-backlinks.md for the full shape.

**Alternatives considered**: A single flat list mixing formal and
semantic with a `kind` field per entry (rejected — the doc's own
illustrative JSON already groups them into separate `formal`/`semantic`
arrays, and FR-003 requires them "clearly distinguishable," which
grouping satisfies more directly than a per-entry field a caller would
otherwise have to filter on).

## Summary of Go footprint

- `internal/ids/scan.go` (or a new small file): + `ResolveTarget`
  (exported), + regression test.
- `internal/validation/wikilinks.go`: refactor `classifyWikilink` to
  call `ids.ResolveTarget` instead of its own inline `ParseAny`+`Scan` —
  behavior-identical; 011's full existing suite is this refactor's own
  regression gate, exactly like 011's own `splitFrontmatter` widening
  was gated by `parser_test.go`.
- `internal/operations/{references.go, references_test.go, backlinks.go,
  backlinks_test.go}` (new).
- `internal/cli/internalcmd/{references.go, backlinks.go}` (new);
  `internal/cli/internal.go` (+2 lines, wiring); zero changes to
  `errors.go`'s `classify` (Decision 4).
- `internal/example`: one new end-to-end quickstart test.
- No new external dependency. No new package.
