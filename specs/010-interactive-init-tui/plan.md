# Implementation Plan: Interactive Init TUI

**Branch**: `010-interactive-init-tui` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/010-interactive-init-tui/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build the interactive `misterspec init` flow `docs/architecture-specification.md`
§36-37 describes: a new `internal/tui` package (Bubble Tea) implementing
§37's own seven-screen state model — Inspect, a combined Non-Empty/
Already-Initialized Warning, Agent Selection, Preview, Installing,
Success, Error — reusing 007-project-bootstrap's `Bootstrap`/`Inspect`
and 006-agent-adapter's `Registry`/`Adapter` entirely unmodified beneath
one additive field (`InspectResult.Empty`, research.md) needed for the
non-empty-directory warning. `internal/cli/init.go` (008-cli-cobra)
gains one branch: `--agent` provided still runs the existing
non-interactive path byte-for-byte; absent, in an interactive terminal,
launches the TUI; absent, without one, fails with the exact same JSON
error shape a missing `--agent` already produced. No new deterministic
operation — the Preview screen composes already-exported, read-only
calls directly (research.md). This is the project's first feature with
a genuine human end-user at a terminal, and its first dependency beyond
`gopkg.in/yaml.v3` and `cobra`.

## Technical Context

**Language/Version**: Go 1.23.4 (existing module, unchanged).
**Primary Dependencies**: `github.com/charmbracelet/bubbletea` v1.3.10, `github.com/charmbracelet/lipgloss` v1.1.0 (new — the TUI runtime and styling §32's own `styles.go` file implies), `golang.org/x/term` (new — interactive-terminal detection, FR-010). `gopkg.in/yaml.v3` and `github.com/spf13/cobra` (existing) unchanged.
**Storage**: Filesystem — unchanged writers. This feature's only mutation anywhere is the single `bootstrap.Bootstrap` call `ScreenPreview`'s confirmation triggers; every other screen (including the Preview's own content) is built from already-proven read-only calls (research.md, FR-011).
**Testing**: `go test` — `internal/tui`'s `Model.Update`/`View` tested directly with synthetic `tea.Msg` values (no `teatest` dependency, research.md); `internal/bootstrap`'s new `Empty` field tested against fixture directory states; 007-project-bootstrap's and 008-cli-cobra's full suites re-run unmodified as named regression gates (the non-interactive `init` path must not change at all). Per Constitution Principle V.
**Target Platform**: Cross-platform Go module, unchanged — Bubble Tea itself is cross-platform (Windows/macOS/Linux terminal support is exactly its own core purpose).
**Project Type**: Single Go module. One new package, `internal/tui` (flat — no `screens/` subpackage, research.md). `internal/bootstrap` extended additively (one field). `internal/cli/init.go` modified (one new branch; existing non-interactive path untouched).
**Performance Goals**: Trivial — a handful of screens, one deterministic bootstrap call; no performance-sensitive path. Bubble Tea's own render loop is already proven at far larger scales than this flow needs.
**Constraints**: The non-interactive path (`--agent` provided) must be byte-for-byte unchanged (SC-005) — every one of 008-cli-cobra's own `TestInitCmd_*` tests is an explicit, named regression gate, not just "the usual suite." `internal/tui.Model` holds UI state only, never filesystem business rules (§37's own explicit rule) — every actual decision (is this valid, is this safe) still comes from `bootstrap`/`agents`/`installer`, never reimplemented in `tui`. No new deterministic operation, no new JSON/exit-code contract change to `internal <op>` (spec.md Assumptions).
**Scale/Scope**: `internal/tui/{model.go, update.go, view.go, styles.go}` plus tests; `internal/bootstrap/inspect.go` modified (`Empty` field) plus test; `internal/cli/init.go` modified (one new branch) plus tests. No new adapter, no change to `internal`'s ten commands, no canonical-Skill-content change.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Semantic/Deterministic Separation | **Pass.** The TUI presents choices and collects a decision (which agent, confirm or not) — it makes no judgment about project structure or content; every actual decision about what's safe or valid still comes from `bootstrap`/`agents`. |
| II. Deterministic Ops as Only Mutation Primitive | **Pass, reinforced.** The only mutation anywhere in this feature is one call to `bootstrap.Bootstrap` — the exact same deterministic primitive 007-project-bootstrap and 008-cli-cobra's non-interactive path already use; the TUI adds no second way to write anything. |
| III. Filesystem Is Single Source of Truth | **Pass.** `Inspect` is re-run fresh at `ScreenInspect`; the Preview reflects live reads, never a cached or assumed state. |
| IV. Simplicity — YAGNI & Minimal Configuration | **Pass, actively exercised.** No `screens/` subpackage for seven closely related states (research.md); no `teatest` dependency when direct `Update`/`View` testing already suffices; no per-adapter `Preview()` method when every adapter's Skills source is already uniform. |
| V. Test-First Discipline | **Gate carried into tasks.** `Model.Update`/`View` tests are mandatory and directly runnable without a real terminal; 007's and 008's full suites are named regression gates, not just "the usual suite." |
| VI. Clean Code & SOLID | **Pass, actively exercised.** The Preview screen reuses `installer.List`/`ListFS`/`Adapter.TargetPath` directly rather than a parallel preview-specific implementation — the same "compose, don't duplicate" discipline every prior feature has applied. |
| VII. Explicit Mutation Boundaries | **Pass, reinforced.** `internal/tui` itself performs zero filesystem writes — its one `bootstrap.Bootstrap` call is the package boundary where mutation happens, identical to how `internal/cli/internalcmd`'s `create`/`create-artifact` commands already delegate every write to `operations`. |
| VIII. Safety by Construction | **Pass, reinforced.** Nothing is written before explicit confirmation (FR-005); an already-initialized target is still rejected by `Bootstrap` itself even if a user chooses "continue anyway" past the warning — the TUI cannot bypass 007's own atomic, no-silent-overwrite guarantee by construction, only reach it. |
| IX. Transparent, Machine-Readable Contracts | **Pass, reinforced for the machine-caller case.** The one new failure mode this feature introduces (no `--agent`, no TTY) reuses `internal`'s existing JSON error envelope exactly (research.md) — a machine caller never sees a second, inconsistent failure shape. |

**Result**: No violations. Complexity Tracking table below is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/010-interactive-init-tui/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/
│   └── tui.md              # Phase 1 output (/speckit-plan command)
├── checklists/
│   └── requirements.md   # /speckit-specify quality checklist
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
misterspec/
├── go.mod, go.sum                       # MODIFIED: + bubbletea, lipgloss, golang.org/x/term
├── cmd/misterspec/                       # unchanged
├── kit/                                   # unchanged
└── internal/
    ├── project/, artifacts/, ids/, testutil/, example/   # unchanged
    ├── lock/, templates/, validation/, operations/         # unchanged
    ├── agents/, agents/claude/, agents/builtin/              # unchanged
    ├── installer/                                             # unchanged
    ├── bootstrap/
    │   ├── inspect.go              # MODIFIED: + InspectResult.Empty
    │   ├── inspect_test.go         # MODIFIED: + Empty test cases
    │   └── bootstrap.go, verify.go, config.go, doc.go, *_test.go  # unchanged
    ├── cli/
    │   ├── init.go                  # MODIFIED: + interactive/no-TTY branch
    │   ├── init_test.go             # MODIFIED: + new branch's test cases
    │   └── root.go, internal.go, internalcmd/  # unchanged
    └── tui/                          # NEW package
        ├── doc.go
        ├── model.go                   # Screen enum, Model struct, NewModel, Init
        ├── update.go                   # Update — per-screen transition logic
        ├── view.go                      # View — per-screen rendering
        ├── styles.go                     # lipgloss styles
        ├── model_test.go
        ├── update_test.go
        └── view_test.go
```

**Structure Decision**: One new package (`internal/tui`, flat — no
`screens/` subpackage, research.md), two existing files modified
additively (`bootstrap/inspect.go` gains one field;
`cli/init.go` gains one branch, its existing path untouched). No new
top-level composition package — `internal/tui` sits directly above
`bootstrap`/`agents`/`installer`, the same layering `internal/cli`
already established for the non-interactive path.

## Complexity Tracking

> Not applicable — the Constitution Check above recorded no violations.
