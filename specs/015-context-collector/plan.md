# Implementation Plan: Context Collector and Retrieval

**Branch**: `015-context-collector` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/015-context-collector/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 6 ("Context
Collector and Retrieval") exactly: `internal/context` gains its first
real files — `Request`/`Intent` (no `Budget` field, deliberately
deferred to Phase 7), `Candidate`/`Reason`/`Tier`/`CandidateSet`, and
`Collect` — a deterministic pipeline producing a deduplicated,
tier-ordered candidate set labeled with why each item was included.
Tiers 0-3 (mandatory Constitution/target, formal relationships,
semantic connections) are computed directly from project files, reusing
011/012/013 exactly, so their correctness never depends on 014's
disposable index existing; only Tier 4 (free-text matches) queries it.
Tier 5 (second-hop) is a single, bounded, always-attempted outgoing
expansion from Tier 2/3's own first-hop artifacts — no ranking, no
budget, no rendering, no CLI command. No new package, no new
dependency.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: Go standard library only — no new dependency.
**Storage**: Filesystem, unchanged, read-only (FR-012). `internal/context/index`'s existing SQLite database is read (never written) for Tier 4 only (research.md #5).
**Testing**: `go test` — new `internal/context` tests covering `docs/context-engine-implementation.md` §29.6's own list (target always wins over unrelated text hits, formal/semantic connections correctly labeled, text matches surfaced, deduplication across overlapping discovery paths, bounded second-hop with no third hop, out-of-scope/unknown target rejection); 011's, 012's, 013's, and 014's own full suites re-run unmodified as the named regression gate, since this feature adds no call site into any of them requiring a change. Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. `internal/context` gains its first real files (`request.go`, `result.go`, `collector.go`) alongside the `index/` subpackage 014 already built; no existing package modified at all.
**Performance Goals**: Bounded by the same per-artifact read cost 012/014 already accept, times the size of the first- and second-hop connected set — at this project's own stated scale (hundreds to low thousands of artifacts), the same order of cost `Backlinks`/`Sync` already pay today.
**Constraints**: `Collect` is strictly read-only (FR-012) — no mutation of any project artifact, the reference graph, or the search index. `Request.Target` must be one of the five entity types `operations.References` already covers (research.md #3) — no widening to 014's own broader indexing scope. No numeric ranking, no budget enforcement (FR-013) — both explicitly deferred to Phase 7. Second-hop expansion never proceeds a third hop, and is outgoing-only from first-hop nodes (research.md #7).
**Scale/Scope**: `internal/context/{request.go, request_test.go, result.go, result_test.go, collector.go, collector_test.go}` (new). No existing package modified, no CLI change, no new dependency.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Collection is purely mechanical (resolve, read, traverse a fixed number of hops, deduplicate, sort) — no judgment about which content matters more happens anywhere in this feature; that remains Phase 7's own numeric ranking. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** Zero mutation — `Collect` joins `References`/`Backlinks`/`Search` as a fourth read-only capability layered on the same deterministic core. |
| III. Filesystem Is Single Source of Truth | **Pass, actively reinforced.** Tiers 0-3 are computed directly from project files, never from the disposable index (research.md #5) — the same "index is a performance optimization, never a correctness dependency" guarantee 012 and 014 already established, now proven a third time under a genuinely more complex consumer. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised repeatedly.** No `engine.go` ahead of Phase 7/8's own real need (research.md #1); no `Budget` field before Phase 7 actually enforces one (research.md #2); target scope reuses 012's own five types rather than widening to a new list (research.md #3); Intent constants reused verbatim from the source document rather than speculatively renamed (research.md #4); second-hop expansion kept outgoing-only and single-hop rather than a more elaborate traversal with no demonstrated need (research.md #7). |
| V. Test-First Discipline | **Gate carried into tasks.** `docs/context-engine-implementation.md` §29.6's own Retrieval test list is the explicit floor; 011/012/013/014's full suites are named regression gates. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `Collect` composes `ReadBody`/`ParseDocument`/`Chunks`/`Inspect`/`References`/`Backlinks`/`Search` directly rather than re-deriving any of their logic; deduplication reuses `Chunk`'s own `(Path, StartLine, EndLine)` identity rather than inventing a new one (research.md #8). |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes nothing at all — the fourth consecutive read-only capability in this project's own reference/retrieval line (012, 014, now 015). |
| VIII. Safety by Construction | **Pass (N/A).** No write path exists in this feature to make safe or unsafe. |
| IX. Transparent, Machine-Readable Contracts | **Pass (N/A this feature).** No CLI command is added — Phase 8's own "internal context" command is the eventual, later machine-readable surface this capability feeds into. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/015-context-collector/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md           # Phase 1 output (/speckit-plan command)
├── quickstart.md            # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── collector.md          # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md    # /speckit-specify quality checklist
└── tasks.md                 # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # unchanged — no new dependency
├── cmd/misterspec/, internal/cli/         # unchanged — no CLI surface added
├── internal/
│   ├── project/, ids/, testutil/, example/   # example/ extended (+1 quickstart test)
│   ├── artifacts/, validation/, operations/    # unchanged — reused, not modified
│   └── context/
│       ├── index/                                # unchanged (014) — reused for Tier 4 only
│       ├── request.go                              # NEW — Request, Intent
│       ├── request_test.go                          # NEW
│       ├── result.go                                   # NEW — Tier, Reason, Candidate, CandidateSet
│       ├── result_test.go                               # NEW
│       ├── collector.go                                   # NEW — Collect
│       └── collector_test.go                               # NEW
└── kit/                                    # unchanged
```

**Structure Decision**: `internal/context` gains its first real files
— no new package, no existing package modified at all, matching 014's
own "zero blast radius" precedent exactly. The most integrative feature
so far (composing four prior features' own outputs) with the smallest
possible Go footprint for that integration.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
