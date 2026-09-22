---

description: "Task list for Evidências de Execução e Validade por Fingerprint"
---

# Tasks: Evidências de Execução e Validade por Fingerprint

**Input**: Design documents from `/specs/041-task-evidence-fingerprint/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/task-evidence-contract.md, quickstart.md

**Tests**: Per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE), test tasks are included for every user story — this feature is entirely deterministic parsing/computation/validation logic, exactly what Principle V requires tests for.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1-US4)
- File paths are exact and repo-relative

## Path Conventions

Single Go project. New package `internal/evidence` (parsing/state only); widened `internal/vcs`, `internal/operations`, `internal/prepare`, `internal/validation`, `internal/cli/internalcmd`; generated `kit/skills/mister-implement/SKILL.md` (via `internal/skillgen`, Spec 039).

---

## Phase 1: Setup

**Purpose**: Scaffold the one new package this feature adds.

- [X] T001 Create `internal/evidence/doc.go` with package doc explaining `internal/evidence` parses a Task's own `Evidence-*:` body lines and derives its real completeness state (`EvidenceState`) — a development-time-scale package, mirroring `internal/prepare`'s own shape (plan.md "Scale/Scope").

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The Git-fact, fingerprint, and evidence-parsing primitives every user story builds on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundational

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T002 [P] Add cases to `internal/vcs/vcs_test.go` (write first): `HeadCommit` returns the current short SHA with `available == true` in a real repo with commits; `available == false`, no error, when root is not a Git repository (mirrors `CommitsSinceFileAdded`'s own convention). Expected to FAIL until T004.
- [X] T003 [P] Add cases to `internal/vcs/vcs_test.go` (write first): `IsWorkingTreeDirty` returns `false` in a freshly-committed repo with no local changes, `true` after writing an uncommitted file, and a non-nil error when root is not a Git repository. Expected to FAIL until T005.
- [X] T006 [P] Add cases to `internal/operations/fingerprint_test.go` (write first): `TaskContentFingerprint(task, body)` is pure (no I/O), deterministic (same `body` → same digest), and produces a different digest for a different `body`; the existing file-based `Fingerprint`'s own behavior is unchanged (regression case). Expected to FAIL until T007-T008.
- [X] T009 [P] Add `internal/evidence/fields_test.go` (write first): `ParseEvidenceFields` correctly extracts `Result`, `Origin`, `By`, `CapturedAt`, `Command`, `GitRevision`, `WorkingTree`, `Fingerprint`, `Log` when their respective `Evidence-*:` lines are present; every field is its zero value (`""`) when none of these lines exist at all — a Task authored before this feature exists is never treated as an error (data-model.md "EvidenceFields", spec Assumptions). Expected to FAIL until T010.
- [X] T011 [P] Add `internal/evidence/state_test.go` (write first), one case per data-model.md "EvidenceState" row in the documented precedence order: `Result == ""` → `Unverified`; `Result == "fail"` → `Failed` (even if the fingerprint would otherwise match); `Result == "pass"` with a mismatched `Fingerprint` → `Stale`; `Result == "pass"` with a matching `Fingerprint` → `Verified`. Expected to FAIL until T012.

### Implementation for Foundational

- [X] T004 Add `vcs.HeadCommit(root string) (sha string, available bool, err error)` to `internal/vcs/vcs.go` per contracts §1 (depends on T002).
- [X] T005 Add `vcs.IsWorkingTreeDirty(root string) (bool, error)` to `internal/vcs/vcs.go`, running `git status --porcelain` and reporting `true` iff it produces any output (research.md #7 — repository-wide, not path-scoped) (depends on T003).
- [X] T007 Extract the SHA-256-of-bytes core already inside `internal/operations/fingerprint.go`'s `Fingerprint` into an unexported `digestBytes(data []byte) FileFingerprint` helper, with `Fingerprint`'s own public behavior unchanged (depends on T006).
- [X] T008 Add `operations.TaskContentFingerprint(task ids.EntityID, body []byte) FileFingerprint` in `fingerprint.go`, built on `digestBytes` (contracts §2) (depends on T007).
- [X] T010 Implement `internal/evidence/fields.go`: the `EvidenceFields` struct and `ParseEvidenceFields(task ids.EntityID, body []byte) EvidenceFields`, one regex per `Evidence-*:` label, mirroring `internal/prepare/fields.go`'s exact idiom (contracts §3, data-model.md "EvidenceFields") (depends on T009).
- [X] T012 Implement `internal/evidence/state.go`: the `EvidenceState` type (`Unverified`/`Failed`/`Stale`/`Verified`) and `DeriveState(fields EvidenceFields, currentFingerprint string) EvidenceState`, applying data-model.md's exact precedence (contracts §3) (depends on T008, T011).

**Checkpoint**: Foundation ready — Git facts, Task-content fingerprints, and evidence-state derivation are all available; nothing downstream consumes them yet.

---

## Phase 3: User Story 1 - Concluir uma tarefa exige evidência, não só um checkbox (Priority: P1) 🎯 MVP

**Goal**: A Task's checkbox alone never means "complete" — completion requires an associated, non-failed evidence record; capturing valid (declared) evidence makes it genuinely complete.

**Independent Test**: Mark a Task's checkbox with no evidence and confirm it isn't treated as verified; capture declared evidence describing a passing result and confirm it becomes verified.

**Scope note**: This phase implements `CaptureEvidence`'s `--origin declared` path only. `--origin automated` (running a real command) is Phase 6 (User Story 4)'s own addition — `CaptureEvidence` rejects `--origin automated` with a clear, explicit error until then, so the command never silently does the wrong thing mid-build-out.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T013 [P] [US1] Add `internal/operations/evidence_test.go` (write first): `CaptureEvidence` with `Origin: "declared"`, `Result: "pass"` (and separately `"fail"`) returns `EvidenceFields` with `Task`, `Result`, `Origin`, `By`, `CapturedAt` (RFC 3339), `GitRevision`/`WorkingTree` (from the Foundational `vcs` helpers), and `Fingerprint` (`TaskContentFingerprint` of the Task's own current body) all populated, with `Command`/`Log` left empty; a request with `Origin: "declared"` and `Command` set is rejected with a descriptive error (contracts §2's "forbidden otherwise"); a request with `Origin: "declared"` and no `Result` is rejected. Expected to FAIL until T017.
- [X] T014 [P] [US1] Add cases to `internal/prepare/scan_test.go` (write first): a Task whose checkbox is checked but has no `Evidence-Result:` line reports `TaskInfo.Status == "pending"` (not `"complete"`, contrary to today's behavior); a Task whose checkbox is checked and carries `Evidence-Result: pass` with a `Evidence-Fingerprint:` matching its own current content reports `Status == "complete"` and `Evidence == evidence.Verified`. Expected to FAIL until T021.
- [X] T015 [P] [US1] Add `internal/validation/task_evidence_test.go` (write first): a checked Task with no `Evidence-*:` lines produces exactly one Finding with `Code == CodeUnverifiedTask`; a checked Task with `Evidence-Result: fail` produces exactly one Finding with `Code == CodeFailedTaskEvidence`; an unchecked Task produces neither code regardless of its evidence fields (data-model.md "Validation Codes"). Expected to FAIL until T019-T020.
- [X] T016 [P] [US1] Add `internal/cli/internalcmd/capture_evidence_test.go` (write first): the declared-origin golden-envelope shape from contracts §4's first example; an error envelope when `--origin declared` is combined with `--command`. Expected to FAIL until T018.

### Implementation for User Story 1

- [X] T017 [US1] Implement `operations.CaptureEvidence`'s declared-origin path in new `internal/operations/evidence.go` (contracts §2): validate `Origin`/`By` are always required and exactly one of `{Result}` (declared) / `{Command,...}` (automated) applies, per data-model.md's `EvidenceCaptureRequest`; for `Origin == "declared"`, compute `GitRevision`/`WorkingTree` (`vcs.HeadCommit`/`IsWorkingTreeDirty`) and `Fingerprint` (`operations.TaskContentFingerprint` against the named Task's current `Section.Body`, located via `prepare.ScanSpecTasks`), set `CapturedAt` to the current UTC time in RFC 3339, and return the populated `evidence.EvidenceFields`; for `Origin == "automated"`, return a clear, explicit "not yet implemented — see User Story 4" error (removed in T031) (depends on T004, T005, T008, T010, T013).
- [X] T018 [US1] Implement `internal/cli/internalcmd/capture_evidence.go`: `"internal capture-evidence SPEC-### --task TASK-###"` per contracts §4, wiring `--origin`/`--by`/`--result` (plus `--command`/`--args`/`--dir`/`--timeout`, accepted now but only meaningful once T031 lands), resolving the Spec via `operations.Resolve` and the Task via `prepare.ScanSpecTasks` exactly as `internal prepare` already does (depends on T016, T017).
- [X] T019 [US1] Add `CodeUnverifiedTask`, `CodeFailedTaskEvidence`, and `CodeStaleTaskEvidence` constants together to `internal/validation/findings.go` (all three added now since they share one const block; `CodeStaleTaskEvidence`'s own dispatch logic is added later, in User Story 3's T027) (depends on T015).
- [X] T020 [US1] Implement `internal/validation/task_evidence.go`'s `taskEvidenceFindings(root, cfg, spec, specDir) ([]Finding, error)`, raising `CodeUnverifiedTask`/`CodeFailedTaskEvidence` per data-model.md's table (the `Stale` branch is added in T027), and wire it into `ValidateProject`/`ValidateEntity` in `internal/validation/validator.go` at `specCoverageFindings`'s own two existing call sites (depends on T012, T019).
- [X] T021 [US1] In `internal/prepare/scan.go`, add `TaskInfo.Evidence evidence.EvidenceState`, computed via `evidence.ParseEvidenceFields` + `evidence.DeriveState` (against `operations.TaskContentFingerprint` of the Task's own current body) alongside `Coverage`/`Dependency`/`Fields`; redefine `taskStatus` (or its call site) so `Status == "complete"` additionally requires `Evidence == evidence.Verified` (data-model.md "TaskInfo (extended)") (depends on T010, T012, T014).

**Checkpoint**: User Story 1 is independently functional — a checked-but-unverified Task is correctly `"pending"`; declared evidence with a passing result makes it genuinely `"complete"`.

---

## Phase 4: User Story 2 - Dependências entre tarefas respeitam evidência real (Priority: P1)

**Goal**: A Task that depends on an upstream Task is only reported ready when that upstream Task's real evidence state is `Verified` — never merely because its checkbox is checked.

**Independent Test**: An upstream Task checked but unverified leaves its dependent not-ready; capturing valid evidence for the upstream Task makes the dependent ready.

### Tests for User Story 2

> **Write these tests FIRST.**

- [X] T022 [P] [US2] Add cases to `internal/prepare/readiness_test.go` (write first): a dependent Task is NOT reported ready while its upstream dependency is checked-but-`Unverified`; it becomes ready once the upstream dependency's `Status` is `"complete"` (i.e., `Verified`, per T021). Expected to FAIL until T021 lands (Phase 3) — if run after Phase 3, should already PASS with no further code change (research.md #1's own claim, verified here).
- [X] T023 [P] [US2] Add cases to `internal/prepare/selection_test.go` (write first): automatic Task selection offers a checked-but-`Unverified` Task again (it is not actually done) rather than skipping it as already complete. Same expectation as T022 — should already pass once Phase 3 is done, with no further code change.

### Implementation for User Story 2

- [X] T024 [US2] Verification task, no production code change expected (research.md #1): run T022/T023 against Phase 3's completed work and confirm both pass with zero changes to `readiness.go`/`selection.go`. If either genuinely requires a code change, that is a defect in Phase 3's T021 (the `Status` redefinition), to be fixed there — record which, and why, rather than adding new logic to `readiness.go`/`selection.go` themselves (depends on T021, T022, T023).

**Checkpoint**: User Stories 1 and 2 are both independently functional — the dependency-readiness gap this feature exists to close is confirmed closed with no new readiness/selection logic.

---

## Phase 5: User Story 3 - Mudar os insumos torna a evidência antiga desatualizada (Priority: P2)

**Goal**: Editing a Task's own verified content flags its evidence as stale; re-capturing evidence against the new content clears the staleness.

**Independent Test**: Capture evidence, edit the verified Task's own body text, confirm staleness is flagged; re-capture and confirm it clears.

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T025 [P] [US3] Add a case to `internal/validation/task_evidence_test.go` (write first): a checked Task with `Verified` evidence, whose body is then edited (so its current `TaskContentFingerprint` no longer matches the recorded `Evidence-Fingerprint:`), produces exactly one Finding with `Code == CodeStaleTaskEvidence`; re-capturing evidence (fresh `Fingerprint` matching current content) removes that Finding. Expected to FAIL until T027.
- [X] T026 [P] [US3] Add a case to `internal/prepare/scan_test.go` (write first): a Task whose `Evidence` is `evidence.Stale` reports `Status == "pending"`, consistent with T014's pattern — should already pass once Phase 3's T021 lands (no new code needed here beyond T027's validation wiring); this test documents and locks in that expectation explicitly.

### Implementation for User Story 3

- [X] T027 [US3] Extend `internal/validation/task_evidence.go`'s `taskEvidenceFindings` to also raise `CodeStaleTaskEvidence` per data-model.md's `Stale` row — the state itself is already computed by `evidence.DeriveState` (Foundational T012); this task only adds the third dispatch branch (depends on T012, T019, T025).

**Checkpoint**: User Stories 1, 2, and 3 are all independently functional — editing verified content is visibly caught, without any manual staleness check.

---

## Phase 6: User Story 4 - Execução automatizada é explícita, controlada e registra o Git real (Priority: P2)

**Goal**: `--origin automated` runs exactly the given command/args/dir/timeout — never Markdown text implicitly — and records the real Git revision and working-tree dirtiness, including local uncommitted changes.

**Independent Test**: Capture automated evidence with explicit command/args/dir/timeout and confirm only that command runs; make an uncommitted local change and confirm the capture reflects a dirty working tree, not just the commit SHA.

### Tests for User Story 4

> **Write these tests FIRST, ensure they FAIL before the corresponding implementation task.**

- [X] T028 [P] [US4] Add cases to `internal/operations/evidence_test.go` (write first): `Origin: "automated"` runs exactly the given `Command`/`Args` inside `Dir`, derives `Result` from the exit code (`0` → `"pass"`, non-zero → `"fail"`), writes the command's combined stdout+stderr to a new file under `<specDir>/evidence/`, and returns that path in `Log`; a command still running past `Timeout` is terminated and reported as `"fail"` (not left hanging); a `Dir` that would resolve outside the project root is rejected before any subprocess ever starts (Principle VIII, reusing `artifacts.RelativeWithinRoot`, contracts §2). Expected to FAIL until T031.
- [X] T029 [P] [US4] Add a case to `internal/operations/evidence_test.go` (write first, per quickstart.md §5): capturing automated evidence while an unrelated file in the repository has an uncommitted change yields `WorkingTree == "dirty"`; after reverting that change, a fresh capture yields `WorkingTree == "clean"`. Expected to FAIL until T031 (though `vcs.IsWorkingTreeDirty` itself already exists from Foundational — this proves `CaptureEvidence`'s automated path actually calls it at the right time).
- [X] T030 [P] [US4] Add cases to `internal/cli/internalcmd/capture_evidence_test.go` (write first): the automated-origin golden-envelope shape from contracts §4's second example; an error envelope for `--dir` escaping the project root; an error envelope for `--origin automated` given without `--command`. Expected to FAIL until T032.

### Implementation for User Story 4

- [X] T031 [US4] Extend `operations.CaptureEvidence` (`internal/operations/evidence.go`) with the real automated-origin path, removing T017's placeholder rejection: run `Command`+`Args` via `exec.CommandContext` bounded by `Timeout`, with `Dir` resolved and validated via `artifacts.RelativeWithinRoot`; capture combined output; derive `Result` from the exit code; write the log to `<specDir>/evidence/<Task>-<RFC3339-compact-timestamp>.log` via the existing `writeAtomic` helper (`internal/operations/atomic_write.go`); populate `Command`/`Log` (depends on T028, T029).
- [X] T032 [US4] Update `internal/cli/internalcmd/capture_evidence.go` to fully surface `--command`/`--args`/`--dir`/`--timeout` (flags already registered by T018) and the new error envelopes from contracts §4 (depends on T030, T031).

**Checkpoint**: All four user stories are independently functional.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Bring the framework's own documentation and Skills up to date with the capability this feature adds, and verify the whole repository end-to-end.

- [X] T033 [P] Update `internal/skillgen/manifests_data.go`'s `mister-implement` manifest (its Deterministic Operations / Procedure / Completion Contract `Bespoke` Parts) to name `internal capture-evidence SPEC-### --task TASK-###` as the deterministic operation backing its own existing prose — `kit/skills/mister-implement/SKILL.md` already says "mark it complete only once that evidence exists" and "its own completion checkbox and evidence for the Task just completed" (lines ~31-32, ~98-99, ~161-163) without any real capability behind that claim today. Regenerate via `go generate ./internal/skillgen/...` and confirm `internal/example/skillgen_drift_test.go` still passes (039-lean-skills-integration-contracts precedent — never hand-edit the generated `SKILL.md` directly).
- [X] T034 [P] Add `"capture-evidence"` to `internal/example/skills_content_test.go`'s `knownInternalCommands` and to `skillOperationsAllowlist["mister-implement"]` (039 precedent) so T033's updated Skill text passes the existing command-existence check; add a matching `CapabilityClaim` entry to `skills_capability_test.go`'s `skillVerificationAllowlist`, backing mister-implement's evidence-verification promise with `BackingOp: "capture-evidence"` (depends on T033).
- [X] T035 Update `docs/architecture-specification.md`'s "Task status should primarily use Markdown checkboxes rather than duplicate lifecycle metadata" note (around line 1336) to record this feature's amendment: the checkbox remains the visible marker, but real completion now additionally requires a valid, non-stale `Evidence-Result:` — a checkbox is necessary, no longer sufficient.
- [X] T036 [P] Manually walk through every command in `quickstart.md` (§1-§5) against a real built binary and confirm actual output matches each section's "Expected" description, updating quickstart.md if any wording drifted from the final implementation.
- [X] T037 Run `go build ./...`, `go vet ./...`, and `go test ./...` for the whole repository and resolve any regression before considering the feature complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: None.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on other stories. 🎯 MVP.
- **User Story 2 (Phase 4)**: Depends on Phase 3's T021 (the redefined `Status`) — its own tests are largely confirmation of an emergent property, not new logic.
- **User Story 3 (Phase 5)**: Depends on Foundational and on Phase 3's `taskEvidenceFindings`/`CodeStaleTaskEvidence` scaffolding (T019, T020).
- **User Story 4 (Phase 6)**: Depends on Phase 3's T017 (the placeholder it replaces) and T018 (the CLI command it extends).
- **Polish (Phase 7)**: Depends on Phases 3-6 all being complete.

### Within Each User Story

- Tests written and failing before their corresponding implementation task (Principle V).
- `internal/vcs`/`internal/operations`/`internal/evidence` (Foundational) before anything that consumes them.
- `CaptureEvidence`'s request validation before its declared-origin assembly before its automated-origin execution (US1 before US4, by design — research.md #4's own incremental split).
- `TaskInfo.Evidence`/redefined `Status` (US1) before any readiness/selection confirmation (US2) or staleness check (US3).

### Parallel Opportunities

- T002, T003, T006, T009, T011 (Foundational tests) run in parallel — different files.
- T013, T014, T015, T016 (Phase 3 tests) run in parallel.
- T022, T023 (Phase 4 tests) run in parallel.
- T025, T026 (Phase 5 tests) run in parallel.
- T028, T029, T030 (Phase 6 tests) run in parallel.
- T033, T034, T036 (Phase 7) run in parallel.

---

## Parallel Example: Foundational

```bash
# All five foundational test files can be written together:
Task: "Add HeadCommit cases to internal/vcs/vcs_test.go"
Task: "Add IsWorkingTreeDirty cases to internal/vcs/vcs_test.go"
Task: "Add TaskContentFingerprint cases to internal/operations/fingerprint_test.go"
Task: "Add ParseEvidenceFields cases to internal/evidence/fields_test.go"
Task: "Add DeriveState precedence cases to internal/evidence/state_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1 (declared-origin evidence only).
4. **STOP and VALIDATE**: run quickstart.md §1-§2 — a checked-but-unevidenced Task is never "complete"; declared evidence with a passing result makes it genuinely complete.

### Incremental Delivery

1. Foundational → Git facts, fingerprints, evidence parsing/state all available.
2. User Story 1 → checkbox-alone completion closed (MVP), declared evidence only.
3. User Story 2 → confirms dependency readiness inherits it for free.
4. User Story 3 → staleness detection on content change.
5. User Story 4 → automated execution, explicit and Git-state-aware.
6. Polish → Skills/docs updated, full suite green.

### Parallel Team Strategy

After Phase 2: one contributor takes Phase 3 (US1)'s declared-origin path + validation wiring, a second prepares Phase 6 (US4)'s automated-execution tests against Phase 3's placeholder rejection (able to write tests immediately, implement once T017/T018 land). Phase 4 (US2) and Phase 5 (US3) are each small enough for a third contributor to pick up sequentially right after Phase 3 completes.

---

## Notes

- [P] tasks touch different files with no ordering dependency between them.
- Every "write the test first" task above is explicitly called out; per Principle V this is NON-NEGOTIABLE for this feature's entirely-deterministic scope.
- Commit after each task or logical group.
- Stop at each Checkpoint to validate that story's Independent Test from spec.md before moving on.
- Avoid: having `CaptureEvidence` or the CLI command write into `tasks.md` itself (that stays `/mister-implement`'s own job, research.md #4); inferring a verification command from a Task's own `Verify:` prose automatically (spec FR-003 — always explicit); scoping `IsWorkingTreeDirty` to "relevant" files (research.md #7's deliberate, documented simplification — repo-wide is the correct default here).

## Post-Implementation Fixes (found by code review, not part of the original 37 tasks)

- [X] T038 Fix `internal/operations/inspect.go`'s `inspectTask` (Task metadata lookup behind `internal inspect`), which still derived Task `Status` from the checkbox alone via a raw line scan (`taskCheckboxStatus`) — a second, divergent definition of Task completion from `internal/prepare/scan.go`'s already-corrected one. A Task checked but `Unverified`/`Stale`/`Failed` would report `"complete"` via `internal inspect` while `internal prepare`/`internal validate` correctly reported it as not-ready/flagged, for the identical Task. Fixed by extracting the one real definition into `evidence.TaskStatus(checked bool, state EvidenceState) string` (the sole leaf package both `internal/prepare` and `internal/operations` can import without a cycle) and rewriting `inspectTask` to use `artifacts.ParseDocument` (so the Task's full `Section.Body` is available for evidence parsing/fingerprinting, not just its checkbox line) instead of the old bespoke `bufio.Scanner` walk. Updated `internal/operations/inspect_test.go`'s pre-existing fixtures (`TestInspect_TaskNarrowedMetadata`, `TestInspect_CompositeTaskReferenceScopedToOwningSpec`) to carry real, matching evidence — the same class of fixture fix `prepare_test.go`'s own `markTaskComplete` already needed.
- [X] T039 Fix `internal/vcs.HeadCommit`, which treated *any* failure of `git rev-parse --short HEAD` as "not available, no error" — masking a genuine Git failure (corrupted repo, permission error, git binary issue) as a normal, empty-repo absence, inconsistent with this same file's `CurrentBranch`, which propagates exec errors verbatim. Fixed by checking "no commits yet" first, by exit code alone (`git rev-parse --verify --quiet HEAD`, the same technique `EnsureBranch` already uses in this file) rather than matching the short-SHA command's own stderr text, which is Git-version-specific (verified empirically: git 2.43.0 prints "fatal: Needed a single revision" for this exact case, not the more commonly assumed "unknown revision" message) — every other failure now propagates as a real error.
- [X] T040 Fix `internal/operations.CaptureEvidence`'s request validation, which rejected a declared-origin request also setting `Command` ("forbidden otherwise") but had no symmetric check for an automated-origin request also setting `Result` — the caller-supplied `Result` was silently discarded and overwritten by the command's own exit-code-derived outcome instead of being flagged as a likely mistake. Added the missing symmetric validation, matching `ErrInvalidEvidenceRequest`'s existing pattern.
- [X] T041 Fix `internal/evidence.DeriveState`, which treated any `Result` value other than `""` or `"fail"` as a pass, instead of requiring it to equal exactly `"pass"` per data-model.md's own documented `EvidenceState` table. A hand-typed `Evidence-Result: PASS` (or any other typo/malformed value) whose `Fingerprint` happened to match current content — e.g. copy-pasted from a prior valid record — was silently reported `Verified`, unblocking dependents in `readiness.go`, changing `internal inspect`'s reported status, and suppressing all three validation Findings, since none of `scan.go`/`inspect.go`/`task_evidence.go` did any additional `Result`-value check of their own before calling `DeriveState`. Fixed at the single source (`DeriveState` itself, not each of its three callers) so an unrecognized `Result` value is always `Unverified`, never silently trusted as a pass. Added a regression case to `internal/evidence/state_test.go`'s existing `TestDeriveState_Precedence` table.
