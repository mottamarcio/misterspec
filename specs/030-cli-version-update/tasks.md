---

description: "Task list for feature implementation"
---

# Tasks: `misterspec --version` and `misterspec --update`

**Input**: Design documents from `/specs/030-cli-version-update/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/version-update.md, quickstart.md

**Tests**: Per Constitution Principle V (Test-First, NON-NEGOTIABLE), every new piece of Go logic (`internal/selfupdate`'s version/semver/release/checksum/replace mechanics, `internal/installer`'s mode-aware `WriteAtomicFile`, `internal/cli`'s new flags) gets real tests written first and confirmed failing, per this project's own established convention.

**Organization**: Tasks are grouped by user story, with a real dependency: both stories share the `internal/selfupdate.Version` variable (Foundational), but User Story 2 (`--update`) is otherwise self-contained and does not depend on User Story 1's own `--version` flag wiring — the two can be built in either order after Foundational completes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Single Go project. New package: `internal/selfupdate/`. Extended: `internal/cli/root.go`, `internal/installer/filesystem.go` (+ its 4 existing callers), `internal/agents/record.go` (bonus fix, Polish), `.github/workflows/release.yml`.

---

## Phase 1: Setup

**Purpose**: Confirm the exact conventions this feature extends, before touching them.

- [X] T001 [P] Read `internal/cli/root.go` and `internal/cli/init.go` in full — the root command currently has no persistent flags and no `RunE`; `init` (a "public" command) already uses the `internalcmd.WriteSuccess`/`WriteError` JSON envelope, `builtin.Default()`, and `kit.SkillsFS` — the exact pattern `--update`'s Skill-refresh step reuses (baseline for both stories).
- [X] T002 [P] Read `internal/installer/filesystem.go`'s `WriteAtomicFile` and its 4 existing call sites (`internal/bootstrap/config.go`, `internal/agents/record.go`, `internal/bootstrap/scaffold.go`, `internal/installer/installer.go`) — baseline for adding a file-mode parameter without changing any existing caller's actual behavior (User Story 2).
- [X] T003 [P] Read `internal/bootstrap/bootstrap.go`'s `Bootstrap(targetDir, agentID string, registry *agents.Registry, skills fs.FS)` signature and `internal/agents/record.go`'s `InstallRecord`/`RecordInstall` — baseline for `--update`'s Skill-reinstall step, and for the `FrameworkVersion = "0.1.0"` hardcoded constant discovered there (`agents/record.go`'s own doc comment: "no version-detection machinery exists yet" — this feature's own `selfupdate.Version` is exactly that machinery; tracked as a Polish task, not required by any FR).
- [X] T004 [P] Read `.github/workflows/release.yml` in full — confirm the exact existing asset-naming step (`misterspec-<goos>-<goarch>[.exe]`) and where the `-ldflags` version-embedding and `SHA256SUMS`-generation steps need to be inserted.

**Checkpoint**: Baseline established; no files changed yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The one piece both stories share — the embedded version variable itself.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 In `internal/selfupdate/version_test.go` (new), add a test asserting a `Report()` function returns the literal `Version` value when set, and `"development build"` when `Version == ""`. Run `go test ./internal/selfupdate/...` and confirm it **fails** (package doesn't exist yet).
- [X] T006 Create `internal/selfupdate/version.go`: `var Version string` (empty by default, set via `-ldflags` at build time — research.md's own decision) and `func Report() string` returning `Version` or `"development build"`. Run `go test ./internal/selfupdate/...` and confirm T005 now **passes**.

**Checkpoint**: `internal/selfupdate.Version`/`Report()` exist; both User Stories can now proceed, in either order.

---

## Phase 3: User Story 1 - Check the installed version (Priority: P1)

**Goal**: `misterspec --version` prints the embedded release version (or "development build"), with no network access and no filesystem changes.

**Independent Test**: Follow `quickstart.md` steps 1–3 — build with and without `-ldflags`, confirm `--version` reports correctly in both cases, offline.

### Tests for User Story 1 (write first, confirm failing)

- [X] T007 [US1] In `internal/cli/root_test.go` (new or extended), add a test that builds the root command, sets `--version`, and asserts the JSON output matches `contracts/version-update.md`'s own `{"ok": true, "version": "..."}` shape (both the set-`Version` and empty-`Version`/"development build" cases, using `selfupdate.Version` as an overridable package variable for the test). Run `go test ./internal/cli/...` and confirm it **fails** (flag doesn't exist yet).

### Implementation for User Story 1

- [X] T008 [US1] In `internal/cli/root.go`, add a `--version` persistent bool flag to the root command and a root-level `RunE` (root has neither today) that, when `--version` is set, calls `internalcmd.WriteSuccess` with `{"version": selfupdate.Report()}` (FR-001, FR-002) and returns — no network call, no filesystem write anywhere in this path. Run `go test ./internal/cli/...` and confirm T007 now **passes**.
- [X] T009 [US1] In `.github/workflows/release.yml`'s existing `go build -o "${{ steps.asset.outputs.name }}" ./cmd/misterspec` step, add `-ldflags "-X github.com/mottamarcio/misterspec/internal/selfupdate.Version=${{ needs.<tag-job>.outputs.tag || github.ref_name }}"` (matching the workflow's own existing tag-resolution logic from its "Determine release tag" step) so every published binary has its own real version embedded.

**Checkpoint**: `misterspec --version` is fully functional and independently testable via `quickstart.md` steps 1–3.

---

## Phase 4: User Story 2 - Check for and apply an available update (Priority: P1) 🎯 MVP

**Goal**: `misterspec --update` checks GitHub, reports current vs. latest, requires confirmation, verifies a checksum, atomically replaces the binary, and — inside an already-initialized project — reinstalls that project's own Skills.

**Independent Test**: Follow `quickstart.md` steps 4–9 — up-to-date/available/declined/confirmed/checksum-mismatch/with-and-without-project paths.

### Tests for User Story 2 (write first, confirm failing)

- [X] T010 [P] [US2] In `internal/selfupdate/semver_test.go` (new), add tests for a `Newer(current, latest string) bool` helper using `golang.org/x/mod/semver`, covering this project's own real tag history ordering: `Newer("v1.1.5-alpha", "v1.2.0")` → `true`; `Newer("v1.2.0", "v1.2.0")` → `false`; `Newer("v1.2.0", "v1.1.5-alpha")` → `false`. Run `go test ./internal/selfupdate/...` and confirm it **fails** (function doesn't exist, dependency not added).
- [X] T011 [P] [US2] In `internal/selfupdate/release_test.go` (new), add tests for a `LatestRelease(ctx, httpClient) (Release, error)`-shaped function (exact signature is implementation's own choice) using an injected fake `http.RoundTripper` returning a canned GitHub API JSON response — asserting the parsed `TagName` and the correctly-selected asset name for a given `GOOS`/`GOARCH` (`misterspec-<goos>-<goarch>[.exe]`, matching `release.yml`'s own naming exactly), plus a case where no asset matches the current platform. Run `go test ./internal/selfupdate/...` and confirm it **fails**.
- [X] T012 [P] [US2] In `internal/selfupdate/checksum_test.go` (new), add tests for parsing a `SHA256SUMS`-format file (`contracts/version-update.md`'s own documented format) and verifying downloaded content against it — both the match and mismatch cases. Run `go test ./internal/selfupdate/...` and confirm it **fails**.
- [X] T013 [P] [US2] In `internal/selfupdate/replace_test.go` (new), add tests for the atomic binary-replace function: the Unix direct-rename path (verify the file at the target path has the new content and `0o755` permissions afterward), and the Windows rename-aside path (via an injectable "current OS" seam — research.md's own decision — asserting the old binary is renamed aside and the new one lands at the original path, without requiring an actual Windows runner). Run `go test ./internal/selfupdate/...` and confirm it **fails**.
- [X] T014 [US2] In `internal/installer/filesystem_test.go`, add a test asserting `WriteAtomicFile`'s extended signature (with a new `os.FileMode` parameter) writes a file with exactly the requested mode. Run `go test ./internal/installer/...` and confirm it **fails** (signature doesn't exist yet).
- [X] T015 [US2] In `internal/cli/root_test.go`, add tests for `--update`'s own JSON output shapes from `contracts/version-update.md` (up-to-date, declined, updated-with-skills-refreshed, updated-without-project, and each error shape: `checksum_mismatch`, `release_check_failed`, `no_compatible_asset`, `permission_denied`) — using injected fakes for the HTTP layer and the confirmation prompt so no test is interactive or hits the network. Run `go test ./internal/cli/...` and confirm it **fails**.

### Implementation for User Story 2

- [X] T016 [US2] Run `go get golang.org/x/mod` (research.md's own justified, minimal new dependency) and `go mod tidy`.
- [X] T017 [P] [US2] Implement `internal/selfupdate/semver.go`'s `Newer` function using `golang.org/x/mod/semver`. Run `go test ./internal/selfupdate/...` and confirm T010 now **passes**.
- [X] T018 [P] [US2] Implement `internal/selfupdate/release.go`: `GET https://api.github.com/repos/mottamarcio/misterspec/releases/latest` via an injected `*http.Client`, parse the JSON response into a `Release{TagName string, Assets []Asset}` shape, and select the asset matching the current `runtime.GOOS`/`runtime.GOARCH` per `release.yml`'s own naming convention — return `FR-012`'s own "no compatible asset" / "release check failed" conditions as distinct, identifiable outcomes (never a generic error only). Run `go test ./internal/selfupdate/...` and confirm T011 now **passes**.
- [X] T019 [P] [US2] Implement `internal/selfupdate/checksum.go`: download and parse a `SHA256SUMS`-format file, compute a downloaded binary's own SHA-256, and compare. Run `go test ./internal/selfupdate/...` and confirm T012 now **passes**.
- [X] T020 [US2] In `internal/installer/filesystem.go`, extend `WriteAtomicFile` with an `os.FileMode` parameter (`chmod` the temp file to the requested mode before the final rename); update all 4 existing call sites (`internal/bootstrap/config.go`, `internal/agents/record.go`, `internal/bootstrap/scaffold.go`, `internal/installer/installer.go`) to pass their current implicit default explicitly, preserving today's behavior byte-for-byte. Run `go test ./internal/installer/... ./internal/bootstrap/... ./internal/agents/...` and confirm T014 now **passes** and every pre-existing test in those three packages still passes unchanged.
- [X] T021 [US2] Implement `internal/selfupdate/replace.go`: given a verified temp binary path and the currently running executable's own path (`os.Executable()`), atomically replace it — `runtime.GOOS == "windows"`: rename the current exe aside to `<name>.old` first (best-effort delete of any pre-existing stale `.old` from a prior run), then rename the temp file into the original path; every other `GOOS`: a single `os.Rename` directly, reusing T020's mode-aware `WriteAtomicFile`-style recipe to land the temp file at `0o755` before the swap. Run `go test ./internal/selfupdate/...` and confirm T013 now **passes**.
- [X] T022 [US2] In `internal/cli/root.go`, add a `--update` persistent bool flag and extend the root `RunE` (T008) with the full `--update` flow: check current (`selfupdate.Report()`) vs. latest (`selfupdate.LatestRelease`) via `Newer`; if not newer, `WriteSuccess` the `status: "up_to_date"` shape and return (FR-004, no download, no prompt); if newer, print both versions and prompt for confirmation (reusing `internal/cli/init.go`'s own `isInteractiveTerminal` convention for consistency) — on decline, `WriteSuccess` the `status: "declined"` shape, nothing changed (FR-006); on confirm, download the selected asset to a temp file, verify its checksum (T019) — on mismatch, `WriteError` with `checksum_mismatch`, nothing changed (FR-008); on match, call `replace.go` (T021) to atomically replace the running binary (FR-009); then check for `.misterspec/install.json` in the current project (`project.Detect`, matching every other command's own resolution) — if present, call `bootstrap.Bootstrap` with its own recorded `agent_id`, `builtin.Default()`, and `kit.SkillsFS` to refresh Skills (FR-010, reusing `init`'s own exact path, not duplicating it); if absent, report `skills_refreshed: false` plainly, not a failure (FR-011). Map GitHub-check failures and no-compatible-asset outcomes (T018) to `WriteError`'s own `release_check_failed`/`no_compatible_asset` codes, and a replace-permission failure (T021) to `permission_denied` (FR-012, FR-013). Run `go test ./internal/cli/...` and confirm T015 now **passes**.

**Checkpoint**: `misterspec --update` is fully functional — checked, confirmed, verified, replaced, and (where applicable) Skills refreshed, exactly per `contracts/version-update.md`.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Close the related gap T003 surfaced, publish checksums, and confirm the whole feature regresses nothing.

- [X] T023 [P] In `.github/workflows/release.yml`'s `release` job, add a step generating a `SHA256SUMS` file (standard `sha256sum`-compatible format, one line per downloaded asset in `dist/`) and include it in the `files:` list `softprops/action-gh-release@v2` already uploads (`contracts/version-update.md`'s own documented format).
- [X] T024 [P] In `internal/agents/record.go`, replace the hardcoded `FrameworkVersion = "0.1.0"` constant with a reference to `selfupdate.Report()` (or `selfupdate.Version` directly, falling back consistently with `Report()`'s own "development build" convention), so `install.json`'s own `misterspec_version` field reflects the actual running binary's version instead of a permanently stale literal — a bonus fix this feature's own new version-reporting mechanism makes possible, not required by any FR but directly adjacent to what T005/T006 already built. Run `go test ./internal/agents/...` and confirm no existing test asserting the literal `"0.1.0"` breaks unexpectedly (update any that do, since the stale literal was the bug being fixed).
- [X] T025 [P] Run `go build ./...` and `go test ./...` in full and confirm no regressions anywhere in the repository, not only in the packages this feature touched.
- [X] T026 Execute `quickstart.md`'s manual validation steps end-to-end against a real local build (or, for the parts requiring an actually-published newer GitHub release, a structural trace against the implemented code plus the already-passing injected-fake tests, noting explicitly which form of validation was performed). Confirm every acceptance scenario in `spec.md` (User Stories 1–2) is satisfied.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — all four tasks are `[P]` (different files, read-only).
- **Foundational (Phase 2)**: Depends on Setup. Blocks both User Stories — both need `selfupdate.Version`/`Report()` to exist.
- **User Story 1 (Phase 3)**: Depends on Foundational only. Does not depend on User Story 2.
- **User Story 2 (Phase 4)**: Depends on Foundational only. Does not depend on User Story 1 — though both stories edit `internal/cli/root.go`'s own `RunE`, so in practice whichever is implemented second must merge into the other's already-edited function (a same-file, not a logical, dependency).
- **Polish (Phase 5)**: Depends on both User Stories (T024 specifically needs T006's `Report()` to exist; T025/T026 need the whole feature present to be meaningful).

### Parallel Opportunities

- **T001–T004** (Setup): fully parallel — four different files, all read-only.
- **T010–T013** (User Story 2 tests): fully parallel — four different new test files, independent concerns (semver, release lookup, checksum, replace).
- **T017–T019** (User Story 2 implementation, matching T010–T012's own tests): fully parallel — three different new files, no dependency between them.
- **T023, T024, T025** (Polish): parallel — three different files/concerns.
- Everything touching `internal/cli/root.go` (T008, T022) and `internal/installer/filesystem.go` + its 4 callers (T020) is inherently sequential by file, regardless of story.

---

## Parallel Example: User Story 2's Own Test-First Wave

```bash
# Four people, four independent new test files, all before any implementation:
Task: "internal/selfupdate/semver_test.go — Newer() covering v1.1.5-alpha vs v1.2.0"
Task: "internal/selfupdate/release_test.go — GitHub API parsing + asset selection"
Task: "internal/selfupdate/checksum_test.go — SHA256SUMS parsing + verification"
Task: "internal/selfupdate/replace_test.go — Unix rename vs Windows rename-aside"
```

---

## Implementation Strategy

### MVP First (User Story 2 Alone Is the Real MVP)

Unusually for this session, User Story 2 (`--update`) — not User Story 1 — is the feature's own actual point; `--version` is almost a prerequisite reporting mechanism `--update` itself needs internally. Either can still be built and shipped independently:

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (`selfupdate.Version`/`Report()`).
3. Complete Phase 4: User Story 2 (`--update`) — deliverable on its own even before Phase 3.
4. **STOP and VALIDATE**: run `quickstart.md` steps 4–9.

### Incremental Delivery

1. Setup + Foundational → `selfupdate.Report()` exists.
2. Add User Story 1 → validate independently (`quickstart.md` steps 1–3) → `--version` usable.
3. Add User Story 2 → validate independently (`quickstart.md` steps 4–9) → `--update` usable (the actual point of this feature).
4. Polish → checksums published, `install.json`'s stale version fixed, full-repo regression.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- T005, T007, T010–T015 must each fail before their corresponding implementation task makes them pass (test-first, Constitution Principle V).
- Commit after each phase (Setup, Foundational, User Story 1, User Story 2, Polish) rather than after every single task, so each phase's diff stays reviewable as one coherent unit.
- Avoid: a hard-coded `http.DefaultClient`/real network call anywhere in `internal/selfupdate`'s own tests (research.md's own explicit decision — always inject the transport); modifying `WriteAtomicFile`'s 4 existing callers' actual on-disk behavior while adding the mode parameter (T020) — only the explicit mode changes, nothing else.
