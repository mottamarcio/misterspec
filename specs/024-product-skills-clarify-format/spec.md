# Feature Specification: Interactive Clarification and Richer Output for Product Skills

**Feature Branch**: `024-product-skills-clarify-format`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "criar um nova spec para as sugestões acima [do relatório comparando o spec-kit com os Skills do produto misterspec]. Vale pontuar que em vez de fazer inferencias e resolver problemas sozinho, sempre perguntar para o usuário o que ele deseja fazer (dando uma lista de opções) e adicionando 'recomendado' em alguma sugestão que a IA julgar melhor. Acho que podemos aproveitar para melhorar/enriquecer o formatação output (talvez usando tabela, bullet points, etc) do resumo do que foi feito para cada 'slash command'"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ambiguous requirements get asked about, not silently resolved (Priority: P1)

A user runs `/create-specs` against a Feature whose scope leaves a
requirement's exact boundary genuinely ambiguous. Today, the Skill
resolves this on its own — narrowing the requirement and quietly
recording an open question — without ever asking the user what they
actually want. Instead, the Skill should stop and ask, presenting
concrete options (with one clearly marked as recommended) rather than
guessing.

**Why this priority**: This is the most direct fix for a confirmed gap
in the product's own core Skill, and it's exactly what the user asked
for: never infer silently when a real choice with different
implications exists — always surface it.

**Independent Test**: Run `/create-specs` against a Feature with at
least one requirement whose boundary is genuinely ambiguous; confirm
the Skill presents the ambiguity as a question with 2-4 concrete
options (one marked recommended) and waits for the user's choice before
writing the Spec, rather than silently narrowing and moving on.

**Acceptance Scenarios**:

1. **Given** a requirement boundary is genuinely ambiguous (multiple
   reasonable interpretations with materially different implications),
   **When** `/create-specs` reaches that requirement, **Then** it
   presents the ambiguity as a clear question, with concrete answer
   options and one marked "(Recommended)" along with a one-sentence
   reason, and waits for the user's answer before finalizing that
   requirement.
2. **Given** the user answers, **When** `/create-specs` continues,
   **Then** the chosen answer is reflected directly in the written
   requirement — never left as a vague placeholder plus a silent
   annotation.
3. **Given** a requirement's boundary has an obvious, low-stakes default
   (not a genuine fork with materially different implications), **When**
   `/create-specs` reaches it, **Then** it applies the default and
   records the assumption, without interrupting the user — this
   behavior is unchanged from today, per the existing quota and
   prioritization rules (scope > security/privacy > user experience >
   technical detail; a small, bounded number of questions per run).
4. **Given** a question was never asked (because the interactive quota
   was reached) but a real ambiguity remains, **When** the Spec is
   written, **Then** it is recorded under the Spec's own existing
   `## Unresolved Questions` section — never silently dropped.

---

### User Story 2 - Planning-level forks get asked about, not silently resolved (Priority: P1)

A user runs `/create-plan` against a Spec whose implementation strategy
has more than one reasonable option with materially different
tradeoffs (e.g. two genuinely different architectural approaches).
Today, the Skill picks one and records the tradeoff as a note in Risks
or Assumptions, without asking. Instead, it should ask, the same way
User Story 1 fixes `/create-specs`.

**Why this priority**: Same value and same mechanism as User Story 1,
just at the planning stage — bundled at the same priority since it's
the same fix applied a second time.

**Independent Test**: Run `/create-plan` against a Spec whose
implementation has a genuine strategic fork; confirm the Skill presents
it as a question with concrete options (one recommended) instead of
silently choosing.

**Acceptance Scenarios**:

1. **Given** a Plan decision has more than one reasonable strategy with
   materially different tradeoffs, **When** `/create-plan` reaches that
   decision, **Then** it presents the fork as a question with concrete
   options and one marked "(Recommended)" with a one-sentence reason,
   and waits for the user's answer.
2. **Given** the user answers, **When** the Plan is written, **Then**
   the chosen strategy is reflected directly, not left as a hedge across
   multiple options.
3. **Given** a Plan decision has only one reasonable approach (no
   genuine fork), **When** `/create-plan` reaches it, **Then** it
   proceeds without interrupting the user — unchanged from today.

---

### User Story 3 - Tasks quote a requirement's own constraint, not just its ID (Priority: P2)

A user runs `/create-tasks` against a Spec whose requirements include
explicit constraints (a specific limit, format, or measurable
threshold). Today, a Task only references the requirement by ID
(`SPEC-###:R#`) — the constraint's own text has to be re-read from the
Spec later, at implementation time, risking drift or reinterpretation.

**Why this priority**: A real, previously-identified drift risk, but
narrower in scope than User Stories 1-2 (a wording fix to one Skill's
own task-writing step).

**Independent Test**: Run `/create-tasks` against a Spec with at least
one requirement that carries an explicit constraint; confirm the
resulting Task's own description quotes that constraint verbatim, not
just the requirement's ID.

**Acceptance Scenarios**:

1. **Given** a Requirement states an explicit constraint (e.g. a
   specific limit, required format, or measurable threshold), **When**
   `/create-tasks` writes a Task touching that Requirement, **Then** the
   Task's own description quotes the constraint's exact text, in
   addition to (not instead of) the `SPEC-###:R#` reference.
2. **Given** a Requirement carries no explicit constraint (purely
   qualitative), **When** `/create-tasks` writes a Task touching it,
   **Then** behavior is unchanged from today (ID reference only).

---

### User Story 4 - Analyze can offer to track a remaining gap as a Task, with the user's confirmation (Priority: P3)

A user runs `/analyze` and it finds that the codebase does not yet
satisfy a requirement — today, `/analyze` can only recommend re-running
`/implement` in prose, with no tracked record of exactly what's
missing. Instead, when `/analyze` finds a gap whose responsible layer is
"implementation incomplete," it should offer to append a new Task
naming that specific gap — asking the user first, never writing on its
own initiative.

**Why this priority**: The most valuable remaining gap from the prior
comparison, but also the most sensitive — it requires carefully,
narrowly widening what `/analyze` is allowed to write, so it ships last
and only with an explicit, always-ask confirmation step, never an
automatic write.

**Independent Test**: Run `/analyze` against a feature with a known,
deliberately introduced implementation gap; confirm it presents the gap
and asks whether to append a tracking Task (with a recommended answer),
appends exactly one new Task only on explicit confirmation, and makes
no other change to any existing artifact.

**Acceptance Scenarios**:

1. **Given** `/analyze` finds a requirement whose responsible layer is
   "implementation incomplete," **When** it reports that finding,
   **Then** it also asks the user whether to append a new Task tracking
   that specific gap, presenting the choice with a recommended option
   and a one-sentence reason.
2. **Given** the user confirms, **When** `/analyze` proceeds, **Then**
   exactly one new Task is appended to the existing tasks.md, naming the
   specific gap and its evidence, and no existing Task, Spec, or Plan
   content is modified.
3. **Given** the user declines, **When** `/analyze` proceeds, **Then**
   no write occurs at all — the finding stays as a prose recommendation,
   exactly like today.
4. **Given** `/analyze` finds a gap whose responsible layer is the Spec
   or Plan itself (not implementation), **When** it reports that
   finding, **Then** it does not offer to append a Task — this
   escape hatch is scoped only to "implementation incomplete" findings,
   never to gaps that call for rewriting a Spec or Plan.

---

### User Story 5 - Every slash command's completion summary is scannable, not a wall of prose (Priority: P1)

A user runs any of misterspec's own slash commands (`/create-program`,
`/create-feature`, `/create-specs`, `/create-plan`, `/create-tasks`,
`/implement`, `/analyze`, `/create-constitution`,
`/create-knowledge-base`). Today, each Skill's own completion report is
a paragraph of prose. Instead, each should present what was
created/changed using structured formatting (tables and/or bullet
lists) so the user can scan the outcome at a glance.

**Why this priority**: Applies to every Skill uniformly and directly
improves the experience of every single command invocation — high
value, and low technical risk (formatting only, no behavior change).

**Independent Test**: Run any of the nine Skills to completion; confirm
its own completion report uses a table and/or bullet list to summarize
what was created or changed, rather than a single prose paragraph.

**Acceptance Scenarios**:

1. **Given** any of the nine product Skills completes its work, **When**
   it reports completion, **Then** the report presents the artifact(s)
   created or modified (ID, type, path, and relevant status) as a table
   or bullet list, not embedded in a prose paragraph.
2. **Given** a Skill's completion involves more than one distinct kind
   of information (e.g. `/implement`'s own per-Task outcomes plus an
   overall status), **When** it reports, **Then** each kind of
   information gets its own clearly labeled table or list, not
   interleaved prose.
3. **Given** a Skill fails or stops early, **When** it reports that
   outcome, **Then** the same structured-formatting expectation applies
   to the failure/stop report as to a success report.

---

### Edge Cases

- What happens when an "ambiguous requirement" (User Story 1) and a
  "planning-level fork" (User Story 2) both occur in the same
  `/create-specs`→`/create-plan` sequence, exceeding the existing
  per-run question quota shared across both stages? The existing
  prioritization rule (scope > security/privacy > user experience >
  technical detail) still governs which questions get asked
  interactively; anything past the quota is recorded under the
  relevant artifact's own existing unresolved-questions section, per
  User Story 1's own Acceptance Scenario 4 — never silently dropped.
- What happens if the user's answer to a presented question doesn't map
  to any offered option? The Skill asks for a quick disambiguation
  (consistent with existing interaction conventions elsewhere in this
  project) rather than guessing or discarding the question.
- What happens when `/analyze` (User Story 4) finds multiple
  "implementation incomplete" gaps in the same run? Each is presented
  and confirmed independently — the user may accept some and decline
  others, and only confirmed gaps get a Task appended.
- What happens to a Skill's completion report (User Story 5) when there
  is genuinely nothing to tabulate (e.g. zero artifacts touched)? The
  Skill states that plainly rather than rendering an empty table.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `/create-specs` MUST present a genuinely ambiguous
  requirement boundary as a question with concrete answer options
  (2-4 options) and MUST mark exactly one option as recommended with a
  one-sentence reason, rather than silently narrowing the requirement
  and recording an assumption.
- **FR-002**: `/create-specs` MUST continue applying its own existing
  default-and-annotate behavior for genuinely low-stakes ambiguity that
  does not rise to a real fork with materially different implications —
  User Story 1 changes what happens for genuine forks only.
- **FR-003**: `/create-specs` MUST record any real ambiguity that goes
  unasked (interactive quota reached) under the Spec artifact's own
  existing `## Unresolved Questions` section.
- **FR-004**: `/create-plan` MUST present a genuine planning-level
  strategic fork (more than one reasonable option with materially
  different tradeoffs) as a question with concrete options and one
  marked recommended, rather than silently choosing and recording the
  tradeoff as a note.
- **FR-005**: `/create-tasks` MUST quote a Requirement's own explicit
  constraint (a specific limit, required format, or measurable
  threshold) verbatim in the description of any Task touching that
  Requirement, in addition to the existing `SPEC-###:R#` reference.
- **FR-006**: `/analyze` MUST, for any finding whose responsible layer
  is "implementation incomplete," ask the user whether to append a
  tracking Task for that specific gap, presenting the choice with a
  recommended option, before making any write.
- **FR-007**: `/analyze` MUST append a new Task only on the user's
  explicit confirmation, and that append MUST be its only write — no
  existing Task, Spec, or Plan content may be modified as part of this
  behavior.
- **FR-008**: `/analyze` MUST NOT offer to append a Task for a finding
  whose responsible layer is the Spec or Plan itself.
- **FR-009**: Every one of misterspec's own nine product Skills MUST
  present its completion (and failure/stop) report using structured
  formatting — a table and/or bullet list naming the specific
  artifact(s) affected — rather than a single prose paragraph.

### Key Entities

- **Clarification Question** (new concept, not a persisted entity): an
  in-session question presented during `/create-specs` or
  `/create-plan`, with 2-4 options, one marked recommended: not written
  to any artifact directly — only its resolution is (either integrated
  into the requirement/decision, or recorded under the artifact's own
  existing unresolved-questions section when deferred).
- **Task** (existing entity, gains one new writer under a narrow
  condition): `/analyze` may now append one new Task line, under
  User Story 4's own explicit-confirmation gate — no schema change.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `/create-specs` against a Feature with a genuinely
  ambiguous requirement results in the user being asked, with a
  recommended option, in 100% of such cases — never a silent
  resolution.
- **SC-002**: Running `/create-plan` against a Spec with a genuine
  strategic fork results in the same — the user is asked, with a
  recommended option, in 100% of such cases.
- **SC-003**: A Task written by `/create-tasks` for a constrained
  Requirement always contains that constraint's own verbatim text,
  verified across multiple Specs with different constraint types (limit,
  format, threshold).
- **SC-004**: `/analyze` never appends a Task without the user's prior,
  explicit confirmation for that specific finding.
- **SC-005**: Every one of the nine product Skills' own completion
  reports contains at least one table or bullet list summarizing the
  affected artifact(s), verified by running each Skill at least once.

## Assumptions

- Verified during specification (not implementation): the Spec artifact
  template (`kit/templates/spec.md.tmpl`) already has an
  `## Unresolved Questions` section — the prior comparison report's
  claim that only the Feature artifact has an open-questions field was
  incorrect. No new section or template change is needed for User Story
  1's own deferred-question recording; it reuses what already exists.
- The interactive question mechanics (recommended-option format,
  disambiguation-on-mismatch, per-run quota and prioritization) reuse
  the same proven pattern already adopted for this project's own
  dev-tooling (spec 023's `speckit-clarify` question-quality rules) —
  no new interaction pattern is being invented, only applied to two
  more Skills (`create-specs`, `create-plan`) and, narrowly, `analyze`'s
  own new confirmation step.
- User Story 4's scope is deliberately narrow: only "implementation
  incomplete" findings, only Task-appending, only with explicit
  confirmation. Broader remediation (auto-fixing code, rewriting a
  Spec/Plan to match reality) remains explicitly out of scope.
- User Story 5 (output formatting) is a presentation-only change — no
  Skill's own Allowed Reads/Creates/Modifications, Preconditions, or
  Deterministic Operations sections change as a result.
- This feature touches only `kit/skills/*/SKILL.md` (the product Skills
  misterspec ships to its own end users) and, if needed for consistency
  during implementation, `kit/templates/*.tmpl` — it does not touch
  `internal/`/`cmd/` Go code, since none of these changes require a new
  deterministic operation (Constitution Principle I/II already
  delegate ambiguity resolution and formatting choices to the agent
  layer, not the Go binary).
