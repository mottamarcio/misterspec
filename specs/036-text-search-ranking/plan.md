# Implementation Plan: Robust Text Search and Measurable Ranking

**Branch**: `036-text-search-ranking` | **Date**: 2026-09-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/036-text-search-ranking/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Tier 4 free-text retrieval (`internal/context/collector.go`) passes `req.Query`/`req.Task` straight into `internal/context/index/search.go`'s FTS5 `MATCH` clause with no escaping, so ordinary punctuation can throw a syntax error instead of returning results. Separately, `SearchResult.Rank` already carries SQLite's own `bm25()` value, but `internal/context/rank.go`'s `textRelevance` ignores it and recomputes a capped term-occurrence count instead — the real relevance signal is computed then thrown away. The technical approach: (1) tokenize/escape free-text queries into a safe FTS5 `MATCH` expression before they reach `Search`, with an explicit, separate advanced-syntax path for callers who want raw FTS5 syntax; (2) thread `SearchResult.Rank` through `Candidate`/`ScoredCandidate` so same-tier ordering uses the real BM25 value instead of `textRelevance`'s heuristic, while leaving the tier-first sort invariant (016-ranking-budgeting) untouched; (3) add an opt-in diagnostic view that exposes score components (tier, BM25, relation weight, intent bonus) and a `ranking_version` marker on context output so weight changes are identifiable and gated by 019's existing evaluation protocol before becoming default.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `modernc.org/sqlite` v1.36.3 (FTS5 virtual tables), `github.com/spf13/cobra` v1.10.2 (CLI)
**Storage**: SQLite, disposable/reconstructable index at `internal/context/index` (Constitution Principle III — no authoritative secondary state)
**Testing**: `go test` — unit tests for query tokenization/escaping and score-component composition, filesystem/index integration tests via the existing `fixture_test.go`/`sqlite_test.go` pattern, golden-style assertions on the CLI JSON envelope
**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), no network access required
**Project Type**: Single Go project (CLI + internal libraries)
**Performance Goals**: No new performance target introduced; preserves the existing `textSearchLimit = 100` bound on Tier 4 candidates (016-ranking-budgeting)
**Constraints**: Deterministic and reproducible output (Constitution Principle I, V); no new persistent/authoritative state (Principle III); contract changes to the `context` command's JSON envelope MUST bump `contextSchemaVersion` (currently 2) and stay backward-compatible for existing fields (033-context-pack-output-contract precedent)
**Scale/Scope**: Confined to `internal/context/index/search.go`, `internal/context/rank.go`, `internal/context/collector.go`, `internal/context/index/store.go`, and the `internal/cli/internalcmd/context.go` output envelope; no new package layer

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Query escaping, BM25 plumbing, and tie-break ordering are all mechanical/computable — they belong in the Go binary, not agent reasoning. No new semantic judgment is being pushed onto the LLM or pulled into the binary.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. This feature is read-only retrieval; it creates no entities, IDs, or paths.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. The FTS5 index remains disposable and reconstructable (`schemaVersion` in `schema.go`); no new authoritative counter or store is introduced. A `ranking_version` value is a constant tag on output, not persisted state.
- **Principle IV (Simplicity First — YAGNI)**: PASS. No new package layer; advanced-syntax mode is a second, explicit path through the existing `Search` call, not a new subsystem. Diagnostic score-component output extends the existing `--mode`/diagnostics machinery from 033/035 rather than inventing a new one.
- **Principle V (Test-First Discipline)**: Applies — tokenization/escaping and score-composition are exactly the kind of deterministic logic this principle requires unit tests for, written alongside implementation. Planned in Phase 1/tasks.
- **Principle VI (Clean Code & SOLID)**: PASS. Query-mode handling stays in `index`/`context`, owned by the packages that already own search and ranking respectively; no cross-package logic duplication.
- **Principle VII (Explicit Mutation Boundaries)**: N/A. No artifact files are written by this feature.
- **Principle VIII (Safety by Construction)**: PASS. No filesystem mutation; only defends against malformed query strings reaching FTS5.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — FR-004 (clear diagnostic for malformed advanced queries), FR-008 (score components in diagnostic output), and FR-009 (versioned ranking) are this principle's requirements applied to search/ranking specifically. The `context` command's `schema_version` bump path (033/035 precedent) is the mechanism.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/036-text-search-ranking/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── context/
│   ├── index/
│   │   ├── search.go       # FTS5 Search() — query mode handling, escaping lives here
│   │   ├── search_test.go
│   │   └── store.go        # SearchResult already carries Rank (bm25); Store interface
│   ├── collector.go         # Tier 4 free-text collection — passes query + mode to Search
│   ├── rank.go               # Score() / textRelevance — consumes SearchResult.Rank instead of recomputing
│   └── rank_test.go
└── cli/
    └── internalcmd/
        └── context.go        # --query-mode flag (or equivalent), diagnostic score components,
                                # ranking_version, contextSchemaVersion bump

tests/                        # existing repo-level integration/fixture tests, extended in place
```

**Structure Decision**: Single Go project, no new top-level directory. This feature is a targeted, mechanical fix confined to the existing `internal/context` and `internal/context/index` packages plus the CLI envelope in `internal/cli/internalcmd/context.go`, consistent with Constitution Principle IV (no speculative new layers).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations — table intentionally omitted.
