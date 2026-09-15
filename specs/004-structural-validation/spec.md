# Feature Specification: Structural Validation and Project Status

**Feature Branch**: `004-structural-validation`
**Created**: 2026-09-11
**Status**: Draft
**Input**: User description: "pode seguir com a sugestão da fase 2 acima"
(proceed with the remaining Phase 2 read-only deterministic operations
recommended after 003-entity-creation: `validate` and `status`)

## User Scenarios & Testing *(mandatory)*

<!--
  As with the prior three features, this has no end-user UI. Its "users"
  are the coding agent and misterspec's own future CLI layer, which — per
  docs/architecture-specification.md §16-19 — must call a structural
  validation primitive instead of an LLM eyeballing whether a project
  "looks right." Every story composes 001-core-foundation's,
  002-read-operations's, and 003-entity-creation's existing primitives
  rather than re-implementing structural checks those already make
  possible.
-->

### User Story 1 - Validate a Single Entity (Priority: P1)

Given one entity's ID, any consumer can run every structural check that
applies to it — frontmatter well-formedness, required fields, ID syntax,
declared lifecycle state validity, parent existence and correct type, and
(for a Spec) that every dependency reference resolves — and get back the
complete list of problems found, or an empty list if there are none.

**Why this priority**: This is the single-entity check every Skill that
just created or modified one artifact needs to run before moving on
(`/create-feature`, `/create-specs`, `/create-plan` all validate the one
entity they just touched per their own contracts in
`docs/architecture-specification.md` §43-47). It is also the smallest
independently useful slice — project-wide validation (User Story 2) is
this same logic applied to every entity, not different logic.

**Independent Test**: Can be fully tested by populating a fixture project
with a mix of well-formed and deliberately broken entities (a missing
required field, a nonexistent parent, a parent of the wrong type, an
invalid status value, a Spec depending on a Spec that doesn't exist) and
confirming each one's validation returns exactly the findings that apply
to it, and that a well-formed entity returns none — no other story's code
required.

**Acceptance Scenarios**:

1. **Given** a well-formed Spec with a valid parent and resolvable
   dependencies, **When** it is validated, **Then** an empty findings list
   is returned and the run itself is reported as successful.
2. **Given** an entity whose declared parent ID does not exist, **When**
   it is validated, **Then** a finding naming that specific problem
   (code, path, message) is returned.
3. **Given** an entity whose declared parent exists but is the wrong
   type (e.g. a Spec's parent resolves to a Program instead of a
   Feature), **When** it is validated, **Then** a finding reporting the
   type mismatch is returned.
4. **Given** an entity with a declared status value that is not one of
   the allowed lifecycle states for its type, **When** it is validated,
   **Then** a finding reporting the invalid state is returned.
5. **Given** a Spec whose `depends_on` names an ID that does not exist,
   **When** it is validated, **Then** a finding reporting the unresolved
   dependency is returned.
6. **Given** an entity with more than one problem at once, **When** it is
   validated, **Then** every applicable finding is returned together, not
   only the first one encountered.

---

### User Story 2 - Validate the Whole Project (Priority: P2)

Any consumer can run structural validation across every entity in the
project in a single call, receiving the complete, aggregated list of every
finding from every entity — including problems only visible at the
whole-project level, like two different artifacts claiming the same ID.

**Why this priority**: This is what a Skill or a human runs before trusting
the project's structure broadly — after a batch of changes, or as a
sanity check. It builds directly on User Story 1 (the same per-entity
checks), extended to also catch cross-entity problems no single-entity
validation could see on its own.

**Independent Test**: Can be fully tested against a fixture project mixing
several valid entities, several individually-broken ones, and two entities
of the same type sharing one ID under different parents — confirming the
aggregated findings list contains every individual-entity problem plus the
duplicate-ID problem, and that a project with zero problems returns zero
findings.

**Acceptance Scenarios**:

1. **Given** a project where every entity is well-formed, **When** the
   whole project is validated, **Then** an empty findings list is
   returned.
2. **Given** a project containing several entities with different,
   individual problems, **When** the whole project is validated, **Then**
   every one of those findings is present in the result.
3. **Given** two artifacts of the same type independently claiming the
   same numeric ID under different parents, **When** the whole project is
   validated, **Then** a duplicate-ID finding is returned for that
   number.
4. **Given** a freshly initialized project with no entities at all,
   **When** it is validated, **Then** an empty findings list is returned,
   not an error.

---

### User Story 3 - Get an Aggregate Project Status Summary (Priority: P3)

Any consumer can request a single summary of the project's current health:
how many entities exist of each type, how many Specs are in each lifecycle
state, and how many structural findings the project currently has —
without needing to run and parse a full validation pass to get that
overview.

**Why this priority**: This is a fast, cheap "how's the project doing"
check — useful on its own, and independent of User Story 1/2's detailed
findings, though it reuses the same counting/validation machinery
underneath.

**Independent Test**: Can be fully tested against a fixture project with a
known mix of Programs, Features, and Specs in various lifecycle states,
confirming the reported counts match exactly, and that the
structural-finding count matches what a full validation run would report
for the same project.

**Acceptance Scenarios**:

1. **Given** a project with a known number of Programs, Features, and
   Specs, **When** status is requested, **Then** the reported counts match
   exactly.
2. **Given** Specs in several different lifecycle states, **When** status
   is requested, **Then** the per-state Spec counts match exactly.
3. **Given** a project with known structural problems, **When** status is
   requested, **Then** the reported structural-finding count matches the
   number of findings a full project validation would return.
4. **Given** a project with zero entities, **When** status is requested,
   **Then** every count is reported as zero, not an error.

---

### Edge Cases

- A project with zero entities of any kind (freshly initialized).
- An entity whose parent field is syntactically well-formed but names an
  ID that does not exist anywhere in the project.
- An entity whose parent exists but is the wrong entity type.
- Two artifacts of the same type sharing the same numeric ID under
  different parents, discovered only at the whole-project level.
- An artifact whose frontmatter `id` field names a different ID than the
  one implied by its own canonical directory or filename.
- An entity with more than one problem simultaneously (e.g. an invalid
  status value *and* a nonexistent parent).
- A Spec's `depends_on` or `supersedes` list naming an ID that is
  syntactically valid but does not resolve to any existing Spec.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST validate a single named entity, checking at
  least: frontmatter well-formedness, required fields for its type, entity
  ID syntax, declared lifecycle state validity for its type, and — where
  applicable — parent existence and correct parent type.
- **FR-002**: For a Spec, System MUST additionally check that every
  `depends_on` and `supersedes` entry resolves to an existing Spec.
- **FR-003**: System MUST validate the entire project by applying every
  single-entity check to every discovered entity in one pass, plus
  detecting duplicate IDs across the whole project.
- **FR-004**: A validation run — single-entity or whole-project — MUST
  return every finding it discovers; it MUST NOT stop at the first
  problem found.
- **FR-005**: Each finding MUST carry a stable code, a severity, the
  affected path, and a specific, human-readable message.
- **FR-006**: System MUST distinguish "the validation command completed"
  from "the project structure is valid" as two independently reportable
  facts — a run that discovers real problems is still a successfully
  completed run, not a failed one.
- **FR-007**: System MUST detect an entity whose declared frontmatter ID
  does not match the ID implied by its own canonical filesystem location.
- **FR-008**: System MUST detect a declared lifecycle status value that is
  not among the fixed set of allowed states for that entity's type.
- **FR-009**: System MUST detect a declared parent that does not exist, or
  that exists as an entity type other than the one required.
- **FR-010**: System MUST detect two or more artifacts of the same entity
  type independently claiming the same numeric ID, during whole-project
  validation.
- **FR-011**: System MUST return an empty findings list — not an error —
  for an entity or project with no problems.
- **FR-012**: System MUST report, in a single call: the count of entities
  by type, the count of Specs by lifecycle state, and the total count of
  structural findings currently present in the project.
- **FR-013**: A status summary MUST be computed from the project's current
  filesystem state at the moment it is requested — never a cached or
  precomputed value.
- **FR-014**: Every operation in this feature MUST be read-only: no
  validation or status run may create, modify, or delete any file.

### Key Entities

- **Finding**: One detected structural problem — its stable code,
  severity, the affected artifact's path, and a specific message.
- **Validation Result**: The outcome of a validation run — the complete
  list of Findings (possibly empty) for the entity or project validated.
- **Status Summary**: An aggregate snapshot — entity counts by type, Spec
  counts by lifecycle state, and the total structural-finding count.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For any entity with N distinct problems, a single validation
  call returns all N findings, never fewer.
- **SC-002**: A project with zero structural problems always reports zero
  findings and zero structural errors in its status summary — 100% of the
  time, across projects of any size.
- **SC-003**: Every duplicate ID present in a project is caught by a
  whole-project validation run, with zero false negatives.
- **SC-004**: A status summary's entity and state counts exactly match a
  direct count of the same project's artifacts, 100% of the time.
- **SC-005**: Across all validation and status operations, zero files are
  ever created, modified, or deleted as a side effect.

## Assumptions

- This feature builds directly on `001-core-foundation`'s,
  `002-read-operations`'s, and `003-entity-creation`'s packages — no new
  external dependency, still no CLI/JSON output surface.
- The allowed lifecycle states per entity type are exactly those already
  frozen in `docs/architecture-specification.md`: Program/Feature —
  `draft, active, done, cancelled`; Spec — `draft, ready, in_progress,
  validated, blocked, superseded, cancelled`; Learning — `candidate,
  promoted, dismissed`. Knowledge's only state named anywhere in that
  document is `active`; this feature treats `active` as the only allowed
  Knowledge state for now, and that set can be extended later without
  breaking anything if a real need for more states appears.
- This feature explicitly does **not** implement checks that require
  parsing an artifact's Markdown *body* content rather than its
  frontmatter or filesystem structure — specifically: duplicate
  requirement markers inside a Spec, and a Task's references to
  requirement markers. Those depend on the same body-parsing capability
  the deferred `references` operation
  (`docs/architecture-specification.md` §15, already flagged as
  out-of-scope in `002-read-operations`'s research) would need, and are
  left to a later feature rather than built piecemeal here.
- Similarly, "missing Plan/Tasks/Validation where Spec state requires
  one" is deferred: the architecture specification does not pin down
  which specific Spec states require which subordinate artifact, and
  resolving that ambiguity productively belongs together with the
  deferred body-content checks above, not as a guess made in isolation
  here.
- The consumers ("users") of this feature are, in priority order: (1) the
  coding agent and the Skills it runs, indirectly via (2) misterspec's own
  future deterministic-operations CLI layer — unchanged from the prior
  three features.
