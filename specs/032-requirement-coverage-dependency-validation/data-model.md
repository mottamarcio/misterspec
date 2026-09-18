# Phase 1 Data Model: Validação de cobertura de requisitos e dependências do SDD

No persisted storage is added (Constitution Principle III). Every type below is an in-memory Go value computed fresh from `spec.md`/`tasks.md` content on each `ValidateProject`/`ValidateEntity` call, never cached or written back.

## RequirementRef

A resolved reference to one Requirement, wherever a `Serves:` line inside a Task's body names one (spec.md "Coverage Reference").

| Field | Type | Description |
|---|---|---|
| `Spec` | `ids.EntityID` (`Type == ids.Spec`) | The Spec the reference names, exactly as written (`SPEC-###`) — not necessarily the Task's own owning Spec; validity of that relationship is checked separately (see "Cross-Spec rejection" below). |
| `Number` | `int` | The Requirement's own number (`3` for `R3`) — no zero-padding, per research.md Decision 1. |

**`String()`**: renders `"<Spec>:R<Number>"` (e.g. `"SPEC-014:R3"`).

**Validation rules** (spec FR-004, FR-005):
- A malformed reference (missing `R`, non-numeric suffix, missing Spec half, wrong Spec-ID syntax) is a parse error — reported as an `unknown_requirement_reference`-shaped Finding naming the raw text, not silently skipped.
- A syntactically valid reference whose `Spec` differs from the Task's own owning Spec is rejected as invalid (spec FR-005) — reported as a distinct Finding code (`cross_spec_requirement_reference`), never treated as valid coverage of the other Spec's requirement.
- A syntactically valid, same-Spec reference whose `Number` does not appear in that Spec's own declared Requirement set is reported as `unknown_requirement_reference`.

## SpecRequirements (per-Spec computed set)

The result of scanning one Spec's own `spec.md` body for `### R<N>` headings (any heading level, matching `kit/templates/spec.md.tmpl`'s `### R1 — <Requirement>` convention).

| Field | Type | Description |
|---|---|---|
| `Spec` | `ids.EntityID` | The owning Spec. |
| `Numbers` | `[]int` | Every distinct Requirement number declared, in document order. |
| `Duplicates` | `[]int` | Any number declared more than once (spec FR-002) — a Requirement heading repeated within the same Spec. |

## TaskCoverage (per-Task computed value)

The result of scanning one Task's own `Section.Body` (from `tasks.md`, via `artifacts.ParseDocument`) for `Serves:` line(s).

| Field | Type | Description |
|---|---|---|
| `Task` | `ids.EntityID` (`Type == ids.Task`) | The Task's own local identity, within its owning Spec's `tasks.md` (031-canonical-task-identity's existing per-Spec Task identity). |
| `References` | `[]RequirementRef` | Every reference found across all `Serves:` line(s) in this Task's body (research.md Decision 3 — comma-separated and/or repeated lines, unioned). Empty when the Task has no `Serves:` line at all (spec FR-006). |
| `MalformedReferences` | `[]string` | Raw text of any `Serves:` entry that failed to parse as `SPEC-###:R#`. |

## RequirementCoverageReport (per-Spec aggregate)

What `ValidateProject`/`ValidateEntity` actually consumes to produce Findings for one Spec — the join of `SpecRequirements` and every `TaskCoverage` belonging to that Spec's own `tasks.md`.

| Field | Type | Description |
|---|---|---|
| `Spec` | `ids.EntityID` | The Spec under evaluation. |
| `UncoveredRequirements` | `[]int` | Numbers in `SpecRequirements.Numbers` referenced by zero same-Spec `Serves:` line project-wide (spec FR-007). |
| `TasksWithoutCoverage` | `[]ids.EntityID` | Tasks (this Spec's own) whose `TaskCoverage.References` is empty (spec FR-006). |
| `InvalidReferences` | list of `{Task ids.EntityID; Ref string; Reason "unknown"\|"cross_spec"\|"malformed"}` | Every `Serves:` entry that failed one of RequirementRef's validation rules above (spec FR-004, FR-005). |

## DependencyGraph (project-wide, ephemeral)

Built once per validation run from every discovered Spec's `Metadata.DependsOn` (spec.md "Dependency Graph").

| Field | Type | Description |
|---|---|---|
| `Nodes` | `[]ids.EntityID` | Every Spec found project-wide (from the existing `ids.Scan(root, cfg, ids.Spec)`). |
| `Edges` | `map[int][]int` | Spec number → the Spec numbers it `depends_on`, filtered to entries of `Type == ids.Spec` (non-Spec `depends_on` entries are already flagged elsewhere by the existing `checkDependencyList`, out of this feature's scope). |

**Derived**: `Cycles []DependencyCycle`, computed via research.md Decision 5's DFS walk.

## DependencyCycle

One detected cycle (spec FR-009), including the degenerate self-reference case.

| Field | Type | Description |
|---|---|---|
| `Path` | `[]ids.EntityID` | The full ordered cycle, starting and ending at the same Spec (e.g. `[SPEC-001, SPEC-002, SPEC-003, SPEC-001]`) — a self-reference renders as `[SPEC-001, SPEC-001]`. |

## SpecPhaseGate (per-Spec computed value)

The result of applying research.md Decision 6's rule to one Spec (spec.md "Spec Phase Gate").

| Field | Type | Description |
|---|---|---|
| `Spec` | `ids.EntityID` | The Spec under evaluation. |
| `Status` | `string` | The Spec's own declared `status` (already parsed via `artifacts.Metadata`). |
| `Exempt` | `bool` | `true` iff `Status == "draft"` (spec FR-011). |
| `Blocked` | `bool` | `true` iff `!Exempt` and (`RequirementCoverageReport.UncoveredRequirements` or `.TasksWithoutCoverage` is non-empty) — the condition that produces a blocking Finding (spec FR-012). |

## State / Lifecycle

No new lifecycle states are introduced; `SpecPhaseGate.Exempt`/`.Blocked` are read-only derived values computed fresh each run from the existing `status` field — no state transition, no persisted flag.
