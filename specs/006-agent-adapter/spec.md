# Feature Specification: Agent Adapter Layer

**Feature Branch**: `006-agent-adapter`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "seguir com a fase 4 sugerida acima" (proceed
with Phase 4: Agent Adapters — an adapter interface, a registry, and at
least one working adapter, per
`docs/architecture-specification.md` §34-35, §68)

## User Scenarios & Testing *(mandatory)*

<!--
  As with the prior five features, this has no end-user UI. Its "users"
  are the coding agent and misterspec's own future `misterspec init`
  command, which — per docs/architecture-specification.md §34-35 — must
  be able to discover which coding agents misterspec can install for,
  and materialize the framework's Skills into that agent's own
  integration location, without inventing its own installation logic per
  agent. Every story builds on 005-embedded-kit's proven materialization
  primitives rather than starting over.
-->

### User Story 1 - Discover Available Adapters (Priority: P1)

Any consumer can list every coding agent misterspec currently knows how
to install for — each one's stable ID, human-readable name, and target
integration location — without installing anything or touching the
filesystem.

**Why this priority**: Before installing anything for a specific agent, a
consumer (eventually `misterspec init`'s own agent-selection step) needs
to know which agents are actually supported. It's also the smallest
possible slice — pure, side-effect-free enumeration of what the binary
knows about.

**Independent Test**: Can be fully tested by listing the registered
adapters and confirming at least the Claude Code adapter is present, with
its correct ID, name, and target integration path — no fixture project,
no target directory, and no filesystem writes involved at all.

**Acceptance Scenarios**:

1. **Given** the compiled binary's adapter registry, **When** available
   adapters are listed, **Then** each one's ID, name, and target
   integration location are returned, with no filesystem access.
2. **Given** a specific, known adapter ID, **When** that adapter is
   selected by ID, **Then** the same adapter listed in Scenario 1 is
   returned.
3. **Given** an ID that does not match any registered adapter, **When**
   selection is attempted, **Then** a distinct, specific "not found"
   result is returned rather than an arbitrary default adapter.

---

### User Story 2 - Install the Kit for a Selected Adapter (Priority: P2)

Given a target project directory and a selected adapter, any consumer can
materialize the framework's Skill resources into that adapter's own
target integration location, and get back a record of exactly what was
installed — which adapter, which framework version, and where — so a
later step can know what happened without re-deriving it from directory
contents.

**Why this priority**: This is the actual "make this agent aware of
misterspec's Skills" capability a future `misterspec init` will
orchestrate. It builds directly on 005-embedded-kit's proven
materialization guarantees (atomic writes, no silent overwrite) rather
than reimplementing them, and is independently testable against a
fixture set of Skill resources.

**Independent Test**: Can be fully tested by installing a fixture set of
Skill resources for the Claude Code adapter into a fresh temporary
directory, confirming every resource lands at the adapter's target
location byte-for-byte correctly, and that a resulting installation
record accurately names the adapter and location used.

**Acceptance Scenarios**:

1. **Given** a selected adapter and a fresh target directory, **When**
   installation runs, **Then** every Skill resource is materialized at
   the adapter's target integration location, and an installation record
   is produced naming the adapter and that location.
2. **Given** a target directory already fully installed for an adapter,
   **When** installation is run again without requesting overwrite,
   **Then** nothing already present is silently modified.
3. **Given** an install operation that is interrupted partway, **When**
   the target directory is inspected afterward, **Then** no resource is
   found partially written.

---

### User Story 3 - Detect the Currently Installed Adapter (Priority: P3)

Given a project, any consumer can read back which adapter (if any) is
currently installed — without re-installing, and without having to infer
it by guessing from directory contents.

**Why this priority**: This is what lets a future `misterspec init`
re-run recognize "this project already has Claude Code installed" versus
"this project has never been set up" versus "a different agent is
installed here" — informing its own decisions without duplicating
installation logic to find out. It is independently testable against
User Story 2's installation record alone.

**Independent Test**: Can be fully tested by installing an adapter into a
fixture project, then reading back the currently-installed adapter and
confirming it matches exactly what was installed; and by querying a
project that was never installed and confirming a distinct "not
installed" result is returned, not an error.

**Acceptance Scenarios**:

1. **Given** a project with an adapter already installed, **When** the
   currently-installed adapter is queried, **Then** its exact ID and
   target location are returned, matching what installation actually
   produced.
2. **Given** a project that has never had any adapter installed,
   **When** the currently-installed adapter is queried, **Then** a
   distinct "not installed" result is returned, not an error.

---

### Edge Cases

- Selecting an adapter ID that does not match any registered adapter.
- Installing into a target directory that is already fully populated for
  the same adapter, without requesting overwrite.
- An install operation interrupted partway through.
- Querying the currently-installed adapter for a project that has never
  had one installed.
- Listing adapters against an otherwise-empty registry query (no target
  project or directory involved at all).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST maintain a registry of available agent
  adapters, each identified by a stable ID and a human-readable name.
- **FR-002**: System MUST allow a consumer to list every registered
  adapter — its ID, name, and target integration location — without
  installing anything or writing to disk.
- **FR-003**: System MUST allow a consumer to select one registered
  adapter by its ID, returning a distinct, specific "not found" result —
  never an arbitrary default — when the requested ID is not registered.
- **FR-004**: System MUST provide at least one working adapter (Claude
  Code), whose target integration location matches this project's own
  established convention.
- **FR-005**: An adapter's install operation MUST materialize Skill
  resources into the adapter's target integration location under the
  same guarantees already established for the framework's resource
  installer: atomic per-resource writes, and no silent overwrite of an
  already-installed resource.
- **FR-006**: An adapter's install operation MUST produce a record of
  what was installed — the adapter's ID and its target integration
  location, at minimum — as a single, machine-readable result.
- **FR-007**: System MUST allow a consumer to read back a project's
  installation record and report which adapter is currently installed,
  or a distinct "not installed" result — without re-installing anything
  or inferring the answer from directory contents alone.
- **FR-008**: An adapter MUST NOT alter the meaning of any Skill resource
  it installs, MUST NOT modify any project-specific requirement, and MUST
  NOT become an authoritative source of project state — its installation
  record describes what happened; it is not project state itself.
- **FR-009**: An adapter's target integration location MUST be
  discoverable from the adapter itself (via listing, FR-002) — a
  consumer MUST NOT need to already know it in advance.

### Key Entities

- **Adapter**: One registered coding-agent integration — its stable ID,
  human-readable name, and target integration location within a project.
- **Installation Record**: The outcome of one install operation — which
  adapter was installed and where, recorded as a single, readable result.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every registered adapter is discoverable via listing, with
  zero adapters requiring a consumer to already know their target
  location in advance.
- **SC-002**: Installing the same adapter twice without requesting
  overwrite changes nothing already present on the second run, 100% of
  the time.
- **SC-003**: After installation, reading back the installation record
  always reports the exact adapter ID and target location that was
  actually installed.
- **SC-004**: Selecting an unregistered adapter ID is rejected with a
  specific, distinguishable result in 100% of attempts — never silently
  substituting a default adapter.
- **SC-005**: Every Skill resource an adapter installs is byte-for-byte
  identical to its source — zero content is altered by the install
  operation.

## Assumptions

- This feature builds one concrete adapter only: **Claude Code**,
  matching this project's own already-established `.claude/skills`
  integration convention. Additional adapters (Codex, Gemini CLI,
  OpenCode, Antigravity) remain, per
  `docs/architecture-specification.md` §34 itself, "a release decision"
  — out of scope here, added by later features once each is actually
  needed.
- Canonical Skill content (`kit/skills/`) does not exist in this project
  yet — that is Phase 6's job. This feature's install operation is
  therefore built and proven against a fixture set of Skill resources,
  not empty production content; the mechanism wires up to real canonical
  Skills unchanged the moment that content exists, the same relationship
  005-embedded-kit's installer already has with real template content.
- "Install" here still means the adapter-level materialization + record
  primitive only — the full interactive `misterspec init` flow (agent
  selection UI, preview, confirmation) remains out of scope, building on
  top of this feature in a later one (Phase 5).
- This feature builds directly on `005-embedded-kit`'s resource-
  installation guarantees (atomic writes, no silent overwrite) — no new
  external dependency, still no CLI/JSON surface.
- The consumers ("users") of this feature are, in priority order: (1) the
  coding agent and the Skills it runs, indirectly via (2) misterspec's
  own future `misterspec init` command — unchanged from the prior five
  features.
