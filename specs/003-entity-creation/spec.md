# Feature Specification: Atomic Entity Creation

**Feature Branch**: `003-entity-creation`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "seguir com a recomendação acima" (proceed
with the mutation layer recommended after 002-read-operations: atomic
`create` for independently-identified entities, `create-artifact` for
Plan/Tasks/Validation, protected by a short-lived recoverable lock)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-core-foundation and 002-read-operations, this feature has no
  end-user UI. Its "users" are the coding agent and misterspec's own
  future CLI layer, which — per docs/architecture-specification.md §10-12,
  §57-59 — must call an atomic creation primitive instead of allocating an
  ID and scaffolding a file as two separate, interruptible steps. This is
  the first feature in this project's lifecycle that writes to the
  filesystem; every story is independently testable against fixture
  projects, without any CLI/JSON surface existing yet.
-->

### User Story 1 - Atomically Create a New Entity (Priority: P1)

Given an entity type (Program, Feature, Spec, Knowledge, or Learning) and,
where that type requires one, a parent ID, any consumer can create a new
instance of it as a single all-or-nothing operation: the next available ID
is allocated, its canonical directory is created, and its initial artifact
file is written with every required frontmatter field populated from that
type's fixed template — or none of that happens at all.

**Why this priority**: This is the one capability every "create a Program"
/ "create a Feature" / "create a Spec" / "add to the knowledge base" /
"record a Learning" Skill depends on. Without it, nothing in the project's
lifecycle beyond read-only inspection can happen. It is also the riskiest
capability in the project so far — the first to write files — so it is
the anchor the other two stories build safety around.

**Independent Test**: Can be fully tested by creating a fixture project
and requesting creation of each supported entity type (including a
Feature under a valid Program, and one under a Program ID that does not
exist) and confirming: the ID allocated is exactly one more than the
highest existing ID of that type; the resulting file passes structural
inspection immediately; and an invalid parent is rejected with nothing
written to disk — with no other story's code required.

**Acceptance Scenarios**:

1. **Given** a project with no existing Specs under a Feature, **When** a
   new Spec is created for that Feature, **Then** it is allocated ID
   `SPEC-001`, its canonical directory and `spec.md` are created, and its
   frontmatter has every required field for a Spec populated.
2. **Given** a project with Specs `SPEC-001` and `SPEC-002` already
   existing under a Feature, **When** another Spec is created for the same
   Feature, **Then** it is allocated `SPEC-003`.
3. **Given** a request to create a Feature under a Program ID that does
   not exist, **When** creation is attempted, **Then** it is rejected with
   a specific "invalid parent" result and no directory or file is created.
4. **Given** a request to create an entity type whose canonical directory
   would already exist, or whose target file already exists, **When**
   creation is attempted, **Then** it is rejected rather than overwriting
   what is there.
5. **Given** a newly created entity of any supported type, **When** its
   artifact is inspected immediately afterward, **Then** it parses
   successfully with no structural errors.

---

### User Story 2 - Create Subordinate Spec Artifacts (Priority: P2)

Given an existing Spec's ID, any consumer can create its Plan, Tasks, or
Validation artifact at that artifact's fixed, predetermined location from
the appropriate template — without allocating any new independent ID,
since these artifacts are identified only by the Spec they belong to.

**Why this priority**: This is what `/create-plan`, `/create-tasks`, and
`/analyze` depend on to produce their subordinate documents. It builds on
User Story 1's atomic-write and template machinery but is scoped
narrowly enough to test independently against a fixture Spec that already
exists.

**Independent Test**: Can be fully tested against a fixture project with
one existing Spec — requesting a Plan, then a Tasks artifact, then a
Validation artifact for it, confirming each lands at its exact
predetermined path with the correct initial frontmatter, and that
requesting a Plan for a nonexistent Spec, or a second Plan for a Spec that
already has one, is rejected.

**Acceptance Scenarios**:

1. **Given** an existing Spec with no Plan yet, **When** a Plan is
   created for it, **Then** `plan.md` is written at that Spec's fixed
   location with `for:` correctly naming the Spec.
2. **Given** a Spec ID that does not exist, **When** a Tasks artifact is
   requested for it, **Then** creation is rejected with a specific
   "invalid parent" result.
3. **Given** a Spec that already has a Validation artifact, **When**
   another Validation artifact is requested for the same Spec, **Then**
   creation is rejected rather than overwriting the existing one.

---

### User Story 3 - Safe, Recoverable Concurrency Protection (Priority: P3)

Any creation request (of either kind, from either story above) is
protected by a short-lived mechanism that guarantees two requests
submitted in close succession for the same entity type never receive the
same ID or corrupt each other's output — and if a previous creation
attempt was interrupted (for example, its process crashed) partway
through, the very next creation request recovers automatically, without
any manual cleanup step.

**Why this priority**: Correctness under concurrency and interruption is
what makes User Story 1 and 2 trustworthy rather than merely
"usually fine." It is independently testable by simulating overlapping
requests and a pre-existing stale protection marker, on top of the
creation machinery already built.

**Independent Test**: Can be fully tested by issuing two creation requests
for the same entity type in immediate succession and confirming they
receive different, sequential IDs with no corrupted output from either;
and by pre-placing a stale protection marker (simulating a crashed prior
attempt) and confirming the next creation request still succeeds without
manual intervention.

**Acceptance Scenarios**:

1. **Given** two requests to create a Spec under the same Feature,
   submitted immediately after one another, **When** both complete,
   **Then** they hold different, sequential IDs and both artifacts are
   fully and correctly written.
2. **Given** a leftover protection marker from a previous creation attempt
   that never completed (e.g. the process was killed), **When** a new
   creation request is made, **Then** it succeeds without requiring any
   manual removal of that leftover marker.
3. **Given** the protection mechanism is currently held by a creation
   request, **When** an overlapping request for the same entity type
   arrives, **Then** it waits for or is rejected in favor of the first
   request — it is never allowed to proceed and allocate a colliding ID.

---

### Edge Cases

- Two near-simultaneous requests to create a Spec under the same Feature.
- Creating a Feature under a Program ID that does not exist.
- Creating a Spec whose declared Feature parent exists but actually
  belongs to a different Program than the one implied by the request.
- A stale protection marker left behind by a previous crashed or
  interrupted creation attempt.
- Requesting a Plan, Tasks, or Validation artifact for a Spec that already
  has one.
- Requesting creation of an entity type this feature does not support.
- A creation request whose computed canonical path would fall outside the
  project root.
- A write failure partway through creating an artifact (e.g. disk full) —
  must never leave a partially written file at the canonical path.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allocate a new entity's ID and create its
  canonical directory and initial artifact file as a single atomic
  operation — never as separate steps a caller or a second concurrent
  request could interleave with.
- **FR-002**: The allocated ID MUST be the highest existing numeric ID of
  that entity type, plus one, determined by scanning existing artifacts —
  never a stored counter.
- **FR-003**: System MUST write the new artifact's initial content from
  that entity type's fixed template, with every field required by that
  type's schema populated.
- **FR-004**: System MUST reject creation with a specific, distinguishable
  result when the specified parent does not exist, is not the correct
  type, or is missing where required — and must write nothing to disk in
  that case.
- **FR-005**: System MUST guarantee that two creation requests for the
  same entity type, submitted in close succession, never receive the same
  ID.
- **FR-006**: System MUST protect creation with a short-lived mechanism
  that is not itself authoritative project state, and that recovers
  automatically — without manual intervention — if a previous creation
  attempt was interrupted before completing.
- **FR-007**: System MUST write every new artifact file atomically: a
  reader must never observe a partially written file, and an interrupted
  or failed creation must never leave an empty, truncated, or corrupted
  file at the canonical path.
- **FR-008**: System MUST reject any creation whose target canonical path
  would resolve outside the project root, or would overwrite an artifact
  that already exists at that path.
- **FR-009**: System MUST support creating each of the following as an
  independently-identified entity, each requiring its own new ID:
  Program, Feature, Spec, Knowledge, Learning.
- **FR-010**: System MUST support creating each of the following as a
  subordinate artifact of an existing Spec, at that artifact's fixed,
  predetermined location — without allocating any new independent ID for
  it: Plan, Tasks, Validation.
- **FR-011**: System MUST reject creating a subordinate artifact for a
  Spec ID that does not exist, with a specific "invalid parent" result.
- **FR-012**: System MUST reject creating a subordinate artifact that
  already exists at its target path, rather than silently overwriting it.
- **FR-013**: Every successfully created artifact MUST be immediately and
  correctly readable by this project's existing inspection capability —
  its structured metadata parses without error the moment creation
  succeeds.
- **FR-014**: System MUST report a specific, distinguishable reason for
  every rejected creation (invalid parent, invalid or unsupported type,
  target already exists, path outside project, protection-mechanism
  failure) rather than a single generic failure.

### Key Entities

- **Creation Request**: The input to a creation operation — the entity or
  artifact type requested, its parent (where applicable), and any
  type-specific input needed to populate its template.
- **Creation Result**: The outcome of a successful creation — the new
  entity's ID (where it has one) and its canonical path.
- **Creation Protection**: A short-lived, non-authoritative mechanism held
  only for the duration of one creation operation, guaranteeing no two
  concurrent creations of the same entity type collide, and safely
  recoverable if abandoned by an interrupted attempt.
- **Artifact Template**: The fixed initial content and required
  frontmatter shape for one entity or artifact type, used to populate a
  newly created artifact.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every created entity's ID is unique across the entire
  project, 100% of the time, including when creation requests occur in
  rapid succession.
- **SC-002**: No creation attempt ever leaves a partially created entity —
  an ID allocated with no corresponding file, or a file with incomplete or
  corrupted content — across all supported entity and artifact types.
- **SC-003**: A creation request interrupted mid-operation never
  permanently blocks subsequent creation requests; the very next attempt
  succeeds without any manual recovery step.
- **SC-004**: 100% of newly created artifacts pass structural inspection
  immediately after creation, with zero manual fix-up required.
- **SC-005**: 100% of invalid creation requests (missing or wrong-type
  parent, duplicate target, path outside the project) are rejected, never
  silently accepted or partially applied.

## Assumptions

- This feature builds directly on `001-core-foundation`'s and
  `002-read-operations`'s packages (project detection/configuration,
  canonical path resolution, ID parsing/scanning, metadata parsing, and
  the resolve/inspect operations used to validate a declared parent
  exists and is the correct type).
- This feature defines and embeds only the specific artifact body
  templates it needs (Program, Feature, Spec, Knowledge, Learning, Plan,
  Tasks, Validation) — it does not implement the broader "Embedded Kit"
  system (canonical Skills, agent integration templates) that
  `misterspec init` will need; that remains a distinct, later feature.
- Still no CLI or JSON output surface — this spec covers the internal Go
  creation layer only, the same scoping `001-core-foundation` and
  `002-read-operations` used.
- Knowledge and Learning artifact filenames include an agent-chosen slug
  (`docs/architecture-specification.md` §61); this feature accepts that
  slug as part of the creation request rather than inventing one, since
  choosing a meaningful slug is a semantic decision that belongs to the
  agent, not this deterministic layer.
- The "protection mechanism" (Creation Protection) is a filesystem-level
  concept, not a network or multi-machine coordination mechanism — this
  project remains a single local repository worked on by one active
  workflow at a time, consistent with `docs/architecture-specification.md`
  §58's MVP concurrency scope.
- The consumers ("users") of this feature are, in priority order: (1) the
  coding agent and the Skills it runs, indirectly via (2) misterspec's own
  future deterministic-operations CLI layer, which is this feature's
  direct caller — unchanged from the prior two features.
