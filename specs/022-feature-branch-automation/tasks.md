---

description: "Task list template for feature implementation"
---

# Tasks: Feature-Level Git Branch Automation

**Input**: Design documents from `/specs/022-feature-branch-automation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/create-git.md, quickstart.md

**Tests**: Included and REQUIRED (Constitution Principle V, NON-NEGOTIABLE) — every change here touches deterministic logic (`internal/vcs`, `operations.Create`).

**Organization**: Tasks are grouped by user story. The new `internal/vcs` package and the `project.Configuration` toggle are genuinely shared prerequisites for all three stories (every story's own behavior is gated by "is this a Git repo" and "is automation enabled"), so they sit in Foundational rather than being duplicated per story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US3)
- Every task names its exact file path

## Path Conventions

```text
internal/vcs/                # NEW package
internal/operations/         # create.go (existing, modified)
internal/project/            # config.go (existing, modified)
internal/cli/internalcmd/    # create.go (existing, modified)
specs/022-feature-branch-automation/contracts/create-git.md
```

---

## Phase 1: Setup

**Purpose**: No new dependency (stdlib `os/exec` only) — nothing to initialize before Foundational work begins.

**Checkpoint**: Nothing to verify — proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The `internal/vcs` package and the `git_branch_automation` config toggle are used by every user story below — none of them is independently meaningful without both existing first.

**⚠️ CRITICAL**: No user story's tasks can start until this phase is complete.

### Tests for Foundational

> Write these tests FIRST; confirm they fail before implementing T002-T003.

- [X] T001 Create `internal/vcs/vcs_test.go`: using a fixture Git repository created via the real `git` CLI inside `t.TempDir()` (`git init`, `git -c user.email=... -c user.name=... commit --allow-empty -m init`), test `IsRepo` (true inside the fixture, false in a plain non-Git `t.TempDir()`), `BranchName("FEAT-007")` (pure function, no I/O, returns `"feat/FEAT-007"`), `CurrentBranch` (returns the fixture's initial branch name), and `EnsureBranch` (creates-and-checks-out a new branch, returns `created=true`; a second call with the same name checks it out without error and returns `created=false`; confirm via `git branch --show-current` after each call that the working tree actually moved). Confirm these fail (the package doesn't exist yet).

### Implementation for Foundational

- [X] T002 Implement `internal/vcs/vcs.go` (research.md #1-#5, data-model.md): `IsRepo`, `CurrentBranch`, `BranchName`, `EnsureBranch`, each shelling out to the system `git` binary via `os/exec.Command` (never through a shell), with the target project root as the command's working directory. `EnsureBranch` MUST check existence first (`git rev-parse --verify --quiet refs/heads/<name>`) and use plain `checkout <name>` to resume or `checkout -b <name>` to create — never `checkout -B` (research.md #4's non-destructive guarantee). Depends on T001.
- [X] T003 [P] In `internal/project/config.go`: add `GitBranchAutomation bool` (YAML `git_branch_automation`) to `Configuration` and to the pointer-based `configYAML` struct, defaulting to `true` when the key is absent from the file — following the exact pattern every other optional field in `Load` already uses (data-model.md). Add/update `internal/project/config_test.go` covering both "absent → true" and an explicit `false` in the YAML.
- [X] T004 Regression checkpoint: `go build ./...` succeeds; `go test ./internal/vcs/... ./internal/project/...` green, including T001's own new assertions now passing.

**Checkpoint**: `internal/vcs` and the config toggle exist and are independently tested — every user story below can now build on them.

---

## Phase 3: User Story 1 - A new Feature starts on its own branch (Priority: P1) 🎯 MVP

**Goal**: `misterspec internal create feature` creates-and-checks-out a dedicated branch (or resumes an existing one) before writing the Feature's own artifact, when the project is a Git repository and automation is enabled; degrades to exactly today's behavior otherwise.

**Independent Test**: In a Git-tracked project on branch `dev`, create a Feature; confirm a new branch now exists and is checked out, and the Feature's own artifact commit does not appear in `dev`'s own history.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T006-T008.

- [X] T005 [US1] In `internal/operations/create_test.go`, using a Git-fixture temp project (per T001's own fixture pattern): test that creating a Feature (a) creates and checks out a new branch named `feat/<new Feature ID>`, with `CreateResult.GitBranchCreated == true`; (b) a second Feature creation whose target branch name was pre-created manually resumes it (`GitBranchCreated == false`) instead of failing; (c) in a **non**-Git temp project, Feature creation succeeds unchanged with `GitSkippedReason == "not_a_git_repo"` and no `git` side effects; (d) with `GitBranchAutomation: false` in the loaded config, Feature creation succeeds unchanged with `GitSkippedReason == "disabled"`. Confirm these fail against the current, unmodified `Create`.

### Implementation for User Story 1

- [X] T006 [US1] In `internal/operations/create.go`: add `GitBranch`, `GitBranchCreated`, `GitSkippedReason` fields to `CreateResult` (data-model.md). For `Type == ids.Feature`, after the new ID is allocated and before the artifact file is written, check `proj.Config.GitBranchAutomation` and `vcs.IsRepo(root)`; when both hold, call `vcs.EnsureBranch(root, vcs.BranchName(newID.String()))` and populate `GitBranch`/`GitBranchCreated`; otherwise set `GitSkippedReason` to `"disabled"` or `"not_a_git_repo"` respectively. Depends on T002, T003, T005.
- [X] T007 [US1] In `internal/cli/internalcmd/create.go`: add a `"git"` object to the JSON success payload for `Type == feature`, shaped per `contracts/create-git.md` (`branch`+`created` when automation ran, `skipped_reason` when it didn't — mutually exclusive, never both). Depends on T006.
- [X] T008 [US1] Regression checkpoint: `go build ./...` succeeds; `go test ./internal/operations/... ./internal/cli/...` green, including T005's own new assertions now passing; run `quickstart.md` §1 manually against a real temp Git project.

**Checkpoint**: User Story 1 is independently complete and testable — creating a Feature isolates its own work onto a dedicated branch.

---

## Phase 4: User Story 2 - Work under a Feature stays on that Feature's branch (Priority: P2)

**Goal**: Creating a Spec never creates a new branch of its own; when the currently checked-out branch doesn't match its parent Feature's own branch, the response clearly flags the mismatch without blocking the creation or switching branches on the caller's behalf.

**Independent Test**: On a Feature's own branch, create a Spec under it — no new branch, no warning. Switch to another branch, create a Spec under the same Feature — creation still succeeds, response carries a mismatch warning.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T010-T011.

- [X] T009 [US2] In `internal/operations/create_test.go`: test that creating a Spec (a) while checked out on its parent Feature's own branch produces `GitBranch` equal to that branch and an empty `GitWarning`, with no new branch created (`git branch --show-current` unchanged before/after); (b) while checked out elsewhere (e.g. `dev`) produces a non-empty `GitWarning` naming both the current branch and the parent Feature's own branch, while still succeeding and leaving the current branch untouched; (c) in a non-Git project or with automation disabled, produces the same `GitSkippedReason` behavior as Feature creation (T005c/d), with no warning attempted. Confirm these fail against the current, unmodified `Create`.

### Implementation for User Story 2

- [X] T010 [US2] In `internal/operations/create.go`: add a `GitWarning` field to `CreateResult`. For `Type == ids.Spec`, after resolving the validated parent Feature ID and before writing the artifact, when `proj.Config.GitBranchAutomation` and `vcs.IsRepo(root)` both hold, compute `expected := vcs.BranchName(parentFeatureID)`, read `actual, _ := vcs.CurrentBranch(root)`, set `GitBranch = expected`, and set `GitWarning` only when `actual != "" && actual != expected` — never call `vcs.EnsureBranch` for Spec creation (no branch is ever created here). Otherwise set `GitSkippedReason` exactly as Feature creation does. Depends on T006, T009.
- [X] T011 [US2] In `internal/cli/internalcmd/create.go`: add the same `"git"` object to the JSON success payload for `Type == spec`, per `contracts/create-git.md` (`branch` alone, `branch`+`warning`, or `skipped_reason` — never `created`, which is Feature-only). Depends on T007, T010.
- [X] T012 [US2] Regression checkpoint: `go build ./...` succeeds; `go test ./internal/operations/... ./internal/cli/...` green, including T009's own new assertions now passing; run `quickstart.md` §2-3 manually.

**Checkpoint**: User Story 2 is independently complete and testable — Specs never fragment a Feature's work across branches, and a mismatch is always visible, never silent.

---

## Phase 5: User Story 3 - Teams that manage branches themselves can opt out (Priority: P3)

**Goal**: A project can fully disable this feature via a single configuration change, with Feature/Spec creation then behaving exactly as it did before this feature existed.

**Independent Test**: Set `git_branch_automation: false`; create a Feature and a Spec; confirm no branch is created, no warning is produced, and both artifacts land on whatever branch was already checked out.

> Note: the `disabled` code path itself was already implemented and tested per-story in T006/T009 (Foundational's own config toggle, T003, gates both). This phase's own tasks close the remaining gap: an explicit, end-to-end test proving the *opt-out* experience — not just the individual `GitSkippedReason` value — matches today's pre-feature behavior exactly.

### Tests for User Story 3

> Write this test FIRST; confirm it fails only in the sense that the behavior it checks didn't exist as an explicit guarantee before T006/T009 landed (it should already pass mechanically once Phase 3/4 are done — this task exists to pin the guarantee down as its own regression test, not to drive new production code).

- [X] T013 [US3] In `internal/operations/create_test.go`: an end-to-end test that, with `GitBranchAutomation: false`, creates a Feature and then a Spec under it in the same Git-tracked fixture project, and asserts the resulting artifact paths/IDs are identical to what the exact same two calls would have produced before this feature existed (i.e. no `git` side effects at all: `git branch --show-current` before and after both calls is unchanged, no new local branches exist per `git branch --list`).

### Implementation for User Story 3

- [X] T014 [US3] No production code change expected — this phase should already pass given T006/T009's own `proj.Config.GitBranchAutomation` gate. If T013 fails, fix `internal/operations/create.go` so the gate is checked before any `vcs` call for both Feature and Spec paths (not just before `EnsureBranch`/`CurrentBranch` individually).
- [X] T015 [US3] Regression checkpoint: `go test ./internal/operations/...` green including T013; run `quickstart.md` §6 manually.

**Checkpoint**: User Story 3 is independently complete and testable — opting out is a single config change with zero remaining Git side effects.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency check across all three stories.

- [X] T016 [P] Run `gofmt -l` and `go vet ./...` across the entire module and fix any findings.
- [X] T017 Full regression run: `go test ./...` (001 through 022) green, `go build ./cmd/misterspec` succeeds.
- [X] T018 Reconcile `specs/022-feature-branch-automation/contracts/create-git.md` against the actual implementation; fix any drift.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do.
- **Foundational (Phase 2)**: No dependency — blocks every user story below.
- **User Story 1 (Phase 3)**: Depends on Foundational (T002, T003).
- **User Story 2 (Phase 4)**: Depends on Foundational (T002, T003) and User Story 1's own `CreateResult`/JSON payload additions (T006, T007) — same struct, same function, same file.
- **User Story 3 (Phase 5)**: Depends on User Story 1 and 2's own gate logic (T006, T009) already existing to verify against.
- **Polish (Phase 6)**: Depends on all three stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational — the true starting point beyond Phase 2.
- **User Story 2 (P2)**: Depends on User Story 1 (same `CreateResult`/JSON payload).
- **User Story 3 (P3)**: Depends on User Story 1 and 2 (verifies their own gate, adds no new gate of its own).

### Within Each User Story

- Every story: tests written and failing before implementation (Constitution Principle V).

### Parallel Opportunities

- T001 (vcs tests) and T003 (config field) can proceed in parallel — different files, no shared dependency.
- Within Polish: T016 alone is parallel-safe; T017/T018 are sequential checks.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational.
2. Complete Phase 3: User Story 1.
3. **STOP and VALIDATE**: `go test ./internal/vcs/... ./internal/operations/... ./internal/cli/...` green; a real `internal create feature` in a Git-tracked temp project lands on a new branch.

### Incremental Delivery

1. Foundational → `internal/vcs` and the config toggle exist, independently tested.
2. User Story 1 → Feature creation gets its own branch (MVP).
3. User Story 2 → Spec creation stays on the Feature's branch, warns on mismatch.
4. User Story 3 → the opt-out is pinned down as its own explicit regression test.
5. Polish (Phase 6).

---

## Phase 7: Amendment — Readable Branch Names via an Optional Slug

**Trigger**: post-implementation validation found `feat/FEAT-001` hard to
recognize at a glance in `git branch --list` (research.md #3, amended).

- [X] T019 In `internal/vcs/vcs_test.go`: add tests for the amended
  `BranchName(featureID, slug string) string` (empty slug → bare ID;
  free-text slug → slugified and appended; unusable slug → falls back
  to bare ID), plus new `BranchForID` (finds a branch by ID prefix
  regardless of slug; returns `""` when none exists; does not
  false-positive-match a different, longer ID) and `EnsureFeatureBranch`
  (creates using the slug when no branch exists yet for that ID; resumes
  the existing branch, ignoring a differently-slugged request, when one
  already does).
- [X] T020 In `internal/vcs/vcs.go`: implement the amended `BranchName`
  (sanitizing free text via a conservative alphanumeric+hyphen
  allow-list), `BranchForID` (lists local branches, matches by exact ID
  or `<ID>-` prefix), and `EnsureFeatureBranch` (looks up via
  `BranchForID` first, only computing a new name via `BranchName` when
  none exists). In `internal/operations/create.go`: thread `req.Slug`
  through to `ensureFeatureBranch` (Feature creation) unvalidated (no
  path-safety check needed — `vcs.BranchName` sanitizes regardless of
  input); change `checkSpecBranch` (Spec creation) to look up the parent
  Feature's actual branch via `vcs.BranchForID` instead of recomputing a
  name, so mismatch detection keeps working correctly regardless of
  what slug (if any) the Feature was created with. In
  `internal/cli/internalcmd/create.go`: update the `--slug` flag's help
  text to note it's optional for `feature`.
- [X] T021 Regression checkpoint: `go build ./...`, `gofmt -l .`,
  `go vet ./...`, `go test ./...` all clean; manually create a Feature
  with `--slug "Git Branch Automation"` against a real Git-tracked temp
  project and confirm the resulting branch reads
  `feat/FEAT-001-git-branch-automation`.

**Checkpoint**: Branch names are readable at a glance without weakening
any of User Story 1/2/3's own guarantees — uniqueness still comes from
the ID alone, lookup still requires no new persisted state.

---

## Notes

- Every Git operation this feature performs is non-destructive: `rev-parse`, `branch --show-current`, and plain `checkout`/`checkout -b` only — never `reset --hard`, `push --force`, `checkout -B`, or branch deletion (Constitution Principle VIII, research.md #4).
- `internal/vcs` shells out to the system `git` binary via `os/exec` — no new Go module dependency (research.md #1).
- This feature's scope is `misterspec internal create`'s own `feature`/`spec` paths only — `program`, `task`, `knowledge`, and `learning` creation are untouched and gain no `git` key in their JSON payload at all.
- Commit after each phase.
