# Implementation Plan: Orçamento sobre a saída efetiva de contexto

**Branch**: `035-context-budget-accuracy` | **Date**: 2026-09-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/035-context-budget-accuracy/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

`internal/artifacts/tokens.go` already declares an `Estimator` interface and a `DefaultEstimator`, explicitly commented "kept as an interface so a real provider-specific tokenizer can be substituted later" — but every real call site (`internal/context/budget.go:79`, `pack.go:92,100`) bypasses it and calls the free function `artifacts.EstimateTokens` directly; nothing identifies which estimator produced a number. Separately, `016-ranking-budgeting/research.md` Decision #9 documents today's single `Request.Budget` value serving two conflated roles — the target for optional content *and* the ceiling that flags mandatory content as "exceeded" — and its own deliberate "stop entirely at the first optional item that doesn't fit" policy. This plan: (1) actually injects `Estimator` into the pipeline and names it in the response; (2) splits the single budget into a soft limit (optional-content target, existing `Request.Budget`) and a new, separate hard limit (mandatory-content ceiling); (3) replaces "stop at first miss" with per-tier backfill (skip and keep trying smaller same-tier candidates) while never reordering across tiers; (4) adds coherent, code-block/requirement-preserving splitting for an oversized single section. `budget_exceeded`'s meaning changes (now compares against the new hard limit, not the soft one) — a deliberate, spec-mandated redefinition, versioned via 033's own `schema_version` bump to `2`.

## Technical Context

**Language/Version**: Go 1.23.4 (existing `go.mod`)
**Primary Dependencies**: existing internal packages only — `internal/artifacts` (`Estimator`, new coherent-splitting function, reusing the existing unexported `fencedLines`), `internal/context` (`budget.go`, `request.go`, `pack.go`), `internal/cli/internalcmd`. No new external dependency — no real provider tokenizer is implemented in this feature (spec Assumptions).
**Storage**: N/A — every computation is per-request, in memory (Constitution Principle III); nothing new is persisted.
**Testing**: Go `testing` — unit tests for the estimator wiring, hard/soft limit resolution, the backfill algorithm, and coherent-unit splitting (pure functions), plus CLI-level tests for the new flags/diagnostics fields, matching existing `internal/context`/`internal/cli/internalcmd` conventions. A dedicated regression test locks the *unrelated* fields (schema_version aside) to stay unchanged.
**Target Platform**: `misterspec` CLI binary, via `internal context`.
**Project Type**: Single Go module — additive changes to `internal/artifacts` and `internal/context`, plus `internal/cli/internalcmd/context.go`.
**Performance Goals**: No regression; backfill is still a single linear pass over the already-sorted candidate list (no new sorting, no knapsack search).
**Constraints**: MUST NOT change `Estimator`'s existing method set in a way that breaks `DefaultEstimator` (additive `Name() string` method only); MUST NOT silently truncate mandatory content under any hard limit (spec FR-007); MUST NOT let a lower-tier candidate be considered before every higher-tier candidate in `ranked`'s own existing order has been decided (spec FR-009); `budget_exceeded`/`overage`'s redefinition is the one deliberate exception to 033's otherwise-strict "no behavior change to existing fields" discipline, and MUST be called out via `schema_version: 2`.
**Scale/Scope**: Touches `internal/artifacts/tokens.go` (Estimator interface), new `internal/artifacts/split.go` (coherent-unit splitting), `internal/context/request.go` (`HardLimit` field), `internal/context/budget.go` (estimator injection, hard-limit resolution, backfill, exclusions, coherent-split fallback), `internal/context/pack.go` (estimator injection for payload-size estimation), `internal/cli/internalcmd/context.go` (`--hard-limit` flag, new diagnostics fields, `schema_version: 2`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Semantic/Deterministic Separation** — PASS. Every new behavior (estimation, limit resolution, backfill, splitting) is purely mechanical; no judgment about *which* content matters is added — `Collect`/`Rank` remain untouched.
- **II. Deterministic Operations Are the Only Mutation Primitive** — PASS. Entirely read-only response shaping; no mutation, no ID allocation.
- **III. Filesystem Is the Single Source of Truth** — PASS. Nothing new is persisted; every number is recomputed per request.
- **IV. Simplicity First — YAGNI & Minimal Configuration** — PASS. No new package layer: `Estimator` already lives in `internal/artifacts` (this plan finally wires the interface that package already owns); coherent splitting is one small additive function in the same package, reusing its existing fence-tracking rather than a second implementation. `--hard-limit` is a request-scoped flag, not a new persistent config field.
- **V. Test-First Discipline** — Applies. The backfill algorithm, hard/soft limit resolution, estimator naming, and coherent splitting MUST get unit tests before/alongside implementation.
- **VI. Clean Code & SOLID** — PASS. `internal/artifacts` keeps owning "how is Markdown structured / how is text sized"; `internal/context` keeps owning "what does a Context Pack response look like" — no blur.
- **VII. Explicit Mutation Boundaries** — N/A. Read-only.
- **VIII. Safety by Construction** — N/A. No new filesystem write path.
- **IX. Transparent, Machine-Readable Contracts** — Applies directly: this is the feature that makes the estimator identifiable and the hard-limit-exceeded condition an explicit, structured diagnostic rather than a conflated single number — and the one deliberate semantic change to an existing field (`budget_exceeded`) is exactly what `schema_version` exists to signal.

No violations requiring justification; Complexity Tracking is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/035-context-budget-accuracy/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── artifacts/
│   ├── tokens.go          # Estimator gains Name() string; DefaultEstimator
│   │                        # implements it ("default")
│   ├── split.go            # NEW: SplitIntoCoherentUnits(content string) []string,
│   │                        # reusing the existing unexported fencedLines
│   └── split_test.go       # NEW
├── context/
│   ├── request.go          # + HardLimit *int
│   ├── budget.go           # ApplyBudget gains an Estimator parameter; hard-limit
│   │                        # resolution; backfill replaces stop-at-first-miss;
│   │                        # Exclusions; coherent-split fallback for an
│   │                        # oversized single candidate
│   ├── budget_test.go       # extended
│   └── pack.go              # PayloadTokens takes the same injected Estimator
└── cli/internalcmd/
    ├── context.go            # --hard-limit flag; estimator/exclusions in
    │                          # diagnostics; schema_version: 2
    └── context_test.go       # extended
```

**Structure Decision**: No new package. `Estimator` wiring and coherent splitting stay inside `internal/artifacts` (the existing owner of both text-sizing and Markdown-structure concerns); the budget/limit/backfill logic stays inside `internal/context` (the existing owner of Context Pack shaping) — matching Constitution Principle IV/VI's existing boundaries.

## Complexity Tracking

*No Constitution Check violations — this section is intentionally empty.*
