# Feature Specification: References and Backlinks

**Feature Branch**: `012-references-backlinks`
**Created**: 2026-09-14
**Status**: Draft
**Input**: User description: "seguir com a sugestão 'phase 2: references
and backlinks'" (proceed with Phase 2 — References and Backlinks — of
`docs/context-engine-implementation.md`, itself
`docs/architecture-specification.md`'s Phase 7 "Second Brain", building
directly on 011-wikilink-foundation's wikilink parsing/validation)

## User Scenarios & Testing *(mandatory)*

<!--
  As with 001-011, this feature's "users" are misterspec's own
  deterministic layer and future callers — the existing "internal"
  command surface (008-cli-cobra), and eventually Phase 4+'s disposable
  index and Phase 6's context collector, rather than an end-user UI of
  its own. Per context-engine-implementation.md's own Phase 2 scoping,
  this covers querying the artifact graph that already exists on disk
  (formal relationships plus 011's semantic wikilinks) — no SQLite
  index, no chunking, no ranking, no context retrieval.
-->

### User Story 1 - Discover Everything an Artifact Points To (Priority: P1)

Given the ID of an existing artifact, a caller (an agent, a script, or a
future context-retrieval capability) can ask what that artifact
explicitly relates to — every formal relationship already recorded in
its frontmatter (its parent, dependencies, and supersession) and every
semantic wikilink written in its body — as one combined, clearly
labeled answer, without having to separately parse frontmatter and body
themselves.

**Why this priority**: This is the smaller, more foundational half of
the graph query surface — a straightforward "read what's already there"
operation that both directly stands alone (useful today, by hand or by
an agent inspecting a Spec before starting work) and is the exact
building block Backlinks (User Story 2) inverts.

**Independent Test**: Create an artifact with a declared parent, a
`depends_on` entry, and a body containing a wikilink to a third
artifact. Query its references and confirm the parent and dependency
appear labeled as formal relationships, and the wikilink appears labeled
as a semantic one, each naming its exact target.

**Acceptance Scenarios**:

1. **Given** a Spec with a declared parent Feature and a `depends_on`
   entry naming another Spec, **When** its references are queried,
   **Then** both appear, each labeled as a formal relationship naming
   its specific kind (parent, depends_on) and its target.
2. **Given** an artifact whose body contains a wikilink to a Knowledge
   artifact, **When** its references are queried, **Then** the link
   appears labeled as a semantic relationship naming its target.
3. **Given** an artifact with both formal relationships and wikilinks,
   **When** its references are queried, **Then** the result contains
   every one of them, correctly separated into formal and semantic.
4. **Given** an artifact with no formal relationships and no wikilinks
   anywhere, **When** its references are queried, **Then** the result is
   a well-formed, empty answer — not an error.

---

### User Story 2 - Discover Everything That Points To an Artifact (Priority: P2)

Given the ID of an existing artifact, a caller can ask which other
artifacts in the project explicitly reference it — whether through a
formal relationship (for example, another Spec that declares it as a
dependency) or a semantic wikilink — without having to scan the entire
project by hand.

**Why this priority**: Knowing what an artifact depends on (User Story
1) is only half the picture; knowing what depends on it is what makes
the graph actually useful for judging impact before a change, or for
surfacing related context later. It depends on the same underlying
relationship data User Story 1 exposes, just inverted across the whole
project.

**Independent Test**: Create three artifacts where two of them reference
a third — one formally (a `depends_on` entry), one semantically (a
wikilink). Query the third artifact's backlinks and confirm both
referencing artifacts appear, each correctly labeled by how it refers to
it.

**Acceptance Scenarios**:

1. **Given** an artifact that another artifact formally depends on,
   **When** its backlinks are queried, **Then** the referencing artifact
   appears, labeled as a formal incoming relationship.
2. **Given** an artifact that another artifact links to via a wikilink,
   **When** its backlinks are queried, **Then** the referencing artifact
   appears, labeled as a semantic incoming relationship.
3. **Given** an artifact nothing in the project refers to, **When** its
   backlinks are queried, **Then** the result is a well-formed, empty
   answer — not an error.
4. **Given** an artifact referenced by several other artifacts, **When**
   its backlinks are queried, **Then** every one of them appears, with
   no omissions.

---

### User Story 3 - Existing Capabilities and Link-Free Artifacts Remain Unaffected (Priority: P3)

Every artifact and every existing capability (creation, inspection,
structural validation) continues to behave exactly as it did before this
feature existed. Querying references or backlinks is a new, read-only
way to ask a question about the project that already exists on disk — it
changes nothing about how artifacts are authored, stored, or validated.

**Why this priority**: The same discipline 011 established for
wikilinks themselves: a query-only capability layered on top of
existing, already-trusted data must not become a second source of truth
or a hidden side effect. Verified last, once User Story 1 and 2 exist to
prove there's something correct to not regress.

**Independent Test**: Run the full existing suite of creation,
inspection, and structural-validation scenarios (from every prior
capability, 001 through 011) unmodified, and confirm every result is
identical to before this feature existed.

**Acceptance Scenarios**:

1. **Given** a project exactly as it existed before this feature,
   **When** any existing operation (create, inspect, validate) is used,
   **Then** its result is identical to before this feature existed.
2. **Given** an artifact with a broken or malformed wikilink (per 011),
   **When** its references are queried, **Then** the query itself does
   not fail — the same integrity problem 011's validation already
   reports is simply not counted as a resolved reference.

---

### Edge Cases

- An artifact references itself (a self-referencing wikilink, or a
  formal field naming its own ID) — must be reported without error, not
  silently dropped or treated as a crash condition.
- The same target is referenced more than once by the same artifact (for
  example, two separate wikilinks to the same target in different
  sections) — each occurrence is preserved, not silently collapsed into
  one.
- A wikilink whose target is broken (doesn't exist), malformed, or
  ambiguous (resolves to more than one artifact claiming the same ID —
  per 011) — excluded from both "references" and "backlinks" in every
  case: none of the three resolves confidently to exactly one real
  artifact, and 011's own validation already reports each as its own
  distinct integrity problem separately. A query naming an ambiguous ID
  directly (as the queried artifact itself, not as a link's target) is
  rejected the same way every other existing operation already rejects
  an ambiguous ID (see FR-008) — it is never reachable as a query
  target at all.
- An artifact type that doesn't participate in the standalone entity
  graph (for example, Task) is queried — handled consistently with how
  the project's existing operations already treat unsupported targets.
- A large project where many artifacts reference the same popular
  target (for example, a widely depended-on Spec) — the backlink query
  still returns a complete, correctly ordered answer.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST, for any given existing artifact, report every
  outgoing formal relationship already recorded in its frontmatter
  (parent, depends_on, supersedes), each labeled with its specific
  relationship kind and its target.
- **FR-002**: System MUST, for any given existing artifact, report every
  outgoing semantic relationship (011's wikilinks) whose target
  resolves to exactly one real artifact, labeled as semantic and naming
  its target.
- **FR-003**: System MUST combine formal and semantic outgoing
  relationships into one answer per artifact, while keeping the two
  kinds clearly distinguishable from each other.
- **FR-004**: System MUST, for any given existing artifact, report every
  other artifact in the project that formally or semantically references
  it, each labeled by which kind of relationship it uses.
- **FR-005**: System MUST compute both references and backlinks directly
  from the project's current filesystem state — neither MUST depend on
  any separate index or database existing.
- **FR-006**: System MUST return a well-formed, empty answer — never an
  error — for an artifact with no outgoing relationships, and separately
  for an artifact with no incoming ones.
- **FR-007**: System MUST return references and backlinks in a stable,
  deterministic order for the same project state, so repeated queries
  produce identical results.
- **FR-008**: System MUST reject a request naming a malformed or
  unsupported target consistently with how every other existing internal
  operation already reports that same problem — never a second,
  differently shaped error.
- **FR-009**: System MUST NOT count a broken or malformed wikilink (per
  011's own validation) as a resolved reference or backlink — only a
  wikilink whose target actually exists contributes to either answer.
- **FR-010**: System MUST NOT automatically create, modify, or remove any
  formal relationship or wikilink as a side effect of querying references
  or backlinks — this capability is strictly read-only.

### Key Entities

- **Reference**: One outgoing relationship from a source artifact to a
  target artifact, labeled as either formal (parent, depends_on,
  supersedes — the project's existing recorded relationships) or
  semantic (a resolved wikilink, per 011) — never a new third kind of
  relationship.
- **Backlink**: The inverse view of a Reference — for a given target
  artifact, every other artifact that names it, whether formally or
  semantically, discovered by inspecting the whole project rather than a
  single artifact.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a representative set of artifacts with formal
  relationships, semantic wikilinks, or both, querying references
  returns every relationship with 100% accuracy against a known-correct
  expected set, correctly separated into formal and semantic.
- **SC-002**: For a representative set of artifacts referenced by other
  artifacts, querying backlinks returns every referencing artifact with
  100% accuracy against a known-correct expected set — zero omissions.
- **SC-003**: 100% of creation, inspection, and structural-validation
  scenarios that existed before this feature was added continue to
  produce identical results afterward — zero regressions.
- **SC-004**: References and backlinks are available immediately from
  project state on disk, with no separate build or indexing step
  required before the first query.
- **SC-005**: An artifact with zero outgoing or zero incoming
  relationships always produces a well-formed empty answer, with zero
  observed error cases across tested fixtures.

## Assumptions

- This feature is scoped to exactly
  `docs/context-engine-implementation.md`'s own Phase 2 ("References and
  Backlinks"): querying the artifact graph that already exists on disk.
  It explicitly excludes that document's later phases — Markdown
  chunking (Phase 3), the disposable SQLite/FTS5 index (Phase 4),
  incremental synchronization (Phase 5), the context collector and
  ranking (Phase 6-7), and the internal context command (Phase 8) —
  each remains its own, later feature, continuing this project's
  established one-phase-at-a-time discipline (every prior feature, 001
  through 011, has kept an equivalent boundary).
- This feature depends directly on 011-wikilink-foundation's
  `ExtractWikiLinks` (for semantic relationships) and on the project's
  already-existing frontmatter fields — `parent`, `depends_on`,
  `supersedes` — for formal ones (docs/architecture-specification.md
  §5.1's own list). It introduces no new relationship kind beyond those
  two.
- This feature's graph covers exactly the same five entity types
  011-wikilink-foundation's own wikilink validation already covers —
  Program, Feature, Spec, Knowledge, Learning. Plan, Tasks, Validation
  (addressed via a `for:` field naming the Spec they belong to, rather
  than an independent ID of their own), Task, and the Constitution are
  out of scope for this feature's reference/backlink graph — the same
  boundary 011 already established for wikilink validation, kept
  consistent rather than redrawn per feature. A later feature MAY extend
  either graph to them if a real, demonstrated need emerges (Constitution
  Principle IV, YAGNI).
- Backlink discovery may initially scan the project's artifacts directly
  (matching this project's "filesystem as source of truth" discipline,
  001-010); once a later phase's disposable index exists, backlink
  discovery may use it purely as a performance optimization, but its
  correctness must never depend on that index existing —
  context-engine-implementation.md §7.2's own explicit requirement.
- References and backlinks are exposed the same way every prior
  deterministic capability has been — as an internal, machine-readable
  operation following this project's already-established conventions
  (008-cli-cobra) — not as a new kind of user-facing surface with its
  own separate conventions.
- A broken, malformed, or ambiguous wikilink (per 011) is a structural
  problem 011's validation already reports distinctly; this feature does
  not re-report that same problem a second time in a different shape —
  it simply omits what doesn't resolve confidently to exactly one real
  artifact from both "references" and "backlinks" alike (see Edge
  Cases).
