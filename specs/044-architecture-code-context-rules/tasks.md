---

description: "Task list for Regras de Arquitetura e Contexto de Código"
---

# Tasks: Regras de Arquitetura e Contexto de Código

**Input**: Design documents from `/specs/044-architecture-code-context-rules/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/check-architecture-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic parsing/evaluation/indexing logic, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US3)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. Two new packages: `internal/gosource` (shared Go-source parsing leaf) and `internal/architecture` (rule evaluation). Additive changes to `internal/context/index`, `internal/prepare`, `internal/project`, `internal/cli/internalcmd`.

---

## Phase 1: Setup

**Purpose**: Scaffold the two new packages this feature adds.

- [X] T001 [P] Create `internal/gosource/doc.go` with a package doc explaining `internal/gosource` is a pure, single-purpose leaf parsing one `.go` file's own imports and top-level declaration signatures via the standard library (`go/parser`) — no I/O beyond reading its own input file, no dependency on any other `internal/*` package, shared by both `internal/architecture` and `internal/context/index` (research.md #1/#2, plan.md "Scale/Scope").
- [X] T002 [P] Create `internal/architecture/doc.go` with a package doc explaining `internal/architecture` evaluates project-declared rules (forbidden dependency, layer boundary, required contract) against the current Go source tree via a per-language adapter, returning a strict `pass`/`fail`/`not_evaluated` per rule — never silently promoting an unsupported rule or language to `pass` (data-model.md "Result", spec FR-004/FR-005).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Go-source parsing, declared configuration (rules + exclusions), and the code index tables — every user story builds on all three.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundational

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T003 [P] Add `internal/gosource/imports_test.go` (write first): `Imports(path)` returns every import path and its own file-absolute line for a fixture `.go` file with multiple imports; returns an empty, non-nil slice for a file with no imports; returns an error for a syntactically invalid `.go` file (contracts §1). Expected to FAIL until T009.
- [X] T004 [P] Add `internal/gosource/declarations_test.go` (write first): `Declarations(path)` returns one `Declaration` per top-level `func`/`type`/`const`/`var`, each with `Name`, `Kind`, `Signature` (including its own doc comment when present), `Body` (the full declaration text), and file-absolute `StartLine`/`EndLine`; a file with only a package clause and comments returns an empty, non-nil slice (data-model.md "CodeDeclaration", contracts §1). Expected to FAIL until T010.
- [X] T005 [P] Add cases to `internal/project/config_test.go` (write first): `Load` accepts a `.misterspec/config.yaml` with a well-formed `architecture_rules` list (one `forbidden_dependency` entry with `from`/`to`) and a `code_exclusions` list, populating `Configuration.ArchitectureRules`/`Configuration.CodeExclusions`; both are empty/nil when the keys are absent (never an error — data-model.md "ArchitectureRule (config)"/"CodeExclusions (config)"); a rule entry with an unrecognized `kind` or an empty `from` returns an error wrapping `project.ErrInvalidConfiguration` naming the specific problem, the same convention every other malformed field already uses (contracts §7 "Error envelope"). Expected to FAIL until T011.
- [X] T006 [P] Add `internal/context/index/schema_test.go` cases (write first): after `ensureSchema` runs on a fresh database, `code_files` (columns `path` unique, `fingerprint`, `indexed_at`) and `code_declarations` (columns `code_file_id`, `name`, `kind`, `signature`, `body` nullable, `start_line`, `end_line`, `is_test`) tables exist; a database created under the previous `schemaVersion` is dropped and recreated with both present, mirroring the existing version-mismatch-triggers-rebuild pattern this file already has (data-model.md "CodeFile"/"CodeDeclaration", research.md #6). Expected to FAIL until T012.
- [X] T007 [P] Add `internal/context/index/sqlite_test.go` cases (write first): `SyncCode` indexes every `.go` file under root except paths matching `exclusions`, one `code_files` row plus one `code_declarations` row per top-level declaration; a second `SyncCode` call with no filesystem change leaves `code_files.fingerprint` unchanged for every file; a changed file's own declarations are replaced, not duplicated; a deleted file's own rows are removed. Expected to FAIL until T013.
- [X] T008 [P] Add cases to `internal/context/index/sqlite_test.go` (write first): `DeclarationsForFiles(paths)` returns every declaration whose own file is in `paths`, plus every declaration from that file's own associated `_test.go` file(s) in the same directory (`is_test: true`); a path with no indexed file returns no rows for it, no error (contracts §3, spec FR-007). Expected to FAIL until T013.

### Implementation for Foundational

- [X] T009 [P] Implement `internal/gosource/imports.go`: `ImportRef` and `Imports(path string) ([]ImportRef, error)`, via `go/parser.ParseFile(fset, path, nil, parser.ImportsOnly)` (contracts §1, research.md #2) (depends on T003).
- [X] T010 [P] Implement `internal/gosource/declarations.go`: `Declaration` and `Declarations(path string) ([]Declaration, error)`, via `go/parser.ParseFile(fset, path, nil, parser.ParseComments)`, rendering each top-level declaration's own signature text (with doc comment) and full body text via `go/format`/source-slicing against `fset`'s own recorded positions (contracts §1, research.md #2) (depends on T004).
- [X] T011 [P] Add `ArchitectureRule` type and `ArchitectureRules []ArchitectureRule`/`CodeExclusions []string` fields to `Configuration` in `internal/project/config.go`; add matching pointer-free fields to `configYAML` (`architecture_rules`/`code_exclusions` — `dec.KnownFields(true)` rejects unknown keys, so both must be declared here too); extend `validateConfig` to validate each declared rule's own `kind` against the fixed set (`forbidden_dependency`/`layer_boundary`/`required_contract`) and reject an empty `from`, returning `*ConfigError` (wrapping `project.ErrInvalidConfiguration`) on either problem — the same convention `stringOrDefault` already establishes (data-model.md, contracts §7, research.md #3/#4) (depends on T005).
- [X] T012 [P] Add `code_files`/`code_declarations` to `internal/context/index/schema.go`'s `schemaStatements`/`dropStatements`; bump `schemaVersion` (data-model.md "CodeFile"/"CodeDeclaration", research.md #6) (depends on T006).
- [X] T013 Add `CodeFile`/`CodeDeclaration` types to `internal/context/index/store.go`; add `SyncCode(root string, exclusions []string) (SyncReport, error)` and `DeclarationsForFiles(paths []string) ([]CodeDeclaration, error)` to the `Store` interface; implement both on `*sqliteStore` in `internal/context/index/sqlite.go`, `SyncCode` calling `internal/gosource.Declarations` per walked `.go` file and reusing the same new/changed/deleted reconciliation shape `Sync` already applies to Markdown (contracts §3, research.md #6) (depends on T009, T010, T012, T007, T008).

**Checkpoint**: Foundation ready — Go source can be parsed, rules/exclusions can be declared and loaded, and code can be indexed/queried; nothing yet evaluates a rule or renders code into a response.

---

## Phase 3: User Story 1 - Detectar uma dependência arquitetural proibida (Priority: P1) 🎯 MVP

**Goal**: A declared `forbidden_dependency`/`layer_boundary` rule violated by a real import produces a `fail` result with a reproducible file+line; a rule with no violation produces `pass`.

**Independent Test**: Declare a rule forbidding module A from importing module B, introduce an import of B inside A, run `check-architecture`, and confirm a `fail` result naming that exact file and line — repeatable across runs with no code change.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T014 [P] [US1] Add `internal/architecture/go_adapter_test.go` (write first): `CheckArchitecture` against a fixture Go module with a `forbidden_dependency` rule violated by one real import returns one `Result{Status: "fail"}` naming that import's own file and line; the same rule with no violating import returns `Result{Status: "pass"}`; running `CheckArchitecture` twice against unchanged fixture code returns byte-identical `Path`/`Line`/`Message` (data-model.md "Result", contracts §2, spec US1 Acceptance Scenarios 1-3). Expected to FAIL until T016.
- [X] T015 [P] [US1] Add cases to `internal/architecture/go_adapter_test.go` (write first): a `layer_boundary` rule (an "outer" pattern forbidden from importing an "inner" pattern's own reverse direction) is evaluated the same way a `forbidden_dependency` rule is — one `fail` per violating import, `pass` otherwise (data-model.md "ArchitectureRule"). Expected to FAIL until T016.

### Implementation for User Story 1

- [X] T016 [US1] Implement `internal/architecture/rule.go` (`Rule` type, `project.Configuration.ArchitectureRules` → `[]Rule` mapping), `internal/architecture/result.go` (`Result`/`Report` types), and `internal/architecture/go_adapter.go`'s `CheckArchitecture(root string, rules []Rule, exclusions []string) (Report, error)` for `forbidden_dependency`/`layer_boundary`: walks `.go` files under root (skipping `exclusions`), calls `gosource.Imports` per file, matches each import against every applicable rule's own `From`/`To` patterns, and emits one `Result` per rule (`pass` when no violation found, one `fail` `Result` per violating import otherwise) (contracts §2, research.md #2) (depends on T009, T011, T014, T015).
- [X] T017 [US1] Implement `internal/cli/internalcmd/check_architecture.go`: `misterspec internal check-architecture [--dir <path>]`, resolving `proj.Config.ArchitectureRules`/`CodeExclusions`, calling `architecture.CheckArchitecture`, and rendering `{"ok": true, "architecture": {"adapter", "results": [...]}}` per contracts §5; register it in `internal/cli/internal.go` (contracts §5) (depends on T016).

**Checkpoint**: User Story 1 is independently functional — a forbidden import or a layer-boundary violation is reported with a reproducible location via `misterspec internal check-architecture`.

---

## Phase 4: User Story 2 - Nunca aprovar falsamente uma linguagem ou regra sem suporte (Priority: P1)

**Goal**: A project with no `go.mod`, or a rule kind the Go adapter doesn't support (`required_contract`), always reports `not_evaluated` with a distinguishing `reason` — never `pass`.

**Independent Test**: Point `check-architecture` at a project with no `go.mod` and confirm every rule reports `not_evaluated`; declare a `required_contract` rule in a Go project and confirm that one rule alone reports `not_evaluated` while `forbidden_dependency`/`layer_boundary` rules in the same run still report `pass`/`fail` normally.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T018 [P] [US2] Add cases to `internal/architecture/go_adapter_test.go` (write first): `CheckArchitecture` against a project directory with no `go.mod` returns `Report{Adapter: ""}` with every rule in `rules` producing `Result{Status: "not_evaluated", Reason: "no_adapter_for_project"}`, never `"pass"` (data-model.md "Report", contracts §2, spec US2 Acceptance Scenario 1). Expected to FAIL until T019.
- [X] T019 [US2] Extend `internal/architecture/go_adapter.go`'s `CheckArchitecture`: detect `go.mod` presence at root first — absent, return `Report{Adapter: ""}` with every rule as `not_evaluated`/`"no_adapter_for_project"`, without attempting any file walk; present, set `Adapter: "go"` and proceed as T016 already does (research.md #8) (depends on T016, T018).
- [X] T020 [P] [US2] Add cases to `internal/architecture/go_adapter_test.go` (write first): a `required_contract` rule declared alongside a `forbidden_dependency` rule in the same `rules` list, run against a Go project, produces `Result{Status: "not_evaluated", Reason: "rule_kind_unsupported"}` for the `required_contract` entry and a normal `pass`/`fail` for the `forbidden_dependency` entry in the same `Report` (spec US2 Acceptance Scenario 2). Expected to FAIL until T021.
- [X] T021 [US2] Extend `internal/architecture/go_adapter.go`'s per-rule dispatch: a `Rule.Kind` of `"required_contract"` (not yet implemented by the Go adapter) always produces `Result{Status: "not_evaluated", Reason: "rule_kind_unsupported"}`, independent of and never blocking evaluation of any other rule in the same `rules` list (data-model.md "Result") (depends on T019, T020).

**Checkpoint**: User Story 2 is independently functional — an unsupported project or rule kind is always explicit and distinguishable (`reason`), never silently reported as compliant.

---

## Phase 5: User Story 3 - Recuperar código relevante para uma Tarefa sem ler o projeto inteiro (Priority: P2)

**Goal**: `internal prepare`'s response gains `code_context` (Scope-declared declarations, signature tier by default, full body under a fixed size threshold) and `code_scope_not_found` (any Scope path that didn't resolve).

**Independent Test**: Prepare a Task whose `Scope:` names a small file subset; confirm `code_context` includes those declarations (and their associated tests) without needing the whole project read, and that a nonexistent Scope path is named in `code_scope_not_found`.

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T022 [P] [US3] Add `internal/prepare/code_context_test.go` (write first): `ResolveCodeContext(store, scope, estimator)` for a `Scope:` naming one indexed `.go` file returns one `contextengine.PackageItem` per declaration in that file (via `DeclarationsForFiles`) plus every declaration from its own associated `_test.go` file, each `Content` set to the declaration's own `Signature` when the owning file's total estimated size exceeds `maxInlineCodeBodySize`, or its own full `Body` when at or under it; a `Scope:` path with no indexed match is returned in `notFound`, not silently omitted (data-model.md "TaskPreparation (extended)", contracts §4, spec FR-007/FR-008, Edge Case). Expected to FAIL until T023.

### Implementation for User Story 3

- [X] T023 [US3] Implement `internal/prepare/code_context.go`: `maxInlineCodeBodySize` constant and `ResolveCodeContext(store index.Store, scope string, estimator artifacts.Estimator) (items []contextengine.PackageItem, notFound []string, err error)` — parses `scope` into file paths (whitespace/comma-separated, mirroring `TaskFields.Scope`'s own existing free-text convention), calls `store.DeclarationsForFiles`, renders each `index.CodeDeclaration` as a `PackageItem` (`Heading` = declaration `Name`, `Content` = `Signature` or `Body` per the fixed threshold, `Location`/`Fingerprint` from the declaration's own file), and collects any Scope path matching no returned declaration into `notFound` (contracts §4, research.md #7) (depends on T013, T022).
- [X] T024 [US3] Add `CodeContext []contextengine.PackageItem` and `CodeScopeNotFound []string` to `TaskPreparation` in `internal/prepare/context.go`; extend `BuildTaskPreparation` (or its caller) to accept and set both (data-model.md "TaskPreparation (extended)") (depends on T023).
- [X] T025 [US3] Wire `internal/cli/internalcmd/prepare.go`: open/sync the code index (mirroring how `context.go` already opens `.misterspec/cache/context.db` via `index.Open`/`store.SyncCode`), call `prepare.ResolveCodeContext` with the target Task's own `Fields.Scope`, and pass the results into `BuildTaskPreparation`; extend `renderPreparation` to add `"code_context"`/`"code_scope_not_found"` keys, each a non-nil (possibly empty) array (contracts §6) (depends on T024).

**Checkpoint**: User Story 3 is independently functional — `internal prepare` returns Scope-relevant code (signatures, associated tests, and full bodies where the fixed threshold allows) without a separate call or manual file read; an unresolved Scope path is always named, never dropped.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification across all three stories together, and the dogfooding validation spec FR-006 requires.

- [X] T026 [P] Add a dogfooding integration test (e.g. `internal/architecture/dogfood_test.go` or a `internal/cli/internalcmd` case) running `check-architecture`/`CheckArchitecture` against MisterSpec's own repository root with a real declared rule reflecting an actual boundary this codebase already relies on (e.g. `internal/validation` MUST NOT import `internal/operations`, per Constitution Principle VI's own stated package boundaries) — confirms the Go adapter runs cleanly end-to-end on a real, non-trivial module (spec FR-006).
- [X] T027 [P] Run `gofmt -l` and `go vet ./...` across every new/changed file in this feature and fix any findings.
- [X] T028 Manually run all 6 scenarios in `quickstart.md` against a built `misterspec` binary in a real fixture Go project, confirming each "Expected" outcome.
- [X] T029 [P] Update `site/commands.html` with `internal check-architecture` (new command) and the additive `code_context`/`code_scope_not_found` fields on `internal prepare`, following the same documentation precedent 042/043's own polish passes established.
- [X] T030 [P] Final pass on every new exported symbol's doc comment in `internal/gosource`, `internal/architecture`, and the extended `internal/context/index`/`internal/prepare` files, confirming they describe the shipped behavior (three-way `pass`/`fail`/`not_evaluated`, fixed-threshold code sizing) rather than phase-by-phase build-out scope notes.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion. MVP — delivers rule violation detection.
- **User Story 2 (Phase 4)**: Depends on Foundational completion; its implementation tasks (T019, T021) extend the same `go_adapter.go` US1's T016 creates, so implement after US1.
- **User Story 3 (Phase 5)**: Depends on Foundational completion only — touches entirely different files (`internal/prepare`, not `internal/architecture`) from US1/US2, so it is genuinely independent of them and could be implemented in parallel by a different contributor.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other stories.
- **User Story 2 (P1)**: Independently testable once implemented, but its implementation tasks (T019, T021) touch the same file US1's T016 creates — sequence by file, not by required feature order.
- **User Story 3 (P2)**: No dependency on US1/US2 at all — different package (`internal/prepare` vs. `internal/architecture`), different underlying data (code index vs. import graph). Can be built in parallel with US1/US2 once Foundational is done.

### Within Each User Story

- Tests written and failing before implementation (Principle V).
- `internal/gosource`/`internal/architecture` types before the CLI command that exposes them (US1/US2); `internal/prepare`'s own resolution function before `TaskPreparation`'s own field before the CLI wiring that renders it (US3).

### Parallel Opportunities

- T001-T002 (Setup) can run in parallel.
- T003-T008 (all Foundational tests) can run in parallel — different files.
- T009-T012 (four of the five Foundational implementations) can run in parallel — different files, no shared dependency between them; T013 depends on T009/T010/T012/T007/T008.
- T014-T015 (US1 tests) can run in parallel.
- T018 and T020 (US2 tests) can run in parallel.
- **User Story 3 as a whole can run in parallel with User Stories 1+2** (different files, no shared dependency beyond Foundational).
- T026-T027, T029-T030 (Polish) can run in parallel.

---

## Parallel Example: Foundational

```bash
# Launch all Foundational tests together:
Task: "Add internal/gosource/imports_test.go"
Task: "Add internal/gosource/declarations_test.go"
Task: "Add architecture_rules/code_exclusions cases to internal/project/config_test.go"
Task: "Add code_files/code_declarations cases to internal/context/index/schema_test.go"
Task: "Add SyncCode cases to internal/context/index/sqlite_test.go"
Task: "Add DeclarationsForFiles cases to internal/context/index/sqlite_test.go"

# Once tests are in place, launch the independent implementation files together:
Task: "Implement internal/gosource/imports.go"
Task: "Implement internal/gosource/declarations.go"
Task: "Add ArchitectureRules/CodeExclusions to internal/project/config.go"
Task: "Add code_files/code_declarations to internal/context/index/schema.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run `quickstart.md` scenario 1 against a real fixture Go project.
5. `misterspec internal check-architecture` is now genuinely useful for the most common case (a real forbidden-dependency check) even before the `not_evaluated` guarantees (US2) or code-context retrieval (US3) exist.

### Incremental Delivery

1. Setup + Foundational → Go-source parsing, declared config, and the code index are ready.
2. Add User Story 1 → forbidden-dependency/layer-boundary detection works → validate independently (MVP!).
3. Add User Story 2 → `not_evaluated` guarantees made explicit and tested → validate independently (quickstart scenarios 2-3).
4. Add User Story 3 (parallelizable with 2/3 above) → `internal prepare` gains Scope-relevant code → validate independently (quickstart scenarios 4-6).
5. Polish → dogfooding test against MisterSpec's own repo, full quickstart re-run, `site/commands.html` update, doc-comment accuracy pass.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- Unlike prior multi-story Specs in this project (e.g. 042, 043), User Story 3 here is genuinely independent of User Stories 1/2 — different package, different data — so it is a real candidate for parallel work by a second contributor, not merely independently testable.
- Verify each story's tests fail before implementing that story's tasks.
- Commit after each task or logical group.
- Stop at any checkpoint to validate a story independently.
