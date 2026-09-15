# Implementation Plan: Init Scaffolding and Binary Distribution

**Branch**: `021-init-scaffold-distribution` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/021-init-scaffold-distribution/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Three related onboarding/distribution fixes for a v1.0.1-flavored
patch: (1) stop `bootstrap.Bootstrap` from materializing the
framework's own artifact templates into every initialized project —
verified dead weight, since rendering already reads them from the
binary's own embedded copy; (2) have `Bootstrap` scaffold every
directory `project.Configuration`'s own defaults name (not only the
chosen agent's Skills directory), writing a `.gitkeep` into any that
is genuinely empty so the structure is committable from the first
commit; (3) publish pre-built, cross-compiled binaries as GitHub
Release assets via a new Actions workflow, linked from both `README.md`
and the docs site alongside (not replacing) the existing
build-from-source instructions. Implementation and review stay on
`dev` — no promotion to `stg`/`main` as part of this feature (research.md #8).

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged) for (1) and (2); GitHub Actions YAML + Go's own native cross-compilation for (3); Markdown/HTML for the README/docs-site link updates.
**Primary Dependencies**: None new. `os.MkdirAll`/already-existing `installer.WriteAtomicFile` for (2); `softprops/action-gh-release` (a third-party Action, not a Go module dependency) for (3)'s own release-asset upload.
**Storage**: N/A beyond the target project's own filesystem (unchanged shape of what already gets written, minus templates, plus directories) and GitHub's own Release-asset storage for (3).
**Testing**: `go test` — `internal/bootstrap` gains coverage for the new `scaffoldDirectories` (every default directory created, `.gitkeep` present only in genuinely empty directories, idempotent against pre-existing content) and for `Bootstrap`'s own updated `BootstrapOutcome` shape (no `TemplateOutcomes`, a populated `DirectoriesScaffolded`); `internal/cli`/`internal/tui` tests updated for the new JSON/Preview/Success shapes. (3) has no Go test — verified by actually running the release workflow and downloading a resulting binary (quickstart.md). 001-020's own full suites re-run unmodified as the standing regression gate.
**Target Platform**: Cross-platform Go module, unchanged; (3) specifically targets linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 (research.md #6).
**Project Type**: Single Go module CLI, plus one new GitHub Actions workflow and two documentation edits (README, docs site).
**Performance Goals**: N/A — one-time `init` setup cost is unchanged in order of magnitude (a handful of `MkdirAll`/empty-file writes replacing a handful of template-file writes).
**Constraints**: FR-002/FR-005/FR-006 require zero behavioral regression in artifact creation and zero disturbance of any real content already on disk — `scaffoldDirectories` must be provably idempotent and additive-only. FR-008 requires a downloaded binary to be behaviorally identical to a source build at the same version — achieved by building from the exact same tagged commit with no build-flag differences beyond `GOOS`/`GOARCH`.
**Scale/Scope**: `internal/bootstrap/{scaffold.go, bootstrap.go, config.go}` (one new file, one modified); `internal/tui/{model.go, update.go, view.go}` (Preview/Success content updated); `internal/cli/init.go` (JSON shape updated); `.github/workflows/release.yml` (new); `README.md` and `site/getting-started.html` (download-link additions).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass (N/A).** All three changes are mechanical: stop writing unused files, write a fixed set of directories, publish a build artifact. No semantic judgment involved anywhere. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass.** `scaffoldDirectories` allocates nothing (no ID, no path guess) — it creates exactly the fixed directory names `project.Configuration`'s own already-canonical `Default*` constants already name; no new deterministic operation is introduced, only one new internal helper composing existing constants. |
| III. Filesystem Is Single Source of Truth | **Pass (N/A).** No new persisted/authoritative state; `.gitkeep` files carry no project meaning misterspec itself ever reads back. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** The generic `installer.Install`/`List` primitives are kept, not deleted, since removing them isn't required by any FR (research.md #1); no new package for one small function (research.md #2); no new build tool (goreleaser) where plain `go build` already suffices (research.md #6); no archive/compression step (research.md #6). |
| V. Test-First Discipline | **Gate carried into tasks.** New coverage for `scaffoldDirectories` and the updated `BootstrapOutcome`/JSON/Preview shapes is required before/alongside implementation. |
| VI. Clean Code & SOLID | **Pass.** `scaffoldDirectories` lives in the one package that already owns bootstrap-time, one-shot setup (`internal/bootstrap`) rather than a new layer (SRP); it reuses `project.Default*` constants and `installer.WriteAtomicFile` rather than re-deriving paths or reimplementing atomic writes. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** `scaffoldDirectories` only ever creates directories/`.gitkeep` files it owns; it never touches a directory's own existing real content (FR-006). |
| VIII. Safety by Construction | **Pass (N/A new surface).** Every directory scaffolded is one of the project's own already-validated, fixed configuration paths — no new user-supplied path is involved. |
| IX. Transparent, Machine-Readable Contracts | **Pass, directly exercised.** `init`'s own JSON success payload is corrected to remove a now-meaningless `"templates"` field rather than leave a misleading empty one (research.md #4) — matching this Principle's own "distinguish... ok from... result" spirit of the contract meaning what it says. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/021-init-scaffold-distribution/
├── plan.md                          # This file (/speckit-plan command output)
├── research.md                      # Phase 0 output (/speckit-plan command)
├── data-model.md                    # Phase 1 output (/speckit-plan command)
├── quickstart.md                    # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── init-outcome.md              # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md              # /speckit-specify quality checklist
└── tasks.md                         # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                          # unchanged — no new Go dependency
├── internal/
│   ├── bootstrap/
│   │   ├── bootstrap.go                       # MODIFIED — remove templates install, add scaffold call, new outcome field
│   │   ├── config.go                          # unchanged — writeDefaultConfig reused as-is
│   │   ├── scaffold.go                        # NEW — scaffoldDirectories
│   │   └── scaffold_test.go                   # NEW
│   ├── tui/
│   │   ├── model.go                           # MODIFIED — previewContent.directories replaces .templates
│   │   ├── update.go                          # MODIFIED — buildPreview computes directories, not installer.List()
│   │   └── view.go                            # MODIFIED — Preview/Success screens render directories
│   ├── cli/
│   │   └── init.go                            # MODIFIED — JSON payload: "directories" replaces "templates"
│   └── installer/                             # unchanged — Install/List kept, just no longer called from bootstrap (research.md #1)
├── .github/workflows/
│   └── release.yml                            # NEW — cross-compiled binaries attached to tagged GitHub Releases
├── README.md                                  # MODIFIED — binary-download install option added
└── site/getting-started.html                  # MODIFIED — same addition on the docs site
```

**Structure Decision**: All Go changes stay inside the three packages
that already own this behavior (`bootstrap`, `tui`, `cli`) — no new
package. Distribution is one new, self-contained workflow file plus
two documentation edits. Nothing under `kit/`, `internal/context/`, or
any other feature's own package is touched.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
