# Phase 0 Research: Interactive Init TUI

`docs/architecture-specification.md` §32, §36-37 give this feature a
concrete envelope (a `Screen` enum, a suggested package layout, a
transaction model) but leave the actual TUI library choice and a few
composition details open. This document resolves them.

## Decision: Bubble Tea + Lip Gloss, the project's first genuinely interactive dependency

- **Decision**: `github.com/charmbracelet/bubbletea` v1.3.10 (the
  `Model`/`Update`/`View` runtime §32 already names) plus
  `github.com/charmbracelet/lipgloss` v1.1.0 (terminal styling — the
  reason §32's own layout includes a `styles.go` file).
- **Rationale**: §32 already illustrates this exact shape
  (`internal/tui/{model.go, update.go, view.go, styles.go}`) — Bubble
  Tea is the de facto standard Go TUI runtime this illustration is
  clearly modeled on, and is the same ecosystem `cobra` (008-cli-cobra)
  already coexists with cleanly (no overlap: Cobra parses flags/routes
  commands, Bubble Tea owns the screen once `init` decides to run
  interactively).
- **Alternatives considered**: Hand-rolling a raw-terminal prompt loop
  with `golang.org/x/term` alone — rejected; §37's own `Screen` state
  model already implies a real state machine over multiple rendered
  views, which is exactly Bubble Tea's designed use case, and hand-
  rolling it would duplicate a well-proven library for no benefit
  (Constitution Principle IV, YAGNI in the "don't reinvent" direction).

## Decision: `golang.org/x/term` for the interactive-terminal check (FR-010)

- **Decision**: `term.IsTerminal(int(os.Stdin.Fd()))` (and `Stdout`)
  gates whether `misterspec init` (no `--agent`) launches the TUI at
  all — checked in `internal/cli/init.go` *before* `tui.RunInit` is
  ever called.
- **Rationale**: Bubble Tea itself assumes a real terminal; running it
  against a pipe produces confusing, hanging, or garbled behavior
  rather than a clear error. `golang.org/x/term` is the standard
  library's own extended module for exactly this check — minimal,
  already a transitive dependency of several terminal-aware tools, and
  avoids hand-rolling terminal-detection logic.
- **Alternatives considered**: Letting Bubble Tea itself fail and
  surfacing whatever error it produces — rejected; FR-010 requires a
  *clear*, specific message ("`--agent` is required outside an
  interactive terminal"), not Bubble Tea's own internal error text.

## Decision: `misterspec init`'s no-TTY fallback reuses `internalcmd`'s JSON error envelope

- **Decision**: When `--agent` is absent and no interactive terminal is
  available, `newInitCmd`'s `RunE` calls
  `internalcmd.WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: --agent is
  required outside an interactive terminal", internalcmd.ErrInvalidArgument))`
  — the exact same JSON shape and exit code (`invalid_argument`, 2) a
  missing `--agent` already produced in 008-cli-cobra, before this
  feature existed.
- **Rationale**: A caller with no TTY attached (a script, CI, another
  program shelling out) is definitionally a machine caller, not a human
  — the same audience `misterspec internal <op>`'s JSON contract
  already serves. Reusing the identical envelope means there is still
  exactly one machine-facing failure shape in this codebase, not two
  (one for `internal` commands, a different one for this fallback).
  `misterspec init`'s only genuinely new output mode is the interactive
  TUI itself, reached only when a human is actually present.
- **Alternatives considered**: A bespoke plain-text error for this
  specific case — rejected; would introduce a second, inconsistent
  failure shape for what is functionally the same condition
  008-cli-cobra's own `TestInitCmd_MissingAgentFlag` already covers.

## Decision: `internal/bootstrap.InspectResult` gains an additive `Empty bool` field

- **Decision**: `Inspect` (007-project-bootstrap) gains one new field on
  its existing result type — `Empty bool`, `true` when `!Initialized`
  and `targetDir` contains no entries (or does not exist yet), `false`
  otherwise. `Inspect`'s exported signature is unchanged; every existing
  caller (008-cli-cobra's `internal` commands, 007's own tests) is
  unaffected by an additive struct field.
- **Rationale**: FR-002 requires warning on a non-empty-but-
  uninitialized target — a condition `Inspect` did not previously
  report (it only ever answered "is this already a misterspec
  project," not "is this directory empty at all"). This is a read-only
  query, not a new mutation primitive, so it fits naturally as an
  extension of `Inspect`'s own existing "is it safe to bootstrap here"
  domain (007-project-bootstrap) rather than a TUI-local one-off check
  — the same reusability reasoning 007's own `Verify` reusing `Inspect`
  already established, applied one level further.
- **Alternatives considered**: A TUI-package-local directory-emptiness
  check — rejected; this is a legitimately reusable, read-only
  capability (any future non-interactive caller wanting the same
  warning would otherwise have to reimplement it), and
  `internal/bootstrap` is where "is this target safe to bootstrap"
  already lives.

## Decision: The installation preview composes existing read-only capabilities directly — no new "preview" operation

- **Decision**: The Preview screen's content is assembled from three
  already-exported, already-read-only calls: `installer.List()` (kit
  templates), `installer.ListFS(kit.SkillsFS, ".", "skill")` (the
  selected agent's Skills — the same source every adapter installs
  from, per 006-agent-adapter's own convention), and
  `adapter.TargetPath()` (where Skills land for the selected agent).
  No new function is added to `internal/bootstrap`, `internal/agents`,
  or `internal/installer` to support this.
- **Rationale**: Every one of these three already exists and is
  already read-only — composing them in the TUI layer is exactly what
  a presentation layer is for, and matches FR-011's "no new way to
  mutate a project" by construction (nothing here writes anything).
- **Alternatives considered**: Adding a `Preview()` method to the
  `Adapter` interface — rejected; every concrete adapter's Skills
  source is already `kit.SkillsFS` by convention (006's own design), so
  a per-adapter preview method would be speculative generality for a
  behavior that's already uniform across the one adapter that exists.

## Decision: `internal/tui` stays flat — no `screens/` subpackage

- **Decision**: `internal/tui/{model.go, update.go, view.go, styles.go}`
  — no `screens/` subdirectory, deviating from §32's own illustrative
  listing.
- **Rationale**: §32 itself says its structure "may evolve based on
  demonstrated complexity" and warns against introducing speculative
  layers "unless actual implementation pressure requires them." Seven
  `Screen` states sharing one `Model` and one linear transition
  sequence (§36's own transaction model, not a branching graph) does
  not demonstrate enough complexity to justify a subpackage per screen
  — a `switch` on the current `Screen` inside `update.go`/`view.go` is
  simpler and equally clear at this scale.
- **Alternatives considered**: One file per screen under `screens/`
  matching §32's illustration literally — rejected as premature
  structure for seven closely related, linearly sequenced states;
  revisit if a later feature (e.g. supporting branching flows for
  multiple agents with materially different installation steps)
  actually demonstrates the need.

## Decision: `Update`/`View` are tested directly — no `teatest` dependency

- **Decision**: Tests construct a `Model` value, call `Update` with
  synthetic `tea.KeyMsg`/`tea.Cmd` results directly, and assert the
  resulting `Model`'s fields and `View()`'s string output — no
  `github.com/charmbracelet/x/exp/teatest` (or similar) dependency.
- **Rationale**: `bubbletea.Model`'s `Update`/`View` are plain,
  synchronous Go functions — testable directly without spinning up a
  real (or simulated) terminal program. Adding a dependency
  specifically for TUI-driving test scaffolding is unnecessary when the
  interface itself is already this directly testable (Constitution
  Principle IV, YAGNI).
- **Alternatives considered**: `teatest` — rejected; adds a dependency
  (and an unclear pseudo-versioned import path, confirmed while
  checking available versions) for a testing convenience this project
  doesn't need at this scale.

## Decision: `misterspec init`'s exit code on a completed interactive session is always `0`, regardless of the user's own choice inside it

- **Decision**: `tui.RunInit` returns a non-nil `error` only for a
  genuinely unexpected failure (the Bubble Tea program itself failing
  to start or run) — never to represent "the user declined" or "the
  bootstrap attempt failed," both of which are shown on-screen (the
  Error screen, FR-008) and end the program normally. `newInitCmd`'s
  `RunE` returns `nil` (exit `0`) whenever `RunInit` itself completed,
  whatever the user decided inside it.
- **Rationale**: An interactive program that ran correctly and let the
  user see and act on every outcome — including declining, or a clearly
  reported failure — did its job; conflating "the TUI ran successfully"
  with "the user chose to bootstrap" would make a deliberate "no" look
  like a process crash to any script checking the exit code, which
  contradicts FR-009's "exit cleanly" requirement.
- **Alternatives considered**: Exit code reflecting whether a project
  was actually bootstrapped — rejected; this is a human-facing
  interactive command, not `internal`'s machine-facing contract
  (008-cli-cobra), so there is no established convention this would
  need to match, and a non-zero exit for a deliberate, successful "no"
  would be actively misleading.

## Output

All unknowns resolved. No `NEEDS CLARIFICATION` markers remain.
