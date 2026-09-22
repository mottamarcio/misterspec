# Research: Análise de Impacto e Revisão Incremental

Input: `specs/042-impact-analysis-review/spec.md`. This feature adds
one new deterministic operation, `internal analyze-impact`, that
consumes capabilities already delivered by prior Specs rather than
inventing its own fingerprinting, relation-resolution, or evidence
model. Each decision below names the existing code it reuses and the
one small extension point each reuse requires.

## Decision 1 — Change detection is a scoped two-revision Git diff, not a persisted snapshot

**Decision**: `internal analyze-impact` takes `--from <revision>` and
an optional `--to <revision>` (default: working tree), and derives the
Change Set (FR-001) from `git diff --name-status <from> [<to>]`,
filtered to paths the project already recognizes as one of the five
referenceable entity types (`operations.referenceableTypes`: Program,
Feature, Spec, Knowledge, Learning) plus each Spec's own `plan.md` and
`tasks.md`. No new persisted "last analyzed state" is introduced.

**Rationale**: Constitution Principle III forbids a persistent
authoritative counter or snapshot store. Two Git revisions are already
the project's own record of "state at a point in time" — the same
source `vcs.CommitsSinceFileAdded` and `capture-evidence`'s own
`Evidence-GitRevision` field already rely on. This makes "what changed"
answerable from ordinary Git history, with no new state to keep in
sync or reconstruct.

**Alternatives considered**: A stored `.misterspec/impact-baseline`
file recording the last-analyzed fingerprint per artifact — rejected,
reintroduces exactly the authoritative-secondary-state Principle III
forbids, and would itself need invalidation logic. Content-hash
comparison against the artifact's *current on-disk* bytes only (no
Git) — rejected, cannot express "what changed since X" at all, only
"is this file different from some string I already have," which pushes
the entire "what is the prior state" problem back onto the caller with
no help from the framework.

## Decision 2 — Two new small `internal/vcs` primitives, no new dependency

**Decision**: Add `vcs.DiffNameStatus(root, from, to string) ([]DiffEntry, error)`
(`to == ""` diffs against the working tree, matching plain `git diff
<from>`) and `vcs.FileAtRevision(root, path, rev string) (content []byte, found bool, err error)`
(via `git show <rev>:<path>`; `rev == ""` reads the working-tree file
directly). Both shell out to the same `git` binary
`CommitsSinceFileAdded`/`HeadCommit`/`IsWorkingTreeDirty` already use.

**Rationale**: `internal/vcs` already owns every Git shell-out in this
codebase (Constitution Principle VI, SRP) — extending it two small,
single-purpose functions is strictly additive, not a rewrite of its
existing contract (Principle VI, OCP). No new third-party Git library
is introduced; `os/exec` + `git` is the same portability story
`041-task-evidence-fingerprint`'s plan.md already accepted.

**Alternatives considered**: A Go Git library (`go-git`) for
programmatic diff/show — rejected, a new third-party dependency for
two operations the shell-out idiom already covers cheaply, with no
stated performance requirement that would justify it (Principle IV).

## Decision 3 — Requirement-level granularity is computed, not persisted, via two `RequirementSections` snapshots

**Decision**: For a changed Spec whose `spec.md` diff-status is
"modified," `internal/impact` reads `spec.md`'s body at both
`--from` and `--to` (via `vcs.FileAtRevision`), and calls a new
exported `validation.RequirementSections(spec ids.EntityID, body []byte) map[int]artifacts.Section`
(a thin, additive export of the same `requirementHeadingPattern`
walk `parseSpecRequirements` already performs, now also returning each
`R<N>`'s own `Section` rather than only its number) against each
snapshot. A Requirement number present in both snapshots with a
different `Section.Body` content fingerprint is a changed Requirement;
present only in the newer snapshot is an added Requirement; present
only in the older is a removed Requirement.

**Rationale**: This is the one genuinely new fingerprint scope this
feature needs — the codebase already fingerprints whole files
(`operations.Fingerprint`) and a Task's own body
(`operations.TaskContentFingerprint`/`evidence.ContentFingerprint`),
but never one Requirement's own text in isolation, which is exactly
the granularity a Task's `Serves: SPEC-###:R#` reference names
(spec User Story 1, FR-002). Reusing `requirementHeadingPattern` and
`artifacts.ParseDocument` (already applied to arbitrary body bytes,
not tied to a live file) avoids a second Requirement-parsing
implementation (Principle VI, DRY).

**Alternatives considered**: Fingerprinting the whole `spec.md` file
and reporting "some Requirement in this Spec may have changed" without
naming which one — rejected, this is exactly the coarse signal the
spec's User Story 3 (explainable propagation path) exists to avoid;
every affected Task would be reported for every Spec edit, including
unrelated prose changes, defeating the feature's own purpose.

## Decision 4 — Reverse relation walking reuses `operations.Backlinks` verbatim; no parallel graph is built

**Decision**: For every whole-artifact `ChangedElement` (a changed/
added/removed Program, Feature, Spec, Knowledge, or Learning),
`internal/impact` calls the already-existing `operations.Backlinks`
once per element and consumes its `Formal` (parent/depends_on/
supersedes) and `Semantic` (wikilink, with `SourcePath`/`SourceSection`/
`SourceLine`/`TargetAnchor` provenance from 038) lists directly.

**Rationale**: `Backlinks` already computes exactly "what points at
this artifact, by which relation, from where" — the reverse-traversal
primitive PROP-11 needs (spec FR-002). Reimplementing that scan inside
`internal/impact` would duplicate `012-references-backlinks`'s and
`038-wikilink-chunk-provenance`'s own logic (Principle VI, DRY).

**Alternatives considered**: A new project-wide in-memory relation
graph built once and queried repeatedly — rejected as premature for
this feature's scope (Principle IV); `Backlinks` is already a full
filesystem scan per call and this feature's own traversal is bounded
by a visited-set (Decision 6), so the number of `Backlinks` calls per
`analyze-impact` run is bounded by the number of distinct artifacts
ever reached, not unbounded.

## Decision 5 — Requirement coverage's reverse edge (Requirement → Task) is a new, narrowly-scoped index, reusing `ParseTaskCoverage`

**Decision**: For each Requirement number found changed (Decision 3)
in Spec S, `internal/impact` reads S's own `tasks.md` (if present),
calls the already-exported `validation.ParseTaskCoverage` per Task
section (the same parse `specCoverageFindings` already performs), and
collects every Task whose `References` contains `{Spec: S, Number: <that R>}`
— `CodeCrossSpecRequirementReference` already guarantees a valid
`Serves:` entry never names a different Spec, so this lookup never
needs to search outside S's own `tasks.md`.

**Rationale**: The forward direction ("does this Task's `Serves:`
cover a real Requirement") is exactly what `032-requirement-coverage-
dependency-validation` built; this feature only needs the same parse
read backwards (Principle VI, DRY) — no new coverage syntax or
storage.

**Alternatives considered**: Extending `operations.Backlinks` itself
to also report coverage edges — rejected; `Backlinks` is scoped to the
five standalone-referenceable entity types and their formal/semantic
relations (`referenceableTypes`), while a Task is scanned per-Spec via
a different mechanism (`ids.ScanTasks`) with its own identity rules
(031-canonical-task-identity) — folding a sixth relation kind into
`Backlinks`'s existing two-list shape would widen a stable contract
for a caller-specific need, better kept local to `internal/impact`
(Principle VI, ISP).

## Decision 6 — Deterministic termination via a visited-element set, not a fixed hop cap

**Decision**: `internal/impact`'s propagation walk keeps a
`map[ids.EntityID]bool` of elements already expanded in the current
`analyze-impact` run; an element is never expanded (its own
`Backlinks`/coverage lookup never re-run) a second time, regardless of
how many distinct paths reach it. There is no fixed maximum hop count.

**Rationale**: Spec FR-006 requires termination in the presence of
cycles, not a bounded *depth* — unlike `038-wikilink-chunk-
provenance`'s deliberate 2-hop retrieval-cost cap (a context-budget
concern that does not apply here), this feature's job is completeness
of the impact picture, so an artifact three formal-dependency hops
away from the change must still be reported. A visited-set alone
already guarantees termination on a finite project graph: each element
is expanded at most once, so the walk cannot loop.

**Alternatives considered**: Reusing 038's fixed 2-hop cap verbatim —
rejected; it would silently under-report impact for any dependency
chain longer than two hops, contradicting spec Edge Case "cadeia de
relações com mais de um salto" (User Story 3, FR-005: full path
required).

## Decision 7 — Classification (deterministic invalidation vs. suggestion) is a fixed table keyed by relation type, not a score

**Decision**: A `PropagationHop`'s `Relation` value maps to exactly one
`Classification` via a small, fixed table:
`depends_on`/`parent`/`supersedes` (formal) → deterministic;
`coverage` (Requirement → Task, Decision 5) → deterministic;
`evidence` (an already-`Stale`/`Failed`/`Unverified` Task per
`evidence.DeriveState`, reused not recomputed) → deterministic;
`wikilink` (038's Semantic backlink) → suggestion, always, with no
override path. `Severity` is a second small fixed table keyed by
`(Classification, Relation)` — e.g. deterministic/evidence = highest
(a Task is concretely unverifiable now), deterministic/formal =
high, suggestion/wikilink = low — computed the same way for the same
input every time (spec FR-010).

**Rationale**: Spec FR-003/FR-004 are explicit and binary: a wikilink
mention MUST NOT become a deterministic invalidation under any
condition. A fixed lookup table makes that guarantee structural
(there is no code path that could promote `wikilink` to
deterministic) rather than a rule that has to be remembered and kept
correct across future changes (Principle IV — no scoring engine or
configurable weighting is introduced; that class of change is exactly
what `036-text-search-ranking`'s own promotion discipline exists to
gate, and it does not apply to this binary classification).

**Alternatives considered**: A numeric confidence score derived from
relation count/recency — rejected; nothing in the spec asks for
ranking affected items by a continuous score, only for a legible,
explainable severity (spec User Story 3), and a numeric score would
need its own evaluation/promotion process (037) this feature has no
need to bootstrap.

## Decision 8 — `plan.md` changes are attributed to their owning Spec; `internal/impact` adds no new addressable entity

**Decision**: `plan.md` is not one of the five standalone-referenceable
entity types (it has no own `ids.EntityID`, no wikilink target). A
changed `plan.md` is recorded in the Change Set as belonging to its
owning Spec's `ids.EntityID` (derived from its path, the same way
`032`'s `specCoverageFindings` already locates a Spec's own
`tasks.md`), and reuses that Spec's own `Backlinks` result for
propagation — no separate "Plan" element type or identity is
introduced.

**Rationale**: Matches spec Assumptions and Principle IV: inventing a
new addressable identity for Plan content is unjustified scope for
this feature; a Plan is already conceptually part of its Spec in the
canonical lifecycle (Constitution "Canonical lifecycle"), and every
formal/semantic backlink a Plan edit could realistically invalidate is
already reachable through the Spec it belongs to.

**Alternatives considered**: Section-level fingerprinting of `plan.md`
mirroring Decision 3's Requirement-level granularity — rejected for
this feature's first version; `plan.md` sections are not referenced by
a structured field the way `Serves:` names a Requirement, so there is
no existing reverse index to walk from a changed Plan section, only
from the Spec as a whole. Revisiting finer Plan granularity is left to
a future Spec if usage shows the Spec-level signal is too coarse
(spec Assumptions defers PROP-13-dependent code-context granularity
the same way).

## Decision 9 — Code-mapped elements: conservative, explicit gap, not silently skipped

**Decision**: `internal/impact` never attempts to resolve a changed
source-code path to a Requirement — no such mapping exists yet
(PROP-13 is unspecified). When the Change Set (from the same `git
diff`) includes paths outside the five recognized artifact types and
their `plan.md`/`tasks.md`, `analyze-impact`'s response includes a
`code_changes_not_mapped: <count>` field (or omits it at 0) alongside
the Elements it *could* classify, rather than silently dropping those
paths from the diff output entirely.

**Rationale**: Spec FR-007 requires an explicit declared limitation,
not silent omission, and Edge Case "vínculo entre código e requisito
está incompleto" names exactly this gap. Surfacing a count (not a
false "0 requirements affected by this code change" claim) keeps the
report honest about what it could not evaluate, satisfying FR-008's
"nenhuma relação conhecida" vs. "impacto verificado como inexistente"
distinction for code paths specifically.

**Alternatives considered**: Treating every non-artifact path change
as affecting every Requirement in the project (maximally conservative)
— rejected; spec FR-007's "estratégia conservadora explícita" asks for
declaring the gap, not manufacturing false-positive impact noise
across the whole project for an unrelated code change (e.g. a
formatting-only diff in an unrelated file).

## Decision 10 — Evidence relation reuses `evidence.DeriveState` as-is; this feature does not change what makes a Task's own evidence stale

**Decision**: When a Task is reached via a `coverage` hop (Decision 5)
because a Requirement it serves changed, `internal/impact` reports
that Task as a deterministic-invalidation item *regardless* of its own
`evidence.EvidenceState` (even if that Task's own body-content
fingerprint still matches, per 041) — the reported reason is "the
Requirement this Task serves changed," a distinct reason from 041's
own "this Task's own content changed" staleness. `internal/impact`
does not modify, re-derive, or wrap `evidence.DeriveState`; a Task
already `Stale`/`Failed`/`Unverified` for its own 041 reasons is
additionally surfaced (its existing state is included in the
`PropagationHop`'s reason text, not recomputed).

**Rationale**: 041 already answers "did this Task's own recorded
evidence go stale against its own content" — this feature answers a
different question, "did something this Task's evidence was never
computed against (the Requirement text itself) change." Conflating the
two into one recomputed state would blur exactly the distinction
spec FR-005's "motivo" field exists to keep legible, and would risk
silently changing 041's own established `EvidenceState` semantics
(Constitution Principle VII — this feature must not casually rewrite
041's contract).

**Alternatives considered**: Extending `EvidenceFields`/
`DeriveState` to also fingerprint every `Serves:`-referenced
Requirement's text at capture time — rejected for this feature's
scope; it would require changing `capture-evidence`'s own written
`Evidence-*:` fields (041's contract, owned by
`/mister-implement` per Constitution Principle VII), a heavier,
cross-feature change better left to a dedicated follow-up if 042's
Requirement-change signal alone proves insufficient in practice.
