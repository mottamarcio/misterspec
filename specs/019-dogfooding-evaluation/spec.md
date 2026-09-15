# Feature Specification: Dogfooding and Evaluation

**Feature Branch**: `019-dogfooding-evaluation`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "podemos seguir para a 'Phase 9
(Dogfooding and Evaluation)'. Se precisar testar explicitamente os
agents, aqui eu tenho acesso ao claude e o agy. Os demais não"
(proceed with Phase 9 — Dogfooding and Evaluation — of
`docs/context-engine-implementation.md`, run against this project's
own real development history and, where a live agent session is
needed, against exactly the two coding agents the user can personally
exercise: Claude Code and Antigravity)

## Note on sequencing relative to the source document

`docs/context-engine-implementation.md` §30 orders Phase 9 (Dogfooding)
before Phase 10 (Skill Integration), on the reasoning that Skills
should only be updated once retrieval quality is proven manually
first. This project deliberately built Phase 10
(018-multi-agent-skill-integration) before this feature, at the user's
own explicit direction. This feature therefore evaluates the
already-integrated system as it exists in production today — the four
Context-Pack-Aware Skills (`/implement`, `/create-plan`,
`/create-tasks`, `/analyze`) actually requesting and using Context
Packs — rather than the standalone `internal context` command in
isolation. This is a strictly more informative evaluation than the
source document's own original ordering would have produced, not a
narrower one: real Skill behavior is dogfooded end-to-end here, not
merely the underlying command.

## User Scenarios & Testing *(mandatory)*

<!--
  This feature's "users" are misterspec's own maintainers — the people
  who need real, measured evidence that the Context Engine (011-017)
  and its Skill integration (018) actually reduce repeated exploration
  without missing anything important, before trusting it for real
  work and before touching any ranking weight. Per
  context-engine-implementation.md §30 Phase 9's own explicit
  instruction, "tune ranking only from observed failures, not
  intuition alone" — this feature's entire purpose is producing that
  evidence, not assuming it. Per the user's own explicit constraint,
  every scenario requiring a live coding-agent session uses only
  Claude Code and Antigravity — the two agents actually available for
  hands-on testing; the other four agents installed by 018 (Codex CLI,
  GitHub Copilot, Cursor, Devin for Terminal) share byte-identical
  Skill content and installation mechanism (already verified in 018)
  but are not live-dogfooded in this feature (see Assumptions).
-->

### User Story 1 - Validate Retrieval Against This Project's Own Real History (Priority: P1)

For a representative set of this project's own already-completed
Specs (011 through 018 — real, finished development work with a real
Plan, real Tasks, and real Learnings already on record), request a
Context Pack for each one and compare what it returns against what
that Spec's own real historical work actually needed and used —
without requiring any live agent session, since the ground truth
(what dependency, Knowledge, or Learning that Spec's real
implementation actually relied on) is already known and recorded.

**Why this priority**: This is the one piece of evidence that doesn't
depend on agent availability at all — it can be produced immediately,
against real (not synthetic) project history, and it is the most
direct possible test of the source document's own stated purpose:
"find and assemble the smallest sufficient project context." Every
other story in this feature either extends or acts on what this one
establishes.

**Independent Test**: For each of the 011-018 Specs, request a
Context Pack using that Spec's own real, historical intent (e.g.
`implementation` for a Spec whose real work was primarily coding);
confirm every dependency, wikilinked Knowledge entry, and directly
relevant Learning that Spec's own real Plan/Tasks/Learnings already
document as having been used is present in the returned pack, and note
any that is missing.

**Acceptance Scenarios**:

1. **Given** one of the 011-018 Specs and its own real, already-
   recorded dependencies, **When** a Context Pack is requested for it,
   **Then** every one of those real dependencies appears among the
   returned items.
2. **Given** the same Spec, **When** the returned pack is compared
   against that Spec's own real Plan/Tasks content, **Then** any
   content present in the pack that turns out to be irrelevant to that
   Spec's own real work is recorded as a specific finding, not silently
   ignored.
3. **Given** all 011-018 Specs have been evaluated this way, **When**
   the results are reviewed together, **Then** a specific, itemized
   list of every found omission or irrelevant inclusion exists — never
   a vague overall impression.

---

### User Story 2 - Dogfood a Real Skill Invocation With Claude Code (Priority: P2)

Using Claude Code — the primary agent the user can exercise directly —
run at least one of the four Context-Pack-Aware Skills
(`/implement`, `/create-plan`, `/create-tasks`, `/analyze`) against a
piece of this project's own real, current development work, and
observe directly: whether the Skill actually requested a Context Pack
first, whether that pack was sufficient on its own, whether any
additional manual exploration was needed beyond it, and how long the
whole request round-trip took.

**Why this priority**: User Story 1 proves the underlying retrieval is
sound in principle; this story proves the actual, already-shipped
Skill instructions (018) really do invoke it correctly in a real,
live session and that doing so doesn't slow the agent down or leave it
stuck. It depends on User Story 1 only in that a known-good retrieval
result makes it easier to judge whether any gap observed live is a new
problem or already-known and accepted.

**Independent Test**: Invoke one of the four updated Skills through
Claude Code against a real Spec in this repository; directly observe
and record whether it requested a Context Pack, what that pack
contained, whether further exploration was needed, and the elapsed
time for the request.

**Acceptance Scenarios**:

1. **Given** a real Spec with executable work remaining, **When** a
   Context-Pack-Aware Skill is invoked through Claude Code, **Then**
   the Skill's own request for a Context Pack is directly observed to
   occur before other exploration.
2. **Given** that same invocation, **When** the agent's own subsequent
   actions are reviewed, **Then** it is recorded whether the pack alone
   was sufficient or whether further exploration was genuinely needed,
   and specifically what was missing if so.
3. **Given** that same invocation, **When** its total elapsed time is
   measured, **Then** it is recorded as a specific, comparable figure —
   not a qualitative impression.

---

### User Story 3 - Confirm the Same Behavior With a Second, Independent Agent (Priority: P3)

Repeat User Story 2's own live invocation using Antigravity — the
second agent the user can exercise directly — to confirm the observed
behavior (Context Pack requested first, sufficiency, elapsed time)
is not specific to Claude Code alone, since both agents install and
run the identical canonical Skill content (018).

**Why this priority**: A single agent's behavior could reflect
something specific to how that one agent reads instructions rather
than the Skill content itself; a second, independently-behaving agent
producing the same result is real corroborating evidence, not a
repeat of the same test. It depends directly on User Story 2's own
scenario existing to repeat.

**Independent Test**: Invoke the same Skill against the same (or an
equivalent) real Spec through Antigravity; confirm the same three
observations User Story 2 recorded (Context Pack requested first,
sufficiency, elapsed time), and note any difference between the two
agents' own behavior.

**Acceptance Scenarios**:

1. **Given** the same or an equivalent real Spec, **When** the same
   Skill is invoked through Antigravity, **Then** it is directly
   observed to request a Context Pack before other exploration, same
   as Claude Code did.
2. **Given** both agents' own recorded observations (User Story 2 and
   this story), **When** they are compared, **Then** any material
   difference in outcome between the two agents is specifically noted,
   not assumed away.

---

### User Story 4 - Decide Whether Ranking Needs Tuning, Based Only on Observed Evidence (Priority: P4)

Using only the specific, itemized findings from User Stories 1-3 —
never intuition or a hypothetical scenario — decide whether 016's own
ranking/budgeting weights need adjustment, and record that decision
with its own supporting evidence either way.

**Why this priority**: This is `docs/context-engine-
implementation.md` §30 Phase 9's own explicit governing rule ("tune
ranking only from observed failures, not intuition alone") and this
feature's own actual exit criterion — it cannot be evaluated before
User Stories 1-3 have actually produced evidence to judge.

**Independent Test**: Review every finding recorded by User Stories
1-3; for each one, either connect it to a specific ranking/budgeting
behavior that plausibly caused it, or record that no such connection
exists; conclude with an explicit decision (no change needed, or a
specific, evidence-cited change) rather than a general impression.

**Acceptance Scenarios**:

1. **Given** every finding from User Stories 1-3, **When** they are
   reviewed, **Then** each is explicitly classified as either
   attributable to ranking/budgeting behavior or not.
2. **Given** zero findings are attributable to ranking/budgeting
   behavior, **When** a conclusion is recorded, **Then** it states
   plainly that no ranking change is warranted by the evidence
   gathered, without inventing a change to justify the effort spent.
3. **Given** one or more findings are attributable to ranking/
   budgeting behavior, **When** a conclusion is recorded, **Then** it
   names the specific behavior, the specific evidence, and the
   specific proposed change — never a vague "improve ranking"
   statement.

---

### Edge Cases

- A Spec among 011-018 whose real historical work involved very little
  cross-referencing (few or no dependencies/wikilinks) — still
  evaluated; a pack containing only mandatory content for such a Spec
  is a pass, not a gap.
- A live Skill invocation (User Story 2/3) that fails to complete for
  reasons unrelated to the Context Engine (e.g. an unrelated
  environment issue) — recorded as an inconclusive observation, not
  miscounted as a retrieval failure.
- The two agents (User Story 2 vs. 3) producing different elapsed
  times for equivalent requests — expected and recorded as a
  data point, not treated as a defect, since absolute latency is not
  itself a pass/fail criterion here (see Assumptions).
- A finding that could plausibly be explained by either a ranking
  issue or a collection (015) issue — recorded against both
  possibilities explicitly rather than arbitrarily attributed to only
  one.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: This effort MUST evaluate Context Pack retrieval against
  every one of Specs 011 through 018 in this repository, using each
  Spec's own real, already-recorded dependencies, wikilinks, and
  Learnings as ground truth.
- **FR-002**: For each Spec evaluated under FR-001, every real
  dependency/wikilink/Learning that Spec's own historical work is
  already documented as having used MUST be checked for presence in
  the returned Context Pack, and any absence MUST be recorded as a
  specific, named finding.
- **FR-003**: Any content returned in a Context Pack that is judged
  irrelevant to the Spec it was requested for MUST also be recorded as
  a specific, named finding — omission and over-inclusion are both
  tracked, not only the former.
- **FR-004**: At least one live invocation of a Context-Pack-Aware
  Skill (018) MUST be performed and directly observed through Claude
  Code against a real piece of this project's own current work,
  recording whether a Context Pack was requested first, whether it was
  sufficient, and its elapsed time.
- **FR-005**: The same observation MUST be repeated through
  Antigravity, and any material difference between the two agents'
  own behavior MUST be explicitly recorded.
- **FR-006**: This effort MUST NOT perform a live agent-session
  observation through any agent other than Claude Code and
  Antigravity — the other four agents installed by 018 remain
  structurally verified (018's own test suite) but not live-dogfooded
  here.
- **FR-007**: Every finding produced by User Stories 1-3 MUST be
  reviewed and explicitly classified as either attributable to a
  specific ranking/budgeting behavior or not, before any ranking
  change is considered.
- **FR-008**: A ranking/budgeting change MUST NOT be made under this
  effort unless it is directly justified by at least one specific,
  recorded finding — intuition or a hypothetical scenario is not
  sufficient justification (docs/context-engine-implementation.md §30
  Phase 9's own explicit rule).
- **FR-009**: This effort MUST produce one durable, reviewable record
  (a report) of every finding, every classification, and the final
  tuning decision — never leave the evaluation's own outcome only in
  ephemeral conversation or terminal output.
- **FR-010**: This effort MUST NOT modify the Context Engine (011-017)
  or the Skill Integration (018) except for a ranking/budgeting change
  explicitly justified under FR-007/FR-008 — the evaluation itself is
  read-only against the rest of the system.

### Key Entities

- **Dogfooding Finding**: One specific, recorded observation from
  User Story 1, 2, or 3 — a missing dependency, an irrelevant
  inclusion, a Skill's own request behavior, a sufficiency judgment,
  or an elapsed-time measurement — always concrete, never a general
  impression.
- **Tuning Decision**: The final, evidence-cited conclusion from User
  Story 4 — either "no change warranted" or a specific proposed
  ranking/budgeting adjustment, each finding it is based on named
  explicitly.
- **Dogfooding Report**: The one durable record (FR-009) collecting
  every Finding and the final Tuning Decision, reviewable after this
  effort concludes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of Specs 011-018 have a recorded Context Pack
  evaluation, each with an explicit pass/fail judgment against its own
  real historical dependencies.
- **SC-002**: 0 real, historically-used dependencies/wikilinks/
  Learnings among the evaluated Specs are missing from their own
  Context Pack, or every such gap found is explicitly recorded as a
  named finding (not silently accepted).
- **SC-003**: At least one live Skill invocation is directly observed
  and recorded for each of Claude Code and Antigravity, each reporting
  a concrete elapsed-time figure.
- **SC-004**: 100% of findings produced by this effort are explicitly
  classified as ranking-attributable or not, with the classification
  itself recorded, not merely implied.
- **SC-005**: Exactly one Tuning Decision is recorded, and it is
  traceable to specific findings — reviewable by a maintainer without
  needing to re-run any part of the evaluation.
- **SC-006**: 0 changes are made to the Context Engine or Skill
  Integration content that are not directly traceable to a recorded
  finding under this effort's own Tuning Decision.

## Assumptions

- This feature evaluates already-completed Specs 011-018 as its
  primary evidence source (User Story 1) precisely because their real
  outcomes are already known — this avoids the circularity of
  "dogfooding" a system using only speculative future work, and gives
  every retrieval judgment an actual, historical ground truth to check
  against.
- "Live agent session" observations (User Story 2/3) are scoped to
  exactly Claude Code and Antigravity, per the user's own explicit,
  stated access constraint. Codex CLI, GitHub Copilot, Cursor, and
  Devin for Terminal remain structurally verified by 018's own test
  suite (identical Skill content, identical installation mechanism)
  but are not independently dogfooded live in this feature; extending
  live dogfooding to them is deferred until they become available for
  hands-on testing.
- Elapsed time (FR-004/FR-005) is recorded as a data point for later
  reference, not compared against a specific numeric threshold in this
  feature — no existing performance target has been established for
  `internal context` requests to judge against yet (context-engine-
  implementation.md §27 sets only qualitative expectations); a
  concrete latency budget, if ever needed, is future work informed by
  this feature's own recorded baseline.
- "Index size" (mentioned as a useful dogfooding metric in
  context-engine-implementation.md §22) is not tracked as a separate
  metric in this feature — none of the Specs evaluated exercise a
  large enough corpus for index size to be a meaningful signal yet;
  revisiting this is deferred until the project's own real corpus
  grows enough to make it informative.
- The Dogfooding Report (FR-009) is a Markdown document produced
  alongside this feature's own planning/implementation artifacts —
  its exact location and structure are an implementation detail left
  to planning, not user-facing product behavior.
- This feature does not implement any new deterministic capability —
  it is an evaluation exercise over 011-018's own already-shipped
  behavior, per docs/context-engine-implementation.md §30 Phase 9's
  own scope, and per this project's own established one-phase-at-a-
  time discipline. A ranking/budgeting change is in scope only if
  FR-007/FR-008's own evidence bar is met — otherwise this feature's
  correct, successful outcome is "no code change," which is not a
  failure of the feature.
