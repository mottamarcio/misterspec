# Feature Specification: Spec-Kit Tooling Improvements

**Feature Branch**: `023-speckit-tooling-improvements`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "podemos trabalhar na especificação e implementação dos 8 itens mencionados acima" — the 8 prioritized findings from comparing upstream `github/spec-kit`'s `templates/` and `templates/commands/` against misterspec's own `.specify/` dev-tooling (the Spec-Driven-Development workflow used to build misterspec itself, distinct from `kit/skills/`, the product misterspec ships to its own end users).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A mandatory hook that's described is a hook that runs (Priority: P1)

A developer (or coding agent) runs any `/speckit-*` command with a
mandatory pre/post hook registered (e.g. `before_specify`'s
git-branch-creation hook). Today, every skill's own instructions stop at
"emit this block, then wait for the result" — nothing tells the agent
that emitting the block is not the same as actually invoking the
named slash command. This exact ambiguity caused the git-branch-creation
hook to be described without being run for several features in a row
earlier in this project's own history.

**Why this priority**: This is a demonstrated, already-observed bug, not
a hypothetical gap — it directly undermines every other hook-dependent
guarantee this tooling relies on (automatic branching, automatic
commits). Fixing it is the highest-value, lowest-risk change in this set
(text-only, no new logic).

**Independent Test**: Trigger any command with a mandatory hook
registered; confirm the hook's own command actually executes (its
observable effect — e.g. a new git branch — exists), not just that its
description was printed.

**Acceptance Scenarios**:

1. **Given** `.specify/extensions.yml` registers a mandatory
   (`optional: false`) hook for a command, **When** that command runs,
   **Then** the hook's own slash command is actually invoked as a real
   step, and its effect is verifiable afterward — not merely described
   in the response.
2. **Given** `.specify/extensions.yml` is present but cannot be parsed
   (malformed YAML), **When** any `/speckit-*` command runs, **Then**
   the response explicitly states that hook checking failed, including
   the parse error, and names which hooks (if any are known to be
   mandatory) could not be checked — never a silent, unannounced skip.

---

### User Story 2 - Checklists stay a read-only quality gate during implementation (Priority: P1)

A developer runs `/speckit-implement` against a feature whose
`checklists/` directory already has items marked complete from an
earlier validation pass. Nothing today stops the implementation skill
from "helpfully" checking off additional checklist items as it works,
which would corrupt the checklist's own meaning as an independent
review record.

**Why this priority**: A checklist's value is entirely that its
checkmarks reflect deliberate review, not incidental side effects of
unrelated work — losing that guarantee is a real, silent
data-integrity gap in the same review record `/speckit-implement`
already reads as a gate.

**Independent Test**: Run `/speckit-implement` against a feature with a
partially-checked checklist; confirm every checkbox's state is
byte-for-byte unchanged afterward, regardless of what implementation
work occurred.

**Acceptance Scenarios**:

1. **Given** a checklist file with some items checked and some
   unchecked, **When** `/speckit-implement` runs to completion, **Then**
   no checkbox in that file changes state.
2. **Given** a freshly generated checklist from `/speckit-checklist`,
   **When** it is written, **Then** every new item starts unchecked
   (`[ ]`) — checkbox state is the reviewer's to set, never the
   generating skill's own.

---

### User Story 3 - Re-running task-to-issue sync never duplicates issues (Priority: P2)

A developer runs `/speckit-taskstoissues` after `tasks.md` already had
some of its tasks synced to GitHub issues in an earlier run (e.g. new
tasks were appended by a later phase, or the command was re-run after
an interruption). Today, every invocation creates a fresh issue per
task with no memory of what was already created — a second run
duplicates every already-synced task.

**Why this priority**: A real, reproducible correctness bug the moment
this command is used more than once against the same feature — not a
hypothetical edge case.

**Independent Test**: Run `/speckit-taskstoissues` twice in a row
against the same `tasks.md` with no new tasks added between runs;
confirm the second run creates zero new issues.

**Acceptance Scenarios**:

1. **Given** every task in `tasks.md` already has a corresponding open
   GitHub issue, **When** `/speckit-taskstoissues` runs again, **Then**
   no new issues are created for those tasks.
2. **Given** `tasks.md` gained new tasks since the last sync, **When**
   `/speckit-taskstoissues` runs, **Then** only the new tasks get new
   issues — already-synced tasks are left alone.

---

### User Story 4 - Constitution updates never silently absorb unrelated work (Priority: P2)

A developer's prompt to `/speckit-constitution` mixes a genuine
governance change with an unrelated implementation, code-generation, or
deployment request (deliberately or by accident). Today nothing
distinguishes the two — the command could either silently ignore the
unrelated request or, worse, act on it under the wrong command's
authority.

**Why this priority**: A direct, close-to-exact match for this
project's own Constitution Principle VII (Explicit Mutation
Boundaries) — a governance command acting outside its own lane is
exactly the failure mode that principle exists to prevent.

**Independent Test**: Send `/speckit-constitution` a prompt containing
both a real governance change and an unrelated "also implement X"
request; confirm the governance change is applied, the unrelated
request is explicitly deferred (never silently executed and never
silently dropped), and it is visibly surfaced back to the developer.

**Acceptance Scenarios**:

1. **Given** a `/speckit-constitution` prompt containing an unrelated
   implementation/code/deploy request, **When** the command runs,
   **Then** it does not perform that unrelated work, and the response
   names it explicitly as deferred, not silently dropped.
2. **Given** a `/speckit-constitution` prompt containing only a genuine
   governance change, **When** the command runs, **Then** it proceeds
   exactly as it does today — this guard never blocks legitimate
   constitution work.

---

### User Story 5 - Each core lifecycle command self-checks before reporting done (Priority: P3)

A developer runs `/speckit-specify`, `/speckit-plan`, `/speckit-tasks`,
`/speckit-implement`, or `/speckit-clarify`. Today each one reports
completion once its own steps have executed, with no final, explicit
self-verification that its own required outputs actually exist and are
well-formed before saying so.

**Why this priority**: Valuable defense-in-depth, but each of these
commands already has its own internal validation logic (e.g.
`/speckit-specify`'s own checklist loop) — this is a final belt-and-
suspenders check, not a fix for a demonstrated failure.

**Independent Test**: Run any of the five commands; confirm its final
report includes an explicit self-check against its own required
outputs (e.g. "spec.md exists and has no remaining
`[NEEDS CLARIFICATION]` markers") before declaring success.

**Acceptance Scenarios**:

1. **Given** one of the five core lifecycle commands completes its main
   work, **When** it reports completion, **Then** the report includes
   an explicit checklist of that command's own required outputs, each
   confirmed present and well-formed.
2. **Given** one of those required outputs is missing or malformed at
   the end of a run, **When** the command reaches its own completion
   step, **Then** it reports the gap explicitly instead of declaring
   unconditional success.

---

### User Story 6 - A drift check between intent and code, run on demand (Priority: P3)

After one or more `/speckit-implement` passes against a feature, a
developer wants to know whether the actual codebase still matches what
`spec.md`/`plan.md`/`tasks.md` describe — without re-reading everything
by hand, and without any tool silently rewriting those already-approved
documents to "fix" a mismatch it finds.

**Why this priority**: Genuinely valuable for long-running or
multi-session features, but it is a new, standalone capability
(closest to an addition, not a fix to an existing gap), and every
other item in this set is either a bug fix or a safety hardening of
something that already exists.

**Independent Test**: After a feature's implementation is believed
complete, run the new command against it; confirm it reports any real
gap between intent and code (or reports none, if there truly is none)
without modifying `spec.md`, `plan.md`, or any already-existing task.

**Acceptance Scenarios**:

1. **Given** a feature whose code already fully satisfies its own
   spec/plan/tasks, **When** the new command runs, **Then** it reports
   no gaps and leaves every existing file byte-for-byte unchanged.
2. **Given** a feature where the actual code is missing, contradicts, or
   goes beyond what its own spec/plan/tasks describe, **When** the new
   command runs, **Then** it reports each gap with a clear severity, and
   its only permitted write is appending new tasks to the end of
   `tasks.md` — it never rewrites `spec.md`, `plan.md`, or any
   already-existing task.
3. **Given** a reported gap is a Constitution violation, **When** the
   new command classifies it, **Then** that gap is always reported at
   the highest severity, listed first.

---

### User Story 7 - Clarification questions are genuinely questions, and checklists stay current (Priority: P3)

A developer runs `/speckit-clarify` against a spec with ambiguous
requirements. Today a "question" can end up being little more than a
restated requirement ID or topic label rather than an actual question,
and once the spec is updated from the developer's answers, the spec
quality checklist `/speckit-specify` originally generated is never
revisited — it can silently go stale relative to the now-updated spec.

**Why this priority**: A real quality-of-output gap, but narrower in
blast radius than the P1/P2 items — it affects clarification quality,
not correctness or safety of what's already written.

**Independent Test**: Run `/speckit-clarify` against a spec with a
genuinely ambiguous requirement; confirm each generated question is a
complete interrogative sentence (never a bare label or ID) with a
one-sentence rationale, and that the spec quality checklist is
re-validated against the updated spec once all questions are answered.

**Acceptance Scenarios**:

1. **Given** `/speckit-clarify` identifies an ambiguity, **When** it
   presents a question, **Then** the question is a complete sentence
   ending in "?", never a bare requirement ID or topic label, and is
   preceded by a one-sentence statement of why it matters.
2. **Given** all clarification questions in a round have been answered
   and the spec updated, **When** that round completes, **Then** the
   spec's own quality checklist is re-validated against the updated
   spec, and its results (newly-passing items, and any regressions) are
   reported.

---

### User Story 8 - Plans and task lists stay precise and appropriately scoped (Priority: P3)

A developer runs `/speckit-plan` or `/speckit-tasks`. Today
`quickstart.md` (a Phase 1 planning output) has no explicit content
boundary, risking it growing into a second, redundant task list; and
individual tasks that touch a field with real constraints
(max length, required/nullable, enum values, validation rules) don't
have to quote those constraints, leaving them to be rediscovered or
reinvented at implementation time.

**Why this priority**: Smallest blast radius of the eight — a
precision/drift-reduction improvement to already-working commands, not
a correctness or safety fix.

**Independent Test**: Generate a plan whose data model has at least one
constrained field; confirm `quickstart.md` contains no full
implementation code, service bodies, migrations, or a complete test
suite, and confirm the resulting task list quotes that field's own
constraint verbatim in its task description.

**Acceptance Scenarios**:

1. **Given** `/speckit-plan` generates a `quickstart.md`, **When** it is
   written, **Then** it contains manual verification steps and example
   invocations only — no full implementation code, model/service
   bodies, migrations, or complete test suites.
2. **Given** `/speckit-tasks` generates a task touching a field with a
   documented constraint in `data-model.md`, **When** that task's
   description is written, **Then** it quotes the constraint verbatim
   rather than leaving it to be inferred later.

---

### Edge Cases

- What happens when a hook is registered as mandatory but its own
  underlying extension command no longer exists (e.g. removed)? The
  visible-failure behavior from User Story 1 applies here too — the gap
  is reported, never silently ignored.
- What happens when `/speckit-implement` needs to reference checklist
  content (e.g. to decide what "done" means) without ever writing to
  it? Reading remains fully permitted — only writes to existing
  checkbox markers are restricted (User Story 2).
- What happens when `/speckit-taskstoissues`'s issue-matching (User
  Story 3) finds an issue whose title looks like it matches a task ID
  but was actually created manually, unrelated to this tooling? The
  matching rule is scoped to this project's own canonical task-ID title
  format, minimizing (though not perfectly eliminating) this
  possibility — an explicitly accepted, low-probability tradeoff rather
  than an unsolved gap.
- What happens when `/speckit-converge` (User Story 6) is run against a
  feature that was never planned or tasked through this tooling at all
  (no `spec.md`/`plan.md`/`tasks.md` to compare against)? It reports
  that it has no intent artifacts to compare against and performs no
  writes, rather than guessing at intent from code alone.
- What happens when a `/speckit-constitution` prompt (User Story 4) is
  ambiguous about whether a request is governance or unrelated work?
  The guard favors deferring and surfacing over silently guessing —
  an explicit, visible question to the developer is preferred over a
  wrong automatic classification in either direction.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every `/speckit-*` command with a hook registered in
  `.specify/extensions.yml` MUST state, as part of executing a
  mandatory hook, that the hook's own command must actually be invoked
  as a real step — not merely described — before the command's own
  work proceeds.
- **FR-002**: Every `/speckit-*` command MUST, when
  `.specify/extensions.yml` exists but cannot be parsed, explicitly
  report that hook checking failed (including the parse error) rather
  than silently continuing as if no hooks were registered.
- **FR-003**: `/speckit-implement` MUST NOT modify any checkbox marker
  in any file under a feature's own `checklists/` directory.
- **FR-004**: `/speckit-checklist` MUST write every newly generated
  checklist item as unchecked.
- **FR-005**: `/speckit-taskstoissues` MUST determine, before creating
  any issue, which tasks in the current `tasks.md` already have a
  corresponding issue, and MUST create issues only for tasks that do
  not.
- **FR-006**: `/speckit-constitution` MUST identify any request in its
  own input that is not itself a governance/constitution change (e.g.
  a feature implementation, code generation, or deployment request),
  MUST NOT perform that request, and MUST explicitly surface it back to
  the developer as deferred.
- **FR-007**: `/speckit-specify`, `/speckit-plan`, `/speckit-tasks`,
  `/speckit-implement`, and `/speckit-clarify` MUST each report, as
  part of declaring completion, an explicit self-check confirming that
  command's own required outputs exist and are well-formed.
- **FR-008**: A new command, `/speckit-converge`, MUST compare a
  feature's own `spec.md`, `plan.md`, and `tasks.md` (plus the project
  Constitution) against the actual codebase, classify any gap by
  severity, and report every gap found.
- **FR-009**: `/speckit-converge`'s only permitted write MUST be
  appending new tasks to the end of an existing `tasks.md`; it MUST
  NOT modify `spec.md`, `plan.md`, or any task that already existed
  before it ran, and MUST make no write at all when it finds no gap.
- **FR-010**: `/speckit-converge` MUST always classify a Constitution
  violation at its highest severity level and list such violations
  first in its report.
- **FR-011**: `/speckit-clarify` MUST reject a candidate clarification
  question that is not a complete interrogative sentence (e.g. a bare
  requirement ID or topic label used as if it were the question), and
  MUST precede each presented question with a one-sentence statement of
  why it matters.
- **FR-012**: `/speckit-clarify` MUST, once every question in a
  clarification round has been answered and the spec updated,
  re-validate that feature's own spec quality checklist against the
  updated spec and report the result.
- **FR-013**: `/speckit-plan` MUST NOT include full implementation
  code, complete model/service bodies, database migrations, or a
  complete test suite in a `quickstart.md` it generates.
- **FR-014**: `/speckit-tasks` MUST quote a field's own documented
  constraint (from `data-model.md`) verbatim in the description of any
  task that touches that field, whenever `data-model.md` documents one.

### Key Entities

- **Hook** (existing, `.specify/extensions.yml`): gains a stronger
  execution guarantee (FR-001) and a mandatory visible-failure path
  when its own registry can't be read (FR-002); no new fields.
- **Checklist** (existing, `checklists/*.md`): gains an explicit
  read-only contract during implementation (FR-003, FR-004); no new
  fields.
- **Convergence Report** (new, produced by `/speckit-converge`, not
  persisted as its own file): a set of classified gaps (severity,
  description, affected artifact), plus the new tasks appended to
  `tasks.md` as a result, if any.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A mandatory hook registered for any `/speckit-*` command
  has an observable effect (e.g. an actual new git branch) in 100% of
  runs where its trigger condition is met — not merely a described
  intention to run it.
- **SC-002**: Checklist files under any feature's `checklists/`
  directory are never modified by `/speckit-implement`, verified across
  repeated implementation runs against the same feature.
- **SC-003**: Running `/speckit-taskstoissues` twice in a row against
  an unchanged `tasks.md` creates zero duplicate issues on the second
  run.
- **SC-004**: A `/speckit-constitution` prompt mixing governance and
  unrelated work results in the unrelated work being visibly deferred,
  never silently performed and never silently dropped, in 100% of such
  cases.
- **SC-005**: Each of the five core lifecycle commands' own completion
  reports include an explicit self-check against that command's
  required outputs.
- **SC-006**: `/speckit-converge`, run against a feature with a known,
  deliberately introduced gap between code and spec, reports that gap
  without modifying any pre-existing file other than appending to
  `tasks.md`.
- **SC-007**: Every clarification question `/speckit-clarify` presents
  is a grammatically complete question, verified across multiple
  clarification rounds on different specs.
- **SC-008**: A generated `quickstart.md` contains no code block longer
  than a short example invocation or manual verification step.

## Assumptions

- This feature changes only this repository's own dev-tooling
  (`.specify/templates/`, `.claude/skills/speckit-*/SKILL.md`, and
  `.specify/extensions.yml`'s own hook-related documentation) — it does
  not touch `kit/skills/` (the Skills misterspec ships to its own end
  users) or any Go source under `internal/`, since those are a
  different, unrelated artifact model.
- `/speckit-converge` (User Story 6) is scoped to reporting and
  append-only task generation for this iteration; deeper remediation
  (e.g. automatically fixing a found gap) is explicitly out of scope,
  matching the same read-only-until-asked posture the rest of this
  project's tooling already follows.
- The exact severity taxonomy and matching rules for
  `/speckit-taskstoissues`'s deduplication (User Story 3) and
  `/speckit-converge`'s gap classification (User Story 6) are
  implementation-level decisions properly resolved during
  `/speckit-plan`, not this specification.
- All 8 user stories are independent of one another (each touches a
  different skill/template file, or an orthogonal concern within one
  file) and may be implemented, tested, and delivered in any order or
  in parallel.
