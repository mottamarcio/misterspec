---

description: "Task list for Reutilização Incremental de Context Packs"
---

# Tasks: Reutilização Incremental de Context Packs

**Input**: Design documents from `/specs/043-incremental-context-reuse/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/incremental-context-reuse-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic hashing/diffing/storage logic, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US3)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. No new package. Additive changes to `internal/context` (`pack.go`), `internal/context/index` (`schema.go`, `store.go`, `sqlite.go`), and `internal/cli/internalcmd` (`context.go`).

---

## Phase 1: Setup

**Purpose**: No new package to scaffold — this feature only extends existing files. Setup is limited to confirming the extension points are where plan.md expects.

- [X] T001 Confirm `internal/context/pack.go`, `internal/context/index/schema.go`/`store.go`/`sqlite.go`, and `internal/cli/internalcmd/context.go` exist and match plan.md's "Project Structure" (no action if they already do — this is a verification-only task, no file changes expected).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Pack identity hashing (`ConfigHash`/`PackID`) and the `packs` storage table — every user story needs both.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundational

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T002 [P] Add `internal/context/pack_test.go` cases (write first): `ComputeConfigHash` returns the identical hash for two `ConfigIdentity` values with identical fields, and a different hash when any single field (`Target`, `Intent`, `Task`, `Query`, `QueryMode`, `Budget`, `HardLimit`, `PreferSection`, `RankingVersion`, `ContextSchemaVersion`, `Estimator`) differs, one field at a time (data-model.md "ConfigIdentity", contracts §1). Expected to FAIL until T006.
- [X] T003 [P] Add cases to `internal/context/pack_test.go` (write first): `ComputePackID` returns the identical `PackID` for two `PackIdentity` values with identical `Config` and `Items`; returns a different `PackID` when `Config` differs (same items) or when `Items` differs (same config) — including when only one item's own `Fingerprint` differs while `Path`/`StartLine`/`EndLine` stay the same (data-model.md "ConfigHash / PackID"). Expected to FAIL until T006.
- [X] T004 [P] Add `internal/context/index/schema_test.go` cases (write first): after `ensureSchema` runs on a fresh database, a `packs` table exists with columns `pack_id` (primary key), `config_hash`, `target`, `created_at`, `items_json`; a database created under the previous `schemaVersion` (3) is dropped and recreated with `packs` present, mirroring the existing version-mismatch-triggers-rebuild test pattern this file already has for `chunks`/`links` (data-model.md "StoredPack", research.md #3). Expected to FAIL until T007.
- [X] T005 [P] Add `internal/context/index/sqlite_test.go` cases (write first): `SavePack` followed by `LookupPack` with the same `pack_id` returns `found == true` and a `StoredPack` matching every field saved (round-trip); `LookupPack` with an unknown `pack_id` returns `found == false` and `err == nil` (contracts §2's "bool separate from error" convention); saving more packs than the fixed cap evicts the oldest (by `created_at`) first, keeping the table bounded (research.md #3). Expected to FAIL until T008.

### Implementation for Foundational

- [X] T006 [P] Implement `internal/context/pack.go` additions: `ConfigHash`/`PackID` (`"sha256:<hex>"` strings), `ConfigIdentity`/`PackIdentity`/`PackItemIdentity` structs, `ComputeConfigHash(cfg ConfigIdentity) ConfigHash`, and `ComputePackID(id PackIdentity) PackID` (hash of `ComputeConfigHash(id.Config)` concatenated with `id.Items`'s own canonical serialization) — reusing the SHA-256 approach `Fingerprint` already establishes in this same file, never a second hashing scheme (data-model.md, contracts §1, research.md #2/#3) (depends on T002, T003).
- [X] T007 [P] Add the `packs` table to `internal/context/index/schema.go`'s `schemaStatements`/`dropStatements` (columns: `pack_id TEXT PRIMARY KEY`, `config_hash TEXT NOT NULL`, `target TEXT NOT NULL`, `created_at INTEGER NOT NULL`, `items_json TEXT NOT NULL`); bump `schemaVersion` from 3 to 4 (data-model.md "StoredPack", contracts §2) (depends on T004).
- [X] T008 Add `StoredPack`/`StoredPackItem` types to `internal/context/index/store.go`; add `SavePack(pack StoredPack) error` and `LookupPack(packID string) (pack StoredPack, found bool, err error)` to the `Store` interface; implement both on `*sqliteStore` in `internal/context/index/sqlite.go` — `SavePack` upserts by `pack_id`, JSON-encodes `Items` into `items_json`, then evicts rows beyond a fixed cap ordered oldest-`created_at`-first; `LookupPack` JSON-decodes `items_json` back into `[]StoredPackItem`, returning `found == false` with no error when the row doesn't exist (contracts §2, research.md #3/#4) (depends on T006, T007, T005).

**Checkpoint**: Foundation ready — pack identity can be computed and compared, and packs can be persisted/retrieved; nothing yet computes a diff or wires this into the CLI.

---

## Phase 3: User Story 1 - Pedir só o que mudou desde o último pacote (Priority: P1) 🎯 MVP

**Goal**: A caller that already has a base pack and passes `--base <pack_id>` gets back only what changed since that base — an empty diff when nothing changed, a targeted diff when part of the content changed — and applying that diff to the base reproduces the exact full pack.

**Independent Test**: Fetch `--mode package` twice for the same target with no source change in between, the second time with `--base` set to the first response's own `pack_id`; confirm the second response's diff has zero `"added"`/`"modified"` entries. Then change one Requirement's text and repeat; confirm exactly one `"modified"` entry appears.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T009 [P] [US1] Add `internal/context/pack_test.go` cases (write first): `DiffAgainstBase(current, base)` returns an entries list that is all `"reuse"` (no `content` anywhere) when `current` and `base` have identical `(Path, StartLine, EndLine, Fingerprint)` per item, in the same order; returns exactly one `"modified"` entry (carrying the item's own current `Content`) when one item's `Fingerprint` differs but its identity is unchanged, with every other item as `"reuse"`; returns one `"added"` entry for an item whose identity is new to `current`; returns one `RemovedEntry` for a base identity absent from `current` (data-model.md "DiffEntry"/"RemovedEntry"/"PackDiff", contracts §1). Expected to FAIL until T011.
- [X] T010 [P] [US1] Add `internal/context/pack_test.go` cases (write first): reconstructing a full item list from a `PackDiff` (substituting each `"reuse"` entry with `base[BaseIndex]`, keeping every `"added"`/`"modified"` entry's own `Item`, in the diff's own `Entries` order) produces a slice byte-identical, item by item, to `current` — the property spec FR-003 requires (data-model.md "PackDiff"). Expected to FAIL until T011.
- [X] T011 [US1] Implement `DiffAgainstBase(current []PackageItem, base []index.StoredPackItem) PackDiff` in `internal/context/pack.go`: builds an identity index (`Path, StartLine, EndLine`) over `base`, walks `current` in its own final order emitting a `"reuse"` entry (with `BaseIndex`) when an identical-fingerprint match exists in `base`, otherwise an `"added"`/`"modified"` entry carrying the current item's own full data; any `base` identity never matched while walking `current` is appended to `PackDiff.Removed` (data-model.md, contracts §1, research.md #5) (depends on T009, T010, T006).
- [X] T012 [P] [US1] Add `internal/cli/internalcmd/context_test.go` cases (write first): a first `--mode package` call (no `--base`) includes a non-empty `pack_id` in the response; a second call with `--base` set to that `pack_id`, no source change in between, returns `context.diff.entries` containing only `"reuse"` entries and `context.reuse.items_sent == 0`; after editing the source Spec, a third call with the same `--base` returns exactly one `"modified"` diff entry and `context.reuse.items_reused` equal to the item count minus 1 (contracts §3's first three example shapes, quickstart.md steps 1-3). Expected to FAIL until T013.

### Implementation for User Story 1

- [X] T013 [US1] Wire `--base <pack_id>` into `internal/cli/internalcmd/context.go`: reject it (`invalid_argument`) unless `--mode package` is also set; after computing `result`/`BuildPackageItems` as today, always compute this call's own `ConfigIdentity`/`PackIdentity`/`PackID` and persist it via `store.SavePack` (so a *future* call can use it as a base — every `--mode package` call saves a pack, not only ones already using `--base`); when `--base` is empty, render `pack_id` alongside the existing full `items` (unchanged shape otherwise); when `--base` is set and `store.LookupPack` finds a row whose `ConfigHash` matches this call's own freshly computed `ConfigHash`, call `contextengine.DiffAgainstBase` and render `context.diff` + `context.reuse` per contracts §3's second/third examples (depends on T008, T011, T012).

**Checkpoint**: User Story 1 is independently functional — a repeat `--base` call with no change returns an empty diff; a real content change returns a minimal, targeted diff; applying either diff to its own base reproduces the exact full pack.

---

## Phase 4: User Story 2 - Recuperar o pacote completo com segurança quando a base não existe mais (Priority: P1)

**Goal**: A `--base` value the system cannot confirm (unknown `pack_id`, or a config mismatch) always yields the full package, explicitly marked `recovered: true` — never a partial or silently-incomplete result.

**Independent Test**: Pass `--base` with a `pack_id` that was never returned by any prior call; confirm the response is the full package with `recovered: true` and `reason: "unknown_base"`, never an error and never a `diff`.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T014 [P] [US2] Add `internal/cli/internalcmd/context_test.go` cases (write first): `--base` set to a `pack_id`-shaped string never returned by any prior call in this test yields `context.recovered == true`, `context.reason == "unknown_base"`, and `context.items` is the full package (same shape as a `--base`-less call, no `diff` key present) (contracts §3's fourth example, quickstart.md step 4). Expected to FAIL until T015.
- [X] T015 [P] [US2] Add cases to `internal/context/index/sqlite_test.go` (write first): `LookupPack` with a `pack_id` that was never saved returns `found == false, err == nil` (already covered by T005 — this task adds the CLI-facing consequence: a command layer that calls `LookupPack` and gets `found == false` for an unrecognized `--base` never proceeds to `DiffAgainstBase`). Expected to FAIL until T016 (this is a thin CLI-orchestration test, distinct from T005's own storage-layer round-trip test).

### Implementation for User Story 2

- [X] T016 [US2] In `internal/cli/internalcmd/context.go`'s `--base` handling (T013), when `store.LookupPack` returns `found == false`, render `context.recovered = true`, `context.reason = "unknown_base"`, and `context.items` as the full package (the same `renderPackageItems` call already used for a `--base`-less request) instead of attempting `DiffAgainstBase` — `context.reuse` still present, with `items_reused: 0` and `items_sent` equal to the full item count (data-model.md "RecoveryResult"/"ReuseDiagnostics", contracts §3/§4) (depends on T013, T014, T015).

**Checkpoint**: User Story 2 is independently functional — an unrecognized `--base` always yields a complete, explicitly-marked full response; the caller's own claim of possessing a base is never trusted without a matching stored row.

---

## Phase 5: User Story 3 - Nunca reutilizar conteúdo desatualizado por engano (Priority: P2)

**Goal**: A stored base whose own request configuration (intent, budget, ranking version, contract version, estimator, …) no longer matches the current call is treated as invalid and falls back to full recovery — distinct from a pure content change, which stays diffable.

**Independent Test**: Save a base pack with one configuration, then request `--base` against the same `pack_id` using a different `--budget` (or after a hypothetical `contextSchemaVersion` bump); confirm the response is `recovered: true, reason: "invalidated"`, not a diff and not a stale reuse.

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T017 [P] [US3] Add cases to `internal/context/pack_test.go` (write first): `ComputeConfigHash` differs between two `ConfigIdentity` values that differ only in `Budget` (spec Acceptance Scenario 1), and differs between two values that differ only in `ContextSchemaVersion` (spec Acceptance Scenario 2) — the exact two dimensions spec User Story 3 names explicitly. Expected to already PASS after T006 (this task adds targeted regression coverage naming the spec's own scenarios, not new production behavior) — confirm and mark accordingly rather than expecting a failure.
- [X] T018 [P] [US3] Add `internal/cli/internalcmd/context_test.go` cases (write first): a `--base` call reusing a `pack_id` saved under one `--budget` value, now called with a different `--budget`, yields `context.recovered == true, context.reason == "invalidated"` (quickstart.md step 5); a `--base` call reusing a `pack_id` where only the underlying source content changed (same config) still yields a `diff` (not `recovered`) — confirming config-invalidation and content-diffing are distinct, non-overlapping paths (spec Acceptance Scenario 3, US1's T012 already covers the content-only case — this task asserts they remain distinguishable after T019 lands). Expected to FAIL until T019.

### Implementation for User Story 3

- [X] T019 [US3] In `internal/cli/internalcmd/context.go`'s `--base` handling, when `store.LookupPack` finds a row but its `ConfigHash` does not equal this call's own freshly computed `ConfigHash`, render `context.recovered = true`, `context.reason = "invalidated"` (distinct from T016's `"unknown_base"`), and the full package — never attempting `DiffAgainstBase` against a config-incompatible base (data-model.md "RecoveryResult", research.md #4) (depends on T013, T017, T018).

**Checkpoint**: User Story 3 is independently functional — a config-mismatched base always falls back to full recovery with the correct `reason`; a content-only change (same config) remains a normal diff, never conflated with invalidation.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification across all three stories together, plus the documentation this feature's own FR-010 requires.

- [X] T020 [P] Add a case to `internal/context/index/sqlite_test.go` or `schema_test.go`: a `packs` row saved under the previous `schemaVersion` is gone after a version bump (`ensureSchema`'s existing drop-and-recreate path) — confirming pack loss on an unrelated schema change is safe, never a correctness problem, only a missed reuse opportunity (spec FR-011, research.md #3).
- [X] T021 [P] Run `gofmt -l` and `go vet ./...` across every new/changed file in this feature and fix any findings.
- [X] T022 Manually run all 6 scenarios in `quickstart.md` against a built `misterspec` binary in a real fixture project, confirming each "Expected" outcome — including step 6's byte-identical reconstruction check.
- [X] T023 [P] Update `site/commands.html`'s existing `internal context` section (`internal-context` anchor) with `--base`'s own flag description, the `pack_id`/`diff`/`recovered`/`reuse` response fields, and one short example — following the same documentation precedent 042-impact-analysis-review's own polish pass established for a new/changed command, and explicitly stating (spec FR-010) that this reuse mechanism is independent of any LLM provider's own prompt caching.
- [X] T024 [P] Final pass on every new exported symbol's doc comment in `internal/context/pack.go` and `internal/context/index/{store,sqlite}.go`, confirming they describe the shipped behavior (config-hash-gated diffing, bounded eviction) rather than phase-by-phase build-out scope notes.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion. MVP — delivers the diff itself.
- **User Story 2 (Phase 4)**: Depends on Foundational completion; its implementation task (T016) extends the same `context.go` `--base` handling US1's T013 creates, so implement after US1.
- **User Story 3 (Phase 5)**: Depends on Foundational completion; its implementation task (T019) extends the same `context.go` `--base` handling US1/US2 already touch, so implement last among the three stories.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other stories — the diff path itself is self-contained once Foundational lands.
- **User Story 2 (P1)**: Independently testable once implemented, but its implementation task (T016) touches the same `--base` branch US1's T013 creates — sequence by file, not by required feature order.
- **User Story 3 (P2)**: Independently testable once implemented; its implementation task (T019) touches the same `--base` branch US1/US2 already modified — sequence last.

### Within Each User Story

- Tests written and failing before implementation (Principle V).
- `internal/context/pack.go` (pure computation) before `internal/context/index` (storage) before `internal/cli/internalcmd/context.go` (orchestration/wiring) — matches Foundational's own T006→T007/T008 order and each story's own Txxx→CLI-wiring order.

### Parallel Opportunities

- T002-T005 (all Foundational tests) can run in parallel — different files.
- T006 and T007 (pack.go / schema.go implementations) can run in parallel — different files, no shared dependency between them; T008 depends on both plus T005.
- T009-T010 (US1 pure-diff tests) and T012 (US1 CLI test) can run in parallel.
- T014-T015 (US2 tests) can run in parallel.
- T017-T018 (US3 tests) can run in parallel.
- T020-T021, T023-T024 (Polish) can run in parallel.

---

## Parallel Example: Foundational

```bash
# Launch all Foundational tests together:
Task: "Add ComputeConfigHash cases to internal/context/pack_test.go"
Task: "Add ComputePackID cases to internal/context/pack_test.go"
Task: "Add packs-table cases to internal/context/index/schema_test.go"
Task: "Add SavePack/LookupPack round-trip cases to internal/context/index/sqlite_test.go"

# Once tests are in place, launch the two independent implementation files together:
Task: "Implement ConfigHash/PackID/ComputeConfigHash/ComputePackID in internal/context/pack.go"
Task: "Add the packs table to internal/context/index/schema.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run `quickstart.md` scenarios 1-3 and 6 against a real fixture project.
5. `internal context --base` is now genuinely useful for the most common case (repeat calls with no or small content changes) even before the safety/invalidation paths (US2/US3) exist as distinct, tested behavior — though T013 already always falls back to a full package for anything T016/T019 haven't specifically handled yet, so MVP-only is never unsafe, only less explicit about *why* recovery happened.

### Incremental Delivery

1. Setup + Foundational → identity hashing and storage ready.
2. Add User Story 1 → diffing works for the common case → validate independently (MVP!).
3. Add User Story 2 → unknown-base recovery explicitly marked → validate independently (quickstart scenario 4).
4. Add User Story 3 → config-mismatch invalidation explicitly distinguished from unknown-base → validate independently (quickstart scenario 5).
5. Polish → schema-bump-loses-packs regression test, full quickstart re-run, `site/commands.html` update, doc-comment accuracy pass.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- US2 and US3 both extend the same `--base` branch in `internal/cli/internalcmd/context.go` that US1 creates — independently *testable* (each adds its own acceptance-scenario coverage) even though not independently deployable as separate binaries, mirroring `042-impact-analysis-review`'s own precedent for a later story extending an earlier story's own file with a documented scope note.
- Verify each story's tests fail before implementing that story's tasks.
- Commit after each task or logical group.
- Stop at any checkpoint to validate a story independently.
