# Feature Specification: Wikilink Graph Foundation

**Feature Branch**: `011-wikilink-foundation`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir com a sugestão 'phase 1:
wikilinks/graph foundation'" (proceed with Phase 1 — Artifact Graph
Foundation — of `docs/context-engine-implementation.md`, itself
`docs/architecture-specification.md`'s Phase 7 "Second Brain")

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-009, this feature's "users" are misterspec's own
  deterministic layer and future callers — the existing "internal
  validate" capability (008-cli-cobra) and, eventually, canonical
  Skills and Phase 2's reference/backlink operations — rather than an
  end-user UI of its own. This is the foundation everything else in
  Phase 7 (Second Brain) builds on; per
  context-engine-implementation.md's own Phase 1 scoping, it covers
  wikilink parsing and validation only — no SQLite index, no
  reference/backlink query operations (Phase 2), no context retrieval.
-->

### User Story 1 - Author Explicit Semantic Links Between Artifacts (Priority: P1)

An author (a human or a coding agent) writes an explicit reference from
one artifact's body to another artifact's ID — for example, a Spec's
"Relevant Knowledge" section linking to the Knowledge artifact it
actually depends on — and misterspec can deterministically recognize
and parse every such link, in document order, with its exact source
location, without guessing or inferring a relationship from prose.

**Why this priority**: This is the smallest foundational slice —
nothing else in this feature (or in Phase 2's later reference/backlink
operations) can exist without the ability to parse these links first.
It is also independently useful on its own: simply being able to
enumerate an artifact's own explicit links deterministically already
has value.

**Independent Test**: Write an artifact body containing several links —
a plain link, an aliased link, multiple links on one line, links across
multiple lines, and text that looks similar but isn't a valid link
(standard Markdown links, inline code, a single bracket) — and confirm
every valid link is extracted in document order with the correct
target, alias, and line number, and nothing else is misinterpreted as
one.

**Acceptance Scenarios**:

1. **Given** an artifact body containing `[[SPEC-014]]`, **When** its
   links are extracted, **Then** exactly one link is found, targeting
   `SPEC-014`, with no alias.
2. **Given** an artifact body containing `[[SPEC-014|Refresh Token
   Rotation]]`, **When** its links are extracted, **Then** exactly one
   link is found, targeting `SPEC-014`, with that display alias.
3. **Given** an artifact body containing several links across multiple
   sections, **When** its links are extracted, **Then** they are
   returned in the same order they appear in the document, each with
   its own correct source line.
4. **Given** an artifact body containing a standard Markdown link, an
   inline code span showing example link syntax, or a single bracket
   that is not a link, **When** its links are extracted, **Then** none
   of those are returned as links.

---

### User Story 2 - Detect Broken, Malformed, and Ambiguous Links (Priority: P2)

When an artifact contains a link whose target doesn't exist, isn't a
valid reference, or resolves to more than one artifact, misterspec's
existing structural validation capability reports it as a specific,
distinct finding — the same way it already reports every other
structural problem — so the problem is caught before it misleads later
work.

**Why this priority**: An unchecked link is worse than no link at all —
it silently promises a relationship that may not be real. This is what
makes the links User Story 1 parses actually trustworthy enough to
build on. It depends on User Story 1's parsing already existing.

**Independent Test**: Create artifacts containing a link to a real,
existing target; a link to an ID that doesn't exist; a link with
malformed syntax; and (where constructible) a link whose target is
ambiguous. Run the project's existing structural validation and confirm
each produces its own specific, correctly labeled finding — and the
valid link produces none.

**Acceptance Scenarios**:

1. **Given** an artifact with a link to an existing, resolvable target,
   **When** the project is validated, **Then** that link produces no
   finding.
2. **Given** an artifact with a link to an ID that does not exist,
   **When** the project is validated, **Then** a finding specifically
   identifying a broken link is reported, naming the artifact and the
   unresolved target.
3. **Given** an artifact with malformed link syntax, **When** the
   project is validated, **Then** a finding specifically identifying a
   malformed link is reported, distinguishable from a broken link.
4. **Given** an artifact with a link whose target resolves ambiguously,
   **When** the project is validated, **Then** a finding specifically
   identifying an ambiguous link is reported, distinguishable from both
   of the above.

---

### User Story 3 - Existing Artifacts and Validation Remain Unaffected (Priority: P3)

Every artifact that contains no links continues to be created, edited,
and validated exactly as it was before this capability existed — this
feature adds a new, optional way to express a relationship; it changes
nothing about artifacts that don't use it.

**Why this priority**: This guards against the most likely way a
feature like this goes wrong — quietly changing behavior for the
overwhelming majority of artifacts that will never contain a link. It
is the lowest-risk but still essential guarantee, verified once User
Story 1 and 2 exist by confirming nothing else changed.

**Independent Test**: Run the full existing suite of project and
artifact validation scenarios (from every prior capability) unmodified,
and confirm every result is identical to before this feature existed.

**Acceptance Scenarios**:

1. **Given** an artifact with no links in its body, **When** it is
   created, inspected, or validated, **Then** the result is identical
   to what it would have been before this feature existed.
2. **Given** a project where no artifact uses a link anywhere, **When**
   the whole project is validated, **Then** the result is identical to
   before this feature existed.

---

### Edge Cases

- A link's target uses a different ID zero-padding width than the
  project's currently configured width (must still resolve correctly,
  matching how an artifact's existing frontmatter fields already
  tolerate this today).
- A link appears inside a fenced code block or inline code span (for
  example, documentation showing example link syntax) — must not be
  treated as a real link.
- A link appears in an artifact's frontmatter block rather than its
  body — out of scope for this feature's extraction (frontmatter has
  its own, already-established parsing).
- An artifact type with no body content at all, or an empty body.
- Two links on the same line, or a link immediately followed by other
  text with no separating whitespace.
- A link's target syntax is present but incomplete or unterminated
  (for example, a stray `[[` with no closing `]]`) — must be treated as
  not a link, not as a malformed one, since no complete link syntax was
  actually written.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support an explicit, author-written syntax for
  linking from one artifact's body to another artifact's ID, with an
  optional display alias distinct from the target itself.
- **FR-002**: System MUST extract every such link from an artifact's
  body in document order, each with the target it names, its alias (if
  any), and its exact source line.
- **FR-003**: System MUST NOT interpret a standard Markdown link, text
  inside inline code or a fenced code block, or incomplete/unterminated
  link syntax as a valid link.
- **FR-004**: System MUST treat link extraction and link target
  validation as separate steps — extraction never fails or is skipped
  because a target does not (yet) exist.
- **FR-005**: System MUST validate a link's target using the exact same
  entity-ID rules the project already uses everywhere else — never a
  second, independently defined interpretation of what a valid ID is.
- **FR-006**: System's existing structural validation capability MUST be
  able to report, for any given project: every link pointing to a
  target that does not exist, distinctly labeled from every other kind
  of structural problem.
- **FR-007**: System's existing structural validation capability MUST be
  able to report every link with malformed syntax, distinctly labeled
  from a broken (nonexistent-target) link.
- **FR-008**: System's existing structural validation capability MUST be
  able to report every link whose target resolves ambiguously,
  distinctly labeled from both a broken and a malformed link.
- **FR-009**: System MUST NOT require any artifact to contain a link —
  an artifact with none MUST produce identical validation results to
  what it produced before this capability existed.
- **FR-010**: System MUST NOT automatically insert, modify, or remove a
  link in any artifact under any circumstance — creating one remains an
  explicit authoring decision, by a human or a coding agent, never a
  side effect of any deterministic operation.

### Key Entities

- **Link**: An explicit, author-written semantic reference from one
  artifact's body to another artifact's ID, with an optional display
  alias — distinct from the project's existing formal relationships
  (parent, depends_on, supersedes, for), which retain their own,
  unchanged meaning.
- **Link Finding**: A structural validation result naming a specific
  link-integrity problem (broken, malformed, or ambiguous), tied to the
  exact artifact and location where it occurs — reported the same way
  every other structural finding already is.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every link in a representative set of artifact bodies
  (plain, aliased, multi-line, mixed with non-link syntax) is extracted
  with 100% accuracy against a known-correct expected set.
- **SC-002**: 100% of links pointing to a nonexistent target are caught
  by structural validation across tested fixtures — zero false
  negatives.
- **SC-003**: 100% of artifacts and validation scenarios that existed
  before this capability was added continue to produce identical
  results afterward — zero regressions.
- **SC-004**: A broken link, a malformed link, and an ambiguous link are
  each distinguishable from one another and from every other existing
  finding type by reading the validation result alone, with zero need
  to inspect source code to understand which specific problem occurred.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 1 ("Artifact
  Graph Foundation"): link parsing and structural validation of link
  targets. It explicitly excludes that document's later phases — the
  `references`/`backlinks` query operations (Phase 2), the disposable
  search index (Phase 4+), Markdown chunking, ranking, budgeting, and
  the context-retrieval command (later phases) — each remains its own,
  later feature, continuing this project's established one-phase-at-
  a-time discipline (every prior feature, 001 through 010, has kept an
  equivalent boundary).
- Link syntax follows `[[TARGET]]` and `[[TARGET|Alias]]`, where TARGET
  must be a misterspec entity ID — never an artifact's title or its
  filesystem path — matching context-engine-implementation.md §6.1's
  own explicit requirement, so links stay stable if a title or path
  later changes.
- Text inside a fenced code block or inline code span is never
  interpreted as a link — the common convention this kind of link
  syntax already follows elsewhere (for example, Obsidian's own
  behavior), and necessary so an artifact documenting example link
  syntax doesn't itself trigger false findings.
- Link validation extends the project's already-existing structural
  validation capability — it does not introduce a second, parallel
  validation surface or a new user-facing command.
- This feature does not change how any of the project's existing formal
  relationships (parent, depends_on, supersedes, for) are parsed,
  validated, or reported — links are a distinct, additive, optional
  relationship category layered alongside them, never a replacement.
- The consumers ("users") of this feature are, in priority order: (1)
  misterspec's own existing structural validation capability, (2) a
  future reference/backlink capability (Phase 2) that will consume
  these same parsed links, and (3) the human or coding agent authoring
  an artifact — the same framing every prior feature in this project
  has used.
