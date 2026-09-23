# Phase 0 Research: Validação de cobertura de requisitos e dependências do SDD

No `[NEEDS CLARIFICATION]` markers were left in the Technical Context. This document records the design decisions the current code and documented conventions force, verified against the actual codebase rather than assumed from the backlog document alone.

## Decision 1 — Requirement ID grammar is `R<digits>`, not the `EntityID` dash/padding grammar

**Decision**: A Requirement identifier is `R` followed by one or more digits (`R1`, `R2`, `R12`), with **no dash and no zero-padding** — matching `kit/templates/spec.md.tmpl`'s literal `### R1 — <Requirement>` heading, not `internal/ids.EntityID`'s `PREFIX-NNN` convention (e.g. `SPEC-014`). A new small type/parser is added to `internal/validation`, not to `internal/ids`.

**Rationale**: `internal/ids.Parse`/`EntityID` enforce a fixed prefix + `-` + exactly-`width`-digits grammar, verified against `internal/ids/ids_test.go`'s own table (`"SPEC-14"` — wrong width — is rejected). Requirements use a visibly different, unpadded grammar already fixed by the shipped template — reusing `EntityID`'s parser would either reject every real `R1`/`R2` heading or require weakening `EntityID`'s width enforcement for every other entity type. A Requirement also has no canonical file of its own and is never `Scan`/`Resolve`d as a standalone artifact (no `internal inspect R1`, no `internal create requirement`) — it only ever appears already-scoped inside a known Spec's body, so it does not fit `internal/ids`'s "globally scannable entity" shape the way `ids.TaskID` (031-canonical-task-identity) does. It belongs in `internal/validation`, the package that already owns "is this Spec's declared content structurally correct."

**Alternatives considered**:
- Reusing `ids.EntityID` with a new `Requirement` `EntityType` and a relaxed, unpadded width rule just for it — rejected: `EntityID.String()`'s zero-padding (`fmt.Sprintf("%s-%0*d", ...)`) is baked into every other entity type's identity; special-casing one type inside that shared formatter is exactly the kind of blurred responsibility Constitution Principle VI (SOLID/SRP) warns against.
- A composite `RequirementID` in `internal/ids` mirroring `TaskID` exactly — rejected for the same "not a scannable entity" reason above; `TaskID` exists because Tasks *are* scanned project-wide (`ids.ScanTasks`) for identity/duplicate purposes. A Requirement's own duplicate check (spec FR-002) only ever happens within one already-open Spec body, never across the project.

## Decision 2 — Reuse `artifacts.ParseDocument`/`Section` for both Requirement headings and `Serves:` line parsing

**Decision**: Parse a Spec's `### R<N>` headings and a Task's `Serves:` line(s) by calling `artifacts.ReadBody` + `artifacts.ParseDocument` (013-document-model-chunking's existing heading-chunker) and scanning the resulting `Section.Heading`/`Section.Body` — not a new bespoke `bufio.Scanner` state machine like `operations/inspect.go`'s `taskCheckboxStatus`.

**Rationale**: `ParseDocument` already solves "split a Markdown body into heading-bounded sections, fence-aware" correctly and is already exercised by other features. A Task's own `## TASK-NNN — Title` heading and everything until the next heading is already exactly one `Section` — its `Body` is precisely where a `Serves:` line lives. Writing a second custom scanner (as `031-canonical-task-identity`'s Task-heading duplicate detection does, at the heading-only level with no body content needed) would duplicate fence-tracking and heading-matching logic `ParseDocument` already provides, violating Constitution Principle VI's DRY guidance now that body content, not just headings, needs to be read.

**Alternatives considered**: Extending `internal/ids/scan.go`'s existing regex-based Task-heading scanner to also capture body text — rejected: that scanner intentionally stays heading-only (it feeds `ids.Scan`'s project-wide identity/duplicate model, per 031-canonical-task-identity); conflating it with body-content parsing for a completely different purpose (Requirement coverage) would blur that file's one existing responsibility.

## Decision 3 — `Serves:` grammar: one label, comma-separated references, tolerant of repetition

**Decision**: A line matching `^Serves:\s*(.+)$` inside a Task's `Section.Body` is split on commas into one or more `SPEC-###:R#` references. A Task may have more than one `Serves:` line; references from every line are unioned. A Task body with no `Serves:` line at all has zero references (spec FR-006's "task without requirement" case).

**Rationale**: `mister-tasks/SKILL.md:129-130` describes "the requirement(s) it serves" (plural) attached to one Task, and `mister-analyze/SKILL.md:149` shows the singular literal form `Serves: SPEC-###:R#` for one auto-appended Task. Supporting both a comma-separated list on one line and multiple repeated `Serves:` lines covers both documented shapes without forcing existing/future Task authors into one specific style the Skills themselves haven't standardized on.

**Alternatives considered**: Requiring exactly one `Serves:` line per Task with no list support — rejected: would reject a Task genuinely covering two requirements, a case `mister-tasks/SKILL.md` explicitly anticipates ("the requirement(s) it serves").

## Decision 4 — Plan.md's "Requirement Coverage" prose section is out of this feature's mechanical scope

**Decision**: This feature's deterministic checks read `spec.md` (Requirement headings) and `tasks.md` (`Serves:` references) only. `plan.md`'s own "Requirement Coverage" section (`kit/templates/plan.md.tmpl`) stays human-authored prose, not mechanically parsed or cross-checked against the Requirement/Task graph by this feature.

**Rationale**: The enforceable, deterministic signal PROP-02 is closing the gap on is Task→Requirement (`mister-tasks/SKILL.md:135-136`'s broken promise) — Plan's own coverage prose is the agent's own semantic judgment about *how* it intends to cover requirements, still owned by Principle I's semantic side. Mechanically parsing Plan prose to double-check it against the same Task-level facts would be a second, redundant mechanical path to the same answer, adding complexity Constitution Principle IV would flag without a demonstrated need — spec.md's FR-013 requirement that `validate SPEC-###` "include its subordinate artifacts" is satisfied by including `tasks.md`'s coverage facts; `plan.md`'s own existing frontmatter/structural checks (already covered by 004-structural-validation, unchanged here) remain the only mechanical check against it.

**Alternatives considered**: Parsing Plan's "Requirement Coverage" section for its own `R#` mentions and cross-checking against Spec's declared requirements — rejected as premature scope growth; nothing in spec.md's Acceptance Scenarios or Functional Requirements tests Plan-body content specifically, only Spec (requirements) and Tasks (`Serves:`).

## Decision 5 — Cycle detection: DFS with a recursion-stack (three-color) walk over Spec `depends_on` edges

**Decision**: Build a directed graph whose nodes are every Spec found by `ids.Scan(root, cfg, ids.Spec)` and whose edges are each Spec's own `Metadata.DependsOn` entries (already parsed `[]ids.EntityID`, filtered to `Type == ids.Spec` — the only kind `depends_on` may name, per `checkDependencyList`'s existing behavior). Detect cycles with a standard white/gray/black DFS: a back-edge to a node currently on the recursion stack (gray) is a cycle, reported as the full stack slice from that node to the current one, inclusive (spec FR-009). A Spec depending on itself (a length-1 edge list containing its own ID) is the same algorithm's trivial single-node cycle case, not special-cased separately (spec FR-009's "including self-reference").

**Rationale**: A DFS/three-color walk is the standard, well-understood approach for directed-cycle detection with a reportable path — the backlog document itself names "ordenação topológica ou componentes fortemente conexos" as candidates; DFS-with-recursion-stack is the simplest of the reasonable options (Constitution Principle IV, KISS) that still produces the exact cycle path spec.md's Acceptance Scenarios require, unlike a pure topological-sort-fails-if-cyclic approach (which detects *that* a cycle exists but not *which* nodes form it without extra bookkeeping).

**Alternatives considered**: Tarjan's strongly-connected-components algorithm — rejected as more machinery than needed for a project-scale graph (tens, not millions, of Specs) and messier to turn directly into one ordered "here is the cycle" path per component; DFS's recursion stack already *is* the path.

## Decision 6 — Phase gate keys off `allowedStatesByType[ids.Spec]`, `"draft"` vs. everything else

**Decision**: `internal/validation/tables.go`'s existing `allowedStatesByType[ids.Spec]` (`"draft", "ready", "in_progress", "validated", "blocked", "superseded", "cancelled"`) is read as-is — no new state is added. The phase gate (spec FR-011/FR-012) is a simple `status != "draft"` check: `"draft"` is exempt from coverage completeness; every other already-valid status blocks on incomplete coverage.

**Rationale**: Verified `tables.go:12-18` directly — this is the one and only state table for Spec status today, and `draft` is unambiguously the "not yet started" state in that fixed vocabulary (docs/architecture-specification.md §25-31, per `tables.go`'s own comment). No new per-state nuance (e.g. "blocked"/"cancelled" specs exempted too) is introduced without a demonstrated need — spec.md's Assumptions explicitly settled this as "ready onward" and Edge Cases explicitly requires a cycle to still be reported regardless of a participating Spec's status, keeping the two checks (coverage phase-gate vs. cycle detection) independent.

**Alternatives considered**: Exempting `"cancelled"`/`"superseded"` Specs from the phase gate too — rejected as unnecessary scope: spec.md's Assumptions already settled on "any state after draft blocks," and a cancelled/superseded Spec choosing to also carry incomplete coverage is not a case any Acceptance Scenario asks to be tolerated.
