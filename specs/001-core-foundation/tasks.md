---

description: "Task list template for feature implementation"
---

# Tasks: Core Repository Foundation

**Input**: Design documents from `/specs/001-core-foundation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/packages.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE) and the plan's Constitution Check gate — this feature does not waive them.

**Organization**: Tasks are grouped by user story (from spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

Single Go module, library-first (no `cmd/` yet — see plan.md's Structure Decision):

```text
internal/project/     # US1 — Configuration, Detect
internal/artifacts/   # US2 (types, paths) + US3 (metadata, parser)
internal/ids/         # Foundational (types) + US3 (parse, scan)
internal/testutil/    # Foundational — shared fixture helper for all stories' integration tests
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Repository/module initialization — nothing exists yet, this is the first code in the repo.

- [X] T001 Initialize the Go module (`go mod init github.com/mottamarcio/misterspec`, `go` directive `1.23`) at the repository root, and create the `internal/project/`, `internal/artifacts/`, `internal/ids/`, and `internal/testutil/` directories, each with a package-level doc comment stating its responsibility per plan.md's Project Structure.
- [X] T002 [P] Add the `gopkg.in/yaml.v3` dependency (`go get gopkg.in/yaml.v3`, committing the resulting `go.mod`/`go.sum`) per research.md's "YAML frontmatter parsing library" decision.

**Checkpoint**: Module builds (`go build ./...` succeeds on empty packages) before any Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared types every user story's tests and implementation import. Per data-model.md, `EntityType`/`EntityID` and `Configuration` are used by more than one story, so their *type definitions* (not yet their behavior) live here — see plan.md's Constitution Check (Principle IV: no speculative behavior added ahead of the story that owns it).

**⚠️ CRITICAL**: No user story implementation may begin until this phase is complete.

- [X] T003 [P] Define the `Configuration` struct and its documented field defaults (`ArtifactsDir`, `RawDir`, `KnowledgeDir`, `ConstitutionPath`, `LearningsDir`, `ProgramsRoot`, `IDWidth`, `SchemaVersion`, `AgentID`) per data-model.md's Configuration entity, in `internal/project/config.go`. Type only — loading and validation are US1's `T009`.
- [X] T004 [P] Define the `EntityType` enum (`program`, `feature`, `spec`, `task`, `knowledge`, `learning`) and the `EntityID` struct (`Type`, `Prefix`, `Number`, `Width`) per data-model.md's Entity ID entity, in `internal/ids/types.go`. Data-only — `Parse`, `NextID`, and `Scan` are US3's `T018`–`T020`.
- [X] T005 [P] Define the `ArtifactType` enum (`program`, `feature`, `spec`, `plan`, `tasks`, `validation`, `knowledge`, `learning`, `constitution` — the superset described in data-model.md's Artifact entity) in `internal/artifacts/types.go`. Enum only — `ClassifyPath` is US2's `T014`.
- [X] T006 Implement a shared fixture-building test helper in `internal/testutil/fixture.go`: builds a temporary project tree (`t.TempDir()`-based) with a `.misterspec/config.yaml` and arbitrary artifact files (valid, malformed-frontmatter, duplicate-ID, as needed by callers), for reuse by every user story's filesystem-integration tests (Constitution Principle V).

**Checkpoint**: Foundation ready — shared types compile, fixture helper is available; US1, US2, US3 implementation can now begin (US2/US3 still depend on US1's `Detect`/`Load` only insofar as their own tests choose to exercise a fully-detected `Project`, but per contracts/packages.md every US2/US3 function takes `Configuration`/`EntityID` values directly, so they do not block on US1 completing first).

---

## Phase 3: User Story 1 - Reliable Project Detection & Configuration (Priority: P1) 🎯 MVP

**Goal**: Any consumer can be pointed at the project root or any nested subdirectory and reliably learn the project root and resolved configuration, or get a distinct not-initialized / invalid-configuration result.

**Independent Test**: Point `Detect` at fixture directories (valid root, nested subdirectory, uninitialized directory, malformed config) and confirm the correct root/config or distinct error for each — no other story's code required.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T009–T010.

- [X] T007 [P] [US1] Unit tests for configuration validation — missing `SchemaVersion`, empty required directory field, non-positive `IDWidth` — each asserting a distinct, named-field `ErrInvalidConfiguration`, in `internal/project/config_test.go`.
- [X] T008 [P] [US1] Filesystem-integration tests for `Detect()` — valid project root, a directory nested several levels below one, a directory with no project in its ancestry, and a project with a malformed `config.yaml` — using `internal/testutil` fixtures, in `internal/project/root_test.go`.

### Implementation for User Story 1

- [X] T009 [US1] Implement configuration loading and validation (`Load`, returning a populated `Configuration` or `ErrInvalidConfiguration` naming the offending field — never silently defaulting a required field) in `internal/project/config.go`. Depends on T003, T007.
- [X] T010 [US1] Implement `Detect(startDir string) (*Project, error)` — upward directory walk to the first `.misterspec/config.yaml`, returning an identical `Project{Root, Config}` regardless of starting subdirectory depth, or `ErrNotInitialized` — in `internal/project/root.go`. Depends on T009, T008.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/project/...` passes on its own.

---

## Phase 4: User Story 2 - Canonical Path Resolution & Artifact Typing (Priority: P2)

**Goal**: Given an entity type/ID/parent, compute the one canonical path for it; given a path, classify its artifact type; never resolve outside the project root.

**Independent Test**: Feed a fixed project root and a table of (type, ID, parent) inputs and confirm canonical paths match the frozen layout exactly, plus feed traversal inputs and confirm rejection — independent of US1 or US3 being implemented (this story consumes `Configuration`/`EntityID` values directly, not `Detect`'s output).

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T013–T014.

- [X] T011 [P] [US2] Unit tests for `ResolvePath`: a table covering every supported entity type's canonical directory/file per `docs/architecture-specification.md` §20 (program, feature, spec, knowledge, learning), plus cases where a supplied ID or parent would resolve outside the project root, asserting `ErrPathOutsideProject`, in `internal/artifacts/paths_test.go`.
- [X] T012 [P] [US2] Unit tests for `ClassifyPath`: classification by canonical location, and the case where a declared frontmatter `type` disagrees with location, in `internal/artifacts/types_test.go`.

### Implementation for User Story 2

- [X] T013 [US2] Implement `ResolvePath(root string, cfg project.Configuration, t ids.EntityType, id ids.EntityID, parent *ids.EntityID) (CanonicalPath, error)` as a pure function of its inputs, checking containment inside `root` *before* constructing any path (never after), in `internal/artifacts/paths.go`. Depends on T004, T011.
- [X] T014 [US2] Implement `ClassifyPath(root, path string) (ArtifactType, error)` in `internal/artifacts/types.go`. Depends on T005, T012.

**Checkpoint**: User Stories 1 AND 2 both independently pass their own tests.

---

## Phase 5: User Story 3 - Metadata Parsing & ID Discovery (Priority: P3)

**Goal**: Parse an artifact's frontmatter into structured metadata with distinct error conditions; scan the project tree for existing IDs of a type, flagging duplicates; compute the next ID by scanning, never a stored counter.

**Independent Test**: Against a fixture tree with well-formed artifacts, one with missing/invalid frontmatter, and two sharing a duplicate ID — confirm metadata parses where valid, errors are specific where invalid, and the duplicate is flagged without failing the scan.

### Tests for User Story 3

> Write these tests FIRST; confirm they fail before implementing T018–T021.

- [X] T015 [P] [US3] Unit tests for `ids.Parse` (wrong prefix, non-numeric suffix, wrong zero-padding width → `ErrInvalidIDSyntax`; valid input → populated `EntityID`) and `ids.NextID` (empty slice → number 1; a slice with gaps → max+1, gaps not backfilled) in `internal/ids/ids_test.go`.
- [X] T016 [P] [US3] Filesystem-integration tests for `ids.Scan` — several valid IDs of one type, a duplicate ID, and a type with zero existing artifacts (must return an empty result, not an error) — using `internal/testutil` fixtures, in `internal/ids/scan_test.go`.
- [X] T017 [P] [US3] Filesystem-integration tests for `artifacts.ParseMetadata` — well-formed frontmatter, a missing file (`ErrArtifactNotFound`), malformed YAML (`ErrFrontmatterMalformed`), and valid YAML missing a required field (`ErrRequiredFieldMissing`) — in `internal/artifacts/parser_test.go`.

### Implementation for User Story 3

- [X] T018 [US3] Implement `ids.Parse(t EntityType, raw string, width int) (EntityID, error)` and `ErrInvalidIDSyntax` in `internal/ids/ids.go`. Depends on T004, T015.
- [X] T019 [US3] Implement `ids.NextID(existing []EntityID, t EntityType, width int) EntityID` as a pure function reading only its `existing` argument — no persisted counter is read or written — in `internal/ids/ids.go`. Depends on T018.
- [X] T020 [US3] Implement `ids.Scan(root string, cfg project.Configuration, t EntityType) (ScanResult, error)` — tree walk, duplicate collection into `ScanResult.Duplicates` without halting enumeration, empty result (not error) when nothing is found — in `internal/ids/scan.go`. Depends on T019, T016.
- [X] T021 [US3] Implement the `Metadata` struct and `ParseMetadata(path string) (Metadata, error)` — frontmatter extraction, strict `yaml.v3` unmarshal, and the three distinct errors (`ErrArtifactNotFound`, `ErrFrontmatterMalformed`, `ErrRequiredFieldMissing`) — in `internal/artifacts/metadata.go` and `internal/artifacts/parser.go`. Depends on T004, T017.

**Checkpoint**: All three user stories independently pass their own tests — the Phase 1 foundation described in plan.md is complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T022 [P] Run `gofmt -l` and `go vet ./...` across `internal/project`, `internal/artifacts`, `internal/ids`, `internal/testutil`, and fix any findings.
- [X] T023 [P] Add/verify package-level doc comments on `project`, `artifacts`, and `ids` summarizing each package's exported contract, cross-checked against `specs/001-core-foundation/contracts/packages.md`.
- [X] T024 Add a compiled, run-in-CI example (`internal/example` or a `google.golang.org`-style `Example...` test) exercising `quickstart.md`'s full flow end-to-end: `Detect` → `ResolvePath` → `ParseMetadata` → `Scan` → `NextID`.
- [X] T025 Reconcile `specs/001-core-foundation/contracts/packages.md`'s error-code mapping table against the actual implemented sentinel error names; fix any drift introduced during implementation.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Stories (Phase 3–5)**: All depend on Foundational completion. US1, US2, US3 do **not** depend on each other's implementation (each contract function takes plain `Configuration`/`EntityID` values, not another story's output) — they can proceed in parallel or in priority order (P1 → P2 → P3).
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Can start after Foundational. No dependency on US1/US3 (tests build `Configuration`/`EntityID` values directly rather than calling `Detect`/`Parse`).
- **User Story 3 (P3)**: Can start after Foundational. No dependency on US1; uses the `EntityID`/`EntityType` types from Foundational, not from US2's behavior.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational types before story-specific behavior.
- Story complete and its own tests green before moving to the next priority (if working sequentially).

### Parallel Opportunities

- T002 (Setup) has no dependency on T001 completing beyond the module existing — can follow immediately.
- All Foundational tasks marked [P] (T003, T004, T005) touch different files and can run in parallel; T006 depends on none of them completing first but is written to use whichever types exist, so sequence it last in Phase 2 for simplicity.
- Once Foundational completes, US1, US2, and US3 can proceed in parallel by different contributors.
- Within each story, all [P] test tasks can run in parallel (different files); implementation tasks within a story are sequential where noted (same file or explicit dependency).

---

## Parallel Example: Foundational Phase

```bash
# Launch all Foundational type-definition tasks together (different files):
Task: "Define Configuration struct in internal/project/config.go"
Task: "Define EntityType/EntityID in internal/ids/types.go"
Task: "Define ArtifactType enum in internal/artifacts/types.go"
```

## Parallel Example: User Story 3

```bash
# Launch all US3 test tasks together (different files):
Task: "Unit tests for ids.Parse/NextID in internal/ids/ids_test.go"
Task: "Filesystem-integration tests for ids.Scan in internal/ids/scan_test.go"
Task: "Filesystem-integration tests for artifacts.ParseMetadata in internal/artifacts/parser_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (blocks everything).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/project/...` green, independently.
5. This alone already lets any later feature reliably answer "am I in a misterspec project, and what's its config?" — a usable increment even before US2/US3 exist.

### Incremental Delivery

1. Setup + Foundational → shared types and fixture helper ready.
2. Add US1 → validate independently → project detection usable (MVP).
3. Add US2 → validate independently → canonical paths and artifact typing usable.
4. Add US3 → validate independently → metadata parsing and ID discovery usable — Phase 1 of the architecture spec is now fully implemented.
5. Polish (Phase 6) once all three are in.

### Parallel Team Strategy

With more than one contributor:

1. Complete Setup + Foundational together first (it blocks everyone).
2. Once Foundational is done: Developer A takes US1, Developer B takes US2, Developer C takes US3 — no cross-story blocking since each story's contract functions take plain values, not another story's runtime output.
3. Stories integrate at Phase 6 (Polish), and later at the Phase 2 `operations` layer this foundation was built for.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- No task in this feature writes to the filesystem outside of test fixtures under `t.TempDir()` — this feature is read-only by design (see plan.md's Constitution Check, Principle II).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
