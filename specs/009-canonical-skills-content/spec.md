# Feature Specification: Canonical Skills Content

**Feature Branch**: `009-canonical-skills-content`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "vamos seguir com a sugestão acima de
'canonical skills content'" (proceed with canonical Skill content for
the framework's MVP skill set, per
`docs/architecture-specification.md` §38-53)

## User Scenarios & Testing *(mandatory)*

<!--
  Unlike 001-008, this feature is primarily semantic content, not
  deterministic code — the coding agent's own instructions for how to
  drive misterspec, not a new capability misterspec itself performs
  (Constitution Principle I: Semantic/Deterministic Separation). Its
  "users" are the coding agent that loads and follows these Skills, and
  the end user who invokes them (e.g. "/create-plan"), once
  006-agent-adapter's Claude Code adapter installs them via
  008-cli-cobra's `misterspec init`.
-->

### User Story 1 - Start a New Project's Knowledge and Memory (Priority: P1)

A user with raw source material (a PRD, existing docs, notes) in a
freshly bootstrapped project can ask their coding agent to build a
structured Knowledge base from it, and then distill that Knowledge into
a project Constitution — the two Skills every later Skill in the
pipeline depends on existing.

**Why this priority**: `docs/architecture-specification.md` §52's
canonical next-step graph starts here — nothing else in the pipeline is
meaningful without a project's Knowledge and Constitution existing
first. It is also the smallest independently valuable and independently
testable slice: a coding agent equipped with just these two Skills
already turns raw documents into durable, structured project memory,
regardless of whether any later-stage Skill exists yet.

**Independent Test**: Bootstrap a fresh project, place a sample document
under its raw-sources directory, invoke `/create-knowledge-base` and
confirm it produces one or more well-formed Knowledge artifacts (using
only the deterministic operations already available — inventory,
fingerprint, create, inspect, validate); then invoke
`/create-constitution` and confirm it produces a Constitution
referencing that Knowledge — both without any later Skill installed.

**Acceptance Scenarios**:

1. **Given** a bootstrapped project with raw source documents present,
   **When** the user invokes the knowledge-base Skill, **Then** the
   agent inventories and fingerprints the sources, creates one Knowledge
   artifact per identified topic via the project's own entity-creation
   capability, and reports a completion summary naming what was created
   and any unresolved conflicts or unknowns.
2. **Given** a project with an established Knowledge base, **When** the
   user invokes the constitution Skill, **Then** the agent produces a
   Constitution capturing durable invariants distinct from ordinary
   facts, and reports completion with a recommended next step.
3. **Given** a project with no raw source documents yet, **When** the
   user invokes the knowledge-base Skill, **Then** the agent reports
   that nothing was found to process rather than fabricating Knowledge
   content.

---

### User Story 2 - Decompose a Project into Programs, Features, and Specs (Priority: P2)

Once a project has Knowledge and a Constitution, a user can ask their
coding agent to structure the actual work: define a Program, break it
into Features, and write Specs for each Feature — the three Skills that
turn project memory into a concrete, addressable backlog.

**Why this priority**: This is the next pipeline stage after User Story
1 and the one that makes 003-entity-creation's Program/Feature/Spec
creation capability actually reachable through a Skill rather than only
through a raw internal command. It is independently testable against
001-008's already-proven entity-creation and validation operations,
without depending on User Story 1's two Skills actually being invoked
first (a project can already have Knowledge/Constitution content
seeded directly for the purpose of testing this story).

**Independent Test**: Against a bootstrapped project (with or without
real Knowledge-base/Constitution Skill invocations having occurred),
invoke `/create-program`, then `/create-feature` for the resulting
Program, then `/create-specs` for the resulting Feature, confirming
each produces a correctly parented, structurally valid entity and a
completion summary naming its ID and recommended next step.

**Acceptance Scenarios**:

1. **Given** a project ready for a new initiative, **When** the user
   invokes the program Skill, **Then** the agent creates one Program
   entity capturing problem, outcome, and scope, and reports its ID.
2. **Given** an existing Program, **When** the user invokes the feature
   Skill for it, **Then** the agent creates one or more Feature entities
   parented to that Program, each with a clear capability boundary.
3. **Given** an existing Feature, **When** the user invokes the specs
   Skill for it, **Then** the agent creates one or more Spec entities
   parented to that Feature, each with testable requirements and
   acceptance scenarios.
4. **Given** a request that would duplicate an existing Program's
   already-covered problem space, **When** the program Skill is
   invoked, **Then** the agent surfaces the overlap rather than silently
   creating a redundant Program.

---

### User Story 3 - Plan, Break Down, Implement, and Verify a Spec (Priority: P3)

Once a Spec exists, a user can ask their coding agent to design an
implementation plan for it, break that plan into executable tasks,
implement them against the real codebase, and verify the result against
the Spec's own requirements — the four Skills that carry one unit of
work from design through delivered, checked reality.

**Why this priority**: This is the pipeline's final, highest-context
stage — each of these four Skills depends on a real Spec (User Story 2)
existing to operate against, and `/implement`/`/analyze` in particular
only produce meaningful results once there is actual repository code to
work with. It is the natural last increment: valuable and testable on
its own once a Spec exists, but it is the stage every earlier one exists
to feed.

**Independent Test**: Against an existing Spec (seeded directly for
testing, independent of whether User Story 1/2's Skills produced it),
invoke `/create-plan`, then `/create-tasks`, then `/implement` for one
resulting task, then `/analyze`, confirming each produces its documented
artifact or code change and a completion summary — including `/analyze`
correctly reporting a gap when implementation is deliberately left
incomplete, and correctly reporting success when it is not.

**Acceptance Scenarios**:

1. **Given** a Spec with no Plan yet, **When** the user invokes the plan
   Skill, **Then** the agent produces a Plan artifact mapping the Spec's
   requirements to a concrete technical strategy.
2. **Given** a Spec with a Plan, **When** the user invokes the tasks
   Skill, **Then** the agent produces a Tasks artifact decomposing the
   Plan into ordered, verifiable units of work.
3. **Given** a Spec with Tasks, **When** the user invokes the implement
   Skill for one task, **Then** the agent makes the corresponding code
   change and reports what changed and how it was verified.
4. **Given** a Spec whose implementation is complete and correct,
   **When** the user invokes the analyze Skill, **Then** the agent
   produces a Validation artifact confirming the Spec's requirements are
   met.
5. **Given** a Spec whose implementation is incomplete or diverges from
   its requirements, **When** the user invokes the analyze Skill,
   **Then** the agent reports specifically which requirement is unmet
   and which artifact layer (Spec, Plan, or implementation) is
   responsible — never a generic pass/fail with no explanation.

---

### Edge Cases

- A Skill invoked against a project that isn't a misterspec project at
  all (must fail clearly, the same `project_not_initialized` condition
  every internal command already reports — not a confusing partial
  attempt).
- A Skill invoked with a target ID that doesn't exist (must fail
  clearly, reusing the project's existing not-found result rather than
  guessing or fabricating).
- Two Skills invoked back-to-back in the same project (e.g. two Features
  created in quick succession) must not corrupt or collide on ID
  allocation — reusing the project's already-proven concurrency
  protection, not a new one.
- A Skill's completion summary when nothing meaningful happened (e.g.
  no raw sources to process, no anomalies found) must still produce the
  required completion concepts, not silently produce no output.
- A coding agent following a Skill's instructions attempts an action the
  Skill's own stated Authority does not permit (must be refused by the
  Skill's own stated rules, not merely discouraged).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide exactly the nine canonical Skills
  named in the framework's specification: create-knowledge-base,
  create-constitution, create-program, create-feature, create-specs,
  create-plan, create-tasks, implement, analyze.
- **FR-002**: Every Skill MUST document, at minimum: its purpose, what
  triggers it, what it is responsible for, its expected inputs and
  outputs, its preconditions, what context it needs (required,
  optional, and explicitly unnecessary), what it is authorized to read,
  create, and modify, what it is forbidden from mutating, which of the
  project's existing deterministic operations it is expected to use,
  its step-by-step procedure, its decision rules, its validation rules,
  its failure and stop conditions, its success criteria, its
  postconditions, whether re-running it is safe, how it resumes
  partial work, and its completion contract.
- **FR-003**: Every Skill's documented deterministic-operations section
  MUST name only operations the project's command-line interface
  already provides — never an operation that does not yet exist.
- **FR-004**: Every Skill's completion summary MUST report, at minimum:
  the outcome, artifacts produced or changed, important findings,
  unresolved issues needing attention, and a recommended next step
  (naming the specific next Skill to invoke, when one applies).
- **FR-005**: A Skill's recommended next step MUST reflect the project's
  actual current state rather than a fixed, always-the-same suggestion —
  including recommending returning to an earlier-stage Skill when the
  gap it identifies belongs there instead of the next stage forward.
- **FR-006**: Every Skill MUST be installable for a supported coding
  agent using the project's existing agent-installation capability, and
  MUST be discoverable by that agent using its own native Skill-loading
  convention once installed.
- **FR-007**: Every Skill's stated Authority MUST distinguish what it
  may read, what entities/artifacts it may create, what it may modify,
  and what mutations are explicitly forbidden to it — consistent with
  the project's existing mutation-boundary guarantees.
- **FR-008**: No Skill MUST perform, or instruct an agent to perform,
  any mutation the project's existing deterministic operations do not
  already support (e.g. inventing an entity ID by hand, guessing a
  canonical path) when an equivalent deterministic operation exists.
- **FR-009**: The set of installed Skills MUST NOT alter or duplicate
  any existing deterministic operation's behavior — Skills are
  instructions for an agent, never a second implementation of anything
  001-008 already built.

### Key Entities

- **Skill**: One canonical, semantic unit of agent guidance — a named
  capability (e.g. "create-plan") with documented purpose, authority,
  procedure, and completion contract, distinct from any deterministic
  operation it calls.
- **Completion Summary**: The required, structured report a Skill
  invocation ends with — outcome, artifacts, findings, unresolved
  issues, and recommended next step.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All nine canonical Skills are present, installable, and
  discoverable by the project's supported coding agent, with zero
  manual setup steps beyond the project's existing bootstrap command.
- **SC-002**: A user can carry one unit of work from a raw source
  document through a verified implementation using only these nine
  Skills in sequence, with zero direct internal command invocation
  required from the user themselves.
- **SC-003**: Every Skill's completion summary names a recommended next
  step 100% of the time, and that recommendation matches the project's
  actual current state (not a fixed script) in every acceptance
  scenario tested.
- **SC-004**: 100% of each Skill's documented deterministic operations
  are operations the project's command-line interface can actually
  execute — zero references to a not-yet-existing operation.
- **SC-005**: Attempting a mutation outside a Skill's own stated
  Authority is refused 100% of the time in every scenario tested.

## Assumptions

- This feature is semantic content — Skill instructions for a coding
  agent — not new deterministic Go behavior. Where a canonical Skill's
  illustrative operation contract in
  `docs/architecture-specification.md` (§41-49) names an operation the
  project has not yet built (a `references` lookup; Task ID allocation
  separate from Program/Feature/Spec/Knowledge/Learning creation), that
  Skill's content uses only what already exists today (e.g. an
  entity's already-available `depends_on`/`supersedes` metadata in
  place of a dedicated `references` lookup; an agent authoring Task
  entries directly into a Spec's Tasks artifact, the same
  human/agent-authored boundary 003-entity-creation's own Create
  already drew for Task) — not a signal to build the missing operation
  as part of this feature. FR-003/SC-004 make this an explicit,
  verifiable requirement rather than a silent gap.
- This feature does not include the interactive Bubble Tea flow or any
  change to `misterspec init`'s own behavior (008-cli-cobra) — it only
  supplies the Skill content `kit.SkillsFS` installs, replacing that
  feature's placeholder.
- This feature does not add a second coding-agent adapter — Claude Code
  (006-agent-adapter) remains the only one Skills are verified against.
- Skill content is authored once, for the framework's own MVP scope; it
  is not project-specific and carries no project's own domain
  knowledge.
- The consumers ("users") of this feature are, in priority order: (1)
  the coding agent that loads and follows a Skill's instructions,
  indirectly via (2) the end user who invokes it — the same framing
  every prior feature has used, now describing the layer real end users
  actually interact with day to day.
