# Phase 0 Research: Structural Validation and Project Status

As with the prior features, the frozen envelope
(`docs/architecture-specification.md` §16-19, §32, §70 and the ratified
constitution) leaves no open `NEEDS CLARIFICATION`. This document records
one significant structural decision — discovered while planning, not
assumed up front — plus several smaller ones.

## Decision: `internal/validation` must NOT depend on `internal/operations` — a real import-cycle constraint

- **Decision**: `internal/validation` is a new package depending only on
  `internal/ids`, `internal/artifacts`, and `internal/project` — never on
  `internal/operations`. `internal/operations/status.go` (new file in the
  existing package) is the one direction of dependency:
  `operations` → `validation`.
- **Rationale**: `docs/architecture-specification.md` §32 places `status.go`
  under `operations/` and validation's own files under a separate
  `validation/` package. `Status` needs `validation.ValidateProject`'s
  finding count — so `operations` must be able to import `validation`.
  If `validation` also imported `operations` (e.g. to reuse `Resolve`/
  `Inspect` for entity lookup), that would be an import cycle, which Go
  refuses to compile. One direction has to give, and per §32's own
  placement, `operations` is the side allowed to depend on `validation`,
  not the reverse.
- **A second, independent reason this is the *right* direction, not just
  the *necessary* one**: `operations.Resolve`/`Inspect` are deliberately
  fail-fast — an ambiguous or missing entity becomes a Go error
  (`ErrEntityAmbiguous`, `ErrEntityNotFound`) the caller must handle or
  propagate. Validation's entire purpose is the opposite: every anomaly
  — not found, ambiguous, wrong parent type, invalid state — must become
  a `Finding` in a returned list, never an error that stops collection
  (FR-004). Reusing `Resolve`/`Inspect` as-is would mean catching their
  errors and translating them back into Findings at every call site,
  which is more indirection than just having `validation` walk
  `ids.Scan`'s results directly — `Scan` already exposes exactly the
  "0 matches / 1 match / >1 matches (Duplicates)" shape validation needs,
  without an error-throwing layer in between.
- **Alternatives considered**: Making `validation` depend on `operations`
  and `operations/status.go` NOT depend on `validation` (compute
  `structural_errors` some other way) — rejected because there is no
  other correct way to compute that count than running the same
  validation logic; duplicating it would violate Principle VI (DRY) far
  more than the layering choice here does. Merging `validation` into
  `operations` entirely (no separate package) — rejected as contradicting
  §32's explicit separation and Principle VI's SRP (rule-checking is a
  distinct responsibility from composing primitives into an answer, the
  same reasoning `operations` itself was created for in
  002-read-operations's research).

## Decision: Single-entity validation resolves independently of `operations.Resolve`

- **Decision**: `validation.ValidateEntity(root, cfg, rawID)` parses
  `rawID` itself (`ids.ParseAny` then a strict `ids.Parse` against
  `cfg.IDWidth`, mirroring `operations.Resolve`'s own two-pass approach)
  and looks up `ids.Scan(...).Paths[number]` directly, turning 0 matches
  into a "not found" Finding and >1 matches into a "duplicate/ambiguous"
  Finding — never a Go error for either case. A Go error is reserved for
  a genuinely unusable request: syntactically invalid `rawID`, or a type
  this feature doesn't validate as a single entity (Task, Plan, Tasks,
  Validation, Constitution — mirroring `Create`'s own supported-type
  boundary).
- **Rationale**: See the previous decision — this is the concrete
  consequence of "collect anomalies as data" for the single-entity case
  specifically.
- **Alternatives considered**: Treating "not found" as an error from
  `ValidateEntity` (matching `Resolve`'s behavior) — rejected; a
  not-found entity is exactly the kind of structural problem a validation
  caller wants reported as a Finding, not raised as an exception it must
  catch separately from every other kind of problem.

## Decision: Fixed lifecycle-state tables, hardcoded per type

- **Decision**: `validation` hardcodes the allowed status values per
  entity type, exactly as frozen in `docs/architecture-specification.md`:
  Program/Feature: `draft, active, done, cancelled`; Spec: `draft, ready,
  in_progress, validated, blocked, superseded, cancelled`; Learning:
  `candidate, promoted, dismissed`; Knowledge: `active` only (per
  spec.md's Assumptions — the only state that document ever names for
  Knowledge).
- **Rationale**: These are frozen product facts, not configuration — no
  project customizes its own lifecycle vocabulary, so a config field would
  violate Principle IV (YAGNI: don't add configuration for something that
  never needs to vary).
- **Alternatives considered**: Making allowed states configurable via
  `.misterspec/config.yaml` — rejected as unjustified configuration
  surface for a fixed, frozen vocabulary.

## Decision: Parent-type table lives in `validation`, not shared with `operations/children.go`

- **Decision**: `validation` defines its own small
  `requiredParentType(childType) (ids.EntityType, bool)` table (Feature →
  Program, Spec → Feature; Program/Knowledge/Learning → none), the
  logical inverse of `operations/children.go`'s unexported
  `childTypesFor`.
- **Rationale**: Sharing it would require exporting `childTypesFor` from
  `operations` and importing `operations` from `validation` — exactly the
  cycle the first decision above rules out. A four-line table duplicated
  in spirit (not logic — one maps parent→children, the other child→parent)
  is a smaller cost than restructuring package boundaries for it.
- **Alternatives considered**: Moving both tables into a third, lower
  package (e.g. `internal/ids`) — rejected as premature generalization
  (Principle IV) for two four-line tables used by exactly one caller each
  today; revisit only if a third consumer needs the same mapping.

## Decision: Whole-project validation includes Task in duplicate-ID detection only

- **Decision**: `ValidateProject` calls `ids.Scan` (and reports its
  `Duplicates`) for Task alongside Program/Feature/Spec/Knowledge/
  Learning, satisfying FR-010 for every entity type `Scan` already
  supports. It does **not** run single-entity checks (frontmatter,
  parent, lifecycle state) against Task — Task has none of those
  concepts in this project's model on its own (see
  002-read-operations's Task handling: its "parent" and "status" are
  narrowly derived from its owning Spec's `tasks.md`, not independent
  frontmatter).
- **Rationale**: Task duplicate-ID detection is "free" — it's exactly
  what `ids.Scan(Task)` (built in 001-core-foundation) already returns,
  no new body-parsing capability required. Extending full single-entity
  validation semantics to Task would require the same body-content
  parsing this feature's Assumptions already deferred.
- **Alternatives considered**: Excluding Task entirely from validation —
  rejected; duplicate Task IDs are a real, cheaply detectable structural
  problem (FR-010 doesn't scope itself to only some entity types), and
  omitting a check this feature can already perform for free would be
  under-delivering without a real reason.

## Decision: `operations.Status` composes counts via `Inspect`, findings via `validation.ValidateProject`

- **Decision**: `Status` computes entity-type counts via `ids.Scan(...).IDs`
  per type, Spec-by-state counts via `operations.Inspect` on every found
  Spec, and `structural_errors` via `len(validation.ValidateProject(...))`
  — composing three already-correct primitives rather than a fourth,
  parallel implementation of any of them.
- **Rationale**: Constitution Principle VI (DRY) — `Status` is a
  read-only aggregation over data other primitives already know how to
  produce correctly.
- **Alternatives considered**: Having `Status` re-derive Spec states by
  reading frontmatter directly instead of calling `Inspect` — rejected as
  unjustified duplication of `Inspect`'s already-correct metadata parsing.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
