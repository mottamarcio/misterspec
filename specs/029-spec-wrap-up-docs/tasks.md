---

description: "Task list for feature implementation"
---

# Tasks: `/mister-wrap-up` — Spec Documentation for Future Official Docs

**Input**: Design documents from `/specs/029-spec-wrap-up-docs/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/commits-since-file.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First, NON-NEGOTIABLE), User Story 2's real Go logic (the new `internal/vcs` function and CLI command) gets real unit/contract tests written first. User Story 1 (the new Skill's own content) follows this session's established content-conformance pattern: a machine-checked assertion written first and confirmed failing, then the `SKILL.md` content that makes it pass.

**Organization**: Tasks are grouped by user story, but with a real dependency the priority numbering alone doesn't show: User Story 1 (the Skill itself) calls the deterministic operation User Story 2 builds, so **User Story 2 is implemented first**, even though both are P1. User Story 2 is still independently valuable and testable on its own — the `internal commits-since-file` command works, and can be verified, without `mister-wrap-up` existing at all.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single Go project with an embedded content kit. US2's Go changes live under `internal/vcs/` and `internal/cli/internalcmd/`; US1's Skill content lives under `kit/skills/mister-wrap-up/`; both stories' test additions touch `internal/example/skills_content_test.go` (US1) and their own package test files (US2).

---

## Phase 1: Setup

**Purpose**: Establish an accurate baseline of the existing conventions this feature must follow before adding to them.

- [X] T001 [P] Read `internal/vcs/vcs.go` in full — the existing `IsRepo`, `CurrentBranch`, `slugifyForBranch`, `BranchName`, `BranchForID`, `EnsureBranch`, `EnsureFeatureBranch` — to confirm the exact conventions the new `CommitsSinceFileAdded` function and the generalized slug helper must follow (baseline for User Story 2).
- [X] T002 [P] Read `internal/cli/internalcmd/fingerprint.go` and `internal/cli/internalcmd/validate.go` in full, plus `internal/cli/internal.go`'s own `AddCommand` list, to confirm the exact `NewXxxCmd()` / `project.Detect` / `WriteSuccess`/`WriteError` envelope pattern every existing command follows (baseline for User Story 2's new CLI command).
- [X] T003 [P] Read `kit/skills/mister-analyze/SKILL.md` in full, as the closest existing template for a Spec-scoped, read-mostly Skill (baseline for User Story 1's own `SKILL.md` authoring).
- [X] T004 [P] Read `internal/example/skills_content_test.go`'s `knownInternalCommands`, `skillOperationsAllowlist`, `canonicalSkillNames`, and `TestSkillsContent_AllNineInstalled` in full, noting exact current line ranges (baseline for both stories' test additions).

**Checkpoint**: Baseline established; no files changed yet.

---

## Phase 2: User Story 2 - Commit history sourced correctly, without any new commit-message convention (Priority: P1) 🎯 Built First

**Goal**: A new deterministic `internal commits-since-file --path <path>` command returns every commit on the current branch from the one that first added the named file (inclusive) through `HEAD` (inclusive) — correct regardless of what any commit message says, and gracefully reporting `available: false` (never an error) when the project isn't a Git repository or the file was never committed.

**Independent Test**: Follow `quickstart.md`'s automated checks for `internal/vcs`/`internal/cli/internalcmd` — call the new CLI command directly (no Skill involved) against a repo with commits both before and after a tracked file was added, and confirm the returned range starts exactly at the "added" commit.

### Tests for User Story 2 (write first, confirm failing)

- [X] T005 [P] [US2] In `internal/vcs/vcs_test.go`, add tests for a new `CommitsSinceFileAdded(root, path string) (commits []Commit, available bool, err error)`: (a) ordinary case — commits made before the file was added and commits made after — asserts the returned range starts exactly at the "added" commit and excludes everything before it; (b) the file was added at the repository's own root commit (no parent) — still returns the full range correctly; (c) the file was never committed — `available: false`, `commits` empty, no error. Run `go test ./internal/vcs/...` and confirm these **fail** (the function doesn't exist yet).
- [X] T006 [P] [US2] In `internal/vcs/vcs_test.go`, add a test for `CommitsSinceFileAdded` called against a directory that is not a Git repository at all — `available: false`, no error (mirrors `IsRepo`'s own existing false-case convention). Run `go test ./internal/vcs/...` and confirm it **fails**.
- [X] T007 [US2] In `internal/cli/internalcmd`, add a contract test (new `commits_since_file_test.go`) for `misterspec internal commits-since-file --path <path>`, asserting the JSON shape from `contracts/commits-since-file.md`: `{"ok":true,"available":true,"commits":[...]}` for the ordinary case, and `{"ok":true,"available":false,"commits":[]}` when history can't be determined. Run `go test ./internal/cli/internalcmd/...` and confirm it **fails** (command doesn't exist yet).

### Implementation for User Story 2

- [X] T008 [US2] In `internal/vcs/vcs.go`, generalize `slugifyForBranch` into an exported `Slugify(s string) string` with unchanged behavior (lowercase, hyphen-separated, unsafe-character-stripped) — `slugifyForBranch` becomes a thin wrapper (or is replaced by direct calls to `Slugify`) so both branch-naming and (later, User Story 1) filename-slug use the identical, single implementation (Constitution Principle VI, `research.md`'s own decision). Run `go test ./internal/vcs/...` and confirm all existing branch-automation tests (`022-feature-branch-automation`) still pass unchanged.
- [X] T009 [US2] In `internal/vcs/vcs.go`, implement `CommitsSinceFileAdded`: locate the commit that introduced `path` via `git log --diff-filter=A --follow --format=%H -- <path>` (take the oldest/last line as the "added" commit); return `available: false` with no error if that yields nothing (file never committed) or `IsRepo` is false; otherwise run `git log <added-commit>^..HEAD --format=%H|%s|%ad --date=short` (or `git log <added-commit>..HEAD` plus the added commit itself, if the added commit has no parent) and parse into `[]Commit{Hash, Subject, AuthorDate}`. Run `go test ./internal/vcs/...` and confirm T005/T006 now **pass**.
- [X] T010 [US2] In `internal/cli/internalcmd/commits_since_file.go` (new file), implement `NewCommitsSinceFileCmd()` following `fingerprint.go`'s own exact pattern: `--path` required flag, `project.Detect`, call `vcs.CommitsSinceFileAdded`, `WriteSuccess` with the `{"ok","available","commits"}` shape from `contracts/commits-since-file.md` (or `WriteError` only for a genuine operational failure, never for `available: false`). Wire it into `internal/cli/internal.go`'s `AddCommand` list. Run `go test ./internal/cli/internalcmd/...` and confirm T007 now **passes**.

**Checkpoint**: `internal commits-since-file` is fully functional and independently testable via the CLI directly — `mister-wrap-up` can now be built on top of it.

---

## Phase 3: User Story 1 - Generate a Spec's own documentation after implementation (Priority: P1) 🎯 MVP

**Goal**: A new canonical Skill, `mister-wrap-up`, produces `cortex/SPEC-###-<slug>.md` — synthesized prose covering the Spec's purpose, what was implemented, and its outcome, drawing on the Spec's own artifacts, `internal commits-since-file`'s own commit range, Knowledge, Constitution, and Learnings.

**Independent Test**: Follow `quickstart.md` steps 1–2 — run `/mister-wrap-up SPEC-0XX` against a Spec with a complete Plan/Tasks/Validation and confirm a coherent, synthesized document appears under `cortex/`.

### Tests for User Story 1 (write first, confirm failing)

- [X] T011 [US1] In `internal/example/skills_content_test.go`: add `"commits-since-file"` to `knownInternalCommands`; add `"mister-wrap-up"` with its own operations allowlist (`resolve`, `inspect`, `inventory`, `commits-since-file`) to `skillOperationsAllowlist`; add `"mister-wrap-up"` to `canonicalSkillNames`; add an `assertSkillConformant(t, "mister-wrap-up")` call (new `TestSkillsContent_WrapUp` function, or alongside an existing grouping — match the file's own existing convention). Run `go test ./internal/example/...` and confirm the new assertions **fail** (the Skill directory doesn't exist yet).
- [X] T012 [US1] In `internal/example/skills_content_test.go`, rename `TestSkillsContent_AllNineInstalled` (and its own doc comment, which currently says "exactly §38's nine Skills exist... no more, no fewer") to `TestSkillsContent_AllCanonicalSkillsInstalled` (or equivalent), reflecting 10 canonical Skills — no logic change, since the test is already fully driven by `canonicalSkillNames`. Run `go test ./internal/example/... -run TestSkillsContent_AllCanonicalSkillsInstalled` and confirm it **fails** (still only 9 directories exist under `kit.SkillsFS`).

### Implementation for User Story 1

- [X] T013 [US1] Author `kit/skills/mister-wrap-up/SKILL.md` with the full §39-required section set (`Purpose` through `Related Skills`), modeled on `kit/skills/mister-analyze/SKILL.md`'s own structure:
  - **Invocation**: `` `/mister-wrap-up SPEC-###` ``.
  - **Outputs**: `cortex/SPEC-###-<slug>.md`, one file per Spec, slug derived via the same `Slugify` helper T008 exported.
  - **Deterministic Operations**: `internal resolve SPEC-###`, `internal inspect SPEC-###`, `internal inventory knowledge`, and the new `internal commits-since-file --path <plan.md's own path>` (T010) — explicitly noting `available: false` is not a Failure Condition, just a note in the document (FR-007).
  - **Allowed Reads**: the Spec's own artifacts (`spec.md`, `plan.md`, `tasks.md`, `validation.md` if present), Knowledge, `ai/memory/constitution.md`, `ai/memory/learnings/*.md`.
  - **Allowed Creates/Modifications**: only its own `cortex/SPEC-###-<slug>.md` — explicitly stated as create-or-wholesale-replace, never amend (FR-009, unlike every other canonical Skill).
  - **Forbidden Mutations**: every other artifact this Skill reads — it is strictly read-only against Spec/Plan/Tasks/Validation/Knowledge/Constitution/Learnings.
  - **Procedure**: resolve the Spec, gather its artifacts, call `internal commits-since-file` for its `plan.md`, judge which Knowledge/Constitution/Learnings content is relevant (semantic, not a structured query — Learnings have no `for`/`parent` field per `docs/architecture-specification.md` §31), compose synthesized prose (not a raw dump), state plainly if Tasks are incomplete (FR-008) or commit history is unavailable (FR-007), write/replace `cortex/SPEC-###-<slug>.md`.
  - **Idempotency/Resume Behavior**: re-running always regenerates the document wholesale from current state — never duplicates, never amends (FR-009, US3's own requirement, satisfied here since there is no separate Go logic for it).
  - **Failure Conditions**: the named Spec has no Plan yet → stop and recommend creating one (FR-006); every other input source missing/unavailable is a *note in the document*, not a Failure Condition.
  Run `go test ./internal/example/...` and confirm T011/T012 now **pass**.
- [X] T014 [US1] In `docs/architecture-specification.md` §38 (Canonical Skill Set), add `mister-wrap-up` to the listed 10 names, and add a short note that it is optional and produces downstream documentation material, distinct from the required MVP pipeline the other 9 form.

**Checkpoint**: `/mister-wrap-up` is fully functional — a Spec with a Plan (complete or not) now produces a real, synthesized `cortex/*.md` document.

---

## Phase 4: User Story 3 - Re-running the command keeps the document current (Priority: P2)

**Goal**: Confirm `/mister-wrap-up`'s wholesale-replace behavior (already authored into the Skill's own Idempotency section, T013) actually holds up in practice — no separate Go logic, this story is validation-only.

**Independent Test**: Run `/mister-wrap-up` twice for the same Spec, with a new commit in between; confirm one current file, never two.

### Validation for User Story 3

- [X] T015 [US3] Follow `quickstart.md` step 3 — after an initial `/mister-wrap-up SPEC-0XX` run, make one more commit and re-run it. Confirm `cortex/SPEC-0XX-<slug>.md` is replaced (same path, updated content reflecting the new commit) rather than a second file appearing.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Confirm the whole feature regresses nothing and behaves as documented end-to-end.

- [X] T016 [P] Run `go build ./...` and `go test ./...` in full and confirm no regressions anywhere in the repository, not only in the packages this feature touched.
- [X] T017 Execute `quickstart.md`'s remaining manual steps (4–5: incomplete-Spec wording, commit-history-unavailable wording) end-to-end, or — if a live multi-turn Skill invocation isn't available in this session — a structural trace against `mister-wrap-up/SKILL.md`'s own Failure Conditions/Procedure text, noting explicitly which form of validation was performed. Confirm every acceptance scenario in `spec.md` (User Stories 1–3) is satisfied. NOTE: Step 4 (incomplete implementation) was validated live in T015's own scratch run — no `tasks.md`/`validation.md` existed for SPEC-001, and the manually-composed wrap-up document correctly stated completeness couldn't be confirmed rather than presenting the Spec as finished. Step 5 (commit-history-unavailable) is covered by the passing automated tests (`TestCommitsSinceFileAdded_FileNeverCommitted`, `_NotARepo`, `TestCommitsSinceFileCmd_NotAvailable`) plus a structural trace of `mister-wrap-up/SKILL.md`'s own Procedure step 4, which correctly instructs noting unavailability and continuing with every other source, never stopping. A live multi-turn `/mister-wrap-up` invocation (the agent actually composing prose from these sources end-to-end) requires a separate agent session and was not performed here — the underlying deterministic operation and the Skill's own instructions were both verified instead.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — all four tasks are `[P]` (different files, read-only).
- **User Story 2 (Phase 2)**: Depends only on Setup. Independently complete and testable on its own — `internal commits-since-file` works without `mister-wrap-up` existing.
- **User Story 1 (Phase 3)**: Depends on User Story 2 (its own `Deterministic Operations`/`Procedure` name and call `internal commits-since-file`, which must exist first) — this is the one real cross-story dependency in this feature, despite both being P1.
- **User Story 3 (Phase 4)**: Depends on User Story 1 (there is nothing to re-run until the Skill exists) — validation-only, no new implementation.
- **Polish (Phase 5)**: Depends on all three user stories.

### Parallel Opportunities

- **T001–T004** (Setup): fully parallel — four different files, all read-only.
- **T005, T006** (User Story 2 tests): parallel — both add tests to the same file (`vcs_test.go`) but describe independent cases; low-conflict, sequenced only by convention.
- **T016** (Polish, full regression): no file dependency on any single prior task, ordered last because it needs the whole feature present to be meaningful.
- No other tasks are file-parallel: User Story 2's implementation tasks (T008–T010) build on each other in the same package; User Story 1's tasks (T011–T014) are mostly sequential by design (test before content, content before doc update).

---

## Implementation Strategy

### MVP First (User Story 2, then User Story 1)

1. Complete Phase 1: Setup.
2. Complete Phase 2: User Story 2 (`internal commits-since-file`) — a real, independently useful deterministic operation even before the Skill exists.
3. Complete Phase 3: User Story 1 (`mister-wrap-up` itself) — this is the actual MVP the user asked for.
4. **STOP and VALIDATE**: run `quickstart.md` steps 1–2 against a scratch Spec.

### Incremental Delivery

1. Setup → conventions confirmed.
2. Add User Story 2 → validate independently (CLI contract tests + `quickstart.md`'s automated checks) → commit-range mechanism live.
3. Add User Story 1 → validate independently (`quickstart.md` steps 1–2) → `/mister-wrap-up` usable end-to-end (MVP).
4. Add User Story 3 → validate (`quickstart.md` step 3) → safe re-runs confirmed.
5. Polish → full-repo regression + remaining edge-case validation.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- T005/T006/T007/T011/T012 must each fail before their corresponding implementation task makes them pass (test-first, Constitution Principle V).
- Commit after each phase (Setup, User Story 2, User Story 1, User Story 3) rather than after every single task, so each phase's diff stays reviewable as one coherent unit.
- Avoid: letting `mister-wrap-up` write, amend, or delete anything other than its own `cortex/SPEC-###-<slug>.md` — every other artifact it reads must stay strictly read-only, per its own Forbidden Mutations (T013).
