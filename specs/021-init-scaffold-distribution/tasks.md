---

description: "Task list template for feature implementation"
---

# Tasks: Init Scaffolding and Binary Distribution

**Input**: Design documents from `/specs/021-init-scaffold-distribution/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/init-outcome.md, quickstart.md

**Tests**: Included and REQUIRED for User Story 1 and User Story 2 (Constitution Principle V) — both touch real Go logic (`bootstrap.Bootstrap`, `scaffoldDirectories`). User Story 3 (a GitHub Actions workflow plus two documentation edits) has no Go test surface; it is verified by actually running the workflow (quickstart.md §5) — the same discipline this project already applied to `deploy-docs.yml`.

**Organization**: Tasks are grouped by user story. **User Story 1** and **User Story 2** both modify `bootstrap.BootstrapOutcome` and the TUI/CLI surfaces that render it — spec.md's own claim that all three stories are "independent enough to ship separately" holds at the *requirements* level, but implementation-wise User Story 2 builds directly on files User Story 1 already touched (research.md's own plan assumes this order). **User Story 3** is fully independent of both — a new workflow file and two documentation edits, touching no Go source at all.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US3)
- Every task names its exact file path

## Path Conventions

```text
internal/bootstrap/          # bootstrap.go, config.go (existing); scaffold.go (new)
internal/tui/                # model.go, update.go, view.go (existing)
internal/cli/                # init.go (existing)
.github/workflows/           # release.yml (new)
README.md, site/getting-started.html
```

---

## Phase 1: Setup

**Purpose**: No new dependency, no new directory to create before either story begins.

**Checkpoint**: Nothing to verify — proceed directly to User Story 1.

---

## Phase 2: User Story 1 - Initialized Projects Contain No Unused Files (Priority: P1) 🎯 MVP

**Goal**: `misterspec init` no longer materializes the framework's own unused `templates/` directory into a target project; `BootstrapOutcome`, the TUI, and the CLI's own JSON payload all stop referencing it.

**Independent Test**: Initialize a fresh project; confirm no `templates/` directory exists, and confirm creating a new entity/artifact afterward still works exactly as before.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T002-T005.

- [X] T001 [US1] Update `internal/bootstrap/bootstrap_test.go`: add a case asserting that after a successful `Bootstrap` call, `<targetDir>/templates` does **not** exist, and that `BootstrapOutcome` no longer has a `TemplateOutcomes` field (a compile-time assertion via the struct literal, since Go itself will refuse to compile a reference to a removed field). Confirm this fails against the current, unmodified source (the directory does exist today).

### Implementation for User Story 1

- [X] T002 [US1] In `internal/bootstrap/bootstrap.go`: remove the `installer.Install(targetDir, false)` call and the `TemplateOutcomes` field from `BootstrapOutcome` (research.md #1 — `internal/installer.Install`/`List` themselves are left untouched, still used by nothing after this change but not deleted). Depends on T001.
- [X] T003 [P] [US1] In `internal/tui/model.go` and `internal/tui/update.go`: remove `templates []resourceSummary` from `previewContent` and the `installer.List()` call from `buildPreview` (research.md #5) — leave the struct/function otherwise intact for User Story 2 to extend. Depends on T002.
- [X] T004 [P] [US1] In `internal/tui/view.go`: remove the "kit templates" line from `viewPreview` and the "kit resources installed" line from `viewSuccess`. Depends on T003.
- [X] T005 [P] [US1] In `internal/cli/init.go`: remove the `"templates": outcomesJSON(outcome.TemplateOutcomes)` line from the JSON success payload (research.md #4 — the key is dropped entirely, not left as an empty array). Depends on T002.
- [X] T006 [US1] Update `internal/tui/*_test.go` and `internal/cli/init_test.go`: remove/adjust any assertion referencing the old templates behavior, JSON key, or Preview/Success wording. Depends on T003, T004, T005.
- [X] T007 [US1] Regression checkpoint: `go build ./...` succeeds; `go test ./internal/bootstrap/... ./internal/tui/... ./internal/cli/...` green, including T001's own new assertion now passing.

**Checkpoint**: User Story 1 is independently complete and testable — a freshly initialized project contains no `templates/` directory, and every existing capability still works.

---

## Phase 3: User Story 2 - A New Project's Full Folder Structure Is Immediately Visible (Priority: P2)

**Goal**: `misterspec init` scaffolds every directory the project's own configuration names, with a `.gitkeep` in each genuinely-empty one, so the structure is committable from the first commit.

**Independent Test**: Initialize a fresh project before creating any artifact; confirm every configured directory exists and is committable as-is; confirm creating a real artifact inside one afterward still succeeds normally.

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T009-T012.

- [X] T008 [P] [US2] Create `internal/bootstrap/scaffold_test.go`: `scaffoldDirectories` creates all six directories (data-model.md's own fixed order) relative to a temp target; each genuinely-empty one contains a zero-byte `.gitkeep`; a directory that already contains real content is left untouched and gets no `.gitkeep` (FR-006); the returned path list matches the fixed order exactly; calling it twice in a row is idempotent (no error, no duplicate `.gitkeep`, no content disturbed). Confirm these fail (the function doesn't exist yet).

### Implementation for User Story 2

- [X] T009 [US2] Implement `scaffoldDirectories` in new file `internal/bootstrap/scaffold.go` (data-model.md): `os.MkdirAll` for each of the six directories from `project.Default*` constants, then `installer.WriteAtomicFile` a zero-byte `.gitkeep` into any that is genuinely empty afterward (`os.ReadDir` length check), returning the six relative paths in fixed order. Depends on T008.
- [X] T010 [US2] Wire `scaffoldDirectories` into `Bootstrap` (`internal/bootstrap/bootstrap.go`): call it right after `writeDefaultConfig` succeeds, populate a new `DirectoriesScaffolded []string` field on `BootstrapOutcome` with its result. Depends on T009, T002 (User Story 1's own field removal on the same struct).
- [X] T011 [P] [US2] In `internal/tui/model.go`/`update.go`: add `directories []string` to `previewContent` (replacing the slot User Story 1 emptied — research.md #5), computed in `buildPreview` directly from `project.Default*` constants (nothing exists on disk yet at Preview time). Depends on T010, T003.
- [X] T012 [P] [US2] In `internal/tui/view.go`: render the six directories in `viewPreview` (in place of the removed templates line) and "N directories scaffolded" in `viewSuccess`. Depends on T011, T004.
- [X] T013 [P] [US2] In `internal/cli/init.go`: add `"directories": outcome.DirectoriesScaffolded` to the JSON success payload (data-model.md's own shape). Depends on T010, T005.
- [X] T014 [US2] Update `internal/tui/*_test.go` and `internal/cli/init_test.go`: assert the new `directories`/`"directories"` content in Preview/Success/JSON output. Depends on T011, T012, T013.
- [X] T015 [US2] Regression checkpoint: `go build ./...` succeeds; `go test ./internal/bootstrap/... ./internal/tui/... ./internal/cli/...` green; run `quickstart.md` §1-3 manually against a real temp project, confirming the directory tree, `.gitkeep` files, `git add -A` staging them, and a subsequent real `internal create` still succeeding.

**Checkpoint**: User Story 2 is independently complete and testable — a freshly initialized project's own full directory hierarchy exists and is git-committable before any artifact is created.

---

## Phase 4: User Story 3 - Get the Binary Without Cloning the Repository (Priority: P3)

**Goal**: A pre-built `misterspec` binary is published per supported OS/architecture on every tagged release, discoverable at a stable URL, alongside (not replacing) the existing build-from-source instructions.

**Independent Test**: Without cloning the repository or installing Go, follow only the published download instructions and confirm a working binary results.

- [X] T016 [P] [US3] Create `.github/workflows/release.yml` (research.md #6): triggered on push of a tag matching `v*` and on `workflow_dispatch` with a tag input; a build matrix over `{linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64}` running `GOOS=<goos> GOARCH=<goarch> CGO_ENABLED=0 go build -o <asset-name> ./cmd/misterspec` (data-model.md's own naming table); attaches every resulting binary to the matching GitHub Release via `softprops/action-gh-release`.
- [X] T017 [P] [US3] Update `README.md`'s own Install section: add a "download a pre-built binary" option linking to `https://github.com/mottamarcio/misterspec/releases/latest` (research.md #7), directly alongside the existing `git clone && go build` instructions — neither replaces the other.
- [X] T018 [P] [US3] Update `site/getting-started.html`'s own Install section with the same addition, matching the README's own content and the site's existing visual style. Depends on T017 (same wording/structure, authored together for consistency).
- [ ] T019 [US3] Validate `release.yml` for real: trigger it via `workflow_dispatch` against the already-existing `v1.0.0-alpha` tag; confirm all 5 binaries attach to that release, download at least one, `chmod +x` it, and confirm `--help`/`init --agent claude-code` behave identically to a local source build at the same commit (FR-008). Depends on T016.

**Checkpoint**: User Story 3 is independently complete and testable — a real, tagged release has five downloadable binaries, verified to actually work, entirely independent of User Story 1/2's own Go changes.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency check across all three stories.

- [X] T020 [P] Run `gofmt -l` and `go vet ./...` across the entire module and fix any findings.
- [X] T021 Full regression run: `go test ./...` (001 through 021) green, `go build ./cmd/misterspec` succeeds.
- [X] T022 Reconcile `specs/021-init-scaffold-distribution/contracts/init-outcome.md` against the actual implementation; fix any drift.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Nothing to do.
- **User Story 1 (Phase 2)**: No dependency — this feature's own MVP.
- **User Story 2 (Phase 3)**: Depends on User Story 1's own file changes to `bootstrap.go`/`model.go`/`update.go`/`view.go`/`init.go` (same structs, same functions — research.md's own planned order).
- **User Story 3 (Phase 4)**: No dependency on User Story 1 or 2 — entirely separate files. May proceed in parallel with either.
- **Polish (Phase 5)**: Depends on all three stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Independent — the true starting point.
- **User Story 2 (P2)**: Depends on User Story 1 (same files, same struct).
- **User Story 3 (P3)**: Independent of both — can be done first, last, or in parallel with either.

### Within Each User Story

- User Story 1/2: tests written and failing before implementation (Constitution Principle V).
- User Story 3: the workflow file itself (T016) before its own real-world validation (T019); README/site edits (T017/T018) are independent of the workflow file's own content.

### Parallel Opportunities

- T003/T005 (User Story 1's TUI and CLI edits) can proceed in parallel once T002 lands.
- T011/T013 (User Story 2's TUI and CLI edits) can proceed in parallel once T010 lands.
- User Story 3 (Phase 4) as a whole can proceed in parallel with User Story 1/2 (Phase 2/3) — zero shared files.
- Within Polish: T020 alone is parallel-safe; T021/T022 are sequential checks.

---

## Parallel Example: User Story 3 (fully independent of 1/2)

```bash
# Can run at any point, in parallel with User Story 1/2's own work:
Task: "Create .github/workflows/release.yml"
Task: "Update README.md's own Install section"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: User Story 1.
2. **STOP and VALIDATE**: `go test ./internal/bootstrap/... ./internal/tui/... ./internal/cli/...` green; a fresh `init` produces no `templates/` directory.

### Incremental Delivery

1. User Story 1 → the unused directory is gone (MVP).
2. User Story 2 (extends the same files) → the full hierarchy is scaffolded with `.gitkeep`.
3. User Story 3 (independent) → pre-built binaries published and verified working.
4. Polish (Phase 5).

### Team Strategy

User Story 3 has zero file overlap with User Story 1/2 and can be
worked on by a different person entirely, in parallel, from the very
start. User Story 1 → User Story 2 is the one genuinely sequential
chain in this feature.

---

## Notes

- This feature's implementation and review stay on `dev` — no PR to `stg` or `main` is part of this task list (research.md #8, explicit user instruction); the user will validate the resulting build on `dev` before any further promotion.
- `internal/installer.Install`/`List` are never deleted — only their one caller (`bootstrap`/`tui`) stops using them (research.md #1).
- Every directory `scaffoldDirectories` creates is one of `project.Configuration`'s own already-validated default paths — no new path-safety surface is introduced (Constitution Principle VIII).
- `"templates"` is removed from `init`'s own JSON payload, not left as an empty array (research.md #4) — a real, deliberate contract change while still pre-1.0/alpha.
- Commit after each phase; User Story 3 can be committed independently of User Story 1/2's own sequence.
