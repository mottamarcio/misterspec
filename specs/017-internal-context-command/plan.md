# Implementation Plan: Internal Context Command

**Branch**: `017-internal-context-command` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/017-internal-context-command/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement `docs/context-engine-implementation.md`'s Phase 8 ("Internal
Context Command") exactly: one new hidden CLI command, `misterspec
internal context <id>`, wired into the existing `internal` command tree
(008-cli-cobra), that orchestrates the pipeline 015/016 already built —
`index.Open` + `Store.Sync` (014, transparently establishing "Index
Readiness" per FR-007 with zero new indexing logic), `contextengine.
Collect` (015), `contextengine.Rank` and `contextengine.ApplyBudget`
(016) — and renders the result as one stable JSON envelope (Constitution
Principle IX), with an optional `--render` flag that additionally
includes a Markdown context pack computed by one new pure function,
`contextengine.Render`. No new ranking or budgeting logic; no Skill
integration (Phase 10, out of scope per spec.md's own Assumptions).

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: Go standard library, `github.com/spf13/cobra` (existing), `modernc.org/sqlite` (existing, via 014's already-vendored `internal/context/index`) — no new dependency.
**Storage**: The existing disposable SQLite index (014) at a new fixed, non-authoritative cache path, `<project root>/.misterspec/cache/context.db` (docs/context-engine-implementation.md §10.2) — opened and incrementally synchronized on every invocation (FR-007), never treated as authoritative (Constitution Principle III). No new schema; `index.Open`'s existing `ensureSchema` already recreates the database from scratch on any schema-version mismatch (research.md #2), so this feature adds no new "rebuild on incompatibility" logic of its own.
**Testing**: `go test` — new `internal/cli/internalcmd` CLI tests following the existing `references_test.go`/`backlinks_test.go` pattern (successful request, unknown target, ambiguous target, unsupported intent, invalid budget flag, missing index rebuilt transparently, stable JSON envelope, `--render` output) plus new `internal/context` tests for the new `Render` function (well-formed Markdown, every item present, grouped by tier, no reordering); 011-016's own full suites re-run unmodified as the named regression gate. Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged.
**Project Type**: Single Go module (CLI). One new file in `internal/cli/internalcmd` (`context.go` + `context_test.go`), one new file in `internal/context` (`render.go` + `render_test.go`), one new exported sentinel + one new `Tier.String()` method on an already-existing file (`request.go`, `candidate.go` — see research.md #4/#5), one new line in `internal/cli/internal.go` registering the command, one new `classify` case in `internal/cli/internalcmd/errors.go`.
**Performance Goals**: Dominated by `Store.Sync`'s own already-established incremental cost (014) — unchanged artifacts are skipped by fingerprint comparison; this feature adds no additional full-corpus scan of its own. `Collect`/`Rank`/`ApplyBudget`/`Render` are the same bounded, in-memory costs 015/016 already established.
**Constraints**: Strictly read-only with respect to authoritative project artifacts (FR-008) — the only write this command ever performs is to the disposable cache database. Output MUST be one stable JSON envelope on both success and failure (FR-006, FR-012, Constitution Principle IX) — `--render` MUST NOT switch to raw Markdown-only stdout (research.md #7). Determinism required end-to-end for identical inputs against unchanged project content (FR-009).
**Scale/Scope**: `internal/cli/internalcmd/{context.go, context_test.go}` (new); `internal/context/{render.go, render_test.go}` (new); `internal/context/request.go` gains one exported sentinel (`ErrUnsupportedIntent`); `internal/context/candidate.go` gains one method (`Tier.String()`); `internal/cli/internalcmd/errors.go` gains one `classify` case; `internal/cli/internal.go` gains one registration line. No existing behavior of 011-016 is modified.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** The command performs zero semantic judgment — it validates input mechanically, opens/syncs the disposable index, and calls three already-deterministic functions (`Collect`, `Rank`, `ApplyBudget`) plus one new deterministic renderer (`Render`) in a fixed order. The coding agent remains responsible for deciding what to do with the returned context. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass.** The command mutates nothing but its own disposable cache database via `Store.Sync`/`index.Open` — both already-existing, already-deterministic 014 operations; no new mutation primitive is introduced. |
| III. Filesystem Is Single Source of Truth | **Pass, directly exercised.** This is the first feature to actually open the disposable index against a real project cache path — explicitly non-authoritative (research.md #1), safe to delete, and transparently rebuilt via `index.Open`'s existing schema-mismatch handling (research.md #2) — exactly the guarantee Phase 8 exists to demonstrate end-to-end. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised repeatedly.** No new subpackage — one file added to each of two already-existing packages (research.md #3); no new configuration field for the cache path, a fixed internal constant instead (research.md #1); "invalid budget" resolved as a Cobra-level int-parse failure rather than a new validation layer, and negative/zero budgets deliberately left to 016's own already-defined, already-tested semantics rather than re-litigated here (research.md #6); "Task" stays untyped free text exactly as 015 already defined it, with no new artifact-existence check invented for this feature (research.md #8); `--render` adds a field to the existing envelope rather than a second output mode or second command (research.md #7); index-sync failures fall through to the existing generic `unexpected_failure` classification rather than a new bespoke error code invented for a case with no current concrete failure mode to test against (research.md #9). |
| V. Test-First Discipline | **Gate carried into tasks.** CLI test list above is the explicit floor; 011-016's full suites are named regression gates. |
| VI. Clean Code & SOLID | **Pass, actively exercised.** `context.go`'s `RunE` is pure orchestration — every real decision (ranking, budgeting, index sync) stays inside the package that already owns it; `Tier.String()` centralizes the tier-label mapping once, reused by both `Render` (Markdown) and `context.go` (JSON `tier` field), rather than duplicating the same switch in two places (research.md #5). |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** This command writes nothing to any authoritative artifact — its only write target is the disposable cache database it also owns opening. |
| VIII. Safety by Construction | **Pass (N/A new surface).** The cache path is a fixed, computed join under the already-validated project root (`project.Detect`'s own existing root-confinement guarantee) — no new user-supplied path is ever accepted for it. |
| IX. Transparent, Machine-Readable Contracts | **Pass, directly exercised.** This is Phase 8's entire purpose: one stable JSON envelope (`ok`/`context` on success, `ok`/`error` on failure) via the existing `WriteSuccess`/`WriteError`/`classify` machinery, with `--render`'s Markdown pack nested as one more field rather than replacing the envelope (research.md #7). |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/017-internal-context-command/
├── plan.md                          # This file (/speckit-plan command output)
├── research.md                      # Phase 0 output (/speckit-plan command)
├── data-model.md                    # Phase 1 output (/speckit-plan command)
├── quickstart.md                    # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── context-command.md           # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md              # /speckit-specify quality checklist
└── tasks.md                         # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                          # unchanged — no new dependency
├── cmd/misterspec/                           # unchanged
├── internal/
│   ├── project/, ids/, testutil/, example/     # unchanged
│   ├── artifacts/, validation/, operations/      # unchanged — reused, not modified
│   ├── context/
│   │   ├── index/                                  # unchanged (014) — reused via Open/Sync, not modified
│   │   ├── request.go                                # MODIFIED — +ErrUnsupportedIntent sentinel, validateIntent wraps it
│   │   ├── request_test.go                             # MODIFIED — +coverage asserting errors.Is(err, ErrUnsupportedIntent)
│   │   ├── candidate.go                                  # MODIFIED — +Tier.String()
│   │   ├── candidate_test.go                               # MODIFIED — +Tier.String() coverage
│   │   ├── rank.go, budget.go, collector.go                 # unchanged (015/016) — reused, not modified
│   │   ├── render.go                                           # NEW — Render(Request, Result) string
│   │   └── render_test.go                                        # NEW
│   └── cli/
│       ├── internal.go                                             # MODIFIED — +NewContextCmd() registration
│       └── internalcmd/
│           ├── errors.go                                             # MODIFIED — +unsupported_intent classify case
│           ├── errors_test.go                                          # MODIFIED — +coverage for the new case
│           ├── references.go, backlinks.go, ...                         # unchanged — reused pattern, not modified
│           ├── context.go                                                  # NEW — NewContextCmd()
│           └── context_test.go                                               # NEW
└── kit/                                                                  # unchanged
```

**Structure Decision**: Two existing packages each gain exactly one new
file (`internal/context/render.go`, `internal/cli/internalcmd/
context.go`) plus the minimal surrounding wiring (one sentinel, one
method, one classify case, one registration line) — no new package, the
smallest footprint for a feature whose entire job is orchestrating
capabilities 011-016 already built, matching this project's own
established one-file-per-user-story-cluster discipline.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
