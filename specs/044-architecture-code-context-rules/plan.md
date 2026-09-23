# Implementation Plan: Regras de Arquitetura e Contexto de Código

**Branch**: `044-architecture-code-context-rules` | **Date**: 2026-09-23 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/044-architecture-code-context-rules/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Today the SDD stops at text: a Spec's Requirements and a Task's own
`Scope:` field (034) declare intent, but nothing checks that the
actual Go code respects declared architectural constraints, and
nothing helps an agent retrieve just the code relevant to a Task
without reading files by hand. The technical approach: (1) one new
leaf package, `internal/gosource`, parses a `.go` file's own imports
and top-level declaration signatures using only the Go standard
library (`go/parser`) — the one new primitive both facets of this
feature share, avoiding a duplicated foundation if this Spec's own two
facets were split later (research.md #1/#2); (2) a new
`internal/architecture` package evaluates project-declared rules
(forbidden dependency, layer boundary, required contract — new,
additive `project.Configuration.ArchitectureRules`) against the
current Go source tree, returning a strict three-way `pass`/`fail`/
`not_evaluated` per rule — never silently promoting an unsupported
rule or an unsupported language to `pass` (research.md #5/#8); (3) a
new, additive `project.Configuration.CodeExclusions` field — the first
exclusion mechanism this project has, since research found none exists
today — is respected by both facets identically (research.md #4); (4)
code-context retrieval extends `internal prepare`'s own response
(never `internal context`'s `Target`, which was never designed to
accept a Task — implementation-time correction during planning,
research.md #6) with a new `code_context` field, resolving a Task's
own `Scope:` against two new sibling tables in the existing disposable
SQLite index (`code_files`/`code_declarations`), rendered as ordinary
`contextengine.PackageItem`s (033) so no new response shape has to be
learned; (5) signature-vs-full-body is a fixed, documented size
threshold, not a reused budget mechanism — research confirmed
`internal prepare` has no budget concept to reuse today (research.md
#7).

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: `go/parser`, `go/ast`, `go/token` (Go standard library — parsing, no type-checking); `golang.org/x/mod` (already a dependency) reused for resolving an import path against the project's own module path. `modernc.org/sqlite` v1.36.3 (existing disposable index — two new tables, no new virtual table). No new third-party dependency.
**Storage**: Filesystem (Constitution Principle III) for `.misterspec/config.yaml`'s new `architecture_rules`/`code_exclusions` sections (declared, not derived); the same disposable, reconstructable SQLite index `internal/context/index` already owns for `code_files`/`code_declarations` (`schemaVersion` bump, existing version-mismatch-triggers-rebuild path reused — no manual migration).
**Testing**: `go test` — unit tests for `gosource.Imports`/`Declarations` (fixture `.go` files, including one with no imports, one with only comments, one with a `_test.go` counterpart); `architecture.CheckArchitecture` (a forbidden-dependency violation, a clean pass, a project with no `go.mod`, an unsupported rule kind); `index.SyncCode`/`DeclarationsForFiles` (new/changed/deleted `.go` files, exclusion patterns, test-file association); `prepare.ResolveCodeContext` (signature-only vs. full-body threshold, a Scope path that doesn't resolve); a dogfooding integration test running `check-architecture` against MisterSpec's own repository with its own declared rules (spec FR-006).
**Target Platform**: cross-platform CLI binary (Linux/macOS/Windows) — `go/parser` and filesystem walking are already platform-independent; no new platform dependency.
**Project Type**: single Go project (CLI + internal libraries)
**Performance Goals**: no new performance target; `check-architecture` and `SyncCode` are both one-time, on-demand full-tree walks (same order of cost as `index.Store.Sync`'s own existing Markdown walk), not a hot path.
**Constraints**: a rule/project combination the adapter cannot evaluate MUST report `not_evaluated`, never `pass` (Principle I, spec FR-004); architecture rules MUST be declared, never inferred from observed code patterns (Principle I, spec Assumptions); code indexed for retrieval MUST remain disposable/reconstructable, carrying its own fingerprint (Principle III, spec FR-009); `CodeExclusions` MUST be honored identically by both facets (spec FR-010); code content MUST count against the same context budget/estimator already established (035) where a budget applies — `internal prepare` itself has none today, so its own code-context sizing uses a fixed, documented threshold instead (research.md #7); `internal context`'s own `Target` contract is unchanged — code retrieval is additive to `internal prepare` only (Principle VII).
**Scale/Scope**: new `internal/gosource` package (`Imports`, `Declarations`); new `internal/architecture` package (`Rule`, `Result`, `Report`, `CheckArchitecture`, the Go adapter); `internal/context/index` (`schemaVersion` bump, `code_files`/`code_declarations` tables, `SyncCode`, `DeclarationsForFiles` on `Store`); `internal/prepare` (`TaskPreparation.CodeContext`/`CodeScopeNotFound`, `ResolveCodeContext`); `internal/project` (`Configuration.ArchitectureRules`, `Configuration.CodeExclusions`, plus `Load`'s own validation of a declared rule's shape); `internal/cli/internalcmd` (new `check_architecture.go` command; `prepare.go` wires the new `code_context`/`code_scope_not_found` response fields — a malformed declared rule wraps the already-existing `project.ErrInvalidConfiguration`, no new error code). No new package beyond the two (`gosource`, `architecture`) both facets genuinely need.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle I (Semantic/Deterministic Separation)**: PASS. Whether an import exists, and whether a declared rule's own pattern matches it, are fully mechanical. *Declaring* a rule at all — deciding which dependencies should be forbidden — remains entirely the maintainer's own judgment, written into configuration; the binary never infers or proposes a rule from observed code (spec Assumptions).
- **Principle II (Deterministic Operations Are the Only Mutation Primitive)**: N/A. No new entity, ID, or canonical path is created; `check-architecture` and code-context retrieval are both read-only reports over already-existing source files.
- **Principle III (Filesystem Is the Single Source of Truth)**: PASS. Declared rules/exclusions live in the existing `.misterspec/config.yaml`; the code index is an explicitly disposable, reconstructable sibling of the existing Markdown index — losing it costs only re-indexing time, never correctness (mirrors 043's own `packs` table precedent).
- **Principle IV (Simplicity First — YAGNI)**: PASS with explicit boundaries (research.md #2, #3, #4, #7, #8): no new third-party static-analysis dependency: `go/parser` alone suffices for the two things needed (imports, signatures); no new artifact type for rules (project config already exists for exactly this); no reused-but-absent budget mechanism forced onto `internal prepare`; no `--language` flag where `go.mod` presence already answers the question unambiguously. Each was considered and deliberately rejected as more than this feature's own gap requires.
- **Principle V (Test-First Discipline)**: Applies — Go-source parsing, rule evaluation, code indexing/sync, and the signature/body threshold are exactly the deterministic logic this principle requires unit/integration tests for, written alongside implementation.
- **Principle VI (Clean Code & SOLID)**: PASS. `gosource` is a pure, single-purpose leaf (parsing only, no I/O beyond reading its own input file) reused by both `architecture` (imports) and `index` (declarations) rather than each reimplementing Go parsing independently (DRY) — the same reasoning this plan used to reject splitting into two Specs (research.md #1).
- **Principle VII (Explicit Mutation Boundaries)**: PASS. Neither new command writes to any project artifact; `internal prepare`'s own existing fields are unchanged, only additive (`code_context`/`code_scope_not_found`) — a caller ignoring the new fields sees byte-identical behavior to before this feature existed.
- **Principle VIII (Safety by Construction)**: PASS. Every file path this feature reads (for rule evaluation or code indexing) resolves through the same `artifacts.RelativeWithinRoot`-style root-confinement discipline `operations.Fingerprint`/`capture-evidence` already establish — a `Scope:`-declared path escaping the project root is rejected, never silently followed.
- **Principle IX (Transparent, Machine-Readable Contracts)**: Applies directly — `internal check-architecture` returns a structured JSON envelope with a stable `not_evaluated`/`reason` vocabulary (never inferred from an empty field); `internal prepare`'s new fields are additive and documented in the same contract style as every other command.

No violations requiring Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/044-architecture-code-context-rules/
├── plan.md                # This file (/speckit-plan command output)
├── research.md             # Phase 0 output (/speckit-plan command)
├── data-model.md           # Phase 1 output (/speckit-plan command)
├── quickstart.md           # Phase 1 output (/speckit-plan command)
├── contracts/              # Phase 1 output (/speckit-plan command)
│   └── check-architecture-contract.md
└── tasks.md                # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── gosource/                        # NEW leaf package
│   ├── doc.go
│   ├── imports.go                   # Imports(path) ([]ImportRef, error)
│   └── declarations.go              # Declarations(path) ([]Declaration, error)
├── architecture/                    # NEW package (Facet A)
│   ├── doc.go
│   ├── rule.go                      # Rule type, project.Configuration mapping
│   ├── result.go                    # Result/Report types
│   └── go_adapter.go                # CheckArchitecture, the Go adapter
├── context/
│   └── index/
│       ├── schema.go                # schemaVersion bump; + code_files/code_declarations
│       ├── store.go                 # Store interface gains SyncCode/DeclarationsForFiles
│       └── sqlite.go                # + implementations on *sqliteStore
├── prepare/
│   ├── context.go                   # TaskPreparation gains CodeContext/CodeScopeNotFound
│   └── code_context.go              # NEW: ResolveCodeContext
├── project/
│   └── config.go                    # + ArchitectureRules, CodeExclusions; Load validates rule shape
└── cli/internalcmd/
    ├── check_architecture.go        # NEW: `misterspec internal check-architecture`
    ├── prepare.go                   # + code_context/code_scope_not_found rendering
    # (no errors.go change — a malformed rule wraps project.ErrInvalidConfiguration)

tests/ (co-located _test.go files per existing repo convention)
├── internal/gosource/*_test.go
├── internal/architecture/*_test.go
├── internal/context/index/sqlite_test.go        # + SyncCode/DeclarationsForFiles cases
├── internal/prepare/code_context_test.go
└── internal/cli/internalcmd/check_architecture_test.go, prepare_test.go (+ cases)
```

**Structure Decision**: Single Go project. Two new packages
(`gosource`, `architecture`) — both genuinely new responsibilities no
existing package owns, and both small/leaf-shaped (Constitution
Principle IV/VI). Every other change is additive to a package that
already owns the corresponding responsibility (`internal/context/index`
for the disposable store, `internal/prepare` for Task-shaped
responses, `internal/project` for declared configuration,
`internal/cli/internalcmd` for CLI wiring) — no speculative new layer
beyond the two packages this feature's own two facets both need.

## Complexity Tracking

*No violations — table omitted.*
