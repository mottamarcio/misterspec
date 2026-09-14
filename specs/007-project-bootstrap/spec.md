# Feature Specification: Project Bootstrap

**Feature Branch**: `007-project-bootstrap`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "seguir com a fase 5 sugerida acima" (proceed
with Phase 5: `misterspec init` — per
`docs/architecture-specification.md` §4, §36-37, §68)

## User Scenarios & Testing *(mandatory)*

<!--
  As with the prior six features, this has no end-user UI of its own yet.
  Its "users" are misterspec's own future `misterspec init` command and
  the coding agent that eventually drives it — per
  docs/architecture-specification.md §36, `misterspec init` needs a
  deterministic, testable "inspect, then bootstrap, then verify" core it
  can call before any interactive presentation is layered on top. Every
  story composes 001-core-foundation's, 005-embedded-kit's, and
  006-agent-adapter's already-proven primitives rather than starting
  over.
-->

### User Story 1 - Inspect a Target Directory Before Bootstrapping (Priority: P1)

Given a target directory, any consumer can determine — without writing
anything — whether it is not yet a misterspec project (safe to
bootstrap), or already is one (and if so, which agent, if any, is
currently installed there).

**Why this priority**: This is the read-only check `misterspec init`
must run before doing anything destructive — the architecture's own
transaction model (§36) starts with "Inspect," before any decision is
collected or any file is written. It is also the smallest independently
useful slice, reusing `001-core-foundation`'s project detection and
`006-agent-adapter`'s installation-record reading directly.

**Independent Test**: Can be fully tested by inspecting a fresh, empty
directory (reported as uninitialized) and a directory with an existing
project and a known installed agent (reported as already initialized,
naming that agent) — no filesystem writes involved in either case.

**Acceptance Scenarios**:

1. **Given** an empty directory with no misterspec project, **When** it
   is inspected, **Then** it is reported as uninitialized and safe to
   bootstrap.
2. **Given** a directory that already has a misterspec project with a
   known agent installed, **When** it is inspected, **Then** it is
   reported as already initialized, naming that agent.
3. **Given** a directory that already has a misterspec project but no
   agent has ever been installed there, **When** it is inspected,
   **Then** it is reported as already initialized, with no agent
   installed.

---

### User Story 2 - Bootstrap a New Project for a Chosen Agent (Priority: P2)

Given an uninitialized target directory and one already-chosen, valid
agent ID, any consumer can atomically bootstrap a new misterspec
project: its configuration is created, the framework's kit resources are
installed, and Skills are installed for that agent — with the specific
outcome of every part reported, never a single aggregate pass/fail.

**Why this priority**: This is the actual "set up misterspec here"
capability a future `misterspec init` will orchestrate, once a user (via
an interactive step this feature doesn't build) has already picked an
agent. It builds directly on User Story 1 (bootstrapping only proceeds
on an uninitialized target) and on every write guarantee
`005-embedded-kit` and `006-agent-adapter` already proved.

**Independent Test**: Can be fully tested by bootstrapping a fresh
temporary directory for a known, registered agent and confirming the
configuration file, kit resources, and agent Skills all land correctly
with specific per-resource outcomes; and by confirming both an
already-initialized target and an unregistered agent ID are rejected
before anything is written.

**Acceptance Scenarios**:

1. **Given** an uninitialized target directory and a registered agent
   ID, **When** bootstrap runs, **Then** a valid project configuration is
   created, the kit's templates are installed, and Skills are installed
   for that agent, each with its own reported outcome.
2. **Given** a target directory that is already an initialized project,
   **When** bootstrap is attempted, **Then** it is rejected with a
   specific, distinguishable result and nothing is overwritten.
3. **Given** an uninitialized target directory but an unregistered agent
   ID, **When** bootstrap is attempted, **Then** it is rejected with a
   specific, distinguishable result before any file is written.

---

### User Story 3 - Verify a Completed Bootstrap (Priority: P3)

After bootstrapping, any consumer can confirm the result is genuinely
usable: the directory now detects as a valid project, and the agent
actually installed matches the one that was requested.

**Why this priority**: This is §36's "Verify" step made concrete — the
difference between "the bootstrap call returned no error" and "the
result can actually be trusted and used." It is independently testable
against User Story 2's output, reusing detection and installation-record
reading rather than inventing a new correctness check.

**Independent Test**: Can be fully tested by bootstrapping a fixture
directory, then verifying it — confirming the project detects
successfully and the recorded agent matches exactly what was requested.

**Acceptance Scenarios**:

1. **Given** a directory that was just successfully bootstrapped for a
   specific agent, **When** it is verified, **Then** the project is
   confirmed detectable and the installed agent is confirmed to match
   the one requested.

---

### Edge Cases

- Bootstrapping a directory that already has an initialized misterspec
  project (must be rejected, never silently overwritten).
- Bootstrapping with an agent ID that isn't registered.
- Bootstrapping into a directory that doesn't exist on disk yet at all
  (must be created as part of bootstrapping).
- A bootstrap attempt interrupted partway through — some resources
  written, others not — verification must correctly reflect the
  incomplete state rather than reporting success.
- Inspecting a directory whose `.misterspec/config.yaml` exists but is
  malformed (mirrors `001-core-foundation`'s own distinct "invalid
  configuration" result, not folded into a generic failure).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST determine, for a target directory, whether it
  is not yet a misterspec project, without writing anything.
- **FR-002**: System MUST determine, for a target directory that already
  is a misterspec project, which agent (if any) is currently installed,
  without writing anything or re-installing.
- **FR-003**: System MUST create a new project's configuration with
  valid, complete defaults as part of bootstrapping an uninitialized
  directory.
- **FR-004**: System MUST reject bootstrapping a directory that is
  already an initialized misterspec project, with a specific,
  distinguishable result — never silently overwriting an existing
  project's configuration.
- **FR-005**: System MUST install the framework's kit resources
  (templates) into the new project as part of bootstrapping.
- **FR-006**: System MUST install Skills for exactly one caller-chosen,
  already-valid agent as part of bootstrapping, using that agent's own
  adapter.
- **FR-007**: System MUST reject bootstrapping with an agent ID that
  isn't registered, with a specific, distinguishable result, before any
  file is written.
- **FR-008**: Bootstrapping MUST report the specific outcome of every
  part attempted (configuration, kit resources, agent Skills) — never a
  single aggregate pass/fail flag for the whole operation.
- **FR-009**: A failed or interrupted bootstrap attempt MUST NOT leave
  any resource partially written.
- **FR-010**: After a successful bootstrap, the resulting project MUST be
  detectable as valid using the project's own existing detection
  capability — not a special case bootstrapping invents for itself.
- **FR-011**: System MUST allow a caller to verify a completed bootstrap:
  confirming the project detects successfully and that the installed
  agent's record matches what was requested.

### Key Entities

- **Inspection Result**: The outcome of inspecting a target directory —
  either "not yet a project," or "already a project," naming the
  currently installed agent if any.
- **Bootstrap Outcome**: The result of one bootstrap attempt — the
  specific outcome of its configuration, kit resource, and agent-install
  parts, never a single pass/fail flag.
- **Verification Result**: The outcome of verifying a completed
  bootstrap — whether the project detects successfully and which agent
  is actually confirmed installed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of bootstrap attempts against an already-initialized
  directory are rejected, never silently overwriting anything already
  there.
- **SC-002**: 100% of bootstrap attempts naming an unregistered agent ID
  are rejected before any file is written.
- **SC-003**: A freshly bootstrapped project is detected as valid by the
  project's detection capability 100% of the time, with zero manual
  fix-up required.
- **SC-004**: Every part of a bootstrap attempt (configuration, kit
  resources, agent Skills) reports its own specific outcome — never
  collapsed into one aggregate pass/fail.
- **SC-005**: Verifying a successful bootstrap always confirms the exact
  agent that was requested, 100% of the time.

## Assumptions

- This feature does **not** include the `cmd/misterspec` binary, Cobra
  command parsing, or the interactive Bubble Tea flow (agent-selection
  UI, preview screen, confirmation prompt) `docs/architecture-specification.md`
  §36-37 describes. Those are presentation-layer concerns that build on
  this feature's deterministic orchestration and remain a distinct,
  later feature — the same "operations before CLI" boundary every prior
  feature in this project has kept (002, 003, 004, 005, 006 all
  explicitly deferred a CLI/interactive surface).
- §36's "collect user decisions," "preview," and "confirm" steps are
  inherently interactive and are therefore out of scope here. This
  feature's bootstrap operation takes an already-decided agent ID as a
  plain argument, the same way `003-entity-creation`'s `Create` takes an
  already-decided entity type rather than prompting for one itself.
- New project configuration defaults mirror `.misterspec/config.yaml`'s
  already-frozen schema (`docs/architecture-specification.md` §21)
  exactly — no new configuration surface is invented here.
- This feature builds directly on `001-core-foundation` (project
  detection/configuration), `005-embedded-kit` (kit resource
  installation), and `006-agent-adapter` (agent registry, adapters,
  installation records) — no new external dependency, still no CLI/JSON
  surface.
- The consumers ("users") of this feature are, in priority order: (1)
  misterspec's own future `misterspec init` command, indirectly via (2)
  the coding agent and end user who will eventually run it — a
  continuation of the same framing every prior feature has used, now one
  step closer to the actual public command.
