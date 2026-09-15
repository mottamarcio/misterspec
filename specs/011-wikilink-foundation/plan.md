# Implementation Plan: Wikilink Graph Foundation

**Branch**: `011-wikilink-foundation` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-wikilink-foundation/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 1 ("Artifact
Graph Foundation") exactly: a `[[TARGET]]`/`[[TARGET|Alias]]` wikilink
syntax an author can write in any artifact's body, a pure lexical
extractor for it, and three new structural-validation finding codes
(`invalid_wikilink`, `broken_wikilink`, `ambiguous_wikilink`) that reuse
`internal/ids`'s already-existing syntax/resolution rules exactly — no
new ID semantics, no SQLite, no reference/backlink query operations
(that document's own Phase 2), no CLI change at all. Two new files in
`internal/artifacts` (`wikilink.go`, `markdown.go`), one small,
backward-compatible widening of an existing unexported helper
(`extractFrontmatter` → `splitFrontmatter`), and one new file in
`internal/validation` (`wikilinks.go`) are the feature's entire Go
footprint. `internal/cli/internalcmd/validate.go`'s already-generic
Finding→JSON mapping needs no change at all — the strongest evidence
yet that this project's established layering (deterministic core,
generic CLI adapter) is paying off.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: Go standard library only — no new dependency (research.md rejects a Markdown-parsing library as disproportionate for this feature's actual scope).
**Storage**: Filesystem, unchanged. This feature reads artifact bodies (a new capability, `artifacts.ReadBody`) but writes nothing — no new mutation anywhere (FR-010, mirroring 004-structural-validation's own read-only nature).
**Testing**: `go test` — `internal/artifacts`'s new extraction/body-reading tests (matching `docs/context-engine-implementation.md` §29.1's own test list: simple/aliased/multiple/multi-line links, malformed and non-link syntax, line-number preservation, ordering, code-span/fenced-block exclusion); `internal/validation`'s new classification tests (§29.2's list: valid, missing, malformed target, wrong configured width, ambiguous target, alias-does-not-affect-resolution); 004-structural-validation's and 008-cli-cobra's own full suites re-run unmodified as named, explicit regression gates for US3 (SC-003). Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module. Two existing packages extended additively (`internal/artifacts`, `internal/validation`); no new package, no new external dependency, no CLI-layer change.
**Performance Goals**: Trivial — the expected corpus is hundreds to thousands of small Markdown artifacts (`docs/context-engine-implementation.md` §27's own stated scale); a per-artifact line scan and a handful of `ids.Scan` calls are already proven cheap at this scale by 004-structural-validation's existing checks.
**Constraints**: `ExtractWikiLinks` must never resolve a target during parsing (FR-004) — resolution is `internal/validation`'s job, strictly separate, per `docs/context-engine-implementation.md` §6.3's own explicit rule. `ParseMetadata`'s existing behavior must remain byte-for-byte unchanged (US3, SC-003). No new Finding code may bypass the existing generic JSON mapping (research.md). Wikilink checking is scoped to exactly the five entity types `checkEntity` already validates — Task, Plan, Tasks, Validation, and Constitution are explicitly out of scope (research.md).
**Scale/Scope**: `internal/artifacts/{wikilink.go, wikilink_test.go, markdown.go, markdown_test.go}` (new); `internal/artifacts/parser.go` (modified: `extractFrontmatter` → `splitFrontmatter`); `internal/validation/{wikilinks.go, wikilinks_test.go}` (new); `internal/validation/{findings.go, validator.go}` (modified: three new Finding codes, one new call site in `checkEntity`). No new package, no CLI change, no new dependency.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** Extraction and classification are purely mechanical (lexical scan, ID syntax/existence checks) — *which* artifacts to link, and what a link means, remains an authoring (human or agent) decision this feature never makes or infers. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass (N/A, distinct boundary).** This feature performs zero mutation — it only reads artifact bodies and reports findings, the same read-only boundary 004-structural-validation itself already established. |
| III. Filesystem Is Single Source of Truth | **Pass.** No derived state is cached anywhere — every validation run re-reads and re-extracts fresh, exactly like every existing `checkEntity` check already does. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** No Markdown-parsing dependency where a small hand-written scanner already fully covers this project's real content (research.md); no new "ambiguous ID" concept invented when `ids.Scan`'s existing duplicate-path notion already covers it exactly; wikilink-checking deliberately not extended to artifact types that have no structural validation at all today. |
| V. Test-First Discipline | **Gate carried into tasks.** `docs/context-engine-implementation.md` §29.1-29.2's own test lists are the explicit floor; 004's and 008's full suites are named regression gates, not just "the usual suite." |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `splitFrontmatter` widens rather than duplicates `extractFrontmatter`'s already-correct delimiter logic; the three new Finding codes reuse `ids.ParseAny`/`ids.Scan` exactly rather than a parallel ID-validity notion — the clearest DRY exercise since 007-project-bootstrap's `writeDefaultConfig`. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This feature writes nothing at all — its only new filesystem interaction (`ReadBody`) is a read, and it is exercised only from within the already-read-only `internal/validation` package. |
| VIII. Safety by Construction | **Pass (N/A).** No write path exists in this feature to make safe or unsafe. |
| IX. Transparent, Machine-Readable Contracts | **Pass, reinforced without any new code.** The three new Finding codes flow through `internal/cli/internalcmd/validate.go`'s already-generic, already-tested JSON mapping (008-cli-cobra) with zero lines changed there — direct evidence the "generic adapter over a deterministic core" contract this project committed to is holding up under a genuinely new kind of finding. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/011-wikilink-foundation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── wikilinks.md        # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # unchanged — no new dependency
├── cmd/misterspec/, internal/cli/         # unchanged
├── internal/
│   ├── project/, ids/, testutil/, example/   # unchanged
│   ├── lock/, templates/, operations/         # unchanged
│   ├── installer/, agents/, bootstrap/, tui/   # unchanged
│   ├── artifacts/
│   │   ├── types.go, metadata.go                # unchanged
│   │   ├── parser.go               # MODIFIED: extractFrontmatter → splitFrontmatter
│   │   ├── parser_test.go          # MODIFIED: + regression coverage for the widened helper
│   │   ├── paths.go                 # unchanged
│   │   ├── markdown.go               # NEW — ReadBody
│   │   ├── markdown_test.go           # NEW
│   │   ├── wikilink.go                 # NEW — WikiLink, ExtractWikiLinks
│   │   └── wikilink_test.go             # NEW
│   └── validation/
│       ├── findings.go              # MODIFIED: + 3 new Code constants
│       ├── tables.go                 # unchanged
│       ├── validator.go               # MODIFIED: + checkWikilinks call site in checkEntity
│       ├── wikilinks.go                # NEW — checkWikilinks
│       └── wikilinks_test.go            # NEW
└── kit/                                    # unchanged
```

**Structure Decision**: Two existing packages extended additively — no
new package, matching this feature's genuinely narrow scope (parsing +
validation only, per `docs/context-engine-implementation.md`'s own
Phase 1 boundary). `internal/cli` is untouched (research.md) — the
first feature since 001-core-foundation to add real new capability
without touching the CLI layer at all, a direct consequence of that
layer's own established genericness.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
