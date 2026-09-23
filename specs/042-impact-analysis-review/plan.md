# Implementation Plan: Análise de Impacto e Revisão Incremental

**Branch**: `042-impact-analysis-review` | **Date**: 2026-09-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/042-impact-analysis-review/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today a maintainer who changes a Requirement's text, a Plan section, a
Knowledge note, or code has no way to ask "what does this affect" —
`operations.Backlinks` (012, extended by 038) already answers "what
references this artifact" for a single artifact, and
`evidence.DeriveState` (041) already flags a Task whose *own* body
content drifted from its recorded evidence fingerprint, but nothing
connects the two: a change to a Requirement's *text* inside `spec.md`
never propagates to the Tasks whose `Serves:` line names it, and there
is no report of *why* an affected item was included, only a raw
backlink list. The technical approach: (1) two new small
`internal/vcs` primitives (`DiffNameStatus`, `FileAtRevision`) let the
project answer "what changed between two Git revisions" without any
new persisted state (research.md #1/#2); (2) a new
`validation.RequirementSections` export (additive sibling to the
existing unexported `parseSpecRequirements`) gives per-`R<N>`
fingerprinting, the one genuinely new granularity this feature needs
(research.md #3); (3) a new `internal/impact` package, layered like
`internal/prepare`, drives the whole computation by reusing
`operations.Backlinks` for formal/wikilink reverse edges,
`validation.ParseTaskCoverage` for the Requirement→Task reverse edge,
and `evidence.DeriveState` as-is for a Task's own staleness — nothing
here recomputes or changes what those three already-shipped features
decide (research.md #4/#5/#10); (4) reverse traversal terminates via a
simple visited-element set, not a hop cap, since this feature's job is
completeness rather than context-budget bounding (research.md #6);
(5) a fixed, unconditional classification table guarantees a wikilink
mention can never become a deterministic invalidation (research.md
#7); (6) one new deterministic operation and CLI command,
`internal analyze-impact`, exposes all of this as a single read-only,
JSON-contract report (contracts/analyze-impact-contract.md).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `github.com/spf13/cobra` v1.10.2 (CLI). `os/exec` for the two new `git diff`/`git show` shell-outs, the same idiom `internal/vcs`'s existing `CommitsSinceFileAdded`/`HeadCommit`/`IsWorkingTreeDirty` already use (research.md #2). No new third-party dependency.
**Storage**: Filesystem + Git history only (Constitution Principle III). No new persisted state of any kind — `analyze-impact` recomputes its entire `ImpactReport` on every call from two Git revisions plus current filesystem content; there is no "last analyzed" marker to keep in sync (research.md #1).
**Testing**: `go test` — unit tests for `vcs.DiffNameStatus`/`FileAtRevision` (added/modified/removed paths, a revision that doesn't resolve, working-tree comparison), `validation.RequirementSections` (added/changed/removed Requirement numbers between two body snapshots), the new `internal/impact` package's classification/severity tables (research.md #7) and visited-set termination on a constructed reference cycle fixture (research.md #6); filesystem+Git integration tests building a real fixture repo with two commits to prove the Requirement→Task and depends_on→Spec propagation end-to-end; a golden-style CLI envelope test for `internal analyze-impact`'s success and error shapes.
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows) — the two new Git shell-outs have the same portability expectations `internal/vcs`'s existing functions already carry.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: no new performance target; `analyze-impact`'s cost is bounded by the number of distinct artifacts the visited-element set ever expands (research.md #6), each expansion being one `operations.Backlinks` full-project scan — the same per-call cost `references`/`backlinks` commands already accept today, not a new hot path.
**Constraints**: read-only — `AnalyzeImpact` MUST NOT mutate the filesystem (Constitution Principle VII); a `wikilink`-only propagation path MUST NOT be classified as `deterministic_invalidation` under any input (Principle I, spec FR-004); classification/severity MUST be a fixed, input-deterministic table, never a tunable score (Principle IV, spec FR-010); a changed source-code path with no Requirement mapping MUST be declared (`unmapped_code_paths`), never silently dropped (spec FR-007); this feature MUST NOT change 041's own `EvidenceState`/`Evidence-*:` semantics (Principle VII, research.md #10).
**Scale/Scope**: `internal/vcs` (two new functions + `DiffEntry`); `internal/validation` (one new additive export, `RequirementSections`); new `internal/impact` package (`ChangeSet`/`ChangedElement`/`PropagationHop`/`PropagationPath`/`AffectedItem`/`ImpactReport` types, `AnalyzeImpact`); `internal/cli/internalcmd` (new `analyze_impact.go` command, `analyze_impact_schema_version = 1`); `internal/cli/internalcmd/errors.go` (two new stable error codes: `revision_not_found`, `not_a_repository`). No new package layer beyond `internal/impact`, which mirrors `internal/prepare`'s existing shape rather than introducing a new architectural pattern (Constitution Principle VI).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. What changed (fingerprint diff), what formally/semantically/by-coverage relates to it, and the fixed classification/severity table are all mechanically computed. Deciding *whether* a suggested-review item actually needs action, and *how* to re-verify a flagged Task, stays the agent's judgment — `reverification_candidate` names a Task, never a command to run automatically.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. `AnalyzeImpact` creates no entity, ID, or canonical path — pure read/report.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. No persisted "last analyzed" state; every `ImpactReport` is recomputed from two Git revisions (already-durable project history) plus current files (research.md #1).
- **Principle IV (Simplicity First — YAGNI)**: PASS with explicit boundaries (research.md #6, #7, #8, #9): no new hop-cap concept borrowed from 038 where it wouldn't fit; no numeric/configurable scoring engine; no new addressable "Plan" entity type; no code→Requirement mapping invented ahead of PROP-13. Each was considered and deliberately deferred or rejected as more than this feature's own gap requires.
- **Principle V (Test-First Discipline)**: Applies — the new `vcs` diff/show primitives, `RequirementSections`, the classification/severity tables, and visited-set termination are exactly the deterministic logic this principle requires unit/integration tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. Git facts stay in `internal/vcs` (which already owns every Git shell-out); Requirement parsing stays in `internal/validation` (which already owns it, via an additive export, not a duplicate parser); reverse-relation orchestration lives in its own new `internal/impact` package rather than bloating `operations`, `validation`, or `evidence` with a fourth/fifth responsibility none of them currently has — `internal/impact` *consumes* all three instead.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. `AnalyzeImpact` performs no filesystem writes at all — pure read/report, the simplest possible case of this principle. It also does not rewrite or reinterpret 041's own `Evidence-*:` fields or `EvidenceState` semantics (research.md #10) — a Task's own evidence staleness is read via the unmodified `evidence.DeriveState`, never recomputed with different rules.
- **Principle VIII (Safety by Construction)**: PASS. `--path` resolution (when supplied) reuses `artifacts.RelativeWithinRoot`, the same guarantee `Fingerprint`/`capture-evidence` already give; the two new `vcs` functions only ever read (`git diff`/`git show`/`os.ReadFile`), never write.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — `internal analyze-impact` returns a versioned (`analyze_impact_schema_version`), structured JSON envelope; `revision_not_found` and `not_a_repository` are stable, distinct error codes rather than a generic failure message.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/042-impact-analysis-review/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md          # Phase 1 output (/speckit-plan command)
├── quickstart.md          # Phase 1 output (/speckit-plan command)
├── contracts/             # Phase 1 output (/speckit-plan command)
│   └── analyze-impact-contract.md
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── vcs/
│   └── vcs.go                       # + DiffNameStatus, FileAtRevision, DiffEntry
├── validation/
│   └── requirements.go              # + RequirementSections (additive; parseSpecRequirements unchanged)
├── impact/                          # NEW package (this feature)
│   ├── doc.go
│   ├── change_set.go                # ChangeSet/ChangedElement, git-diff-driven
│   ├── propagate.go                 # PropagationHop/PropagationPath, visited-set walk
│   ├── classify.go                  # Classification/Severity fixed tables
│   └── analyze.go                   # AnalyzeImpact (orchestrates the above)
└── cli/internalcmd/
    ├── analyze_impact.go            # NEW: `misterspec internal analyze-impact`
    └── errors.go                    # + revision_not_found, not_a_repository codes

tests/ (co-located _test.go files per existing repo convention)
├── internal/vcs/vcs_test.go                     # + diff/show cases
├── internal/validation/requirements_test.go     # + RequirementSections cases
├── internal/impact/*_test.go                    # new unit + fixture-repo integration tests
└── internal/cli/internalcmd/analyze_impact_test.go
```

**Structure Decision**: Single Go project (existing `internal/...`
layout, no new top-level directory). This feature adds exactly one new
package, `internal/impact`, at the same layer as the existing
`internal/prepare` (both sit above `operations`/`validation`/
`evidence`/`vcs`/`ids`/`artifacts`, avoiding the only import-cycle risk
in this codebase: `operations` already imports `validation`, so
`validation` must never import `operations`, and neither may import
`internal/impact`). Every other change is additive to an existing
package the corresponding responsibility already lives in
(`internal/vcs` for Git, `internal/validation` for Requirement
parsing, `internal/cli/internalcmd` for the new command) — no
speculative new layer beyond that one package (Constitution
Principle IV).

## Complexity Tracking

*No violations — table omitted.*
