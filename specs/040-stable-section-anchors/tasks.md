---

description: "Task list for Referências a Seções com Âncoras Estáveis"
---

# Tasks: Referências a Seções com Âncoras Estáveis

**Input**: Design documents from `/specs/040-stable-section-anchors/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/stable-anchors-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic parsing/validation/retrieval logic, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US3)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. This feature widens the same five packages Spec 038 already touched: `internal/artifacts` (parsing), `internal/validation` (classification), `internal/operations` (occurrence plumbing), `internal/context` (+ `index`) (retrieval), `internal/cli/internalcmd` (versioned output). No new package.

---

## Phase 1: Setup

No setup phase — this feature extends five existing packages with no new dependency, package, or scaffolding (plan.md "Scale/Scope").

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The parsing primitives every user story builds on — an anchor must be parseable out of both a heading and a wikilink before anything else in this feature can work.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundational

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T001 [P] Add cases to `internal/artifacts/document_test.go` (write first): a heading line `## Retry Policy {#retry-policy}` parses into `Section.Heading == "Retry Policy"` and `Section.Anchor == "retry-policy"`; a heading with no `{#...}` suffix has `Anchor == ""` and `Heading` unchanged; a heading with a malformed/unterminated suffix (e.g. `## Retry Policy {#` with no closing `}`) is treated as literal heading text — `Anchor == ""`, `Heading` includes the literal `{#` text — never an error, per `ParseDocument`'s existing always-succeeds contract (data-model.md "Section"). Expected to FAIL until T004.
- [X] T002 [P] Add cases to `internal/artifacts/chunk_test.go` (write first): a `Section` with a non-empty `Anchor` produces a `Chunk` with the same `Anchor`; a `Section` with empty `Body` still produces no `Chunk` at all from `Chunks()` (existing, unchanged behavior — data-model.md "Chunk (extended)" note). Expected to FAIL until T005.
- [X] T003 [P] Add cases to `internal/artifacts/wikilink_test.go` (write first): `[[KNOW-003#retry-policy]]` parses into `Target == "KNOW-003"`, `Anchor == "retry-policy"`, `Alias == ""`; `[[KNOW-003#retry-policy|Política de retries]]` parses into all three fields correctly (alias split happens first, then anchor split, per research.md #2); `[[KNOW-003]]` and `[[KNOW-003|Alias]]` are completely unaffected — `Anchor == ""` in both, `Target`/`Alias` exactly as before this feature. Expected to FAIL until T006.

### Implementation for Foundational

- [X] T004 Extend `headingPattern`-based parsing in `internal/artifacts/document.go` to recognize an optional trailing `{#slug}` suffix on a heading line, stripping it into a new `Section.Anchor` field (contracts §1) and leaving `Heading` clean; a malformed/unterminated suffix stays literal text in `Heading` with `Anchor == ""` (depends on T001).
- [X] T005 Add `Anchor string` to `Chunk` in `internal/artifacts/chunk.go`, copied verbatim from the originating `Section.Anchor` in `Chunks()`'s existing loop (depends on T002, T004).
- [X] T006 Add `Anchor string` to `WikiLink` in `internal/artifacts/wikilink.go`; extend the target-parsing helper so `TARGET#anchor` (the pre-alias portion) splits into `Target`/`Anchor` — alias split (`|`) happens first, exactly as today, then the remaining target string splits on the first `#` (research.md #2, contracts §1) (depends on T003).

**Checkpoint**: Foundation ready — an anchor can be declared on a heading and referenced in a wikilink; nothing downstream consumes either field yet.

---

## Phase 3: User Story 1 - Referenciar apenas a seção relevante (Priority: P1) 🎯 MVP

**Goal**: An anchor-qualified wikilink retrieves exactly the referenced section (plus a minimal ancestor-heading breadcrumb), never the whole target artifact; a non-anchor wikilink is completely unaffected.

**Independent Test**: Reference a section by anchor in a multi-section target artifact and confirm the retrieved content is that section alone, not the whole artifact; confirm a plain `[[ID]]` reference to the same artifact is unaffected.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T007 [P] [US1] Add a case to `internal/operations/references_test.go` (write first): `References` on an artifact containing `[[KNOW-003#retry-policy]]` returns a `ReferenceEntry` with `TargetAnchor == "retry-policy"`; a plain `[[KNOW-003]]` reference in the same artifact returns `TargetAnchor == ""` (data-model.md "ReferenceEntry / BacklinkEntry (extended)"). Expected to FAIL until T012.
- [X] T008 [P] [US1] Add the equivalent case to `internal/operations/backlinks_test.go` for `BacklinkEntry.TargetAnchor`. Expected to FAIL until T013.
- [X] T009 [P] [US1] Add cases to `internal/context/collector_test.go` (write first, per contracts §5): an anchor-qualified reference to a multi-section target produces exactly one `Candidate` (the matching Section's own Chunk), with `HeadingPath` set to that Section's ancestor titles outermost-first — not one `Candidate` per Chunk of the whole target (today's `chunkArtifact` behavior); a non-anchor reference to the same target still produces the full per-Chunk candidate set, completely unchanged; an anchor on a Section whose own `Body` is empty (heading immediately followed by a subheading) still resolves to exactly one `Candidate` with empty `Content` — never treated as not-found (spec Edge Cases, data-model.md "Chunk (extended)" note). Expected to FAIL until T015-T016.
- [X] T010 [P] [US1] Add cases to `internal/context/index/schema_test.go` and `sync_test.go` (write first): the `chunks` table's `anchor` column and the `links` table's `target_anchor` column are populated correctly by `indexChunks`/`indexLinks`; a database built under `schemaVersion 2` is detected as stale and fully rebuilt under `schemaVersion 3` (existing rebuild-on-mismatch mechanism, 038 precedent). Expected to FAIL until T017-T018.
- [X] T011 [P] [US1] Add a case to the existing golden-envelope test pattern for `internal/cli/internalcmd/context.go` (write first): the envelope's `schema_version` is `5`; an item produced via an anchor-qualified reference carries `heading_path`/`anchor` keys; an item produced without an anchor omits both keys entirely (not empty-valued — absent, 038's own convention); a non-anchor response is otherwise byte-identical to `schema_version: 4`'s shape aside from the version number (contracts §6). Expected to FAIL until T019.

### Implementation for User Story 1

- [X] T012 [US1] Add `TargetAnchor string` to `ReferenceEntry` in `internal/operations/references.go`, populated as `link.Anchor` in `References`'s existing semantic-entry loop (contracts §3) (depends on T006, T007).
- [X] T013 [US1] Add `TargetAnchor string` to `BacklinkEntry` in `internal/operations/backlinks.go`, populated as `link.Anchor` in `Backlinks`'s existing semantic-matching loop (depends on T006, T008).
- [X] T014 [US1] Add `HeadingPath []string` to `Candidate` and `TargetAnchor string` to `Reason` in `internal/context/result.go` (data-model.md "Candidate / Reason (extended)") (depends on T009).
- [X] T015 [US1] Implement `chunkArtifactAnchor(root, relPath, anchor string, reason Reason) (Candidate, error)` in `internal/context/collector.go` per contracts §5: parse the target's `Document`, locate the `Section` whose `Anchor` matches (including an empty-`Body` Section — resolve against `Document.Sections` directly, not `artifacts.Chunks()`'s already-filtered output, per data-model.md's note), build `HeadingPath` from that Section's ancestor headings (by `Level` nesting, outermost first), and return exactly one `Candidate` (depends on T004, T005, T014).
- [X] T016 [US1] Update `referenceEntryReason`/`backlinkEntryReason` in `internal/context/collector.go` to copy `TargetAnchor` from the driving entry into `Reason`; update `connectedCandidates` to call `chunkArtifactAnchor` instead of `chunkArtifact` whenever the built `Reason.Relation == "wikilink"` and `Reason.TargetAnchor != ""` (depends on T012, T013, T015).
- [X] T017 [US1] Widen `internal/context/index/schema.go`: add `anchor TEXT` to the `chunks` table and `target_anchor TEXT` to the `links` table; bump `schemaVersion` 2 → 3 (data-model.md "Index Schema (extended)", contracts §7) (depends on T010).
- [X] T018 [US1] Update `indexChunks`/`indexLinks` in `internal/context/index/sync.go` to populate the two new columns from `Chunk.Anchor`/`ReferenceEntry.TargetAnchor`/`BacklinkEntry.TargetAnchor` (depends on T005, T012, T013, T017).
- [X] T019 [US1] Bump `contextSchemaVersion` 4 → 5 in `internal/cli/internalcmd/context.go`; add `heading_path`/`anchor` keys to a context item's JSON output, present only when the item came from an anchor-qualified reference (contracts §6) (depends on T011, T014, T016).
- [X] T020 [US1] Add the additive `target_anchor` field to `internal/cli/internalcmd/references.go` and `backlinks.go`'s entry output — no version bump, unversioned envelope, 038 precedent (contracts §4) (depends on T012, T013).

**Checkpoint**: User Story 1 is independently functional — an anchor-qualified reference retrieves exactly one section, plainly distinguishable from a whole-artifact reference, with the index/CLI contract versions bumped consistently.

---

## Phase 4: User Story 2 - Âncora sobrevive a uma pequena edição de título (Priority: P2)

**Goal**: Renaming a heading's title while keeping its declared `{#anchor}` preserves every existing reference to that anchor; two Sections in one artifact declaring the same anchor are caught before being trusted.

**Independent Test**: Declare an anchor, reference it, rename only the heading's title text, and confirm the reference still resolves without any edit; declare the same anchor twice in one artifact and confirm validation flags it.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T021 [P] [US2] Add a case to `internal/validation/wikilinks_test.go` (write first): an artifact with two Sections both declaring `{#retry-policy}` produces exactly one `Finding` with `Code == CodeDuplicateAnchor`, naming the artifact; an artifact with no duplicate anchors produces no such Finding (data-model.md "Validation Codes"). Expected to FAIL until T023-T024.
- [X] T022 [P] [US2] Add a case to `internal/context/collector_test.go` (or extend T009's fixture): starting from a working anchor-qualified reference (T009), rename only the target Section's heading title text while keeping `{#anchor}` unchanged, and confirm the reference still resolves to the same Section's content — no wikilink edit involved (spec User Story 2, Acceptance Scenario 1).

### Implementation for User Story 2

- [X] T023 [US2] Add `CodeDuplicateAnchor = "duplicate_anchor"` to `internal/validation/findings.go` (data-model.md "Validation Codes") (depends on T021).
- [X] T024 [US2] Implement `checkAnchors(root, cfg, filePath)` in `internal/validation/wikilinks.go` (same shape as `checkWikilinks`): scan one artifact's own `Document.Sections` for a repeated non-empty `Anchor`, raising `CodeDuplicateAnchor` — independent of whether anything currently references that anchor (contracts §2); wire it into the same validation pass that already invokes `checkWikilinks` per artifact (depends on T004, T023).
- [X] T025 [US2] Verification task (no production code change expected): run T022's test and quickstart.md §2 manually, confirming a title-only edit with an unchanged `{#anchor}` requires no change anywhere in the retrieval/validation pipeline built by Phase 3 — record the result; if anything *does* need a change, that is a Phase 3 defect to fix there, not new scope here (depends on T015, T022).

**Checkpoint**: User Stories 1 and 2 are both independently functional — anchors survive title edits, and duplicates are caught.

---

## Phase 5: User Story 3 - Âncora quebrada nunca vira link para o documento inteiro (Priority: P2)

**Goal**: A wikilink referencing a nonexistent anchor within an existing target artifact is flagged with a diagnostic distinct from a nonexistent target artifact — never silently treated as a valid whole-document link.

**Independent Test**: Remove a declared anchor without updating its referencing wikilink and confirm validation reports a specific, distinct "unknown anchor" finding, never conflated with "broken wikilink."

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T026 [P] [US3] Add a case to `internal/validation/wikilinks_test.go` (write first): a wikilink `[[KNOW-003#nonexistent]]` where `KNOW-003` exists but declares no Section with that `Anchor` produces exactly one `Finding` with `Code == CodeUnknownAnchor` (data-model.md "Validation Codes"). Expected to FAIL until T028-T029.
- [X] T027 [P] [US3] Add a case to `internal/validation/wikilinks_test.go` proving `CodeBrokenWikilink` (target artifact `SPEC-999` doesn't exist at all, anchor or not) and `CodeUnknownAnchor` (target exists, anchor doesn't) are mutually exclusive — never both raised for the same link, and the two remain distinguishable by `Code` value alone (spec FR-009, Acceptance Scenario US3.2).

### Implementation for User Story 3

- [X] T028 [US3] Add `CodeUnknownAnchor = "unknown_anchor"` to `internal/validation/findings.go` (depends on T026).
- [X] T029 [US3] Extend `classifyWikilink` in `internal/validation/wikilinks.go`: when `matches == 1` (the target artifact exists) and `link.Anchor != ""`, parse that artifact's own `Document` and confirm a `Section` with a matching `Anchor` exists; return `CodeUnknownAnchor` if not, no Finding otherwise (contracts §2) (depends on T004, T027, T028).

**Checkpoint**: All three user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the feature end-to-end and leave the repository's own documentation consistent with the new syntax.

- [X] T030 [P] Update `docs/context-engine-implementation.md` §6.1 (currently lines ~213-220, "Supported syntax") to document `[[ID#anchor]]`/`[[ID#anchor|Alias]]` alongside the existing `[[ID]]`/`[[ID|Alias]]` examples, and §6.4 (currently lines ~280-284, the `broken_wikilink`/`ambiguous_wikilink`/`invalid_wikilink` list) to add `unknown_anchor` and `duplicate_anchor` to the documented validation vocabulary.
- [X] T031 [P] Manually walk through every command in `quickstart.md` (§1-§5) and confirm actual output matches each section's "Expected" description, updating quickstart.md if any wording drifted from the final implementation.
- [X] T032 Run `go build ./...`, `go vet ./...`, and `go test ./...` for the whole repository and resolve any regression before considering the feature complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: None — skipped.
- **Foundational (Phase 2)**: BLOCKS all user stories — an anchor must be parseable before anything can validate, retrieve, or resolve it.
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on other stories.
- **User Story 2 (Phase 4)**: Depends on Foundational and, for T022/T025's verification, on Phase 3's retrieval path (T015/T016) already existing.
- **User Story 3 (Phase 5)**: Depends on Foundational only — its validation-only scope has no dependency on Phase 3/4's retrieval or duplicate-detection work.
- **Polish (Phase 6)**: Depends on Phases 3-5 all being complete.

### Within Each User Story

- Tests written and failing before their corresponding implementation task (Principle V).
- `internal/artifacts` parsing (Foundational) before any package that consumes `Anchor`.
- `ReferenceEntry`/`BacklinkEntry.TargetAnchor` before `Reason.TargetAnchor` before `connectedCandidates`'s branch.
- Index schema widening before sync population before the CLI's version bump (so a rebuilt index always has the columns the new envelope code expects).

### Parallel Opportunities

- T001, T002, T003 (Foundational tests) run in parallel — different files.
- T007, T008, T009, T010, T011 (Phase 3 tests) run in parallel — different files, none depending on another's implementation.
- Phase 5 (US3) can be staffed in parallel with Phase 3/4 after Foundational completes, since its validation-only scope touches only `internal/validation`.
- T030/T031 (Phase 6 docs) run in parallel.

---

## Parallel Example: Foundational

```bash
# All three foundational test files can be written together:
Task: "Add heading-anchor-suffix cases to internal/artifacts/document_test.go"
Task: "Add Chunk.Anchor propagation cases to internal/artifacts/chunk_test.go"
Task: "Add wikilink anchor/alias-splitting cases to internal/artifacts/wikilink_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational.
2. Complete Phase 3: User Story 1.
3. **STOP and VALIDATE**: run quickstart.md §1 — an anchor-qualified reference retrieves exactly one section, a non-anchor reference is unaffected.

### Incremental Delivery

1. Foundational → anchors parseable everywhere.
2. User Story 1 → precise section-scoped retrieval (MVP).
3. User Story 2 → title-rename stability + duplicate-anchor detection.
4. User Story 3 → broken-anchor diagnostic, distinguishable from broken-wikilink.
5. Polish → docs updated, full suite green.

### Parallel Team Strategy

After Phase 2: one contributor takes Phase 3 (US1)'s retrieval/schema/CLI work, a second takes Phase 5 (US3)'s validation-only work in parallel (no shared files with Phase 3 beyond `findings.go`, which is append-only). Phase 4 (US2) is best taken by whoever did Phase 3, since T025 verifies against Phase 3's own retrieval path.

---

## Notes

- [P] tasks touch different files with no ordering dependency between them.
- Every "write the test first" task above is explicitly called out; per Principle V this is NON-NEGOTIABLE for this feature's entirely-deterministic scope.
- Commit after each task or logical group.
- Stop at each Checkpoint to validate that story's Independent Test from spec.md before moving on.
- Avoid: adding a second Markdown/heading parser instead of extending `document.go`'s existing one; conflating `CodeUnknownAnchor` with `CodeBrokenWikilink`; letting `chunkArtifactAnchor` fall back to `chunkArtifact` on a not-found anchor (validation, not retrieval, owns that diagnostic — collector.go's own contract note in contracts §5 assumes the anchor already exists by the time retrieval runs).

## Post-Implementation Fix (found by code review, not part of the original 32 tasks)

- [X] T033 Fix `chunkArtifactAnchor` (`internal/context/collector.go`) silently returning a near-empty, contentless `Candidate` (no Heading/Content, `StartLine`/`EndLine` 0, but a real nonzero score) when a wikilink names an anchor that doesn't exist on an otherwise-real target artifact. This path is reachable in normal use — `operations.References`/`Backlinks` populate `TargetAnchor` from raw wikilink text with no existence check — contradicting contracts §5's original "never reached here in practice" assumption. Fixed by adding a sentinel `errAnchorNotFound`, returned by `chunkArtifactAnchor` and checked with `errors.Is` in `connectedCandidates`, which now silently omits that one reference instead of fabricating a placeholder Candidate or failing the whole `Collect` call. Added `TestCollect_UnknownAnchorReferenceIsOmittedNotFabricated` in `internal/context/collector_test.go`; corrected contracts §5's doc comment and inline note to record the correction.
