# Phase 0 Research: Context Collector and Retrieval

All unknowns below were resolved by grounding
`docs/context-engine-implementation.md` §13-18, §29.6 against the actual
codebase (`internal/operations`, `internal/artifacts`,
`internal/context/index`) rather than against the doc's own
illustrative snippets in isolation, per that document's own §32.1
instruction.

## 1. Package placement: `internal/context` gets its first real files

**Decision**: `Request`/`Intent` in `internal/context/request.go`;
`Candidate`/`Reason`/`Tier`/`CandidateSet` in `internal/context/
result.go`; `Collect` and its helpers in `internal/context/collector.go`.
No `engine.go` yet.

**Rationale**: `docs/context-engine-implementation.md` §10.1/§31's own
illustrative layout already reserves exactly these three filenames
under `internal/context/`, alongside the `index/` subpackage 014
already built. `engine.go` is the eventual top-level orchestrator across
ranking (Phase 7), budgeting (Phase 7), and rendering (Phase 8) — none
of which exist yet; creating it now would be an empty wrapper with
nothing to orchestrate (Constitution Principle IV, YAGNI, the same
restraint 014's research.md #2 already exercised for this exact
package).

**Alternatives considered**: Placing `Collect` inside
`internal/context/index` itself (rejected — collection is a distinct
responsibility from storage/search, and the doc's own layout already
separates them).

## 2. `Request` omits `Budget` for now

**Decision**: `Request` carries `Target` (required), `Task`, `Intent`,
`Query` — no `Budget` field, unlike the source document's own
illustrative struct.

**Rationale**: This feature (Phase 6) never reads or enforces a budget
— that is explicitly Phase 7's job ("Ranking and Budgeting"). An unused
field on a public struct is exactly the kind of premature surface
Constitution Principle IV warns against; 013's own research.md #6 made
the identical call for `Chunk.Tokens` (present in the doc's own
illustrative struct, omitted here for the same reason: no current
consumer). Phase 7 adds `Budget` to `Request` when it actually starts
consuming it.

**Alternatives considered**: Including `Budget` now, unused, "for
forward compatibility" (rejected — the same reasoning 013 already
rejected an eager `Tokens` field for).

## 3. Target scope: reuse 012's five entity types, not a new list

**Decision**: `Collect`'s `Request.Target` must resolve to one of the
five entity types `operations.References`/`Backlinks` already cover
(Program, Feature, Spec, Knowledge, Learning) — validated by calling
`operations.Inspect` and checking the same scope 012 and 014 already
established, reusing `operations.ErrInvalidTarget` for a target outside
it.

**Rationale**: Constitution Principle VI (DRY) and direct continuity —
012 drew this boundary, 014 reused it verbatim for `links` population
(014's own research.md #4); redrawing it a third time for a third
purpose would fragment what "the reference graph" means depending on
which feature is asked. `Collect` cannot compute a structural/semantic
tier (User Story 2) for a target `operations.References` itself
refuses, so accepting a wider target set here would be a promise this
feature cannot keep for those extra types anyway.

**Alternatives considered**: Widening scope to match 014's own broader
*indexing* scope (all 9 `ArtifactType` values, since Plan/Tasks/
Validation/Constitution are indexed too) — rejected: a Plan/Tasks/
Validation/Constitution artifact has no reference-graph computation at
all (research.md #4 in 014), so `Collect` could only ever produce Tier
0/1/4 for it, a materially different (and silently degraded) guarantee
depending on target type — simpler and more honest to keep one
consistent target scope across 012/014/015 than to special-case a
narrower feature set for four extra types with no demonstrated need
yet.

## 4. Intent values: keep the source document's own suggested constants

**Decision**: `Intent` is a string type with exactly the five values
the source document itself suggests: `planning`, `tasks`,
`implementation`, `validation`, `analysis`. An empty `Intent` is valid
(no preference); anything else is rejected (FR-002).

**Rationale**: `docs/context-engine-implementation.md` §13 already
proposes these five as its own illustrative `Intent` constants, noting
only that "exact intents should match actual MisterSpec Skill
terminology" — these five already read as generic, action-oriented
categories (not tied to one specific Skill's exact command name) and
map reasonably onto this project's own canonical Skill pipeline
(`/create-plan`, `/create-tasks`, `/implement`, structural validation,
`/analyze`). Renaming them without a demonstrated mismatch would be
speculative churn (Principle IV) — Phase 7, the first feature to
actually branch behavior on `Intent`, is a better-informed point to
revisit this if a real naming problem surfaces.

**Alternatives considered**: Deriving intent names live from
`kit/skills/*/SKILL.md`'s own frontmatter (rejected — no current
consumer needs this dynamism, and it would couple this feature to the
kit's own file layout for no demonstrated benefit).

## 5. Structural/semantic tiers read the filesystem directly, never the index

**Decision**: Tiers 0-3 (mandatory Constitution/target, formal
relationships, semantic connections) are computed by reading the
project's files directly — `artifacts.ReadBody`/`ParseDocument`/
`Chunks` (011/013) plus `operations.Inspect`/`References`/`Backlinks`
(012) — never through `internal/context/index`. Only Tier 4 (free-text
matches) uses `index.Store.Search`.

**Rationale**: `docs/context-engine-implementation.md` §7.2's own
explicit requirement (already honored by 012 and 014): correctness of
graph-derived information must never depend on the disposable index
existing. `Collect` would silently violate that guarantee the moment
any of its non-text tiers read from the index instead of computing
directly — Tier 4 is the *only* tier whose entire reason for existing
is "text search," a capability with no filesystem-direct equivalent at
all, so it is the sole tier genuinely dependent on `index.Store`.

**Alternatives considered**: Reading Tier 0-3 content from the index's
own `documents`/`chunks` tables as a performance shortcut, since 014
already stores it (rejected — reintroduces exactly the "correctness
depends on a disposable cache" risk 012/014 both went out of their way
to avoid; at this project's own stated scale, a handful of direct
`ReadBody`/`ParseDocument` calls per request is not a demonstrated
problem worth this tradeoff).

## 6. Free-text query derivation: reuse the given Task string verbatim, never fabricate

**Decision**: Tier 4 runs `index.Store.Search` using `Request.Query` if
non-empty; otherwise, if `Request.Task` is non-empty, that string is
used as the search query verbatim; otherwise Tier 4 contributes nothing
(FR-007).

**Rationale**: FR-007 explicitly forbids inventing a question that was
never asked. Reusing `Task`'s own text verbatim is the only derivation
that adds zero invented content — it is simply the same words already
supplied, redirected into a second use, not a new fact this feature
introduces.

**Alternatives considered**: Extracting keywords from `Task` via some
heuristic before searching (rejected — an invented transformation this
feature has no demonstrated need for; verbatim reuse is simpler and
already satisfies every acceptance scenario in spec.md).

## 7. Second-hop expansion: outgoing-only, from first-hop nodes, always attempted

**Decision**: After Tiers 2-3 produce a first-hop set of connected
artifacts, `Collect` computes each first-hop artifact's own *outgoing*
`operations.References` (never that artifact's own backlinks, and never
the original target's own backlinks a second time) as Tier 5 candidates
— always attempted, not conditionally, since Phase 6 has no "meaningful
ranking value" judgment yet (that arrives with Phase 7's own numeric
scoring).

**Rationale**: `docs/context-engine-implementation.md` §14's own Tier 5
description ("second-hop graph relationships only when budget remains
and they have meaningful ranking value") depends on machinery (budget,
ranking) this feature deliberately does not have yet (Assumptions,
research.md #2). The disciplined resolution: collect the full,
deterministic second-hop set now: Phase 7 decides what survives a
budget; Phase 6's own job is only to make it available, correctly
bounded, deduplicated (FR-010, FR-011), and never explored a third hop
out. Restricting to outgoing-only (not also expanding each first-hop
node's own incoming backlinks) keeps the traversal's own fan-out
bounded and directional, matching FR-010's explicit two-hop ceiling
without an unbounded-lookalike incoming×incoming expansion.

**Alternatives considered**: Also expanding first-hop backlinks'
own incoming edges (rejected — quadratic fan-out risk for no
demonstrated value, and not what "second-hop" most naturally means
starting from a directed graph traversal).

## 8. Candidate identity and deduplication key

**Decision**: A `Candidate`'s identity is `(Path, StartLine, EndLine)` —
the exact chunk. Two discoveries mapping to the same triple merge into
one `Candidate`, unioning their `Reasons` (FR-008); the first
`Tier`/`Relation` combination to reach it is not privileged over a
later one — all reasons are kept, and a Tier 2/3 (direct) classification
always wins over a Tier 5 (second-hop) one for the exact same reason
(FR-011), even if second-hop expansion revisits it after the direct
discovery.

**Rationale**: `(Path, StartLine, EndLine)` is exactly the identity
013's own `Chunk` already carries — no new identity concept invented
(Constitution Principle VI). Keeping every reason (rather than only the
first) directly implements §18's own explicit guidance: "preserve
those signals as ranking/explanation metadata rather than repeating the
content."

## 9. Candidate content: every chunk of every included artifact

**Decision**: Whenever an artifact is included for any reason (target,
Constitution, a structural/semantic connection, a second-hop
connection), every one of its own `Chunk`s (013) becomes a `Candidate` —
not a single summarizing entry per artifact, and not a curated subset.

**Rationale**: Phase 6 has no ranking/relevance model yet (research.md
#2) to justify picking *which* of an artifact's own chunks matter more
— that judgment belongs to Phase 7. Including every chunk is the
simplest, most complete, and only truly deterministic choice available
without inventing a selection heuristic this feature has no mandate for
(Constitution Principle IV). A free-text match (Tier 4) is the one
naturally different case: FTS5 itself already returns specific matching
chunks, not whole artifacts, so Tier 4 candidates are exactly whatever
`Search` returns — no further "include every chunk" expansion is
applied there.

## 10. Deterministic ordering

**Decision**: The `CandidateSet`'s own order is: `Tier` ascending
(`TierMandatory` first, `TierSecondHop` last), then by `Path`, then by
`StartLine` — a total, stable order requiring no separate scoring step.

**Rationale**: FR-009's own tier labeling already gives a natural,
meaningful primary sort key; alphabetical path and line-number
tie-breaking are the simplest secondary/tertiary keys available,
matching every prior feature's own "reuse an existing, already-proven
ordering pattern" discipline (e.g. 012's research.md #5) rather than
inventing a new one.

## 11. Package name: `contextengine`, not `context` (discovered during implementation)

**Decision**: The Go files added directly under `internal/context/`
(`request.go`, `result.go`, `collector.go`) declare `package
contextengine` — not the directory-derived default of `package
context`. The import path stays `internal/context` (unchanged from
Decision 1 and from 014's own already-shipped `internal/context/index`,
whose own `package index` is unaffected).

**Rationale**: Go's own standard library `context` package
(`context.Context`, used pervasively for cancellation/timeouts/values)
is common enough that any caller needing both it and this feature's own
package in the same file would be forced into an import alias on every
single such file, forever — a needless, permanent tax this project has
no reason to accept when a distinct package name costs nothing. Package
name and import path directory name are independent in Go (the
identifier a caller uses is whatever `package` line the source
declares, not the directory's own name) — `gopkg.in/yaml.v3` declaring
`package yaml` is a well-known precedent for exactly this kind of
deliberate divergence. Caught before implementation began in earnest
(the very first test file), so no already-shipped code needed any
rename.

**Alternatives considered**: Renaming the directory itself (e.g. to
`internal/contextengine/`) to make package name and directory agree
(rejected — would also force renaming 014's own already-merged
`internal/context/index` import path for zero functional benefit, the
larger and riskier of the two fixes for the same underlying problem).
Accepting the stdlib shadowing (rejected — `docs/context-engine-
implementation.md` §31's own explicit "follow existing repository
naming... when implementation details conflict with this illustrative
layout" already anticipates exactly this kind of justified deviation
from its own illustrative directory tree).

## Summary of Go footprint

- `internal/context/{request.go, result.go, collector.go}` (new), each
  with a matching `_test.go`.
- No change to `internal/artifacts`, `internal/validation`,
  `internal/operations`, `internal/context/index`, or `internal/cli` —
  this feature only reads from them (`ReadBody`, `ParseDocument`,
  `Chunks`, `Inspect`, `References`, `Backlinks`, `Search`); it adds no
  new call site requiring a change to any of them.
- No new package beyond `internal/context` itself gaining its first
  real files (already anticipated by 014's own package-placement
  decision). No new external dependency. No CLI surface — Phase 8's own
  "internal context" command remains the later, actual entry point.
