# Implementation Plan: Reutilização Incremental de Context Packs

**Branch**: `043-incremental-context-reuse` | **Date**: 2026-09-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/043-incremental-context-reuse/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today `internal context --mode package` (033) already returns every
selected item's own full content, location, and fingerprint — but a
caller returning to the same target (a second Task in the same Spec,
or resuming after losing context) always gets the whole package again,
even when almost nothing changed. The technical approach: (1) a new
`--base <pack_id>` flag on the existing `internal context` command,
never a second command, since the full selection/ranking/budgeting
pipeline it needs already exists (research.md #1); (2) `pack_id` is a
SHA-256 hash (reusing `contextengine.Fingerprint`'s existing
convention) over the request's own resolved configuration plus every
selected item's own `(Path, Location, Fingerprint)` — so it changes
automatically the instant either the request or the underlying content
differs (research.md #2); (3) one new `packs` table added to the
already-open, already-disposable SQLite index
(`internal/context/index`), reusing its existing version-mismatch-
triggers-full-rebuild lifecycle rather than inventing a second store
(research.md #3); (4) `--base` is a lookup key only, never a trusted
assertion — an unknown or invalidated `pack_id` always falls back to
the full package, explicitly marked `recovered: true` (research.md
#4); (5) the diff itself is an ordered, identity-keyed
reuse-or-replace list (reusing `Candidate`'s own `(Path, StartLine,
EndLine)` identity), directly reversible by the caller to reconstruct
the exact full pack (research.md #5); (6) a `reuse` diagnostics block
reports per-call reuse counts — no new aggregation command (research.md
#6); (7) documentation states plainly that this local mechanism is
independent of any LLM provider's own prompt caching (research.md #7).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `modernc.org/sqlite` v1.36.3 (existing disposable index — `packs` is a new table in the same database, no new virtual table); `github.com/spf13/cobra` v1.10.2 (CLI). No new third-party dependency.
**Storage**: SQLite, the same disposable/reconstructable index at `internal/context/index` (Constitution Principle III) `internal context` already opens. `schemaVersion` bumps 3→4, triggering the index's own existing automatic rebuild-on-mismatch path — no manual migration code (research.md #3). A stored pack is explicitly disposable: its loss never breaks correctness, only that one call's reuse opportunity (spec FR-011).
**Testing**: `go test` — unit tests for `ComputePackID` (identical inputs → identical ID; any single differing component → a different ID), `DiffAgainstBase` (pure function: no-op reuse, single-item modification, addition, removal, and reconstructing the exact full pack from a diff); filesystem/SQLite integration tests for `Store.SavePack`/`LookupPack` (round-trip, eviction bound, schema-version-triggered loss); a golden-style CLI envelope test for `internal context --base`'s three response shapes (diff, recovery-unknown, recovery-invalidated).
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows), same as every other `internal` command — no new platform dependency.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: no new performance target; `--base` adds one indexed lookup (`packs.pack_id`, the table's own primary key) and one `SavePack` upsert per `--mode package` call — negligible relative to the existing `Collect`/`Rank`/`ApplyBudget` pipeline the command already runs on every call.
**Constraints**: `--base` MUST only be accepted alongside `--mode package` (research.md #1); a `--base` value MUST NEVER be trusted without a matching stored row re-verified against the current request's own freshly recomputed identity (Principle I/VIII, spec FR-004); the `packs` table MUST remain bounded (research.md #3) — no unbounded local growth across a long-lived project; this feature MUST NOT claim or measure LLM-provider-side cost savings, only locally measured reuse counts (spec FR-010, research.md #7); `internal context`'s existing `--base`-less behavior MUST remain byte-identical except for the additive `pack_id` field (Constitution Principle VII).
**Scale/Scope**: `internal/context` (new `pack.go` additions: `ConfigHash`/`PackID`/`ConfigIdentity`/`PackIdentity`/`PackItemIdentity`, `ComputeConfigHash`, `ComputePackID`, `DiffEntry`/`RemovedEntry`/`PackDiff`/`RecoveryResult`/`ReuseDiagnostics`, `DiffAgainstBase`); `internal/context/index` (`schemaVersion` 3→4, new `packs` table with a `config_hash` column, `StoredPack`/`StoredPackItem` types, `SavePack`/`LookupPack` on `Store`); `internal/cli/internalcmd/context.go` (`--base` flag, `contextSchemaVersion` 5→6, new response rendering for `pack_id`/`diff`/`recovered`/`reuse`). No new package layer.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Identity hashing, diff computation, and storage are all fully mechanical. Whether/when an agent chooses to pass `--base` at all — and what to do with a `recovered: true` response — stays the agent's own judgment; the binary never decides on the agent's behalf that a base "should" still be valid.
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. No new entity, ID, or canonical path is created; `pack_id` identifies a *response*, not a project artifact.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. The `packs` table is explicitly disposable — reconstructable behavior (full recovery) is always available with zero data loss risk to the project itself; losing every stored pack only ever costs one call's worth of reuse, never correctness (research.md #3, spec FR-011).
- **Principle IV (Simplicity First — YAGNI)**: PASS with explicit boundaries (research.md #1, #3, #6): no new command, no new storage file/lifecycle, no new aggregation/reporting command — each was considered and deliberately rejected as more than this feature's own gap requires.
- **Principle V (Test-First Discipline)**: Applies — `ComputePackID`'s sensitivity to every input dimension, `DiffAgainstBase`'s reconstruction guarantee, and `SavePack`/`LookupPack`'s round-trip/eviction behavior are exactly the deterministic logic this principle requires unit/integration tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. Identity/diff computation stays in `internal/context` (which already owns `PackageItem`/`Fingerprint`/the selection pipeline); persistence stays in `internal/context/index` (which already owns the one disposable store `internal context` opens) — neither is duplicated into `internal/cli/internalcmd`, which only wires flags and shapes JSON, exactly like every other `internal` command.
- **Principle VII (Explicit Mutation Boundaries)**: PASS. `internal context --base` only ever writes to its own new `packs` table (via `SavePack`), inside the same disposable index it already opens — it never touches any project artifact (`spec.md`, `tasks.md`, etc.). A `--base`-less call's existing output is unaffected beyond the additive `pack_id` field.
- **Principle VIII (Safety by Construction)**: PASS. `--base` is never trusted as an assertion (research.md #4) — the one safety property this whole feature exists to guarantee structurally, not by convention. No new filesystem path resolution is introduced (the `packs` table lives inside the existing `.misterspec/cache/context.db`).
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — `contextSchemaVersion` bumps 5→6 (additive: `pack_id` always present in `--mode package`; `diff`/`recovered`+`reason` only when `--base` is passed); `recovered`/`reason` are stable, explicit fields, never inferred from response shape/size.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/043-incremental-context-reuse/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md          # Phase 1 output (/speckit-plan command)
├── quickstart.md          # Phase 1 output (/speckit-plan command)
├── contracts/             # Phase 1 output (/speckit-plan command)
│   └── incremental-context-reuse-contract.md
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── context/
│   ├── pack.go                      # + ConfigHash/PackID/ConfigIdentity/PackIdentity/
│   │                                 #   PackItemIdentity, ComputeConfigHash, ComputePackID,
│   │                                 #   DiffEntry/RemovedEntry/PackDiff/RecoveryResult/
│   │                                 #   ReuseDiagnostics, DiffAgainstBase
│   └── index/
│       ├── schema.go                # schemaVersion 3->4; + packs table
│       ├── store.go                 # + StoredPack/StoredPackItem types; Store interface
│       │                             #   gains SavePack/LookupPack
│       └── sqlite.go                # + SavePack/LookupPack implementations on *sqliteStore
└── cli/internalcmd/
    └── context.go                   # + --base flag; contextSchemaVersion 5->6;
                                      #   pack_id/diff/recovered/reuse response rendering

tests/ (co-located _test.go files per existing repo convention)
├── internal/context/pack_test.go              # + ComputePackID/DiffAgainstBase cases
├── internal/context/index/sqlite_test.go       # + SavePack/LookupPack round-trip, eviction
├── internal/context/index/schema_test.go      # + packs table present after ensureSchema
└── internal/cli/internalcmd/context_test.go   # + --base golden-envelope cases
```

**Structure Decision**: Single Go project (existing `internal/...`
layout). Every change is additive to a package that already owns the
corresponding responsibility (`internal/context` for pack
identity/diffing, `internal/context/index` for the one disposable
store, `internal/cli/internalcmd` for CLI wiring) — no new package,
no new file beyond what each existing file's own responsibility
already covers (Constitution Principle IV/VI).

## Complexity Tracking

*No violations — table omitted.*
