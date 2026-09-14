---

description: "Task list template for feature implementation"
---

# Tasks: References and Backlinks

**Input**: Design documents from `/specs/012-references-backlinks/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/references-backlinks.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 011-wikilink-foundation's and 008-cli-cobra's full test suites are explicit, named regression gates (User Story 3, Polish), since this feature must not change their behavior at all.

**Organization**: Tasks are grouped by user story. **Foundational blocks both User Story 1 and User Story 2** — both need `ids.ResolveTarget` to exist. **Unlike 011's strictly sequential US1→US2→US3 chain, User Story 1 (References) and User Story 2 (Backlinks) here are independent of each other** — neither's Go code calls the other's; both only depend on Foundational — matching 003's/006's/009's own parallel-story-pair shape, with one exception: their two CLI-wiring tasks (T009, T014) touch the same `internal/cli/internal.go` file, so those two specific tasks are sequential regardless of story order. **User Story 3 depends on both User Story 1 and 2 existing** — there is nothing to prove "unaffected" until both new capabilities exist to be unaffected by.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/ids/                    # existing package — Foundational
internal/validation/               # existing package — Foundational (refactor only)
internal/operations/                  # existing package — US1, US2
internal/cli/, internal/cli/internalcmd/   # existing packages — US1, US2
internal/example/                            # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package and no new dependency (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the one genuinely shared piece of infrastructure both later stories need — `ids.ResolveTarget`, sharing the width-tolerant target-resolution logic `internal/validation`'s `classifyWikilink` already has inline, rather than a third copy of it (research.md #2). This is the only place existing, already-shipped code (011's `classifyWikilink`) changes at all in this feature.

**⚠️ CRITICAL**: No user story implementation may begin until this phase's regression gate (T004) passes.

- [X] T001 [P] Unit tests for `ids.ResolveTarget` in `internal/ids/scan_test.go`: a well-formed target resolving to exactly one artifact; a well-formed target resolving to nothing (broken); a well-formed target resolving to more than one artifact claiming the same ID (ambiguous — a duplicate ID fixture); a malformed target string (not even a well-formed entity ID) returns an error; a target using a different zero-padding width than the project's configured width still resolves correctly (mirrors 011's own `TestCheckWikilinks_DifferentZeroPaddingWidthStillResolves`).
- [X] T002 Implement `ResolveTarget(root string, cfg project.Configuration, raw string) (EntityID, []string, error)` (data-model.md) in `internal/ids/scan.go`, composing `ParseAny` and `Scan` exactly as `internal/validation/wikilinks.go`'s existing `classifyWikilink` already does inline. Depends on T001.
- [X] T003 Refactor `classifyWikilink` in `internal/validation/wikilinks.go` to call `ids.ResolveTarget` instead of its own inline `ids.ParseAny`/`ids.Scan` calls — behavior-identical, zero change to any Finding code or message. Depends on T002.
- [X] T004 Regression checkpoint: run `go test ./internal/ids/...` and `go test ./internal/validation/...` and confirm every existing test (011-wikilink-foundation's own full suite included) passes unmodified — the explicit, named proof this refactor changed nothing observable about `classifyWikilink`. Depends on T003.

**Checkpoint**: Foundation ready — `ids.ResolveTarget` exists, shared and proven behavior-identical to 011's own resolution logic. User story implementation can begin.

---

## Phase 3: User Story 1 - Discover Everything an Artifact Points To (Priority: P1) 🎯 MVP

**Goal**: `operations.References` reports every outgoing formal relationship (parent, depends_on, supersedes, exactly as declared) and every outgoing semantic relationship (a wikilink resolving to exactly one real artifact) for a given artifact of one of the five scoped types.

**Independent Test**: Create an artifact with a declared parent, a `depends_on` entry, and a body containing a wikilink to a third artifact; query its references and confirm the parent and dependency appear labeled formal, and the wikilink appears labeled semantic, each naming its exact target.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T006.

- [X] T005 [P] [US1] Unit tests for `operations.References` in `internal/operations/references_test.go`, covering `docs/context-engine-implementation.md` §29.3's own list plus spec.md's acceptance scenarios: formal relationships only (parent, `depends_on`, `supersedes`, in declaration order); semantic relationships only; both combined and correctly separated; an artifact with no relationships at all returns a well-formed empty result, not an error; a self-referencing formal field and a self-referencing wikilink are both reported, not dropped; two separate wikilinks to the same target are both preserved, not collapsed; a broken, malformed, or ambiguous wikilink target is excluded (data-model.md's resolution table); a target of an out-of-scope type (Task) is rejected via `operations.ErrInvalidTarget`.

### Implementation for User Story 1

- [X] T006 [US1] Implement `ReferenceEntry`, `ReferencesResult`, and `References` in new file `internal/operations/references.go` (data-model.md, contracts/references-backlinks.md) — calls `Inspect` (reusing its existing `Resolve`+`ParseMetadata` composition), rejects a non-scoped type via `ErrInvalidTarget`, builds `Formal` directly from `Metadata`, builds `Semantic` via `ReadBody`+`ExtractWikiLinks`+`ids.ResolveTarget` per link, keeping only exactly-one-match targets. Depends on T005, T004.

### CLI for User Story 1

- [X] T007 [P] [US1] Tests for `misterspec internal references <id>` in `internal/cli/internalcmd/references_test.go`: success JSON shape (`{"ok":true,"target":...,"references":{"formal":[...],"semantic":[...]}}`); `entity_not_found`/`entity_ambiguous`/`invalid_target` failure shapes and exit codes, reusing `operations`'s existing sentinels with zero `classify` change.
- [X] T008 [US1] Implement `NewReferencesCmd` in new file `internal/cli/internalcmd/references.go` — argument parsing and JSON shaping only, calling `operations.References` directly (contracts/references-backlinks.md). Depends on T007, T006.
- [X] T009 [US1] Wire `internalcmd.NewReferencesCmd()` into `newInternalCmd`'s `AddCommand` list in `internal/cli/internal.go`. Depends on T008.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/operations/... -run TestReferences` and `go test ./internal/cli/... -run TestReferences` both pass on their own, with zero dependency on Backlinks.

---

## Phase 4: User Story 2 - Discover Everything That Points To an Artifact (Priority: P2)

**Goal**: `operations.Backlinks` reports every other artifact in the project that formally or semantically references a given artifact of one of the five scoped types, discovered by a full project scan.

**Independent Test**: Create three artifacts where two of them reference a third — one formally (`depends_on`), one semantically (a wikilink); query the third artifact's backlinks and confirm both referencing artifacts appear, each correctly labeled.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T011.

- [X] T010 [P] [US2] Unit tests for `operations.Backlinks` in `internal/operations/backlinks_test.go`, mirroring T005's own list inverted: formal backlinks only; semantic backlinks only; both combined and correctly separated; an artifact nothing references returns a well-formed empty result, not an error; a self-referencing artifact appears as its own backlink; two separate artifacts (or two separate wikilinks from the same artifact) referencing the same popular target both appear, in deterministic order; a broken, malformed, or ambiguous wikilink is excluded in both directions (data-model.md); a target of an out-of-scope type (Task) is rejected via `operations.ErrInvalidTarget`.

### Implementation for User Story 2

- [X] T011 [US2] Implement `BacklinkEntry`, `BacklinksResult`, and `Backlinks` in new file `internal/operations/backlinks.go` (data-model.md, contracts/references-backlinks.md) — calls `Inspect` to resolve and scope-check the queried target, then scans every artifact of the five scoped types (`ids.Scan` per type, `resolve.go`'s existing `canonicalFilename` for the file path, `ParseMetadata`+`ReadBody`+`ExtractWikiLinks`+`ids.ResolveTarget` per candidate source), in the same fixed type-then-number deterministic order `internal/validation/validator.go`'s own `projectEntityTypes`/`sortedNumbers` already establish (research.md #5). Depends on T010, T004.

### CLI for User Story 2

- [X] T012 [P] [US2] Tests for `misterspec internal backlinks <id>` in `internal/cli/internalcmd/backlinks_test.go`: success JSON shape (`{"ok":true,"target":...,"backlinks":{"formal":[...],"semantic":[...]}}`); same failure shapes/exit codes as T007.
- [X] T013 [US2] Implement `NewBacklinksCmd` in new file `internal/cli/internalcmd/backlinks.go` — argument parsing and JSON shaping only, calling `operations.Backlinks` directly. Depends on T012, T011.
- [X] T014 [US2] Wire `internalcmd.NewBacklinksCmd()` into `newInternalCmd`'s `AddCommand` list in `internal/cli/internal.go`. Depends on T013 **and** T009 (same file as User Story 1's own wiring task — sequential to avoid a conflicting edit, regardless of which story's implementation finishes first).

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/operations/... -run TestBacklinks` and `go test ./internal/cli/... -run TestBacklinks` both pass on their own.

---

## Phase 5: User Story 3 - Existing Capabilities and Link-Free Artifacts Remain Unaffected (Priority: P3)

**Goal**: Every artifact and every existing capability (creation, inspection, structural validation) continues to behave exactly as before this feature existed; a broken/malformed wikilink never causes a References/Backlinks query itself to fail.

**Independent Test**: Run the full existing suite of creation, inspection, and structural-validation scenarios (001 through 011) unmodified, and confirm every result is identical to before this feature existed.

### Tests for User Story 3

> Write this test FIRST; confirm it passes only once US1/US2's own implementation is complete and correct (this story is a regression guarantee, not new behavior).

- [X] T015 [US3] Dedicated test in `internal/operations/references_test.go`/`backlinks_test.go` directly proving spec.md's own User Story 3 acceptance scenarios: querying references/backlinks on a project containing a broken, malformed, or ambiguous wikilink succeeds (`ok:true`, no error) rather than failing — the integrity problem is simply absent from the answer, never surfaced as a query-level failure. Depends on T014.

### Regression Gate for User Story 3

- [X] T016 [US3] Regression checkpoint: run `go test ./internal/ids/...`, `go test ./internal/validation/...`, `go test ./internal/operations/...`, and `go test ./internal/cli/...` in full and confirm every pre-existing test (from 002-read-operations, 004-structural-validation, 008-cli-cobra, and 011-wikilink-foundation respectively) passes completely unmodified — the explicit, named proof of SC-003. Depends on T015.

**Checkpoint**: User Story 3 is independently complete and testable — the full pre-existing suites are green, unmodified, alongside both new capabilities.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T017 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 011-wikilink-foundation's packages included) and fix any findings.
- [X] T018 [P] Verify/extend package-level doc comments on `internal/ids`, `internal/operations`, and `internal/cli/internalcmd`, cross-checked against `specs/012-references-backlinks/contracts/references-backlinks.md`.
- [X] T019 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: query references and backlinks against a small fixture graph, confirm a broken wikilink is excluded, confirm an artifact with no relationships returns empty, confirm an out-of-scope target (Task) is rejected — each matching `quickstart.md`'s own documented result exactly.
- [X] T020 Reconcile `specs/012-references-backlinks/contracts/references-backlinks.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T021 Full regression run: `go test ./...` across the entire module (001 through 012) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/ids/... ./internal/operations/... ./internal/validation/... -race` clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS both User Story 1 and User Story 2 (both need `ids.ResolveTarget`).
- **User Story 1 (Phase 3)**: Depends on Foundational only. No functional dependency on User Story 2.
- **User Story 2 (Phase 4)**: Depends on Foundational only. No functional dependency on User Story 1 — except its own CLI-wiring task (T014) must follow User Story 1's (T009), a file-conflict ordering, not a story dependency.
- **User Story 3 (Phase 5)**: Depends on User Story 1 and User Story 2 both being complete — there is nothing to prove "unaffected by" until both new capabilities exist.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational. No dependency on User Story 2.
- **User Story 2 (P2)**: Can start immediately after Foundational, in parallel with User Story 1 — except T014, which must follow T009.
- **User Story 3 (P3)**: Depends on both User Story 1 and User Story 2 being complete.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V) — except User Story 3, which is a regression *guarantee* about already-implemented behavior.
- Foundational's regression gate (T004) passes before either User Story 1 or User Story 2's implementation begins.

### Parallel Opportunities

- T001 (Foundational's own test) has no sibling to parallelize with in this feature (unlike 011's T001+T002 pair) — Foundational here is a single, small, linear chain.
- Once Foundational is done: **User Story 1 and User Story 2 can proceed in parallel** (different files throughout: `references.go`/`references_test.go` vs. `backlinks.go`/`backlinks_test.go`) — the one exception is T009/T014 both touching `internal/cli/internal.go`, which must stay sequential.
- T005 and T010 (the two stories' own test-writing tasks) can run in parallel.
- T007 and T012 (the two stories' own CLI test-writing tasks) can run in parallel.
- Within Polish: T017 and T018 in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (`ids.ResolveTarget`, regression-proven).
3. Complete Phase 3: User Story 1 (`References`, plus its own CLI command).
4. **STOP and VALIDATE**: `go test ./internal/operations/... -run TestReferences` and the CLI's own references tests green, independently.
5. This alone already proves the outgoing half of the graph query surface works.

### Incremental Delivery

1. Setup (nothing) + Foundational (`ids.ResolveTarget`, regression-proven).
2. Add User Story 1 → validate independently → outgoing references usable (MVP).
3. Add User Story 2 (independent of US1, sharing only Foundational) → validate independently → incoming backlinks usable.
4. Add User Story 3 (depends on US1 and US2 both existing) → validate independently → non-regression formally proven.
5. Polish (Phase 6), including the full-module `-race` regression run (T021).

### Team Strategy

Unlike 007's and 011's strictly sequential chains, this feature's two
core stories (References, Backlinks) are genuinely independent once
Foundational is done — two developers could take one each in parallel,
coordinating only on the shared `internal/cli/internal.go` wiring
(T009/T014).

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- A wikilink counts as a graph edge only when it resolves to exactly one real artifact (research.md #3) — broken, malformed, and ambiguous are all excluded identically, in both References and Backlinks.
- A formal relationship is reported exactly as declared, with no existence check of its own — that remains 004-structural-validation's job, never duplicated here.
- `errors.go`'s `classify` needs zero changes anywhere in this feature (research.md #4) — both new commands reuse `operations`'s existing sentinels verbatim.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
