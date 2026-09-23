---

description: "Task list template for feature implementation"
---

# Tasks: Context Pack completo e contrato de saída versionado

**Input**: Design documents from `/specs/033-context-pack-output-contract/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/context-pack-output-contract.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story and MUST be written and confirmed failing before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths are included in every description

## Path Conventions

Single Go module/CLI project (plan.md "Structure Decision") — all paths are relative to the repository root, inside the existing `internal/artifacts`, `internal/context`, and `internal/cli/internalcmd` packages. No new top-level directory; one new file (`internal/context/pack.go`).

---

## Phase 1: Setup

**Purpose**: Establish a clean, verified starting point. No new dependencies, no new project structure.

- [X] T001 Run `go build ./...` and `go test ./...` from the repository root and confirm a clean baseline (no pre-existing failures) before making any change, so any later red test is attributable to this feature's work.

**Checkpoint**: Baseline confirmed green — safe to start Foundational work.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The file-absolute line-offset fix (research.md Decision 1) and the `--mode` flag's own parsing/enum — the two things every user story below builds on.

**⚠️ CRITICAL**: No user story task may start until this phase is complete.

- [X] T002 [P] Write failing unit tests in `internal/artifacts/markdown_test.go` for a new `ReadBodyWithOffset` function: a file with frontmatter returns the body plus the correct 1-indexed file-absolute starting line of the body; a file with no frontmatter returns an offset of `0` (research.md Decision 1's Edge Case, data-model.md "ItemLocation" validation rule).
- [X] T003 Implement `ReadBodyWithOffset` in `internal/artifacts/markdown.go`, reusing `splitFrontmatter` (already computes the frontmatter's own line span) to also report the body's starting line, to make T002 pass.
- [X] T004 [P] Write failing test in `internal/context/collector_test.go`: `chunkArtifact` on a fixture `spec.md` with frontmatter followed by multiple headed sections produces `Candidate`s whose `StartLine`/`EndLine` are file-absolute (verified against the fixture's own real line numbers), not relative to the post-frontmatter body — this is the actual bug confirmed in research.md Decision 1, independent of any CLI flag.
- [X] T005 Update `chunkArtifact` in `internal/context/collector.go` to call `ReadBodyWithOffset` and add the returned offset to every `Chunk.StartLine`/`EndLine` before building each `Candidate`, to make T004 pass (depends on T003).
- [X] T006 [P] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: `internal context <id> --mode package` and `--mode markdown` are accepted; `--mode manifest` and omitting `--mode` are equivalent; `--mode bogus` fails with `invalid_argument` (contracts §1's mode column, data-model.md "OutputMode").
- [X] T007 Add `--mode` flag parsing and enum validation to `internal/cli/internalcmd/context.go` (`manifest` default, `package`, `markdown`; anything else rejected via the existing `ErrInvalidArgument` sentinel), to make T006 pass — no output-shape changes yet.

**Checkpoint**: File-absolute `Candidate` locations and `--mode` flag parsing exist — User Stories 1, 2, and 3 can now proceed in priority order.

---

## Phase 3: User Story 1 - Receber o conteúdo selecionado sem releitura obrigatória (Priority: P1) 🎯 MVP

**Goal**: `--mode package` returns each item's own selected text and a content fingerprint directly in the response; the default (manifest) mode explicitly never includes content, so a caller cannot mistake "metadata-only" for "content included."

**Independent Test**: Request `--mode package` for a Spec with selectable content; every item's `content` matches the corresponding source text without re-reading the file. Request the default mode; confirm no item has a `content` field (quickstart.md Scenario 1).

### Tests for User Story 1 ⚠️

- [X] T008 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: `internal context <id> --mode package` returns items whose `content` field is non-empty and matches the corresponding section of the source fixture file, verified without a second file read (spec Acceptance Scenarios 1–2).
- [X] T009 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: the default (no `--mode`, or `--mode manifest`) response has no `content` key on any item (spec Acceptance Scenario 3).
- [X] T010 [P] [US1] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: `--mode package` items include a `fingerprint` field shaped `"sha256:<hex>"`; the fingerprint changes after editing that item's own source section and is unchanged across two calls with no edit in between (spec Edge Cases — fingerprint detects source drift, research.md Decision 4).

### Implementation for User Story 1

- [X] T011 [US1] Create `internal/context/pack.go` and implement package-mode item content assembly — reusing `ResultItem.Content` already held in memory from `Collect` (no re-read) — to make T008 pass (depends on T007).
- [X] T012 [US1] Implement per-item fingerprint in `internal/context/pack.go`: SHA-256 of the item's own `Content` bytes, rendered `"sha256:<hex>"` (research.md Decision 4, matching `operations.FileFingerprint`'s string format without importing `internal/operations`), to make T010 pass (depends on T011).
- [X] T013 [US1] Wire `--mode package`'s item response in `internal/cli/internalcmd/context.go` to include `content` and `fingerprint` via the new `pack.go` helpers; confirm `manifest` and `markdown` modes remain unaffected (no `content`/`fingerprint` field), to make T008/T009 pass (depends on T011, T012).

**Checkpoint**: User Story 1 is independently functional and testable — full content and a drift-detectable fingerprint are delivered on request, and the lightweight mode stays explicitly content-free.

---

## Phase 4: User Story 2 - Localizar cada trecho no arquivo de origem (Priority: P1)

**Goal**: Every `--mode package` item carries a `location` whose `start_line`/`end_line` are absolute to the whole source file (frontmatter included in the count), correct whether or not the file has frontmatter, and correct for a Task-derived item inside a shared `tasks.md`.

**Independent Test**: Request `--mode package` for a Spec with frontmatter; open the source file at `location.start_line` and confirm it lands exactly on that item's own first line (quickstart.md Scenario 2).

### Tests for User Story 2 ⚠️

- [X] T014 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: for a fixture `spec.md` with frontmatter, `--mode package` items' `location.start_line`/`end_line` match that fixture's own real, file-absolute line numbers (spec Acceptance Scenario 1).
- [X] T015 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: the same check against a fixture artifact with no frontmatter — location is still correct (spec Acceptance Scenario 2, Edge Case).
- [X] T016 [P] [US2] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: a package-mode item derived from a `## TASK-NNN — Title` heading inside a shared `tasks.md` has a `heading` and `location` bounded to exactly that Task's own section, not the whole file (spec FR-009, research.md Decision 7).

### Implementation for User Story 2

- [X] T017 [US2] Add the `location` field (`path`, file-absolute `start_line`/`end_line` — already correct after T005) to package-mode item assembly in `internal/context/pack.go`, and wire it into `context.go`'s `--mode package` response, to make T014–T016 pass (depends on T013, T005).

**Checkpoint**: User Stories 1 AND 2 both work independently — content and its exact file-absolute location are both delivered, including for Task-derived items.

---

## Phase 5: User Story 3 - Consumir um contrato de saída versionado e sem duplicação (Priority: P2)

**Goal**: Every response identifies its `schema_version`; `--mode package` never also includes a duplicate `rendered` Markdown blob; requesting `--render` together with `--mode package` is rejected outright; diagnostics reflect the actual size of what was returned; and the pre-existing default/`--render` output stays byte-for-byte unchanged.

**Independent Test**: Compare a default-mode and a `--render` response against a captured pre-change baseline — identical except for two new additive fields. Request `--mode package --render` together and confirm it is rejected (quickstart.md Scenarios 4–6).

### Tests for User Story 3 ⚠️

- [X] T018 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: `context.schema_version` equals `1` in the manifest, package, and markdown mode responses alike.
- [X] T019 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: a `--mode package` response has `rendered: null` — no item's content is duplicated into a rendered field in the same response (spec Acceptance Scenario 2).
- [X] T020 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: `internal context <id> --mode package --render` fails with `{"ok":false,"error":{"code":"invalid_argument",...}}` and exit code 2 (contracts §1's rejected row, §6).
- [X] T021 [P] [US3] Write failing regression CLI test in `internal/cli/internalcmd/context_test.go`: the default (no flags) and `--render` responses are identical, field-for-field, to a captured pre-change baseline fixture, except for the two new additive fields (`schema_version`, `diagnostics.payload_tokens`) (spec Acceptance Scenario 3, FR-008/SC-003).
- [X] T022 [P] [US3] Write failing CLI test in `internal/cli/internalcmd/context_test.go`: `diagnostics.payload_tokens` is present and `>= diagnostics.tokens_selected` in `--mode package`; `diagnostics.tokens_selected` itself is numerically identical to what the same request returns in manifest mode (spec FR-007).

### Implementation for User Story 3

- [X] T023 [US3] Add `schema_version: 1` to the `context` envelope in `internal/cli/internalcmd/context.go`, present in every mode, to make T018 pass.
- [X] T024 [US3] Implement the `--render` + `--mode package` mutual-exclusivity rejection (`ErrInvalidArgument`) in `internal/cli/internalcmd/context.go` per contracts §1/§6 (research.md Decision 3), to make T020 pass (depends on T007).
- [X] T025 [US3] Implement `payload_tokens` estimation in `internal/context/pack.go` — the serialized size (item metadata plus content for package mode, or the Markdown string for markdown mode) of the actual response for the requested mode — and wire it into `context.go`'s `diagnostics` object, to make T022 pass (depends on T013, T017).
- [X] T026 [US3] Capture a pre-change baseline fixture (default and `--render` responses against a fixed test project) and assert the post-change responses match it except for the two additive fields, to make T021 pass.

**Checkpoint**: All three user stories are independently functional — full content, exact location, and a versioned, non-duplicating, backward-compatible contract.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the whole feature end-to-end and keep the published docs in sync.

- [X] T027 [P] Re-read `kit/skills/mister-{plan,tasks,analyze,implement,wrap-up}/SKILL.md`'s own references to `internal context` and confirm none need changes — none currently pass `--render` or depend on `content`/`location`/`fingerprint` (research.md Decision 2), so this feature ships without requiring any Skill edit; note `--mode package` as a now-available option only if it fits naturally, without rewriting a Skill's own re-read step as part of this feature.
- [X] T028 [P] Update `site/commands.html`'s `internal context` section with the new `--mode` flag, `schema_version`, and a `--mode package` example, mirroring 031/032's precedent of keeping the published command reference in sync with newly shipped behavior.
- [X] T029 Manually execute every scenario in `quickstart.md` (1–6) against a disposable temp project built from the finished binary, and confirm each documented "Pass condition" holds.
- [X] T030 Run `go build ./...`, `go test ./...`, and `go vet ./...` from the repository root and confirm a fully green result across the whole module, not just the packages touched by this feature.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup (T001). BLOCKS every user story (T008–T026).
- **User Story 1 (Phase 3)**: Depends on Foundational (T002–T007) only.
- **User Story 2 (Phase 4)**: Depends on Foundational, and specifically reuses User Story 1's `pack.go` item-assembly scaffold (T011) to attach `location` onto the same package-mode item shape — sequenced after US1's T011–T013, not fully parallel with it.
- **User Story 3 (Phase 5)**: `schema_version` (T023) and the mutual-exclusivity rejection (T024) depend only on Foundational; `payload_tokens` (T025) depends on both US1's content assembly (T013) and US2's location assembly (T017), since it estimates the full package item's serialized size.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on US2 or US3.
- **User Story 2 (P1)**: Builds on US1's `pack.go` scaffold (same `PackageItem` assembly, different field) — not independent of US1 the way 031's three stories were, but independently testable once US1's T011–T013 land.
- **User Story 3 (P2)**: `payload_tokens` depends on US1 and US2 both being complete (it measures their combined output); `schema_version` and the mutual-exclusivity rejection do not.

### Within Each User Story

- Tests are written and confirmed failing before their corresponding implementation task (Principle V).
- Content assembly (T011) before fingerprint (T012) before wiring into the CLI response (T013).
- Location assembly (T017) after T013, since it extends the same package-mode item the CLI response already renders.
- `payload_tokens` (T025) after both `content` (T013) and `location` (T017) exist, since it measures their combined serialized size.

### Parallel Opportunities

- T002 and T004 (Foundational tests, different files/layers — `artifacts` vs. `context`) — run in parallel.
- T006 (Foundational CLI flag test) can run in parallel with T002/T004 — different concern (flag parsing vs. line offset).
- T008, T009, T010 (US1 tests) — independent fixtures within `context_test.go`; run in parallel.
- T014, T015, T016 (US2 tests) — independent fixtures; run in parallel with each other, but only after US1's T011–T013 land.
- T018–T022 (US3 tests) — independent fixtures; run in parallel with each other.
- T027 and T028 (Polish doc updates) — different files; run in parallel.

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together once Phase 2 is complete:
Task: "CLI test: --mode package items include matching content in internal/cli/internalcmd/context_test.go"
Task: "CLI test: default mode has no content field in internal/cli/internalcmd/context_test.go"
Task: "CLI test: fingerprint present, changes on edit, stable otherwise in internal/cli/internalcmd/context_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001).
2. Complete Phase 2: Foundational (T002–T007) — CRITICAL, blocks all stories.
3. Complete Phase 3: User Story 1 (T008–T013).
4. **STOP and VALIDATE**: run quickstart.md Scenario 1 against a temp project.
5. This alone already closes the core gap (content delivered without a mandatory re-read) and is independently shippable — `--mode` defaults to `manifest`, so nothing existing changes.

### Incremental Delivery

1. Setup + Foundational → absolute line offsets and `--mode` flag parsing ready.
2. Add User Story 1 → full content + fingerprint on request → validate independently (MVP).
3. Add User Story 2 → exact file-absolute location on the same package items → validate independently.
4. Add User Story 3 → versioned envelope, no-duplication guard, and proven backward compatibility → validate independently.
5. Polish → confirm Skills/docs match code, run full quickstart, whole-repo test gate.

### Parallel Team Strategy

1. One contributor completes Setup + Foundational (T001–T007) — the shared dependency every story needs.
2. Once Phase 2 lands:
   - Contributor A: User Story 1 (`internal/context/pack.go`'s content+fingerprint, `context.go` wiring).
   - Contributor B: joins after US1's T011–T013 land to add User Story 2's `location` field onto the same scaffold.
   - Contributor C: User Story 3's `schema_version` and mutual-exclusivity rejection can start immediately after Foundational; `payload_tokens` waits on both A and B.
3. Polish (Phase 6) runs once all three land.

---

## Notes

- `[P]` tasks touch different files or independent fixtures within a shared test file — no shared mutable state.
- `[Story]` labels map every Phase 3+ task to its spec.md user story for traceability.
- Confirm each test fails before writing its implementation (Principle V, NON-NEGOTIABLE).
- The default (no `--mode`) and `--render` output MUST remain byte-for-byte unchanged except for the two additive fields — treat any other diff in T021/T026 as a regression to fix, not a baseline to update.
- Avoid: vague tasks, same-file conflicts inside a `[P]` group, and any change to `renderContextItems`'s existing field names/values.
