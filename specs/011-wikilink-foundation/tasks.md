---

description: "Task list template for feature implementation"
---

# Tasks: Wikilink Graph Foundation

**Input**: Design documents from `/specs/011-wikilink-foundation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/wikilinks.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 004-structural-validation's and 008-cli-cobra's full test suites are explicit, named regression gates (User Story 3, Polish), since this feature must not change their behavior at all for any artifact that doesn't use the new capability.

**Organization**: Tasks are grouped by user story. **User Story 1 (`ExtractWikiLinks`) depends only on Foundational** — it is a pure function operating on raw bytes, not on `ReadBody` or any real file. **User Story 2 (the three Finding codes) depends on both User Story 1** (it consumes `ExtractWikiLinks`'s output) **and Foundational** (it needs `ReadBody` to get a real artifact's body). **User Story 3 (non-regression) depends on User Story 1 and 2 existing** — there is nothing to prove "unaffected" until the new capability itself exists to be unaffected by.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
internal/artifacts/            # existing package — Foundational, US1
internal/validation/            # existing package — US2, US3
internal/example/                # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: This feature adds no new package and no new dependency (research.md). There is nothing to front-load before Foundational.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Safely widen `internal/artifacts`'s existing frontmatter-delimiter scanner to also expose an artifact's body — the one piece of genuinely shared, existing-code-touching infrastructure both later stories build on (research.md's `splitFrontmatter` decision). This is the only place existing, already-shipped code changes at all in this feature.

**⚠️ CRITICAL**: No user story implementation may begin until this phase's regression gate (T004) passes.

- [X] T001 [P] Regression tests confirming `ParseMetadata`'s existing behavior is completely unchanged after the planned `extractFrontmatter` → `splitFrontmatter` widening, in `internal/artifacts/parser_test.go` — every existing test in this file is itself part of this gate; add any additional case needed to pin down current behavior precisely before the refactor.
- [X] T002 [P] Tests for `ReadBody` in `internal/artifacts/markdown_test.go`: a normal artifact returns exactly its post-frontmatter body; an artifact with an empty body (nothing after the closing delimiter) returns an empty body, not an error; a file with missing or malformed frontmatter returns an error consistent with `ParseMetadata`'s own existing error vocabulary (`ErrFrontmatterMalformed`/`ErrArtifactNotFound`) rather than a second, parallel one.
- [X] T003 Implement `splitFrontmatter(data []byte) (frontmatter, body []byte, err error)` (widening `extractFrontmatter`, data-model.md) in `internal/artifacts/parser.go`, updating `ParseMetadata`'s own call site to use only the `frontmatter` half; implement `ReadBody(path string) ([]byte, error)` in new file `internal/artifacts/markdown.go`. Depends on T001, T002.
- [X] T004 Regression checkpoint: run `go test ./internal/artifacts/...` and confirm every existing test (002-read-operations through 010-interactive-init-tui's own fixtures) passes unmodified — the explicit, named proof this refactor changed nothing observable about `ParseMetadata`. Depends on T003.

**Checkpoint**: Foundation ready — `internal/artifacts` can produce both an artifact's frontmatter (unchanged) and its body (new), with zero regression. User story implementation can begin.

---

## Phase 3: User Story 1 - Author Explicit Semantic Links Between Artifacts (Priority: P1) 🎯 MVP

**Goal**: `ExtractWikiLinks` deterministically parses every `[[TARGET]]`/`[[TARGET|Alias]]` link from a body of Markdown text, in document order, with correct source line — and correctly ignores everything that isn't one.

**Independent Test**: Feed `ExtractWikiLinks` a body containing plain links, aliased links, multiple links per line, links across multiple lines, and text that looks similar but isn't a valid link (standard Markdown links, inline code, a bare bracket); confirm every valid link is extracted correctly and nothing else is.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T006.

- [X] T005 [P] [US1] Unit tests for `ExtractWikiLinks` in `internal/artifacts/wikilink_test.go`, covering `docs/context-engine-implementation.md` §29.1's own list: a simple link; an aliased link; multiple links on one line; multiple links across multiple lines, in document order, each with its correct line number; a standard Markdown link (`[text](url)`) is not extracted; text inside an inline code span (`` `[[NOT-A-LINK]]` ``) is not extracted; text inside a fenced code block is not extracted; an unterminated `[[` with no closing `]]` is not extracted (spec.md's own Edge Case); a bare single bracket is not extracted; at least one target of each supported entity-ID prefix (`PRG`, `FEAT`, `SPEC`, `TASK`, `KNOW`, `LRN`) round-trips correctly.

### Implementation for User Story 1

- [X] T006 [US1] Implement `WikiLink` and `ExtractWikiLinks` in new file `internal/artifacts/wikilink.go` — a line-based scanner tracking fenced-code-block state across lines and masking single-backtick inline code spans within a line before matching `[[...]]` (research.md); never resolves a target (FR-004). Depends on T005.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/artifacts/... -run TestExtractWikiLinks` passes on its own, with zero dependency on `ReadBody` or `internal/validation`.

---

## Phase 4: User Story 2 - Detect Broken, Malformed, and Ambiguous Links (Priority: P2)

**Goal**: The project's existing structural validation capability reports every link's target as exactly one of `invalid_wikilink`, `broken_wikilink`, `ambiguous_wikilink`, or no finding at all — reusing `internal/ids`'s existing syntax/resolution rules exactly (research.md).

**Independent Test**: Create artifacts with a link to a real target, a link to a nonexistent ID, a link with malformed target syntax, and a link to an ID that resolves to more than one artifact; run structural validation and confirm each produces its own specific, correctly labeled finding — and the valid link produces none.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T008–T010.

- [X] T007 [P] [US2] Unit tests for the new wikilink check in `internal/validation/wikilinks_test.go`, covering `docs/context-engine-implementation.md` §29.2's own list: a valid, resolvable target produces no finding; a target that doesn't exist produces `broken_wikilink`; a target with malformed ID syntax produces `invalid_wikilink`, distinct from `broken_wikilink`; a target using a different (but otherwise valid) ID zero-padding width than the project's configured width still resolves correctly, matching how frontmatter fields already tolerate this; a target resolving to more than one existing artifact produces `ambiguous_wikilink`, distinct from both; an alias never affects resolution — two links with the same target but different aliases classify identically.

### Implementation for User Story 2

- [X] T008 [P] [US2] Add `CodeInvalidWikilink`, `CodeBrokenWikilink`, `CodeAmbiguousWikilink` constants in `internal/validation/findings.go` (data-model.md). Depends on T007.
- [X] T009 [US2] Implement `checkWikilinks(root string, cfg project.Configuration, filePath string) []Finding` in new file `internal/validation/wikilinks.go` — reads the artifact's body via `artifacts.ReadBody`, extracts links via `artifacts.ExtractWikiLinks`, classifies each via `ids.ParseAny`/`ids.Scan` exactly as data-model.md's table specifies. Depends on T006 (US1), T004 (Foundational), T008.
- [X] T010 [US2] Wire `checkWikilinks` into `checkEntity`, `internal/validation/validator.go` — one new call site, for each of the five already-validated entity types (Program, Feature, Spec, Knowledge, Learning; research.md's scope decision). Depends on T009.

**Checkpoint**: User Story 2 is independently complete and testable — `go test ./internal/validation/... -run TestCheckWikilinks` (or equivalent) passes on its own; all three codes correctly distinguished.

---

## Phase 5: User Story 3 - Existing Artifacts and Validation Remain Unaffected (Priority: P3)

**Goal**: An artifact with zero links anywhere in its body produces byte-for-byte identical validation results to before this feature existed — for every artifact, not just a new one written to prove the point.

**Independent Test**: Run the full existing suite of project and artifact validation scenarios (004-structural-validation's own fixtures) unmodified, and confirm every result is identical to before this feature existed.

### Tests for User Story 3

> Write this test FIRST; confirm it passes only once US1/US2's own implementation is complete and correct (this story is a regression guarantee, not new behavior — "red" here means "the guarantee isn't yet provable," not "code is missing").

- [X] T011 [US3] Dedicated test in `internal/validation/wikilinks_test.go` confirming an artifact whose body contains zero links produces zero wikilink-related findings — an explicit, separate positive case beyond what US2's own tests already imply, directly proving spec.md's own User Story 3 acceptance scenario. Depends on T010.

### Regression Gate for User Story 3

- [X] T012 [US3] Regression checkpoint: run `go test ./internal/artifacts/...` and `go test ./internal/validation/...` in full and confirm every pre-existing test (from 002-read-operations and 004-structural-validation respectively) passes completely unmodified — the explicit, named proof of SC-003. Depends on T011.

**Checkpoint**: User Story 3 is independently complete and testable — the full pre-existing suites are green, unmodified, alongside the new capability.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T013 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 010-interactive-init-tui's packages included) and fix any findings.
- [X] T014 [P] Verify/extend package-level doc comments on `internal/artifacts` and `internal/validation`, cross-checked against `specs/011-wikilink-foundation/contracts/wikilinks.md`.
- [X] T015 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: extract a link from a fixture body, then validate a fixture project containing a valid link, a broken link, and a malformed link, confirming each produces the exact finding (or lack thereof) `quickstart.md` documents.
- [X] T016 Reconcile `specs/011-wikilink-foundation/contracts/wikilinks.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T017 Full regression run: `go test ./...` across the entire module (001 through 011) green, `go vet ./...` clean, `gofmt -l .` empty, `go build ./cmd/misterspec` succeeds, `go test ./internal/artifacts/... ./internal/validation/... -race` clean.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do — proceed directly to Foundational.
- **Foundational (Phase 2)**: BLOCKS User Story 2 (needs `ReadBody`) — does not block User Story 1, which needs nothing from this phase beyond the module compiling.
- **User Story 1 (Phase 3)**: Depends on Foundational only in the trivial sense of the module building; no functional dependency on `ReadBody`.
- **User Story 2 (Phase 4)**: Depends on User Story 1 (`ExtractWikiLinks`) and Foundational (`ReadBody`).
- **User Story 3 (Phase 5)**: Depends on User Story 1 and User Story 2 both being complete — there is nothing to prove "unaffected by" until the new capability itself exists.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational (or in parallel with it, in practice, since it doesn't need `ReadBody`). No dependency on US2/US3.
- **User Story 2 (P2)**: Depends on User Story 1's `ExtractWikiLinks` and Foundational's `ReadBody` — cannot start meaningfully before both exist.
- **User Story 3 (P3)**: Depends on User Story 2 being complete — it proves a guarantee *about* the new capability, so the capability must exist first.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V) — except User Story 3, which is a regression *guarantee* about already-implemented behavior, not new code to drive with a failing test in the usual sense (noted explicitly on T011).
- Foundational's regression gate (T004) passes before User Story 2's implementation begins.

### Parallel Opportunities

- T001 and T002 (Foundational's two test files) in parallel.
- Once Foundational is done: **User Story 1 can proceed immediately**; User Story 2 must wait for User Story 1's `ExtractWikiLinks` to exist, so the two are *not* independently parallel the way 003's/006's/009's own story pairs were — this feature's chain (US1 → US2 → US3) is closer to 007-project-bootstrap's own strictly sequential shape.
- Within Polish: T013 and T014 in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (nothing to do).
2. Complete Phase 2: Foundational (`splitFrontmatter`/`ReadBody`, regression-proven).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/artifacts/... -run TestExtractWikiLinks` green, independently.
5. This alone already proves the deterministic core of this feature works — parsing is correct — even before validation consumes it.

### Incremental Delivery

1. Setup (nothing) + Foundational (`ReadBody`, regression-proven).
2. Add US1 → validate independently → link parsing usable (MVP).
3. Add US2 (depends on US1) → validate independently → link integrity checking usable.
4. Add US3 (depends on US2) → validate independently → non-regression formally proven.
5. Polish (Phase 6), including the full-module `-race` regression run (T017).

### Team Strategy

This feature's story chain (US1 → US2 → US3) is strictly sequential,
unlike several prior features' parallel story pairs — a single
developer working through it in order is the natural shape, the same
as 007-project-bootstrap's own Inspect → Bootstrap → Verify chain.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `ExtractWikiLinks` never resolves a target (FR-004) — resolution is `internal/validation`'s job alone, kept strictly separate (plan.md's Constitution Check, Principle I).
- No CLI-layer file is touched anywhere in this feature (research.md) — the three new Finding codes flow through `internal/cli/internalcmd/validate.go`'s already-generic JSON mapping unchanged.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
