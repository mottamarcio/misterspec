# Feature Specification: Multi-Agent Skill Integration

**Feature Branch**: `018-multi-agent-skill-integration`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "vamos seguir com a sugestão de trabalhar
na spec 'Phase 10 — Skill Integration' e depois voltamos para a 'phase
9'. Detalhe muito importante para essa fase 10: os adaptar para outros
modelos sem ser claude ainda não foram implementados e precisamos
implementar. Para esse 'mvp', podemos criar adapters para 'agy
(Antigravity)', 'codex (Codex CLI)', 'copilot (GitHub Copilot)',
'cursor-agent (Cursor)' e 'devin (Devin for Terminal)'" (proceed with
Phase 10 — Skill Integration — of
`docs/context-engine-implementation.md`, with the explicit addition the
user called out: this MVP must also deliver the five coding-agent
adapters — Antigravity, Codex CLI, GitHub Copilot, Cursor, and Devin
for Terminal — that 006-agent-adapter left as "a later release
decision" and that were still missing when this feature started;
without them, Skill Integration would only ever reach Claude Code
users)

## User Scenarios & Testing *(mandatory)*

<!--
  Unlike 011-017's own purely internal, agent-facing capabilities, this
  feature has two distinct kinds of "user": (a) a project owner running
  `misterspec init` who wants misterspec's canonical Skills installed
  for whichever coding agent they actually use, and (b) the coding
  agent itself, executing an installed Skill, which should now begin
  its work from the Context Engine's own budgeted Context Pack
  (011-017) rather than re-exploring the project from scratch every
  time. Per context-engine-implementation.md's own Phase 10 scoping,
  the highest-value workflows come first (implementation, planning,
  task creation, analysis); per §23's own explicit warning, the Context
  Pack is a bootstrap optimization, never a sandbox that blocks an
  agent from reading more of the repository when it genuinely needs to.
-->

### User Story 1 - Install misterspec's Skills for More Coding Agents (Priority: P1)

A project owner using Antigravity, Codex CLI, GitHub Copilot, Cursor,
or Devin for Terminal can select their own coding agent during
`misterspec init` and get the same canonical Skills every Claude Code
user already gets, materialized into that agent's own recognized
integration location — without needing to know or care how any other
agent's integration works.

**Why this priority**: None of this feature's other value — Skills
that consume the Context Pack — reaches a single user of these five
agents until the agents themselves can be selected and installed for
at all. This is the literal prerequisite the user calling this out as
"muito importante" is pointing at, and it stands entirely on its own:
it delivers real value (canonical Skills reaching five new audiences)
even before any Skill's own content changes.

**Independent Test**: For each of the five agents, select it during
installation against a fixture project and confirm misterspec's
canonical Skills are materialized into that agent's own correct,
documented integration location, in a form that agent recognizes as
its own custom commands/instructions — without altering what any Skill
means.

**Acceptance Scenarios**:

1. **Given** a project with no coding-agent integration installed yet,
   **When** a supported agent (Antigravity, Codex CLI, GitHub Copilot,
   Cursor, or Devin for Terminal) is selected and installed for,
   **Then** every canonical Skill is materialized into that agent's own
   integration location in a form that agent recognizes.
2. **Given** any one of these five agents has already been installed
   for, **When** the same project is inspected, **Then** installing for
   a second, different supported agent does not remove or corrupt the
   first agent's own installed integration.
3. **Given** any of these five agent IDs, **When** it is listed among
   available adapters, **Then** its stable ID, human-readable name, and
   target integration location are all reported, matching the same
   discovery contract Claude Code's own adapter already provides.
4. **Given** installing for one of these five agents, **When** the
   installation completes, **Then** no Skill's own meaning, required
   inputs, or completion criteria differ from what Claude Code's own
   installed copy of the same Skill says.

---

### User Story 2 - Implementation Begins From the Context Pack (Priority: P2)

When an agent (any of the six now-supported ones) executes the
implementation Skill for a Spec, it requests a Context Pack from the
Context Engine first and begins its work from that pack — the target
Spec, the current Task, the requirements that Task serves, and
directly relevant Knowledge/Learnings — rather than first re-reading
unrelated project files or Knowledge by hand.

**Why this priority**: Implementation is explicitly named as the
highest-value workflow to convert first
(context-engine-implementation.md §30 Phase 10), and it is the Skill
this whole project's "Second Brain" effort was originally motivated by
— the first place a measurable reduction in repeated exploration
actually reaches real work. It depends on User Story 1 only in that an
agent must be installed at all to run any Skill; it does not depend on
any other Skill's own integration.

**Independent Test**: Run the implementation Skill against a Spec with
a known Task, dependency, and linked Knowledge entry; confirm the
Skill's own instructions direct it to request and consume a Context
Pack before any other project exploration, and that doing so still
results in the same correct implementation outcome as before.

**Acceptance Scenarios**:

1. **Given** an executable Task belonging to a Spec, **When** the
   implementation Skill begins, **Then** its own instructions have it
   request a Context Pack for that Spec/Task before reading any other
   project file by hand.
2. **Given** the returned Context Pack already contains the Task's own
   requirement, a relevant dependency, and a linked Knowledge entry,
   **When** the Skill proceeds, **Then** it uses that content directly
   rather than re-discovering the same information through separate
   exploration.
3. **Given** the Context Pack does not contain something the agent
   determines it genuinely needs, **When** the Skill continues,
   **Then** it remains fully free to read further project files or run
   further deterministic operations — the pack never blocks additional
   exploration.
4. **Given** the same Task implemented both with and without this
   change, **When** the two outcomes are compared, **Then** the
   resulting code change and completion criteria are equivalent — this
   feature changes how context is gathered, never what "done" means.

---

### User Story 3 - Planning, Task Creation, and Analysis Begin From the Context Pack (Priority: P3)

The remaining named highest-value workflows — planning, task creation,
and analysis — are updated the same way as User Story 2: each requests
an intent-appropriate Context Pack first and begins from it, before
falling back to broader manual exploration.

**Why this priority**: Extends the same, already-proven pattern (User
Story 2) to the remaining workflows context-engine-implementation.md
§30 Phase 10 explicitly names, completing this feature's own scoped
rollout without yet touching every other Skill in the framework (which
remains future work, per Assumptions).

**Independent Test**: For each of planning, task creation, and
analysis, run the Skill against a fixture Spec and confirm its own
instructions request an appropriately-intended Context Pack first and
consume it, with the same correct outcome as before this change.

**Acceptance Scenarios**:

1. **Given** the planning Skill is invoked for a Spec, **When** it
   begins, **Then** it requests a Context Pack using the planning
   intent before broader exploration.
2. **Given** the task-creation Skill is invoked for a Spec with an
   existing Plan, **When** it begins, **Then** it requests a Context
   Pack using the tasks intent before broader exploration.
3. **Given** the analysis Skill is invoked for a Spec with existing
   Tasks and evidence, **When** it begins, **Then** it requests a
   Context Pack using the validation/analysis intent before broader
   exploration.
4. **Given** any of these three Skills, **When** its own instructions
   are inspected, **Then** they explicitly state the agent remains free
   to expand context beyond the pack when it determines it needs to.

---

### Edge Cases

- A supported agent selected for installation on a project that
  already has a different, unsupported or unrecognized integration
  present — installation still proceeds for the newly selected agent
  without disturbing the unrelated existing content.
- The Context Engine's own internal context command returns an error
  (e.g. an unresolvable target) when a Skill requests a Context Pack —
  the Skill falls back to its own prior, pre-integration exploration
  behavior rather than failing outright.
- A Context Pack returned with `budget_exceeded: true` — the Skill
  still proceeds using the mandatory content it received, aware that
  some optional content was omitted, per the pack's own diagnostics.
- An agent whose own integration format has no direct equivalent for
  one canonical Skill field (for example, no notion of a machine-
  readable "description") — that agent's adapter still installs the
  Skill in the closest form its own convention supports, never
  silently dropping the Skill entirely.
- Two agents installed in the same project, each then executing the
  same Skill independently — both begin from an equivalent Context
  Pack for the same request; neither agent's own integration affects
  the content of the pack the other receives.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support selecting and installing misterspec's
  canonical Skills for each of five additional coding agents:
  Antigravity, Codex CLI, GitHub Copilot, Cursor, and Devin for
  Terminal — in addition to the already-supported Claude Code.
- **FR-002**: Each of these five agents MUST be discoverable the same
  way Claude Code's own adapter already is: a stable ID, a
  human-readable name, and its own target integration location, all
  without requiring any installation to occur first.
- **FR-003**: Installing for any of these five agents MUST materialize
  every canonical Skill into that agent's own recognized integration
  location, in a form that agent's own tooling recognizes as a custom
  command or instruction set, without altering what any Skill means,
  requires, or considers complete.
- **FR-004**: Installing for one agent MUST NOT remove, overwrite, or
  otherwise corrupt another agent's own already-installed integration
  in the same project.
- **FR-005**: The implementation Skill's own instructions MUST direct
  the agent to request a Context Pack (via the existing internal
  context command) for the Spec/Task being worked on before undertaking
  broader, unguided project exploration.
- **FR-006**: The planning, task-creation, and analysis Skills' own
  instructions MUST each be updated the same way, each requesting a
  Context Pack using the intent matching its own workflow.
- **FR-007**: Every updated Skill's own instructions MUST explicitly
  state that the agent remains free to read further project files or
  invoke further deterministic operations beyond the Context Pack when
  it determines it genuinely needs to — the pack MUST NOT be presented
  as a hard boundary on exploration.
- **FR-008**: If a Context Pack request fails for any reason, the
  affected Skill's own instructions MUST direct the agent to fall back
  to that Skill's own prior exploration approach rather than treating
  the failure as blocking.
- **FR-009**: No Skill outside the four named in FR-005/FR-006
  (implementation, planning, task creation, analysis) is changed by
  this feature — every other Skill's behavior remains exactly as it
  was before this feature.
- **FR-010**: This feature MUST NOT change what any Skill considers a
  correct or complete outcome — only how it gathers project context
  before beginning.

### Key Entities

- **Coding Agent Integration**: One of the six now-supported coding
  agents' own way of recognizing custom commands/instructions — an ID,
  a human-readable name, and a target location, mirroring what
  006-agent-adapter already established for Claude Code.
- **Context-Pack-Aware Skill**: One of the four canonical Skills
  (implementation, planning, task creation, analysis) whose own
  instructions now request an intent-appropriate Context Pack before
  broader exploration, while remaining free to exceed it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A project owner using any of the six supported coding
  agents can install misterspec's full canonical Skill set for their
  own agent in one `misterspec init` run, with 100% of Skills reaching
  that agent's own recognized integration location.
- **SC-002**: For a representative set of implementation, planning,
  task-creation, and analysis runs, 100% begin by requesting a Context
  Pack before any other project exploration, verified by inspecting
  each Skill's own instructions.
- **SC-003**: For the same representative set of runs, the final
  outcome (code changed, Plan produced, Tasks produced, or analysis
  produced) is judged equivalent to the outcome produced before this
  feature, by the same completion criteria each Skill already defined.
- **SC-004**: Across every tested installation combination (any subset
  of the six agents installed together in one project), 0 instances of
  one agent's own installed integration being removed or corrupted by
  installing another.
- **SC-005**: 100% of the four updated Skills' own instructions
  explicitly state the agent may expand beyond the Context Pack when
  it needs to, verified by inspection.

## Assumptions

- This feature is scoped to exactly two things the user explicitly
  bundled together: (a) `docs/architecture-specification.md` §34's own
  "exact adapters supported... remain a release decision" — resolved
  here for five specific, named agents (Antigravity, Codex CLI, GitHub
  Copilot, Cursor, Devin for Terminal) — and (b)
  `docs/context-engine-implementation.md`'s own Phase 10 ("Skill
  Integration") for exactly the four highest-value Skills it names
  (implementation, planning, task creation, analysis). It explicitly
  excludes updating every other canonical Skill (e.g. `/create-spec`,
  `/create-feature`, `/create-knowledge-base`) to consume the Context
  Pack, and excludes Phase 9's own dogfooding/evaluation work, both of
  which remain separate, later features, continuing this project's
  established one-phase-at-a-time discipline.
- Each new agent's own exact integration file format/location (for
  example, where Cursor expects its own custom-command files versus
  where Copilot expects its own prompt files) is an implementation
  detail left to planning — this specification requires only that each
  agent's own installed Skills be materialized into a location and
  form that agent's own tooling actually recognizes (FR-003), matching
  006-agent-adapter's own established pattern of "transform canonical
  Skills where necessary" (`docs/architecture-specification.md` §35).
- "Falls back to prior exploration behavior" (FR-008) means each
  updated Skill's own instructions continue to describe what to do
  without a Context Pack — this feature adds a new first step to four
  Skills' own instructions; it does not remove any capability they
  already had.
- The Context Engine itself (011-017) is not modified by this feature
  — Skills become new callers of the already-existing internal context
  command; no new ranking, budgeting, indexing, or retrieval logic is
  introduced here.
- "Canonical Skills" and "Skill" throughout this document refer to the
  same artifacts 006-agent-adapter and 009-canonical-skills-content
  already established (`.misterspec/skills/`, materialized per-agent by
  each adapter) — this feature changes four of their own instruction
  bodies and adds five new adapters, not the underlying Skill model
  itself.
