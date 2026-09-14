# Feature Specification: Interactive Init TUI

**Feature Branch**: `010-interactive-init-tui`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "interactive Bubble Tea flow for
misterspec init: agent selection, installation preview, confirmation
prompt, per docs/architecture-specification.md §36-37"

## User Scenarios & Testing *(mandatory)*

<!--
  This feature's "user" is, for the first time in this project, a real
  human sitting at a terminal running "misterspec init" with no flags —
  not a coding agent shelling out, and not a future caller. 008-cli-cobra
  already built the deterministic, flag-driven "misterspec init --agent
  <id> [--dir <path>]" — that non-interactive path is unchanged and
  remains available; this feature adds the interactive experience
  docs/architecture-specification.md §36-37 describes for when a human
  runs "misterspec init" without already knowing which agent to name.
-->

### User Story 1 - Bootstrap a Project Interactively, Start to Finish (Priority: P1)

A user with no particular flags in mind runs `misterspec init` in an
empty directory, is shown which coding agent(s) are available, picks
one, sees exactly what will be installed before anything happens, and
confirms — after which the project is bootstrapped and a clear success
summary is shown.

**Why this priority**: This is the actual product experience
`docs/architecture-specification.md` §4.1 promises as misterspec's one
public command — a human should never need to already know an agent's
exact ID string or read documentation before running `misterspec
init`. It is the smallest complete slice: the whole transaction model
(§36) exercised once, successfully, end to end.

**Independent Test**: Run `misterspec init` with no flags in a fresh,
empty, writable directory in an interactive terminal; select the one
available agent when prompted; review the preview screen; confirm; and
observe a success screen naming what was installed — using only
008-cli-cobra's already-proven `bootstrap.Bootstrap` underneath, no new
deterministic behavior.

**Acceptance Scenarios**:

1. **Given** an empty, uninitialized directory and an interactive
   terminal, **When** the user runs `misterspec init` with no flags,
   **Then** they are shown the available coding agent(s) to choose from.
2. **Given** the user has selected an agent, **When** the flow proceeds,
   **Then** they are shown a preview of exactly what will be created and
   installed before anything is written.
3. **Given** the user is looking at the preview, **When** they confirm,
   **Then** the project is bootstrapped and a success screen reports
   what was created, matching what the preview showed.
4. **Given** the user is looking at the preview, **When** they decline
   to confirm, **Then** nothing is written and the program exits
   cleanly.
5. **Given** the user explicitly provides `--agent <id>` (and optionally
   `--dir`), **When** they run `misterspec init`, **Then** the
   non-interactive path (008-cli-cobra) runs exactly as it already does
   — no interactive screens appear.

---

### User Story 2 - Warn Before Touching an Already-Initialized or Non-Empty Directory (Priority: P2)

Before any interactive decision is even offered, the user is warned if
the target directory is already a misterspec project, or already
contains other files, and must explicitly acknowledge before the flow
continues — never silently overwriting or bootstrapping into clutter.

**Why this priority**: `docs/architecture-specification.md` §36's own
transaction model starts with "Inspect," before any decision is
collected — this is that inspection made visible to a human, and the
safety guarantee (never silently overwrite) this project has held at
every layer since 007-project-bootstrap. It depends on User Story 1's
flow existing to be warned *before*.

**Independent Test**: Run `misterspec init` with no flags against a
directory that is already an initialized misterspec project, and
separately against a non-empty directory that isn't; confirm each shows
a clear warning before any agent-selection prompt, and that declining
to proceed leaves the directory untouched.

**Acceptance Scenarios**:

1. **Given** a directory that is already an initialized misterspec
   project, **When** the user runs `misterspec init` there, **Then**
   they are warned it is already initialized (naming the currently
   installed agent, if any) before any further prompt appears.
2. **Given** a directory that is not a misterspec project but already
   contains other files, **When** the user runs `misterspec init`
   there, **Then** they are warned the directory is not empty before
   any further prompt appears.
3. **Given** either warning is shown, **When** the user declines to
   proceed, **Then** the program exits without writing anything.
4. **Given** either warning is shown, **When** the user explicitly
   chooses to proceed anyway, **Then** the flow continues to agent
   selection — an already-initialized target still cannot actually be
   re-bootstrapped (the underlying guarantee from 007-project-bootstrap
   is never bypassed), so proceeding past this warning for an
   already-initialized directory leads to a clear explanation, not a
   silent no-op or a forced overwrite.

---

### User Story 3 - Clear Recovery When Installation Fails Partway (Priority: P3)

If bootstrapping fails after the user has already confirmed — a
filesystem permission error, a disk-full condition, or any other
failure partway through — the user sees specifically what failed and
what (if anything) was actually written, not a crash, a silent exit, or
a misleading success message.

**Why this priority**: This is the lowest-probability but
highest-consequence path — a user who confirmed a preview and then hit
an unclear failure has no way to know whether it's safe to retry. It
depends on User Story 1's flow reaching the installation step to have
anything to fail.

**Independent Test**: Simulate a failure partway through installation
(e.g. a target that becomes unwritable after confirmation) and confirm
the resulting screen names the specific failure and leaves the
directory in a state the existing `misterspec init` (non-interactive)
or a repeat interactive attempt can safely re-run against.

**Acceptance Scenarios**:

1. **Given** the user has confirmed and installation begins, **When** a
   step fails, **Then** the screen names specifically what failed,
   distinct from a generic "something went wrong."
2. **Given** a failure has been reported, **When** the user reads it,
   **Then** it clearly states whether re-running `misterspec init` is
   safe (it always is, per 007-project-bootstrap's own
   inspect-then-reject guarantee — a partial project is still
   detectable and rerunnable, never silently corrupted).

---

### Edge Cases

- `misterspec init` run with no flags outside an interactive terminal
  (output piped, no TTY attached — e.g. inside a script or CI step):
  the interactive flow cannot meaningfully prompt for input; this must
  fail clearly and immediately, explaining that `--agent` is required
  in a non-interactive context, rather than hanging waiting for input
  that will never come.
- Exactly one coding agent is currently registered (Claude Code) — the
  agent-selection screen must still function correctly and clearly with
  only one choice, not assume multiple are always available.
- The user cancels (e.g. Ctrl+C) at any screen, not only the
  confirmation screen — the program must exit cleanly without a partial
  write in every case, not only the ones this spec's acceptance
  scenarios enumerate directly.
- The target directory does not exist yet at all (must be created as
  part of a confirmed installation, matching 007-project-bootstrap's
  own existing behavior — the interactive flow doesn't change this, only
  previews it before it happens).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST launch an interactive flow when `misterspec
  init` is run without an `--agent` flag in an interactive terminal, and
  MUST run the existing non-interactive path unchanged when `--agent` is
  provided.
- **FR-002**: The interactive flow MUST inspect the target directory
  before offering any decision to the user, and MUST warn the user —
  requiring explicit acknowledgment before continuing — when the target
  is already an initialized misterspec project or already contains
  other files.
- **FR-003**: The interactive flow MUST present every currently
  registered coding agent as a choice, and MUST require the user to
  select exactly one before proceeding.
- **FR-004**: The interactive flow MUST show the user a preview of
  exactly what will be created and installed — the project
  configuration, the framework's kit resources, and the selected
  agent's Skills — before writing anything.
- **FR-005**: The interactive flow MUST require an explicit user
  confirmation after the preview before any file is written, and MUST
  write nothing at all if the user declines.
- **FR-006**: Once confirmed, the interactive flow MUST perform the
  actual bootstrap using the project's existing deterministic bootstrap
  capability — never a second, parallel implementation of project
  creation.
- **FR-007**: After a successful bootstrap, the interactive flow MUST
  show a completion summary naming what was actually created, matching
  what the preview showed.
- **FR-008**: If bootstrapping fails after confirmation, the interactive
  flow MUST show the specific failure, distinct from a generic error,
  and MUST make clear that re-running `misterspec init` is safe.
- **FR-009**: The interactive flow MUST exit cleanly, without writing
  anything, if the user cancels at any point before confirmation.
- **FR-010**: When `misterspec init` (no `--agent` flag) is run without
  an interactive terminal available, the system MUST fail immediately
  with a clear message that `--agent` is required in that context —
  never hang waiting for input.
- **FR-011**: The interactive flow MUST introduce no new way to mutate
  a project — every actual filesystem write remains
  007-project-bootstrap's `Bootstrap` and its own already-proven atomic,
  no-silent-overwrite guarantees.

### Key Entities

- **Installation Plan Preview**: The user-facing summary of exactly
  what an interactive `misterspec init` run is about to do —
  configuration, kit resources, and agent Skills — shown before
  confirmation and never differing from what actually gets written once
  confirmed.
- **Screen**: One state of the interactive flow (inspecting, warning,
  agent selection, preview, installing, success, error) — the user is
  always in exactly one at a time, and every screen is reachable only
  through the flow's own defined transitions, never skipped silently.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user with no prior knowledge of misterspec's agent IDs
  or JSON output can successfully bootstrap a project by running
  `misterspec init` alone and following the on-screen prompts, with
  zero need to consult documentation first.
- **SC-002**: 100% of interactive runs against an already-initialized or
  non-empty target show a warning before any further prompt — zero
  silent overwrites, matching 007-project-bootstrap's own already-proven
  guarantee now made visible.
- **SC-003**: 100% of interactive runs write nothing to disk unless the
  user explicitly confirmed the exact preview shown.
- **SC-004**: 100% of installation failures after confirmation report
  a specific cause and explicitly confirm that retrying is safe — zero
  generic, unexplained failures.
- **SC-005**: The non-interactive path (`--agent` provided) behaves
  identically to 008-cli-cobra's own existing behavior 100% of the time
  — zero regression introduced by this feature.

## Assumptions

- This feature builds only on the one coding agent already registered
  (Claude Code, 006-agent-adapter) — additional adapters (a second and
  third coding agent) are explicitly a separate, later feature, not
  bundled into this one. The agent-selection screen must work correctly
  with exactly one choice today, and scale to more without a redesign
  once they exist.
- `misterspec init --agent <id> [--dir <path>]` (008-cli-cobra) remains
  the supported non-interactive entry point unchanged — providing
  `--agent` explicitly always skips the interactive flow entirely,
  matching how flag-driven tools conventionally distinguish "I already
  know what I want" from "prompt me."
- This feature introduces the project's first genuinely interactive,
  full-terminal-UI dependency — the specific technology choice is a
  planning-level decision, not specified here, consistent with this
  specification staying implementation-agnostic.
- Every actual filesystem mutation still flows through
  007-project-bootstrap's `Bootstrap` — this feature is a presentation
  layer over an already-complete, already-tested deterministic core, not
  a reimplementation of any part of it.
- No new coding agent, no new deterministic operation, and no change to
  `misterspec internal <op>`'s JSON/exit-code contract (008-cli-cobra) —
  this feature is scoped entirely to the human-facing `misterspec init`
  experience.
