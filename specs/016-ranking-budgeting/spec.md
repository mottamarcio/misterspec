# Feature Specification: Ranking and Budgeting

**Feature Branch**: `016-ranking-budgeting`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir com a sugestão 'phase 7: ranking
and budgeting'" (proceed with Phase 7 — Ranking and Budgeting — of
`docs/context-engine-implementation.md`, itself `docs/architecture-
specification.md`'s Phase 7 "Second Brain", building directly on 015's
own deduplicated, tier-labeled candidate set)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-015, this feature's "users" are misterspec's own
  deterministic layer and future callers — Phase 8's own internal
  command, rather than an end-user UI of its own. Per context-engine-
  implementation.md's own Phase 7 scoping, this covers scoring 015's
  own candidates and fitting them into a token budget — no rendering,
  no CLI command (Phase 8).
-->

### User Story 1 - Rank Structural Truth Above Textual Coincidence (Priority: P1)

Given the candidate set 015 already collects and labels, assign every
candidate a deterministic relevance score — but never let a strong
textual coincidence outrank an artifact's own mandatory content or its
genuine structural/semantic connections, no matter how well that
coincidence happens to match the request's own words.

**Why this priority**: This is the one guarantee that makes ranking
trustworthy at all — if a lucky text match could ever bump a real
dependency out of the top of the list, every other guarantee in this
feature (and the budgeting built on top of it) would be unreliable.
Nothing else in this feature means anything until ordering itself is
trustworthy.

**Independent Test**: Build a candidate set where a purely textual
match happens to score very strongly, alongside the target's own
mandatory content and a genuine structural dependency; confirm the
mandatory content and the structural dependency both rank above the
text match regardless of its own score.

**Acceptance Scenarios**:

1. **Given** a candidate set containing mandatory content, a structural
   connection, and a strong but purely textual match, **When** the set
   is ranked, **Then** the mandatory content and the structural
   connection both rank above the textual match.
2. **Given** two candidates within the same priority tier, **When** the
   set is ranked, **Then** the more textually relevant one ranks
   higher, and the same relative order is produced every time for the
   same input.
3. **Given** the current task or intent of the request, **When** the
   set is ranked, **Then** the request's own intent may influence
   ordering only within a tier, never in a way that lets a lower tier
   outrank a higher one.

---

### User Story 2 - Fit Within a Token Budget Without Losing What Matters (Priority: P2)

Given a token budget, return the smallest ranked set of context that
fits within it — filling the budget tier by tier, from most to least
essential, rather than simply sorting everything by score and cutting
wherever the numbers run out.

**Why this priority**: This is the entire practical point of the
Context Engine — without a budget, nothing controls how much context an
agent receives, undermining the whole motivation behind this project's
own "Second Brain" effort. Depends on User Story 1's own ordering
already being trustworthy, since budgeting decides what to cut based on
that same order.

**Independent Test**: Build a candidate set larger than a given budget,
spanning several priority tiers; confirm the returned result fits
within the budget, with lower-priority tiers trimmed before any
higher-priority content is ever touched.

**Acceptance Scenarios**:

1. **Given** a candidate set whose combined estimated cost exceeds the
   requested budget, **When** budgeting is applied, **Then** the
   returned result's own combined estimated cost fits within the
   budget.
2. **Given** a budget boundary that falls in the middle of a lower-
   priority tier, **When** budgeting is applied, **Then** content from
   that tier is trimmed before any higher-priority tier's own content
   is touched.
3. **Given** a candidate set that already fits comfortably within the
   requested budget, **When** budgeting is applied, **Then** nothing is
   removed at all.
4. **Given** no budget is specified, **When** budgeting is applied,
   **Then** a fixed, documented default budget is used.

---

### User Story 3 - Never Silently Drop Mandatory Context (Priority: P3)

When the mandatory content alone (the project's Constitution and the
target artifact itself) already exceeds the requested budget, still
return all of it in full, clearly flag that the budget was exceeded,
and report by how much — never silently omit or shorten mandatory
content to force a fit.

**Why this priority**: An agent that silently receives an incomplete
Constitution or an incomplete target artifact is worse off than one
that receives a clear signal the budget couldn't be honored — this is
the one situation where correctness must visibly win over the budget,
not quietly lose to it. Depends on User Story 2's own budgeting
mechanism existing to have something to except from.

**Independent Test**: Construct mandatory content whose own estimated
cost alone exceeds a deliberately small budget; confirm the full
mandatory content is still returned, the result is flagged as having
exceeded budget, and the estimated overage is reported.

**Acceptance Scenarios**:

1. **Given** mandatory content whose own estimated cost exceeds the
   requested budget, **When** budgeting is applied, **Then** all of the
   mandatory content is still returned in full.
2. **Given** that same situation, **When** the result is inspected,
   **Then** it is clearly flagged as having exceeded the budget, with
   the estimated overage reported.
3. **Given** mandatory content alone exceeds the budget, **When**
   budgeting is applied, **Then** every optional (non-mandatory) item is
   still eligible to be omitted to avoid making the overage worse.

---

### User Story 4 - Explain Every Decision (Priority: P4)

For every request, produce measurable diagnostics: how many candidates
existed before budgeting, how many were kept, the estimated tokens
available versus selected versus excluded, and the resulting reduction
percentage — so the value of this whole capability can actually be
measured, not merely assumed.

**Why this priority**: The entire motivation for this project's own
"Second Brain" effort was measured token reduction, not a vague hope of
it — without this story, nobody could ever confirm the other three
stories are actually delivering value in practice. Depends on User
Story 1 and 2 both existing, since there's nothing meaningful to report
before ranking and budgeting actually happen.

**Independent Test**: Run a request against a known candidate set and a
known budget; confirm the reported counts and token estimates are
verifiably accurate against the actual result produced.

**Acceptance Scenarios**:

1. **Given** any completed request, **When** its result is inspected,
   **Then** it reports how many candidates were considered and how many
   were ultimately kept.
2. **Given** any completed request, **When** its result is inspected,
   **Then** it reports the estimated tokens available, selected, and
   excluded, and the resulting reduction percentage, all consistent
   with each other and with the actual result.
3. **Given** any completed request, **When** its result is inspected,
   **Then** every kept item still carries why it was included (per
   015's own reasons) alongside its own estimated token cost.

---

### Edge Cases

- A requested budget of zero, or a negative value — treated the same as
  an extremely small budget: mandatory content is still always
  returned, flagged as exceeding it.
- A candidate set consisting only of mandatory content, well within an
  ample budget — nothing is excluded; diagnostics report zero exclusion
  and zero reduction.
- Two candidates with genuinely identical relevance — a stable,
  deterministic order is still produced, never an arbitrary one that
  could vary between identical requests.
- A budget large enough that nothing is ever excluded — diagnostics
  still report accurate available/selected figures, not merely "not
  applicable."
- The same request repeated against unchanged project content — ranking,
  selection, and diagnostics are identical every time.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST assign every candidate a deterministic
  relevance score derived from its priority tier, its specific
  relationship kind, the request's own intent (if any), and its textual
  relevance where applicable — never a random or externally-learned
  score.
- **FR-002**: A candidate from a higher-priority tier (per 015's own
  ordering: mandatory, then structural, then semantic, then textual,
  then second-hop) MUST always rank above a candidate from a
  lower-priority tier, regardless of any individual relevance score
  difference.
- **FR-003**: Intent MAY influence ordering only among candidates
  already within the same priority tier — it MUST NOT let a
  lower-priority tier outrank a higher one.
- **FR-004**: System MUST accept an optional token budget as part of a
  request, using a fixed, documented default value when none is
  supplied.
- **FR-005**: Given a budget, System MUST return the smallest ranked set
  of candidates whose combined estimated token cost fits within it,
  filling the budget in priority-tier order — never by sorting
  everything by score alone and truncating wherever it runs out.
- **FR-006**: Mandatory context (the project's Constitution and the
  target artifact itself, per 015's own guarantee) MUST always be
  included in the result in full, regardless of the requested budget.
- **FR-007**: When mandatory context alone exceeds the requested budget,
  System MUST still return all of it, MUST flag the result as having
  exceeded the budget, and MUST report the estimated overage.
- **FR-008**: When mandatory context alone exceeds the requested budget,
  System MUST still be free to omit lower-priority optional content —
  exceeding budget for mandatory reasons never forces every optional
  item to be kept too.
- **FR-009**: System MUST report, for every item in the result, why it
  was included and its own estimated token cost.
- **FR-010**: System MUST report, for every request, the number of
  candidates considered, the number selected, the estimated tokens
  available, selected, and excluded, and the resulting reduction
  percentage.
- **FR-011**: System MUST produce identical ranking, selection, and
  diagnostics for the same request against the same, unchanged project
  state — no randomness, no dependency on anything beyond the request
  and current project content.
- **FR-012**: This capability MUST be strictly read-only — ranking and
  budgeting MUST NOT modify any project artifact, the reference graph,
  or the search index.

### Key Entities

- **Scored Candidate**: One of 015's own Candidates, now additionally
  carrying a deterministic relevance score used only to order items
  within its own priority tier — never across tiers (FR-002, FR-003).
- **Budget**: The token ceiling a request is measured against — a fixed
  default when the caller supplies none (FR-004).
- **Context Result**: The final, budgeted answer to one request — the
  kept items (each still labeled with why it was included and its own
  token cost), whether the budget was exceeded and by how much, and the
  overall diagnostics describing how much was considered versus kept.
- **Diagnostics**: The measurable counts and token estimates reported
  for every request (FR-010) — candidates considered, items selected,
  tokens available/selected/excluded, and the resulting reduction
  percentage.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a representative set of requests, 100% of mandatory
  and structural candidates outrank any purely textual match in the
  same result, regardless of the textual match's own score.
- **SC-002**: For a representative set of budgets, 100% of results
  fit within their requested budget whenever mandatory content alone
  does not already exceed it.
- **SC-003**: 100% of requests where mandatory content alone exceeds the
  budget still return that content in full, correctly flagged, with the
  overage reported.
- **SC-004**: 100% of requests report diagnostics that are verifiably
  consistent with the actual result produced (selected count matches
  the number of items returned; available minus excluded equals
  selected, within the estimator's own rounding).
- **SC-005**: Repeating the same request against unchanged project
  content produces byte-for-byte identical ranking, selection, and
  diagnostics, 100% of the time.
- **SC-006**: Zero project artifacts, reference-graph data, or
  search-index data are ever modified by a ranking/budgeting request,
  verified across every tested scenario.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 7 ("Ranking and
  Budgeting"): scoring 015's own candidate set and fitting it into a
  token budget, with diagnostics. It explicitly excludes that
  document's own later phase — the internal context command and any
  rendering of results (Phase 8) — which remains its own, later
  feature, continuing this project's established one-phase-at-a-time
  discipline (every prior feature, 001 through 015, has kept an
  equivalent boundary).
- The exact scoring formula and its specific weights are an
  implementation detail left to planning — this specification requires
  only the *ordering guarantee* (FR-002, FR-003) to hold, not any
  particular numeric value, matching the source document's own explicit
  guidance that tests should focus on ordering guarantees rather than
  hard-coded weights.
- The default token budget is a fixed, internal value — not a new
  user-facing configuration option — consistent with this project's own
  minimal-configuration discipline; it may be adjusted later once real
  usage provides evidence for a better default.
- Token cost for any item is computed using 013's own already-existing
  estimator — this feature introduces no new estimation method.
- This feature decides *what* belongs in the final result and *why* —
  it does not decide how that result is rendered, returned over a CLI,
  or consumed; Phase 8's own later "internal context" command is where
  that happens.
