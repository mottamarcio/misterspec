# Feature Specification: CLI Command Layer (Cobra)

**Feature Branch**: `008-cli-cobra`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir a sugestão acima da camada cli com
cobra" (proceed with the CLI layer built on Cobra, per
`docs/architecture-specification.md` §4-8, §32, §36)

## User Scenarios & Testing *(mandatory)*

<!--
  Unlike 001-007, this feature's "users" now include a genuinely
  external caller for the first time: a terminal invocation (a human, a
  script, or a coding agent shelling out) rather than only another Go
  package. Every deterministic operation 001-007 already built
  (project detection, resolve/inspect/parent/children, create/
  create-artifact, fingerprint/inventory, validate/status, bootstrap)
  gets its first invocable surface here — this feature adds no new
  business logic, only a Cobra-based command tree over what already
  exists.
-->

### User Story 1 - Invoke Deterministic Operations via Stable, Scriptable Commands (Priority: P1)

A caller (a coding agent shelling out, a script, or a human at a
terminal) can invoke any of misterspec's existing deterministic
operations — resolve, inspect, parent, children, create, create an
artifact, fingerprint, list an inventory, validate, and check status —
from the command line against a target project, and receive
machine-readable JSON output with a stable, distinguishable result: a
successful answer, or a structured error naming what went wrong.

**Why this priority**: This is the actual machine-facing contract every
future canonical Skill will call (`docs/architecture-specification.md`
§5-7) — without it, none of 001-007's deterministic core is reachable
except by writing Go code. It is also the largest, most immediately
valuable slice: every operation already exists and is already proven;
this story only needs to expose it.

**Independent Test**: Can be fully tested by invoking each command
against a fixture project directory (as a subprocess) and asserting its
JSON output's shape and its exit code — both for a successful case (a
known entity resolved, a new entity created) and a failure case (an
unknown entity ID, an invalid target) — with no dependency on User
Story 2 or 3.

**Acceptance Scenarios**:

1. **Given** a valid project and a known entity ID, **When** the
   corresponding internal command is invoked, **Then** it prints a
   successful JSON result to stdout and exits with code `0`.
2. **Given** a valid project and an unknown entity ID, **When** the
   corresponding internal command is invoked, **Then** it prints a
   structured JSON error naming a stable error code and exits with a
   non-zero code specific to "not found," distinct from other failure
   categories.
3. **Given** a directory that is not a misterspec project, **When** any
   internal command requiring one is invoked, **Then** it prints a
   structured JSON error with the project-not-initialized error code and
   exits with that failure category's specific code.
4. **Given** any internal command, **When** it succeeds or fails,
   **Then** its output contains no human-oriented decorative styling —
   only the documented JSON shape.

---

### User Story 2 - Bootstrap a New Project Non-Interactively (Priority: P2)

A caller can run misterspec's public initialization command against a
target directory with an already-chosen agent, without answering any
interactive prompts, and have a new project bootstrapped — or be told
precisely why it was rejected — with the outcome reported the same
structured way User Story 1's commands report theirs.

**Why this priority**: This is `docs/architecture-specification.md`
§4.1's promised public entry point, made real for the first time — but
scoped to exactly what 007-project-bootstrap already proved
(`internal/bootstrap.Bootstrap`), not the full interactive experience
`docs/architecture-specification.md` §36-37 eventually describes. It
depends on User Story 1 only in the sense that it reuses the same JSON
envelope and error-reporting conventions — its own underlying operation
(`Bootstrap`) is independent and already complete.

**Independent Test**: Can be fully tested by running the public init
command against a fresh, empty target directory with a registered
agent, confirming the project now exists and detects successfully; and
by running it again against the same directory (rejected, already
initialized) and against a fresh directory with an unregistered agent
(rejected, unknown agent) — both rejections confirmed to leave nothing
written.

**Acceptance Scenarios**:

1. **Given** an empty target directory and a registered agent, **When**
   the public init command is run with that agent specified, **Then** a
   new project is bootstrapped there and a structured JSON success
   result reports the outcome of its configuration, kit resources, and
   agent installation separately.
2. **Given** a target directory that is already an initialized project,
   **When** the public init command is run against it, **Then** it is
   rejected with a stable, distinguishable error code and nothing is
   overwritten.
3. **Given** an empty target directory and an unregistered agent ID,
   **When** the public init command is run, **Then** it is rejected with
   a stable, distinguishable error code before any file is written.

---

### User Story 3 - Keep the Human-Facing Command Surface Minimal (Priority: P3)

A user running the tool's own help output sees only the small, public
surface the product intends to present — never the full internal
operation vocabulary User Story 1 exposes for machine callers.

**Why this priority**: `docs/architecture-specification.md` §4
explicitly requires this separation — "the product must not present
users with a large CLI command vocabulary." It is the lowest-risk,
smallest slice, layered on top of User Story 1 and 2 both already
existing, and independently verifiable once they do.

**Independent Test**: Can be fully tested by running the tool's
top-level help output and confirming it lists only the public command(s)
— never any of User Story 1's internal operation names — while those
internal commands remain fully invocable directly (not removed, only
hidden from discovery).

**Acceptance Scenarios**:

1. **Given** the built CLI, **When** its top-level help is requested,
   **Then** only the public command(s) are listed.
2. **Given** the built CLI, **When** an internal command is invoked
   directly by its known name, **Then** it still runs normally — hidden
   from help, not removed.

---

### Edge Cases

- An internal command invoked with a missing or malformed required
  argument (Edge Case, distinct from a valid but unresolvable one).
- An internal command invoked against an ambiguous, non-unique partial
  identifier (reusing 002-read-operations's existing ambiguity result).
- The public init command invoked with no agent specified at all.
- An internal command's target path resolving outside the project root
  (reusing 002-read-operations's existing containment guarantee).
- A structural validation command run against a project with existing
  anomalies — reported as a distinct failure category from "not found"
  or "invalid invocation."

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose every existing deterministic operation
  (resolve, inspect, parent, children, create, create an artifact,
  fingerprint, inventory, validate, status) as an invocable command.
- **FR-002**: Every such command MUST produce machine-readable JSON on
  success, matching the `{"ok": true, ...}` shape already established
  across 001-007's own contracts documentation.
- **FR-003**: Every such command MUST produce a structured JSON error
  with a stable, named error code and a concise message on failure —
  never an unstructured or prose-only error.
- **FR-004**: Every such command MUST exit with a non-zero code specific
  to its failure's category (invalid invocation, target not found,
  structural validation failure, mutation rejected, project not
  initialized, unexpected failure) — never one generic non-zero code for
  every kind of failure.
- **FR-005**: The machine-facing operation commands MUST NOT appear in
  the tool's top-level help output or be otherwise advertised to a
  normal user browsing available commands.
- **FR-006**: System MUST provide one public command that bootstraps a
  new project for a caller-specified agent, without requiring any
  interactive input.
- **FR-007**: The public bootstrap command MUST reject an
  already-initialized target and an unrecognized agent ID with the same
  non-destructive guarantee the underlying bootstrap capability already
  provides, reported as a stable, distinguishable error.
- **FR-008**: The public bootstrap command's successful outcome MUST
  report the specific outcome of its configuration, kit-resource, and
  agent-installation parts separately — never collapsed into one
  aggregate flag.
- **FR-009**: System MUST NOT introduce a second implementation of any
  existing deterministic operation — every command's behavior is
  supplied entirely by 001-007's already-proven capability, not
  reimplemented.
- **FR-010**: No command's output MUST contain human-oriented decorative
  styling (color, spinners, progress bars) in its machine-facing
  (internal) form.

### Key Entities

- **Command Result**: The outcome of one command invocation — either a
  successful, structured result, or a structured error naming a stable
  code and message. Never ambiguous about which one occurred.
- **Error Code**: A stable, named identifier for one category of
  failure, distinguishable by any caller without parsing prose.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of misterspec's existing deterministic operations are
  invocable from a terminal, with no operation reachable only by writing
  Go code.
- **SC-002**: 100% of command failures produce a structured result
  naming a stable error code — zero unstructured, prose-only failures
  reach a caller.
- **SC-003**: The tool's top-level help output lists only its intended
  public surface, with zero internal operation names appearing in it.
- **SC-004**: A new project can be bootstrapped end-to-end from a single
  terminal command, with zero manual file editing required afterward.
- **SC-005**: 100% of bootstrap attempts against an already-initialized
  target or naming an unregistered agent are rejected via the public
  command, with zero silent overwrites — matching the underlying
  capability's own already-proven guarantee.

## Assumptions

- This feature does **not** include the interactive Bubble Tea flow
  `docs/architecture-specification.md` §36-37 describes (agent-selection
  screen, installation preview, confirmation prompt). The public
  bootstrap command here is flag-driven and non-interactive, wrapping
  007-project-bootstrap's already-complete `Bootstrap` directly. The
  fully interactive experience remains a distinct, later feature — the
  same "deterministic core before interactive presentation" boundary
  every prior feature (002 through 007) has kept, now one layer closer
  to the user-facing product.
- Canonical Skill content and `SKILL.md` authoring
  (`docs/architecture-specification.md` §38-39) remain out of scope —
  this feature exposes the commands Skills will eventually call; it does
  not author any Skill itself.
- This feature adds no new business logic. Every command is a thin
  invocation surface over 001-007's already-implemented, already-tested
  Go packages (`internal/operations`, `internal/validation`,
  `internal/bootstrap`) — consistent with the project constitution's
  Clean Code & SOLID and DRY principles.
- Error codes beyond `docs/architecture-specification.md` §7's
  illustrative list (e.g. for an already-initialized bootstrap target,
  or an unregistered agent) are introduced following that section's own
  stated pattern ("possible stable error codes include") — extending,
  not replacing, the documented set.
- Only one concrete agent (Claude Code, via `internal/agents/builtin`)
  is registered — unchanged from 006-agent-adapter's own scoping;
  additional adapters remain a later, separate decision.
- No authentication, authorization, or network surface is introduced —
  this feature remains entirely local-filesystem-scoped, matching every
  prior feature's boundary.
