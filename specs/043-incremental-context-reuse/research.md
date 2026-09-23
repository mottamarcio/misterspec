# Research: Reutilização Incremental de Context Packs

Input: `specs/043-incremental-context-reuse/spec.md`. This feature
extends the existing `internal context` command and its underlying
`contextengine`/`index` packages rather than introducing a new
command or a second content-selection pipeline — the goal is reusing
what a caller already has, not recomputing selection differently.

## Decision 1 — Extend `internal context` with `--base <pack_id>`, no new command

**Decision**: `internal context` gains one new optional flag, `--base
<pack_id>`, usable only with `--mode package`. When absent, behavior is
byte-identical to today. When present, the response's `items` becomes
a **diff** against the named base instead of the full item list —
unless the base cannot be confirmed (Decision 4), in which case the
full package is still returned, marked as a full recovery. Every
`--mode package` response, whether a diff or a full recovery, also
gains a `pack_id` field the caller can save and pass as a future
`--base`.

**Rationale**: `internal context --mode package` (033) already
computes and returns exactly the content this feature needs to diff —
`contextengine.PackageItem` already carries `Path`, `Location`,
`Content`, and `Fingerprint` per item. Reusing that pipeline avoids a
second selection/ranking/budgeting code path (Constitution Principle
VI, DRY) and keeps one place — `contextSchemaVersion` — as the single
version a caller checks for "did the response shape change"
(Constitution Principle IX).

**Alternatives considered**: A separate `internal context-diff`
command — rejected; it would have to duplicate `Collect`/`Rank`/
`ApplyBudget`/`BuildPackageItems` or awkwardly wrap `internal context`
as a subprocess, for no benefit over one new flag on the existing
command (Principle IV). Returning the diff as a *third* mode value
(`--mode diff`) instead of a flag on `package` — rejected; a diff is
not a new content *shape*, it is the same package shape with reuse
applied, and `--base` naturally implies "act like package mode, but
incrementally" without introducing a fourth thing to keep in sync with
`--mode`.

## Decision 2 — Pack identity is a SHA-256 hash of the request's own resolved configuration plus every selected item's own identity+fingerprint

**Decision**: `pack_id` is computed as `sha256` over a deterministic,
canonical serialization of: the resolved `Target`, `Intent`, `Task`,
`Query`, `QueryMode`, `Budget`, `HardLimit`, `PreferSection` (the
`contextengine.Request`'s own effective fields — spec FR-001's
"alvo/tarefa"), `rankingVersion`, `contextSchemaVersion`, the
estimator's own `Name()` (spec FR-001's "política de ranking... e
estimador"), and the ordered list of every selected item's own
`(Path, Location, Fingerprint)` (spec FR-001's "conteúdo de origem
selecionado"). Two calls with identical resolved inputs always produce
the identical `pack_id`; any difference in any one of these — even an
unrelated later Requirement's own content, if it happened to be
selected — changes it.

**Rationale**: This is exactly spec FR-001's own definition, and reuses
`contextengine.Fingerprint`'s existing SHA-256 convention (033) rather
than inventing a second hashing scheme. Including every selected
item's own fingerprint (not just a whole-file hash) makes `pack_id`
automatically change the instant *any* selected content differs —
the property Decision 4/FR-006 needs for invalidation, computed as a
side effect of identity rather than a second, separate check.

**Alternatives considered**: Hashing only the request configuration
(target/intent/query/etc.), not the selected content — rejected; two
calls with the same configuration but different underlying file
content (the normal "content changed since last time" case) would
collide on the same `pack_id`, defeating FR-006's invalidation
requirement. A random/sequential ID assigned at storage time —
rejected; it cannot be independently recomputed by the caller to
verify "is this really the same pack," and it does not naturally
encode config-equality the way a content-derived hash does.

## Decision 3 — Storage is one new table in the existing disposable context index, not a new store; identity splits into `config_hash` + `pack_id`

**Decision**: `PackIdentity` (Decision 2) splits into two independently
hashed parts: `ConfigIdentity` — every field *except* the selected
items (`Target`, `Intent`, `Task`, `Query`, `QueryMode`, resolved
`Budget`/`HardLimit`, `PreferSection`, `RankingVersion`,
`ContextSchemaVersion`, `Estimator`) — and the item list. `config_hash`
is the SHA-256 of `ConfigIdentity` alone; `pack_id` is the SHA-256 of
`config_hash` plus the ordered item identity+fingerprint list (still
satisfying Decision 2's "any single differing component changes it").
A new `packs` table is added to the same SQLite database `internal
context` already opens at `.misterspec/cache/context.db`
(`internal/context/index`), storing `pack_id` (primary key),
`config_hash`, `target`, `created_at`, and a JSON-serialized ordered
list of each item's own `(Path, Location, Fingerprint, Content)`.
`index.schemaVersion` is bumped; the package's existing version-
mismatch-triggers-full-rebuild path (`ensureSchema`) already drops and
recreates every table on a mismatch — `packs` included, with no
separate migration path of its own. A simple bound (e.g., capped row
count, oldest-first eviction) keeps the table from growing unboundedly
across a long-lived project.

**Correction found during design review**: an earlier draft of this
decision stored only `pack_id` itself (the full, items-inclusive
hash) and proposed validating a `--base` request by "recomputing the
current request's own identity and comparing it" — but a SHA-256 hash
is one-way; there is nothing to recompute *against* without having
stored the config components separately. `config_hash` is the
smallest additional piece of stored state that makes Decision 4's
validation mechanically possible: the *current* call's own freshly
computed `config_hash` (from its own live request, not derived from
the stored hash) is compared against the *stored* row's `config_hash`
column directly, and only a match proceeds to per-item diffing.

**Rationale**: `internal/context/index` is already explicitly disposable
(Constitution Principle III) and already reconstructable on any schema
change — exactly the property spec FR-011 requires ("todo estado usado
para calcular diferenças... MUST ser tratável como descartável").
Adding one table reuses `index.Open`/`Store.Sync`'s existing lifecycle
(opened once per `internal context` call, already on the hot path) —
no second file, no second open/close/lock discipline (Principle VI).
Losing every stored pack on an unrelated schema bump (e.g., a future
change to `chunks`/`links`) is coarser invalidation than strictly
necessary, but it is always *safe* per FR-011 — the caller simply gets
full recovery (Decision 4) that one time, never a correctness problem.

**Alternatives considered**: A separate cache file/database dedicated
to packs — rejected; two independent SQLite lifecycles for one CLI
process is unjustified complexity for a table this feature could add
to the one already open (Principle IV). Storing packs as plain files
under `.misterspec/cache/` (one file per `pack_id`) — rejected; loses
the existing `ensureSchema`/version-mismatch-triggers-rebuild
mechanism for free, and would need its own pruning logic reimplemented
rather than a `DELETE` statement.

## Decision 4 — A named base is only ever trusted after being re-fetched and its own hash re-verified; anything else is full recovery

**Decision**: `--base <pack_id>` is treated purely as a *lookup key*,
never as an assertion. The command looks up `pack_id` in the `packs`
table; if absent, it is unknown (spec FR-004's first case). If present,
the current call's own freshly computed `config_hash` (Decision 3 —
target, intent, query, resolved budget/hard-limit, prefer-section,
ranking version, contract version, estimator; everything except the
selected items) is compared against the stored row's own `config_hash`
column; any mismatch means the stored row no longer applies to this
request shape and is discarded, falling back to full recovery (spec
FR-004's second case, spec FR-006 — this is the check that actually
catches "the caller passed a different `--budget`/`--intent` this
time," which a bare `pack_id` string alone cannot answer). Only on a
`config_hash` match does the command proceed to compute the current
full item selection and diff it, by identity+fingerprint, against the
row's own stored `items_json` (Decision 5) — a per-item content change
is exactly what that diff step surfaces; it is never a reason to
discard the row itself.

**Rationale**: This is spec FR-004's own core safety rule — "a alegação
de posse de quem chama MUST NOT ser aceita sem essa verificação" — made
structural: there is no code path where a `--base` value alone, without
a matching stored row, produces anything but a full response. Every
full-recovery response also sets a `recovered: true` marker (spec
FR-005) so the caller never has to infer this from response size.

**Alternatives considered**: Trusting a `--base` value that matches the
*shape* of a `pack_id` (e.g., a valid-looking hash) without a stored
row to back it — rejected outright; this is exactly the "confiar
cegamente" failure mode spec User Story 2 exists to prevent. A
soft/partial diff when only *some* items match — considered and kept:
Decision 5 already handles this per-item, so a "mismatch" here means
only "the whole base row is unusable as a unit" (wrong target/config),
not "some items within a still-usable base changed" (that is exactly
what a diff computes, not a reason to discard the base).

## Decision 5 — Diff is an identity-keyed comparison (Path+Location) with an explicit final ordering, not a text-level patch

**Decision**: Each stored/current item's identity is `(Path,
StartLine, EndLine)` — the same identity `contextengine.Candidate`
already uses for deduplication (`result.go`, research.md #8 of
015-context-collector). The diff response lists, in the *current*
selection's own final order: for each position, either `"reuse":
<base_index>` (identity and `Fingerprint` unchanged from the base — no
`content` sent) or the item's own full new data (`"added"` if its
identity is new, `"modified"` if its identity existed in the base with
a different `Fingerprint`). Every base identity absent from the
current selection is additionally listed under `"removed"`. Applying
this response against the named base (substituting each `"reuse"`
entry with that base index's own stored item, keeping every explicit
entry's own new data, in the response's own listed order) reproduces
the exact full package a direct request would return (spec FR-003).

**Rationale**: This is the simplest structure that satisfies FR-003
exactly and mechanically — an ordered list of "reuse-or-replace"
instructions is trivially reversible by the caller, unlike a
line-level text patch (which the project has no existing precedent or
need for) or an unordered added/modified/removed triple alone (which
cannot alone reconstruct final ordering per spec FR-007/FR-008 without
extra position bookkeeping). Reusing `Candidate`'s own established
identity convention avoids inventing a second identity scheme for the
same content (Principle VI).

**Alternatives considered**: An unordered three-bucket diff
(added/modified/removed) with no explicit final ordering — rejected;
spec FR-007/FR-008 explicitly require preserving stable ordering for
unaffected parts, which an unordered bucket structure cannot express
without the caller re-deriving order from elsewhere. A byte-level text
diff of the whole rendered pack — rejected; loses per-item
identity/fingerprint tracking that `internal context --mode package`
already carries, and would require diffing serialized JSON text rather
than the already-structured item list the system already has in hand.

## Decision 6 — Reuse measurement is a per-call diagnostic field, not a new command

**Decision**: Every `--base`-driven response (diff or full recovery)
includes a `reuse` diagnostics block: `items_reused` (count of
`"reuse"` entries), `items_sent` (count of `"added"`+`"modified"`
entries, 0 for a pure no-op reuse), and `bytes_saved_estimate` (the
estimator's own token estimate for the reused items' content, had it
been resent). No separate aggregation command is introduced; per-call
reporting is sufficient for spec FR-009's "medir... por integração" —
an integration's own tooling aggregates these numbers across calls if
it wants a rollup, the same way `eval-compare`'s own baseline
comparison already builds on individually-recorded run data (037)
rather than the framework computing rollups itself.

**Rationale**: Matches Constitution Principle IV — no new aggregation
machinery is justified by this feature alone; per-call numbers are the
minimal, sufficient signal, and an integration genuinely wanting
long-term tracking already has 037's own evaluation-harness precedent
to build on if it becomes a real need.

**Alternatives considered**: A dedicated `internal reuse-stats`
command aggregating history from the `packs` table — rejected as
premature scope (Principle IV); nothing in the spec's Success Criteria
requires a standing report, only that per-integration measurement be
*possible* (SC-005).

## Decision 7 — Provider prompt-cache independence is documentation, not a runtime check

**Decision**: Spec FR-010's "não confundir com cache de prompt do
provedor" is satisfied by documentation (command reference, this
plan's own framing) stating plainly that `pack_id`/`--base` reuse is a
local, project-side mechanism entirely independent of whatever
provider-side prompt caching a calling agent's own LLM session may or
may not use — the response never claims a dollar/token cost saving,
only a measured local reuse count (Decision 6). No runtime
integration with any provider's caching API is built.

**Rationale**: The spec's own Assumptions section states this
explicitly as a documentation concern, not a technical integration
point — the two mechanisms operate at different layers (local content
selection vs. a provider's own prompt-token cache) and there is
nothing for misterspec's binary to detect or coordinate with (spec's
own Limite for PROP-12 in the backlog document).

**Alternatives considered**: none — this is a documentation-only
decision with no competing technical approach to weigh.
