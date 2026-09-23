---

description: "Task list for Skills Enxutas e Contratos de Integração Testáveis"
---

# Tasks: Skills Enxutas e Contratos de Integração Testáveis

**Input**: Design documents from `/specs/039-lean-skills-integration-contracts/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/skillgen-and-eval-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic logic and contract-conformance checking, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US4)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. `internal/skillgen/` (new package), `internal/example/` (existing spec-level integration tests), `internal/eval/` (existing evaluation package), `kit/skills/*/SKILL.md` (existing canonical Skills, becoming generator output).

---

## Phase 1: Setup

**Purpose**: Scaffold the new package this feature adds.

- [X] T001 Create `internal/skillgen/doc.go` with package doc explaining `internal/skillgen` composes the canonical `kit/skills/*/SKILL.md` files from named fragments + per-Skill manifests (data-model.md), and is a development-time authoring tool with no CLI wiring (plan.md "Scale/Scope" — no new `misterspec` subcommand).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The composition types and engine every user story's tasks build on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 Define `Fragment`, `Part`, `SectionEntry`, `SkillManifest` types in `internal/skillgen/manifest.go` per contracts/skillgen-and-eval-contract.md §1 and data-model.md "Fragment"/"SkillManifest"/"SectionEntry"/"Part": `Fragment.Name` MUST be unique; `Fragment.Body` MUST NOT reference another fragment's `Name`; `Part` is exactly one of `FragmentRef`/`Bespoke`; `SkillManifest.Sections` MUST cover exactly the 26 headings from `requiredSkillHeadings` (`internal/example/skills_content_test.go`) in that exact order.
- [X] T003 [P] Write `internal/skillgen/manifest_test.go` (write first, confirm it fails): valid manifest passes a `Validate()`-style check; a manifest missing a required heading fails; a manifest with headings out of order fails; a manifest with a duplicate `Fragment.Name` across `KnownFragments` fails.
- [X] T004 Implement `skillgen.Generate(manifest SkillManifest, fragments []Fragment) ([]byte, error)` in `internal/skillgen/generate.go` per contracts §1: composes frontmatter + each `SectionEntry`'s heading + concatenated `Parts` (resolving `FragmentRef` against `fragments`, using `Bespoke` verbatim) into final Markdown bytes; returns an error naming the manifest's `SkillName` and the specific problem when a `FragmentRef` does not resolve, when a resolved `Fragment.Section` disagrees with the enclosing `SectionEntry.Heading`, or when `Sections` does not cover exactly the 26 required headings in order (depends on T002).
- [X] T005 [P] Write `internal/skillgen/generate_test.go` (write first, confirm it fails before T004, pass after): fragment resolves under matching section (pass, exact byte output asserted); `FragmentRef` unresolved (error); `Fragment.Section` mismatch (error); missing/out-of-order required heading (error) (depends on T002; verifies T004).
- [X] T006 [P] Add a documented baseline constant to `internal/example/skills_content_test.go`, e.g. `const preFeatureSkillLineCount = 2378 // kit/skills/*/SKILL.md non-frontmatter lines, recorded by research.md #6 before this feature's changes`, for the Phase 7 size assertion to compare against.

**Checkpoint**: Foundation ready — `internal/skillgen` can compose a manifest into bytes and reports every malformed-input case correctly.

---

## Phase 3: User Story 1 - Manter regras comuns em uma única fonte (Priority: P1) 🎯 MVP

**Goal**: A shared-rule edit made once (in a fragment) is reflected in every Skill that references it, without per-file edits; Skill-specific exceptions are preserved; the generated content stays byte-identical to hand-reviewed Markdown.

**Independent Test**: Change a fragment's `Body`, regenerate, and confirm every referencing Skill's committed file reflects the change while non-referencing Skills are untouched.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T007 [P] [US1] Write `internal/example/skillgen_drift_test.go` per contracts §1's drift-check contract: for every `m := range skillgen.Manifests`, assert `skillgen.Generate(m, skillgen.KnownFragments)` byte-equals the committed content of `kit/skills/<m.SkillName>/SKILL.md`, failing with the Skill name and a diff on mismatch. Expected to FAIL until `Manifests`/`KnownFragments` exist (T009-T017) and are reconciled with committed files (T018).

### Implementation for User Story 1

- [X] T008 [US1] Define `KnownFragments` in `internal/skillgen/fragment.go` with exactly two fragments identified by research.md #1: `resolve-preamble` (Section: "Deterministic Operations", Body: the exact two-line bullet `` `internal resolve SPEC-###` — confirm the Spec exists and locate its canonical file.`` as it appears identically today in `kit/skills/mister-plan/SKILL.md:103-104`, `mister-tasks/SKILL.md:100-101`, `mister-analyze/SKILL.md:107-108`, `mister-wrap-up/SKILL.md:113-114`) and `mechanical-steps-note` (Section: "Deterministic Operations", Body: the exact sentence `Use misterspec operations for every mechanical repository step.` as it appears identically today in `kit/skills/mister-features/SKILL.md:87`, `mister-analyze/SKILL.md:103`, `mister-constitution/SKILL.md:123`, `mister-specify/SKILL.md:87`, `mister-plan/SKILL.md:99`, `mister-program/SKILL.md:94`) (depends on T002).
- [X] T009 [P] [US1] Build the `SkillManifest` for `mister-plan` in `internal/skillgen/manifest.go`, reproducing `kit/skills/mister-plan/SKILL.md`'s current content exactly, with its "Deterministic Operations" `SectionEntry` referencing `resolve-preamble` and `mechanical-steps-note` via `FragmentRef` instead of inlining that text, and every other section as a single `Bespoke` Part (depends on T002, T008).
- [X] T010 [P] [US1] Build the `SkillManifest` for `mister-tasks`, referencing `resolve-preamble` (depends on T002, T008).
- [X] T011 [P] [US1] Build the `SkillManifest` for `mister-analyze`, referencing `resolve-preamble` and `mechanical-steps-note` (depends on T002, T008).
- [X] T012 [P] [US1] Build the `SkillManifest` for `mister-wrap-up`, referencing `resolve-preamble` (depends on T002, T008).
- [X] T013 [P] [US1] Build the `SkillManifest` for `mister-features`, referencing `mechanical-steps-note` (depends on T002, T008).
- [X] T014 [P] [US1] Build the `SkillManifest` for `mister-constitution`, referencing `mechanical-steps-note` (depends on T002, T008).
- [X] T015 [P] [US1] Build the `SkillManifest` for `mister-specify`, referencing `mechanical-steps-note` (depends on T002, T008).
- [X] T016 [P] [US1] Build the `SkillManifest` for `mister-program`, referencing `mechanical-steps-note` (depends on T002, T008).
- [X] T017 [P] [US1] Build the `SkillManifest`s for `mister-implement` and `mister-knowledge-base` (neither shares today's two identified fragments verbatim, per research.md #1) as fully `Bespoke` Parts reproducing their current content section-for-section, so all 10 Skills are generator-managed and covered by T007's drift-check (depends on T002).
- [X] T018 [US1] Audit every Skill's Required/Optional/Deterministic Operations sections for restatement of the pre-034 manual multi-call context-gathering flow now superseded by `internal prepare` (035/034); trim any found into the affected manifest's `Bespoke` Parts. Confirmed by research.md's investigation that `mister-implement` already documents `prepare` as primary with `context` as an explicit fallback (034 already integrated there) — this task records the audit result and applies any trim actually found in the other 9 Skills, rather than assuming duplication exists (spec FR-003) (depends on T009-T017).
- [X] T019 [US1] Populate `Manifests []SkillManifest` in `internal/skillgen/manifest.go` (or a new `internal/skillgen/manifests.go`) as the fixed slice of all 10 Skills' manifests from T009-T018 (depends on T009-T018).
- [X] T020 [US1] Add a `//go:generate` directive plus a small regeneration entrypoint (e.g. `internal/skillgen/cmd/gen/main.go`, run via `go generate ./internal/skillgen/...`) that calls `skillgen.Generate` for every `skillgen.Manifests` entry and writes the result to `kit/skills/<name>/SKILL.md`, matching quickstart.md §1's documented command (depends on T004, T019).
- [X] T021 [US1] Run the regeneration entrypoint from T020 and commit the result: `kit/skills/*/SKILL.md` for Skills untouched by T018's trims MUST stay byte-identical to their pre-feature content (fragment-only substitution is lossless); Skills touched by T018 reflect the trim. Confirm T007's drift-check test now passes (depends on T007, T018, T020).

**Checkpoint**: User Story 1 is independently functional — editing `resolve-preamble` or `mechanical-steps-note` in `internal/skillgen/fragment.go` and re-running T020's generator changes every referencing Skill at once; the drift-check test (T007) guards against silent hand-edit divergence.

---

## Phase 4: User Story 2 - Skills não prometem verificações que o binário não pode cumprir (Priority: P1)

**Goal**: Every `internal <op>` citation (already checked), every example invocation, and every promised verification in a Skill is automatically confirmed against a real binary capability.

**Independent Test**: Introduce a broken citation of each new kind and confirm the test fails naming the Skill and the offending content.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T022 [P] [US2] In `internal/example/skills_content_test.go`, add a failing-fixture-driven test asserting that a fenced `misterspec …` example naming a nonexistent subcommand or flag is rejected — using an in-test synthetic Skill body, not a real `kit/skills` file, so the negative case doesn't require breaking real content (Principle V, regression coverage per spec Acceptance Scenario US2.3). Expected to FAIL until T023 exists.
- [X] T024 [P] [US2] In `internal/example/skills_content_test.go`, add a failing-fixture-driven test asserting that a `CapabilityClaim` with neither `BackingCode` nor `BackingOp` set (or both set) is rejected by the allowlist check (spec Acceptance Scenario US2.2). Expected to FAIL until T025/T026 exist.

### Implementation for User Story 2

- [X] T023 [US2] Implement the example-syntax check in `internal/example/skills_content_test.go` per contracts §2 (FR-005): build the real `*cobra.Command` tree from `internal/cli`'s existing constructor, extract every fenced `misterspec …` invocation from each `kit/skills/*/SKILL.md` body, and assert `Find()` plus every named flag resolve against that real tree; fail naming the Skill and the offending example line on mismatch (depends on T022, and reuses the existing per-Skill parsing already in this file).
- [X] T025 [P] [US2] Define `skillVerificationAllowlist map[string][]CapabilityClaim` in `internal/example/skills_content_test.go` per data-model.md "CapabilityClaim", mapping each Skill's own textual verification/validation promises (drawn from its "Validation Rules"/"Deterministic Operations" sections) to exactly one of: a real `internal/validation.Code*` constant (`BackingCode`, for structural-validation promises) or an entry already present in `knownInternalCommands` (`BackingOp`, for non-validation deterministic capabilities, e.g. `internal fingerprint`).
- [X] T026 [US2] Implement the capability-allowlist check in `internal/example/skills_content_test.go` (FR-006): for each Skill in `skillVerificationAllowlist`, assert every listed claim has exactly one of `BackingCode`/`BackingOp` set and that a `BackingCode` value is a real member of `validation`'s `Code*` set; fail naming the Skill and the unbacked or dual-backed claim (depends on T024, T025).

**Checkpoint**: User Stories 1 and 2 both work independently; a Skill citing a nonexistent command, a broken example, or an unbacked verification promise is now caught three distinct ways.

---

## Phase 5: User Story 3 - Comportamento consistente entre integrações representativas (Priority: P2)

**Goal**: The real `resolve → context/prepare → validate` chain a Skill documents completes correctly for each of three representative agent integrations covering three distinct install targets.

**Independent Test**: Run the smoke test after a Skill-content change; confirm each of the three adapters completes the chain or the failure names the specific adapter and step.

### Tests for User Story 3

> **Write this test FIRST, ensure it FAILS before the corresponding implementation task.**

- [X] T027 [P] [US3] Write `internal/example/skill_smoke_test.go`'s table per data-model.md "SmokeTestScenario", fixed to the three representative adapters from research.md #4: `claude-code` (`TargetPath` `.claude/skills`), `cursor-agent` (`.cursor/skills`), `copilot` (`.github/skills`), each with `CommandChain` `[resolve, context-or-prepare, validate]`. Expected to FAIL (no runner yet) until T028.

### Implementation for User Story 3

- [X] T028 [US3] Implement the smoke test runner in `internal/example/skill_smoke_test.go` per contracts §3 (FR-007): for each `SmokeTestScenario` row, install the (post-Phase-3) generated Skills into a temp project via that adapter's real `Install()` (`internal/agents/{claude,cursoragent,copilot}`), then invoke the real `misterspec internal resolve`/`context` (or `prepare`)/`validate` chain against a small fixture Spec/Task (reusing an existing `internal/example`/`internal/context` fixture pattern), asserting each step exits `0`, returns its documented JSON envelope shape, and that the installed Skill content actually landed at that adapter's own `TargetPath()`; a failing row names the adapter and the failing step (depends on T021 for final generated content, T027).
- [X] T029 [US3] Add a deliberate single-adapter-failure regression case (e.g. temporarily asserting a wrong `TargetPath` for one row) proving the failure is reported for that adapter+step only, without masking the other two rows' pass results (spec Acceptance Scenario US3.1) (depends on T028).

**Checkpoint**: User Stories 1, 2, and 3 are all independently functional.

---

## Phase 6: User Story 4 - Falhas e fallback do Context Engine ficam visíveis (Priority: P2)

**Goal**: A Context Engine fallback occurring during a task is recordable in the existing evaluation `RunRecord` format and distinguishable by comparison, with no new automatic/hidden telemetry.

**Independent Test**: Record two `TaskResult`s differing only in fallback count and confirm the existing comparison logic surfaces the difference.

### Tests for User Story 4

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T030 [P] [US4] Add `internal/eval/record_test.go` cases (write first): `Metrics{ContextFallbacks: 0}.Validate()` and `Metrics{ContextFallbacks: 2}.Validate()` return `nil`; `Metrics{ContextFallbacks: -1}.Validate()` returns an `ErrInvalidRunRecord`-wrapped error. Expected to FAIL until T031.
- [X] T032 [P] [US4] Add an `internal/eval/compare_test.go` case (write first) asserting that two `TaskResult`s identical except for `ContextFallbacks` produce a per-field diff entry via the existing comparison logic, with no new comparison code required. Expected to FAIL until T031 adds the field (compare logic itself needs no change per contracts §4).

### Implementation for User Story 4

- [X] T031 [US4] Add `ContextFallbacks int \`json:"context_fallbacks"\`` to `eval.Metrics` in `internal/eval/record.go` per data-model.md "Metrics (extended)" (a plain count, no `*_estimated` sibling, consistent with `Calls`/`ExtraReads`/`Rework`), and extend `Metrics.Validate()` to reject `ContextFallbacks < 0` (depends on T030).
- [X] T033 [US4] Add a `fallback-tolerance-note` fragment to `internal/skillgen/fragment.go` (Section: "Deterministic Operations") with Body extending today's identical sentence found in `mister-analyze/SKILL.md:114-116`, `mister-plan/SKILL.md:110-112`, `mister-tasks/SKILL.md:105-107` ("If this fails, proceed using this Skill's own Optional Context above instead — it is never a Failure Condition.") with an added instruction to record the occurrence as `context_fallbacks` in the task's own `RunRecord` once one exists, rather than silently absorbing the extra read (spec FR-008) (depends on T008, T031).
- [X] T034 [P] [US4] Update the `mister-analyze`, `mister-plan`, and `mister-tasks` manifests (T011, T009, T010) to reference the new `fallback-tolerance-note` fragment instead of their current inline `Bespoke` text for that sentence (depends on T033).
- [X] T035 [P] [US4] Update `mister-wrap-up`'s equivalent "Required Context" fallback sentence (`SKILL.md:129-130`) and `mister-implement`'s own "This Skill remains free to read further repository files… whenever the pack alone is insufficient" sentence with the same fallback-recording instruction, as edited `Bespoke` Parts in their manifests (T012, T017) — kept bespoke rather than fragmented, since their exact wording differs from the three-Skill fragment (depends on T033, T012, T017).
- [X] T036 [US4] Re-run the regeneration entrypoint (T020) and confirm `skillgen_drift_test.go` (T007) still passes with the fallback-instruction change applied consistently to all five touched Skills (depends on T033, T034, T035).

**Checkpoint**: All four user stories are independently functional; fallback occurrences are now a recordable, comparable `RunRecord` figure instead of a silently-absorbed extra read.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Verify the feature's overall, cross-story success criteria and leave the repository consistent.

- [X] T037 **Amended, not met as originally written** — implemented `internal/example/skills_size_test.go` (`TestSkillsContent_SizeDoesNotRegressUnexpectedly`) as a non-regression guard rather than the 80%-of-baseline assertion this task originally specified. The 20% reduction target (SC-001) was found unachievable by User Story 1's fragment-consolidation approach (deduplication happens in `internal/skillgen/fragment.go`'s authoring source, not in what an agent actually reads — each generated `SKILL.md` stays full, byte-identical Markdown) without cutting real Skill content against FR-010's quality-preservation requirement. Actual corpus: 2393 lines vs. 2378 baseline (+0.6%, driven by User Story 4's fallback-recording text). See spec.md SC-001's amendment note for the recorded decision; a real size-reduction target is deferred to a future Spec.
- [X] T038 [P] Manually walk through every command in `quickstart.md` (§1-§5) and confirm actual output matches each section's "Expected" description, updating quickstart.md if any wording drifted from the final implementation.
- [X] T039 Run `go test ./...` for the full repository and resolve any regression before considering the feature complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on other stories.
- **User Story 2 (Phase 4)**: Depends on Foundational only for its own tests/checks (it does not read `internal/skillgen` output) — independently implementable in parallel with Phase 3, though its later re-verification is most meaningful once Phase 3's trims (T018) have landed.
- **User Story 3 (Phase 5)**: Depends on Foundational; its smoke-test fixture should run against Phase 3's final generated Skill content (T021), so schedule after Phase 3 even though the test code itself could be written earlier.
- **User Story 4 (Phase 6)**: Depends on Foundational and on Phase 3's fragment/manifest infrastructure (T008-T019) to add its own fragment; independent of Phases 4-5's content.
- **Polish (Phase 7)**: Depends on Phases 3-6 all being complete (SC-001's size check is a whole-corpus measurement).

### Within Each User Story

- Tests written and failing before their corresponding implementation task (Principle V).
- Fragment/manifest definitions before the manifests that reference them.
- Manifests before `Manifests` population, generation, and the drift-check.

### Parallel Opportunities

- T003, T005, T006 (Phase 2) run in parallel once T002 lands.
- T009-T017 (Phase 3 manifests) run in parallel once T008 lands — different Skills, independent edits.
- T022/T024 (Phase 4 negative-fixture tests) run in parallel; T025 runs in parallel with T023.
- Phase 4 (US2) and Phase 3 (US1) can be staffed in parallel after Phase 2, since US2's checks operate on whatever Skill content currently exists and are re-run regardless of Phase 3's trims.
- T034/T035 (Phase 6 manifest updates) run in parallel once T033 lands.

---

## Parallel Example: User Story 1

```bash
# After T008 (KnownFragments) lands, build all 9 remaining manifests together:
Task: "Build SkillManifest for mister-plan in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-tasks in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-analyze in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-wrap-up in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-features in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-constitution in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-specify in internal/skillgen/manifest.go"
Task: "Build SkillManifest for mister-program in internal/skillgen/manifest.go"
Task: "Build SkillManifests for mister-implement and mister-knowledge-base in internal/skillgen/manifest.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run `internal/example/skillgen_drift_test.go`; edit a fragment and confirm the change propagates to every referencing Skill on regeneration (quickstart.md §1).

### Incremental Delivery

1. Setup + Foundational → engine ready.
2. User Story 1 → canonical fragments in place, drift-checked (MVP).
3. User Story 2 → citation/example/capability checks extended.
4. User Story 3 → cross-integration smoke coverage.
5. User Story 4 → fallback visibility wired into the existing evaluation format.
6. Polish → whole-corpus size reduction confirmed, full suite green.

### Parallel Team Strategy

After Phase 2: one contributor takes Phase 3 (US1)'s manifest work, a second takes Phase 4 (US2)'s test-file checks (independent of Phase 3's specific trims), a third prepares Phase 5 (US3)'s smoke-test table and fixture — merging Phase 5's implementation once Phase 3's generated content is final. Phase 6 (US4) is best taken by whoever did Phase 3, since it edits the same fragment/manifest files.

---

## Notes

- [P] tasks touch different files or independent parts of the same generator input (fragments vs. per-Skill manifests) with no ordering dependency between them.
- Every "write the test first" task above is explicitly called out; per Principle V this is NON-NEGOTIABLE for this feature's entirely-deterministic scope.
- Commit after each task or logical group (e.g., all of one Skill's manifest + its regenerated file together).
- Stop at each Checkpoint to validate that story's Independent Test from spec.md before moving on.
- Avoid: hand-editing a generated `kit/skills/*/SKILL.md` without updating the source fragment/manifest and regenerating — T007's drift-check exists specifically to catch this.
