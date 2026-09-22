# Implementation Plan: Referências a Seções com Âncoras Estáveis

**Branch**: `040-stable-section-anchors` | **Date**: 2026-09-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/040-stable-section-anchors/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today a wikilink can only ever address a whole target artifact —
`internal/artifacts/wikilink.go`'s `WikiLink.Target` is the entity ID
alone, and `internal/context/collector.go`'s `connectedCandidates` turns
every reference into candidates for *every* Chunk of the target
(`chunkArtifact`), regardless of which passage the author actually
meant. The technical approach: (1) extend the ATX heading parser
(`internal/artifacts/document.go`) to recognize an explicit, opt-in
`{#slug}` anchor suffix on any heading, stripped into a new
`Section.Anchor`/`Chunk.Anchor` field, stable across a title edit
(research.md #1); (2) extend the wikilink lexical scanner
(`internal/artifacts/wikilink.go`) to recognize `[[ID#anchor]]`/
`[[ID#anchor|Alias]]`, adding `WikiLink.Anchor` alongside the unchanged
`Target` (research.md #2); (3) add two validation codes —
`CodeUnknownAnchor` (target artifact exists, anchor doesn't) and
`CodeDuplicateAnchor` (two Sections in one artifact declare the same
anchor) — extending `internal/validation/wikilinks.go`'s existing
three-code classification rather than overloading it (research.md #3);
(4) add `chunkArtifactAnchor` alongside the existing `chunkArtifact` in
`internal/context/collector.go`, so an anchor-qualified reference
contributes exactly one Candidate (that Section's own Chunk, even when
its Body is empty) plus a `HeadingPath` ancestor-title breadcrumb,
never the whole target artifact (research.md #4); (5) widen the
disposable SQLite index's `chunks`/`links` tables and bump
`schemaVersion` 2→3 and `contextSchemaVersion` 4→5, following 038's own
additive-field-bumps-version precedent exactly (research.md #5). Every
new code path only activates when an anchor is actually present —
`WikiLink.Anchor == ""` (all existing content) is provably unchanged
behavior end to end (research.md #6, spec FR-008).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `modernc.org/sqlite` v1.36.3 (existing disposable index — `chunks`/`links` tables widen, no new virtual table); `github.com/spf13/cobra` v1.10.2 (CLI). No new third-party dependency.
**Storage**: SQLite, disposable/reconstructable index at `internal/context/index` (Constitution Principle III). `schemaVersion` bumps 2→3, triggering the index's already-existing version-mismatch rebuild path — no manual migration code, same mechanism 038 already exercised for its own bump.
**Testing**: `go test` — unit tests for heading-anchor-suffix parsing (`internal/artifacts/document_test.go`), wikilink anchor/target/alias splitting (`internal/artifacts/wikilink_test.go`), the two new validation codes (`internal/validation/wikilinks_test.go`), `chunkArtifactAnchor`'s single-Chunk-plus-breadcrumb behavior and its empty-Body-Section case (`internal/context/collector_test.go`), schema/sync widening (`internal/context/index/schema_test.go`, `sync_test.go`), and the additive `context --provenance`/`references`/`backlinks` output fields (existing golden-envelope test patterns).
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows), no network access required.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: no new performance target; an anchor-qualified reference is strictly cheaper than today's whole-artifact chunking (one Candidate instead of N), never more expensive.
**Constraints**: deterministic, reproducible output (Constitution Principle I, V, IX); no new persistent/authoritative state beyond the existing disposable index (Principle III); an anchor reference MUST remain a reference — never embedded/transcluded content (Principle I, spec FR-007); every new behavior is reached only when `WikiLink.Anchor != ""`, so 100% of existing wikilinks/Sections/Chunks are provably unaffected (spec FR-008); `context`/index JSON contract changes bump `contextSchemaVersion`/`schemaVersion` (033/035/036/038 precedent, now 4→5 and 2→3 respectively for this feature).
**Scale/Scope**: `internal/artifacts` (heading-anchor parsing in `document.go`, `Section`/`Chunk` gain `Anchor`; wikilink anchor parsing in `wikilink.go`, `WikiLink` gains `Anchor`); `internal/validation/wikilinks.go` (two new codes, `findings.go` gains their constants); `internal/operations/references.go`/`backlinks.go` (`ReferenceEntry`/`BacklinkEntry` gain `TargetAnchor`); `internal/context/collector.go`/`rank.go`/`result.go`/`pack.go` (`chunkArtifactAnchor`, `Candidate`/`Reason` gain `HeadingPath`/`TargetAnchor`); `internal/context/index/schema.go`/`sync.go` (`chunks`/`links` widened, `schemaVersion` 2→3); `internal/cli/internalcmd/context.go` (`contextSchemaVersion` 4→5); `internal/cli/internalcmd/references.go`/`backlinks.go` (additive `target_anchor` field, no version bump — unversioned envelopes, 038 precedent). No new package layer.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Anchor declaration, parsing, existence/uniqueness checking, and single-Chunk selection are all fully mechanical — no judgment about *which* anchor an author "should" have meant. The HeadingPath breadcrumb is computed straight from already-parsed Section nesting, not inferred.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. No new entity, ID, or canonical path is created; an anchor is author-declared Markdown text, read-side from this feature's own perspective (no new `internal create*` behavior).
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. `Section.Anchor`/`Chunk.Anchor` are derived fresh from Markdown on every parse, never stored authoritatively elsewhere; the widened index rebuilds automatically on the `schemaVersion` mismatch, exactly 038's already-proven mechanism.
- **Principle IV (Simplicity First — YAGNI)**: PASS with an explicit boundary: no new package layer, no second Markdown parser (the `{#slug}` suffix extends the existing `headingPattern`/`Section` model in place); `HeadingPath` is a plain title list, not a richer navigation structure; anchor resolution reuses `ids.ResolveTarget` unchanged rather than inventing a parallel resolution path.
- **Principle V (Test-First Discipline)**: Applies — heading-suffix parsing, wikilink anchor/alias splitting, the two new validation codes, `chunkArtifactAnchor`'s Candidate/breadcrumb shape (including the empty-Body edge case), and schema/sync widening are exactly the deterministic logic this principle requires unit/integration tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. Anchor parsing stays in `internal/artifacts` (which already owns heading/section parsing); anchor validation stays in `internal/validation` (which already owns wikilink classification); anchor-aware retrieval stays in `internal/context` (which already owns Chunk-to-Candidate assembly) — no cross-package logic duplication.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. No new filesystem writes; this feature only enriches read-side parsing/validation/retrieval of already-existing artifacts.
- **Principle VIII (Safety by Construction)**: PASS. No new mutating operation; the index's own existing atomic rebuild-on-mismatch path is reused unchanged.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — `contextSchemaVersion` bumps 4→5 for the new `heading_path`/`anchor` fields (033/035/036/038 precedent); `references`/`backlinks` gain an additive, unversioned `target_anchor` field; the two new validation codes are stable, documented `Code*` constants, not ad hoc message text.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/040-stable-section-anchors/
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
│   ├── document.go            # headingPattern extended for optional trailing "{#slug}"; Section gains Anchor
│   ├── document_test.go
│   ├── chunk.go                # Chunk gains Anchor, copied from Section.Anchor
│   ├── chunk_test.go
│   ├── wikilink.go             # extractLineLinks/splitTargetAlias extended for "TARGET#anchor"; WikiLink gains Anchor
│   └── wikilink_test.go
├── validation/
│   ├── findings.go             # CodeUnknownAnchor, CodeDuplicateAnchor constants
│   ├── wikilinks.go            # classifyWikilink gains the anchor-existence branch; new checkAnchors for duplicate-anchor detection
│   └── wikilinks_test.go
├── operations/
│   ├── references.go           # ReferenceEntry gains TargetAnchor, populated from the resolved WikiLink
│   ├── references_test.go
│   ├── backlinks.go            # BacklinkEntry gains TargetAnchor
│   └── backlinks_test.go
├── context/
│   ├── result.go                # Candidate gains HeadingPath; Reason gains TargetAnchor
│   ├── collector.go             # New chunkArtifactAnchor; connectedCandidates branches to it when the driving entry is an anchor-qualified wikilink
│   ├── collector_test.go
│   └── index/
│       ├── schema.go            # chunks gains anchor column, links gains target_anchor column; schemaVersion 2 → 3
│       ├── schema_test.go
│       ├── sync.go               # indexChunks/indexLinks populate the new columns
│       └── sync_test.go
└── cli/
    └── internalcmd/
        ├── context.go            # contextSchemaVersion 4 → 5; heading_path/anchor fields on context items
        ├── references.go         # additive target_anchor field, no version bump (unversioned envelope, 038 precedent)
        └── backlinks.go          # same additive field
```

**Structure Decision**: Single Go project, no new top-level directory —
this feature widens the same five packages 038 (wikilink chunk
provenance) already touched, following the identical shape: parsing in
`internal/artifacts`, classification in `internal/validation`,
occurrence plumbing in `internal/operations`, retrieval in
`internal/context` (+ its `index` subpackage), and versioned CLI output
in `internal/cli/internalcmd`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. Table intentionally omitted.
