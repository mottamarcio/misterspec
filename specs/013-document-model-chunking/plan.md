# Implementation Plan: Document Model and Chunking

**Branch**: `013-document-model-chunking` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/013-document-model-chunking/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 3 ("Document
Model and Chunking") exactly: a pure, deterministic `ParseDocument`
splitting any artifact body into an ordered, flat list of heading-
bounded `Section`s (never duplicating a parent heading's content into
its children's, research.md #2); `Chunks`, deriving one retrieval-sized,
fully-traceable `Chunk` per non-empty `Section`; and a small, replaceable
`Estimator` abstraction (`EstimateTokens`) for a deterministic,
documented token-count approximation. One small DRY refactor
(`fencedLines`, shared between this feature's heading detection and
011's own wikilink extraction) is the only change to already-shipped
code. No new package (`document.go`/`chunk.go`/`tokens.go` all land in
`internal/artifacts`, matching the source document's own illustrative
layout), no new dependency, no CLI surface — Phase 8's own "internal
context" command is the later, actual CLI entry point.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: Go standard library only — no new dependency.
**Storage**: Filesystem, unchanged. Strictly read-only/pure (FR-010) — `ParseDocument`/`Chunks`/`EstimateTokens` perform no I/O of their own at all (they operate on an already-read `[]byte`/`string`), an even stricter read-only boundary than 011's own `ReadBody`.
**Testing**: `go test` — new `internal/artifacts` tests for `ParseDocument` (headings, nested headings, empty sections, no-heading body, empty body, fenced-block exclusion — `docs/context-engine-implementation.md` §29.4's own list), `Chunks` (one chunk per non-empty section, no chunk for an empty one, provenance correctness, determinism, wikilink-content preservation), and `EstimateTokens`/`Estimator` (determinism, monotonicity, empty-string zero); 011-wikilink-foundation's full existing suite re-run unmodified as the named regression gate for the `fencedLines` extraction. Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. One existing package extended additively (`internal/artifacts`); no new package, no new external dependency, no CLI-layer change.
**Performance Goals**: Trivial — a per-artifact line scan and a handful of string operations, the same order of cost 011's own `ExtractWikiLinks` already proved cheap at this project's stated scale.
**Constraints**: `ParseDocument`/`Chunks`/`EstimateTokens` are pure functions of their inputs — no filesystem access, no path resolution, no entity-ID lookup (research.md #5) — deferring "which artifact types are eligible" to whichever future caller (Phase 4's indexer) actually walks the project. `Chunk` carries no `Tokens` field (research.md #6) — token estimation stays a fully independent, on-demand capability. Section boundaries are ATX headings only (research.md #4); Setext headings and other CommonMark constructs this project's own content never uses are out of scope, matching 011's own Markdown-scope precedent.
**Scale/Scope**: `internal/artifacts/{document.go, document_test.go, chunk.go, chunk_test.go, tokens.go, tokens_test.go}` (new); `internal/artifacts/wikilink.go` (modified: extract `fencedLines`, behavior-identical); `internal/artifacts/wikilink_test.go` (unchanged — 011's own suite is the regression gate). No new package, no CLI change, no new dependency.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Structuring and chunking are purely mechanical (heading-boundary detection, line-range bookkeeping) — deciding what content *means* or how it should eventually be ranked remains entirely out of this feature's scope (later phases). |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A).** Zero mutation — `ParseDocument`/`Chunks`/`EstimateTokens` don't even perform I/O, an even stricter boundary than 011's/012's own read-only operations. |
| III. Filesystem Is Single Source of Truth | **Pass.** No derived state is cached or persisted anywhere — every call re-derives Sections/Chunks fresh from the `body` given to it. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** No new package ahead of Phase 4's own real need (research.md #1); no `ArtifactID`/entity-ID coupling this feature doesn't need (research.md #5); no eager `Tokens` field on `Chunk` before a real consumer (Phase 4's own SQLite row) exists (research.md #6); no large-section subdivision logic without a demonstrated real case (spec.md's own Assumptions). |
| V. Test-First Discipline | **Gate carried into tasks.** `docs/context-engine-implementation.md` §29.4's own Chunker test list is the explicit floor; 011's full existing suite is a named regression gate, not just "the usual suite." |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `fencedLines` factored out of `ExtractWikiLinks`'s existing inline loop rather than duplicated a third time (research.md #3) — direct continuation of 012's own `ids.ResolveTarget` extraction, the same DRY discipline applied twice now to two different kinds of duplication. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes nothing at all, and doesn't even read a file itself — the strictest mutation boundary of any feature so far. |
| VIII. Safety by Construction | **Pass (N/A).** No write or read path exists in this feature to make safe or unsafe. |
| IX. Transparent, Machine-Readable Contracts | **Pass (N/A this feature).** No CLI command is added — Phase 8's own "internal context" command is the eventual, later machine-readable surface this capability feeds into. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/013-document-model-chunking/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md           # Phase 1 output (/speckit-plan command)
├── quickstart.md            # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── document-chunking.md  # Phase 1 output (/speckit-plan command)
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
│   ├── lock/, templates/, operations/         # unchanged
│   ├── installer/, agents/, bootstrap/, tui/   # unchanged
│   ├── validation/                              # unchanged — no new call site
│   └── artifacts/
│       ├── types.go, metadata.go, paths.go        # unchanged
│       ├── parser.go                                # unchanged
│       ├── markdown.go                                # unchanged — ReadBody already exists (011)
│       ├── wikilink.go                                  # MODIFIED: factor out fencedLines
│       ├── wikilink_test.go                              # unchanged — 011's own regression gate
│       ├── document.go                                     # NEW — Section, Document, ParseDocument
│       ├── document_test.go                                 # NEW
│       ├── chunk.go                                           # NEW — Chunk, Chunks
│       ├── chunk_test.go                                       # NEW
│       ├── tokens.go                                             # NEW — Estimator, DefaultEstimator, EstimateTokens
│       └── tokens_test.go                                         # NEW
└── kit/                                    # unchanged
```

**Structure Decision**: One existing package extended additively — no
new package, no new dependency, no CLI change. The smallest Go
footprint of any feature since 011 (which itself added no CLI surface
either) — this phase is purely a new, pure-function data model layered
on top of what 011 already reads.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
