---

description: "Task list template for feature implementation"
---

# Tasks: Embedded Kit and Resource Installer

**Input**: Design documents from `/specs/005-embedded-kit/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/kit-installer.md, quickstart.md

**Tests**: Included and REQUIRED for every user story, per the project constitution (Principle V, Test-First Discipline, NON-NEGOTIABLE). 003-entity-creation's full test suite is an explicit, named regression gate — not just "the usual suite" — since this feature relocates content 003 depends on.

**Organization**: Tasks are grouped by user story, but this feature's real shape doesn't map cleanly onto spec.md's P1/P2/P3 priority order, and that's called out here rather than forced into a false uniform pattern. **The template relocation spec.md frames as User Story 3's Independent Test ("re-run 003's suite, confirm it passes unmodified") is architecturally the Foundational prerequisite for User Story 1 and 2** — `internal/installer.List`/`Install` have nothing real to discover or materialize until `kit.TemplatesFS` exists with actual content. So the relocation *work* lives in Foundational; User Story 3's own phase here is the explicit, separately labeled regression-verification checkpoint spec.md's Independent Test describes. Separately, **User Story 2 depends on User Story 1** — `Install` calls `List` internally to enumerate what to materialize (contracts/kit-installer.md), the same "reuse the enumeration, don't reimplement it" reasoning 002-read-operations's `Children` used.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task names its exact file path

## Path Conventions

```text
kit/                    # NEW top-level package — Foundational
internal/templates/      # existing package — Foundational (modified)
internal/installer/       # NEW package — US1, US2
internal/example/          # existing package — extended in Polish
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the directory skeletons this feature adds.

- [X] T001 Create `kit/` (with a `templates/` subdirectory) and `internal/installer/` directories, each with a `doc.go` package-level doc comment (per plan.md's Project Structure), at the repository root.

**Checkpoint**: Module builds (`go build ./...`) with the new empty packages before Foundational work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Relocate 003-entity-creation's templates into the shared `kit/` root so `kit.TemplatesFS` has real, correct content — without this, `internal/installer`'s `List`/`Install` (User Story 1, 2) would have nothing genuine to discover or materialize, only an empty or fabricated embed.

**⚠️ CRITICAL**: No User Story 1 or 2 implementation may begin until this phase — including the build succeeding — is complete.

- [X] T002 Move the 8 `.tmpl` files from `internal/templates/files/` to `kit/templates/` (preserving content byte-for-byte — `git mv`, not a copy-and-hand-edit).
- [X] T003 Implement `kit/kit.go`: `//go:embed templates/*.tmpl` into an exported `TemplatesFS embed.FS`.
- [X] T004 Modify `internal/templates/templates.go` to read each template's content via `kit.TemplatesFS.ReadFile("templates/"+filename)` instead of its own private embed; remove the old `//go:embed files/*.tmpl` directive, the now-unused `templateFS` variable, and the emptied `internal/templates/files/` directory. `Render`'s exported signature and every `*Data` struct are unchanged.
- [X] T005 Run `go build ./...` and `go test ./internal/templates/...` and confirm `internal/templates`'s own existing test suite (`templates_test.go`, covering all 8 `Kind`s) passes unmodified after the relocation.

**Checkpoint**: Foundation ready — `kit.TemplatesFS` holds real, correct content; `internal/templates` renders identically to before (verified, not assumed); User Story 1 implementation can now begin.

---

## Phase 3: User Story 1 - Discover What the Embedded Kit Provides (Priority: P1) 🎯 MVP

**Goal**: List every resource the embedded kit contains — name, kind, and artifact type — without touching the filesystem.

**Independent Test**: Call `List()` against the compiled binary's embedded kit and confirm it returns exactly the known 8-template set, identically on repeated calls, with no fixture project or filesystem writes involved.

### Tests for User Story 1

> Write these tests FIRST; confirm they fail before implementing T007.

- [X] T006 [P] [US1] Unit tests for `List()` — returns exactly 8 resources; each has `Kind == "template"` and the `ArtifactType` correctly derived from its filename (e.g. `program.md.tmpl` → `"program"`); two calls return identical results; no test here touches any filesystem path outside what `go test` itself manages — in `internal/installer/installer_test.go`.

### Implementation for User Story 1

- [X] T007 [US1] Implement `Resource` and `List()` (enumerates `kit.TemplatesFS` via `fs.WalkDir` or `fs.Glob`, deriving each resource's `Kind`/`ArtifactType` from its filename) in `internal/installer/installer.go`. Depends on T003, T006.

**Checkpoint**: User Story 1 is independently complete and testable — `go test ./internal/installer/... -run TestList` passes on its own.

---

## Phase 4: User Story 2 - Install Embedded Resources Into a Target Directory (Priority: P2)

**Goal**: Materialize every kit resource into a target directory, one atomic write per resource, never silently overwriting.

**Independent Test**: Install into a fresh temporary directory (every file newly written, byte-for-byte identical); install again without overwrite into the now-populated directory (every resource `Skipped`, nothing changed); install with overwrite (everything replaced); attempt installing to a traversal-escaping target (rejected, nothing written). **Note**: `Install`'s implementation genuinely depends on User Story 1 (`List`) — it calls `List()` to enumerate what to materialize rather than re-deriving that enumeration itself (contracts/kit-installer.md).

### Tests for User Story 2

> Write these tests FIRST; confirm they fail before implementing T010–T011.

- [X] T008 [P] [US2] Filesystem-integration tests for `Install` — a fresh target directory (every `Outcome.Status == Installed`, every file byte-identical to its `kit.TemplatesFS` source); a fully-populated target directory with `overwrite: false` (every `Outcome.Status == Skipped`, nothing on disk changes); the same directory with `overwrite: true` (every resource replaced); a target whose computed destination would resolve outside the target directory (`Outcome.Status == Failed`, `Err` wraps `artifacts.ErrPathOutsideProject`, nothing written) — in `internal/installer/installer_test.go` (same file as T006; sequential, not `[P]` with it, since T006/T007 are already complete by the time this runs).
- [X] T009 [P] [US2] Unit tests for the atomic-write helper — a successful write leaves the exact expected content at the target path; a simulated failure partway through (e.g. writing to a read-only directory) leaves no file at the target path at all — in `internal/installer/filesystem_test.go`.

### Implementation for User Story 2

- [X] T010 [US2] Implement the atomic-write helper (temp file in the target directory → complete write → `fsync` → `os.Rename`) — this package's own small implementation, not imported from `internal/operations` (research.md) — in `internal/installer/filesystem.go`. Depends on T009.
- [X] T011 [US2] Implement `OutcomeStatus`, `Outcome`, and `Install(targetDir string, overwrite bool) ([]Outcome, error)` — calls `List()`, checks each destination via `artifacts.RelativeWithinRoot` before any write, applies skip/overwrite logic, delegates the actual write to `filesystem.go`'s helper — in `internal/installer/installer.go`. Depends on T007, T010, T008.

**Checkpoint**: User Stories 1 AND 2 both pass their own tests.

---

## Phase 5: User Story 3 - Templates Have One Source of Truth (Priority: P3)

**Goal**: Prove — not assume — that 003-entity-creation's `Create`/`CreateArtifact` produce unchanged output after Phase 2's relocation.

**Independent Test**: Re-run 003-entity-creation's existing test suite unmodified and confirm every test passes. **Note**: unlike every other story in this project so far, this story's actual *work* already happened in Foundational (T002–T005) — Phase 2 had to relocate the templates correctly for User Story 1/2 to have real content in the first place. This phase is the explicit, separately labeled verification checkpoint spec.md's own Independent Test describes for User Story 3, not new implementation.

### Verification for User Story 3

- [X] T012 [US3] Run `go test ./internal/operations/... ./internal/example/...` (covering `create_test.go`, `create_artifact_test.go`, `create_concurrency_test.go`, and `creation_quickstart_test.go`) and confirm every 003-entity-creation test passes unmodified — the concrete proof of FR-008/SC-005, not an assumption. Depends on T004.

**Checkpoint**: All three user stories verified — `kit.TemplatesFS` is genuinely templates' one source of truth, `List`/`Install` work against it, and 003's behavior is unchanged.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency and validation across all three stories, no new capability.

- [X] T013 [P] Run `gofmt -l` and `go vet ./...` across the entire module (001-core-foundation through 004-structural-validation's packages included) and fix any findings.
- [X] T014 [P] Verify/extend package-level doc comments on `kit` and `internal/installer`, cross-checked against `specs/005-embedded-kit/contracts/kit-installer.md`.
- [X] T015 Add a compiled, run-in-CI example in `internal/example` (extending the existing package) exercising `quickstart.md`'s flow end-to-end: `List`, `Install` into a fresh directory, `Install` again without overwrite (all skipped), `Install` with overwrite, and a `Create` call proving artifact creation still works from the same `kit`.
- [X] T016 Reconcile `specs/005-embedded-kit/contracts/kit-installer.md`'s signatures against the actual implementation; fix any drift introduced during implementation (same discipline as every prior feature's final reconciliation task).
- [X] T017 Full regression run: `go test ./...` across the entire module (001 through 005) green, `go vet ./...` clean, `gofmt -l .` empty.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS User Story 1 and 2, and *is* User Story 3's real work (see Organization note above).
- **User Story 1 (Phase 3)**: Depends on Foundational only.
- **User Story 2 (Phase 4)**: Depends on Foundational and on User Story 1's implementation (`Install` calls `List`).
- **User Story 3 (Phase 5)**: Its verification (T012) depends only on Foundational (T004) — it does not depend on User Story 1 or 2 at all, since it's proving 003's *own* behavior is unchanged, unrelated to the new `installer` package.
- **Polish (Phase 6)**: Depends on all three user stories being complete/verified.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependency on US2/US3.
- **User Story 2 (P2)**: Depends on User Story 1's implementation (T007) — not just Foundational.
- **User Story 3 (P3)**: Depends only on Foundational (T004) — genuinely independent of US1/US2's *implementation*, since it's re-verifying 003's own suite, not installer's new code.

### Within Each User Story

- Tests written and failing before implementation (Constitution Principle V).
- Foundational's relocation complete and its own regression check (T005) green before US1/US2 begin.
- US3's verification (T012) can, in principle, run any time after T004 — it's placed after US1/US2 here mainly for narrative clarity, not because it technically depends on them.

### Parallel Opportunities

- T002–T005 (Foundational) are a strict sequence — each depends on the previous (move files → embed them → repoint the reader → verify) — not parallelizable.
- Once Foundational is done: US3's verification (T012) can run in parallel with all of US1/US2's work, since it depends on nothing from `internal/installer`.
- Within US2: T008 and T009 (tests, different files) in parallel.
- Within Polish: T013 and T014 in parallel.

---

## Parallel Example: After Foundational Completes

```bash
# US3's verification has no dependency on the new installer package,
# so it can run alongside US1/US2's work from here on:
Task: "Run 003-entity-creation's full test suite (T012)"
Task: "Unit tests for List() in internal/installer/installer_test.go (T006)"
```

## Parallel Example: User Story 2's Tests

```bash
Task: "Filesystem-integration tests for Install in internal/installer/installer_test.go"
Task: "Unit tests for the atomic-write helper in internal/installer/filesystem_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (the relocation, verified against `internal/templates`'s own suite).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: `go test ./internal/installer/... -run TestList` green, independently.
5. This alone already proves the embedded kit is real and discoverable — the prerequisite for everything Phase 4/5/6 of the architecture roadmap will build on top of.

### Incremental Delivery

1. Setup + Foundational (relocation, regression-checked) → `kit.TemplatesFS` is real and correct.
2. Add US1 → validate independently → discovery usable (MVP).
3. Add US2 (depends on US1) → validate independently → installation usable.
4. Verify US3 (depends only on Foundational, can run any time after) → 003's behavior proven unchanged.
5. Polish (Phase 6), including the full-module regression run (T017).

### Team Strategy

Once Foundational is done, US3's verification (T012) needs nothing from
the new `installer` package — one contributor can run it immediately
while another starts US1, in parallel. US2 must still wait on US1's
implementation specifically (`Install` calls `List`), so it isn't a
fully independent third lane the way 003-entity-creation's US1/US2 were.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] label maps every story-phase task to its spec.md user story.
- Tests are mandatory here (Constitution Principle V) — write them first, watch them fail, then implement.
- `Install` writes only to its caller-chosen target directory for framework resources — it never touches a project's `ai/` artifact tree (plan.md's Constitution Check, Principle VII).
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
