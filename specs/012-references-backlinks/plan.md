# Implementation Plan: References and Backlinks

**Branch**: `012-references-backlinks` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/012-references-backlinks/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 2 ("References
and Backlinks") exactly: two new read-only `internal/operations`
functions — `References` (what an artifact points to, formal +
semantic) and `Backlinks` (what points at it) — each scoped to the same
five entity types 011's wikilink validation already covers (Program,
Feature, Spec, Knowledge, Learning), each excluding a broken, malformed,
or ambiguous wikilink target exactly as 011's own classification already
distinguishes them. One small, DRY-motivated addition to `internal/ids`
(`ResolveTarget`, sharing the width-tolerant resolution logic
`internal/validation`'s `classifyWikilink` already has, rather than a
third copy of it). Two new `internal/cli/internalcmd` commands
(`references`, `backlinks`) reusing every existing sentinel/JSON/exit-
code convention verbatim — zero change to `errors.go`'s `classify`, the
strongest evidence yet (after 011's own zero-CLI-change precedent) that
this project's generic-adapter-over-deterministic-core contract keeps
paying for itself. No SQLite, no chunking, no ranking — that document's
own later phases.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: Go standard library only — no new dependency.
**Storage**: Filesystem, unchanged. Strictly read-only (FR-010) — no new mutation path anywhere, matching 002-read-operations' and 011's own read-only nature.
**Testing**: `go test` — new unit tests for `ids.ResolveTarget` (mirroring 011's own three-way classification: resolved/broken/ambiguous); new `internal/operations` tests for `References`/`Backlinks` (formal only, semantic only, both, empty, self-reference, duplicate reference not collapsed, broken/malformed/ambiguous exclusion in both directions, unsupported-target rejection, deterministic ordering — `docs/context-engine-implementation.md` §29.3's own test list); new `internal/cli/internalcmd` command tests (JSON shape, exit codes); 011-wikilink-foundation's and 008-cli-cobra's own full suites re-run unmodified as named, explicit regression gates for US3 (SC-003). Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. Three existing packages extended additively (`internal/ids`, `internal/operations`, `internal/cli`/`internalcmd`); no new package, no new external dependency.
**Performance Goals**: `References` is a single-artifact read — trivial. `Backlinks` is a full scan of the five scoped types' artifacts per query (research.md #6, `docs/context-engine-implementation.md` §7.2's own explicit allowance for the MVP) — the same order of cost `ValidateProject` already pays today at this project's own stated scale (hundreds to low thousands of artifacts, §27).
**Constraints**: Scoped to exactly Program/Feature/Spec/Knowledge/Learning (research.md #1) — Task, Plan, Tasks, Validation, Constitution out of scope, rejected via the existing `operations.ErrInvalidTarget`. A wikilink target counts as an edge only when it resolves to exactly one real artifact (research.md #3) — broken, malformed, and ambiguous are all excluded identically, in both References and Backlinks. A formal relationship is reported exactly as declared, with no existence check of its own (that remains 004-structural-validation's job). Deterministic ordering required (FR-007) via existing declaration/document/scan order, not a new sort.
**Scale/Scope**: `internal/ids/scan.go` (or a new small file, + `ResolveTarget`, + test); `internal/validation/wikilinks.go` (refactor `classifyWikilink` to call it — behavior-identical); `internal/operations/{references.go, references_test.go, backlinks.go, backlinks_test.go}` (new); `internal/cli/internalcmd/{references.go, backlinks.go}` (new); `internal/cli/internal.go` (+2 lines wiring); `internal/example` (+1 quickstart test). No new package, no CLI error-code change, no new dependency.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Both operations are purely mechanical (read frontmatter/body, classify against `ids`'s own existing syntax/resolution rules) — deciding *what a relationship means* or *whether to write one* remains an authoring decision this feature never makes. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** Zero mutation — the same read-only boundary 002-read-operations and 011 already established; `References`/`Backlinks` join `Inspect`/`Parent`/`Children` as read-only queries. |
| III. Filesystem Is Single Source of Truth | **Pass, actively exercised.** `Backlinks` performs a full scan on every call rather than persisting a reverse-reference index (research.md #6) — directly implementing §7.2's own explicit "correctness MUST NOT depend on a cache existing" requirement. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** Scope held to exactly 011's own five-type boundary rather than extending to Plan/Tasks/Validation/Task on a hypothetical need (research.md #1); no caching/indexing layer introduced ahead of Phase 4's own dedicated feature (research.md #6). |
| V. Test-First Discipline | **Gate carried into tasks.** `docs/context-engine-implementation.md` §29.3's own test list is the explicit floor; 011's and 008's full suites are named regression gates, not just "the usual suite." |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `ids.ResolveTarget` shares logic that was about to be duplicated a third time (research.md #2) — the clearest DRY exercise since 011's own `splitFrontmatter` widening; `Backlinks` reuses `resolve.go`'s existing `canonicalFilename` directly (same package) rather than a third copy. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes nothing at all. |
| VIII. Safety by Construction | **Pass (N/A).** No write path exists to make safe or unsafe. |
| IX. Transparent, Machine-Readable Contracts | **Pass, reinforced without any `classify` change.** Both new commands reuse `operations.ErrInvalidTarget`/`ErrEntityNotFound`/`ErrEntityAmbiguous` verbatim (research.md #4) — zero lines changed in `errors.go`, direct evidence the existing sentinel-to-code contract already covers a second feature's own new "unsupported/missing/ambiguous target" cases without modification. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/012-references-backlinks/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md           # Phase 1 output (/speckit-plan command)
├── quickstart.md            # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── references-backlinks.md   # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md    # /speckit-specify quality checklist
└── tasks.md                 # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # unchanged — no new dependency
├── cmd/misterspec/                        # unchanged
├── internal/
│   ├── project/, testutil/, example/        # example/ extended (+1 quickstart test)
│   ├── lock/, templates/                      # unchanged
│   ├── installer/, agents/, bootstrap/, tui/   # unchanged
│   ├── artifacts/                               # unchanged — this feature reads via
│   │                                              #   ReadBody/ExtractWikiLinks (011), no new capability needed here
│   ├── ids/
│   │   ├── types.go, ids.go, scan.go               # scan.go MODIFIED: + ResolveTarget
│   │   └── scan_test.go                              # MODIFIED: + ResolveTarget regression coverage
│   ├── validation/
│   │   ├── wikilinks.go               # MODIFIED: classifyWikilink calls ids.ResolveTarget
│   │   ├── wikilinks_test.go            # unchanged — 011's own suite is this refactor's regression gate
│   │   └── ...                            # unchanged otherwise
│   ├── operations/
│   │   ├── resolve.go, inspect.go, status.go, ...    # unchanged — reused, not modified
│   │   ├── references.go                               # NEW — References, ReferenceEntry, ReferencesResult
│   │   ├── references_test.go                           # NEW
│   │   ├── backlinks.go                                   # NEW — Backlinks, BacklinkEntry, BacklinksResult
│   │   └── backlinks_test.go                                # NEW
│   └── cli/
│       ├── internal.go                # MODIFIED: +2 AddCommand lines
│       └── internalcmd/
│           ├── errors.go                # unchanged — zero classify change (research.md #4)
│           ├── references.go              # NEW
│           └── backlinks.go                 # NEW
└── kit/                                    # unchanged
```

**Structure Decision**: Three existing packages extended additively —
no new package. Unlike 011 (which touched no CLI code at all), this
feature does add two thin `internal/cli/internalcmd` wrappers, since
`docs/context-engine-implementation.md` §7 explicitly specifies a CLI
surface (`misterspec internal references`/`backlinks`) for this phase —
but even so, `errors.go`'s sentinel-to-code mapping needs zero changes,
continuing the same "no CLI-layer surprise" trend 011 established.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
