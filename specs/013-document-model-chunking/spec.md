# Feature Specification: Document Model and Chunking

**Feature Branch**: `013-document-model-chunking`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir com a sugestão 'phase 3: document
model and chunking'" (proceed with Phase 3 — Document Model and
Chunking — of `docs/context-engine-implementation.md`, itself
`docs/architecture-specification.md`'s Phase 7 "Second Brain", building
on 011-wikilink-foundation's body/wikilink parsing and
012-references-backlinks' reference graph)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-012, this feature's "users" are misterspec's own
  deterministic layer and future callers — the disposable search index
  (Phase 4) and the context collector (Phase 6) that will consume its
  output, rather than an end-user UI of its own. Per
  context-engine-implementation.md's own Phase 3 scoping, this covers
  turning an artifact's body into a structured, chunked representation
  with provenance and a token estimate — no SQLite, no indexing, no
  search, no ranking, no context retrieval.
-->

### User Story 1 - See an Artifact's Body as Its Real Structure, Not a Blob of Text (Priority: P1)

Given any artifact this project manages, its Markdown body is turned
into a structured representation — an ordered list of sections, each
with its own heading, nesting level, and content — instead of one
undifferentiated block of text, so later capabilities can reason about
"the Acceptance Scenarios section" or "the third requirement" directly.

**Why this priority**: This is the foundational data model everything
else in this feature (and Phase 4's index) builds on — nothing can be
chunked or estimated until the artifact's real structure is known.

**Independent Test**: Take an artifact whose body has several headings
at different nesting levels, some with content and some without;
convert it to its structured form and confirm every heading appears in
document order, at its correct level, with exactly the content that
belongs to it — no more, no less.

**Acceptance Scenarios**:

1. **Given** an artifact body with several top-level and nested
   headings, **When** it is converted to its structured form, **Then**
   every heading appears, in document order, each correctly labeled
   with its own level and exactly the content between it and the next
   heading.
2. **Given** an artifact body with no headings at all (plain prose),
   **When** it is converted, **Then** the entire body still appears as
   one section, rather than being lost or rejected as an error.
3. **Given** an artifact body containing a fenced code block that itself
   shows example heading syntax, **When** it is converted, **Then** the
   text inside the code block is never mistaken for a real section
   boundary.
4. **Given** an artifact with a completely empty body, **When** it is
   converted, **Then** the result is a well-formed structure with no
   sections — not an error.

---

### User Story 2 - Break an Artifact Into Traceable, Retrieval-Sized Pieces (Priority: P2)

Given an artifact's structured form (User Story 1), it is broken into
individual pieces aligned to its own natural sections — each piece
small enough to be useful on its own, and each one able to point back
to exactly where it came from (which artifact, which heading, which
lines) so a later capability can explain why it was surfaced.

**Why this priority**: This is what actually makes an artifact's
content usable by a future search/retrieval capability (Phase 4+) — a
whole Spec is too large and undifferentiated to search or rank
meaningfully; individually traceable pieces are. Depends on User Story
1's structure existing first.

**Independent Test**: Take a structured artifact with several sections,
including one with no content between its heading and the next; break
it into pieces and confirm each non-empty section produces exactly one
traceable piece, no piece is produced for the empty section, and
running the same operation again on the same unchanged artifact
produces an identical result.

**Acceptance Scenarios**:

1. **Given** a structured artifact with three non-empty sections,
   **When** it is broken into pieces, **Then** exactly three pieces
   result, each one naming its own originating artifact, heading, and
   line range.
2. **Given** a section with a heading but no content before the next
   heading, **When** the artifact is broken into pieces, **Then** no
   empty piece is produced for it.
3. **Given** the same, unchanged artifact processed twice, **When** the
   results are compared, **Then** they are identical — same pieces,
   same order, same content.
4. **Given** a piece whose content contains an explicit link to another
   artifact (011's wikilink syntax), **When** the piece is produced,
   **Then** the link's original text is preserved exactly as written,
   unmodified and unresolved.

---

### User Story 3 - Know Roughly How Much a Piece of Content Will Cost (Priority: P3)

Given any piece of text — typically one of the traceable pieces from
User Story 2, but not required to be — get a consistent, approximate
count of how much language-model context it would consume, so a later
capability can eventually stay within a budget.

**Why this priority**: Useful the moment any piece of content exists,
and independently testable without User Story 1 or 2 (it operates on
any text at all) — but only becomes valuable once there is real content
(User Story 2) to estimate. Ordered last because budgeting itself
remains a later feature's job (Phase 6-7); this story only produces the
number.

**Independent Test**: Estimate the same piece of text twice and confirm
the result is identical both times; estimate two pieces of clearly
different length and confirm the longer one's estimate is larger.

**Acceptance Scenarios**:

1. **Given** a specific piece of text, **When** its cost is estimated
   twice, **Then** both estimates are identical.
2. **Given** two pieces of text of clearly different length, **When**
   both are estimated, **Then** the longer piece's estimate is larger.
3. **Given** an empty piece of text, **When** its cost is estimated,
   **Then** the result is zero, not an error.

---

### Edge Cases

- An artifact body with deeply nested headings (a heading under a
  heading under a heading) — every level must still appear correctly,
  and a parent heading's own piece (if it has any content of its own
  before its first child heading) must not duplicate content already
  captured by its children's own pieces.
- A heading appears inside an inline code span or fenced code block —
  never mistaken for a real section boundary (same discipline
  011-wikilink-foundation already established for wikilink detection).
- Two sections with the exact same heading text at different points in
  the same artifact — both must still appear as distinct pieces, never
  merged or treated as duplicates of each other.
- An unusually large section (far more content than any other section in
  the artifact) — still produces exactly one piece for now; further
  subdividing an oversized section is explicitly deferred (see
  Assumptions) rather than guessed at without a demonstrated real case.
- An artifact whose body is only frontmatter with nothing after it
  (011's own "empty body" case) — produces zero sections and zero
  pieces, not an error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST convert any managed artifact's Markdown body
  into a structured form: an ordered list of sections, each with its own
  heading text, nesting level, and body content.
- **FR-002**: System MUST derive section boundaries from Markdown
  headings — the content between one heading (at any level) and the
  next heading (at any level) belongs to the first heading's own
  section.
- **FR-003**: System MUST NOT treat heading-like text inside an inline
  code span or fenced code block as a real section boundary.
- **FR-004**: An artifact body with no headings at all MUST still
  produce exactly one section covering the entire body, rather than
  being rejected or silently dropped.
- **FR-005**: System MUST break a structured artifact into individual,
  retrieval-sized pieces aligned to its own sections — this feature does
  not perform arbitrary fixed-size splitting independent of section
  structure.
- **FR-006**: Every piece produced MUST carry enough information to
  identify exactly where it came from: its originating artifact, the
  heading/section it belongs to, and its exact starting and ending line
  numbers within that artifact's body.
- **FR-007**: A section with no content between its heading and the next
  boundary MUST NOT produce an empty piece.
- **FR-008**: System MUST produce identical structured output and
  identical pieces for the same, unchanged artifact every time
  (determinism) — no randomness, no dependency on anything other than
  the artifact's own current content.
- **FR-009**: System MUST be able to estimate, for any given piece of
  text, an approximate cost in language-model tokens, using one
  documented, consistently applied method across every use in the
  project — not a real provider-specific tokenizer.
- **FR-010**: This capability MUST be strictly read-only — converting an
  artifact into its structured form or into pieces MUST NOT modify that
  artifact or any other file.
- **FR-011**: An artifact with a completely empty body MUST produce a
  structured form with zero sections and zero pieces, not an error.
- **FR-012**: A piece's own content MUST preserve any explicit link
  (011's wikilink syntax) written inside it exactly as authored — this
  feature does not strip, alter, or resolve links found inside a piece's
  content; that remains 011's and 012's own separate concern.

### Key Entities

- **Document**: The structured representation of one artifact's
  Markdown body — an ordered list of Sections. Distinct from an
  artifact's existing Metadata (frontmatter), which already has its own
  separate, unaffected representation.
- **Section**: One heading-bounded region of a Document — its heading
  text, nesting level, body content, and starting line, in document
  order.
- **Chunk**: One retrieval-sized piece derived from a Section, carrying
  enough provenance (originating artifact, heading, line range) to
  explain exactly where it came from — what a later search/retrieval
  capability (Phase 4+) will actually index and rank.
- **Token Estimate**: A deterministic, approximate count of how many
  language-model tokens a given piece of text would cost — not a real
  tokenizer, but one consistent method used everywhere in the project
  that needs this number.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a representative set of artifacts (multiple headings,
  nested headings, no headings, empty body), converting to structured
  form produces the correct section breakdown with 100% accuracy
  against a known-correct expected set.
- **SC-002**: Converting the same, unchanged artifact into pieces
  repeatedly always produces byte-for-byte identical output — zero
  variation observed across tested fixtures.
- **SC-003**: Every piece produced can be traced back to its exact
  originating artifact, heading, and line range with zero ambiguity.
- **SC-004**: A token estimate is produced for 100% of tested pieces of
  content, with zero errors, using one consistent method throughout.
- **SC-005**: Converting an artifact into its structured form or into
  pieces never modifies it — verified across every tested fixture, zero
  exceptions.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 3 ("Document
  Model and Chunking"): structuring an artifact's body and breaking it
  into traceable pieces, plus a token-estimate capability. It explicitly
  excludes that document's later phases — the disposable SQLite/FTS5
  index (Phase 4), incremental synchronization (Phase 5), the context
  collector and ranking (Phase 6-7), and the internal context command
  (Phase 8) — each remains its own, later feature, continuing this
  project's established one-phase-at-a-time discipline (every prior
  feature, 001 through 012, has kept an equivalent boundary).
- "Any managed artifact" means any Markdown file this project already
  reads a body from via 011-wikilink-foundation's own `ReadBody` — this
  feature does not introduce a new notion of which artifact types are
  "eligible"; it operates on a given artifact's content directly, the
  same way 011's own wikilink extraction does, rather than re-deriving
  012-references-backlinks' own separate "which entity types are
  addressable" scope decision, which answers a different question
  (identity/resolution) than this feature needs to.
- Subdividing an unusually large section into multiple smaller pieces is
  explicitly deferred — the source document itself only suggests this
  as a possibility ("MAY"), and no real oversized-section case has been
  demonstrated yet in this project's own content (Constitution Principle
  IV, YAGNI). A whole section becomes exactly one piece for now.
- The token-estimate capability is a documented approximation, not an
  integration with any real language-model provider's tokenizer — the
  source document itself explicitly allows this for the MVP, as long as
  the same method is used consistently everywhere the project needs this
  number.
- This feature introduces no new relationship or identity concept — a
  piece's provenance is expressed by artifact path, heading, and line
  range, never a new kind of ID; 011's and 012's own ID/link/reference
  vocabulary is completely unaffected and unused by this feature except
  as already-existing input (an artifact's already-parsed body).
