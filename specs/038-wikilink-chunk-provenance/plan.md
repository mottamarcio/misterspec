# Implementation Plan: Wikilinks with Chunk-Level Provenance

**Branch**: `038-wikilink-chunk-provenance` | **Date**: 2026-09-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/038-wikilink-chunk-provenance/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today a wikilink is resolved down to `ReferenceEntry{Relation, Target}` (`internal/operations/references.go`) and indexed as a `links` row with only `{source_artifact_id, target_artifact_id, relation}` (`internal/context/index/schema.go`) — `artifacts.WikiLink`'s own `Line` field, and which section it falls under, are both discarded before a Context Pack item's `Reasons` ever get built, so a retrieved item can say "SPEC-014 references this" but never "from its own Requirements section, line 42." The technical approach: (1) correlate each `WikiLink`'s already-known `Line` against the same `artifacts.Chunks` segmentation the codebase already computes everywhere else, producing a `ReferenceOccurrence` (source path, section heading, line) threaded through `ReferenceEntry`/`BacklinkEntry` — no second Markdown parser, no new segmentation logic (Principle VI); (2) widen the index's `links` table with the same fields and bump its `schemaVersion` (1→2), which the existing version-mismatch rebuild path already handles automatically (Principle III, no manual migration); (3) thread occurrence data through `Candidate`/`Reason` into the `context` command's response behind a new opt-in `--provenance` flag (mirroring 036's `--diagnostic-scores` precedent), bumping `contextSchemaVersion` 3→4; (4) add occurrence detail as additive, backward-compatible fields to the already-unversioned `references`/`backlinks` command output; (5) implement the Requirements/active-task-section reference preference (spec User Story 2) as an explicit, off-by-default scoring capability that never changes default ordering until a future Spec promotes it with recorded evidence from the evaluation harness (037-eval-quality-efficiency) — exactly the `ranking_version` promotion discipline 036 already established.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `modernc.org/sqlite` v1.36.3 (existing disposable FTS5 index — `links` table only, no new virtual table); `github.com/spf13/cobra` v1.10.2 (CLI). No new third-party dependency.
**Storage**: SQLite, disposable/reconstructable index at `internal/context/index` (Constitution Principle III). The `links` table's schema widens; `schemaVersion` bumps 1→2, triggering the index's own existing automatic rebuild-on-mismatch path — no manual migration code.
**Testing**: `go test` — unit tests for WikiLink-line-to-section correlation, `ReferenceEntry`/`BacklinkEntry` provenance fields, `links` table schema/sync; filesystem integration tests for cycle- and hub-boundedness regression (existing 2-hop cap, exercised not reinvented); golden-style CLI envelope tests for `context --provenance` and the additive `references`/`backlinks` fields.
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows), no network access required.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: no new performance target; preserves the existing bounded, non-recursive 2-hop expansion (`internal/context/collector.go`'s `secondHopCandidates` — first hop from target, one further hop, never a third) as the mechanism satisfying spec FR-008/SC-002, not a new limiter.
**Constraints**: deterministic, reproducible output (Constitution Principle I, V, IX); no new persistent/authoritative state beyond the existing disposable index (Principle III); a reference occurrence MUST NOT be presented or stored as a formal dependency (Principle I, spec FR-009); the Requirements/active-task-section preference (spec FR-005/FR-006) MUST ship off by default and MUST NOT be promoted to default ordering without a recorded comparison through 037-eval-quality-efficiency's harness (Principle IV, spec FR-006); `context` command JSON contract changes bump `contextSchemaVersion` (033/035/036 precedent, now 3→4 for this feature).
**Scale/Scope**: `internal/artifacts` (new line→enclosing-chunk correlation helper, reusing existing `Chunks`), `internal/operations/references.go` and `backlinks.go` (`ReferenceEntry`/`BacklinkEntry` gain occurrence fields), `internal/context/index/schema.go` and `sync.go` (`links` table widened, `schemaVersion` bump), `internal/context/collector.go`/`rank.go`/`result.go`/`pack.go` (`Reason` carries occurrence data; new off-by-default preference scoring path), `internal/cli/internalcmd/context.go` (`--provenance` flag, `contextSchemaVersion` 4), `internal/cli/internalcmd/references.go` and `backlinks.go` (additive per-entry fields, no version bump needed — backward compatible). No new package layer.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Recording *where* a reference was written (source path/section/line, correlated against already-computed chunk boundaries) is fully mechanical. *Whether* to prefer a reference because of where it was written is a ranking-policy question — this feature computes the signal but ships the preference itself gated and off by default (spec FR-006), never silently decided as "better" by the binary.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. No new entity, ID, or canonical path is created; occurrence data is read-side enrichment of already-existing reference computation.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. The widened `links` table remains fully reconstructable from Markdown; the existing `schemaVersion`/`PRAGMA user_version` mismatch-triggers-rebuild path (unchanged mechanism, `internal/context/index/schema.go`) is reused verbatim for FR-010's "automatically rebuilt, no manual migration" requirement.
- **Principle IV (Simplicity First — YAGNI)**: PASS with an explicit boundary: no new package layer, no second Markdown/section parser — occurrence correlation reuses `artifacts.Chunks`'s existing segmentation. The preference-policy capability (spec User Story 2) is scoped to "compute the signal, expose it opt-in, never apply it by default" for this feature; actually promoting it to default ranking is explicitly deferred to a future Spec once 037's harness records evidence, per Principle IV's own prohibition on speculative behavior changes.
- **Principle V (Test-First Discipline)**: Applies — line-to-section correlation, occurrence field plumbing, schema/sync changes, and the cycle/hub-boundedness regression are exactly the deterministic logic this principle requires unit/integration tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. Correlation logic is added to `internal/artifacts` (which already owns chunking/section segmentation) and `internal/operations` (which already owns reference/backlink computation) — not duplicated into `internal/context` or `internal/cli`.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. No new filesystem writes; this feature only enriches read-side data already computed from existing artifacts.
- **Principle VIII (Safety by Construction)**: PASS. No new mutating operation; the index's own existing atomic-transaction sync/rebuild path is reused unchanged.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — the `context` command's new `--provenance` field is versioned (`contextSchemaVersion` 3→4, opt-in, additive — 036's `--diagnostic-scores` precedent); `references`/`backlinks` gain additive, backward-compatible per-entry fields.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/038-wikilink-chunk-provenance/
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
│   ├── wikilink.go          # WikiLink already has Line; unchanged parsing
│   ├── chunk.go              # Chunks() unchanged; new helper correlates a line to its enclosing Chunk's Heading
│   └── occurrence.go         # New: ReferenceOccurrence type + line→section correlation helper
├── operations/
│   ├── references.go         # ReferenceEntry gains occurrence fields, populated via the new correlation helper
│   ├── references_test.go
│   ├── backlinks.go           # BacklinkEntry gains the same occurrence fields
│   └── backlinks_test.go
├── context/
│   ├── result.go               # Reason gains occurrence fields (source path/section/line)
│   ├── collector.go             # connectedCandidates/secondHopCandidates thread occurrence data through
│   ├── collector_test.go
│   ├── rank.go                  # New, off-by-default preference scoring path (spec FR-005/FR-006) — never affects default ordering
│   ├── rank_test.go
│   └── index/
│       ├── schema.go            # links table widened; schemaVersion 1 → 2
│       ├── schema_test.go
│       ├── sync.go               # indexLinks populates the new columns via the same operations.References/Backlinks call
│       └── sync_test.go
└── cli/
    └── internalcmd/
        ├── context.go            # --provenance flag; contextSchemaVersion 3 → 4
        ├── context_test.go
        ├── references.go          # additive per-entry occurrence fields, no version bump
        ├── references_test.go
        ├── backlinks.go
        └── backlinks_test.go
```

**Structure Decision**: Single Go project, no new top-level directory and no new package layer. This feature is a targeted enrichment of existing packages (`internal/artifacts`, `internal/operations`, `internal/context`, `internal/context/index`, `internal/cli/internalcmd`), consistent with Constitution Principle IV.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations — table intentionally omitted.
