# Phase 1 Contracts: Interactive Init TUI

Like 008-cli-cobra, this feature's contract is partly an external
interface — a terminal UI a human actually sees — so this document
specifies observable screen behavior (data-model.md's transitions)
alongside Go signatures.

**Reconciled against the actual implementation (T030)** — three things
discovered during implementation, not present in the original draft
below:

1. **Dependency versions were pinned lower than planned.**
   `github.com/charmbracelet/bubbletea` is `v1.3.7`, not `v1.3.10` —
   `v1.3.9`/`v1.3.10` require Go ≥ 1.24, which would have silently
   bumped this module's own `go` directive (confirmed by direct
   experiment: `go get`'s automatic toolchain switch pulled `go1.26.8`
   the first time). `golang.org/x/term` is `v0.34.0`, not `@latest`
   (`v0.46.0`), for the identical reason. `github.com/charmbracelet/lipgloss`
   stayed at the planned `v1.1.0` (it only needs Go 1.18). This module's
   `go 1.23.4` directive is unchanged, as plan.md's Technical Context
   required.
2. **`ScreenSuccess`/`ScreenError` auto-quit immediately after
   rendering, rather than waiting for "any key."** The original design
   (below, as first drafted) required a keypress to quit from these two
   terminal screens. Implementing User Story 1's own integration test
   (driving `RunInit` with synthetic input) surfaced a real race: a
   keypress meant for the *next* screen can arrive while `ScreenInstalling`'s
   async `bootstrap.Bootstrap` call is still in flight, and once that
   call's result lands, there was nothing left in the test's input
   stream to trigger the "any key" quit — an indefinite hang, not a
   flaky timing issue confined to tests (the same race exists for any
   caller driving the program non-interactively). Fixed by having
   `handleBootstrapResult` (and the `ScreenInspect`→`ScreenError` path)
   return `tea.Quit` in the same `Update` call that sets the terminal
   screen — Bubble Tea still renders that screen's final `View()` before
   exiting, so nothing is lost, and the program no longer depends on a
   keypress that may never arrive.
3. **A related, narrower race is now explicitly documented, not
   silently possible:** a keypress arriving while still on `ScreenInspect`
   (before `Init()`'s async `bootstrap.Inspect` result lands) has no
   case in `handleKey`'s switch and is silently dropped. Harmless for a
   real human (no one types during a sub-millisecond local filesystem
   call), but real for any synthetic-input caller (this feature's own
   integration tests included) — those tests now synchronize with a
   short delay before sending the first key, documented at the call
   site rather than left as an unexplained sleep.

The `Screen transitions` table below reflects the corrected (auto-quit)
behavior directly, not the originally drafted one.

## `internal/bootstrap` (extended, additive)

```go
package bootstrap

// InspectResult gains one additive field (research.md); every other
// field and Inspect's own exported signature are unchanged.
type InspectResult struct {
    Initialized    bool
    ProjectRoot    string
    InstalledAgent string
    AgentInstalled bool
    Empty          bool // NEW
}

func Inspect(targetDir string) (InspectResult, error) // unchanged signature
```

**Guarantees**: Every existing 007/008/009 test passes unmodified — an
additive struct field changes nothing for a caller using keyed struct
literals (every caller in this codebase already does).

## `internal/tui`

```go
package tui

// RunInit runs the interactive init flow to completion: inspect the
// target, warn if it's already initialized or non-empty, let the user
// pick a registered agent, preview exactly what will be created, and —
// only on explicit confirmation — bootstrap it via bootstrap.Bootstrap
// (the same deterministic core 007-project-bootstrap already proved;
// FR-006, FR-011).
//
// RunInit returns a non-nil error only for a genuine failure to run
// the interactive program itself — never to represent a user's own
// decision (decline, cancel) or a reported bootstrap failure, both of
// which are shown on-screen and end the program normally
// (research.md's exit-code decision).
func RunInit(targetDir string, registry *agents.Registry, skills fs.FS) error
```

```go
package tui

// Screen is one of the seven states this flow can be in
// (docs/architecture-specification.md §37, verbatim).
type Screen int

const (
    ScreenInspect Screen = iota
    ScreenNonEmptyWarning
    ScreenAgentSelection
    ScreenPreview
    ScreenInstalling
    ScreenSuccess
    ScreenError
)

// Model implements tea.Model (Init/Update/View) — UI state only, no
// filesystem business rules of its own (§37's own explicit rule;
// data-model.md).
type Model struct{ /* unexported fields, data-model.md */ }

func NewModel(targetDir string, registry *agents.Registry, skills fs.FS) Model
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

**Guarantees**:
- `Model.Update`/`Model.View` are plain, synchronous functions —
  directly unit-testable with synthetic `tea.Msg` values, no real
  terminal or `teatest`-style harness required (research.md).
- The only filesystem mutation anywhere in this package is the single
  `bootstrap.Bootstrap` call triggered from `ScreenPreview`'s
  confirmation — every other screen (including the Preview screen's own
  content) is built from read-only calls (`bootstrap.Inspect`,
  `agents.Registry.List`, `installer.List`, `installer.ListFS`,
  `Adapter.TargetPath`) already proven elsewhere (FR-011).

## `internal/cli` (extended)

```go
package cli

// newInitCmd's RunE, extended (008-cli-cobra's non-interactive path
// unchanged when --agent is provided):
//
//   if agent != "" {
//       // 008's existing path — unchanged, byte-for-byte.
//   }
//   if !term.IsTerminal(int(os.Stdin.Fd())) {
//       return internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf(
//           "%w: --agent is required outside an interactive terminal",
//           internalcmd.ErrInvalidArgument))
//   }
//   return tui.RunInit(dir, builtin.Default(), kit.SkillsFS)
func newInitCmd() *cobra.Command
```

**Guarantees** (FR-001, FR-010, SC-005):
- `misterspec init --agent <id> [--dir <path>]` behaves identically to
  008-cli-cobra's own existing behavior, 100% unchanged — this feature
  adds a branch before that path, never modifies it.
- `misterspec init` (no `--agent`) outside an interactive terminal fails
  immediately with the exact same JSON error shape and exit code
  (`invalid_argument`, 2) a missing `--agent` already produced before
  this feature existed (research.md) — never a hang, never a different
  failure shape for what is functionally the same condition.

## Screen transitions (data-model.md, the human-facing contract)

| From | User action | To |
|---|---|---|
| `ScreenInspect` | (automatic, on `Inspect` completing) | `ScreenNonEmptyWarning` (if already initialized or non-empty) or `ScreenAgentSelection` |
| `ScreenInspect` | (automatic, on a genuine `Inspect` error) | `ScreenError`, then quit immediately after rendering |
| `ScreenNonEmptyWarning` | Cancel | quit, nothing written |
| `ScreenNonEmptyWarning` | Continue anyway | `ScreenAgentSelection` |
| `ScreenAgentSelection` | pick + confirm | `ScreenPreview` |
| `ScreenAgentSelection` | cancel | quit, nothing written |
| `ScreenPreview` | confirm | `ScreenInstalling` |
| `ScreenPreview` | decline | quit, nothing written (FR-005) |
| `ScreenInstalling` | (automatic, on `Bootstrap` succeeding) | `ScreenSuccess`, then quit immediately after rendering |
| `ScreenInstalling` | (automatic, on `Bootstrap` failing) | `ScreenError`, then quit immediately after rendering |

Cancelling (e.g. Ctrl+C) is honored at every screen, not only the ones
listed with an explicit "cancel" action — `tea.KeyCtrlC` maps to
`quitting = true` globally in `Update`, before any per-screen handling
(FR-009's own edge case).
