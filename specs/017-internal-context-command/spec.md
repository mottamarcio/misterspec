# Feature Specification: Internal Context Command

**Feature Branch**: `017-internal-context-command`
**Created**: 2026-09-15
**Status**: Draft
**Input**: User description: "considerando os arquivos salvos em @docs/
poderia criar a spec da 'Phase 8 — Internal Context Command' do arquivo
@docs/context-engine-implementation.md ?" (create the spec for Phase 8
— Internal Context Command — of `docs/context-engine-implementation.md`,
itself `docs/architecture-specification.md`'s Phase 7 "Second Brain",
exposing 016-ranking-budgeting's own budgeted Context Result through
the existing "internal" command surface)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-016, this feature's "users" are misterspec's own
  deterministic layer and its external callers — a coding agent, a
  Skill, or a script driving the misterspec binary — rather than an
  end-user UI of its own. Per context-engine-implementation.md's own
  Phase 8 scoping, this covers exposing 016's own budgeted Context
  Result as a stable, machine-readable command: no new ranking logic,
  no new budgeting logic, no Skill integration (Phase 10), and the
  optional Markdown `--render` mode only after the JSON shape is
  stable.
-->

### User Story 1 - Request a Context Pack for a Target (Priority: P1)

Given the ID of an existing artifact, a caller can request a budgeted,
ranked context pack for it in one call — naming the target, an intent
(planning, tasks, implementation, validation, or analysis), and
optionally a current task and a free-text query — and receive back
016's own Context Result as stable, structured, machine-readable
output, without separately invoking the collector, ranker, and budgeter
themselves.

**Why this priority**: This is the entire point of Phase 8 — until the
already-built retrieval, ranking, and budgeting pipeline (015, 016) is
reachable through one command, none of it delivers any value to an
actual caller. Every other story in this feature refines what this one
command returns or reports.

**Independent Test**: Invoke the context command against a known
target and intent in a small fixture project; confirm the returned
result matches what directly invoking 016's own ranking/budgeting
pipeline would produce for the same request.

**Acceptance Scenarios**:

1. **Given** a valid target artifact ID and an intent, **When** the
   context command is invoked, **Then** it returns a successful,
   structured result containing the budgeted, ranked context items for
   that target and intent.
2. **Given** a valid target ID, an intent, a current task ID, and a
   free-text query, **When** the context command is invoked, **Then**
   all four inputs are passed through to the underlying request, and
   the returned result reflects their influence on the selected
   context.
3. **Given** no task, no query, or no budget is supplied, **When** the
   context command is invoked, **Then** the command still succeeds,
   using the documented defaults for each omitted input.
4. **Given** the same target, intent, task, query, and budget invoked
   twice against unchanged project content, **When** both results are
   compared, **Then** they are identical.

---

### User Story 2 - Get a Clear, Actionable Error for Invalid Input (Priority: P2)

Given a request naming a target that does not exist, is ambiguous, or
is malformed, or naming an unsupported intent, or supplying an invalid
budget, a caller receives a clear, machine-readable error identifying
exactly what was wrong — never a successful-looking result built from
bad input, and never an unhandled crash.

**Why this priority**: A caller (often an automated agent, not a human
reading a terminal) must be able to branch reliably on failure without
guessing from prose — this is what makes the command safe to depend on
programmatically. It depends on User Story 1's own request shape
already being defined, since these are that same request's failure
paths.

**Independent Test**: Invoke the context command with a nonexistent
target ID, then with an unsupported intent value, then with a negative
budget; confirm each produces a distinct, well-formed error result
rather than a success result or a crash.

**Acceptance Scenarios**:

1. **Given** a target ID that does not resolve to any existing
   artifact, **When** the context command is invoked, **Then** it
   returns a well-formed error identifying the target as not found.
2. **Given** a target ID that resolves ambiguously, **When** the
   context command is invoked, **Then** it returns a well-formed error
   identifying the ambiguity.
3. **Given** an intent value outside the documented supported set,
   **When** the context command is invoked, **Then** it returns a
   well-formed error identifying the unsupported intent.
4. **Given** an invalid budget value (for example, negative), **When**
   the context command is invoked, **Then** it returns a well-formed
   error identifying the invalid budget, without silently substituting
   a default.
5. **Given** any of the above error conditions, **When** the command's
   output is inspected, **Then** it never resembles a successful
   result and never modifies any project artifact.

---

### User Story 3 - Trust the Index Stays Usable Without Manual Repair (Priority: P3)

Given the underlying disposable search index is missing, stale, or
incompatible, a caller invoking the context command still receives a
correct result — the command transparently synchronizes or rebuilds
whatever derived state it needs first, rather than failing and
requiring the caller to run a separate repair step by hand.

**Why this priority**: The context command is meant to be the single
entry point a caller relies on; if it could fail merely because a
disposable cache happened to be stale or absent, every caller would
need to reimplement index-repair logic themselves, undermining Phase
8's own purpose as a stable, self-sufficient entry point. Depends on
User Story 1's own successful path already existing to synchronize
before.

**Independent Test**: Delete the disposable index entirely, then invoke
the context command against a target that exists on disk; confirm the
command still returns a correct, successful result, having rebuilt
whatever derived state it needed.

**Acceptance Scenarios**:

1. **Given** the disposable index does not yet exist, **When** the
   context command is invoked, **Then** it transparently builds it and
   still returns a correct result.
2. **Given** the disposable index is stale relative to current
   artifact content, **When** the context command is invoked, **Then**
   it transparently synchronizes the index and returns a result
   reflecting current content.
3. **Given** the disposable index has an incompatible schema version,
   **When** the context command is invoked, **Then** it transparently
   rebuilds the index rather than failing, and still returns a correct
   result.
4. **Given** any of the above repair paths runs, **When** the command
   completes, **Then** no authoritative project artifact is ever
   modified as a result.

---

### User Story 4 - Render a Human/Agent-Readable Context Pack (Priority: P4)

Given the same request as User Story 1, a caller can optionally
request the result rendered as a compact, readable Markdown context
pack instead of (or alongside) the structured result — grouping the
returned items under clear section headings by tier and origin — for
direct inclusion in an agent's own working context.

**Why this priority**: The structured result (User Story 1) is the
canonical, stable interface and must exist and be trustworthy first;
rendering is a presentation convenience on top of it, explicitly
deferred until after the structured shape is stable, per this
project's own phased delivery discipline. Depends entirely on User
Story 1.

**Independent Test**: Invoke the context command in rendered mode
against a known request; confirm the output is well-formed Markdown
whose sections correspond exactly to the same items the structured
result (User Story 1) would return for the identical request.

**Acceptance Scenarios**:

1. **Given** a valid request, **When** the context command is invoked
   with rendering requested, **Then** it returns well-formed Markdown
   containing every item present in the equivalent structured result,
   grouped under clear section headings.
2. **Given** rendering is not requested, **When** the context command
   is invoked, **Then** the structured result is returned unchanged,
   with no rendered output included.
3. **Given** a request that would produce an error under User Story 2,
   **When** rendering is requested, **Then** the same well-formed
   error is returned instead of a rendered pack.

---

### Edge Cases

- A target ID that is syntactically well-formed but for an entity type
  that does not participate in context retrieval — treated as a normal
  not-found condition, not a crash.
- A budget of zero, or an extremely small budget — handled exactly as
  016 already defines: mandatory content is still returned in full,
  flagged as exceeding budget.
- A query string that is empty, or consists only of whitespace —
  treated as no query supplied, falling back to intent- and
  structure-driven retrieval alone.
- A task ID supplied that does not belong to the target's own Spec, or
  does not exist at all — reported as an invalid input rather than
  silently ignored.
- Concurrent invocations of the command against the same project while
  artifacts are being edited — each invocation still produces a
  correct result for whatever content it observes; no invocation
  corrupts the index for another.
- The command invoked from a project with zero indexable artifacts —
  succeeds with an empty, well-formed result rather than an error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose one internal command that accepts a
  target artifact ID, an intent, an optional current task ID, an
  optional free-text query, and an optional token budget, and returns
  016's own budgeted, ranked Context Result for that request.
- **FR-002**: System MUST support exactly the documented set of
  intents (planning, tasks, implementation, validation, analysis) and
  MUST reject any other intent value with a clear, distinct error.
- **FR-003**: When task, query, or budget are omitted, System MUST
  apply documented defaults (016's own default budget; no task and no
  query otherwise influencing retrieval) rather than failing.
- **FR-004**: System MUST reject a target ID that does not resolve to
  exactly one existing artifact with a clear, distinct error
  identifying whether the target was not found or was ambiguous.
- **FR-005**: System MUST reject an invalid budget (for example,
  negative) with a clear, distinct error, and MUST NOT silently
  substitute the default budget in its place.
- **FR-006**: System MUST return output in a stable, machine-readable,
  structured format for both successful results and errors, following
  this project's existing internal-command envelope conventions.
- **FR-007**: System MUST ensure the disposable search index the
  request depends on is present, current, and schema-compatible before
  serving the request — transparently building, synchronizing, or
  rebuilding it as needed — without requiring the caller to invoke any
  separate repair step.
- **FR-008**: This capability MUST be strictly read-only with respect
  to authoritative project artifacts — invoking it, including any
  index build/synchronize/rebuild it triggers, MUST NOT modify any
  Markdown artifact or its frontmatter.
- **FR-009**: System MUST produce identical structured output for the
  same request (same target, intent, task, query, and budget) against
  the same, unchanged project state — no randomness, no dependency on
  invocation order or prior cache state.
- **FR-010**: System MUST support an optional rendering mode that
  converts a successful Context Result into a well-formed, readable
  Markdown context pack grouping items by tier/origin, containing
  every item present in the equivalent structured result.
- **FR-011**: The rendering mode MUST NOT alter which items are
  selected or how they are ranked or budgeted — it MUST only change
  the presentation of an already-computed Context Result.
- **FR-012**: When a request would fail (per FR-002, FR-004, or
  FR-005), System MUST return the same well-formed error regardless of
  whether rendering was requested, never a partially rendered or
  partially successful output.

### Key Entities

- **Context Command Request**: The caller-supplied input to the
  command — target ID, intent, optional task ID, optional query,
  optional budget, and whether rendering is requested — validated
  before being handed to 016's own request model.
- **Context Command Result**: The command's own output — either 016's
  own structured Context Result (optionally accompanied by its
  rendered Markdown form) on success, or a well-formed, distinct error
  identifying exactly what was wrong with the request.
- **Index Readiness**: The precondition this command establishes
  before serving a request — the disposable search index existing,
  synchronized with current artifact content, and at a compatible
  schema version — established transparently, never exposed as a
  separate step the caller must perform.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a representative set of valid requests, 100% return
  a successful, structured result matching what invoking the
  underlying ranking/budgeting pipeline directly would produce for the
  same inputs.
- **SC-002**: For a representative set of invalid requests (unknown
  target, ambiguous target, unsupported intent, invalid budget), 100%
  return a distinct, well-formed error rather than a success result or
  a crash.
- **SC-003**: 100% of requests issued against a missing, stale, or
  schema-incompatible index still return a correct result, with zero
  requests failing purely due to index staleness or absence.
- **SC-004**: Repeating the same request against unchanged project
  content produces byte-for-byte identical structured output, 100% of
  the time.
- **SC-005**: Zero authoritative project artifacts are ever modified
  by any invocation of this command, verified across every tested
  scenario including index rebuild paths.
- **SC-006**: For a representative set of successful requests, 100% of
  rendered outputs contain every item present in the equivalent
  structured result, with no items added, dropped, re-ranked, or
  re-budgeted by rendering.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 8 ("Internal
  Context Command"): exposing 016-ranking-budgeting's own Context
  Result through one internal command, with an optional Markdown
  rendering mode added only once the structured shape is stable. It
  explicitly excludes that document's own later phases — dogfooding
  and evaluation (Phase 9) and Skill integration (Phase 10) — which
  remain their own, later features, continuing this project's
  established one-phase-at-a-time discipline (every prior feature, 001
  through 016, has kept an equivalent boundary).
- The command's exact JSON field names and envelope shape are an
  implementation detail left to planning — this specification requires
  only that the output be stable, structured, machine-readable, and
  consistent with this project's existing internal-command
  conventions (FR-006), not any particular field naming.
- "Transparently" ensuring index readiness (FR-007) means the command
  itself triggers whatever build/synchronize/rebuild 014 and 015
  already define — this feature introduces no new indexing or
  synchronization logic of its own, only the orchestration that calls
  it before serving a request.
- The rendering mode's exact Markdown formatting is an implementation
  detail left to planning — this specification requires only that it
  be well-formed, readable, and faithful to the underlying structured
  result (FR-010, FR-011), not any particular heading text or layout.
- Supported intents match the set already defined by 015's own request
  model (planning, tasks, implementation, validation, analysis); this
  feature introduces no new intents.
- This feature does not change how any existing Skill behaves — Skills
  continue operating exactly as before until the later, explicit Skill
  Integration phase (Phase 10) deliberately adopts this command.
