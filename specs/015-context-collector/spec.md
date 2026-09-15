# Feature Specification: Context Collector and Retrieval

**Feature Branch**: `015-context-collector`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir com a sugestão 'phase 6: context
collector and retrieval'" (proceed with Phase 6 — Context Collector and
Retrieval — of `docs/context-engine-implementation.md`, itself
`docs/architecture-specification.md`'s Phase 7 "Second Brain", building
on 011's wikilinks, 012's reference graph, 013's document/chunk model,
and 014's disposable search index)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-014, this feature's "users" are misterspec's own
  deterministic layer and future callers — Phase 7's ranking/budgeting
  and Phase 8's own internal command, rather than an end-user UI of its
  own. Per context-engine-implementation.md's own Phase 6 scoping, this
  covers collecting and deduplicating a candidate set, each item
  labeled with why it was included — no scoring/weighting model, no
  token budget enforcement (Phase 7), no rendering or CLI command
  (Phase 8).
-->

### User Story 1 - Always Get the Non-Negotiable Baseline (Priority: P1)

Given a request naming a target artifact — plus, depending on what's
being asked, an optional current task, an optional intent (what kind of
work this context is for), and an optional free-text question — the
project's Constitution and the target artifact itself are always part
of the answer, regardless of what else is found or how the request is
phrased.

**Why this priority**: Everything else this feature does is optional
enrichment; this is the one guarantee nothing may ever crowd out —
later phases' own budget limits (Phase 7) will apply to everything
else, never to this baseline. It is also the foundation every other
story in this feature builds on: there is no candidate set to
deduplicate or expand before this exists.

**Independent Test**: Send a request naming an existing target artifact
with nothing else specified; confirm the answer includes the project's
Constitution content and the target artifact itself, labeled as the
reason each was included.

**Acceptance Scenarios**:

1. **Given** a request naming an existing target artifact and nothing
   else, **When** context is collected, **Then** the project's
   Constitution and the target artifact both appear in the result,
   each labeled as mandatory.
2. **Given** a request naming a target artifact with no connections to
   anything else in the project, **When** context is collected,
   **Then** the result still contains the mandatory baseline — never an
   empty or failed result.
3. **Given** a project with no Constitution recorded yet, **When**
   context is collected for a valid target, **Then** the target itself
   still appears, and the absence of a Constitution is not treated as a
   failure.
4. **Given** a request naming a target that does not exist, **When**
   context is collected, **Then** the request is rejected consistently
   with how every other existing capability already reports an unknown
   target.

---

### User Story 2 - Discover Everything Structurally and Semantically Connected (Priority: P2)

Given the target, also collect everything it's formally related to
(its parent, its dependencies, what it supersedes) and everything it's
semantically linked to or referenced by (explicit links it makes, and
other artifacts that link back to it) — reusing this project's own
already-established reference graph exactly, never a second,
independently-derived notion of "related."

**Why this priority**: This is what makes the answer actually useful
beyond the bare minimum — the structural and semantic neighborhood
around a target is exactly what 011/012 already made deterministically
knowable; this story is where that knowledge starts feeding into an
actual answer. Depends on User Story 1's baseline existing first.

**Independent Test**: Send a request for a target with a known parent,
a known dependency, and a known incoming reference from another
artifact; confirm all three appear in the result, each labeled with
which specific relationship justified including it.

**Acceptance Scenarios**:

1. **Given** a target with a declared parent and a declared dependency,
   **When** context is collected, **Then** both appear, each labeled
   with its specific relationship kind.
2. **Given** a target with an explicit link to another artifact,
   **When** context is collected, **Then** the linked artifact appears,
   labeled as a semantic connection.
3. **Given** an artifact that formally or semantically references the
   target, **When** context is collected for the target, **Then** that
   referencing artifact appears, labeled as an incoming connection.
4. **Given** a target with no formal or semantic connections at all,
   **When** context is collected, **Then** this story contributes
   nothing beyond User Story 1's own baseline — not an error.

---

### User Story 3 - Surface Relevant Text Matches (Priority: P3)

Given a free-text question (supplied directly, or naturally implied by
the current task), also collect content elsewhere in the project whose
text is relevant to it — reusing this project's own local search
capability exactly, never a second, independently-built search.

**Why this priority**: Structural connections (User Story 2) only
surface what's already explicitly linked; a relevant piece of
Knowledge or a past Learning that nobody thought to link explicitly is
exactly what free-text search exists to still surface. Depends on User
Story 1's baseline and is independent of User Story 2's own graph
traversal.

**Independent Test**: Send a request with a free-text question known to
match content in exactly one other artifact; confirm that content
appears in the result, labeled as a text match.

**Acceptance Scenarios**:

1. **Given** a request with a free-text question matching content
   elsewhere in the project, **When** context is collected, **Then**
   that content appears, labeled as a text match.
2. **Given** a request with no free-text question and no task to derive
   one from, **When** context is collected, **Then** this story
   contributes nothing — not an error, and not a requirement to
   fabricate a query from nothing.
3. **Given** a free-text question matching nothing anywhere in the
   project, **When** context is collected, **Then** this story
   contributes nothing, and the overall result is still whatever User
   Story 1/2 already produced — not a failure.

---

### User Story 4 - Never Show the Same Content Twice (Priority: P4)

When the same underlying piece of content is discovered more than once
— for example, an artifact that is both a formal dependency and a text
match — it appears exactly once in the result, carrying every reason it
was found rather than being repeated once per reason.

**Why this priority**: User Stories 2 and 3 can easily rediscover the
same content from different angles; without this guarantee, the result
would misrepresent how much distinct context actually exists, and a
later budget (Phase 7) would waste its limit on duplicates. Depends on
User Story 2 and 3 both existing, since there is nothing to deduplicate
before they produce overlapping candidates.

**Independent Test**: Construct a target with a dependency that is also
a free-text match for the request's own query; confirm the result
contains exactly one entry for it, listing both reasons.

**Acceptance Scenarios**:

1. **Given** the same piece of content discovered as both a structural
   connection (User Story 2) and a text match (User Story 3), **When**
   context is collected, **Then** it appears exactly once, listing
   both reasons.
2. **Given** the same piece of content discovered through two different
   structural relationships (for example, it is both a dependency and
   something that links back to the target), **When** context is
   collected, **Then** it still appears exactly once, listing both
   relationships.

---

### User Story 5 - Optionally Look One Step Further (Priority: P5)

When what's been found so far leaves room to explore further, expand
one additional, bounded step from an already-discovered connection
(never a second step from the original target directly) to surface
second-hop context — still deduplicated against everything already
found, and never allowed to grow without a limit.

**Why this priority**: The lowest-value, most optional part of
collection — genuinely useful sometimes, but never something the other
stories depend on, and the first thing later work (Phase 7's budgeting)
would trim away first. Depends on User Story 2 (there must be a
first-hop connection to expand from) and User Story 4 (an expansion
that only rediscovers what's already known must not duplicate it).

**Independent Test**: Construct a target whose direct dependency itself
has its own further dependency; confirm that further, second-hop
dependency can appear in the result, labeled as a second-hop
connection, distinguishable from a direct one — and confirm expansion
never proceeds a third hop out.

**Acceptance Scenarios**:

1. **Given** a target's direct dependency that itself has its own
   further connection, **When** context is collected, **Then** that
   further connection may appear, labeled distinctly as second-hop —
   never mistaken for a direct connection.
2. **Given** a second-hop candidate that duplicates something already
   found directly, **When** context is collected, **Then** it is not
   added a second time (User Story 4's own guarantee still holds).
3. **Given** a chain of connections three or more hops from the target,
   **When** context is collected, **Then** nothing beyond the second
   hop is ever included.

---

### Edge Cases

- A request naming a target of a type this project's own reference
  graph does not cover (per 012's own established scope) — rejected
  consistently with how 012 itself already rejects such a target, not
  a new, differently-shaped error.
- A request supplying an intent value that isn't one of this project's
  own recognized kinds of work — rejected consistently, not silently
  ignored.
- A request supplying neither a task nor a free-text query at all —
  still valid; the result is whatever User Story 1 and 2 alone produce.
- The same artifact appearing as both a direct (User Story 2) and a
  second-hop (User Story 5) connection — the direct classification
  always wins; it is never demoted to second-hop.
- A project large enough that the fully expanded candidate set (before
  any future budget is applied) is large — collection itself still
  completes and returns a complete, correctly deduplicated set; trimming
  it to fit any limit is explicitly a later feature's job (Phase 7).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a context request naming a target
  artifact (required) and, optionally, a current task, an intent
  (naming what kind of work this context is for), and a free-text
  question.
- **FR-002**: System MUST reject a request naming a target that does
  not exist, or an intent value it does not recognize, consistently
  with how this project's existing capabilities already report an
  invalid or unsupported request.
- **FR-003**: System MUST always include the project's Constitution (if
  one exists) and the target artifact itself in every successful
  result, regardless of what else is found and regardless of any future
  budget constraint.
- **FR-004**: The absence of a Constitution MUST NOT be treated as a
  failure — the target itself still appears.
- **FR-005**: System MUST collect every formal relationship (parent,
  dependency, supersession) and semantic connection (explicit link, and
  incoming reference from another artifact) already known for the
  target, reusing this project's own existing reference-graph
  computation exactly — never a second, independently derived notion of
  relatedness.
- **FR-006**: When a free-text question is supplied or reasonably
  derivable from the current task, System MUST also collect relevant
  content via this project's own existing local search capability —
  never a second, independently built search.
- **FR-007**: System MUST NOT fabricate a free-text question when none
  is supplied and none can reasonably be derived — text-based
  collection simply contributes nothing in that case.
- **FR-008**: System MUST deduplicate the collected set: the same
  underlying content discovered through more than one path appears
  exactly once, carrying every distinct reason it was included.
- **FR-009**: System MUST label every item in the result with why it
  was included — mandatory, a specific formal relationship, a specific
  semantic connection, a text match, or a second-hop expansion — never
  an unlabeled or generically-labeled item.
- **FR-010**: System MAY expand at most one additional step beyond a
  direct connection (never directly from the original target a second
  time), and MUST NOT expand further than that in this feature's own
  scope.
- **FR-011**: A second-hop candidate that duplicates something already
  found directly MUST NOT be added again, and an item found both
  directly and via second-hop expansion MUST always be classified as
  direct.
- **FR-012**: This capability MUST be strictly read-only — collecting
  context MUST NOT modify any project artifact, the reference graph, or
  the search index.
- **FR-013**: System MUST NOT apply any token or size budget to the
  collected set, and MUST NOT rank or numerically score candidates
  beyond the tier-level distinctions FR-009 already requires — both
  remain a later feature's own job.

### Key Entities

- **Context Request**: One caller's question — a required target
  artifact, plus an optional current task, an optional intent, and an
  optional free-text question.
- **Candidate**: One piece of collected context (either the
  Constitution's own relevant content or one of 013's own chunks),
  tagged with every distinct reason it was included — never presented
  without at least one such reason.
- **Context Candidate Set**: The complete, deduplicated collection of
  Candidates produced for one Context Request — this feature's own
  output; not yet ranked by numeric score, not yet trimmed to any
  budget.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of successful requests include the target artifact
  and, when one exists, the project's Constitution — zero observed
  omissions across tested fixtures.
- **SC-002**: For a representative set of targets with known formal and
  semantic connections, 100% of those connections appear in the
  result, each correctly labeled by its specific relationship kind.
- **SC-003**: For a representative set of free-text questions with a
  known matching artifact, 100% of them surface that artifact in the
  result.
- **SC-004**: Zero duplicate entries for the same underlying content
  are ever observed in a result, across every tested scenario that
  deliberately creates an overlap between two or more discovery paths.
- **SC-005**: Zero results are observed to contain a third-hop (or
  further) connection, across every tested multi-hop fixture.
- **SC-006**: Collecting context never modifies any project artifact,
  the reference graph, or the search index — verified across every
  tested scenario, zero exceptions.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 6 ("Context
  Collector and Retrieval"): an explicit request model and
  deterministic, tiered, deduplicated candidate collection with
  inclusion reasons. It explicitly excludes that document's own later
  phases — numeric ranking/weighting and token-budget enforcement
  (Phase 7), and the internal context command and any rendering of
  results (Phase 8) — each remains its own, later feature, continuing
  this project's established one-phase-at-a-time discipline (every
  prior feature, 001 through 014, has kept an equivalent boundary).
- A target artifact must be one of the five entity types this
  project's own reference graph already covers (Program, Feature, Spec,
  Knowledge, Learning — 012's own established scope, reused here
  without redrawing it, the same way 014 already reused it for its own
  `links` population). Extending collection to other artifact types is
  a future feature's own decision, not this one's.
- Recognized intent values match this project's own existing Skill
  vocabulary (for example: planning, task creation, implementation,
  validation, analysis) — this feature only validates and carries the
  value through; using it to change what gets prioritized or kept
  remains Phase 7's own job.
- "Relevant content via this project's own existing local search
  capability" means 014's own already-built full-text search — this
  feature adds no new search mechanism of its own.
- This feature produces structured data only — a labeled, deduplicated
  candidate set — never rendered prose and never a new user-facing
  command; Phase 8's own later "internal context" command is the actual
  entry point that will eventually expose this to a caller.
- Deduplication (User Story 4) and second-hop expansion (User Story 5)
  both operate purely on already-discovered structural/semantic/text
  signals — neither introduces a new way of discovering content beyond
  what 012's reference graph and 014's search already provide.
