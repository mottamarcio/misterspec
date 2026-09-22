# Data Model: Análise de Impacto e Revisão Incremental

All types below live in a new package, `internal/impact`, at the same
layer as `internal/prepare` (imports `internal/operations`,
`internal/validation`, `internal/evidence`, `internal/vcs`,
`internal/ids`, `internal/artifacts`; no import cycle, per
research.md Decision 4/5). Nothing here is persisted (Constitution
Principle III) — every value is recomputed on each `analyze-impact`
call from two Git revisions plus the project's current filesystem
state.

## ChangedElement

One artifact- or Requirement-scoped element found different between
`--from` and `--to` (spec Key Entity "Conjunto de Mudança", FR-001).

| Field | Type | Notes |
|---|---|---|
| `ID` | `ids.EntityID` | The owning artifact's ID. For a Requirement-level entry, still the owning Spec's ID (Requirement identity is `{ID, RequirementNumber}`, not a standalone `ids.EntityID` — Requirements are not wikilink targets). |
| `Path` | `string` | The artifact's canonical file path (or owning Spec's `plan.md`/`tasks.md` path — research.md Decision 8). |
| `Status` | `ChangeStatus` | `added`, `modified`, or `removed`. |
| `RequirementNumber` | `*int` | Non-nil only for a Requirement-level entry inside a modified/added/removed Spec (research.md Decision 3); nil for a whole-artifact entry. |

`ChangeStatus` is a string enum: `"added" | "modified" | "removed"`.

A single `git diff` covering a modified `spec.md` typically yields one
whole-artifact `ChangedElement` (`Status: "modified"`, `RequirementNumber: nil`)
*and* one `ChangedElement` per individually changed Requirement number
inside it — both are reported; the whole-artifact entry still drives
formal/semantic backlink walking (Decision 4), the per-Requirement
entries drive coverage walking (Decision 5).

## ChangeSet

The full result of comparing `--from`/`--to` (spec Key Entity
"Conjunto de Mudança").

| Field | Type | Notes |
|---|---|---|
| `From` | `string` | The resolved `--from` revision (always a concrete SHA, resolved via the same Git call that produced the diff). |
| `To` | `string` | The resolved `--to` revision, or `"working-tree"` when `--to` was omitted. |
| `Elements` | `[]ChangedElement` | Every recognized artifact/Requirement-level change, deterministically ordered (path, then Requirement number ascending, then whole-artifact-before-Requirement-entries). |
| `UnmappedCodePaths` | `int` | Count of changed paths in the diff that are not one of the five referenceable types nor a `plan.md`/`tasks.md` (research.md Decision 9) — 0 when every changed path was recognized. |

## PropagationHop

One edge in a `PropagationPath` — one relation, walked in the reverse
(backlink) direction (spec Key Entity "Caminho de Propagação").

| Field | Type | Notes |
|---|---|---|
| `Relation` | `RelationKind` | `"depends_on" \| "parent" \| "supersedes" \| "coverage" \| "evidence" \| "wikilink"`. |
| `FromID` | `ids.EntityID` | The element this hop originates from (the changed element, or the result of a prior hop). |
| `ToID` | `ids.EntityID` | The element/Task this hop reaches. For a `coverage`/`evidence` hop, a Task's composite ID (`SPEC-###:TASK-###`, 031-canonical-task-identity). |
| `SourcePath` | `string` | The referencing artifact's own path (from `operations.ReferenceEntry`/`BacklinkEntry` for formal/wikilink hops; the owning Spec's `tasks.md` path for coverage/evidence hops). |
| `SourceSection` / `SourceLine` | `string` / `int` | Populated only for a `wikilink` hop (038 provenance), empty/zero otherwise — same population rule `BacklinkEntry` already establishes. |

## PropagationPath

The full ordered sequence of `PropagationHop`s from one `ChangedElement`
to one affected item (spec Key Entity "Caminho de Propagação", FR-005).
Always has length ≥ 1. `Hops[0].FromID` equals the originating
`ChangedElement.ID`; `Hops[len-1].ToID` equals the affected item's own
ID.

## Classification

`"deterministic_invalidation" | "suggested_review"` — derived
mechanically from a `PropagationPath`'s *last* hop's `Relation` via the
fixed table in research.md Decision 7: `depends_on`/`parent`/
`supersedes`/`coverage`/`evidence` → `deterministic_invalidation`;
`wikilink` → `suggested_review`, unconditionally (spec FR-003/FR-004).

## Severity

`"high" | "medium" | "low"` — derived mechanically from
`(Classification, last hop's Relation)` via the fixed table in
research.md Decision 7. Deterministic for identical input (spec
FR-010): recomputing the same `PropagationPath` always yields the same
`Severity`.

## AffectedItem

One artifact or Task found reachable from the Change Set, with its
full explanation (spec Key Entity "Classificação de Impacto" +
"Candidato a Reexecução", combined into one reported row per FR-005).

| Field | Type | Notes |
|---|---|---|
| `ID` | `ids.EntityID` | The affected artifact's or Task's own ID. |
| `Path` | `string` | Its own canonical path (`tasks.md#TASK-NNN` form for a Task, matching `validation.Finding.Path`'s existing `"path#TASK-NNN"` convention). |
| `Classification` | `Classification` | See above. |
| `Severity` | `Severity` | See above. |
| `Reason` | `string` | A specific, human-readable explanation (never generic) — e.g. `"TASK-003 serves SPEC-014:R2, whose content changed between <from> and <to>"`, or, when the Task's own 041 EvidenceState is already non-Verified, that state is appended (research.md Decision 10), e.g. `"...; its own recorded evidence is already stale against its current content"`. |
| `Path_Propagation` (`Paths`) | `[]PropagationPath` | Every distinct path from any `ChangedElement` that reaches this item — a single item can be reached more than once (e.g. by both a formal dependency and a coverage hop) and every such path is kept, never collapsed into one (mirrors 038 FR-004's "each occurrence remains separately identifiable"). |
| `ReverificationCandidate` | `*string` | Non-nil only for a `deterministic_invalidation` item that is a Task: the Task's own composite ID, meant to be re-run through `internal capture-evidence` (041) — a pointer, not a command line, since deciding *how* to re-verify stays the agent's job (Constitution Principle I). |

## ImpactReport

The full response of one `analyze-impact` call (spec Key Entity set,
combined into the operation's own top-level result).

| Field | Type | Notes |
|---|---|---|
| `ChangeSet` | `ChangeSet` | As above. |
| `AffectedItems` | `[]AffectedItem` | Deterministically ordered: `Severity` (high → low), then `Path` ascending — never a nondeterministic map-iteration order. |
| `NoKnownRelationElements` | `[]ids.EntityID` | Every `ChangedElement.ID` for which the traversal (Decision 4-6) found zero backlinks/coverage/evidence edges at all — reported explicitly, distinct from `AffectedItems` being empty for an unrelated reason (spec FR-008; Edge Case "nenhuma relação reversa conhecida"). |

`AffectedItems` being empty and `NoKnownRelationElements` being empty
are two different, both-legitimate report shapes: the former with a
non-empty `ChangeSet.Elements` and every element appearing in
`NoKnownRelationElements` means "nothing known to be affected, and
that is stated, not merely absent from the report" (FR-008); an
`ImpactReport` MUST NOT be interpreted as "impact verified as absent"
from the mere absence of `AffectedItems` alone — only the explicit
`NoKnownRelationElements` membership carries that meaning.

## Reused, unmodified types (no new definitions)

- `evidence.EvidenceState` (041) — read, never recomputed by this
  feature (research.md Decision 10).
- `operations.ReferenceEntry` / `BacklinkEntry` (012, 038, 040) — the
  source of every formal/`wikilink` `PropagationHop`.
- `validation.RequirementRef` / `TaskCoverage` (032) — the source of
  every `coverage` `PropagationHop`, via the reused
  `validation.ParseTaskCoverage`.
- `ids.EntityID` (031) — every element/Task identity in this feature.

## New, additive exports required in existing packages

- `internal/vcs`: `DiffNameStatus`, `FileAtRevision`, `DiffEntry`
  (research.md Decision 2).
- `internal/validation`: `RequirementSections(spec ids.EntityID, body []byte) map[int]artifacts.Section`,
  an additive sibling to the existing unexported `parseSpecRequirements`
  (research.md Decision 3) — `parseSpecRequirements` itself is
  unchanged.

No existing exported signature changes; no existing Finding `Code`
changes meaning (Constitution Principle VII).
