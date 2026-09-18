# Phase 0 Research: Modelo canônico de tarefas e identidade composta

No `[NEEDS CLARIFICATION]` markers were left in the Technical Context — this feature is a scoped fix inside an existing, well-understood codebase, and the spec's Assumptions section already resolved the open questions a fresh unknown-tech feature would need research for. This document instead records the design decisions the current code's actual behavior forces, so Phase 1 can proceed from verified facts rather than assumptions.

## Decision 1 — Where the composite identity type lives

**Decision**: Add a `TaskID` type to `internal/ids` (alongside `EntityID`), not a new package.

**Rationale**: `internal/ids` already owns `EntityID`, `Parse`, `ParseAny`, and `Scan` — the exact responsibilities a composite Spec+Task identity needs (parsing, scanning, formatting). Constitution Principle IV forbids new package layers without demonstrated pressure; none exists here. Constitution Principle VI (SRP) is satisfied because `ids` already has "identity representation and parsing" as its one reason to change.

**Alternatives considered**:
- A new `internal/tasks` package — rejected: no other behavior (creation, listing) currently justifies a dedicated Task package; would duplicate `ids.EntityID`'s parsing machinery.
- Embedding the Spec directly inside `EntityID` for all types — rejected: every other entity type (Program, Feature, Spec, Knowledge, Learning) has a genuinely global, non-composite identity; forcing a `SpecID` field onto all of them for Task's sake violates ISP (Principle VI) and would ripple into every existing `EntityID` consumer.

## Decision 2 — How per-Spec scanning is scoped

**Decision**: `ids.Scan(root, cfg, ids.Task)` keeps its current signature and semantics (global scan, all `tasks.md` files) for backward-compatible callers, but its internal claims map is rekeyed to `(specNumber, taskNumber)` pairs instead of bare `taskNumber`. Two new derived views are exposed:
- a same-spec duplicate check (feeds `validation.ValidateProject`'s per-Spec Task loop), and
- a global-by-number index (feeds `operations.Resolve`'s "no Spec context given" path) that groups paths by bare Task number *across* Specs, distinguishing "one Spec claims it" (resolvable) from "more than one Spec claims it" (requires explicit Spec context, spec FR-003) from "one Spec claims it twice" (a true duplicate, spec FR-004).

**Rationale**: This is the one internal representation change that lets `ValidateProject` stop treating "same number, different Specs" as a duplicate (User Story 2) while still letting `Resolve` serve a bare `TASK-NNN` when it happens to be globally unique (spec FR-009's "local form stays valid" + FR-003's ambiguity carve-out). Recomputing per-Spec scans from scratch for every Spec (e.g. N separate `filepath.Glob` calls) was rejected as it reintroduces the "who calls which scan" inconsistency PROP-01 exists to remove (spec FR-005: one shared parser/resolver).

**Alternatives considered**:
- Scanning only the caller's named Spec directory directly (`filepath.Glob(specDir/tasks.md)`), bypassing `ids.Scan` for the Task type entirely — rejected: this is exactly what `mister-implement` currently expects, but doing it as a second, separate code path (rather than a view over the same underlying scan) violates FR-005 (single shared resolver) and Principle VI's DRY guidance.

## Decision 3 — CLI/argument shape for the composite reference

**Decision**: Accept `SPEC-###:TASK-###` as a single positional argument wherever a task ID is expected today (`inspect <id>`, and any future `prepare`/task-scoped command), parsed by a new `ids.ParseTaskRef` that recognizes the `:` separator and splits into two `ids.Parse` calls. A bare `TASK-###` remains accepted as before; only when it is ambiguous does resolution fail with an actionable error asking for the composite form.

**Rationale**: `kit/skills/mister-implement/SKILL.md` already documents `/mister-implement SPEC-### TASK-NNN` (two space-separated arguments at the Skill/slash-command layer) — the composite `:` form is additive at the Go CLI/`ids` layer, giving both the Skill (which already has both pieces as separate arguments) and a future single-string consumer (e.g. a Context Pack manifest referencing a task) a way to name the same tuple unambiguously in one token. This does not require changing the Skill's existing two-argument invocation.

**Alternatives considered**:
- Space-separated two-argument-only (`inspect SPEC-014 TASK-003`) — rejected as the sole form: doesn't serve a caller that only has one string in hand (e.g. from a wikilink target, a JSON field, a log line) without the caller inventing its own separator convention, which is exactly the kind of guessing Constitution Principle II forbids Skills from doing.
- A different separator (`SPEC-014/TASK-003`, `SPEC-014.TASK-003`) — rejected: `:` matches the proposal's own illustrative example (`SPEC-014:TASK-003`) already used in the spec and in PROP-01's source document, and doesn't collide with existing path-like (`/`) or extension-like (`.`) syntax already meaningful elsewhere (e.g. `KNOW-003#retry-policy` uses `#` for a different, unrelated future extension — PROP-09 — so `:` avoids that collision too).

## Decision 4 — Migration diagnostic scope

**Decision**: The migration diagnostic (spec FR-008, User Story 3) is a read-only report layered on top of the same rescoped scan (Decision 2): for every Task number that is claimed by more than one Spec project-wide, list the Specs and paths involved, labeled as "previously collided under global scanning, now valid under per-Spec scoping" — distinct in wording and code from an actual same-Spec duplicate Finding.

**Rationale**: Spec FR-008 requires identifying affected cases without mutating anything; this is a pure read over data `ValidateProject` and `Resolve` already compute for Decision 2, so no new scanning logic is needed — only a new report shape distinguishing "was ambiguous under the old global model" from "is a genuine duplicate under the new model."

**Alternatives considered**: A separate one-off migration script outside the `misterspec internal` namespace — rejected: Constitution's "Public command surface" constraint keeps the internal deterministic API under `misterspec internal …`; a bespoke external script would duplicate scanning logic and bypass Principle IX's machine-readable contract requirement.
